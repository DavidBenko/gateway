package model

import (
	"errors"
	"fmt"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

const (
	ProxyEndpointComponentTypeSingle = "single"
	ProxyEndpointComponentTypeMulti  = "multi"
	ProxyEndpointComponentTypeJS     = "js"

	// TypeDiscStandard indicates an ordinary ProxyEndpointComponent.
	TypeDiscStandard = "standard"

	// TypeDiscShared indicates a ProxyEndpointComponent with a synthetic
	// reference to a SharedComponent.
	TypeDiscShared = "shared"
)

// ProxyEndpointComponent represents a step of a ProxyEndpoint's workflow.  It
// may inherit and override a SharedComponent.
type ProxyEndpointComponent struct {
	// ID is the ID of the entity in the proxy_endpoint_components table.
	// It is represented in the API as proxy_endpoint_component_id.
	ID int64 `json:"proxy_endpoint_component_id,omitempty" db:"id"`

	// ProxyEndpointComponentReferenceID is the ID of the entity in the
	// proxy_endpoint_component_references table.  SharedComponents do not
	// populate this field in the SharedComponent API, so it is nullable.
	ProxyEndpointComponentReferenceID *int64 `json:"proxy_endpoint_component_reference_id,omitempty" db:"proxy_endpoint_component_reference_id"`

	Type                  string                         `json:"type"`
	Conditional           string                         `json:"conditional"`
	ConditionalPositive   bool                           `json:"conditional_positive" db:"conditional_positive"`
	BeforeTransformations []*ProxyEndpointTransformation `json:"before,omitempty"`
	AfterTransformations  []*ProxyEndpointTransformation `json:"after,omitempty"`
	Call                  *ProxyEndpointCall             `json:"call,omitempty"`
	Calls                 []*ProxyEndpointCall           `json:"calls,omitempty"`
	Data                  types.JsonText                 `json:"data,omitempty"`
	SharedComponentID     *int64                         `json:"shared_component_id,omitempty" db:"-"`

	// TypeDiscriminator can be "standard" or "shared" to indicate whether
	// the component has a synthetic reference to a shared component.
	// Shared components live in the same table as ordinary components, but
	// have different properties.  This value is not expressed in the API.
	TypeDiscriminator string `json:"-" db:"type_discriminator"`
}

// Validate validates the model.
func (c *ProxyEndpointComponent) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)

	if !isInsert && c.ProxyEndpointComponentReferenceID == nil {
		errors.Add(
			"proxy_endpoint_component_reference_id",
			"must not be undefined",
		)
	}
	if c.SharedComponentID != nil {
		return errors
	}

	errors.AddErrors(c.validateType())
	errors.AddErrors(c.validateTransformations())

	return errors
}

func (c *ProxyEndpointComponent) validateType() aperrors.Errors {
	errors := make(aperrors.Errors)

	switch c.Type {
	case ProxyEndpointComponentTypeSingle:
	case ProxyEndpointComponentTypeMulti:
	case ProxyEndpointComponentTypeJS:
	default:
		errors.Add("type", "must be one of 'single', or 'multi', or 'js'")
	}

	return errors
}

func (c *ProxyEndpointComponent) validateTransformations() aperrors.Errors {
	errors := make(aperrors.Errors)

	for i, t := range c.BeforeTransformations {
		tErrors := t.Validate()
		if !tErrors.Empty() {
			errors.Add("before", fmt.Sprintf("%d is invalid: %v", i, tErrors))
		}
	}

	for i, t := range c.AfterTransformations {
		tErrors := t.Validate()
		if !tErrors.Empty() {
			errors.Add("after", fmt.Sprintf("%d is invalid: %v", i, tErrors))
		}
	}

	return errors
}

// AllCalls provides a common interface to iterate through single and multi-call
// components' calls.
func (c *ProxyEndpointComponent) AllCalls() []*ProxyEndpointCall {
	if c.Type == ProxyEndpointComponentTypeSingle {
		return []*ProxyEndpointCall{c.Call}
	}

	return c.Calls
}

