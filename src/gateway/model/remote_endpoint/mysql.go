package remote_endpoint

import (
	"encoding/json"
	"fmt"

	"gateway/db"
	sql "gateway/db/sql"

	"github.com/jmoiron/sqlx/types"
)

type MySQL struct {
	Config      *sql.MySQLSpec `json:"config"`
	Tx          bool           `json:"transactions"`
	MaxOpenConn int            `json:"maxOpenConn,omitempty"`
	MaxIdleConn int            `json:"maxIdleConn,omitempty"`
}

// MySQLConfig gets a "gateway/db/sql" MySQLSpec and returns any errors.
func MySQLConfig(data types.JsonText) (db.Specifier, error) {
	var conf MySQL
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for MySQL config: %s", err.Error())
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
