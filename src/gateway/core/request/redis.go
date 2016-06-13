package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gateway/db/pools"
	"gateway/db/redis"
	"gateway/model"

	redigo "github.com/garyburd/redigo/redis"
)

type RedisRequest struct {
	Config     *redis.Spec `json:"config"`
	Parameters []interface{}
	conn       redigo.Conn
}

type RedisResponse struct {
	Data  []interface{} `json:"data"`
	Error string        `json:"error,omitempty"`
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

func (r *RedisRequest) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *RedisRequest) Perform() Response {
	response := &RedisResponse{}

	if len(r.Parameters) == 0 {
		response.Error = "missing command parameter"
		return response
	}

	command, ok := r.Parameters[0].(string)
	if !ok {
		response.Error = fmt.Sprintf("invalid command parameter type %T", command)
		return response
	}

	result, err := r.conn.Do(command, r.Parameters[1:]...)

	if err != nil {
		response.Error = err.Error()
		return response
	}

	response.Error = err.Error()

	scanned, err := redigo.Values(result, nil)

	if err != nil {
		response.Error = "failed to extract redis results"
		return response
	}

	response.Data = scanned

	return response
}

func (r *RedisRequest) Log(devMode bool) string {
	if devMode {
		var buffer bytes.Buffer
		// TODO: Log some things
		return buffer.String()
	}
	return ""
}

func (r *RedisRequest) updateWith(endpointData *RedisRequest) {
	if endpointData.Config != nil {
		if r.Config == nil {
			r.Config = &redis.Spec{}
		}
		r.Config.UpdateWith(endpointData.Config)
	}
}

func (r *RedisResponse) JSON() ([]byte, error) {
	return json.Marshal(r)
}

func (r *RedisResponse) Log() string {
	return fmt.Sprintf("Add some response logging")
}
