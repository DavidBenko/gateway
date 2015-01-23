package model

import (
	"encoding/json"
	"gateway/config"
	apsql "gateway/sql"
	"log"
	"strings"
)

// ProxyEndpointTransformation describes a transformation around a proxy call.
type ProxyEndpointTransformation struct {
	ID          int64           `json:"id"`
	ComponentID *int64          `json:"-" db:"component_id"`
	CallID      *int64          `json:"-" db:"call_id"`
	Before      bool            `json:"-" db:"before"`
	Type        string          `json:"type"`
	Position    int64           `json:"-"`
	Data        json.RawMessage `json:"data,omitempty"`
}

// AllProxyEndpointTransformationsForComponentIDsAndCallIDs returns all
// transformations for a set of endpoint component.
func AllProxyEndpointTransformationsForComponentIDsAndCallIDs(db *apsql.DB,
	componentIDs, callIDs []int64) ([]*ProxyEndpointTransformation, error) {

	transformations := []*ProxyEndpointTransformation{}

	numComponentIDs := len(componentIDs)
	numCallIDs := len(callIDs)
	if numComponentIDs == 0 && numCallIDs == 0 {
		return transformations, nil
	}

	whereClauses := []string{}
	if numComponentIDs > 0 {
		whereClauses = append(whereClauses,
			"`component_id` IN ("+apsql.NQs(numComponentIDs)+")")
	}
	if numCallIDs > 0 {
		whereClauses = append(whereClauses,
			"`call_id` IN ("+apsql.NQs(numCallIDs)+")")
	}

	var args []interface{}
	for _, id := range componentIDs {
		args = append(args, id)
	}
	for _, id := range callIDs {
		args = append(args, id)
	}

	err := db.Select(&transformations,
		"SELECT "+
			"  `id`, `component_id`, `call_id`, `before`, `type`, `data` "+
			"FROM `proxy_endpoint_transformations` "+
			"WHERE "+strings.Join(whereClauses, " OR ")+" "+
			"ORDER BY `before` DESC, `position` ASC;",
		args...)
	return transformations, err
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
