package sql

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	aperrors "gateway/errors"
	"gateway/stats"
)

// sampleQuery generates the SQL query used to get the sample from the given
// Constraints.  It always appends ORDER BY timestamp, node.
func sampleQuery(
	paramVals func(int) []string,
	constraints []stats.Constraint,
	vars []string,
) string {
	var (
		// At least two time constraints, plus any tags.
		fixedVars        = make([]string, len(vars))
		queryConstraints = make([]string, len(constraints))
		ps               = paramVals(len(constraints))
	)

	for i, v := range vars {
		fixedVars[i] = strings.Replace(v, ".", "_", -1)
	}

	for i, c := range constraints {
		queryConstraints[i] = fmt.Sprintf(
			"%s %s %s",
			strings.Replace(c.Key, ".", "_", -1),
			c.Operator,
			ps[i],
		)
	}

	anyConstraints := ""
	if len(constraints) > 0 {
		anyConstraints = "\nWHERE " +
			strings.Join(queryConstraints, "\n  AND ")
	}

	return fmt.Sprintf(`
SELECT
  %s
FROM stats%s
ORDER BY timestamp, node`[1:],
		strings.Join(fixedVars, "\n  , "),
		anyConstraints,
	)
}

// Sample implements stats.Sampler on SQL.  Querying with no named vars will
// return an error.  All vars must be valid measurements or sample names.
func (s *SQL) Sample(
	constraints []stats.Constraint,
	terminate <-chan struct{},
	vars ...string,
) (stats.Result, error) {
	if len(vars) < 1 {
		return nil, errors.New("no vars given")
	}

	// Make sure all requested vars are legit.  If asked for node or
	// timestamp, we'll set those separately later, so set flags.
	var desiredValues []string
	wantsNode, wantsTimestamp := false, false
	for _, v := range vars {
		if !stats.ValidSample(v) {
			return nil, fmt.Errorf("unknown var %q", v)
		}
		switch {
		case v == "node":
			wantsNode = true
		case v == "timestamp":
			wantsTimestamp = true
		default:
			desiredValues = append(desiredValues, v)
		}
	}

	// Generate the query given the constraints and vars.
	query := sampleQuery(s.Parameters, constraints, vars)

	// Set the constraint args.
	args := make([]interface{}, len(constraints))
	for i, c := range constraints {
		args[i] = c.Value
	}

	// Make the query.
	rows, err := s.Queryx(query, args...)
	switch {
	case err != nil:
		return nil, err
	case rows == nil:
		return nil, errors.New("no rows for stats query")
	}

	// If rows initialized, close when finished.
	defer rows.Close()

	// Get the max procs without changing the setting.
	procs := runtime.GOMAXPROCS(-1)
	var (
		// numRows will determine the size of the stats.Result workers
		// will write to.
		numRows = 0

		inCh       = make(chan ithSQLRow, procs)
		ready, die = make(chan struct{}), make(chan struct{})
		outCh      = make(chan ithStatsRow, procs)
	)

	// If user already requested terminate, stop now.
	select {
	case <-terminate:
		return nil, errors.New("Sample terminated")
	default:
	}

	// Make a worker for each proc which will receive sql.Row's and make
	// them into stats.Row's.  If terminate is closed, the workers will
	// return.
	for i := 0; i < procs; i++ {
		go receiveRows(
			ready, terminate, die,
			inCh, outCh,
			wantsNode, wantsTimestamp,
			desiredValues,
		)
	}

	// Iterate over rows, sending them via channel to workers.
	for i := 0; rows.Next(); i++ {
		row := new(Row)
		if err = rows.StructScan(row); err != nil {
			return nil, aperrors.NewWrapped("failed to scan", err)
		}

		select {
		case <-terminate:
			return nil, errors.New("Sample terminated")
		case inCh <- ithSQLRow{i: i, row: row}:
			numRows++
		}
	}

	if err = rows.Err(); err != nil {
		// terminate is used externally, die is used internally to stop
		// workers.
		close(die)
		return nil, aperrors.NewWrapped("sql stats rows had error", err)
	}

	// Let our workers know they can start sending values.
	close(ready)

	result := make(stats.Result, numRows)

	for i := 0; i < numRows; i++ {
		select {
		case row := <-outCh:
			// Note results won't be received in order.
			result[row.i] = *row.row
		case <-terminate:
			// If a send from a worker to outCh was blocking, it
			// will select its terminate case.
			return nil, errors.New("Sample terminated")
		}
	}

	// Not necessary to stop workers, they've all finished looping and
	// returned by now.
	return result, nil
}

// Note that a SQLRow is not the same as a stats.Row.
type ithSQLRow struct {
	i   int
	row *Row
}

type ithStatsRow struct {
	i   int
	row *stats.Row
}

// receiveRows is a worker function which receives sql.Row's and uses getRow to
// turn them into stats.Row's.
func receiveRows(
	ready, terminate, die <-chan struct{},
	inCh <-chan ithSQLRow, outCh chan<- ithStatsRow,
	wantsNode, wantsTimestamp bool,
	desiredValues []string,
) {
	// Buffer received values in received.
	var received []ithStatsRow

receiveAll:
	for {
		select {
		case in := <-inCh:
			received = append(received, ithStatsRow{
				i: in.i,
				row: getRow(
					in.row,
					wantsNode,
					wantsTimestamp,
					desiredValues,
				),
			})
		case <-terminate:
			return
		case <-die:
			return
		case <-ready:
			// start sending
			break receiveAll
		}
	}

	for _, ithRow := range received {
		select {
		case <-terminate:
			return
		case outCh <- ithRow:
		}
	}
}

// getRow makes a stats.Row out of a sql.Row given the vars the user requested
// and whether to get the node and timestamp.
func getRow(
	row *Row,
	wantsNode, wantsTimestamp bool,
	desiredValues []string,
) *stats.Row {
	var (
		node      string
		timestamp time.Time
	)

	var resultValues map[string]interface{}
	if len(desiredValues) > 0 {
		resultValues = make(map[string]interface{})
	}

	for _, v := range desiredValues {
		resultValues[v] = row.value(v)
	}

	if wantsNode {
		node = row.Node
	}
	if wantsTimestamp {
		timestamp = row.Timestamp.UTC()
	}

	return &stats.Row{
		Node:      node,
		Timestamp: timestamp,
		Values:    resultValues,
	}
}
