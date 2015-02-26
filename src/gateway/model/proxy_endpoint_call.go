package model

import (
	"errors"
	apsql "gateway/sql"
)

type ProxyEndpointCall struct {
	ID                    int64                          `json:"id"`
	ComponentID           int64                          `json:"-" db:"component_id"`
	RemoteEndpointID      int64                          `json:"remote_endpoint_id" db:"remote_endpoint_id"`
	EndpointNameOverride  string                         `json:"endpoint_name_override" db:"endpoint_name_override"`
	Conditional           string                         `json:"conditional"`
	ConditionalPositive   bool                           `json:"conditional_positive" db:"conditional_positive"`
	Position              int64                          `json:"-"`
	BeforeTransformations []*ProxyEndpointTransformation `json:"before,omitempty"`
	AfterTransformations  []*ProxyEndpointTransformation `json:"after,omitempty"`

	// RemoteEndpoint is used in proxy to cache remote endpoint data for execution
	RemoteEndpoint *RemoteEndpoint `json:"-"`
}

// AllProxyEndpointCallsForEndpointID returns all calls of a set of endpoint component.
func AllProxyEndpointCallsForComponentIDs(db *apsql.DB, componentIDs []int64) ([]*ProxyEndpointCall, error) {
	calls := []*ProxyEndpointCall{}
	numIDs := len(componentIDs)
	if numIDs == 0 {
		return calls, nil
	}

	var ids []interface{}
	for _, id := range componentIDs {
		ids = append(ids, id)
	}

	err := db.Select(&calls,
		`SELECT
			id, component_id, remote_endpoint_id,
			endpoint_name_override, conditional, conditional_positive
		FROM proxy_endpoint_calls
		WHERE component_id IN (`+apsql.NQs(numIDs)+`)
		ORDER BY position ASC;`,
		ids...)
	return calls, err
}

// DeleteProxyEndpointCallsWithComponentIDAndNotInList
func DeleteProxyEndpointCallsWithComponentIDAndNotInList(tx *apsql.Tx,
	componentID int64, validIDs []int64) error {

	args := []interface{}{componentID}
	var validIDQuery string
	if len(validIDs) > 0 {
		validIDQuery = " AND id NOT IN (" + apsql.NQs(len(validIDs)) + ")"
		for _, id := range validIDs {
			args = append(args, id)
		}
	}
	_, err := tx.Exec(
		`DELETE FROM proxy_endpoint_calls
		WHERE component_id = ?`+validIDQuery+`;`,
		args...)
	return err
}

// Insert inserts the call into the database as a new row.
func (c *ProxyEndpointCall) Insert(tx *apsql.Tx, componentID, apiID int64,
	position int) error {
	var err error
	c.ID, err = tx.InsertOne(
		`INSERT INTO proxy_endpoint_calls
			(component_id, remote_endpoint_id, endpoint_name_override,
			 conditional, conditional_positive, position)
		VALUES (?,(SELECT id FROM remote_endpoints WHERE id = ? AND api_id = ?),
						?, ?, ?, ?)`,
		componentID, c.RemoteEndpointID, apiID, c.EndpointNameOverride,
		c.Conditional, c.ConditionalPositive, position)
	if err != nil {
		return err
	}

	for position, transform := range c.BeforeTransformations {
		err = transform.InsertForCall(tx, c.ID, true, position)
		if err != nil {
			return err
		}
	}
	for position, transform := range c.AfterTransformations {
		err = transform.InsertForCall(tx, c.ID, false, position)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update updates the call into the database in place.
func (c *ProxyEndpointCall) Update(tx *apsql.Tx, componentID, apiID int64,
	position int) error {
	err := tx.UpdateOne(
		`UPDATE proxy_endpoint_calls
		SET
			remote_endpoint_id =
				(SELECT id FROM remote_endpoints WHERE id = ? AND api_id = ?),
			endpoint_name_override = ?,
			conditional = ?,
			conditional_positive = ?,
			position = ?
		WHERE id = ? AND component_id = ?;`,
		c.RemoteEndpointID, apiID, c.EndpointNameOverride,
		c.Conditional, c.ConditionalPositive, position, c.ID, componentID)
	if err != nil {
		return err
	}

	var validTransformationIDs []int64
	for position, transformation := range c.BeforeTransformations {
		if transformation.ID == 0 {
			err = transformation.InsertForCall(tx, c.ID, true, position)
			if err != nil {
				return err
			}
		} else {
			err = transformation.UpdateForCall(tx, c.ID, true, position)
			if err != nil {
				return err
			}
		}
		validTransformationIDs = append(validTransformationIDs, transformation.ID)
	}
	for position, transformation := range c.AfterTransformations {
		if transformation.ID == 0 {
			err = transformation.InsertForCall(tx, c.ID, false, position)
			if err != nil {
				return err
			}
		} else {
			err = transformation.UpdateForCall(tx, c.ID, false, position)
			if err != nil {
				return err
			}
		}
		validTransformationIDs = append(validTransformationIDs, transformation.ID)
	}
	err = DeleteProxyEndpointTransformationsWithCallIDAndNotInList(tx,
		c.ID, validTransformationIDs)
	if err != nil {
		return err
	}

	return nil
}

func (c *ProxyEndpointCall) Name() (string, error) {
	name := c.EndpointNameOverride
	if name == "" {
		if c.RemoteEndpoint == nil {
			return name, errors.New("Remote endpoint is not loaded.")
		}
		name = c.RemoteEndpoint.Codename
	}
	return name, nil
}
