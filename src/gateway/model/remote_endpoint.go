package model

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"gateway/code"
	"gateway/config"
	"gateway/db"
	aperrors "gateway/errors"
	aphttp "gateway/http"
	"gateway/logreport"
	re "gateway/model/remote_endpoint"
	"gateway/soap"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
	"github.com/vincent-petithory/dataurl"
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

type RemoteEndpointEnvironmentDataLinks struct {
	ScratchPads string `json:"scratch_pads"`
}

// RemoteEndpointEnvironmentData contains per-environment endpoint data
type RemoteEndpointEnvironmentData struct {
	// TODO: add type to remote_endpoint_environment_data table
	ID               int64          `json:"id,omitempty"`
	RemoteEndpointID int64          `json:"remote_endpoint_id" db:"remote_endpoint_id"`
	EnvironmentID    int64          `json:"environment_id,omitempty" db:"environment_id"`
	Name             string         `json:"-"`
	Type             string         `json:"type"`
	Data             types.JsonText `json:"data"`

	Links *RemoteEndpointEnvironmentDataLinks `json:"links,omitempty"`

	ExportEnvironmentIndex int `json:"environment_index,omitempty"`
}

func (e *RemoteEndpointEnvironmentData) AddLinks(apiID int64) {
	e.Links = &RemoteEndpointEnvironmentDataLinks{
		ScratchPads: fmt.Sprintf("/apis/%v/remote_endpoints/%v/environment_data/%v/scratch_pads",
			apiID, e.RemoteEndpointID, e.ID),
	}
}

// HTTPRequest encapsulates a request made over HTTP(s).
type HTTPRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Body    string                 `json:"body"`
	Headers map[string]interface{} `json:"headers"`
	Query   map[string]string      `json:"query"`
}

// Validate validates the model.
func (e *RemoteEndpoint) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if e.Name == "" {
		errors.Add("name", "must not be blank")
	}
	if e.Codename == "" {
		errors.Add("codename", "must not be blank")
	}
	if code.IsReserved(e.Codename) {
		errors.Add("codename", "is a reserved word and may not be used")
	}
	if !code.IsValidVariableIdentifier(e.Codename) {
		errors.Add("codename", "must start with A-Z a-z _ and may only contain A-Z a-z 0-9 _")
	}
	switch e.Type {
	case RemoteEndpointTypeHTTP:
		e.ValidateHTTP(errors)
	case RemoteEndpointTypeSoap:
		e.ValidateSOAP(errors, isInsert)
	case RemoteEndpointTypeStore:
	case RemoteEndpointTypeScript:
		e.ValidateScript(errors)
	case RemoteEndpointTypeMySQL, RemoteEndpointTypeSQLServer,
		RemoteEndpointTypePostgres, RemoteEndpointTypeMongo, RemoteEndpointTypeHana,
		RemoteEndpointTypeRedis, RemoteEndpointTypeOracle:
		_, err := e.DBConfig()
		if err != nil {
			errors.Add("base", fmt.Sprintf("error in database config: %s", err))
		}
	case RemoteEndpointTypeLDAP:
		e.ValidateLDAP(errors)
	case RemoteEndpointTypePush:
		e.ValidatePush(errors)
	case RemoteEndpointTypeSMTP:
		e.ValidateSMTP(errors)
	case RemoteEndpointTypeDocker:
		e.ValidateDocker(errors)
	default:
		errors.Add("base", fmt.Sprintf("unknown endpoint type %q", e.Type))
	}
	if status := e.Status; status.Valid {
		val, _ := status.Value() // error always nil
		switch val {
		case RemoteEndpointStatusFailed, RemoteEndpointStatusProcessing,
			RemoteEndpointStatusSuccess, RemoteEndpointStatusPending:
		default:
			errors.Add("status", fmt.Sprintf("Invalid value for status: %v", val))
		}
	}

	return errors
}

