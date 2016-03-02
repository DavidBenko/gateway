package ldap

import "github.com/go-ldap/ldap"

// Operation represents an operation against an LDAP server.
type Operation interface {
	Invoke(*ldap.Conn) (*Response, error)
	PrettyString() string
}
