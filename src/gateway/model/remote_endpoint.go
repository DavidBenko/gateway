package model

import (
	"errors"
	"fmt"

	"gateway/code"
	"gateway/db"
	re "gateway/model/remote_endpoint"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

const (
	// RemoteEndpointTypeHTTP denotes that a remote endpoint is an HTTP endpoint
	RemoteEndpointTypeHTTP = "http"
	// RemoteEndpointTypeSQLServer denotes that a remote endpoint is a MS SQL Server database
	RemoteEndpointTypeSQLServer = "sqlserver"
	RemoteEndpointTypeMySQL     = "mysql"
	RemoteEndpointTypePostgres  = "postgres"
	RemoteEndpointTypeMongo     = "mongodb"
	// RemoteEndpointTypeSoap denotes that a remote endpoint is a SOAP service
	RemoteEndpointTypeSoap = "soap"
)

const (
	// RemoteEndpointStatusPending is one of the possible statuses for the Status field on
	// the RemoteEndpoint struct.  Pending indicates that no processing has yet been attempted
	// on a RemoteEndpoint
	RemoteEndpointStatusPending = "Pending"
	// RemoteEndpointStatusProcessing is one of the possible statuses for the Status field on
	// the RemoteEndpoint struct.  Processing indicates that the WSDL file is actively being processed,
	// and is pending final outcome. (originally for use in soap remote endpoints)
	RemoteEndpointStatusProcessing = "Processing"
	// RemoteEndpointStatusFailed is one of the possible statuses for the Status field on the
	// RemoteEndpoint struct.  Failed indicates that there was a failure encountered processing the
	// WSDL.  The user ought to correct the problem with their WSDL and attempt to process it again.
	// (originally for use in soap remote endpoints)
	RemoteEndpointStatusFailed = "Failed"
	// RemoteEndpointStatusSuccess is one of the possible statuses for the Status field on the
	// RemoteEndpoint struct.  Success indicates taht processing on the WSDL file has been completed
	// successfully, and the SOAP service is ready to be invoked. (originally for use in soap remote endpoints)
	RemoteEndpointStatusSuccess = "Success"
)

// RemoteEndpoint is an endpoint that a proxy endpoint delegates to.
type RemoteEndpoint struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id"`

	ID            int64            `json:"id,omitempty"`
	Name          string           `json:"name"`
	Codename      string           `json:"codename"`
	Description   string           `json:"description"`
	Type          string           `json:"type"`
	Status        apsql.NullString `json:"status,omitempty"`
	StatusMessage apsql.NullString `json:"status_message,omitempty" db:"status_message"`

	Data            types.JsonText                   `json:"data" db:"data"`
	EnvironmentData []*RemoteEndpointEnvironmentData `json:"environment_data"`

	// Proxy Data Cache
	SelectedEnvironmentData *types.JsonText `json:"-" db:"selected_env_data"`

	// Soap specific attributes
	Soap *SoapRemoteEndpoint `json:"-"`
}

// RemoteEndpointEnvironmentData contains per-environment endpoint data
type RemoteEndpointEnvironmentData struct {
	RemoteEndpointID int64          `json:"-" db:"remote_endpoint_id"`
	EnvironmentID    int64          `json:"environment_id,omitempty" db:"environment_id"`
	Data             types.JsonText `json:"data"`

	ExportEnvironmentIndex int `json:"environment_index,omitempty"`
}

