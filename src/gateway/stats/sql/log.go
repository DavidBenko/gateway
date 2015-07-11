package sql

import (
	"errors"
	"fmt"
	"log"
	"strings"

	gwerr "gateway/errors"
	"gateway/stats"
)

var fixMeasurements = make(map[string]string)

// totalLen is all measurements plus "ms", "timestamp", and "node"
var totalLen = len(measurements) + 3

func init() {
	for _, v := range measurements {
		fixMeasurements[v] = strings.Replace(v, "_", ".", -1)
	}
}

func logQuery(
	paramVals func(int) []string,
	node string,
	num int,
) string {
	paramNames := strings.Join(
		append([]string{"node", "timestamp", "ms"}, measurements...),
		"\n  , ",
	)

	numParams := totalLen - 1

	paramVs := paramVals(num * numParams)
	params := make([]string, num)

	for i := 0; i < num; i++ {
		params[i] = fmt.Sprintf(
			"('%s', %s)",
			node,
			strings.Join(
				// Subtract 1 since node is accounted for.
				paramVs[i*numParams:(i+1)*numParams],
				", ",
			),
		)
	}

	return fmt.Sprintf(`
INSERT INTO stats (
  %s
) VALUES
  %s
`[1:],
		paramNames,
		strings.Join(params, "\n  , "),
	)
}

func getArgs(ps ...stats.Point) ([]interface{}, error) {
	if len(ps) < 1 {
		return nil, errors.New("must pass at least one stats.Point")
	}

	argLen := totalLen - 1
	// totalLen - 1 accounts for "node" being a constant.
	args := make([]interface{}, len(ps)*argLen)

	for i, p := range ps {
		offset := i * argLen

		// For each Point, the first argument will be timestamp in UTC.
		args[offset] = p.Timestamp.UTC()

		// The next argument will be millis in day.  This is simply for
		// table partitioning, and will not be retrieved in the select.
		args[offset+1] = dayMillis(p.Timestamp.UTC())

		offset += 2

		// All Points must have the full set of Measurements.
		for j, k := range measurements {
			rK := fixMeasurements[k]
			if v, ok := p.Values[rK]; ok {
				args[offset+j] = v
			} else {
				return nil, fmt.Errorf(
					"point missing measurement %q", rK,
				)
			}
		}
	}

	return args, nil
}

// Log implements stats.Logger on SQL.  Note that all Points passed must have
// all measurement values populated, or an error will be returned.
func (s *SQL) Log(ps ...stats.Point) error {
	node := "global"
	if s.ID != "" {
		node = s.ID
	}

	args, err := getArgs(ps...)
	if err != nil {
		return gwerr.NewWrapped(
			"failed to get args for stats query", err,
		)
	}

	query := logQuery(
		s.Parameters,
		node,
		len(ps),
	)

	_, err = s.Exec(query, args...)
	if err != nil {
		log.Println("\n" + query)
		return gwerr.NewWrapped("failed to exec stats query", err)
	}

	return nil
}
