package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

// SharedComponent models a Proxy Endpoint Component that can be defined
// globally for an API and selected for a Proxy Endpoint component.
type SharedComponent struct {
	ProxyEndpointComponent

	// ID overrides ProxyEndpointComponent's ID field.
	ID int64 `json:"id" db:"id"`

	AccountID int64 `json:"-"`
	UserID    int64 `json:"-"`
	APIID     int64 `json:"api_id,omitempty" db:"api_id"`

	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate validates the base ProxyEndpointComponent, name, and ID.  isInsert
// has no effect here, but is required by the controller.
func (s *SharedComponent) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)

	if s.Name == "" {
		errors.Add("name", "must not be blank")
	}

	pec := s.ProxyEndpointComponent

	if pec.SharedComponentID != nil {
		errors.Add("shared_component_id", "must not be defined")
	}

	if pec.ProxyEndpointComponentReferenceID != nil {
		errors.Add("proxy_endpoint_component_reference_id",
			"must not be defined",
		)
	}

	// Validate base Component.
	errors.AddErrors(pec.validateType())
	errors.AddErrors(pec.validateTransformations())
	errors.AddErrors(pec.validateCalls())

	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (s *SharedComponent) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(
		err, "proxy_endpoint_components", "api_id", "name",
	) {
		errors.Add("name", "is already taken")
	}
	return errors
}

func populateSharedComponents(db *apsql.DB, shared []*SharedComponent) error {
	// Populate each ProxyEndpointComponent for the SharedComponents.
	var componentIDs []int64
	componentsByID := make(map[int64]*ProxyEndpointComponent)
	for _, sh := range shared {
		componentIDs = append(componentIDs, sh.ID)
		pec := &(sh.ProxyEndpointComponent)
		pec.ID = sh.ID
		componentsByID[sh.ID] = pec
	}

	return PopulateComponents(db, componentIDs, componentsByID)
}

// AllSharedComponentsForAPI returns a map of all Shared Components for the
// given API ID with all relationships populated, keyed by ID.
func AllSharedComponentsForAPI(db *apsql.DB, apiID int64) (
	map[int64]*SharedComponent, error,
) {
	sharedCs := []*SharedComponent{}

	err := db.Select(
		&sharedCs,
		db.SQL("proxy_endpoint_components/all_shared_api"),
		apiID,
	)

	if err != nil {
		return nil, err
	}

	if err = populateSharedComponents(db, sharedCs); err != nil {
		return nil, err
	}

	mapCs := make(map[int64]*SharedComponent)
	for _, c := range sharedCs {
		mapCs[c.ID] = c
	}

	return mapCs, err
}

// AllSharedComponentsForAPIIDAndAccountID returns all shared components on the
// Account's API in default order, with all relationships populated.
func AllSharedComponentsForAPIIDAndAccountID(
	db *apsql.DB,
	apiID, accountID int64,
) ([]*SharedComponent, error) {
	sharedCs := []*SharedComponent{}

	err := db.Select(
		&sharedCs,
		db.SQL("proxy_endpoint_components/all_shared_api_acc"),
		apiID, accountID,
	)

	if err != nil {
		return nil, err
	}

	err = populateSharedComponents(db, sharedCs)

	return sharedCs, err
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
		db.SQL("proxy_endpoint_components/find_shared"),
		id, apiID, accountID,
	)

	if err != nil {
		return nil, err
	}

	err = populateSharedComponents(db, []*SharedComponent{shared})

	return shared, err
}

// DeleteSharedComponentForAPIIDAndAccountID deletes the shared component with
// the id, api_id, and account_id specified.
func DeleteSharedComponentForAPIIDAndAccountID(
	tx *apsql.Tx,
	id, apiID, accountID, userID int64,
) error {
	err := tx.DeleteOne(
		tx.SQL("proxy_endpoint_components/delete_shared"),
		id, apiID, accountID,
	)

	if err != nil {
		return err
	}

	return tx.Notify(
		"proxy_endpoint_components",
		accountID, userID, apiID, id,
		0, // ProxyEndpointID
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
		tx.SQL("proxy_endpoint_components/insert_shared"),
		s.Conditional, s.ConditionalPositive, s.Type, data,
		s.Name, s.Description,
		s.APIID, s.AccountID,
	)
	if err != nil {
		return aperrors.NewWrapped("Inserting shared component", err)
	}

	pec := s.ProxyEndpointComponent
	pec.ID = s.ID
	if err = pec.InsertRelationships(tx, s.APIID); err != nil {
		return aperrors.NewWrapped("Inserting shared component relationships", err)
	}

	return tx.Notify(
		"proxy_endpoint_components",
		s.AccountID, s.UserID, s.APIID, s.ID,
		0, // ProxyEndpointID
		apsql.Insert,
	)
}

// Update updates the SharedComponent in the database.
func (s *SharedComponent) Update(tx *apsql.Tx) error {
	data, err := marshaledForStorage(s.Data)
	if err != nil {
		return err
	}
	err = tx.UpdateOne(
		tx.SQL("proxy_endpoint_components/update_shared"),
		s.Conditional, s.ConditionalPositive, s.Type, data,
		s.Name, s.Description,
		s.ID,
		s.APIID, s.AccountID,
	)
	if err != nil {
		return err
	}

	pec := s.ProxyEndpointComponent
	pec.ID = s.ID
	if err = pec.UpdateRelationships(tx, s.APIID); err != nil {
		return err
	}

	return tx.Notify(
		"proxy_endpoint_components",
		s.AccountID, s.UserID, s.APIID, s.ID,
		0, // ProxyEndpointID
		apsql.Update,
	)
}
