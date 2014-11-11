package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Environment holds configuration values for the proxy.
type Environment struct {
	IDField int64                  `json:"id"`
	Name    string                 `json:"name" index:"true" unique:"true"`
	Values  map[string]interface{} `json:"values"`
}

// CollectionName returns a system-unique name that can be used to reference
// collections of this model, e.g. for URLs or database table names.
func (e *Environment) CollectionName() string {
	return "environments"
}

// ID returns the id of the model.
func (e *Environment) ID() interface{} {
	return e.IDField
}

// EmptyInstance returns an empty instance of the model.
func (e *Environment) EmptyInstance() Model {
	return &Environment{}
}

// Valid identifies whether or not the instance can be persisted.
func (e *Environment) Valid() (bool, error) {
	if e.Name == "" {
		return false, fmt.Errorf("Name must not be blank")
	}

	return true, nil
}

// Less is a case insensitive comparison on Name.
func (e *Environment) Less(other Model) bool {
	l2 := other.(*Environment)
	return strings.ToUpper(e.Name) < strings.ToUpper(l2.Name)
}

// MarshalToJSON returns a JSON representation of an instance or slice.
func (e *Environment) MarshalToJSON(data interface{}) ([]byte, error) {
	switch dataType := data.(type) {
	case *Environment:
		data = &environmentInstanceWrapper{dataType}
	case []Model:
		endpoints := make([]*Environment, len(dataType))
		for i, v := range dataType {
			endpoints[i] = v.(*Environment)
		}
		data = &environmentCollectionWrapper{endpoints}
	}
	return json.MarshalIndent(data, "", "    ")
}

// UnmarshalFromJSON returns an instance created from the passed JSON.
func (e *Environment) UnmarshalFromJSON(data []byte) (Model, error) {
	wrapper := environmentInstanceWrapper{}
	err := json.Unmarshal(data, &wrapper)
	return wrapper.Endpoint, err
}

// UnmarshalFromJSONWithID returns an instance created from the passed JSON,
// with its ID overridden.
func (e *Environment) UnmarshalFromJSONWithID(data []byte, id interface{}) (Model, error) {
	wrapper := environmentInstanceWrapper{}
	err := json.Unmarshal(data, &wrapper)
	wrapper.Endpoint.IDField = id.(int64)
	return wrapper.Endpoint, err
}

type environmentCollectionWrapper struct {
	Endpoints []*Environment `json:"environments"`
}

type environmentInstanceWrapper struct {
	Endpoint *Environment `json:"environment"`
}
