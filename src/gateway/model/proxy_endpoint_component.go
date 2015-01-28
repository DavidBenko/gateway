package model

import (
	"encoding/json"
	apsql "gateway/sql"
)

const (
	ProxyEndpointComponentTypeSingle = "single"
	ProxyEndpointComponentTypeMulti  = "multi"
	ProxyEndpointComponentTypeJS     = "js"
)

type ProxyEndpointComponent struct {
	ID                    int64                          `json:"id"`
	Conditional           string                         `json:"conditional"`
	ConditionalPositive   bool                           `json:"conditional_positive" db:"conditional_positive"`
	Type                  string                         `json:"type"`
	BeforeTransformations []*ProxyEndpointTransformation `json:"before,omitempty"`
	AfterTransformations  []*ProxyEndpointTransformation `json:"after,omitempty"`
	Call                  *ProxyEndpointCall             `json:"call,omitempty"`
	Calls                 []*ProxyEndpointCall           `json:"calls,omitempty"`
	Data                  json.RawMessage                `json:"data,omitempty"`
}

// AllProxyEndpointsForAPIIDAndAccountID returns all components of an endpoint.
func AllProxyEndpointComponentsForEndpointID(db *apsql.DB, endpointID int64) ([]*ProxyEndpointComponent, error) {
	components := []*ProxyEndpointComponent{}
	err := db.Select(&components,
		"SELECT "+
			"  `id`, `conditional`, `conditional_positive`, `type`, `data` "+
			"FROM `proxy_endpoint_components` "+
			"WHERE `endpoint_id` = ? "+
			"ORDER BY `position` ASC;",
		endpointID)
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
		validIDQuery = " AND `id` NOT IN (" + apsql.NQs(len(validIDs)) + ")"
		for _, id := range validIDs {
			args = append(args, id)
		}
	}
	_, err := tx.Exec(
		"DELETE FROM `proxy_endpoint_components` "+
			"WHERE `endpoint_id` = ?"+validIDQuery+";",
		args...)
	return err
}

// Insert inserts the component into the database as a new row.
func (c *ProxyEndpointComponent) Insert(tx *apsql.Tx, endpointID, apiID int64,
	position int) error {

	data, err := c.Data.MarshalJSON()
	if err != nil {
		return err
	}
	c.ID, err = tx.InsertOne(
		"INSERT INTO `proxy_endpoint_components` "+
			"(`endpoint_id`, `conditional`, `conditional_positive`, "+
			" `position`, `type`, `data`) "+
			"VALUES (?, ?, ?, ?, ?, ?);",
		endpointID, c.Conditional, c.ConditionalPositive,
		position, c.Type, string(data))
	if err != nil {
		return err
	}

	for position, transform := range c.BeforeTransformations {
		err = transform.InsertForComponent(tx, c.ID, true, position)
		if err != nil {
			return err
		}
	}
	for position, transform := range c.AfterTransformations {
		err = transform.InsertForComponent(tx, c.ID, false, position)
		if err != nil {
			return err
		}
	}

	switch c.Type {
	case ProxyEndpointComponentTypeSingle:
		err = c.Call.Insert(tx, c.ID, apiID, 0)
		if err != nil {
			return err
		}
	case ProxyEndpointComponentTypeMulti:
		for position, call := range c.Calls {
			err = call.Insert(tx, c.ID, apiID, position)
			if err != nil {
				return err
			}
		}
	default:
	}

	return nil
}

// Update updates the component in place.
func (c *ProxyEndpointComponent) Update(tx *apsql.Tx, endpointID, apiID int64,
	position int) error {

	data, err := c.Data.MarshalJSON()
	if err != nil {
		return err
	}
	err = tx.UpdateOne(
		"UPDATE `proxy_endpoint_components` "+
			"SET `conditional` = ?, "+
			"    `conditional_positive` = ?, "+
			"    `position` = ?, "+
			"    `type` = ?, "+
			"    `data` = ? "+
			"WHERE `id` = ? AND `endpoint_id` = ?;",
		c.Conditional, c.ConditionalPositive,
		position, c.Type, string(data),
		c.ID, endpointID)
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
