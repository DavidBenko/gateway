package ldap

import (
	"bytes"
	"fmt"

	"github.com/go-ldap/ldap"
)

// DeleteOperation encapsulates an LDAP add operation
type DeleteOperation struct {
	DistinguishedName string `json:"distinguishedName"`
}

// PrettyString returns a pretty string representation of the SearchOperation
func (d *DeleteOperation) PrettyString() string {
	var buf bytes.Buffer
	buf.WriteString("DeleteOperation:\n")

	buf.WriteString(fmt.Sprintf("  %s: %+v\n", "DistinguishedName", d.DistinguishedName))
	return buf.String()
}

// Invoke satisfies apldap.Operation's Invoke method
func (d DeleteOperation) Invoke(conn *ldap.Conn) (*Response, error) {
	err := conn.Del(ldap.NewDelRequest(d.DistinguishedName, []ldap.Control{}))

	if err != nil {
		var e *ldap.Error
		var ok bool
		if e, ok = err.(*ldap.Error); !ok {
			return nil, err
		}
		return &Response{
			StatusCode:        e.ResultCode,
			StatusDescription: ldap.LDAPResultCodeMap[e.ResultCode],
		}, nil
	}

	return &Response{
		StatusCode:        ldap.LDAPResultSuccess,
		StatusDescription: ldap.LDAPResultCodeMap[ldap.LDAPResultSuccess],
	}, nil
}