// PopulateComponents populates all relationships for a map of IDs to
// ProxyEndpointComponents.  This function mutates the values of the given map.
func PopulateComponents(
	db *apsql.DB,
	componentIDs []int64,
	componentsByID map[int64]*ProxyEndpointComponent,
) error {
	calls, err := AllProxyEndpointCallsForComponentIDs(db, componentIDs)
	if err != nil {
		return err
	}

	var callIDs []int64
	callsByID := make(map[int64]*ProxyEndpointCall)
	for _, call := range calls {
		callIDs = append(callIDs, call.ID)
		callsByID[call.ID] = call
		component := componentsByID[call.ComponentID]
		// Populate the Component's Call or Calls.
		switch component.Type {
		case ProxyEndpointComponentTypeSingle:
			component.Call = call
		case ProxyEndpointComponentTypeMulti:
			component.Calls = append(component.Calls, call)
		}
	}

	transforms, err := TransformationsForComponentIDsAndCallIDs(
		db, componentIDs, callIDs,
	)
	if err != nil {
		return err
	}

	for _, transform := range transforms {
		if transform.ComponentID != nil {
			component := componentsByID[*transform.ComponentID]
			// Populate the Component's Before and After transforms.
			if transform.Before {
				component.BeforeTransformations = append(
					component.BeforeTransformations, transform,
				)
			} else {
				component.AfterTransformations = append(
					component.AfterTransformations, transform,
				)
			}
		} else if transform.CallID != nil {
			call := callsByID[*transform.CallID]
			// Populate the Call's Before and After transforms.
			if transform.Before {
				call.BeforeTransformations = append(
					call.BeforeTransformations, transform,
				)
			} else {
				call.AfterTransformations = append(
					call.AfterTransformations, transform,
				)
			}
		}
	}

	return nil
}

// AllProxyEndpointComponentsForEndpointID returns all components of a
// ProxyEndpoint given its ID.  Only standard (non-shared) Components will have
// their relationships populated.
func AllProxyEndpointComponentsForEndpointID(
	db *apsql.DB, endpointID int64,
) ([]*ProxyEndpointComponent, error) {
	components := []*ProxyEndpointComponent{}
	err := db.Select(
		&components,
		db.SQL("proxy_endpoint_component_references/all_endpoint"),
		endpointID,
	)

	var componentIDs []int64
	componentsByID := make(map[int64]*ProxyEndpointComponent)
	for _, component := range components {
		switch component.TypeDiscriminator {
		case TypeDiscShared:
			sharedID := new(int64)
			*sharedID = component.ID
			// ProxyEndpointComponent.ID is a reference into the
			// proxy_endpoint_components table.
			component.SharedComponentID = sharedID
		case TypeDiscStandard:
			// Populate only components without a Shared reference.
			componentIDs = append(componentIDs, component.ID)
			componentsByID[component.ID] = component
		}
	}

	err = PopulateComponents(db, componentIDs, componentsByID)

	return components, err
}

// DeleteProxyEndpointComponentsWithEndpointIDAndRefNotInSlice deletes
// from proxy_endpoint_component_references with the given endpointID, which do
// not have a ProxyEndpointComponentReferenceID in validIDs.  This also deletes
// the entry from proxy_endpoint_component, iff it is a non-shared component.
func DeleteProxyEndpointComponentsWithEndpointIDAndRefNotInSlice(
	tx *apsql.Tx,
	endpointID int64,
	validIDs []int64,
) error {
	args := []interface{}{endpointID}
	var validIDQuery = ""
	if len(validIDs) > 0 {
		validIDQuery = fmt.Sprintf(" AND id NOT IN (%s)",
			apsql.NQs(len(validIDs)),
		)
		for _, id := range validIDs {
			args = append(args, id)
		}
	}

	_, err := tx.Exec(`
DELETE FROM proxy_endpoint_component_references
  WHERE proxy_endpoint_id = ?
  `[1:]+validIDQuery+`;`,
		args...,
	)

	return err
}

// Insert inserts the component into the database as a new row.
func (c *ProxyEndpointComponent) Insert(
	tx *apsql.Tx,
	endpointID, apiID int64,
	position int,
) error {
	if c.SharedComponentID != nil {
		return c.insertWithShared(tx, endpointID, apiID, position)
	}

	return c.insertWithoutShared(tx, endpointID, apiID, position)
}

