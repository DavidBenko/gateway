package model

import (
	aperrors "gateway/errors"
	apsql "gateway/sql"
)

type ProxyEndpointTestPair struct {
  ID     int64  `json:"id,omitempty"`
	Key    string `json:"key"`
  Value  string `json:"value"`
}

func (p *ProxyEndpointTestPair) Insert(tx *apsql.Tx, testID int64) error {
  var err error
  p.ID, err = tx.InsertOne(tx.SQL("pairs/insert"), testID, p.Key, p.Value)
  if err != nil {
    return aperrors.NewWrapped("Inserting pair", err)
  }

  return nil
}

func (p *ProxyEndpointTestPair) Update(tx *apsql.Tx, testID int64) error {
  err := tx.UpdateOne(tx.SQL("pairs/update"), p.Key, p.Value, p.ID, testID)
  if err != nil {
    return aperrors.NewWrapped("Updating pair", err)
  }

  return nil
}

func AllProxyEndpointTestPairsForTestID(db *apsql.DB, testID int64) ([]*ProxyEndpointTestPair, error) {
  pairs := []*ProxyEndpointTestPair{}
  err := db.Select(&pairs, db.SQL("pairs/all"), testID)
  if err != nil {
    return nil, err
  }

  return pairs, err
}

func DeleteProxyEndpointTestPairsWithTestIDAndNotInList(tx *apsql.Tx,
	testID int64, validIDs []int64) error {

	args := []interface{}{testID}
	var validIDQuery string
	if len(validIDs) > 0 {
		validIDQuery = " AND id NOT IN (" + apsql.NQs(len(validIDs)) + ")"
		for _, id := range validIDs {
			args = append(args, id)
		}
	}
	_, err := tx.Exec(
		`DELETE FROM proxy_endpoint_test_pairs
		WHERE test_id = ?`+validIDQuery+`;`,
		args...)
	return err
}
