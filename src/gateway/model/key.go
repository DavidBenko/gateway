package model

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	aperrors "gateway/errors"
	apsql "gateway/sql"
	"strings"
	"time"
)

type Key struct {
	ID        int64  `json:"id,omitempty" path:"id"`
	AccountID int64  `json:"-" db:"account_id"`
	Name      string `json:"name" db:"name"`
	Key       []byte `json:"-"`
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
	if apsql.IsUniqueConstraint(err, "keys", "account_id", "name") {
		errors.Add("name", "is already taken")
	}
	return errors
}

func FindKeysForAccount(accountID int64, db *apsql.DB) ([]*Key, error) {
	keys := []*Key{}
	err := db.Select(&keys, db.SQL("keys/find_all_for_account"), accountID)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func FindKeyByAccountIdAndName(accountID int64, name string, db *apsql.DB) (*Key, error) {
	key := Key{}
	if err := db.Get(&key, db.SQL("keys/find_by_account_name"), name, accountID); err != nil {
		return nil, err
	}
	return &key, nil
}

func (k *Key) Insert(accountID, userID, apiID int64, tx *apsql.Tx) (err error) {
	if k.ID, err = tx.InsertOne(tx.SQL("keys/insert"), k.Name, k.Key, k.AccountID, time.Now().UTC()); err != nil {
		return
	}
	err = afterKeyInsert(k, accountID, userID, apiID, tx)
	return
}

func (k *Key) Delete(accountID, userID, apiID int64, tx *apsql.Tx) (err error) {
	if err = tx.DeleteOne(tx.SQL("keys/delete"), k.ID, accountID); err != nil {
		return
	}
	err = afterKeyDelete(k, accountID, userID, apiID, tx)
	return
}

func afterKeyInsert(key *Key, accountID, userID, apiID int64, tx *apsql.Tx) error {
	return tx.Notify("keys", accountID, userID, apiID, 0, key.ID, apsql.Insert, key.ID, key.Name)
}

func afterKeyDelete(key *Key, accountID, userID, apiID int64, tx *apsql.Tx) error {
	return tx.Notify("keys", accountID, userID, apiID, 0, key.ID, apsql.Delete, key.ID)
}

func (k *Key) GetParsedKey() (interface{}, error) {
	block, err := parsePem(k.Key)
	if err != nil {
		return nil, err
	}
	key, _, err := ParseToKey(block)
	return key, err
}

// data should be a single pem block, i.e. from the opening
// -----BEGIN TYPE----- to the closing -----END TYPE-----
func parsePem(data []byte) (*pem.Block, error) {
	block, remainder := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid format")
	}
	if len(remainder) > 0 {
		return nil, errors.New("additional information in pem, file should contain a single public or private key")
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
