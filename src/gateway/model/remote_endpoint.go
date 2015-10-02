package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"gateway/code"
	"gateway/config"
	"gateway/db"
	aperrors "gateway/errors"
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
	UserID    int64 `json:"-"`
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
		remote_endpoints.status_message as status_message,
		soap_remote_endpoints.id as soap_id
	FROM remote_endpoints
	LEFT JOIN remote_endpoint_environment_data
		ON remote_endpoints.id = remote_endpoint_environment_data.remote_endpoint_id
	 AND remote_endpoint_environment_data.environment_id = ?
	LEFT JOIN soap_remote_endpoints
	  ON remote_endpoints.id = soap_remote_endpoints.remote_endpoint_id
	WHERE remote_endpoints.id IN (` + idQuery + `);`

	args := []interface{}{environmentID}
	for _, id := range ids {
		args = append(args, id)
	}

	return mapRemoteEndpoints(db, query, args...)
}

func mapRemoteEndpoints(db *apsql.DB, query string, args ...interface{}) ([]*RemoteEndpoint, error) {
	remoteEndpoints := []*RemoteEndpoint{}
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return nil, aperrors.NewWrapped("[model/remote_endpoint.go] Error fetching all remote endpoints for IDS in environment", err)
	}
	for rows.Next() {
		rowResult := make(map[string]interface{})
		remoteEndpoint := new(RemoteEndpoint)
		err = rows.MapScan(rowResult)
		if err != nil {
			return nil, aperrors.NewWrapped("[model/remote_endpoint.go] Error scanning row while getting all remote endpoints", err)
		}
		if apiID, ok := rowResult["api_id"].(int64); ok {
			remoteEndpoint.APIID = apiID
		}
		if id, ok := rowResult["id"].(int64); ok {
			remoteEndpoint.ID = id
		}
		if name, ok := rowResult["name"].([]byte); ok {
			remoteEndpoint.Name = string(name)
		}
		if codename, ok := rowResult["codename"].([]byte); ok {
			remoteEndpoint.Codename = string(codename)
		}
		if description, ok := rowResult["description"].([]byte); ok {
			remoteEndpoint.Description = string(description)
		}
		if _type, ok := rowResult["type"].([]byte); ok {
			remoteEndpoint.Type = string(_type)
		}
		if status, ok := rowResult["status"].([]byte); ok {
			remoteEndpoint.Status = apsql.MakeNullString(string(status))
		}
		if statusMessage, ok := rowResult["statusMessage"].([]byte); ok {
			remoteEndpoint.StatusMessage = apsql.MakeNullString(string(statusMessage))
		}
		if data, ok := rowResult["data"].([]byte); ok {
			remoteEndpoint.Data = types.JsonText(json.RawMessage(data))
		}
		if selectedEnvData, ok := rowResult["selected_env_data"].([]byte); ok {
			envData := types.JsonText(json.RawMessage(selectedEnvData))
			remoteEndpoint.SelectedEnvironmentData = &envData
		}
		if soapID, ok := rowResult["soap_id"].(int64); ok {
			remoteEndpoint.Soap = new(SoapRemoteEndpoint)
			remoteEndpoint.Soap.ID = soapID
		}
		if wsdl, ok := rowResult["wsdl"].([]byte); ok {
			remoteEndpoint.Soap.Wsdl = string(wsdl)
		}

		remoteEndpoints = append(remoteEndpoints, remoteEndpoint)
	}
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
		remote_endpoints.status_message as status_message,
		soap_remote_endpoints.id as soap_id,
		soap_remote_endpoints.wsdl as wsdl
	FROM remote_endpoints, apis
	LEFT JOIN soap_remote_endpoints ON remote_endpoints.id = soap_remote_endpoints.remote_endpoint_id
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

	remoteEndpoints, err := mapRemoteEndpoints(db, query, args...)
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

func beforeDelete(remoteEndpoint *RemoteEndpoint, tx *apsql.Tx) error {
	if remoteEndpoint.Status.String == RemoteEndpointStatusPending {
		return fmt.Errorf("Unable to delete remote endpoint -- status is currently %s", RemoteEndpointStatusPending)
	}

	if remoteEndpoint.Type != RemoteEndpointTypeSoap {
		return nil
	}

	soapRemoteEndpoint, err := FindSoapRemoteEndpointByRemoteEndpointID(tx.DB, remoteEndpoint.ID)
	if err != nil {
		return aperrors.NewWrapped("[model/remote_endpoint.go] Unable to find soap remote endpoint by remote endpoint ID", err)
	}

	remoteEndpoint.Soap = soapRemoteEndpoint

	return nil
}

// DeleteRemoteEndpointForAPIIDAndAccountID deletes the remoteEndpoint with the id, api_id and account_id specified.
func DeleteRemoteEndpointForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID, userID int64) error {
	var endpoints []*RemoteEndpoint
	err := tx.Select(&endpoints,
		`SELECT remote_endpoints.id as id,
		  remote_endpoints.type as type,
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

	if err := beforeDelete(endpoint, tx); err != nil {
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

	if err = afterDelete(endpoint, accountID, userID, apiID, tx); err != nil {
		return err
	}

	return tx.Notify("remote_endpoints", accountID, userID, apiID, id, apsql.Delete, msg)
}

