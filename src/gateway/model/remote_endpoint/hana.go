package remote_endpoint

import (
	"encoding/json"
	"fmt"

	"gateway/db"
	sql "gateway/db/sql"

	"github.com/jmoiron/sqlx/types"
)

// Hana type is a HanaSpec config and a boolean to enable transactions
type Hana struct {
	Config *sql.HanaSpec `json:"config"`
	Tx     bool          `json:"transactions"`
}

// HanaConfig gets a "gateway/db/hana" Config and returns any errors.
func HanaConfig(data types.JsonText) (db.Specifier, error) {
	var conf Hana
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for Hana config: %s", err.Error())
	}

	spec, err := sql.Config(
		sql.Connection(conf.Config),
	)
	if err != nil {
		return nil, err
	}
	return spec, nil
}
