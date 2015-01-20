package sql

import "strings"

// NQs returns n comma separated '?'s
func NQs(n int) string {
	return strings.Join(strings.Split(strings.Repeat("?", n), ""), ",")
}
