package ldap

import (
	"encoding/json"

	"github.com/go-ldap/ldap"
)

// Response TODO
type Response struct {
	SearchResult *ldap.SearchResult `json:"searchResults"`
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
