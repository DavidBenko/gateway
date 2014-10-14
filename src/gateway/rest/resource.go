package rest

// Resource is a REST-fully accessible resource.
type Resource interface {
	Name() string

	Index() (resources interface{}, err error)
	Create(data interface{}) (resource interface{}, err error)
	Show(id interface{}) (resource interface{}, err error)
	Update(id interface{}, data interface{}) (resource interface{}, err error)
	Delete(id interface{}) error
}