func (c *ProxyEndpointComponent) insertWithShared(
	tx *apsql.Tx,
	endpointID, apiID int64,
	position int,
) error {
	crID, err := tx.InsertOne(`
INSERT INTO proxy_endpoint_component_references
  (proxy_endpoint_id
  , proxy_endpoint_component_id
  , position)
VALUES (
  (SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?)
  , (SELECT id FROM proxy_endpoint_components WHERE id = ? AND api_id = ?)
  , ?
)`[1:],
		endpointID, apiID,
		c.SharedComponentID, apiID,
		position,
	)

	if err != nil {
		return aperrors.NewWrapped("Inserting component", err)
	}

	c.ProxyEndpointComponentReferenceID = new(int64)
	*c.ProxyEndpointComponentReferenceID = crID
	c.ID = *c.SharedComponentID

	return nil
}

func (c *ProxyEndpointComponent) insertWithoutShared(
	tx *apsql.Tx,
	endpointID, apiID int64,
	position int,
) error {
	data, err := marshaledForStorage(c.Data)
	if err != nil {
		return aperrors.NewWrapped("Marshaling component JSON", err)
	}

	c.ID, err = tx.InsertOne(`
INSERT INTO proxy_endpoint_components
  (conditional
  , conditional_positive
  , type
  , data
  , type_discriminator)
  VALUES (?, ?, ?, ?, ?)`[1:],
		c.Conditional,
		c.ConditionalPositive,
		c.Type,
		data,
		TypeDiscStandard,
	)

	if err != nil {
		return aperrors.NewWrapped("Inserting component", err)
	}

	crID, err := tx.InsertOne(`
INSERT INTO proxy_endpoint_component_references
  (proxy_endpoint_id
  , proxy_endpoint_component_id
  , position)
  VALUES (
    (SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?)
    , ?, ?
  )`[1:],
		endpointID, apiID,
		c.ID, position,
	)

	if err != nil {
		return aperrors.NewWrapped("Inserting component reference", err)
	}

	c.ProxyEndpointComponentReferenceID = new(int64)
	*c.ProxyEndpointComponentReferenceID = crID

	if err = c.InsertRelationships(tx, apiID); err != nil {
		return aperrors.NewWrapped(
			"Inserting component relationships", err,
		)
	}

	return nil
}

// InsertRelationships performs a database Insert on each relationship item in
// the ProxyEndpointComponent, i.e. Transformations and Calls.
func (c *ProxyEndpointComponent) InsertRelationships(
	tx *apsql.Tx, apiID int64,
) error {
	var err error
	for tPosition, transform := range c.BeforeTransformations {
		err = transform.InsertForComponent(tx, c.ID, true, tPosition)
		if err != nil {
			return aperrors.NewWrapped(
				"Inserting before transformation", err,
			)
		}
	}
	for tPosition, transform := range c.AfterTransformations {
		err = transform.InsertForComponent(tx, c.ID, false, tPosition)
		if err != nil {
			return aperrors.NewWrapped(
				"Inserting after transformation", err,
			)
		}
	}

	switch c.Type {
	case ProxyEndpointComponentTypeSingle:
		if call := c.Call; call != nil {
			if err = call.Insert(tx, c.ID, apiID, 0); err != nil {
				return aperrors.NewWrapped(
					"Inserting single call", err,
				)
			}
		}
	case ProxyEndpointComponentTypeMulti:
		for position, call := range c.Calls {
			err = call.Insert(tx, c.ID, apiID, position)
			if err != nil {
				return aperrors.NewWrapped(
					"Inserting multi call", err,
				)
			}
		}
	}

	return nil
}

// Update updates the component in place.
func (c *ProxyEndpointComponent) Update(
	tx *apsql.Tx,
	endpointID, apiID int64,
	position int,
) error {
	if c.ProxyEndpointComponentReferenceID == nil {
		return errors.New("cannot update ProxyEndpointComponent with nil reference ID")
	}

	if c.SharedComponentID != nil {
		return c.updateWithShared(tx, endpointID, apiID, position)
	}

	return c.updateWithoutShared(tx, endpointID, apiID, position)
}

