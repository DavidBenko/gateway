package db

import "gateway/model"

// DB defines the interface of a backing datastore.
type DB interface {
	Router() model.Router
	UpdateRouter(script string) (model.Router, error)

	List(instance model.Model) ([]interface{}, error)
	Insert(instance model.Model) error
	Get(m model.Model, id interface{}) (model.Model, error)
	Find(m model.Model, findByFieldName string, id interface{}) (model.Model, error)
	Update(instance model.Model) error
	Delete(m model.Model, id interface{}) error
}
