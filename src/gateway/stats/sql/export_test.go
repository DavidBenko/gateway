package sql

import (
	"gateway/stats"
	"time"
)

// LogQuery exports logQuery for testing.
func LogQuery(f func(int) []string, n int) string {
	return logQuery(f, n)
}

// SampleQuery exports sampleQuery for testing.
func SampleQuery(
	f func(int) []string,
	constraints []stats.Constraint,
	vars []string,
) string {
	return sampleQuery(f, constraints, vars)
}

// GetArgs exports getArgs for testing.
func GetArgs(node string, ps ...stats.Point) ([]interface{}, error) {
	return getArgs(node, ps...)
}

// DayMillis exports dayMillis for testing.
func DayMillis(t time.Time) int64 {
	return dayMillis(t)
}

// Value exports Row.value for testing.
func (r *Row) Value(k string) interface{} {
	return r.value(k)
}

func (r *Row) Values(ks ...string) map[string]interface{} {
	if len(ks) == 0 {
		return nil
	}

	result := make(map[string]interface{})
	for _, k := range ks {
		result[k] = r.Value(k)
	}
	return result
}
