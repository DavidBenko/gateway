package ldap

import (
	"bytes"
	"fmt"

	"github.com/go-ldap/ldap"
)

// CompareOperation encapsulates an LDAP compare operation
type CompareOperation struct {
	DistinguishedName string `json:"distinguishedName"`
	Attribute         string `json:"attribute"`
	Value             string `json:"value"`
}

// PrettyString returns a pretty string representation of the SearchOperation
func (c *CompareOperation) PrettyString() string {
	var buf bytes.Buffer
	kv := "  %s: %+v\n"
	buf.WriteString("CompareOperation:\n")
	buf.WriteString(fmt.Sprintf(kv, "DistinguishedName", c.DistinguishedName))
	buf.WriteString(fmt.Sprintf(kv, "Attribute", c.Attribute))
	buf.WriteString(fmt.Sprintf(kv, "Value", c.Value))
	return buf.String()
}

// Invoke satisfies apldap.Operation's Invoke method
func (c CompareOperation) Invoke(conn *ldap.Conn) (*Response, error) {
	matches, err := conn.Compare(c.DistinguishedName, c.Attribute, c.Value)

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
		CompareResult:     &CompareResult{matches},
		StatusCode:        ldap.LDAPResultSuccess,
		StatusDescription: ldap.LDAPResultCodeMap[ldap.LDAPResultSuccess],
	}, nil
}

// CompareResult encapsulates the results of a compare operation
type CompareResult struct {
	Matches bool `json:"matches"`
}
