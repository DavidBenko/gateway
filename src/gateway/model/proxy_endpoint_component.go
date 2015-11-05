package model

import (
	"fmt"

	aperrors "gateway/errors"
	apsql "gateway/sql"

	"github.com/jmoiron/sqlx/types"
)

const (
	ProxyEndpointComponentTypeSingle = "single"
	ProxyEndpointComponentTypeMulti  = "multi"
	ProxyEndpointComponentTypeJS     = "js"
)

// ProxyEndpointComponent represents a step of a ProxyEndpoint's workflow.  It
// may inherit and override a SharedComponent.
type ProxyEndpointComponent struct {
	ID                    int64                          `json:"id,omitempty"`
	Type                  string                         `json:"type"`
	Conditional           string                         `json:"conditional"`
	ConditionalPositive   bool                           `json:"conditional_positive" db:"conditional_positive"`
	BeforeTransformations []*ProxyEndpointTransformation `json:"before,omitempty"`
	AfterTransformations  []*ProxyEndpointTransformation `json:"after,omitempty"`
	Call                  *ProxyEndpointCall             `json:"call,omitempty"`
	Calls                 []*ProxyEndpointCall           `json:"calls,omitempty"`
	Data                  types.JsonText                 `json:"data,omitempty"`

	// SharedComponentID is the database ID of the SharedComponent.
	SharedComponentID *int64 `json:"shared_component_id,omitempty" db:"shared_component_id"`

	// SharedComponentHandle will be fetched before Update or Insert.
	SharedComponentHandle *SharedComponent `json:"-"`
}

// Validate validates the model.
//
// This method will panic if a SharedComponent was not prepopulated for a
// component having a non-nil SharedComponentID.
func (c *ProxyEndpointComponent) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)

	switch {
	case c.SharedComponentID != nil:
		errors.AddErrors(c.validateSharedBase())
		errors.AddErrors(c.validateAgainstParent())
	default:
		errors.AddErrors(c.validateBase())
	}

	return errors
}

// validateBase validates a ProxyEndpointComponent which is not inherited from a
// SharedComponent.
func (c *ProxyEndpointComponent) validateBase() aperrors.Errors {
	errors := make(aperrors.Errors)

	switch c.Type {
	case ProxyEndpointComponentTypeSingle:
	case ProxyEndpointComponentTypeMulti:
	case ProxyEndpointComponentTypeJS:
	default:
		errors.Add("type", "must be one of 'single', or 'multi', or 'js'")
	}

	errors.AddErrors(c.validateTransformations())

	return errors
}

// validateSharedBase validates the ProxyEndpointComponent as a base for a
// SharedComponent.
func (c *ProxyEndpointComponent) validateSharedBase() aperrors.Errors {
	errors := make(aperrors.Errors)
	if c.Type != "" {
		errors.Add("type", "must not override SharedComponent's type")
	}

	errors.AddErrors(c.validateTransformations())

	return errors
}

// validateAgainstParent assumes the SharedComponentHandle is fully populated.
func (c *ProxyEndpointComponent) validateAgainstParent() aperrors.Errors {
	switch c.SharedComponentHandle.ProxyEndpointComponent.Type {
	case ProxyEndpointComponentTypeSingle:
		return c.validateSingle()
	case ProxyEndpointComponentTypeMulti:
		return c.validateMulti()
	case ProxyEndpointComponentTypeJS:
		return c.validateJS()
	default:
		return aperrors.Errors{
			"type": []string{"must be one of 'single', or 'multi', or 'js'"},
		}
	}
}

func (c *ProxyEndpointComponent) validateSingle() aperrors.Errors {
	s := ProxyEndpointComponentTypeSingle
	return aperrors.ValidateCases([]aperrors.TestCase{
		{c.Calls == nil, "calls", "type " + s + " must not have multi calls"},
		{c.Data == nil, "data", "type " + s + " must not have js"},
	}...)
}

