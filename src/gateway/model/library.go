package model

import (
	"encoding/json"
	"strconv"
	"strings"
)

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

// Valid identifies whether or not the instance can be persisted.
func (l Library) Valid() (bool, error) {
	return true, nil
}

// Less is a case insensitive comparison on Name.
func (l Library) Less(other Model) bool {
	l2 := other.(Library)
	return strings.ToUpper(l.Name) < strings.ToUpper(l2.Name)
}

// MarshalToJSON returns a JSON representation of an instance or slice.
func (l Library) MarshalToJSON(data interface{}) ([]byte, error) {
	switch dataType := data.(type) {
	case Library:
		data = &libraryInstanceWrapper{dataType}
	case []Model:
		libraries := make([]Library, len(dataType))
		for i, v := range dataType {
			libraries[i] = v.(Library)
		}
		data = &libraryCollectionWrapper{libraries}
	}
	return json.MarshalIndent(data, "", "    ")
}

// UnmarshalFromJSON returns an instance created from the passed JSON.
func (l Library) UnmarshalFromJSON(data []byte) (Model, error) {
	instance := Library{}
	err := json.Unmarshal(data, &instance)
	return instance, err
}

// UnmarshalFromJSONWithID returns an instance created from the passed JSON,
// with its ID overridden.
func (l Library) UnmarshalFromJSONWithID(data []byte, id interface{}) (Model, error) {
	wrapper := libraryInstanceWrapper{}
	err := json.Unmarshal(data, &wrapper)
	wrapper.Library.IDField, err = strconv.ParseInt(id.(string), 10, 64)
	return wrapper.Library, err
}

type libraryCollectionWrapper struct {
	Libraries []Library `json:"libraries"`
}

type libraryInstanceWrapper struct {
	Library Library `json:"library"`
}
