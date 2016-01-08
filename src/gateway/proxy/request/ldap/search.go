package ldap

import (
	"bytes"
	"fmt"

	"github.com/go-ldap/ldap"
)

// SearchOperation encapsulates an LDAP search
type SearchOperation struct {
	BaseDistinguishedName string      `json:"baseDistinguishedName"`
	Scope                 Scope       `json:"scope"`
	DereferenceAliases    Dereference `json:"dereferenceAliases"`
	SizeLimit             int         `json:"sizeLimit"`
	TimeLimit             int         `json:"timeLimit"`
	TypesOnly             bool        `json:"typesOnly"`
	Filter                string      `json:"filter"`
	Attributes            []string    `json:"attributes"`

	IncludeByteValue bool `json:"-"`
}

// NewSearchOperation creates a new SearchOperation, examining the options and
// setting appropriate values on the created struct
func NewSearchOperation(options map[string]interface{}) *SearchOperation {
	search := new(SearchOperation)
	if value, ok := options["includeByteValues"]; ok {
		if boolVal, ok := value.(bool); ok {
			search.IncludeByteValue = boolVal
		}
	}

	return search
}

// PrettyString returns a pretty string representation of the SearchOperation
func (s *SearchOperation) PrettyString() string {
	var buf bytes.Buffer
	kv := "  %s: %+v\n"
	buf.WriteString("SearchOperation:\n")
	buf.WriteString(fmt.Sprintf(kv, "BaseDistinguishedName", s.BaseDistinguishedName))
	buf.WriteString(fmt.Sprintf(kv, "Scope", s.Scope))
	buf.WriteString(fmt.Sprintf(kv, "DereferenceAliases", s.DereferenceAliases))
	buf.WriteString(fmt.Sprintf(kv, "SizeLimit", s.SizeLimit))
	buf.WriteString(fmt.Sprintf(kv, "TimeLimit", s.TimeLimit))
	buf.WriteString(fmt.Sprintf(kv, "Filter", s.Filter))
	buf.WriteString(fmt.Sprintf(kv, "Attributes", s.Attributes))
	buf.WriteString(fmt.Sprintf(kv, "IncludeByteValue", s.IncludeByteValue))
	return buf.String()
}

// Invoke satisfies LDAPOperation's Invoke method
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
		nil,
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

// SearchResult represents the results of an LDAP search operation
type SearchResult struct {
	Entries          []*Entry `json:"entries"`
	SearchReferences []string `json:"searchReferences,omitempty"` // search references are references to another LDAP server
	Controls         []string `json:"controls,omitempty"`
}

// NewSearchResult creates a new SearchResult
func NewSearchResult(sr *ldap.SearchResult, includeByteValues bool) *SearchResult {
	res := new(SearchResult)
	for _, entry := range sr.Entries {
		res.Entries = append(res.Entries, NewEntry(entry, includeByteValues))
	}
	for _, referral := range sr.Referrals {
		res.SearchReferences = append(res.SearchReferences, referral)
	}
	for _, control := range sr.Controls {
		res.Controls = append(res.Controls, control.String())
	}
	return res
}

// Entry represents an individual entry in the results returned by LDAP search
type Entry struct {
	DistinguishedName string            `json:"distinguishedName"`
	Attributes        []*EntryAttribute `json:"attributes"`
}

// NewEntry creates a new Entry
func NewEntry(le *ldap.Entry, includeByteValues bool) *Entry {
	e := new(Entry)
	e.DistinguishedName = le.DN
	for _, entryAttribute := range le.Attributes {
		e.Attributes = append(e.Attributes, NewEntryAttribute(entryAttribute, includeByteValues))
	}
	return e
}

// EntryAttribute represents an attribute on an individual search result Entry
type EntryAttribute struct {
	Name       string   `json:"name"`
	Values     []string `json:"values"`
	ByteValues [][]byte `json:"byteValues,omitempty"`
}

// NewEntryAttribute creates a new EntryAttribute
func NewEntryAttribute(lea *ldap.EntryAttribute, includeByteValues bool) *EntryAttribute {
	ea := new(EntryAttribute)
	ea.Name = lea.Name
	ea.Values = lea.Values
	if includeByteValues {
		ea.ByteValues = lea.ByteValues
	}
	return ea
}
