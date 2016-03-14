package store

import (
	"encoding/json"
	"errors"
	"time"

	"gateway/config"
	aperrors "gateway/errors"
	"gateway/logreport"
	apsql "gateway/sql"

	"github.com/boltdb/bolt"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
)

//go:generate peg jql.peg

const (
	StoreTypeBolt     = "boltdb"
	StoreTypePostgres = "postgres"
)

var (
	ErrCollectionExists      = errors.New("store collection already exists")
	ErrCollectionDoesntExist = errors.New("store collection doesn't exist")
)

type Collection struct {
	UserID int64 `json:"-"`

	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id,omitempty" db:"account_id"`
	Name      string `json:"name"`
}

type Object struct {
	UserID int64 `json:"-"`

	ID           int64          `json:"id"`
	AccountID    int64          `json:"account_id,omitempty" db:"account_id"`
	CollectionID int64          `json:"store_collection_id" db:"collection_id"`
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
	RegisterListener(l apsql.Listener)
	Shutdown()
	StoreAdmin
	StoreEndpoint
}

func (c *Collection) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)
	if c.Name == "" {
		errors.Add("name", "must not be blank")
	}
	return errors
}

func (o *Object) Validate() aperrors.Errors {
	errors := make(aperrors.Errors)
	var data interface{}
	if err := o.Data.Unmarshal(&data); err != nil {
		errors.Add("data", "must be valid json")
	}
	return errors
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

		listener := pq.NewListener(conf.ConnectionString,
			2*time.Second,
			time.Minute,
			p.listenerConnectionEvent)
		err = listener.Listen(postgresNotifyChannel)
		if err != nil {
			return nil, err
		}
		go func() {
			for {
				select {
				case pgNotification := <-listener.Notify:
					if pgNotification.Channel == postgresNotifyChannel {
						var notification apsql.Notification
						err := json.Unmarshal([]byte(pgNotification.Extra), &notification)
						if err != nil {
							logreport.Printf("%s Error parsing notification '%s': %v",
								config.System, pgNotification.Extra, err)
							continue
						}
						p.notifyListeners(&notification)
					} else {
						p.notifyListenersOfReconnection()
					}
				case <-time.After(90 * time.Second):
					go func() {
						listener.Ping()
					}()
				}
			}
		}()

		return &p, nil
	}

	return nil, nil
}
