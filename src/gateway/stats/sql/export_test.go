package sql

import (
	"gateway/stats"
	"time"
)

// LogQuery exports logQuery for testing.
func LogQuery(
	f func(int) []string,
	node string,
	n int,
) string {
	return logQuery(f, node, n)
}

// SampleQuery exports sampleQuery for testing.
func SampleQuery(
	f func(int) []string,
	vars, tags []string,
) string {
	return sampleQuery(f, vars, tags)
}

// GetArgs exports getArgs for testing.
func GetArgs(ps ...stats.Point) ([]interface{}, error) {
	return getArgs(ps...)
}

// DayMillis exports dayMillis for testing.
func DayMillis(t time.Time) int64 {
	return dayMillis(t)
}
