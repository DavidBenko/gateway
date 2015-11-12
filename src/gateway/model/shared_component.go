package model

import (
	"fmt"
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

	if s.SharedComponentID != nil {
		errors.Add("shared_component_id", "must not be defined")
	}

	errors.AddErrors(s.ProxyEndpointComponent.Validate())

	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (s *SharedComponent) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "shared_components", "api_id", "name") {
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

// SharedComponentsByIDs takes a slice of IDs of SharedComponents and
// returns a slice of the given SharedComponents with all relationships
// populated.
func SharedComponentsByIDs(
	db *apsql.DB,
	ids []int64,
) ([]*SharedComponent, error) {
	// Fetch the SharedComponents for this set of owner IDs.
	var shared []*SharedComponent

	interfaceIDs := make([]interface{}, len(ids))
	for i, id := range ids {
		interfaceIDs[i] = id
	}

	err := db.Select(
		&shared,
		`
SELECT
  id
  , conditional
  , conditional_positive
  , type
  , data
  , name
  , description
FROM shared_components
WHERE id IN (`[1:]+apsql.NQs(len(ids))+")",
		interfaceIDs...,
	)

	if err != nil {
		return nil, err
	}

	// We should find the same number of SharedComponents as we asked for.
	if len(shared) != len(ids) {
		return nil, fmt.Errorf(
			"tried to find %d SharedComponents but only found %d",
			len(ids), len(shared),
		)
	}

	if err := PopulateSharedComponents(db, shared); err != nil {
		return nil, err
	}

	return shared, nil
}

// PopulateSharedComponents takes a DB handle and a slice of SharedComponents
// and populates all relationships in the given SharedComponents.  This function
// mutates the members of the given slice.
func PopulateSharedComponents(
	db *apsql.DB, components []*SharedComponent,
) error {
	// Populate each ProxyEndpointComponent for the SharedComponents.
	var componentIDs []int64
	componentsByID := make(map[int64]*ProxyEndpointComponent)
	for _, sharedComponent := range components {
		componentIDs = append(componentIDs, sharedComponent.ID)
		componentsByID[sharedComponent.ID] = &sharedComponent.ProxyEndpointComponent
	}

	return PopulateComponents(db, componentIDs, componentsByID)
}

// AllSharedComponentsForAPIID returns all shared components on the API in
// default order.
func AllSharedComponentsForAPIID(
	db *apsql.DB, apiID int64,
) ([]*SharedComponent, error) {
	shared := []*SharedComponent{}
	err := db.Select(&shared, db.SQL("shared_components/all_api"), apiID)

	if err != nil {
		return nil, err
	}

	// Populate each ProxyEndpointComponent for the SharedComponents.
	var componentIDs []int64
	componentsByID := make(map[int64]*ProxyEndpointComponent)
	for _, sharedComponent := range shared {
		componentIDs = append(componentIDs, sharedComponent.ID)
		componentsByID[sharedComponent.ID] = &sharedComponent.ProxyEndpointComponent
	}

	err = PopulateComponents(db, componentIDs, componentsByID)

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

	if err != nil {
		return nil, err
	}

	// Populate each ProxyEndpointComponent for the SharedComponents.
	componentIDs := []int64{shared.ID}
	componentsByID := map[int64]*ProxyEndpointComponent{
		shared.ID: &(shared.ProxyEndpointComponent),
	}

	err = PopulateComponents(db, componentIDs, componentsByID)

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
	// TODO(@binary132): add all the Update calls as in Insert
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}
	err = tx.UpdateOne(
		tx.SQL("shared_components/update"),
		s.Conditional, s.ConditionalPositive, s.Type, data,
		s.APIID, s.Name, s.Description,
		s.ID,
		s.APIID, s.AccountID,
	)
	if err != nil {
		return err
	}
	return tx.Notify(
		"shared_components",
		s.AccountID, s.UserID, s.APIID, s.ID,
		apsql.Update,
	)
}
