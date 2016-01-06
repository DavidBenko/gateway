package ldap

import (
	"fmt"
	"log"
	"strings"

	ld "github.com/go-ldap/ldap"
)

const (
	// ScopeBase TODO
	ScopeBase = Scope("base")
	// ScopeOne TODO
	ScopeOne = Scope("one")
	// ScopeSingle TODO
	ScopeSingle = Scope("single")
	// ScopeSubtree TODO
	ScopeSubtree = Scope("subtree")
)

// Scope TODO
type Scope string

// IntValue TODO
func (s Scope) IntValue() (int, error) {
	switch s {
	case "base":
		return ld.ScopeBaseObject, nil
	case "one":
		return 0, fmt.Errorf("No support for 'one' yet")
	case "single":
		return ld.ScopeSingleLevel, nil
	case "subtree":
		return ld.ScopeWholeSubtree, nil
	default:
		log.Println("error invoking IntValue unmarshaling ")
		return 0, fmt.Errorf("Invalid scope %s", s)
	}
}

// UnmarshalJSON TODO
func (s *Scope) UnmarshalJSON(data []byte) error {
	content := strings.Trim(string(data), `"`)
	switch content {
	case "base":
		*s = ScopeBase
	case "one":
		*s = ScopeOne
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