func (e *RemoteEndpoint) ValidateSOAP(errors aperrors.Errors, isInsert bool) {
	if !soap.Available() {
		errors.Add("base", "SOAP is not currently available.  Requisite dependencies must be met")
	}

	var (
		sc  *re.Soap
		err error
	)
	if sc, err = re.SoapConfig(e.Data); err != nil {
		errors.Add("base", fmt.Sprintf("Unable to validate soap configuration: %v", err))
	}

	if sc.WSDL == "" && isInsert {
		errors.Add("wsdl", "WSDL is required for new SOAP endpoints")
		return
	}

	e.Soap = NewSoapRemoteEndpoint(e)
	if sc.WSDL != "" {
		decoded, err := dataurl.DecodeString(sc.WSDL)
		if err != nil {
			errors.Add("wsdl", "Unable to decode WSDL.")
			return
		}

		wsdlDoc := struct {
			XMLName xml.Name `xml:"definitions"`
		}{}
		if err = xml.Unmarshal(decoded.Data, &wsdlDoc); err != nil {
			errors.Add("wsdl", "Must be a valid WSDL file")
			return
		}

		xmlName := xml.Name{Space: "http://schemas.xmlsoap.org/wsdl/", Local: "definitions"}
		if wsdlDoc.XMLName != xmlName {
			errors.Add("wsdl", "Must be a valid WSDL file")
			return
		}

		e.Soap.Wsdl = string(decoded.Data)
	}

}

func ValidateURL(rurl string) aperrors.Errors {
	errors := make(aperrors.Errors)
	if len(rurl) > 0 {
		if !strings.HasPrefix(rurl, "http://") && !strings.HasPrefix(rurl, "https://") {
			errors.Add("url", "url must start with 'http://' or 'https://'")
			return errors
		}
		purl, err := url.ParseRequestURI(rurl)
		if err != nil {
			errors.Add("url", fmt.Sprintf("error parsing url: %s", err))
			return errors
		}
		switch purl.Scheme {
		case "http", "https":
		default:
			errors.Add("url", "url scheme must be 'http' or 'https'")
			return errors
		}
	}
	return nil
}

func (e *RemoteEndpoint) ValidateHTTP(errors aperrors.Errors) {
	request := &HTTPRequest{}
	if err := json.Unmarshal(e.Data, request); err != nil {
		errors.Add("base", fmt.Sprintf("error in http config: %s", err))
		return
	}

	if errs := ValidateURL(request.URL); errs != nil {
		errors.AddAll(errs)
		return
	}
	for _, environment := range e.EnvironmentData {
		request := &HTTPRequest{}
		if err := json.Unmarshal(environment.Data, request); err != nil {
			errors.Add("environment_data", fmt.Sprintf("error in environment http config: %s", err))
			return
		}
		if errs := ValidateURL(request.URL); errs != nil {
			errs.MoveAllToName("environment_data")
			errors.AddAll(errs)
			return
		}
	}
}

func (e *RemoteEndpoint) ValidateScript(errors aperrors.Errors) {
	script := &re.Script{}
	if err := json.Unmarshal(e.Data, script); err != nil {
		errors.Add("base", fmt.Sprintf("error in script config: %s", err))
		return
	}
	if errs := script.Validate(); errs != nil {
		errors.AddAll(errs)
	}

	for _, environment := range e.EnvironmentData {
		escript := &re.Script{}
		if err := json.Unmarshal(environment.Data, escript); err != nil {
			errors.Add("environment_data", fmt.Sprintf("error in script config: %s", err))
			return
		}
		script_copy := &re.Script{}
		*script_copy = *script
		script_copy.UpdateWith(escript)
		if errs := script_copy.Validate(); errs != nil {
			errs.MoveAllToName("environment_data")
			errors.AddAll(errs)
		}
	}
}

// Validate LDAP endpoint configuration
func (e *RemoteEndpoint) ValidateLDAP(errors aperrors.Errors) {
	ldap := re.LDAP{}
	if err := json.Unmarshal(e.Data, &ldap); err != nil {
		errors.Add("base", "error in ldap config")
		return
	}

	if errs := ldap.Validate(); errs != nil {
		errors.AddAll(errs)
		return
	}

	data, err := json.Marshal(ldap)
	if err != nil {
		errors.Add("base", "error re-encoding data")
		return
	}
	e.Data = data
}

