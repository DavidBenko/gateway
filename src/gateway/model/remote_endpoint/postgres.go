package remote_endpoint

import (
	"encoding/json"
	"fmt"

	"gateway/db"
	sql "gateway/db/sql"

	"github.com/jmoiron/sqlx/types"
)

type Postgres struct {
	Config      *sql.PostgresSpec `json:"config"`
	Tx          bool              `json:"transactions"`
	MaxOpenConn int               `json:"maxOpenConn,omitempty"`
	MaxIdleConn int               `json:"maxIdleConn,omitempty"`
}

// PostgresConfig gets a "gateway/db/postgres" Config and returns any errors.
func PostgresConfig(data types.JsonText) (db.Specifier, error) {
	var conf Postgres
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for Postgres config: %s", err.Error())
	}

	// default sslmode to 'prefer' if not provided
	if conf.Config.SSLMode == "" {
		conf.Config.SSLMode = "prefer"
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
