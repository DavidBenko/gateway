package ldap

import (
	"fmt"
	"log"
	"strings"

	ld "github.com/go-ldap/ldap"
)

const (
	// ScopeBase indicates that search will be restricted to the base value
	ScopeBase = Scope("base")
	// ScopeSingle indicates that search will be restricted to a single level
	ScopeSingle = Scope("single")
	// ScopeSubtree indicates that search will be conducted on the entire subtree
	ScopeSubtree = Scope("subtree")
)

// Scope represents possible scopes that can be specified in a search
type Scope string

// IntValue returns the go-ldap/ldap library's corresponding int value that matches
// the given Scope value
func (s Scope) IntValue() (int, error) {
	switch s {
	case "base":
		return ld.ScopeBaseObject, nil
	case "single":
		return ld.ScopeSingleLevel, nil
	case "subtree":
		return ld.ScopeWholeSubtree, nil
	default:
		log.Println("error invoking IntValue unmarshaling ")
		return 0, fmt.Errorf("Invalid scope %s", s)
	}
}

// UnmarshalJSON unmarshals the JSON into a Scope
func (s *Scope) UnmarshalJSON(data []byte) error {
	content := strings.Trim(string(data), `"`)
	switch content {
	case "base":
		*s = ScopeBase
	case "single":
		*s = ScopeSingle
	case "subtree":
		*s = ScopeSubtree
	default:
		log.Println("Invalid scope unmarshaling")
		return fmt.Errorf("Invalid scope: %s", content)
	}
	return nil
}
