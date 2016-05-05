package sql

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	"gateway/db"
	"gateway/logreport"

	_ "github.com/denisenkom/go-mssqldb"
)

// SQLServerSpec implements db.Specifier for MSSQL connection string parameters.
type SQLServerSpec struct {
	spec
	Server   string `json:"server"`
	Port     int    `json:"port"`
	UserId   string `json:"user id"`
	Password string `json:"password"`
	Database string `json:"database"`
	Schema   string `json:"schema,omitempty"`
	Timeout  int    `json:"connection timeout,omitempty"`
	Encrypt  string `json:"encrypt,omitempty"`
}

func (s *SQLServerSpec) driver() driver {
	return mssql
}

func (s *SQLServerSpec) validate() error {
	return validate(s, []validation{
		{kw: "port", errCond: s.Port < 0, val: s.Port},
		{kw: "user id", errCond: s.UserId == "", val: s.UserId},
		{kw: "password", errCond: s.Password == "", val: s.Password},
		{kw: "database", errCond: s.Database == "", val: s.Database},
		{kw: "server", errCond: s.Server == "", val: s.Server},
		{kw: "timeout", errCond: s.Timeout < 0, val: s.Timeout},
		{kw: "encrypt", errCond: s.Encrypt != "" && s.Encrypt != "true" && s.Encrypt != "false" && s.Encrypt != "disable", val: s.Encrypt},
	})
}

func (s *SQLServerSpec) ConnectionString() string {
	m := map[string]interface{}{
		"port":     s.Port,
		"user id":  s.UserId,
		"password": s.Password,
		"database": s.Database,
		"server":   s.Server,
	}

	if s.Encrypt != "" {
		m["encrypt"] = s.Encrypt
	}
	if s.Schema != "" {
		m["schema"] = s.Schema
	}

	if s.Timeout > 0 {
		m["timeout"] = s.Timeout
	}

	return s.serialize(m)
}

func (s *SQLServerSpec) UniqueServer() string {
	return s.serialize(map[string]interface{}{
		"port":     s.Port,
		"user id":  s.UserId,
		"password": s.Password,
		"dbname":   s.Database,
		"host":     s.Server,
	})
}

func (s *SQLServerSpec) serialize(m map[string]interface{}) string {
	keys := make([]string, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	// Iterate over the sorted config and get values.  Escape and quote
	// as needed.
	var buffer bytes.Buffer
	for _, key := range keys {
		val := fmt.Sprintf("%v", m[key])
		buffer.WriteString(fmt.Sprintf("%s=%s;", key, val))
	}

	str := buffer.String()
	return str[:len(str)-1]
}

func (s *SQLServerSpec) NewDB() (db.DB, error) {
	return newDB(s)
}

func (s *SQLServerSpec) NeedsUpdate(sp db.Specifier) bool {
	if sp == nil {
		logreport.Panicf("tried to compare to nil db.Specifier!")
		return false
	}
	if spec, ok := sp.(*SQLServerSpec); ok {
		return spec.Timeout != s.Timeout || spec.Encrypt != s.Encrypt || spec.Schema != s.Schema || s.spec.NeedsUpdate(sp)
	}

	logreport.Panicf("tried to compare %T to %T!", s, sp)
	return false
}

func (s *SQLServerSpec) Update(sp db.Specifier) error {
	spec, ok := sp.(*SQLServerSpec)
	if !ok {
		return fmt.Errorf("can't update SQLServerSpec with %T", s)
	}

	err := spec.validate()
	if err != nil {
		return err
	}

	s.Timeout = spec.Timeout
	s.Encrypt = spec.Encrypt
	s.Schema = spec.Schema
	return nil
}

// UpdateWith validates the given sqlsSpec and updates the caller with it.
func (s *SQLServerSpec) UpdateWith(sqlsSpec *SQLServerSpec) error {
	if sqlsSpec == nil {
		return errors.New("cannot update a SQLServerSpec with a nil Specifier")
	}
	if err := sqlsSpec.validate(); err != nil {
		return err
	}
	*s = *sqlsSpec
	return nil
}
