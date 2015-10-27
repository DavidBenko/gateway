package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

// SharedComponent models a Proxy Endpoint Component that can be defined
// globally for an API and selected for a Proxy Endpoint component.
type SharedComponent struct {
	ProxyEndpointComponent
	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id"`

	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate validates the modes.
func (s *SharedComponent) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	if s.Name == "" {
		errors.Add("name", "must not be blank")
	}

	if s.SharedComponentID != 0 {
		errors.Add("shared_component_id", "must not be defined")
	}

	errors.AddErrors(s.ProxyEndpointComponent.Validate())

	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (s *SharedComponent) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if err.Error() == "UNIQUE constraint failed: shared_components.api_id, shared_components.name" ||
		err.Error() == `pq: duplicate key value violates unique constraint "shared_components_api_id_name_key"` {
		errors.Add("name", "is already taken")
	}
	return errors
}

// AllSharedComponentsForAPIIDAndAccountID returns all shared components on the
// Account's API in default order.
func AllSharedComponentsForAPIIDAndAccountID(
	db *apsql.DB,
	apiID, accountID int64,
) ([]*SharedComponent, error) {
	shared := []*SharedComponent{}

	err := db.Select(
		&shared,
		db.SQL("shared_components/all"),
		apiID, accountID,
	)

	return shared, err
}

// AllSharedComponentsForProxy returns all shared components on the API in
// default order.
func AllSharedComponentsForAPIID(
	db *apsql.DB, apiID int64,
) ([]*SharedComponent, error) {
	shared := []*SharedComponent{}
	err := db.Select(&shared, db.SQL("shared_components/all_api"), apiID)
	return shared, err
}

// FindSharedComponentForAPIIDAndAccountID returns the shared component with the
// id, api id, and account_id specified.
func FindSharedComponentForAPIIDAndAccountID(
	db *apsql.DB,
	id, apiID, accountID int64,
) (*SharedComponent, error) {
	shared := &SharedComponent{}
	err := db.Get(
		shared,
		db.SQL("shared_components/find"),
		id, apiID, accountID,
	)
	return shared, err
}

// DeleteSharedComponentForAPIIDAndAccountID deletes the shared component with
// the id, api_id and account_id specified.
func DeleteSharedComponentForAPIIDAndAccountID(
	tx *apsql.Tx,
	id, apiID, accountID, userID int64,
) error {
	err := tx.DeleteOne(
		tx.SQL("shared_components/delete"),
		id, apiID, accountID,
	)
	if err != nil {
		return err
	}
	return tx.Notify(
		"shared_components",
		accountID, userID, apiID, id,
		apsql.Delete,
	)
}

// Insert inserts the shared component into the database as a new row.
func (s *SharedComponent) Insert(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return aperrors.NewWrapped(
			"Marshaling shared component JSON", err,
		)
	}

	s.ID, err = tx.InsertOne(
		tx.SQL("shared_components/insert"),
		s.Conditional, s.ConditionalPositive, s.Type, data,
		s.APIID, s.AccountID, s.Name, s.Description,
	)
	if err != nil {
		return aperrors.NewWrapped("Inserting shared component", err)
	}

	for tPosition, transform := range s.BeforeTransformations {
		err = transform.InsertForComponent(tx, s.ID, true, tPosition)
		if err != nil {
			return aperrors.NewWrapped(
				"Inserting before transformation", err,
			)
		}
	}

	for tPosition, transform := range s.AfterTransformations {
		err = transform.InsertForComponent(tx, s.ID, false, tPosition)
		if err != nil {
			return aperrors.NewWrapped(
				"Inserting after transformation", err,
			)
		}
	}

	switch s.Type {
	case ProxyEndpointComponentTypeSingle:
		if err = s.Call.Insert(tx, s.ID, s.APIID, 0); err != nil {
			return aperrors.NewWrapped("Inserting single call", err)
		}
	case ProxyEndpointComponentTypeMulti:
		for callPosition, call := range s.Calls {
			err = call.Insert(tx, s.ID, s.APIID, callPosition)
			if err != nil {
				return aperrors.NewWrapped(
					"Inserting multi call", err,
				)
			}
		}
	default:
	}

	return tx.Notify(
		"shared_components",
		s.AccountID, s.UserID, s.APIID, s.ID,
		apsql.Insert,
	)
}

// Update updates the library in the databass.
func (s *SharedComponent) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}
	err = tx.UpdateOne(tx.SQL("libraries/update"),
		s.Name, s.Description, data, s.ID, s.APIID, s.AccountID)
	if err != nil {
		return err
	}
	return tx.Notify(
		"shared_components",
		s.AccountID, s.UserID, s.APIID, s.ID,
		apsql.Update,
	)
}
