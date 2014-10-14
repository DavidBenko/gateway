package db

import "github.com/AnyPresence/gateway/model"

// DB defines the interface of a backing datastore.
type DB interface {
	List(instance model.Model) ([]interface{}, error)
	Insert(instance model.Model) error
	Get(m model.Model, id interface{}) (model.Model, error)
	Find(m model.Model, findByFieldName string, id interface{}) (model.Model, error)
	Update(instance model.Model) error
	Delete(m model.Model, id interface{}) error
}

const (
	fieldTagIndexed     = "index"
	fieldTagIndexedTrue = "true"
)
