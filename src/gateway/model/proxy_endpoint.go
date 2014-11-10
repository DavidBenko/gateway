package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Endpoint represents an endpoint that the Gateway should handle.
type Endpoint struct {
	IDField int64  `json:"id"`
	Name    string `json:"name" index:"true" unique:"true"`
	Script  string `json:"script"`
}

// CollectionName returns a system-unique name that can be used to reference
// collections of this model, e.g. for URLs or database table names.
func (p *Endpoint) CollectionName() string {
	return "endpoints"
}

// ID returns the id of the model.
func (p *Endpoint) ID() interface{} {
	return p.IDField
}

// EmptyInstance returns an empty instance of the model.
func (p *Endpoint) EmptyInstance() Model {
	return &Endpoint{}
}

// Valid identifies whether or not the instance can be persisted.
func (p *Endpoint) Valid() (bool, error) {
	if p.Name == "" {
		return false, fmt.Errorf("Name must not be blank")
	}

	return true, nil
}

// Less is a case insensitive comparison on Name.
func (p *Endpoint) Less(other Model) bool {
	p2 := other.(*Endpoint)
	return strings.ToUpper(p.Name) < strings.ToUpper(p2.Name)
}

// MarshalToJSON returns a JSON representation of an instance or slice.
func (p *Endpoint) MarshalToJSON(data interface{}) ([]byte, error) {
	switch dataType := data.(type) {
	case *Endpoint:
		data = &endpointInstanceWrapper{dataType}
	case []Model:
		endpoints := make([]*Endpoint, len(dataType))
		for i, v := range dataType {
			endpoints[i] = v.(*Endpoint)
		}
		data = &endpointCollectionWrapper{endpoints}
	}
	return json.MarshalIndent(data, "", "    ")
}

// UnmarshalFromJSON returns an instance created from the passed JSON.
func (p *Endpoint) UnmarshalFromJSON(data []byte) (Model, error) {
	wrapper := endpointInstanceWrapper{}
	err := json.Unmarshal(data, &wrapper)
	return wrapper.Endpoint, err
}

// UnmarshalFromJSONWithID returns an instance created from the passed JSON,
// with its ID overridden.
func (p *Endpoint) UnmarshalFromJSONWithID(data []byte, id interface{}) (Model, error) {
	wrapper := endpointInstanceWrapper{}
	err := json.Unmarshal(data, &wrapper)
	wrapper.Endpoint.IDField = id.(int64)
	return wrapper.Endpoint, err
}

type endpointCollectionWrapper struct {
	Endpoints []*Endpoint `json:"endpoints"`
}

type endpointInstanceWrapper struct {
	Endpoint *Endpoint `json:"endpoint"`
}
