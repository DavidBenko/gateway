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
	vars, tags []string,
) string {
	var (
		// At least two time constraints, plus any tags.
		ps          = paramVals(len(tags) + 2)
		fixedVars   = make([]string, len(vars))
		constraints = make([]string, len(tags)+1)
	)

	for i, v := range vars {
		fixedVars[i] = strings.Replace(v, ".", "_", -1)
	}

	for i, t := range tags {
		constraints[i] = fmt.Sprintf(
			"%s = %s",
			strings.Replace(t, ".", "_", -1),
			ps[i],
		)
	}

	constraints[len(constraints)-1] = fmt.Sprintf(
		"timestamp >= %s AND timestamp < %s",
		ps[len(ps)-2],
		ps[len(ps)-1],
	)

	return fmt.Sprintf(`
SELECT
  %s
FROM stats
WHERE %s
ORDER BY timestamp, node`[1:],
		strings.Join(fixedVars, "\n  , "),
		strings.Join(constraints, "\n  AND "),
	)
}

// Sample implements stats.Sampler on SQL.  Note that it only supports certain
// tags: "node", "api_id", and "user_id".  It auto-generates the query based on
// passed tags and vars -- don't expose these to the user, or validate their
// values if you must expose them.
func (s *SQL) Sample(
	tags map[string]interface{},
	from time.Time,
	to time.Time,
	measurements ...string,
) (stats.Result, error) {
	// Make sure timestamps are in correct order.
	if !from.Before(to) {
		return nil, fmt.Errorf("time %s is not after %s", to, from)
	}

	if len(measurements) < 1 {
		return nil, errors.New("no measurements given")
	}

	var desiredValues []string
	wantsNode, wantsTimestamp := false, false
	for _, m := range measurements {
		switch {
		case m == "node":
			wantsNode = true
		case m == "timestamp":
			wantsTimestamp = true
		default:
			desiredValues = append(desiredValues, m)
		}
	}

	var tagParams []string
	for t := range tags {
		tagParams = append(tagParams, t)
	}

	query := sampleQuery(s.Parameters, measurements, tagParams)

	args := make([]interface{}, len(tags))
	for i, t := range tagParams {
		args[i] = tags[t]
	}

	rows, err := s.Queryx(query, append(args, from.UTC(), to.UTC())...)
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
