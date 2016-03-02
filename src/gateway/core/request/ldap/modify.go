package ldap

import (
	"bytes"
	"fmt"

	"github.com/go-ldap/ldap"
)

// ModifyOperation encapsulates an LDAP add operation
type ModifyOperation struct {
	DistinguishedName string      `json:"distinguishedName"`
	AddAttributes     []Attribute `json:"addAttributes"`
	DeleteAttributes  []Attribute `json:"deleteAttributes"`
	ReplaceAttributes []Attribute `json:"replaceAttributes"`
}

// ConvertToLDAPModifyRequest converts a ModifyOperation to an ldap.ModifyRequest
func ConvertToLDAPModifyRequest(modOp *ModifyOperation) *ldap.ModifyRequest {
	modReq := ldap.NewModifyRequest(modOp.DistinguishedName)
	for _, add := range modOp.AddAttributes {
		modReq.Add(add.Type, add.Values)
	}
	for _, del := range modOp.DeleteAttributes {
		modReq.Delete(del.Type, del.Values)
	}
	for _, repl := range modOp.ReplaceAttributes {
		modReq.Replace(repl.Type, repl.Values)
	}
	return modReq
}

// PrettyString returns a pretty string representation of the SearchOperation
func (m *ModifyOperation) PrettyString() string {
	var buf bytes.Buffer
	buf.WriteString("ModifyOperation:\n")

	buf.WriteString(fmt.Sprintf("  %s: %+v\n", "DistinguishedName", m.DistinguishedName))
	buf.WriteString("  AddAttributes:\n")
	for _, attr := range m.AddAttributes {
		buf.WriteString("    Attribute:\n")
		buf.WriteString(fmt.Sprintf("      Type: %s\n", attr.Type))
		buf.WriteString("      Values:\n")
		for _, v := range attr.Values {
			buf.WriteString(fmt.Sprintf("        Value: %s\n", v))
		}
	}
	for _, attr := range m.DeleteAttributes {
		buf.WriteString("    DeleteAttribute:\n")
		buf.WriteString(fmt.Sprintf("      Type: %s\n", attr.Type))
		buf.WriteString("      Values:\n")
		for _, v := range attr.Values {
			buf.WriteString(fmt.Sprintf("        Value: %s\n", v))
		}
	}
	for _, attr := range m.ReplaceAttributes {
		buf.WriteString("    ReplaceAttribute:\n")
		buf.WriteString(fmt.Sprintf("      Type: %s\n", attr.Type))
		buf.WriteString("      Values:\n")
		for _, v := range attr.Values {
			buf.WriteString(fmt.Sprintf("        Value: %s\n", v))
		}
	}

	return buf.String()
}

// Invoke satisfies apldap.Operation's Invoke method
func (m ModifyOperation) Invoke(conn *ldap.Conn) (*Response, error) {
	err := conn.Modify(ConvertToLDAPModifyRequest(&m))

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
