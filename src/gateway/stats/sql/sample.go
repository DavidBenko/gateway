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

// Sample implements stats.Sampler on SQL.
func (s *SQL) Sample(
	constraints []stats.Constraint,
	terminate <-chan struct{},
	vars ...string,
) (stats.Result, error) {
	if len(vars) < 1 {
		return nil, errors.New("no vars given")
	}

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

	query := sampleQuery(s.Parameters, constraints, vars)

	args := make([]interface{}, len(constraints))
	for i, c := range constraints {
		args[i] = c.Value
	}

	rows, err := s.Queryx(query, args...)
	switch {
	case err != nil:
		return nil, err
	case rows == nil:
		return nil, errors.New("no rows for stats query")
	}

	defer rows.Close()

	// Get the max procs without changing the setting.
	procs := runtime.GOMAXPROCS(-1)
	var (
		numRows = 0

		inCh  = make(chan ithSQLRow, procs)
		ready = make(chan struct{})
		outCh = make(chan ithStatsRow, procs)
	)

	select {
	case <-terminate:
		return nil, errors.New("Sample terminated")
	default:
	}

	for i := 0; i < procs; i++ {
		go receiveRows(
			ready,
			terminate,
			inCh,
			outCh,
			wantsNode,
			wantsTimestamp,
			desiredValues,
		)
	}

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
		return nil, aperrors.NewWrapped("sql stats rows had error", err)
	}

	close(ready)

	result := make(stats.Result, numRows)

	for i := 0; i < numRows; i++ {
		select {
		case row := <-outCh:
			result[row.i] = *row.row
		case <-terminate:
			// If a send from a worker to outCh was blocking, it
			// will select its terminate case.
			return nil, errors.New("Sample terminated")
		}
	}

	return result, nil
}

type ithSQLRow struct {
	i   int
	row *Row
}

type ithStatsRow struct {
	i   int
	row *stats.Row
}

func receiveRows(
	ready, terminate <-chan struct{},
	inCh <-chan ithSQLRow,
	outCh chan<- ithStatsRow,
	wantsNode, wantsTimestamp bool,
	desiredValues []string,
) {
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
