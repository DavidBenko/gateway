package ldap

import (
	"fmt"
	"strings"

	ld "github.com/go-ldap/ldap"
)

// Dereference TODO
type Dereference string

const (
	// DereferenceNever TODO
	DereferenceNever = Dereference("never")
	// DereferenceInSearch TODO
	DereferenceInSearch = Dereference("search")
	// DereferenceFindingBaseObj TODO
	DereferenceFindingBaseObj = Dereference("find")
	// DereferenceAlways TODO
	DereferenceAlways = Dereference("always")
)

// IntValue TODO
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

// UnmarshalJSON TODO
func (d *Dereference) UnmarshalJSON(data []byte) error {
	//log.Printf("-- UnmarshalJSON -- data is %s", string(data))
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