func (e *RemoteEndpoint) ValidatePush(errors aperrors.Errors) {
	push := &re.Push{}
	if err := json.Unmarshal(e.Data, push); err != nil {
		errors.Add("push_platforms", fmt.Sprintf("error in push config: %s", err))
		return
	}
	if errs := push.Validate(); errs != nil {
		errs.MoveAllToName("push_platforms")
		errors.AddAll(errs)
	}

	for _, environment := range e.EnvironmentData {
		epush := &re.Push{}
		if err := json.Unmarshal(environment.Data, epush); err != nil {
			errors.Add("push_platforms", fmt.Sprintf("error in push config: %s", err))
			return
		}
		epush.UpdateWith(push)
		if errs := epush.Validate(); errs != nil {
			errs.MoveAllToName("push_platforms")
			errors.AddAll(errs)
		}
	}
}

func (e *RemoteEndpoint) ValidateSMTP(errors aperrors.Errors) {
	smtp := &re.Smtp{}

	if err := json.Unmarshal(e.Data, smtp); err != nil {
		errors.Add("smtp", fmt.Sprintf("error in smtp config: %s", err))
	}

	if errs := smtp.Validate(); errs != nil {
		errors.AddAll(errs)
		return
	}
}

// Validate Docker endpoint configuration
func (e *RemoteEndpoint) ValidateDocker(errors aperrors.Errors) {
	docker := &re.Docker{}

	if err := json.Unmarshal(e.Data, docker); err != nil {
		errors.Add("docker", fmt.Sprintf("error in docker config: %s", err))
	}

	if errs := docker.Validate(); errs != nil {
		errors.AddAll(errs)
		return
	}
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *RemoteEndpoint) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "remote_endpoints", "api_id", "name") {
		errors.Add("name", "is already taken")
	}
	if apsql.IsNotNullConstraint(err, "remote_endpoint_environment_data", "environment_id") {
		errors.Add("environment_data", "must include a valid environment in this API")
	}
	if apsql.IsUniqueConstraint(err, "remote_endpoint_environment_data", "remote_endpoint_id", "environment_id") {
		errors.Add("environment_data", "environment is already taken")
	}
	return errors
}

// AllRemoteEndpointsForAPIIDAndAccountID returns all remoteEndpoints on the Account's API in default order.
func AllRemoteEndpointsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*RemoteEndpoint, error) {
	return _remoteEndpoints(db, "", 0, apiID, accountID)
}

func WriteAllScriptFiles(db *apsql.DB) error {
	endpoints, err := _remoteEndpoints(db, "", 0, 0, 0)
	if err != nil {
		return err
	}
	for _, endpoint := range endpoints {
		if err := endpoint.WriteScript(); err != nil {
			return err
		}
	}
	return nil
}