// Validate validates the model.
func (e *RemoteEndpoint) Validate() Errors {
	errors := make(Errors)
	if e.Name == "" {
		errors.add("name", "must not be blank")
	}
	if e.Codename == "" {
		errors.add("codename", "must not be blank")
	}
	if code.IsReserved(e.Codename) {
		errors.add("codename", "is a reserved word and may not be used")
	}
	if !code.IsValidVariableIdentifier(e.Codename) {
		errors.add("codename", "must start with A-Z a-z _ and may only contain A-Z a-z 0-9 _")
	}
	switch e.Type {
	case RemoteEndpointTypeHTTP, RemoteEndpointTypeSoap:
	case RemoteEndpointTypeMySQL, RemoteEndpointTypeSQLServer,
		RemoteEndpointTypePostgres, RemoteEndpointTypeMongo:
		_, err := e.DBConfig()
		if err != nil {
			errors.add("base", fmt.Sprintf("error in database config: %s", err))
		}
	default:
		errors.add("base", fmt.Sprintf("unknown endpoint type %q", e.Type))
	}
	if e.Status.Valid {
		val, _ := e.Status.Value()
		switch val {
		case RemoteEndpointStatusFailed, RemoteEndpointStatusProcessing,
			RemoteEndpointStatusSuccess, RemoteEndpointStatusPending:
		default:
			errors.add("status", fmt.Sprintf("Invalid value for status: %v", val))
		}
	}

	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *RemoteEndpoint) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if apsql.IsUniqueConstraint(err, "remote_endpoints", "api_id", "name") {
		errors.add("name", "is already taken")
	}
	if apsql.IsNotNullConstraint(err, "remote_endpoint_environment_data", "environment_id") {
		errors.add("environment_data", "must include a valid environment in this API")
	}
	return errors
}

// AllRemoteEndpointsForAPIIDAndAccountID returns all remoteEndpoints on the Account's API in default order.
func AllRemoteEndpointsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*RemoteEndpoint, error) {
	return _remoteEndpoints(db, 0, apiID, accountID)
}

// AllRemoteEndpointsForIDsInEnvironment returns all remoteEndpoints with id specified,
// populated with environment data
func AllRemoteEndpointsForIDsInEnvironment(db *apsql.DB, ids []int64, environmentID int64) ([]*RemoteEndpoint, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	idQuery := apsql.NQs(len(ids))
	query := `SELECT
		remote_endpoints.api_id as api_id,
		remote_endpoints.id as id,
		remote_endpoints.name as name,
		remote_endpoints.codename as codename,
		remote_endpoints.type as type,
		remote_endpoints.data as data,
		remote_endpoint_environment_data.data as selected_env_data,
		remote_endpoints.status as status,
		remote_endpoints.status_message as status_message
	FROM remote_endpoints
	LEFT JOIN remote_endpoint_environment_data
		ON remote_endpoints.id = remote_endpoint_environment_data.remote_endpoint_id
	 AND remote_endpoint_environment_data.environment_id = ?
	WHERE remote_endpoints.id IN (` + idQuery + `);`
	args := []interface{}{environmentID}
	for _, id := range ids {
		args = append(args, id)
	}
	remoteEndpoints := []*RemoteEndpoint{}
	err := db.Select(&remoteEndpoints, query, args...)
	return remoteEndpoints, err
}

// FindRemoteEndpointForAPIIDAndAccountID returns the remoteEndpoint with the id, api id, and account_id specified.
func FindRemoteEndpointForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*RemoteEndpoint, error) {
	endpoints, err := _remoteEndpoints(db, id, apiID, accountID)
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("No endpoint with id %d found", id)
	}
	return endpoints[0], nil
}

