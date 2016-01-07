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

	IncludeByteValue bool `json:"-"`
}

// NewSearchOperation TODO
func NewSearchOperation(options map[string]interface{}) *SearchOperation {
	search := new(SearchOperation)
	if value, ok := options["includeByteValues"]; ok {
		if boolVal, ok := value.(bool); ok {
			search.IncludeByteValue = boolVal
		}
	}
	return search
}

// SearchResult TODO
type SearchResult struct {
	Entries []*Entry `json:"entries"`
}

// NewSearchResult TODO
func NewSearchResult(sr *ldap.SearchResult, includeByteValues bool) *SearchResult {
	res := new(SearchResult)
	for _, entry := range sr.Entries {
		res.Entries = append(res.Entries, NewEntry(entry, includeByteValues))
	}
	return res
}

// Entry TODO
type Entry struct {
	DistinguishedName string            `json:"distinguishedName"`
	Attributes        []*EntryAttribute `json:"attributes"`
}

// NewEntry TODO
func NewEntry(le *ldap.Entry, includeByteValues bool) *Entry {
	e := new(Entry)
	e.DistinguishedName = le.DN
	for _, entryAttribute := range le.Attributes {
		e.Attributes = append(e.Attributes, NewEntryAttribute(entryAttribute, includeByteValues))
	}
	return e
}

// EntryAttribute TODO
type EntryAttribute struct {
	Name       string   `json:"name"`
	Values     []string `json:"values"`
	ByteValues [][]byte `json:"byteValues,omitempty"`
}

// NewEntryAttribute TODO
func NewEntryAttribute(lea *ldap.EntryAttribute, includeByteValues bool) *EntryAttribute {
	ea := new(EntryAttribute)
	ea.Name = lea.Name
	ea.Values = lea.Values
	if includeByteValues {
		ea.ByteValues = lea.ByteValues
	}
	return ea
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
		SearchResult:      NewSearchResult(result, s.IncludeByteValue),
		StatusCode:        ldap.LDAPResultSuccess,
		StatusDescription: ldap.LDAPResultCodeMap[ldap.LDAPResultSuccess],
	}, nil
}