func afterDelete(remoteEndpoint *RemoteEndpoint, accountID, userID, apiID int64, tx *apsql.Tx) error {

	if remoteEndpoint.Type != RemoteEndpointTypeSoap {
		return nil
	}

	err := DeleteJarFile(remoteEndpoint.Soap.ID)
	if err != nil {
		log.Printf("%s Unable to delete jar file for SoapRemoteEndpoint: %v", config.System, err)
	}

	// trigger a notification for soap_remote_endpoints
	err = tx.Notify("soap_remote_endpoints", accountID, userID, apiID, remoteEndpoint.Soap.ID, apsql.Delete, remoteEndpoint.Soap.ID)
	if err != nil {
		return fmt.Errorf("%s Failed to send notification that soap_remote_endpoint was deleted for id %d: %v", config.System, remoteEndpoint.Soap.ID, err)
	}

	return nil
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

func removeJSONField(jsonText types.JsonText, fieldName string) (types.JsonText, error) {
	dataAsByteArray := []byte(json.RawMessage(jsonText))
	targetMap := make(map[string]interface{})
	err := json.Unmarshal(dataAsByteArray, &targetMap)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode data: %v", err)
	}

	delete(targetMap, fieldName)
	result, err := json.Marshal(targetMap)
	if err != nil {
		return nil, fmt.Errorf("Unable to encode data: %v", err)
	}

	return types.JsonText(json.RawMessage(result)), nil
}

func (e *RemoteEndpoint) beforeInsert(tx *apsql.Tx) error {
	if e.Type != RemoteEndpointTypeSoap {
		return nil
	}

	e.Status = apsql.MakeNullString(RemoteEndpointStatusPending)
	soap, err := NewSoapRemoteEndpoint(e)
	if err != nil {
		return fmt.Errorf("Unable to construct SoapRemoteEndpoint object: %v", err)
	}

	e.Soap = &soap

	var newVal types.JsonText
	if newVal, err = removeJSONField(e.Data, "wsdl"); err != nil {
		return err
	}
	e.Data = newVal

	return nil
}

