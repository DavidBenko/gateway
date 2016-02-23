package store

import (
	"errors"

	"gateway/config"

	"github.com/boltdb/bolt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	_ "github.com/lib/pq"
)

const (
	StoreTypeBolt     = "boltdb"
	StoreTypePostgres = "postgres"
)

var (
	ErrCollectionExists      = errors.New("collection already exists")
	ErrCollectionDoesntExist = errors.New("collection doesn't exist")
)

type Collection struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id" db:"account_id"`
	Name      string `json:"name"`
}

type Object struct {
	ID           int64          `json:"id"`
	AccountID    int64          `json:"account_id" db:"account_id"`
	CollectionID int64          `json:"collection_id" db:"collection_id"`
	Data         types.JsonText `json:"data"`
}

type StoreAdmin interface {
	ListCollection(collection *Collection, collections *[]*Collection) error
	CreateCollection(collection *Collection) error
	ShowCollection(collection *Collection) error
	UpdateCollection(collection *Collection) error
	DeleteCollection(collection *Collection) error
	ListObject(object *Object, objects *[]*Object) error
	CreateObject(object *Object) error
	ShowObject(object *Object) error
	UpdateObject(object *Object) error
	DeleteObject(object *Object) error
}

type StoreEndpoint interface {
	SelectByID(accountID int64, collection string, id uint64) (interface{}, error)
	UpdateByID(accountID int64, collection string, id uint64, object interface{}) (interface{}, error)
	DeleteByID(accountID int64, collection string, id uint64) (interface{}, error)
	Insert(accountID int64, collection string, object interface{}) ([]interface{}, error)
	Delete(accountID int64, collection string, query string, params ...interface{}) ([]interface{}, error)
	Select(accountID int64, collection string, query string, params ...interface{}) ([]interface{}, error)
}

type Store interface {
	Migrate() error
	Clear() error
	Shutdown()
	StoreAdmin
	StoreEndpoint
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
	} else if conf.Type == StoreTypePostgres {
		p := PostgresStore{conf: conf}
		var err error
		p.db, err = sqlx.Open("postgres", conf.ConnectionString)
		if err != nil {
			return nil, err
		}

		p.db.SetMaxOpenConns(int(conf.MaxConnections))

		return &p, nil
	}

	return nil, nil
}
