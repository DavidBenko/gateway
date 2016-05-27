package sql

import (
	"errors"
	"fmt"

	"gateway/db"

	_ "github.com/jackc/pgx/stdlib"
)

type sslMode string

const (
	sslModeDisable    sslMode = "disable"
	sslModeAllow      sslMode = "allow"
	sslModePrefer     sslMode = "prefer"
	sslModeRequire    sslMode = "require"
	sslModeVerifyCA   sslMode = "verify-ca"
	sslModeVerifyFull sslMode = "verify-full"
)

var sslModes = map[sslMode]bool{
	sslModeDisable:    true,
	sslModeAllow:      true,
	sslModePrefer:     true,
	sslModeRequire:    true,
	sslModeVerifyCA:   true,
	sslModeVerifyFull: true,
}

// PostgresSpec implements db.Specifier for Postgres connection parameters.
type PostgresSpec struct {
	spec
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbname"`
	Host     string `json:"host"`
	SSLMode  string `json:"sslmode"`
}

func (p *PostgresSpec) validate() error {
	return validate(p, []validation{
		{kw: "port", errCond: p.Port < 0, val: p.Port},
		{kw: "user", errCond: p.User == "", val: p.User},
		{kw: "password", errCond: p.Password == "", val: p.Password},
		{kw: "dbname", errCond: p.DbName == "", val: p.DbName},
		{kw: "host", errCond: p.Host == "", val: p.Host},
		{kw: "sslmode", errCond: !sslModes[sslMode(p.SSLMode)], val: p.SSLMode},
	})
}

func (p *PostgresSpec) driver() driver {
	return postgres
}

func (p *PostgresSpec) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User,
		p.Password,
		p.Host,
		p.Port,
		p.DbName,
		p.SSLMode,
	)
}

func (p *PostgresSpec) UniqueServer() string {
	return p.ConnectionString()
}

func (p *PostgresSpec) NewDB() (db.DB, error) {
	return newDB(p)
}

// UpdateWith validates `pSpec` and updates `p` with its contents if it is
// valid.  This also sets a default SSL mode if one is not provided.
func (p *PostgresSpec) UpdateWith(pSpec *PostgresSpec) error {
	if pSpec == nil {
		return errors.New("cannot update a PostgresSpec with a nil Specifier")
	}
	if pSpec.SSLMode == "" {
		pSpec.SSLMode = string(sslModePrefer)
	}
	if err := pSpec.validate(); err != nil {
		return err
	}
	*p = *pSpec
	return nil
}
