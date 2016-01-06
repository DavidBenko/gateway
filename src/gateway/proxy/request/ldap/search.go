package ldap

import "github.com/go-ldap/ldap"

// SearchOperation TODO
type SearchOperation struct {
	BaseDistinguishedName string      `json:"baseDistinguishedName"`
	Scope                 Scope       `json:"scope"`
	DereferenceAliases    Dereference `json:"dereferenceAliases"`
	SizeLimit             int         `json:"sizeLimit"`
	TimeLimit             int         `json:"timeLimit"`
	TypesOnly             bool        `json:"typesOnly"`
	Filter                string      `json:"filter"`
	Attributes            []string    `json:"attributes"`
	Controls              []string    `json:"controls"`
}

// Invoke TODO
func (s SearchOperation) Invoke(conn *ldap.Conn) (*Response, error) {
	var scope int
	var deref int
	var err error

	if scope, err = s.Scope.IntValue(); err != nil {
		return nil, err
	}

	if deref, err = s.DereferenceAliases.IntValue(); err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		s.BaseDistinguishedName,
		scope,
		deref,
		s.SizeLimit,
		s.TimeLimit,
		s.TypesOnly,
		s.Filter,
		s.Attributes,
		nil, // TODO - add controls
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	return &Response{SearchResult: result}, nil
}
