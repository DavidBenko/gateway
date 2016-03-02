package ldap

// Attribute represents an LDAP attribute
type Attribute struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
}
