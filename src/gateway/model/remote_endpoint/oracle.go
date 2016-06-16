package remote_endpoint

import (
	"encoding/json"
	"fmt"

	"gateway/db"
	sql "gateway/db/sql"

	"github.com/jmoiron/sqlx/types"
)

type Oracle struct {
	Config      *sql.OracleSpec `json:"config"`
	Tx          bool            `json:"transactions"`
	MaxOpenConn int             `json:"maxOpenConn,omitempty"`
	MaxIdleConn int             `json:"maxIdleConn,omitempty"`
}

// OracleConfig gets a "gateway/db/oracle" Config and returns any errors.
func OracleConfig(data types.JsonText) (db.Specifier, error) {
	var conf Oracle
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for Oracle config: %s", err.Error())
	}

	spec, err := sql.Config(
		sql.Connection(conf.Config),
		sql.MaxOpenIdle(conf.MaxOpenConn, conf.MaxIdleConn),
	)
	if err != nil {
		return nil, err
	}
	return spec, nil
}