func _remoteEndpoints(db *apsql.DB, id, apiID, accountID int64) ([]*RemoteEndpoint, error) {
	query := `SELECT
		remote_endpoints.api_id as api_id,
	  remote_endpoints.id as id,
	  remote_endpoints.name as name,
		remote_endpoints.codename as codename,
	  remote_endpoints.description as description,
	  remote_endpoints.type as type,
		remote_endpoints.data as data,
		remote_endpoints.status as status,
		remote_endpoints.status_message as status_message
	FROM remote_endpoints, apis
	WHERE `
	args := []interface{}{}
	if id != 0 {
		query = query + "remote_endpoints.id = ? AND "
		args = append(args, id)
	}
	query = query +
		`   remote_endpoints.api_id = ?
	  AND remote_endpoints.api_id = apis.id
	  AND apis.account_id = ?
  ORDER BY
	  remote_endpoints.name ASC,
		remote_endpoints.id ASC;`
	args = append(args, apiID, accountID)
	remoteEndpoints := []*RemoteEndpoint{}
	err := db.Select(&remoteEndpoints, query, args...)
	if err != nil {
		return nil, err
	}
	if len(remoteEndpoints) == 0 {
		return remoteEndpoints, nil
	}

	var endpointIDs []interface{}
	for _, endpoint := range remoteEndpoints {
		endpointIDs = append(endpointIDs, endpoint.ID)
	}
	idQuery := apsql.NQs(len(remoteEndpoints))
	environmentData := []*RemoteEndpointEnvironmentData{}
	err = db.Select(&environmentData,
		`SELECT
			remote_endpoint_environment_data.remote_endpoint_id as remote_endpoint_id,
			remote_endpoint_environment_data.environment_id as environment_id,
			remote_endpoint_environment_data.data as data
		FROM remote_endpoint_environment_data, remote_endpoints, environments
		WHERE remote_endpoint_environment_data.remote_endpoint_id IN (`+idQuery+`)
			AND remote_endpoint_environment_data.environment_id = environments.id
			AND remote_endpoint_environment_data.remote_endpoint_id = remote_endpoints.id
		ORDER BY
			remote_endpoints.name ASC,
			remote_endpoints.id ASC,
			environments.name ASC;`,
		endpointIDs...)
	if err != nil {
		return nil, err
	}
	var endpointIndex int64
	for _, envData := range environmentData {
		for remoteEndpoints[endpointIndex].ID != envData.RemoteEndpointID {
			endpointIndex++
		}
		endpoint := remoteEndpoints[endpointIndex]
		endpoint.EnvironmentData = append(endpoint.EnvironmentData, envData)
	}
	return remoteEndpoints, err
}

// CanDeleteRemoteEndpoint checks whether deleting would violate any constraints
func CanDeleteRemoteEndpoint(tx *apsql.Tx, id int64) error {
	var count int64
	if err := tx.Get(&count,
		`SELECT COUNT(id) FROM proxy_endpoint_calls
		 WHERE remote_endpoint_id = ?;`, id); err != nil {
		return errors.New("Could not check if endpoint could be deleted.")
	}

	if count > 0 {
		return errors.New("There are proxy endpoint calls that reference this endpoint.")
	}

	return nil
}

func beforeDelete(remoteEndpoint *RemoteEndpoint) error {
	if remoteEndpoint.Status.String == RemoteEndpointStatusPending {
		return fmt.Errorf("Unable to delete remote endpoint -- status is currently %s", RemoteEndpointStatusPending)
	}
	return nil
}

// DeleteRemoteEndpointForAPIIDAndAccountID deletes the remoteEndpoint with the id, api_id and account_id specified.
func DeleteRemoteEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	var endpoints []*RemoteEndpoint
	err := tx.Select(&endpoints,
		`SELECT remote_endpoints.type as type,
			remote_endpoints.data as data,
			remote_endpoints.status as status
		FROM remote_endpoints
		WHERE remote_endpoints.id = ?
		AND remote_endpoints.api_id IN
			(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		id, apiID, accountID)
	if err != nil {
		return err
	}

	if len(endpoints) != 1 {
		return fmt.Errorf("found multiple remote endpoints for endpoint id %d", id)
	}
	endpoint := endpoints[0]
	var msg interface{}
	if endpoint.Type != RemoteEndpointTypeHTTP && endpoint.Type != RemoteEndpointTypeSoap {
		conf, err := endpoint.DBConfig()
		switch err {
		case nil:
			msg = conf
		default:
			msg = err
		}
	}

	if err = beforeDelete(endpoint); err != nil {
		return err
	}

	err = tx.DeleteOne(
		`DELETE FROM remote_endpoints
		WHERE remote_endpoints.id = ?
			AND remote_endpoints.api_id IN
				(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		id, apiID, accountID)
	if err != nil {
		return err
	}

	return tx.Notify("remote_endpoints", apiID, apsql.Delete, msg)
}

// DBConfig gets a DB Specifier for database endpoints, or nil for non database
// endpoints, and returns any validation errors generated for the Config.
func (e *RemoteEndpoint) DBConfig() (db.Specifier, error) {
	switch e.Type {
	case RemoteEndpointTypeSQLServer:
		return re.SQLServerConfig(e.Data)
	case RemoteEndpointTypePostgres:
		return re.PostgresConfig(e.Data)
	case RemoteEndpointTypeMySQL:
		return re.MySQLConfig(e.Data)
	case RemoteEndpointTypeMongo:
		return re.MongoConfig(e.Data)
	default:
		return nil, fmt.Errorf("unknown database endpoint type %q", e.Type)
	}
}

