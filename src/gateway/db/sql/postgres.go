package sql

import (
	"errors"
	"fmt"
	"regexp"

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

var spaces *regexp.Regexp
var escapeChars *regexp.Regexp
var sslModes *regexp.Regexp

// init compiles non-unique keys when the package is loaded.
func init() {
	spaces = regexp.MustCompile(" ")
	escapeChars = regexp.MustCompile("'")

	sslModeRe := ""
	for _, mode := range []string{
		string(sslModeDisable),
		string(sslModeAllow),
		string(sslModePrefer),
		string(sslModeRequire),
		string(sslModeVerifyCA),
		string(sslModeVerifyFull),
	} {
		sslModeRe += fmt.Sprintf("%s|", mode)
	}
	sslModes = regexp.MustCompile(sslModeRe[:len(sslModeRe)-1])
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
	if p.SSLMode == "" {
		p.SSLMode = string(sslModePrefer)
	}

	sslModeOk := !sslModes.MatchString(p.SSLMode)

	return validate(p, []validation{
		{kw: "port", errCond: p.Port < 0, val: p.Port},
		{kw: "user", errCond: p.User == "", val: p.User},
		{kw: "password", errCond: p.Password == "", val: p.Password},
		{kw: "dbname", errCond: p.DbName == "", val: p.DbName},
		{kw: "host", errCond: p.Host == "", val: p.Host},
		{kw: "sslmode", errCond: sslModeOk, val: p.SSLMode},
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
// valid.
func (p *PostgresSpec) UpdateWith(pSpec *PostgresSpec) error {
	if pSpec == nil {
		return errors.New("cannot update a PostgresSpec with a nil Specifier")
	}
	if err := pSpec.validate(); err != nil {
		return err
	}
	*p = *pSpec
	return nil
}
