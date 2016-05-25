package testing

import "gateway/stats"

// Logger is a stats.Logger which "logs" points to a slice which can be checked.
// An error on Log can be artificially triggered by setting its Error field.
type Logger struct {
	Error  error
	Buffer []stats.Point
}

// Log implements stats.Logger.Log on Logger.  It is assumed to be used
// synchronously.
func (l *Logger) Log(ps ...stats.Point) error {
	l.Buffer = append(l.Buffer, ps...)
	return l.Error
}