// Insert inserts the remoteEndpoint into the database as a new row.
func (e *RemoteEndpoint) Insert(tx *apsql.Tx) error {
	encodedData, err := marshaledForStorage(e.Data)
	if err != nil {
		return err
	}
	e.ID, err = tx.InsertOne(
		`INSERT INTO remote_endpoints (api_id, name, codename, description, type, status, status_message, data)
		VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?,?,?,?,?,?,?)`,
		e.APIID, e.AccountID, e.Name, e.Codename, e.Description, e.Type, e.Status, e.StatusMessage, encodedData)
	if err != nil {
		return err
	}
	for _, envData := range e.EnvironmentData {
		encodedData, err := marshaledForStorage(envData.Data)
		if err != nil {
			return err
		}
		err = _insertRemoteEndpointEnvironmentData(tx, e.ID, envData.EnvironmentID,
			e.APIID, encodedData)
		if err != nil {
			return err
		}
	}
	return nil
}

// Update updates the remoteEndpoint in the database.
func (e *RemoteEndpoint) Update(tx *apsql.Tx) error {
	// Get any database config for Flushing if needed.
	var msg interface{}
	if e.Type != RemoteEndpointTypeHTTP {
		conf, err := e.DBConfig()
		switch err {
		case nil:
			msg = conf
		default:
			msg = err
		}
	}

	encodedData, err := marshaledForStorage(e.Data)
	if err != nil {
		return err
	}
	err = tx.UpdateOne(
		`UPDATE remote_endpoints
		SET name = ?, codename = ?, description = ?, status = ?, status_message = ?, data = ?
		WHERE remote_endpoints.id = ?
			AND remote_endpoints.api_id IN
				(SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		e.Name, e.Codename, e.Description, e.Status, e.StatusMessage, encodedData, e.ID, e.APIID, e.AccountID)
	if err != nil {
		return err
	}

	var existingEnvIDs []int64
	err = tx.Select(&existingEnvIDs,
		`SELECT environment_id
		FROM remote_endpoint_environment_data
		WHERE remote_endpoint_id = ?
		ORDER BY environment_id ASC;`,
		e.ID)
	if err != nil {
		return err
	}

	for _, envData := range e.EnvironmentData {
		encodedData, err := marshaledForStorage(envData.Data)
		if err != nil {
			return err
		}

		var found bool
		existingEnvIDs, found = popID(envData.EnvironmentID, existingEnvIDs)
		if found {
			_, err = tx.Exec(
				`UPDATE remote_endpoint_environment_data
				  SET data = ?
				WHERE remote_endpoint_id = ?
				  AND environment_id = ?;`,
				encodedData, e.ID, envData.EnvironmentID)
			if err != nil {
				return err
			}
		} else {
			err = _insertRemoteEndpointEnvironmentData(tx, e.ID, envData.EnvironmentID,
				e.APIID, encodedData)
			if err != nil {
				return err
			}
		}
	}

	if len(existingEnvIDs) == 0 {
		return nil
	}

	args := []interface{}{e.ID}
	for _, envID := range existingEnvIDs {
		args = append(args, envID)
	}
	idQuery := apsql.NQs(len(existingEnvIDs))
	_, err = tx.Exec(
		`DELETE FROM remote_endpoint_environment_data
		WHERE remote_endpoint_id = ? AND environment_id IN (`+idQuery+`);`,
		args...)

	if err != nil {
		return err
	}
	return tx.Notify("remote_endpoints", e.APIID, apsql.Update, msg)
}

func _insertRemoteEndpointEnvironmentData(tx *apsql.Tx, rID, eID, apiID int64,
	data string) error {
	_, err := tx.Exec(
		`INSERT INTO remote_endpoint_environment_data
			(remote_endpoint_id, environment_id, data)
			VALUES (?, (SELECT id FROM environments WHERE id = ? AND api_id = ?), ?);`,
		rID, eID, apiID, data)
	return err
}
