package sql

import (
	"errors"
	"fmt"
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
	results := make(stats.Result, 0)
	// Iterate over rows.
	for i := 0; rows.Next(); i++ {
		row := new(Row)
		if err = rows.StructScan(row); err != nil {
			return nil, aperrors.NewWrapped("failed to scan", err)
		}
		results = append(results, *convertRow(row, wantsNode, wantsTimestamp, desiredValues))
	}

	if err = rows.Err(); err != nil {
		return nil, aperrors.NewWrapped("sql stats rows had error", err)
	}

	return results, nil
}

// convertRow makes a stats.Row out of a sql.Row given the vars the user requested
// and whether to get the node and timestamp.
func convertRow(
	row *Row,
	wantsNode bool,
	wantsTimestamp bool,
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
