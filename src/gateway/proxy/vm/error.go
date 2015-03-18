package vm

import "fmt"

type jsError struct {
	err  error
	code interface{}
}

func (e *jsError) Error() string {
	return fmt.Sprintf("JavaScript Error: %v\n\n--\n\n%v", e.err, e.code)
}
