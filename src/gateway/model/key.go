package model

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"
	"strings"
)

type Key struct {
	ID   int64  `json:"id,omitempty" path:"id"`
	Name string `json:"name" db:"name"`
	Key  []byte `json:"-"`
}

func (k *Key) Validate(isInsert bool) aperrors.Errors {
	errors := make(aperrors.Errors)
	if k.Name == "" || strings.TrimSpace(k.Name) == "" {
		errors.Add("name", "must not be blank")
	}

	block, err := parsePem(k.Key)
	if err != nil {
		errors.Add("key", err.Error())
		return errors
	}

	_, _, err = ParseToKey(block)
	if err != nil {
		errors.Add("key", err.Error())
		return errors
	}
	return errors
}

func (k *Key) ValidateFromDatabaseError(err error) aperrors.Errors {
	errors := make(aperrors.Errors)
	if apsql.IsUniqueConstraint(err, "keys", "api_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func FindKeys(db *apsql.DB) ([]*Key, error) {
	keys := []*Key{}
	err := db.Select(&keys, db.SQL("keys/find_all"))
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (k *Key) Insert(accountID, userID, apiID int64, tx *apsql.Tx) (err error) {
	if k.ID, err = tx.InsertOne(tx.SQL("keys/insert"), k.Name, k.Key); err != nil {
		return
	}
	err = afterInsert(k, accountID, userID, apiID, tx)
	return
}

func (k *Key) Delete(tx *apsql.Tx) (err error) {
	err = tx.DeleteOne(tx.SQL("keys/delete"), k.ID)
	return
}

func afterInsert(key *Key, accountID, userID, apiID int64, tx *apsql.Tx) error {
	return tx.Notify("keys", accountID, userID, apiID, 0, key.ID, apsql.Insert, key.ID)
}

// data should be a single pem block, i.e. from the opening
// -----BEGIN TYPE----- to the closing -----END TYPE-----
func parsePem(data []byte) (*pem.Block, error) {
	block, remainder := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid format")
	}
	if len(remainder) > 0 {
		return nil, errors.New("additional information in pem")
	}
	return block, nil
}

func parsePrivateKey(block *pem.Block) (interface{}, bool) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		k, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, false
		}
		return k, true
	case "EC PRIVATE KEY":
		k, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, false
		}
		return k, true
	default:
		return nil, false
	}
}

func parsePublicKey(block *pem.Block) (interface{}, bool) {
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, false
	}
	return key, true
}

func ParseToKey(block *pem.Block) (interface{}, bool, error) {
	if key, ok := parsePrivateKey(block); ok {
		return key, false, nil
	}

	if key, ok := parsePublicKey(block); ok {
		return key, true, nil
	}

	return nil, false, errors.New("invalid or unsupported key type")
}
