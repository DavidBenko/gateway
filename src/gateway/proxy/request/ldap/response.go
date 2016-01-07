package ldap

import "encoding/json"

// Response TODO
type Response struct {
	SearchResult      *SearchResult `json:"searchResults,omitempty"`
	StatusCode        uint8         `json:"statusCode"`
	StatusDescription string        `json:"statusDescription,omitempty"`
}

// JSON TODO
func (r *Response) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// Log TODO
func (r *Response) Log() string {
	// TODO
	return "Response TODO"
}
