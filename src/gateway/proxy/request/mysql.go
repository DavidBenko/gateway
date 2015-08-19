package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"gateway/db/pools"
	sql "gateway/db/sql"
	"gateway/model"
)

// MySQLRequest encapsulates a request made to a MySQL endpoint.
type MySQLRequest struct {
	sqlRequest
	Config *sql.MySQLSpec `json:"config"`
}

// Perform executes the sqlRequest and returns its response
func (r *MySQLRequest) Perform() Response {
	isQuery, isExec := r.Query != "", r.Execute != ""

	switch {
	case isQuery && !r.Tx:
		return r.performQuery()
	case isQuery && r.Tx:
		return r.transactQuery()
	case isExec:
		return r.sqlRequest.Perform()
	default:
		return NewErrorResponse(errors.New("no SQL query or execute specified"))
	}
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

// performQuery is like sql.performQuery, but uses prepared statements.
func (r *MySQLRequest) performQuery() Response {
	log.Printf("Params are %v", r.Parameters)

	if r.conn == nil {
		return NewSQLErrorResponse(errors.New("nil database connection"), "nil database connection")
	}

	q, err := r.conn.Preparex(r.Query)
	if q != nil {
		defer q.Close()
	}

	if err != nil {
		return NewSQLErrorResponse(err, "failed to prepare SQL query")
	}

	rows, err := q.Queryx(r.Parameters...)
	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		return NewSQLErrorResponse(err, "failed to execute SQL query")
	}

	var dataRows []map[string]interface{}

	for rowNum := 0; rows.Next(); rowNum++ {
		newMap := make(map[string]interface{})

		err = rows.MapScan(newMap)
		if err != nil {
			return NewSQLErrorResponse(err, "failed to extract results of SQL query")
		}

		dataRows = append(dataRows, newMap)
	}

	err = rows.Err()
	if err != nil {
		return NewSQLErrorResponse(err, "failed to iterate over rows in SQL query response")
	}

	return &sqlResponse{Data: dataRows}
}

// transactQuery is like sql.transactQuery, but uses prepared statements.
func (r *MySQLRequest) transactQuery() Response {
	log.Printf("Params are %v", r.Parameters)

	if r.conn == nil {
		return NewSQLErrorResponse(errors.New("nil database connection"), "nil database connection")
	}

	// Begin transaction
	tx, err := r.conn.Beginx()
	if err != nil {
		return NewSQLErrorResponse(err, "failed to get SQL transaction handle")
	}

	q, err := tx.Preparex(r.Query) //, r.Parameters...)

	if q != nil {
		defer q.Close()
	}

	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLErrorResponse(newErr, "failed to roll back SQL query after error: "+err.Error())
		}

		return NewSQLErrorResponse(err, "failed to prepare SQL query")
	}

	rows, err := q.Queryx(r.Parameters...)

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLErrorResponse(newErr, "failed to roll back SQL query after error: "+err.Error())
		}

		return NewSQLErrorResponse(err, "failed to execute SQL query")
	}

	var dataRows []map[string]interface{}

	for rowNum := 0; rows.Next(); rowNum++ {
		newMap := make(map[string]interface{})
		err := rows.MapScan(newMap)
		if err != nil {
			newErr := tx.Rollback()
			if newErr != nil {
				return NewSQLErrorResponse(newErr, "failed to roll back SQL query after error: "+err.Error())
			}
			return NewSQLErrorResponse(err, "failed to extract results of SQL query")
		}
		dataRows = append(dataRows, newMap)
	}

	err = rows.Err()
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLErrorResponse(newErr, "failed to roll back SQL query after error: "+err.Error())
		}
		return NewSQLErrorResponse(err, "failed to iterate over rows in SQL query response")
	}

	err = tx.Commit()
	if err != nil {
		return NewSQLErrorResponse(err, "failed to commit SQL query")
	}
	return &sqlResponse{Data: dataRows}
}
