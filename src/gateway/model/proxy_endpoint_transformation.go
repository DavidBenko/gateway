package model

import (
	apsql "gateway/sql"
	"strings"

	"github.com/jmoiron/sqlx/types"
)

const (
	ProxyEndpointTransformationTypeJS = "js"
)

/* TODO: Add validation of Type */

// ProxyEndpointTransformation describes a transformation around a proxy call.
type ProxyEndpointTransformation struct {
	ID          int64          `json:"id"`
	ComponentID *int64         `json:"-" db:"component_id"`
	CallID      *int64         `json:"-" db:"call_id"`
	Before      bool           `json:"-" db:"before"`
	Type        string         `json:"type"`
	Position    int64          `json:"-"`
	Data        types.JsonText `json:"data,omitempty"`
}

// Validate validates the model.
func (t *ProxyEndpointTransformation) Validate() Errors {
	errors := make(Errors)
	switch t.Type {
	case ProxyEndpointTransformationTypeJS:
	default:
		errors.add("type", "must be 'js'")
	}
	return errors
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
			"component_id IN ("+apsql.NQs(numComponentIDs)+")")
	}
	if numCallIDs > 0 {
		whereClauses = append(whereClauses,
			"call_id IN ("+apsql.NQs(numCallIDs)+")")
	}

	var args []interface{}
	for _, id := range componentIDs {
		args = append(args, id)
	}
	for _, id := range callIDs {
		args = append(args, id)
	}

	err := db.Select(&transformations,
		`SELECT
			id, component_id, call_id, before, type, data
		FROM proxy_endpoint_transformations
		WHERE `+strings.Join(whereClauses, " OR ")+`
		ORDER BY before DESC, position ASC;`,
		args...)
	return transformations, err
}

func DeleteProxyEndpointTransformationsWithComponentIDAndNotInList(tx *apsql.Tx,
	componentID int64, validIDs []int64) error {
	return _deleteProxyEndpointTransformations(tx, "component_id", componentID, validIDs)
}

func DeleteProxyEndpointTransformationsWithCallIDAndNotInList(tx *apsql.Tx,
	callID int64, validIDs []int64) error {
	return _deleteProxyEndpointTransformations(tx, "call_id", callID, validIDs)
}

func _deleteProxyEndpointTransformations(tx *apsql.Tx, ownerCol string,
	ownerID int64, validIDs []int64) error {

	args := []interface{}{ownerID}
	var validIDQuery string
	if len(validIDs) > 0 {
		validIDQuery = " AND id NOT IN (" + apsql.NQs(len(validIDs)) + ")"
		for _, id := range validIDs {
			args = append(args, id)
		}
	}
	_, err := tx.Exec(
		`DELETE FROM proxy_endpoint_transformations
		WHERE `+ownerCol+" = ?"+validIDQuery+";",
		args...)
	return err
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

	data, err := marshaledForStorage(t.Data)
	if err != nil {
		return err
	}
	t.ID, err = tx.InsertOne(
		`INSERT INTO proxy_endpoint_transformations
			(`+ownerCol+`, before, position, type, data)
		VALUES (?, ?, ?, ?, ?)`,
		ownerID, before, position, t.Type, data)

	return err
}

// InsertForComponent inserts the transformation into the database as a new row
// owned by a proxy endpoint component.
func (t *ProxyEndpointTransformation) UpdateForComponent(tx *apsql.Tx,
	componentID int64, before bool, position int) error {
	return t.update(tx, "component_id", componentID, before, position)
}

// InsertForComponent inserts the transformation into the database as a new row
// owned by a proxy endpoint component.
func (t *ProxyEndpointTransformation) UpdateForCall(tx *apsql.Tx,
	callID int64, before bool, position int) error {
	return t.update(tx, "call_id", callID, before, position)
}

// Insert inserts the transformation into the database as a new row.
func (t *ProxyEndpointTransformation) update(tx *apsql.Tx, ownerCol string,
	ownerID int64, before bool, position int) error {

	data, err := marshaledForStorage(t.Data)
	if err != nil {
		return err
	}
	return tx.UpdateOne(
		`UPDATE proxy_endpoint_transformations
		 SET before = ?, position = ?, type = ?, data = ?
		 WHERE id = ? AND `+ownerCol+" = ?",
		before, position, t.Type, data, t.ID, ownerID)
}
