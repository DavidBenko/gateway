package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"gateway/db/mongo"
	"gateway/db/pools"
	sqls "gateway/db/sqlserver"
	"gateway/model"
)

func (s *Server) prepareRequest(endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	switch endpoint.Type {
	case model.RemoteEndpointTypeHTTP:
		return newHTTPRequest(s.httpClient, endpoint, data)
	case model.RemoteEndpointTypeSQLServer:
		return newSQLServerRequest(s.dbPools, endpoint, data)
	case model.RemoteEndpointTypeMongo:
		return newMongoRequest(s.dbPools, endpoint, data)
	}
	return nil, fmt.Errorf("%s is not a valid call type", endpoint.Type)
}

func newMongoRequest(pools *pools.Pools, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &MongoRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal request json: %v", err)
	}

	endpointData := &MongoRequest{}
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

	conn, err := pools.Connect(mongo.Config(
		mongo.Connection(request.Config),
		mongo.PoolLimit(request.Limit),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if mongoConn, ok := conn.(*mongo.DB); ok {
		request.conn = mongoConn
		return request, nil
	}

	return nil, fmt.Errorf("need Mongo connection, got %T", conn)
}

func (r *MongoRequest) updateWith(endpointData *MongoRequest) {
	if endpointData.Arguments != nil {
		r.Arguments = endpointData.Arguments
	}

	if endpointData.Config != nil {
		if r.Config == nil {
			r.Config = mongo.Conn{}
		}
		for key, value := range endpointData.Config {
			r.Config[key] = value
		}
	}

	if r.Limit != endpointData.Limit {
		r.Limit = endpointData.Limit
	}
}

func newSQLServerRequest(pools *pools.Pools, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
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
		request.conn = sqlConn
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

func newHTTPRequest(client *http.Client, endpoint *model.RemoteEndpoint, data *json.RawMessage) (Request, error) {
	request := &HTTPRequest{}
	if err := json.Unmarshal(*data, request); err != nil {
		return nil, err
	}

	endpointData := &HTTPRequest{}
	if err := json.Unmarshal(endpoint.Data, endpointData); err != nil {
		return nil, err
	}
	request.updateWith(endpointData)

	if endpoint.SelectedEnvironmentData != nil {
		if err := json.Unmarshal(*endpoint.SelectedEnvironmentData, endpointData); err != nil {
			return nil, err
		}
		request.updateWith(endpointData)
	}

	if client == nil {
		return nil, errors.New("no client defined")
	}

	request.client = client

	return request, nil
}

func (r *HTTPRequest) updateWith(endpointData *HTTPRequest) {
	if endpointData.Method != "" {
		r.Method = endpointData.Method
	}
	if endpointData.URL != "" {
		r.URL = endpointData.URL
	}
	if endpointData.Body != "" {
		r.Body = endpointData.Body
	}
	for name, value := range endpointData.Query {
		if r.Query == nil {
			r.Query = make(map[string]string)
		}
		r.Query[name] = value
	}
	for name, value := range endpointData.Headers {
		if r.Headers == nil {
			r.Headers = make(map[string]interface{})
		}
		r.Headers[name] = value
	}
}
