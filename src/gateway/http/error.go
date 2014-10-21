package http

import (
	"fmt"
	"net/http"
)

// Error is an interface that describes an error case.
type Error interface {
	Error() error
	Code() int
}

type httpError struct {
	err  error
	code int
}

// Error returns the underlying error.
func (h *httpError) Error() error {
	return h.err
}

// Code returns the HTTP status code of the error. Defaults to 500.
func (h *httpError) Code() int {
	if h.code == 0 {
		return http.StatusInternalServerError
	}
	return h.code
}

// NewError returns a new error to use with this library.
func NewError(err error, code int) Error {
	return &httpError{err: err, code: code}
}

// NewServerError returns a new error with standard code.
func NewServerError(err error) Error {
	return NewError(err, 0)
}

// ErrorReturningHandler is an http.Handler that can return an error
type ErrorReturningHandler func(w http.ResponseWriter, r *http.Request) Error

// ErrorCatchingHandler catches an error a handler throws and responds with it.
func ErrorCatchingHandler(handler ErrorReturningHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			if err != nil {
				http.Error(w, fmt.Sprintf("%v\n", err.Error()), err.Code())
			}
		}
	})
}