// AllRemoteEndpointsForIDsInEnvironment returns all remoteEndpoints with id specified,
// populated with environment data
func AllRemoteEndpointsForIDsInEnvironment(db *apsql.DB, ids []int64, environmentID int64) ([]*RemoteEndpoint, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	idQuery := apsql.NQs(len(ids))
	query := `SELECT
		apis.account_id as account_id,
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
	JOIN apis ON remote_endpoints.api_id = apis.id
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

func newRemoteEndpoint(rowResult map[string]interface{}) *RemoteEndpoint {
	remoteEndpoint := new(RemoteEndpoint)

	if accountID, ok := rowResult["account_id"].(int64); ok {
		remoteEndpoint.AccountID = accountID
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

	return remoteEndpoint
}

func mapRemoteEndpoints(db *apsql.DB, query string, args ...interface{}) ([]*RemoteEndpoint, error) {
	remoteEndpoints := []*RemoteEndpoint{}
	rows, err := db.Queryx(query, args...)
	if err != nil {
		return nil, aperrors.NewWrapped("[model/remote_endpoint.go] Error fetching all remote endpoints for IDS in environment", err)
	}
	for rows.Next() {
		rowResult := make(map[string]interface{})
		err := rows.MapScan(rowResult)
		if err != nil {
			return nil, aperrors.NewWrapped("[model/remote_endpoint.go] Error scanning row while getting all remote endpoints", err)
		}

		remoteEndpoints = append(remoteEndpoints, newRemoteEndpoint(rowResult))
	}
	return remoteEndpoints, err
}

// FindRemoteEndpointForAPIIDAndAccountID returns the remoteEndpoint with the id, api id, and account_id specified.
func FindRemoteEndpointForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*RemoteEndpoint, error) {
	endpoints, err := _remoteEndpoints(db, "", id, apiID, accountID)
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("No endpoint with id %d found", id)
	}
	return endpoints[0], nil
}

func FindRemoteEndpointForCodenameAndAPIIDAndAccountID(db *apsql.DB, codename string, apiID, accountID int64) (*RemoteEndpoint, error) {
	endpoints, err := _remoteEndpoints(db, codename, 0, apiID, accountID)
	if err != nil {
		return nil, err
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("No endpoint with codename %v found", codename)
	}
	return endpoints[0], nil
}

func _remoteEndpoints(db *apsql.DB, codename string, id, apiID, accountID int64) ([]*RemoteEndpoint, error) {
	args := []interface{}{}
	query := `SELECT`
	if accountID != 0 {
		query += ` apis.account_id as account_id,`
	}
	query += ` remote_endpoints.api_id as api_id,
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
	FROM remote_endpoints`
	if accountID != 0 {
		query += " JOIN apis ON remote_endpoints.api_id = apis.id AND apis.account_id = ?"
		args = append(args, accountID)
	}
	query += " LEFT JOIN soap_remote_endpoints ON remote_endpoints.id = soap_remote_endpoints.remote_endpoint_id"
	if id != 0 || apiID != 0 {
		query += " WHERE"

		if id != 0 {
			query += " remote_endpoints.id = ?"
			args = append(args, id)
		} else if codename != "" {
			query += " remote_endpoints.codename = ?"
			args = append(args, codename)
		}

		if apiID != 0 {
			if id != 0 || codename != "" {
				query += " AND"
			}
			query += " remote_endpoints.api_id = ?"
			args = append(args, apiID)
		}
	}
	query += " ORDER BY remote_endpoints.name ASC, remote_endpoints.id ASC;"

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
			remote_endpoint_environment_data.id as id,
			remote_endpoint_environment_data.remote_endpoint_id as remote_endpoint_id,
			remote_endpoint_environment_data.environment_id as environment_id,
			remote_endpoint_environment_data.data as data,
			environments.name as name
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
		envData.Type = endpoint.Type
		envData.AddLinks(apiID)
		endpoint.EnvironmentData = append(endpoint.EnvironmentData, envData)
	}
	return remoteEndpoints, err
}

// CanDeleteRemoteEndpoint checks whether deleting would violate any constraints
func CanDeleteRemoteEndpoint(tx *apsql.Tx, id, accountID int64, auth aphttp.AuthType) error {
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
	if endpoint.Type != RemoteEndpointTypeHTTP &&
		endpoint.Type != RemoteEndpointTypeSoap &&
		endpoint.Type != RemoteEndpointTypePush {
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

	return tx.Notify("remote_endpoints", accountID, userID, apiID, 0, id, apsql.Delete, msg)
}

func afterDelete(remoteEndpoint *RemoteEndpoint, accountID, userID, apiID int64, tx *apsql.Tx) error {

	if remoteEndpoint.Type != RemoteEndpointTypeSoap {
		return nil
	}

	err := DeleteJarFile(remoteEndpoint.Soap.ID)
	if err != nil && !os.IsNotExist(err) {
		logreport.Printf("%s Unable to delete jar file for SoapRemoteEndpoint: %v", config.System, err)
	}

	// trigger a notification for soap_remote_endpoints
	err = tx.Notify("soap_remote_endpoints", accountID, userID, apiID, 0, remoteEndpoint.Soap.ID, apsql.Delete, remoteEndpoint.Soap.ID)
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
	case RemoteEndpointTypeHana:
		return re.HanaConfig(e.Data)
	case RemoteEndpointTypeRedis:
		return re.RedisConfig(e.Data)
	case RemoteEndpointTypeOracle:
		return re.OracleConfig(e.Data)
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

	var newVal types.JsonText
	var err error
	if newVal, err = removeJSONField(e.Data, "wsdl"); err != nil {
		return err
	}
	e.Data = newVal

	return nil
}

func (e *RemoteEndpoint) WriteScript() error {
	if e.Type != RemoteEndpointTypeScript {
		return nil
	}

	script := &re.Script{}
	if err := json.Unmarshal(e.Data, script); err != nil {
		return err
	}

	if err := script.WriteFile(); err != nil {
		return err
	}

	for _, environment := range e.EnvironmentData {
		escript := &re.Script{}
		if err := json.Unmarshal(environment.Data, escript); err != nil {
			return err
		}

		script_copy := &re.Script{}
		*script_copy = *script
		script_copy.UpdateWith(escript)

		if err := script_copy.WriteFile(); err != nil {
			return err
		}
	}

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
		envData.ID, err = _insertRemoteEndpointEnvironmentData(tx, e.ID, envData.EnvironmentID,
			e.APIID, encodedData)
		if err != nil {
			return err
		}
		envData.AddLinks(e.APIID)
	}

	if err := e.WriteScript(); err != nil {
		return err
	}

	return tx.Notify("remote_endpoints", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Insert)
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

	newWsdlValue := e.Soap.Wsdl

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

	if newWsdlValue == "" {
		return nil
	}

	soapRemoteEndpoint.Wsdl = newWsdlValue
	soapRemoteEndpoint.GeneratedJarThumbprint = apsql.MakeNullStringNull()
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
	if e.Type != RemoteEndpointTypeHTTP &&
		e.Type != RemoteEndpointTypeSoap &&
		e.Type != RemoteEndpointTypeLDAP &&
		e.Type != RemoteEndpointTypePush {
		conf, err := e.DBConfig()
		switch err {
		case nil:
			msg = conf
		default:
			msg = err
		}
	}

	if fireLifecycleHooks {
		if err := e.beforeUpdate(tx); err != nil {
			return err
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

	if fireLifecycleHooks {
		if err := e.afterUpdate(tx); err != nil {
			return err
		}
	}

	var existingEnvIDs []int64
	err = tx.Select(&existingEnvIDs,
		`SELECT id
		FROM remote_endpoint_environment_data
		WHERE remote_endpoint_id = ?
		ORDER BY id ASC;`,
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
		existingEnvIDs, found = popID(envData.ID, existingEnvIDs)
		if found {
			_, err = tx.Exec(
				`UPDATE remote_endpoint_environment_data
				  SET data = ?, environment_id = ?
				WHERE remote_endpoint_id = ?
				  AND id = ?;`,
				encodedData, envData.EnvironmentID, e.ID, envData.ID)
			if err != nil {
				return err
			}
		} else {
			envData.ID, err = _insertRemoteEndpointEnvironmentData(tx, e.ID, envData.EnvironmentID,
				e.APIID, encodedData)
			if err != nil {
				return err
			}
		}
		envData.AddLinks(e.APIID)
	}

	done := func() error {
		if err := e.WriteScript(); err != nil {
			return err
		}
		return tx.Notify("remote_endpoints", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Update, msg)
	}

	if len(existingEnvIDs) == 0 {
		return done()
	}

	args := []interface{}{e.ID}
	for _, envID := range existingEnvIDs {
		args = append(args, envID)
	}
	idQuery := apsql.NQs(len(existingEnvIDs))
	_, err = tx.Exec(
		`DELETE FROM remote_endpoint_environment_data
		WHERE remote_endpoint_id = ? AND id IN (`+idQuery+`);`,
		args...)

	if err != nil {
		return err
	}
	return done()
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
	data string) (int64, error) {
	return tx.InsertOne(
		`INSERT INTO remote_endpoint_environment_data
			(remote_endpoint_id, environment_id, data)
			VALUES (?, (SELECT id FROM environments WHERE id = ? AND api_id = ?), ?)`,
		rID, eID, apiID, data)
}

func (e *RemoteEndpoint) encodeWsdlForExport() error {

	encodedWsdlStr := dataurl.EncodeBytes([]byte(e.Soap.Wsdl))
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(quote)
	buf.WriteString(encodedWsdlStr)
	buf.WriteString(quote)
	encodedWsdl := json.RawMessage(buf.Bytes())

	dataPayload := make(map[string]*json.RawMessage)
	err := json.Unmarshal(e.Data, &dataPayload)
	if err != nil {
		return aperrors.NewWrapped("[model/api_import_export.go] Unmarshaling data for encoding", err)
	}

	dataPayload["wsdl"] = &encodedWsdl
	bytes, err := json.Marshal(&dataPayload)
	if err != nil {
		return aperrors.NewWrapped("[model/api_import_export.go] Marshaling data for encoding", err)
	}

	newData := json.RawMessage(bytes)
	e.Data = types.JsonText(newData)

	return nil
}
