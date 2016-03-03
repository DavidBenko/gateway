package ldap

import "github.com/go-ldap/ldap"

// ConnectionAdapter is a wrapper for an ldap.Conn which implements the io.Closer interface
type ConnectionAdapter struct {
	Conn *ldap.Conn
}

// Close closes the ldap.Conn
func (a *ConnectionAdapter) Close() error {
	a.Conn.Close()
	return nil
}
