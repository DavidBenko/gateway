package remote_endpoint

import (
	"encoding/json"
	"fmt"

	"gateway/db"
	pq "gateway/db/postgres"

	"github.com/jmoiron/sqlx/types"
)

type Postgres struct {
	Config      pq.Conn `json:"config"`
	Tx          bool    `json:"transactions"`
	MaxOpenConn int     `json:"maxOpenConn,omitempty"`
	MaxIdleConn int     `json:"maxIdleConn,omitempty"`
}

// PostgresConfig gets a "gateway/db/postgres" Config and returns any errors.
func PostgresConfig(data types.JsonText) (db.Specifier, error) {
	var conf Postgres
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for Postgres config: %s", err.Error())
	}

	spec, err := pq.Config(
		pq.Connection(conf.Config),
		pq.MaxOpenIdle(conf.MaxOpenConn, conf.MaxIdleConn),
	)
	if err != nil {
		return nil, err
	}
	return spec, nil
}
