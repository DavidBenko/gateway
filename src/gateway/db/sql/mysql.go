package sql

import (
	"errors"
	"fmt"
	"time"

	"gateway/db"
	"gateway/logger"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLSpec implements db.Specifier for MySQL connections.
type MySQLSpec struct {
	spec
	Username string `json:"username"`
	Password string `json:"password"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Timeout  string `json:"timeout,omitempty"`
	DbName   string `json:"dbname"`
}

func (s *MySQLSpec) driver() driver {
	return mysql
}

func (m *MySQLSpec) validate() error {
	var err error
	if m.Timeout != "" {
		_, err = time.ParseDuration(m.Timeout)
	}
	return validate(m, []validation{
		{kw: "port", errCond: m.Port < 0, val: m.Port},
		{kw: "username", errCond: m.Username == "", val: m.Username},
		{kw: "password", errCond: m.Password == "", val: m.Password},
		{kw: "dbname", errCond: m.DbName == "", val: m.DbName},
		{kw: "server", errCond: m.Server == "", val: m.Server},
		{kw: "timeout", errCond: err != nil, val: m.Timeout, err: err},
	})
}

func (m *MySQLSpec) ConnectionString() string {
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	// Example: user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname?timeout=90s&collation=utf8mb4_unicode_ci
	conn := m.UniqueServer()
	if m.Timeout != "" {
		conn += fmt.Sprintf("?timeout=%s", m.Timeout)
	}

	return conn
}

func (m *MySQLSpec) UniqueServer() string {
	conn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s",
		m.Username,
		m.Password,
		"tcp",
		m.Server,
		m.Port,
		m.DbName,
	)
	return conn
}

func (m *MySQLSpec) NewDB() (db.DB, error) {
	return newDB(m)
}

func (m *MySQLSpec) NeedsUpdate(s db.Specifier) bool {
	if s == nil {
		logger.Panicf("tried to compare to nil db.Specifier!")
		return false
	}
	if tSpec, ok := s.(*MySQLSpec); ok {
		return m.Timeout != tSpec.Timeout || m.spec.NeedsUpdate(s)
	}
	logger.Panicf("tried to compare wrong database kinds: %T and %T", m, s)
	return false
}

func (m *MySQLSpec) Update(s db.Specifier) error {
	spec, ok := s.(*MySQLSpec)
	if !ok {
		return fmt.Errorf("can't update MySQLSpec with %T", s)
	}

	err := spec.validate()
	if err != nil {
		return err
	}

	if spec.Timeout != m.Timeout {
		m.Timeout = spec.Timeout
	}

	return nil
}

// UpdateWith validates `mysqlSpec` and updates `m` with its contents if it is
// valid.
func (m *MySQLSpec) UpdateWith(mysqlSpec *MySQLSpec) error {
	if mysqlSpec == nil {
		return errors.New("cannot update a MySQLSpec with a nil Specifier")
	}
	if err := mysqlSpec.validate(); err != nil {
		return err
	}
	*m = *mysqlSpec
	return nil
}
