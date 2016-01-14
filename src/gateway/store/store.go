package store

import (
	"gateway/config"

	"github.com/boltdb/bolt"
)

const (
	StoreTypeBolt     = "boltdb"
	StoreTypePostgres = "postgres"
)

type Store interface {
	SelectByID(accountID int64, collection string, id uint64) (interface{}, error)
	UpdateByID(accountID int64, collection string, id uint64, object interface{}) (interface{}, error)
	DeleteByID(accountID int64, collection string, id uint64) (interface{}, error)
	Insert(accountID int64, collection string, object interface{}) ([]interface{}, error)
	Delete(accountID int64, collection string, query string, params ...interface{}) ([]interface{}, error)
	Select(accountID int64, collection string, query string, params ...interface{}) ([]interface{}, error)
	Shutdown()
}

func Configure(conf config.Store) (Store, error) {
	if conf.Type == StoreTypeBolt {
		s := BoltDBStore{conf: conf}
		var err error
		s.boltdb, err = bolt.Open(conf.ConnectionString, 0600, nil)
		if err != nil {
			return nil, err
		}
		return &s, nil
	}

	return nil, nil
}
