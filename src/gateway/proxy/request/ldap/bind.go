package ldap

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap"
)

// BindOperation encapsulates an LDAP bind, which authenticates a user for the
// current session
type BindOperation struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PrettyString returns a pretty string representation of the SearchOperation
func (b *BindOperation) PrettyString() string {
	var buf bytes.Buffer
	kv := "  %s: %+v\n"
	buf.WriteString("BindOperation:\n")
	buf.WriteString(fmt.Sprintf(kv, "Username", b.Username))
	buf.WriteString(fmt.Sprintf(kv, "Password", strings.Repeat("*", len(b.Password))))
	return buf.String()
}

// Invoke satisfies apldap.Operation's Invoke method
func (b BindOperation) Invoke(conn *ldap.Conn) (*Response, error) {

	err := conn.Bind(b.Username, b.Password)
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
