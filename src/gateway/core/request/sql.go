package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"gateway/logreport"

	"github.com/jmoiron/sqlx"
)

// sqlRequest is a generic SQL implementation without a config or generator.
// Extend this to implement other SQL database types e.g. MySQL, Postgres
type sqlRequest struct {
	Query       string        `json:"queryStatement"`
	Execute     string        `json:"executeStatement"`
	Parameters  []interface{} `json:"parameters"`
	Tx          bool          `json:"transactions"`
	MaxOpenConn int           `json:"maxOpenConn,omitempty"`
	MaxIdleConn int           `json:"maxIdleConn,omitempty"`
	conn        *sqlx.DB
}

// sqlResponse encapsulates a response from a sqlRequest.
type sqlResponse struct {
	Data         []map[string]interface{} `json:"data"`
	InsertID     int64                    `json:"insertId,omitempty"`
	RowsAffected int64                    `json:"rowsAffected,omitempty"`
}

// Perform executes the sqlRequest and returns its response
func (r *sqlRequest) Perform() Response {
	isQuery, isExec := r.Query != "", r.Execute != ""

	switch {
	case isQuery && !r.Tx:
		return r.performQuery()
	case isQuery && r.Tx:
		return r.transactQuery()
	case isExec && !r.Tx:
		return r.performExecute()
	case isExec && r.Tx:
		return r.transactExecute()
	default:
		return NewErrorResponse(errors.New("no SQL query or execute specified"))
	}
}

func (r *sqlRequest) performQuery() Response {
	logreport.Printf("Params are %v", r.Parameters)

	if r.conn == nil {
		return NewSQLErrorResponse(errors.New("nil database connection"), "nil database connection")
	}

	rows, err := r.conn.Queryx(r.Query, r.Parameters...)

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

func (r *sqlRequest) transactQuery() Response {
	logreport.Printf("Params are %v", r.Parameters)

	// Begin transaction
	tx, err := r.conn.Beginx()
	if err != nil {
		return NewSQLErrorResponse(err, "failed to get SQL transaction handle")
	}

	rows, err := tx.Queryx(r.Query, r.Parameters...)

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

func (r *sqlRequest) performExecute() Response {
	result, err := r.conn.Exec(r.Execute, r.Parameters...)
	if err != nil {
		return NewSQLErrorResponse(err, "Failed to execute SQL update")
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		return NewSQLErrorResponse(err, "Failed checking for SQL insert ID after update")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewSQLErrorResponse(err, "Failed checking for SQL number of rows affected")
	}

	return &sqlResponse{
		RowsAffected: rowsAffected,
		InsertID:     insertID,
	}
}

func (r *sqlRequest) transactExecute() Response {
	tx, err := r.conn.Beginx()
	if err != nil {
		return NewSQLErrorResponse(err, "Failed to get exec transaction handle")
	}

	result, err := tx.Exec(r.Execute, r.Parameters...)
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLErrorResponse(newErr, "failed to roll back SQL exec after error: "+err.Error())
		}
		return NewSQLErrorResponse(err, "Failed to execute SQL update")
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLErrorResponse(newErr, "failed to roll back SQL exec after error: "+err.Error())
		}
		return NewSQLErrorResponse(err, "Failed checking for SQL insert ID after update")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLErrorResponse(newErr, "failed to roll back SQL exec after error: "+err.Error())
		}
		return NewSQLErrorResponse(err, "Failed checking for SQL number of rows affected")
	}

	err = tx.Commit()
	if err != nil {
		return NewSQLErrorResponse(err, "failed to commit MS SQL exec")
	}

	return &sqlResponse{
		RowsAffected: rowsAffected,
		InsertID:     insertID,
	}
}

// Log returns the SQL request basics, It returns the SQL statement when in server mode.
// When in dev mode the query parameters are also returned.
func (request *sqlRequest) Log(devMode bool) string {
	var buffer bytes.Buffer

	if request.Query != "" {
		buffer.WriteString(request.Query)
	} else if request.Execute != "" {
		buffer.WriteString(request.Execute)
	}

	if devMode {
		if len(request.Parameters) > 0 {
			buffer.WriteString(fmt.Sprintf("\nParameters: %v", request.Parameters))
		}
		buffer.WriteString(fmt.Sprintf("\nTransactional: %t", request.Tx))
	}
	return buffer.String()
}

// JSON converts this response to JSON format.
func (r *sqlResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// Log returns the number of records affected by the statement
func (r *sqlResponse) Log() string {
	if r.Data != nil {
		return fmt.Sprintf("Records found: %d", len(r.Data))
	}

	return fmt.Sprintf("Records affected: %d", r.RowsAffected)
}

func (r *sqlRequest) updateWith(endpointData sqlRequest) {
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
