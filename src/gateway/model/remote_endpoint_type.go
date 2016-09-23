package model

import (
	"fmt"

	"gateway/config"
	aperrors "gateway/errors"
	"gateway/sql"
)

// RemoteEndpointType represents a type of remote endpoint.  Enabled indicates
// whether or not the syatem currently supports that endpoint.
type RemoteEndpointType struct {
	ID      int64  `json:"id"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

const (
	// RemoteEndpointTypeHTTP denotes that a remote endpoint is an HTTP endpoint
	RemoteEndpointTypeHTTP = "http"
	// RemoteEndpointTypeSQLServer denotes that a remote endpoint is a MS SQL Server database
	RemoteEndpointTypeSQLServer = "sqlserver"
	// RemoteEndpointTypeMySQL denotes that a remote endpoint is a MySQL Server database
	RemoteEndpointTypeMySQL = "mysql"
	// RemoteEndpointTypePostgres denotes that a remote endpoint is a PostgreSQL Server database
	RemoteEndpointTypePostgres = "postgres"
	// RemoteEndpointTypeMongo denotes that a remote endpoint is a MongoDB database
	RemoteEndpointTypeMongo = "mongodb"
	// RemoteEndpointTypeScript denotes that a remote endpoint is a SOAP service
	RemoteEndpointTypeScript = "script"
	// RemoteEndpointTypeSoap denotes that a remote endpoint is a SOAP service
	RemoteEndpointTypeSoap = "soap"
	// RemoteEndpointTypeStore denotes that a remote endpoint is an Object Store database
	RemoteEndpointTypeStore = "store"
	// RemoteEndpointTypeLDAP denotes that a remote endpoint is an LDAP service
	RemoteEndpointTypeLDAP = "ldap"
	// RemoteEndpointTypeHana denotes that a remote endpoint is an SAP Hana database
	RemoteEndpointTypeHana = "hana"
	// RemoteEndpointTypePush denotes that a remote endpoint is an Push service
	RemoteEndpointTypePush = "push"
	// RemoteEndpointTypeRedis denotes that a remote endpoint is a Redis database
	RemoteEndpointTypeRedis = "redis"
	// RemoteEndpointTypeSMTP denotes that a remote endpoint is an SMTP service
	RemoteEndpointTypeSMTP = "smtp"
	// RemoteEndpointTypeDocker denotes that a remote endpoint is a docker endpoint
	RemoteEndpointTypeDocker = "docker"
	// RemoteEndpointTypejob denotes that a remote endpoint is a job endpoint
	RemoteEndpointTypeJob = "job"
)

var remoteEndpointTypes map[string]*RemoteEndpointType

// InitializeRemoteEndpointTypes configures which remote endpoints are currently supported within the system
func InitializeRemoteEndpointTypes(reConf config.RemoteEndpoint) {
	remoteEndpointTypes = map[string]*RemoteEndpointType{
		RemoteEndpointTypeHTTP:      &RemoteEndpointType{ID: 1, Value: RemoteEndpointTypeHTTP, Enabled: reConf.HTTPEnabled},
		RemoteEndpointTypeSQLServer: &RemoteEndpointType{ID: 2, Value: RemoteEndpointTypeSQLServer, Enabled: reConf.SQLServerEnabled},
		RemoteEndpointTypeMySQL:     &RemoteEndpointType{ID: 3, Value: RemoteEndpointTypeMySQL, Enabled: reConf.MySQLEnabled},
		RemoteEndpointTypePostgres:  &RemoteEndpointType{ID: 4, Value: RemoteEndpointTypePostgres, Enabled: reConf.PostgreSQLEnabled},
		RemoteEndpointTypeMongo:     &RemoteEndpointType{ID: 5, Value: RemoteEndpointTypeMongo, Enabled: reConf.MongoDBEnabled},
		RemoteEndpointTypeScript:    &RemoteEndpointType{ID: 6, Value: RemoteEndpointTypeScript, Enabled: reConf.ScriptEnabled},
		RemoteEndpointTypeSoap:      &RemoteEndpointType{ID: 7, Value: RemoteEndpointTypeSoap, Enabled: reConf.SoapEnabled},
		RemoteEndpointTypeLDAP:      &RemoteEndpointType{ID: 8, Value: RemoteEndpointTypeLDAP, Enabled: reConf.LDAPEnabled},
		RemoteEndpointTypeStore:     &RemoteEndpointType{ID: 9, Value: RemoteEndpointTypeStore, Enabled: reConf.StoreEnabled},
		RemoteEndpointTypeHana:      &RemoteEndpointType{ID: 10, Value: RemoteEndpointTypeHana, Enabled: reConf.HanaEnabled},
		RemoteEndpointTypePush:      &RemoteEndpointType{ID: 11, Value: RemoteEndpointTypePush, Enabled: reConf.PushEnabled},
		RemoteEndpointTypeRedis:     &RemoteEndpointType{ID: 12, Value: RemoteEndpointTypeRedis, Enabled: reConf.RedisEnabled},
		RemoteEndpointTypeSMTP:      &RemoteEndpointType{ID: 13, Value: RemoteEndpointTypeSMTP, Enabled: reConf.SMTPEnabled},
		RemoteEndpointTypeDocker:    &RemoteEndpointType{ID: 14, Value: RemoteEndpointTypeDocker, Enabled: reConf.DockerEnabled},
		RemoteEndpointTypeJob:       &RemoteEndpointType{ID: 15, Value: RemoteEndpointTypeJob, Enabled: reConf.JobEnabled},
	}
}

// IsRemoteEndpointTypeEnabled tells whether a specific remote endpoint type
// is enabled
func IsRemoteEndpointTypeEnabled(reType string) bool {
	var (
		endpointType *RemoteEndpointType
		found        bool
	)

	if endpointType, found = remoteEndpointTypes[reType]; !found {
		return true
	}

	return endpointType.Enabled
}

// AllRemoteEndpointTypes finds all RemoteEndpointTypes
func AllRemoteEndpointTypes(db *sql.DB) ([]*RemoteEndpointType, error) {
	all, i := make([]*RemoteEndpointType, len(remoteEndpointTypes)), 0
	for _, v := range remoteEndpointTypes {
		all[i] = v
		i++
	}

	return all, nil
}

// FindRemoteEndpointType finds a the RemoteEndpointType with the given id
func FindRemoteEndpointType(db *sql.DB, id int64) (*RemoteEndpointType, error) {
	// we can safely swallow this error since the target function doesn't actually throw one
	for _, ret := range remoteEndpointTypes {
		if ret.ID == id {
			return ret, nil
		}
	}
	return nil, fmt.Errorf("Not found")
}

// DeleteRemoteEndpointType is not supported
func DeleteRemoteEndpointType(tx *sql.Tx, id int64) error {
	return fmt.Errorf("DeleteRemoteEndpointType is not supported")
}

// Insert is not supported
func (r *RemoteEndpointType) Insert(tx *sql.Tx) (err error) {
	return fmt.Errorf("Insert is not supported")
}

// Update is not supported
func (r *RemoteEndpointType) Update(tx *sql.Tx) error {
	return fmt.Errorf("Update is not supported")
}

// Validate validates the model.  This implementation doesn't really do anything
// since inserts and updates are not supported.  It simply returns an empty
// errors object.
func (r *RemoteEndpointType) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.  This implementation doesn't really do anything since
// inserts and updates are not supported.  It simply returns an empty errors
// object
func (r *RemoteEndpointType) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	return errors
}
