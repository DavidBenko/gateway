package model

import (
	"errors"
	"fmt"
	aperrors "gateway/errors"
	"gateway/license"
	"gateway/logreport"
	apsql "gateway/sql"
)

var defaultAPIAccessScheme string

// ConfigureDefaultAPIAccessScheme sets the value to be used as the BaseURL on
// each API.
func ConfigureDefaultAPIAccessScheme(value string) {
	defaultAPIAccessScheme = value
}

// API represents a top level grouping of endpoints accessible at a host.
type API struct {
	AccountID int64 `json:"-" db:"account_id"`
	UserID    int64 `json:"-"`

	ID                   int64  `json:"id,omitempty"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	CORSAllowOrigin      string `json:"cors_allow_origin" db:"cors_allow_origin"`
	CORSAllowHeaders     string `json:"cors_allow_headers" db:"cors_allow_headers"`
	CORSAllowCredentials bool   `json:"cors_allow_credentials" db:"cors_allow_credentials"`
	CORSRequestHeaders   string `json:"cors_request_headers" db:"cors_request_headers"`
	CORSMaxAge           int64  `json:"cors_max_age" db:"cors_max_age"`
	EnableSwagger        bool   `json:"enable_swagger" db:"enable_swagger"`
	Export               string `json:"export,omitempty" db:"-"`

	Hosts                []*Host                `json:"-"`
	Environments         []*Environment         `json:"environments,omitempty"`
	EndpointGroups       []*EndpointGroup       `json:"endpoint_groups,omitempty"`
	Libraries            []*Library             `json:"libraries,omitempty"`
	RemoteEndpoints      []*RemoteEndpoint      `json:"remote_endpoints,omitempty"`
	ProxyEndpoints       []*ProxyEndpoint       `json:"proxy_endpoints,omitempty"`
	ProxyEndpointSchemas []*ProxyEndpointSchema `json:"proxy_endpoint_schemas,omitempty"`
	ScratchPads          []*ScratchPad          `json:"scratch_pads,omitempty"`
	ExportVersion        int64                  `json:"export_version,omitempty"`
}

// CopyFrom copies all attributes except for AccountID, ID, and Name from other
// into this API.  Also, Export is set to the empty string and ExportVersion is set
// to 0
func (a *API) CopyFrom(other *API, copyEmbeddedObjects bool) {
	a.Description = other.Description
	a.CORSAllowOrigin = other.CORSAllowOrigin
	a.CORSAllowHeaders = other.CORSAllowHeaders
	a.CORSAllowCredentials = other.CORSAllowCredentials
	a.CORSRequestHeaders = other.CORSRequestHeaders
	a.CORSMaxAge = other.CORSMaxAge
	a.Export = ""
	a.ExportVersion = 0
	if copyEmbeddedObjects {
		a.Environments = other.Environments
		a.EndpointGroups = other.EndpointGroups
		a.Libraries = other.Libraries
		a.RemoteEndpoints = other.RemoteEndpoints
		a.ProxyEndpoints = other.ProxyEndpoints
	}
}

// Normalize normalizes an API by zeroing out all references to related objects
// and eliminates references to exports
func (a *API) Normalize() {
	a.Export = ""
	a.Environments = []*Environment{}
	a.EndpointGroups = []*EndpointGroup{}
	a.Libraries = []*Library{}
	a.RemoteEndpoints = []*RemoteEndpoint{}
	a.ProxyEndpoints = []*ProxyEndpoint{}
	a.ExportVersion = 0
}

// Validate validates the model.
func (a *API) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if a.Name == "" {
		errors.Add("name", "must not be blank")
	}
	if a.CORSAllowOrigin == "" {
		errors.Add("cors_allow_origin", "must not be blank (use '*' for everything)")
	}
	if a.CORSAllowHeaders == "" {
		errors.Add("cors_allow_headers", "must not be blank (use '*' for everything)")
	}
	if a.CORSRequestHeaders == "" {
		errors.Add("cors_request_headers", "must not be blank (use '*' for everything)")
	}

	for _, re := range a.RemoteEndpoints {
		logreport.Printf("Validating remote endpoints")
		if err := re.Validate(isInsert); !err.Empty() {
			logreport.Printf("Validation not ok!")
			if base, ok := err["base"]; ok {
				errors.Add("base", fmt.Sprintf("associated remote endpoint is invalid -- %v", base))
			} else {
				errors.Add("base", fmt.Sprintf("associated remote endpoint is invalid -- %v", err))
			}
		} else {
			logreport.Printf("Validation ok!")
		}
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (a *API) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "apis", "account_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

// AllAPIsForAccountID returns all apis on the Account in default order.
func AllAPIsForAccountID(db *apsql.DB, accountID int64) ([]*API, error) {
	apis := []*API{}
	err := db.Select(&apis, db.SQL("apis/all"), accountID)
	return apis, err
}

// AllAPIs returns all apis.
func AllAPIs(db *apsql.DB) ([]*API, error) {
	apis := []*API{}
	err := db.Select(&apis, db.SQL("apis/all_apis"))
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

func FindAPIForAccountIDForSwagger(db *apsql.DB, id, accountID int64) (*API, error) {
	api, err := FindAPIForAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Finding API", err)
	}

	api.ProxyEndpoints, err = AllProxyEndpointsForAPIIDAndAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching proxy endpoints", err)
	}
	for index, endpoint := range api.ProxyEndpoints {
		api.ProxyEndpoints[index], err = FindProxyEndpointForAPIIDAndAccountID(db, endpoint.ID, id, accountID)
		if err != nil {
			return nil, aperrors.NewWrapped("Fetching proxy endpoint", err)
		}
	}

	api.ProxyEndpointSchemas, err = AllProxyEndpointSchemasForAPIIDAndAccountID(db, id, accountID)
	if err != nil {
		return nil, aperrors.NewWrapped("Fetching proxy endpoint schemas", err)
	}

	return api, nil
}

// DeleteAPIForAccountID deletes the api with the id and account_id specified.
func DeleteAPIForAccountID(tx *apsql.Tx, id, accountID, userID int64) error {
	err := tx.DeleteOne(tx.SQL("apis/delete"), id, accountID)
	if err != nil {
		return err
	}
	return tx.Notify("apis", accountID, userID, id, 0, id, apsql.Delete)
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

	if a.Export != "" {
		tx.PushTag(apsql.NotificationTagImport)
		defer tx.PopTag()
	}
	a.ID, err = tx.InsertOne(tx.SQL("apis/insert"),
		a.AccountID, a.Name, a.Description, a.CORSAllowOrigin, a.CORSAllowHeaders,
		a.CORSAllowCredentials, a.CORSRequestHeaders, a.CORSMaxAge, a.EnableSwagger)
	if err != nil {
		return
	}

	err = tx.Notify("apis", a.AccountID, a.UserID, a.ID, 0, a.ID, apsql.Insert)
	return
}

// Update updates the api in the database.
func (a *API) Update(tx *apsql.Tx) error {
	err := tx.UpdateOne(tx.SQL("apis/update"),
		a.Name, a.Description, a.CORSAllowOrigin, a.CORSAllowHeaders,
		a.CORSAllowCredentials, a.CORSRequestHeaders, a.CORSMaxAge, a.EnableSwagger,
		a.ID, a.AccountID)
	if err != nil {
		return err
	}

	return tx.Notify("apis", a.AccountID, a.UserID, a.ID, 0, a.ID, apsql.Update)
}
