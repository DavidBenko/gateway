package requests

import "encoding/json"

// ErrorResponse holds an error string.
type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON converts this response to JSON format.
func (r *ErrorResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// NewErrorResponse returns a new response that wraps the error.
func NewErrorResponse(err error) Response {
	return &ErrorResponse{Error: err.Error()}
}