func (c *ProxyEndpointComponent) updateWithShared(
	tx *apsql.Tx,
	endpointID, apiID int64,
	position int,
) error {
	return tx.UpdateOne(`
UPDATE proxy_endpoint_component_references
  SET position = ?
  , proxy_endpoint_component_id = (
    SELECT id FROM proxy_endpoint_components WHERE id = ? AND api_id = ?
  )
  WHERE proxy_endpoint_id = (
    SELECT id FROM proxy_endpoints WHERE id = ? AND api_id = ?
  ) AND id = ?;`[1:],
		position,
		c.SharedComponentID, apiID,
		endpointID, apiID,
		c.ProxyEndpointComponentReferenceID,
	)
}

func (c *ProxyEndpointComponent) updateWithoutShared(
	tx *apsql.Tx,
	endpointID, apiID int64,
	position int,
) error {
	data, err := marshaledForStorage(c.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(`
UPDATE proxy_endpoint_components
  SET conditional = ?
    , conditional_positive = ?
    , type = ?
    , data = ?
  WHERE id = (
    SELECT cr.proxy_endpoint_component_id
    FROM proxy_endpoint_component_references AS cr,
      proxy_endpoint_components AS pc
    WHERE pc.id = cr.proxy_endpoint_component_id
      AND cr.id = ?
      AND pc.id = ?
  );`[1:],
		c.Conditional,
		c.ConditionalPositive,
		c.Type,
		data,
		c.ProxyEndpointComponentReferenceID,
		c.ID,
	)

	if err != nil {
		return err
	}

	err = tx.UpdateOne(`
UPDATE proxy_endpoint_component_references
  SET position = ?
  WHERE proxy_endpoint_id = ?
  AND id = (
    SELECT cr.id
    FROM proxy_endpoint_component_references AS cr,
      proxy_endpoint_components AS pc
    WHERE pc.id = cr.proxy_endpoint_component_id
      AND cr.id = ?
      AND pc.id = ?
  );`[1:],
		position,
		endpointID,
		c.ProxyEndpointComponentReferenceID,
		c.ID,
	)

	if err != nil {
		return err
	}

	return c.UpdateRelationships(tx, apiID)
}

// UpdateRelationships performs a database Update on each relationship item in
// the ProxyEndpointComponent, i.e. Transformations and Calls.
func (c *ProxyEndpointComponent) UpdateRelationships(
	tx *apsql.Tx, apiID int64,
) error {
	var validTransformationIDs []int64
	var err error

	for position, transformation := range c.BeforeTransformations {
		if transformation.ID == 0 {
			err = transformation.InsertForComponent(
				tx, c.ID, true, position,
			)
		} else {
			err = transformation.UpdateForComponent(
				tx, c.ID, true, position,
			)
		}

		if err != nil {
			return err
		}

		validTransformationIDs = append(
			validTransformationIDs, transformation.ID,
		)
	}

	for position, transformation := range c.AfterTransformations {
		if transformation.ID == 0 {
			err = transformation.InsertForComponent(
				tx, c.ID, true, position,
			)
		} else {
			err = transformation.UpdateForComponent(
				tx, c.ID, true, position,
			)
		}

		if err != nil {
			return err
		}

		validTransformationIDs = append(
			validTransformationIDs, transformation.ID,
		)
	}

	err = DeleteProxyEndpointTransformationsWithComponentIDAndNotInList(
		tx, c.ID, validTransformationIDs,
	)

	if err != nil {
		return err
	}

	var validCallIDs []int64
	switch c.Type {
	case ProxyEndpointComponentTypeSingle:
		if call := c.Call; call != nil {
			if call.ID == 0 {
				err = call.Insert(tx, c.ID, apiID, 0)
			} else {
				err = call.Update(tx, c.ID, apiID, 0)
			}

			if err != nil {
				return err
			}

			validCallIDs = append(validCallIDs, call.ID)
		}
	case ProxyEndpointComponentTypeMulti:
		for position, call := range c.Calls {
			if call.ID == 0 {
				err = call.Insert(tx, c.ID, apiID, position)
			} else {
				err = call.Update(tx, c.ID, apiID, position)
			}

			if err != nil {
				return err
			}

			validCallIDs = append(validCallIDs, call.ID)
		}
	default:
	}

	return DeleteProxyEndpointCallsWithComponentIDAndNotInList(
		tx, c.ID, validCallIDs,
	)
}
