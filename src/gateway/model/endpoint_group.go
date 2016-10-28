package model

import (
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

// EndpointGroup is an optional grouping of proxy endpoints.
type EndpointGroup struct {
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id"`

	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate validates the model.
func (e *EndpointGroup) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if e.Name == "" {
		errors.Add("name", "must not be blank")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *EndpointGroup) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "endpoint_groups", "api_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

// AllEndpointGroupsForAPIIDAndAccountID returns all endpointGroups on the Account's API in default order.
func AllEndpointGroupsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*EndpointGroup, error) {
	endpointGroups := []*EndpointGroup{}
	var err error
	if apiID > 0 && accountID > 0 {
		err = db.Select(&endpointGroups, db.SQL("endpoint_groups/all"), apiID, accountID)
	} else if accountID > 0 {
		err = db.Select(&endpointGroups, db.SQL("endpoint_groups/all_account"), accountID)
	} else {
		return nil, errors.New("Not enough information for endpoint group all.")
	}
	return endpointGroups, err
}

// FindEndpointGroupForAPIIDAndAccountID returns the endpointGroup with the id, api id, and account_id specified.
func FindEndpointGroupForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*EndpointGroup, error) {
	endpointGroup := EndpointGroup{}
	err := db.Get(&endpointGroup, db.SQL("endpoint_groups/find"), id, apiID, accountID)
	return &endpointGroup, err
}

// DeleteEndpointGroupForAPIIDAndAccountID deletes the endpointGroup with the id, api_id and account_id specified.
func DeleteEndpointGroupForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID, userID int64) error {
	err := tx.DeleteOne(tx.SQL("endpoint_groups/delete"), id, apiID, accountID)
	if err != nil {
		return err
	}
	return tx.Notify("endpoint_groups", accountID, userID, apiID, 0, id, apsql.Delete)
}

// Insert inserts the endpointGroup into the database as a new row.
func (e *EndpointGroup) Insert(tx *apsql.Tx) (err error) {
	e.ID, err = tx.InsertOne(tx.SQL("endpoint_groups/insert"),
		e.APIID, e.AccountID, e.Name, e.Description)
	if err != nil {
		return
	}
	err = tx.Notify("endpoint_groups", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Insert)
	return
}

// Update updates the endpointGroup in the database.
func (e *EndpointGroup) Update(tx *apsql.Tx) error {
	err := tx.UpdateOne(tx.SQL("endpoint_groups/update"),
		e.Name, e.Description, e.ID, e.APIID, e.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify("endpoint_groups", e.AccountID, e.UserID, e.APIID, 0, e.ID, apsql.Update)
}
