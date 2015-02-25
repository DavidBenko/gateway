package errors

import "fmt"

type WrappedError struct {
	context string
	err     error
}

func NewWrapped(context string, err error) *WrappedError {
	return &WrappedError{context: context, err: err}
}

func (w *WrappedError) Error() string {
	return fmt.Sprintf("%s: %s", w.context, w.err.Error())
}

func (w *WrappedError) PrettyError() string {
	return fmt.Sprintf("%s: \n%s", w.context, w.err.Error())
}
