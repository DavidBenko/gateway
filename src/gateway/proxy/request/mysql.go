package request

import (
	"encoding/json"
	"errors"
	"fmt"

	"gateway/db/pools"
	sql "gateway/db/sql"
	"gateway/model"
)

// MySQLRequest encapsulates a request made to a MySQL endpoint.
type MySQLRequest struct {
	sqlRequest
	Config *sql.MySQLSpec `json:"config"`
}

func (r *MySQLRequest) Log(devMode bool) string {
	s := r.sqlRequest.Log(devMode)
	if devMode {
		s += fmt.Sprintf("\nConnection: %+v", r.Config)
	}
	return s
}

func NewMySQLRequest(pools *pools.Pools, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &MySQLRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := &MySQLRequest{}
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal endpoint configuration: %v", err)
	}
	request.updateWith(endpointData)

	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, endpointData); err != nil {
			return nil, err
		}
		request.updateWith(endpointData)
	}

	if pools == nil {
		return nil, errors.New("database pools not set up")
	}

	conn, err := pools.Connect(sql.Config(
		sql.Connection(request.Config),
		sql.MaxOpenIdle(request.MaxOpenConn, request.MaxIdleConn),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if pqConn, ok := conn.(*sql.DB); ok {
		if pqConn.DB == nil {
			return nil, fmt.Errorf("got nil database connection")
		}
		request.conn = pqConn.DB
		return request, nil
	}

	return nil, fmt.Errorf("need MySQL connection, got %T", conn)
}

func (r *MySQLRequest) updateWith(endpointData *MySQLRequest) {
	if endpointData.Config != nil {
		if r.Config == nil {
			r.Config = &sql.MySQLSpec{}
		}
		r.Config.UpdateWith(endpointData.Config)
	}

	r.sqlRequest.updateWith(endpointData.sqlRequest)
}
