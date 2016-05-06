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
