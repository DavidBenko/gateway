package ldap

import (
	"fmt"
	"strings"

	ld "github.com/go-ldap/ldap"
)

// Dereference is an enumeration of possible values for whether or not to
// dereference values
type Dereference string

const (
	// DereferenceNever means that values are never dereferenced
	DereferenceNever = Dereference("never")
	// DereferenceInSearch means that values are dereferenced in search results, but not otherwise
	DereferenceInSearch = Dereference("search")
	// DereferenceFindingBaseObj means that values are dereferenced when finding the base object, but otherwise not
	DereferenceFindingBaseObj = Dereference("find")
	// DereferenceAlways means that values are always dereferenced
	DereferenceAlways = Dereference("always")
)

// IntValue returns the appropriate integer value for the given possible value of type Dereference
func (d Dereference) IntValue() (int, error) {
	switch d {
	case "never":
		return ld.NeverDerefAliases, nil
	case "search":
		return ld.DerefInSearching, nil
	case "find":
		return ld.DerefFindingBaseObj, nil
	case "always":
		return ld.DerefAlways, nil
	default:
		return 0, fmt.Errorf("Invalid dereference %s", d)
	}
}

// UnmarshalJSON unmarshals a JSON value into a Dereference
func (d *Dereference) UnmarshalJSON(data []byte) error {
	content := strings.Trim(string(data), `"`)
	switch content {
	case "never":
		*d = DereferenceNever
	case "search":
		*d = DereferenceInSearch
	case "find":
		*d = DereferenceFindingBaseObj
	case "always":
		*d = DereferenceAlways
	default:

		return fmt.Errorf("Invalid dereference: %s", content)
	}
	return nil
}
