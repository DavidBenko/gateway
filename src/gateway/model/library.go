package model

import "encoding/json"

// Library holds reusable scripts for the proxy.
type Library struct {
	IDField int64  `json:"id"`
	Name    string `json:"name"`
	Script  string `json:"script"`
}

// CollectionName returns a system-unique name that can be used to reference
// collections of this model, e.g. for URLs or database table names.
func (l Library) CollectionName() string {
	return "libraries"
}

// ID returns the id of the model.
func (l Library) ID() interface{} {
	return l.IDField
}

// EmptyInstance returns an empty instance of the model.
func (l Library) EmptyInstance() Model {
	return Library{}
}

// UnmarshalFromJSON returns an instance created from the passed JSON.
func (l Library) UnmarshalFromJSON(data []byte) (Model, error) {
	instance := Library{}
	err := json.Unmarshal(data, &instance)
	return instance, err
}