// Insert inserts the remoteEndpoint into the database as a new row.
func (e *RemoteEndpoint) Insert(tx *apsql.Tx) error {
	encodedData, err := marshaledForStorage(e.Data)
	if err != nil {
		return err
	}

	if err := e.beforeInsert(tx); err != nil {
		return err
	}

	e.ID, err = tx.InsertOne(
		`INSERT INTO remote_endpoints (api_id, name, codename, description, type, status, status_message, data)
		VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?),?,?,?,?,?,?,?)`,
		e.APIID, e.AccountID, e.Name, e.Codename, e.Description, e.Type, e.Status, e.StatusMessage, encodedData)
	if err != nil {
		return err
	}

	if err := e.afterInsert(tx); err != nil {
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
	return tx.Notify("remote_endpoints", e.AccountID, e.UserID, e.APIID, e.ID, apsql.Insert)
}

func (e *RemoteEndpoint) afterInsert(tx *apsql.Tx) error {
	if e.Type != RemoteEndpointTypeSoap {
		return nil
	}

	e.Soap.RemoteEndpointID = e.ID
	err := e.Soap.Insert(tx)
	if err != nil {
		return fmt.Errorf("Unable to insert SoapRemoteEndpoint: %v", err)
	}

	return nil
}

func (e *RemoteEndpoint) beforeUpdate(tx *apsql.Tx) error {
	existingRemoteEndpoint, err := FindRemoteEndpointForAPIIDAndAccountID(tx.DB, e.ID, e.APIID, e.AccountID)
	if err != nil {
		return aperrors.NewWrapped("[remote_endpoints.go BeforeUpdate] Unable to fetch existing remote endpoint with id %d, api ID %d, account ID %d", err)
	}

	if existingRemoteEndpoint.Status.String == RemoteEndpointStatusPending {
		return fmt.Errorf("Unable to update remote endpoint %d -- status is currently %s", e.ID, RemoteEndpointStatusPending)
	}

	e.Status = existingRemoteEndpoint.Status
	e.StatusMessage = existingRemoteEndpoint.StatusMessage

	if e.Type != RemoteEndpointTypeSoap {
		return nil
	}

	soap, err := NewSoapRemoteEndpoint(e)
	if err != nil {
		return fmt.Errorf("Unable to construct SoapRemoteEndpoint object for update: %v", err)
	}

	soapRemoteEndpoint, err := FindSoapRemoteEndpointByRemoteEndpointID(tx.DB, e.ID)
	if err != nil {
		return fmt.Errorf("Unable to fetch SoapRemoteEndpoint with remote_endpoint_id of %d: %v", e.ID, err)
	}

	e.Soap = soapRemoteEndpoint
	var newVal types.JsonText
	if newVal, err = removeJSONField(e.Data, "wsdl"); err != nil {
		return err
	}
	e.Data = newVal

	if soap.Wsdl == "" {
		return nil
	}

	soapRemoteEndpoint.Wsdl = soap.Wsdl
	soapRemoteEndpoint.GeneratedJarThumbprint = ""
	soapRemoteEndpoint.RemoteEndpoint = e

	e.Status = apsql.MakeNullString(RemoteEndpointStatusPending)
	e.StatusMessage = apsql.MakeNullStringNull()

	return nil
}

// Update updates the remoteEndpoint in the database.
func (e *RemoteEndpoint) Update(tx *apsql.Tx) error {
	return e.update(tx, true)
}

func (e *RemoteEndpoint) update(tx *apsql.Tx, fireLifecycleHooks bool) error {
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

	if fireLifecycleHooks {
		if err := e.beforeUpdate(tx); err != nil {
			return err
		}
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

	if fireLifecycleHooks {
		if err := e.afterUpdate(tx); err != nil {
			return err
		}
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
		return tx.Notify("remote_endpoints", e.AccountID, e.UserID, e.APIID, e.ID, apsql.Update, msg)
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
	return tx.Notify("remote_endpoints", e.AccountID, e.UserID, e.APIID, e.ID, apsql.Update, msg)
}

func (e *RemoteEndpoint) afterUpdate(tx *apsql.Tx) error {
	if e.Type != RemoteEndpointTypeSoap {
		return nil
	}

	if e.Soap.Wsdl != "" {
		err := e.Soap.Update(tx)
		if err != nil {
			return fmt.Errorf("Unable to update SoapRemoteEndpoint: %v", err)
		}
	}

	return nil
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
