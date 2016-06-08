package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/db/pools"
	"gateway/db/redis"
	"gateway/model"

	redigo "github.com/garyburd/redigo/redis"
)

type RedisRequest struct {
	Config    *redis.Spec `json:"config"`
	MaxActive int
	MaxIdle   int
	conn      redigo.Conn
}

type RedisResponse struct {
}

func NewRedisRequest(pools *pools.Pools, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &RedisRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := &RedisRequest{}
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal endpoint configuration: %v", err)
	}

	request.updateWith(endpointData)

	if pools == nil {
		return nil, errors.New("database pools not set up")
	}

	c, err := pools.Connect(redis.Config(
		redis.Connection(request.Config),
		redis.MaxActive(request.MaxActive),
		redis.MaxIdle(request.MaxIdle),
	))

	if err != nil {
		return nil, fmt.Errorf("failed to get redis connection pool: %v", err)
	}

	if redisPool, ok := c.(*redis.DB); ok {
		// Grab an available connection from the redis connection pool
		request.conn = redisPool.Pool.Get()
		return request, nil
	}

	return nil, fmt.Errorf("need Redis connection, got %T", c)
}

func (r *RedisRequest) updateWith(endpointData *RedisRequest) {
	if endpointData.Config != nil {
		if r.Config == nil {
			r.Config = &redis.Spec{}
		}
		r.Config.UpdateWith(endpointData.Config)
	}

	if r.MaxActive != endpointData.MaxActive {
		r.MaxActive = endpointData.MaxActive
	}
	if r.MaxIdle != endpointData.MaxIdle {
		r.MaxIdle = endpointData.MaxIdle
	}
}
