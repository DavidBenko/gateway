package rest

import "net/http"

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
