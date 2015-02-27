package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse holds an error string.
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
}

// JSON converts this response to JSON format.
func (r *ErrorResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// Log returns the error message
func (r *ErrorResponse) Log() string {
	return fmt.Sprintf("Error: '%s'", r.Error)
}

// NewErrorResponse returns a new response that wraps the error.
func NewErrorResponse(err error) Response {
	return &ErrorResponse{http.StatusInternalServerError, err.Error()}
}
