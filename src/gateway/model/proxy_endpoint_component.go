package model

import (
	"encoding/json"
	"gateway/config"
	apsql "gateway/sql"
	"log"
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

// Insert inserts the component into the database as a new row.
func (c *ProxyEndpointComponent) Insert(tx *apsql.Tx, endpointID, apiID int64,
	position int) error {

	data, err := c.Data.MarshalJSON()
	if err != nil {
		return err
	}
	result, err := tx.Exec(
		"INSERT INTO `proxy_endpoint_components` "+
			"(`endpoint_id`, `conditional`, `conditional_positive`, "+
			" `position`, `type`, `data`) "+
			"VALUES (?, ?, ?, ?, ?, ?);",
		endpointID, c.Conditional, c.ConditionalPositive,
		position, c.Type, string(data))
	if err != nil {
		return err
	}
	c.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for proxy endpoint component: %v",
			config.System, err)
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
