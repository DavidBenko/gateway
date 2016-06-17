package remote_endpoint

import (
	"encoding/json"
	"fmt"
	"gateway/db"
	"gateway/db/redis"

	"github.com/jmoiron/sqlx/types"
)

type Redis struct {
	Config    *redis.Spec `json:"config"`
	MaxActive int         `json:"maxActive"`
	MaxIdle   int         `json:"maxIdle"`
}

func RedisConfig(data types.JsonText) (db.Specifier, error) {
	var conf Redis
	err := json.Unmarshal(data, &conf)
	if err != nil {
		return nil, fmt.Errorf("bad JSON for Redis config: %s", err.Error())
	}

	spec, err := redis.Config(
		redis.Connection(conf.Config),
		redis.MaxActive(conf.MaxActive),
		redis.MaxIdle(conf.MaxIdle),
	)

	if err != nil {
		return nil, err
	}

	return spec, nil
}
