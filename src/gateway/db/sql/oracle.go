package sql

import (
	"errors"
	"fmt"

	"gateway/db"

	_ "github.com/mattn/go-oci8"
)

// OracleSpec implements db.Specifier for Oracle connection parameters.
type OracleSpec struct {
	spec
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbname"`
	Host     string `json:"host"`
}

func (o *OracleSpec) validate() error {
	return validate(o, []validation{
		{kw: "port", errCond: o.Port < 0, val: o.Port},
		{kw: "user", errCond: o.User == "", val: o.User},
		{kw: "password", errCond: o.Password == "", val: o.Password},
		{kw: "dbname", errCond: o.DbName == "", val: o.DbName},
		{kw: "host", errCond: o.Host == "", val: o.Host},
	})
}

func (o *OracleSpec) driver() driver {
	return oracle
}

func (o *OracleSpec) ConnectionString() string {
	return fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		o.User,
		o.Password,
		o.Host,
		o.Port,
		o.DbName,
	)
}

func (o *OracleSpec) UniqueServer() string {
	return o.ConnectionString()
}

func (o *OracleSpec) NewDB() (db.DB, error) {
	return newDB(o)
}

// UpdateWith validates `oSpec` and updates `o` with its contents if it is
// valid.
func (o *OracleSpec) UpdateWith(oSpec *OracleSpec) error {
	if oSpec == nil {
		return errors.New("cannot update an OracleSpec with a nil Specifier")
	}
	if err := oSpec.validate(); err != nil {
		return err
	}
	*o = *oSpec
	return nil
}
