package ldap

import (
	"bytes"
	"fmt"

	"github.com/go-ldap/ldap"
)

// AddOperation encapsulates an LDAP add operation
type AddOperation struct {
	DistinguishedName string       `json:"distinguishedName"`
	Attributes        []*Attribute `json:"attributes"`
}

func ConvertToLDAPAddRequest(addOp *AddOperation) *ldap.AddRequest {
	addReq := ldap.NewAddRequest(addOp.DistinguishedName)
	for _, attr := range addOp.Attributes {
		addReq.Attribute(attr.Type, attr.Values)
	}
	return addReq
}

// Attribute represents an LDAP attribute
type Attribute struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}

// PrettyString returns a pretty string representation of the SearchOperation
func (a *AddOperation) PrettyString() string {
	var buf bytes.Buffer
	buf.WriteString("AddOperation:\n")
	buf.WriteString(fmt.Sprintf("  %s: %+v\n", "DistinguishedName", a.DistinguishedName))
	buf.WriteString("  Attributes:\n")
	for _, attr := range a.Attributes {
		buf.WriteString("    Attribute:\n")
		buf.WriteString(fmt.Sprintf("      Type: %s\n", attr.Type))
		buf.WriteString("      Values:\n")
		for _, v := range attr.Values {
			buf.WriteString(fmt.Sprintf("        Value: %s\n", v))
		}
	}
	return buf.String()
}

// Invoke satisfies apldap.Operation's Invoke method
func (a AddOperation) Invoke(conn *ldap.Conn) (*Response, error) {
	addReq := ConvertToLDAPAddRequest(&a)
	err := conn.Add(addReq)

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
