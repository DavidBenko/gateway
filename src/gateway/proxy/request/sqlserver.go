package request

import (
	"encoding/json"
	"errors"
	"fmt"

	"gateway/db/pools"
	sqls "gateway/db/sqlserver"
	"gateway/model"
)

// SQLServerRequest encapsulates a request made to a SQLServer endpoint.
type SQLServerRequest struct {
	sqlRequest
	Config sqls.Conn `json:"config"`
}

func (r *SQLServerRequest) Log(devMode bool) string {
	s := r.sqlRequest.Log(devMode)
	if devMode {
		s += fmt.Sprintf("\nConnection: %+v", r.Config)
	}
	return s
}

func NewSQLServerRequest(pools *pools.Pools, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &SQLServerRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := &SQLServerRequest{}
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

	conn, err := pools.Connect(sqls.Config(
		sqls.Connection(request.Config),
		sqls.MaxOpenIdle(request.MaxOpenConn, request.MaxIdleConn),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if sqlConn, ok := conn.(*sqls.DB); ok {
		if sqlConn.DB == nil {
			return nil, fmt.Errorf("got nil database connection")
		}
		request.conn = sqlConn.DB
		return request, nil
	}

	return nil, fmt.Errorf("need MSSQL connection, got %T", conn)
}

// TODO - refactor to DRY this code up across different data sources
func (r *SQLServerRequest) updateWith(endpointData *SQLServerRequest) {
	if endpointData.Config != nil {
		if r.Config == nil {
			r.Config = sqls.Conn{}
		}
		for key, value := range endpointData.Config {
			r.Config[key] = value
		}
	}

	if endpointData.Execute != "" {
		r.Execute = endpointData.Execute
	}

	if endpointData.Query != "" {
		r.Query = endpointData.Query
	}

	if endpointData.Parameters != nil {
		r.Parameters = endpointData.Parameters
	}

	if endpointData.Tx {
		r.Tx = endpointData.Tx
	}

	if r.MaxOpenConn != endpointData.MaxOpenConn {
		r.MaxOpenConn = endpointData.MaxOpenConn
	}

	if r.MaxIdleConn != endpointData.MaxIdleConn {
		r.MaxIdleConn = endpointData.MaxIdleConn
	}
}
