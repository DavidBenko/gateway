package ldap

import (
	"encoding/json"
	"fmt"
)

// Response represents the result of an LDAP operation
type Response struct {
	SearchResult      *SearchResult `json:"searchResults,omitempty"`
	StatusCode        uint8         `json:"statusCode"`
	StatusDescription string        `json:"statusDescription,omitempty"`
}

// JSON satisfies JSON method of request.Response
func (r *Response) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// Log satisfies Log method of request.Response
func (r *Response) Log() string {
	return fmt.Sprintf("%d %s", r.StatusCode, r.StatusDescription)
}
