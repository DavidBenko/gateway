package errors

import "fmt"

type WrappedError struct {
	context string
	err     error
}

func NewWrapped(context string, err error) *WrappedError {
	return &WrappedError{context: context, err: err}
}

// WrapErrors wraps an old error with a new error and an explanation.
func WrapErrors(context string, old, err error) *WrappedError {
	return &WrappedError{
		context: context,
		err:     NewWrapped(err.Error(), old),
	}
}

func (w *WrappedError) Error() string {
	return fmt.Sprintf("%s: %s", w.context, w.err.Error())
}

func (w *WrappedError) PrettyError() string {
	return fmt.Sprintf("%s: \n%s", w.context, w.err.Error())
}
