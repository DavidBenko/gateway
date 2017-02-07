package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/db/pools"
	"gateway/db/redis"
	"gateway/logreport"
	"gateway/model"
	"strings"

	redigo "github.com/garyburd/redigo/redis"
)

type RedisRequest struct {
	Config    *redis.Spec `json:"config"`
	Arguments []string    `json:"arguments"`
	conn      redigo.Conn
}

type RedisResponse struct {
	Data  []interface{} `json:"data"`
	Error string        `json:"error,omitempty"`
}

func splitStatement(statement string) []string {
	return strings.Split(statement, " ")
}

func toEmptyInterfaceSlice(s []string) []interface{} {
	a := make([]interface{}, len(s))

	for i := range s {
		a[i] = s[i]
	}

	return a
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

	if endpoint.SelectedEnvironmentData != nil {
		endpointData := &RedisRequest{}
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, endpointData); err != nil {
			return nil, err
		}
		request.updateWith(endpointData)

	}

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

	if len(r.Arguments) == 0 {
		response.Error = "missing command parameter"
		return response
	}

	// The command will be the first element in the string slice
	command := r.Arguments[0]

	// Pass the command and all parameters after the first (the first is the command)

	d, err := r.conn.Do(command, toEmptyInterfaceSlice(r.Arguments)[1:]...)

	if err != nil {
		logreport.Printf(err.Error())
		response.Error = err.Error()
		return response
	}

	switch t := d.(type) {
	case int64, string:
		data := make([]interface{}, 1)
		data[0] = d
		response.Data = data
	case []byte:
		data := make([]interface{}, 1)
		data[0] = string(d.([]byte)[:])
		response.Data = data
	case []interface{}:
		values, err := redigo.Values(d, nil)

		if err != nil {
			response.Error = err.Error()
			break
		}

		data := make([]interface{}, len(values))

		for i, v := range values {
			switch tt := v.(type) {
			case []byte:
				data[i] = string(v.([]byte)[:])
			default:
				_ = tt
				data[i] = v
			}
		}

		response.Data = data
	default:
		_ = t
	}

	return response
}

func (r *RedisRequest) Log(devMode bool) string {
	return strings.Join(r.Arguments, " ")
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
	return ""
}
