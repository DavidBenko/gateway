package model

import (
	"errors"
	"fmt"
	"gateway/license"
	apsql "gateway/sql"
)

// API represents a top level grouping of endpoints accessible at a host.
type API struct {
	AccountID int64 `json:"-" db:"account_id"`

	ID                   int64  `json:"id,omitempty"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	CORSAllowOrigin      string `json:"cors_allow_origin" db:"cors_allow_origin"`
	CORSAllowHeaders     string `json:"cors_allow_headers" db:"cors_allow_headers"`
	CORSAllowCredentials bool   `json:"cors_allow_credentials" db:"cors_allow_credentials"`
	CORSRequestHeaders   string `json:"cors_request_headers" db:"cors_request_headers"`
	CORSMaxAge           int64  `json:"cors_max_age" db:"cors_max_age"`
	Export               string `json:"export" db:"-"`

	Environments    []*Environment    `json:"environments,omitempty"`
	EndpointGroups  []*EndpointGroup  `json:"endpoint_groups,omitempty"`
	Libraries       []*Library        `json:"libraries,omitempty"`
	RemoteEndpoints []*RemoteEndpoint `json:"remote_endpoints,omitempty"`
	ProxyEndpoints  []*ProxyEndpoint  `json:"proxy_endpoints,omitempty"`

	ExportVersion int64 `json:"export_version,omitempty"`
}

// Validate validates the model.
func (a *API) Validate() Errors {
	errors := make(Errors)
	if a.Name == "" {
		errors.add("name", "must not be blank")
	}
	if a.CORSAllowOrigin == "" {
		errors.add("cors_allow_origin", "must not be blank (use '*' for everything)")
	}
	if a.CORSAllowHeaders == "" {
		errors.add("cors_allow_headers", "must not be blank (use '*' for everything)")
	}
	if a.CORSRequestHeaders == "" {
		errors.add("cors_request_headers", "must not be blank (use '*' for everything)")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (a *API) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	if apsql.IsUniqueConstraint(err, "apis", "account_id", "name") {
		errors.add("name", "is already taken")
	}
	return errors
}

// AllAPIsForAccountID returns all apis on the Account in default order.
func AllAPIsForAccountID(db *apsql.DB, accountID int64) ([]*API, error) {
	apis := []*API{}
	err := db.Select(&apis, db.SQL("apis/all"), accountID)
	return apis, err
}

// FindAPIForAccountID returns the api with the id and account_id specified.
func FindAPIForAccountID(db *apsql.DB, id, accountID int64) (*API, error) {
	api := API{}
	err := db.Get(&api, db.SQL("apis/find"), id, accountID)
	return &api, err
}

// FindAPIForAccountID returns the api with the id specified.
func FindAPIForProxy(db *apsql.DB, id int64) (*API, error) {
	api := API{}
	err := db.Get(&api, db.SQL("apis/find_proxy"), id)
	return &api, err
}

// DeleteAPIForAccountID deletes the api with the id and account_id specified.
func DeleteAPIForAccountID(tx *apsql.Tx, id, accountID int64) error {
	err := tx.DeleteOne(tx.SQL("apis/delete"), id, accountID)
	if err != nil {
		return err
	}
	return tx.Notify("apis", id, apsql.Delete)
}

// Insert inserts the api into the database as a new row.
func (a *API) Insert(tx *apsql.Tx) (err error) {
	if license.DeveloperVersion {
		var count int
		tx.Get(&count, tx.SQL("apis/count"), a.AccountID)
		if count >= license.DeveloperVersionAPIs {
			return errors.New(fmt.Sprintf("Developer version allows %v api(s).", license.DeveloperVersionAPIs))
		}
	}

	a.ID, err = tx.InsertOne(tx.SQL("apis/insert"),
		a.AccountID, a.Name, a.Description, a.CORSAllowOrigin, a.CORSAllowHeaders,
		a.CORSAllowCredentials, a.CORSRequestHeaders, a.CORSMaxAge)
	return
}

// Update updates the api in the database.
func (a *API) Update(tx *apsql.Tx) error {
	return tx.UpdateOne(tx.SQL("apis/update"),
		a.Name, a.Description, a.CORSAllowOrigin, a.CORSAllowHeaders,
		a.CORSAllowCredentials, a.CORSRequestHeaders, a.CORSMaxAge,
		a.ID, a.AccountID)
}
