package model

// Model is the base struct for all models that clients can manage.
type Model interface {
	ID() interface{}
	CollectionName() string
}
