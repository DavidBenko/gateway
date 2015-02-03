package model

import apsql "gateway/sql"

// EndpointGroup is an optional grouping of proxy endpoints.
type EndpointGroup struct {
	AccountID int64 `json:"-"`
	APIID     int64 `json:"-" db:"api_id"`

	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate validates the model.
func (e *EndpointGroup) Validate() Errors {
	errors := make(Errors)
	if e.Name == "" {
		errors.add("name", "must not be blank")
	}
	return errors
}

// ValidateFromDatabaseError translates possible database constraint errors
// into validation errors.
func (e *EndpointGroup) ValidateFromDatabaseError(err error) Errors {
	errors := make(Errors)
	return errors
}

// AllEndpointGroupsForAPIIDAndAccountID returns all endpointGroups on the Account's API in default order.
func AllEndpointGroupsForAPIIDAndAccountID(db *apsql.DB, apiID, accountID int64) ([]*EndpointGroup, error) {
	endpointGroups := []*EndpointGroup{}
	err := db.Select(&endpointGroups,
		`SELECT
			endpoint_groups.id as id,
			endpoint_groups.name as name,
			endpoint_groups.description as description
		FROM endpoint_groups, apis
		WHERE endpoint_groups.api_id = ?
			AND endpoint_groups.api_id = apis.id
			AND apis.account_id = ?
		ORDER BY endpoint_groups.name ASC;`,
		apiID, accountID)
	return endpointGroups, err
}

// FindEndpointGroupForAPIIDAndAccountID returns the endpointGroup with the id, api id, and account_id specified.
func FindEndpointGroupForAPIIDAndAccountID(db *apsql.DB, id, apiID, accountID int64) (*EndpointGroup, error) {
	endpointGroup := EndpointGroup{}
	err := db.Get(&endpointGroup,
		`SELECT
		  endpoint_groups.id as id,
		  endpoint_groups.name as name,
		  endpoint_groups.description as description
		 FROM endpoint_groups, apis
		 WHERE endpoint_groups.id = ?
			AND endpoint_groups.api_id = ?
			AND endpoint_groups.api_id = apis.id
			AND apis.account_id = ?;`,
		id, apiID, accountID)
	return &endpointGroup, err
}

// DeleteEndpointGroupForAPIIDAndAccountID deletes the endpointGroup with the id, api_id and account_id specified.
func DeleteEndpointGroupForAPIIDAndAccountID(tx *apsql.Tx, id, apiID, accountID int64) error {
	return tx.DeleteOne(
		`DELETE FROM endpoint_groups
		 WHERE endpoint_groups.id = ?
		 AND endpoint_groups.api_id IN
		   (SELECT id FROM apis WHERE id = ? AND account_id = ?);`,
		id, apiID, accountID)
}

// Insert inserts the endpointGroup into the database as a new row.
func (e *EndpointGroup) Insert(tx *apsql.Tx) (err error) {
	e.ID, err = tx.InsertOne(
		`INSERT INTO endpoint_groups (api_id, name, description)
		VALUES ((SELECT id FROM apis WHERE id = ? AND account_id = ?), ?, ?)`,
		e.APIID, e.AccountID, e.Name, e.Description)
	return
}

// Update updates the endpointGroup in the database.
func (e *EndpointGroup) Update(tx *apsql.Tx) error {
	return tx.UpdateOne(
		`UPDATE endpoint_groups
		 SET name = ?, description = ?
		 WHERE endpoint_groups.id = ?
		   AND endpoint_groups.api_id IN
		       (SELECT id FROM apis WHERE id = ? AND account_id = ?)`,
		e.Name, e.Description, e.ID, e.APIID, e.AccountID)
}
