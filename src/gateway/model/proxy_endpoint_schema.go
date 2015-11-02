package model

import (
	"fmt"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
	"github.com/xeipuuv/gojsonschema"
)

type ProxyEndpointSchema struct {
	AccountID       int64 `json:"-"`
	UserID          int64 `json:"-"`
	APIID           int64 `json:"-" db:"api_id"`
	ProxyEndpointID int64 `json:"endpoint_id,omitempty" db:"endpoint_id"`

	ID                    int64  `json:"id,omitempty"`
	Name                  string `json:"name"`
	RequestSchemaID       *int64 `json:"request_schema_id,omitempty" db:"request_schema_id"`
	RequestType           string `json:"request_type" db:"request_type"`
	RequestSchema         string `json:"request_schema" db:"request_schema"`
	ResponseSameAsRequest bool   `json:"response_same_as_request" db:"response_same_as_request"`
	ResponseSchemaID      *int64 `json:"response_schema_id,omitempty" db:"response_schema_id"`
	ResponseType          string `json:"response_type" db:"response_type"`
	ResponseSchema        string `json:"response_schema" db:"response_schema"`

	Data types.JsonText `json:"data" db:"data"`
}

func (s *ProxyEndpointSchema) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)
	if s.Name == "" {
		errors.Add("name", "must have a name")
	}
	if s.RequestType != "json" {
		errors.Add("request_type", "must be 'json'")
	}
	if s.ResponseType != "json" {
		errors.Add("response_type", "must be 'json'")
	}
	if s.RequestSchema != "" {
		schema := gojsonschema.NewStringLoader(s.RequestSchema)
		_, err := gojsonschema.NewSchema(schema)
		errors.Add("request_schema", fmt.Sprintf("schema error: %v", err))
	}
	if s.ResponseSchema != "" {
		schema := gojsonschema.NewStringLoader(s.ResponseSchema)
		_, err := gojsonschema.NewSchema(schema)
		errors.Add("response_schema", fmt.Sprintf("schema error: %v", err))
	}
	return errors
}

func (s *ProxyEndpointSchema) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "proxy_endpoints", "endpoint_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func AllProxyEndpointSchemasForProxyEndpointIDAndAPIIDAndAccountID(db *apsql.DB,
	proxyEndpointID, apiID, accountID int64) ([]*ProxyEndpointSchema, error) {
	schemas := []*ProxyEndpointSchema{}
	err := db.Select(&schemas, db.SQL("proxy_endpoint_schemas/all"), proxyEndpointID, apiID, accountID)
	return schemas, err
}

func FindProxyEndpointSchemasForProxy(db *apsql.DB, proxyEndpointID, apiID int64) ([]*ProxyEndpointSchema, error) {
	schemas := []*ProxyEndpointSchema{}
	err := db.Select(&schemas, db.SQL("proxy_endpoint_schemas/find_proxy"), proxyEndpointID, apiID)
	return schemas, err
}

func FindProxyEndpointSchemaForProxyEndpointIDAndAPIIDAndAccountID(db *apsql.DB,
	id, proxyEndpointID, apiID, accountID int64) (*ProxyEndpointSchema, error) {
	schema := ProxyEndpointSchema{}
	err := db.Get(&schema, db.SQL("proxy_endpoint_schemas/find"), id, proxyEndpointID, apiID, accountID)
	return &schema, err
}

func DeleteProxyEndpointSchemaForProxyEndpointIDAndAPIIDAndAccountID(tx *apsql.Tx,
	id, proxyEndpointID, apiID, accountID, userID int64) error {
	err := tx.DeleteOne(tx.SQL("proxy_endpoint_schemas/delete"), id, proxyEndpointID, apiID)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoint_schemas", accountID, userID, apiID, id, apsql.Delete)
}

func (s *ProxyEndpointSchema) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	s.ID, err = tx.InsertOne(tx.SQL("proxy_endpoint_schemas/insert"), s.ProxyEndpointID,
		s.APIID, s.Name, s.RequestSchemaID, s.APIID, s.RequestType, s.RequestSchema,
		s.ResponseSameAsRequest, s.ResponseSchemaID, s.APIID, s.ResponseType,
		s.ResponseSchema, data)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoint_schemas", s.AccountID, s.UserID, s.APIID, s.ID, apsql.Insert)
}

func (s *ProxyEndpointSchema) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(tx.SQL("proxy_endpoint_schemas/update"), s.Name, s.RequestSchemaID,
		s.APIID, s.RequestType, s.RequestSchema, s.ResponseSameAsRequest, s.ResponseSchemaID,
		s.APIID, s.ResponseType, s.ResponseSchema, data, s.ID, s.ProxyEndpointID, s.APIID)
	if err != nil {
		return err
	}
	return tx.Notify("proxy_endpoint_schemas", s.AccountID, s.UserID, s.APIID, s.ID, apsql.Update)
}