func (c *ProxyEndpointComponent) validateMulti() aperrors.Errors {
	m := ProxyEndpointComponentTypeMulti
	return aperrors.ValidateCases([]aperrors.TestCase{
		{c.Call == nil, "call", "type " + m + " must not have single call"},
		{c.Data == nil, "data", "type " + m + " must not have js"},
	}...)
}

func (c *ProxyEndpointComponent) validateJS() aperrors.Errors {
	j := ProxyEndpointTransformationTypeJS
	return aperrors.ValidateCases([]aperrors.TestCase{
		{c.Call == nil, "call", "type " + j + " must not have single call"},
		{c.Calls == nil, "calls", "type " + j + " must not have multi calls"},
	}...)
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

// AllProxyEndpointsForAPIIDAndAccountID returns all components of an endpoint.
func AllProxyEndpointComponentsForEndpointID(db *apsql.DB, endpointID int64) ([]*ProxyEndpointComponent, error) {
	components := []*ProxyEndpointComponent{}
	err := db.Select(
		&components,
		`SELECT
			id, conditional, conditional_positive, type, data,
			shared_component_id
		FROM proxy_endpoint_components
		WHERE endpoint_id = ?
		ORDER BY position ASC;`,
		endpointID,
	)
	if err != nil {
		return nil, err
	}

	var componentIDs []int64
	componentsByID := make(map[int64]*ProxyEndpointComponent)
	for _, component := range components {
		componentIDs = append(componentIDs, component.ID)
		componentsByID[component.ID] = component
	}

	calls, err := AllProxyEndpointCallsForComponentIDs(db, componentIDs)
	if err != nil {
		return nil, err
	}

	var callIDs []int64
	callsByID := make(map[int64]*ProxyEndpointCall)
	for _, call := range calls {
		callIDs = append(callIDs, call.ID)
		callsByID[call.ID] = call
		component := componentsByID[call.ComponentID]
		switch component.Type {
		case ProxyEndpointComponentTypeSingle:
			component.Call = call
		case ProxyEndpointComponentTypeMulti:
			component.Calls = append(component.Calls, call)
		}
	}

	transforms, err := AllProxyEndpointTransformationsForComponentIDsAndCallIDs(db,
		componentIDs, callIDs)
	if err != nil {
		return nil, err
	}

	for _, transform := range transforms {
		if transform.ComponentID != nil {
			component := componentsByID[*transform.ComponentID]
			if transform.Before {
				component.BeforeTransformations = append(component.BeforeTransformations, transform)
			} else {
				component.AfterTransformations = append(component.AfterTransformations, transform)
			}
		} else if transform.CallID != nil {
			call := callsByID[*transform.CallID]
			if transform.Before {
				call.BeforeTransformations = append(call.BeforeTransformations, transform)
			} else {
				call.AfterTransformations = append(call.AfterTransformations, transform)
			}
		}
	}

	return components, err
}

// DeleteProxyEndpointComponentsWithEndpointIDAndNotInList
func DeleteProxyEndpointComponentsWithEndpointIDAndNotInList(tx *apsql.Tx,
	endpointID int64, validIDs []int64) error {

	args := []interface{}{endpointID}
	var validIDQuery string
	if len(validIDs) > 0 {
		validIDQuery = " AND id NOT IN (" + apsql.NQs(len(validIDs)) + ")"
		for _, id := range validIDs {
			args = append(args, id)
		}
	}
	_, err := tx.Exec(
		`DELETE FROM proxy_endpoint_components
		WHERE endpoint_id = ?`+validIDQuery+`;`,
		args...)
	return err
}

// Insert inserts the component into the database as a new row.
func (c *ProxyEndpointComponent) Insert(
	tx *apsql.Tx,
	endpointID, apiID int64,
	position int,
) error {
	data, err := marshaledForStorage(c.Data)
	if err != nil {
		return aperrors.NewWrapped("Marshaling component JSON", err)
	}

	c.ID, err = tx.InsertOne(
		`INSERT INTO proxy_endpoint_components
			(endpoint_id, conditional, conditional_positive,
			 position, type, data, shared_component_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		endpointID, c.Conditional, c.ConditionalPositive,
		position, c.Type, data, c.SharedComponentID)
	if err != nil {
		return aperrors.NewWrapped("Inserting component", err)
	}

	for tPosition, transform := range c.BeforeTransformations {
		err = transform.InsertForComponent(tx, c.ID, true, tPosition)
		if err != nil {
			return aperrors.NewWrapped("Inserting before transformation", err)
		}
	}
	for tPosition, transform := range c.AfterTransformations {
		err = transform.InsertForComponent(tx, c.ID, false, tPosition)
		if err != nil {
			return aperrors.NewWrapped("Inserting after transformation", err)
		}
	}

	switch c.Type {
	case ProxyEndpointComponentTypeSingle:
		if err = c.Call.Insert(tx, c.ID, apiID, 0); err != nil {
			return aperrors.NewWrapped("Inserting single call", err)
		}
	case ProxyEndpointComponentTypeMulti:
		for position, call := range c.Calls {
			if err = call.Insert(tx, c.ID, apiID, position); err != nil {
				return aperrors.NewWrapped("Inserting multi call", err)
			}
		}
	default:
	}

	return nil
}

// Update updates the component in place.
func (c *ProxyEndpointComponent) Update(tx *apsql.Tx, endpointID, apiID int64,
	position int) error {
	data, err := marshaledForStorage(c.Data)
	if err != nil {
		return err
	}

	err = tx.UpdateOne(
		`UPDATE proxy_endpoint_components
		SET
			conditional = ?,
			conditional_positive = ?,
			position = ?,
			type = ?,
			data = ?,
			shared_component_id = ?
		WHERE id = ? AND endpoint_id = ?;`,
		c.Conditional,
		c.ConditionalPositive,
		position,
		c.Type,
		data,
		c.SharedComponentID,
		c.ID,
		endpointID,
	)
	if err != nil {
		return err
	}

	var validTransformationIDs []int64
	for position, transformation := range c.BeforeTransformations {
		if transformation.ID == 0 {
			err = transformation.InsertForComponent(tx, c.ID, true, position)
			if err != nil {
				return err
			}
		} else {
			err = transformation.UpdateForComponent(tx, c.ID, true, position)
			if err != nil {
				return err
			}
		}
		validTransformationIDs = append(validTransformationIDs, transformation.ID)
	}
	for position, transformation := range c.AfterTransformations {
		if transformation.ID == 0 {
			err = transformation.InsertForComponent(tx, c.ID, false, position)
			if err != nil {
				return err
			}
		} else {
			err = transformation.UpdateForComponent(tx, c.ID, false, position)
			if err != nil {
				return err
			}
		}
		validTransformationIDs = append(validTransformationIDs, transformation.ID)
	}
	err = DeleteProxyEndpointTransformationsWithComponentIDAndNotInList(tx,
		c.ID, validTransformationIDs)
	if err != nil {
		return err
	}

	var validCallIDs []int64
	switch c.Type {
	case ProxyEndpointComponentTypeSingle:
		if c.Call.ID == 0 {
			err = c.Call.Insert(tx, c.ID, apiID, 0)
			if err != nil {
				return err
			}
		} else {
			err = c.Call.Update(tx, c.ID, apiID, 0)
			if err != nil {
				return err
			}
		}
		validCallIDs = append(validCallIDs, c.Call.ID)
	case ProxyEndpointComponentTypeMulti:
		for position, call := range c.Calls {
			if call.ID == 0 {
				err = call.Insert(tx, c.ID, apiID, position)
				if err != nil {
					return err
				}
			} else {
				err = call.Update(tx, c.ID, apiID, position)
				if err != nil {
					return err
				}
			}
			validCallIDs = append(validCallIDs, call.ID)
		}
	default:
	}

	err = DeleteProxyEndpointCallsWithComponentIDAndNotInList(tx,
		c.ID, validCallIDs)
	if err != nil {
		return err
	}

	return nil
}
