package sql

import (
	"errors"
	"fmt"
	"strings"
	"time"

	gwerr "gateway/errors"
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

// Sample implements stats.Sampler on SQL.  Note that it only supports certain
// tags: "node", "api_id", and "user_id".  It auto-generates the query based on
// passed constraints and vars -- don't expose these to the user, or validate
// their values if you must expose them.
func (s *SQL) Sample(
	constraints []stats.Constraint,
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

	var result stats.Result

	for rowNum := 0; rows.Next(); rowNum++ {
		var (
			row       Row
			node      string
			timestamp time.Time
		)

		if err = rows.StructScan(&row); err != nil {
			return nil, gwerr.NewWrapped("failed to scan", err)
		}

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

		result = append(result, stats.Row{
			Node:      node,
			Timestamp: timestamp,
			Values:    resultValues,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, gwerr.NewWrapped("sql stats rows had error", err)
	}

	return result, nil
}
