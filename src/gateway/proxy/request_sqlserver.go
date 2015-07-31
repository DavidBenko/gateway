package proxy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	sqls "gateway/db/sqlserver"
)

// SQLServerRequest encapsulates a request made to a SQLServer endpoint
type SQLServerRequest struct {
	Query       string        `json:"queryStatement"`
	Execute     string        `json:"executeStatement"`
	Parameters  []interface{} `json:"parameters"`
	Config      sqls.Conn     `json:"config"`
	Tx          bool          `json:"transactions"`
	MaxOpenConn int           `json:"maxOpenConn,omitempty"`
	MaxIdleConn int           `json:"maxIdleConn,omitempty"`
	conn        *sqls.DB
}

// SQLServerResponse encapsulates a response from a SQLServerRequest.
// It also contains its transaction, in case it needs to be rolled back.
type SQLServerResponse struct {
	Data         []map[string]interface{} `json:"data"`
	InsertID     int64                    `json:"insertId,omitempty"`
	RowsAffected int64                    `json:"rowsAffected,omitempty"`
}

// Perform executes the SQLServer request and returns its response
func (r *SQLServerRequest) Perform() Response {
	if r.Query != "" {
		return r.performQuery()
	} else if r.Execute != "" {
		return r.performExecute()
	}

	return NewErrorResponse(errors.New("no SQL query or execute specified"))
}

func (r *SQLServerRequest) performQuery() Response {
	// TODO(binary132): Much of this could be in package gateway/db.
	// TODO(binary132): two-phase commits?
	if r.Tx {
		return r.transactQuery()
	}

	log.Printf("Params are %v", r.Parameters)

	if r.conn == nil {
		return NewSQLServerErrorResponse(errors.New("nil database connection"), "nil database connection")
	}

	rows, err := r.conn.Queryx(r.Query, r.Parameters...)

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		return NewSQLServerErrorResponse(err, "failed to execute MSSQL query")
	}

	var dataRows []map[string]interface{}

	for rowNum := 0; rows.Next(); rowNum++ {
		newMap := make(map[string]interface{})
		err := rows.MapScan(newMap)
		if err != nil {
			return NewSQLServerErrorResponse(err, "failed to extract results of MSSQL query")
		}
		dataRows = append(dataRows, newMap)
	}

	err = rows.Err()
	if err != nil {
		return NewSQLServerErrorResponse(err, "failed to iterate over rows in MSSQL query response")
	}

	return &SQLServerResponse{Data: dataRows}
}

func (r *SQLServerRequest) transactQuery() Response {
	log.Printf("Params are %v", r.Parameters)

	// Begin transaction
	tx, err := r.conn.Beginx()
	if err != nil {
		return NewSQLServerErrorResponse(err, "failed to get MSSQL transaction handle")
	}

	rows, err := tx.Queryx(r.Query, r.Parameters...)

	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLServerErrorResponse(newErr, "failed to roll back MSSQL query after error: "+err.Error())
		}

		return NewSQLServerErrorResponse(err, "failed to execute MSSQL query")
	}

	var dataRows []map[string]interface{}

	for rowNum := 0; rows.Next(); rowNum++ {
		newMap := make(map[string]interface{})
		err := rows.MapScan(newMap)
		if err != nil {
			newErr := tx.Rollback()
			if newErr != nil {
				return NewSQLServerErrorResponse(newErr, "failed to roll back MSSQL query after error: "+err.Error())
			}
			return NewSQLServerErrorResponse(err, "failed to extract results of MSSQL query")
		}
		dataRows = append(dataRows, newMap)
	}

	err = rows.Err()
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLServerErrorResponse(newErr, "failed to roll back MSSQL query after error: "+err.Error())
		}
		return NewSQLServerErrorResponse(err, "failed to iterate over rows in MSSQL query response")
	}

	err = tx.Commit()
	if err != nil {
		return NewSQLServerErrorResponse(err, "failed to commit MSSQL query")
	}
	return &SQLServerResponse{Data: dataRows}
}

func (r *SQLServerRequest) performExecute() Response {
	if r.Tx {
		return r.transactExecute()
	}
	result, err := r.conn.Exec(r.Execute, r.Parameters...)
	if err != nil {
		return NewSQLServerErrorResponse(err, "Failed to execute MSSQL update")
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		return NewSQLServerErrorResponse(err, "Failed checking for MSSQL insert ID after update")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewSQLServerErrorResponse(err, "Failed checking for MSSQL number of rows affected")
	}

	return &SQLServerResponse{
		RowsAffected: rowsAffected,
		InsertID:     insertID,
	}
}

func (r *SQLServerRequest) transactExecute() Response {
	tx, err := r.conn.Beginx()
	if err != nil {
		return NewSQLServerErrorResponse(err, "Failed to get exec transaction handle")
	}

	result, err := tx.Exec(r.Execute, r.Parameters...)
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLServerErrorResponse(newErr, "failed to roll back MSSQL exec after error: "+err.Error())
		}
		return NewSQLServerErrorResponse(err, "Failed to execute MSSQL update")
	}

	insertID, err := result.LastInsertId()
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLServerErrorResponse(newErr, "failed to roll back MSSQL exec after error: "+err.Error())
		}
		return NewSQLServerErrorResponse(err, "Failed checking for MSSQL insert ID after update")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return NewSQLServerErrorResponse(newErr, "failed to roll back MSSQL exec after error: "+err.Error())
		}
		return NewSQLServerErrorResponse(err, "Failed checking for MSSQL number of rows affected")
	}

	err = tx.Commit()
	if err != nil {
		return NewSQLServerErrorResponse(err, "failed to commit MS SQL exec")
	}

	return &SQLServerResponse{
		RowsAffected: rowsAffected,
		InsertID:     insertID,
	}
}

// Log returns the SQLServer request basics, It returns the SQL statement when in server mode.
// When in dev mode the query parameters are also returned.
func (request *SQLServerRequest) Log(devMode bool) string {
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
		buffer.WriteString(fmt.Sprintf("\nConnection: %s", request.Config))
		buffer.WriteString(fmt.Sprintf("\nTransactional: %t", request.Tx))
	}
	return buffer.String()
}

// JSON converts this response to JSON format.
func (r *SQLServerResponse) JSON() ([]byte, error) {
	return json.Marshal(&r)
}

// Log returns the number of records affected by the statement
func (r *SQLServerResponse) Log() string {
	if r.Data != nil {
		return fmt.Sprintf("Records found: %d", len(r.Data))
	}

	return fmt.Sprintf("Records affected: %d", r.RowsAffected)
}
