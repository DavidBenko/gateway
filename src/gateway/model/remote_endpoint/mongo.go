package remote_endpoint

import (
	"encoding/json"
	"fmt"

	"gateway/db"
	"gateway/db/mongo"

	"github.com/jmoiron/sqlx/types"
)

type Mongo struct {
	Config mongo.Conn `json:"config"`
	Limit  int        `json:"limit"`
}

func MongoConfig(data types.JsonText) (db.Specifier, error) {
	var conf Mongo
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for Mongo config: %s", err.Error())
	}

	spec, err := mongo.Config(
		mongo.Connection(conf.Config),
		mongo.PoolLimit(conf.Limit),
	)
	if err != nil {
		return nil, err
	}
	return spec, nil
}
