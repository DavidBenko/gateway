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
	UnmarshalFromJSONWithID(data []byte, id interface{}) (Model, error)
	CollectionName() string
	EmptyInstance() Model
	Valid() (bool, error)
	Less(Model) bool
	MarshalToJSON(interface{}) ([]byte, error)
	UnmarshalFromJSON([]byte) (Model, error)
}

// SortableModels is a type that defines the sort interface on []Model
type SortableModels []Model

// Len is the number of elements in the collection.
func (s SortableModels) Len() int {
	return len(s)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s SortableModels) Less(i, j int) bool {
	return s[i].Less(s[j])
}

// Swap swaps the elements with indexes i and j.
func (s SortableModels) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
