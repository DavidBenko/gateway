package model

import (
	"encoding/json"
	"gateway/config"
	apsql "gateway/sql"
	"log"
)

// ProxyEndpointTransformation describes a transformation around a proxy call.
type ProxyEndpointTransformation struct {
	ID       int64           `json:"id"`
	Type     string          `json:"type"`
	Position int64           `json:"-"`
	Data     json.RawMessage `json:"data,omitempty"`
}

// InsertForComponent inserts the transformation into the database as a new row
// owned by a proxy endpoint component.
func (t *ProxyEndpointTransformation) InsertForComponent(tx *apsql.Tx,
	componentID int64, before bool, position int) error {
	return t.insert(tx, "component_id", componentID, before, position)
}

// InsertForCall inserts the transformation into the database as a new row
// owned by a proxy endpoint call.
func (t *ProxyEndpointTransformation) InsertForCall(tx *apsql.Tx,
	callID int64, before bool, position int) error {
	return t.insert(tx, "call_id", callID, before, position)
}

// Insert inserts the transformation into the database as a new row.
func (t *ProxyEndpointTransformation) insert(tx *apsql.Tx, ownerCol string,
	ownerID int64, before bool, position int) error {

	data, err := t.Data.MarshalJSON()
	if err != nil {
		return err
	}
	result, err := tx.Exec(
		"INSERT INTO `proxy_endpoint_transformations` "+
			"(`"+ownerCol+"`, `before`, `position`, `type`, `data`) "+
			"VALUES (?, ?, ?, ?, ?);",
		ownerID, before, position, t.Type, string(data))
	if err != nil {
		return err
	}
	t.ID, err = result.LastInsertId()
	if err != nil {
		log.Printf("%s Error getting last insert ID for proxy endpoint tranform: %v",
			config.System, err)
		return err
	}

	return nil
}
