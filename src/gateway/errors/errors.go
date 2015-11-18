package errors

import "fmt"

// Errors represents API-serializable validation errors.
type Errors map[string][]string

func (e Errors) Add(name, message string) {
	e[name] = append(e[name], message)
}

// Empty reports if there are no errors.
func (e Errors) Empty() bool {
	return len(e) == 0
}

func (e Errors) String() string {
	errs, sep := "", ""
	for _, i := range e {
		for _, j := range i {
			errs += sep + j
			sep = " "
		}
	}
	return errs
}

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
