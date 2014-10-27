package model

const (
	// FieldTagIndexed is the name of the field tag to use to mark it as indexed
	FieldTagIndexed = "index"

	// FieldTagUnique is the name of the field tag to use to mark it as unique
	FieldTagUnique = "unique"
)

// Model is the base struct for all models that clients can manage.
type Model interface {
	ID() interface{}
	CollectionName() string
	EmptyInstance() Model
	Valid() (bool, error)
	UnmarshalFromJSON([]byte) (Model, error)
}
