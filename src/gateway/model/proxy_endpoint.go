package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ProxyEndpoint represents an endpoint that the Gateway should handle.
type ProxyEndpoint struct {
	IDField int64  `json:"id"`
	Name    string `json:"name" index:"true" unique:"true"`
	Script  string `json:"script"`
}

// CollectionName returns a system-unique name that can be used to reference
// collections of this model, e.g. for URLs or database table names.
func (p *ProxyEndpoint) CollectionName() string {
	return "proxyEndpoints"
}

// ID returns the id of the model.
func (p *ProxyEndpoint) ID() interface{} {
	return p.IDField
}

// EmptyInstance returns an empty instance of the model.
func (p *ProxyEndpoint) EmptyInstance() Model {
	return &ProxyEndpoint{}
}

// Valid identifies whether or not the instance can be persisted.
func (p *ProxyEndpoint) Valid() (bool, error) {
	if p.Name == "" {
		return false, fmt.Errorf("Name must not be blank")
	}

	return true, nil
}

// Less is a case insensitive comparison on Name.
func (p *ProxyEndpoint) Less(other Model) bool {
	p2 := other.(*ProxyEndpoint)
	return strings.ToUpper(p.Name) < strings.ToUpper(p2.Name)
}

// MarshalToJSON returns a JSON representation of an instance or slice.
func (p *ProxyEndpoint) MarshalToJSON(data interface{}) ([]byte, error) {
	switch dataType := data.(type) {
	case *ProxyEndpoint:
		data = &proxyEndpointInstanceWrapper{dataType}
	case []Model:
		endpoints := make([]*ProxyEndpoint, len(dataType))
		for i, v := range dataType {
			endpoints[i] = v.(*ProxyEndpoint)
		}
		data = &proxyEndpointCollectionWrapper{endpoints}
	}
	return json.MarshalIndent(data, "", "    ")
}

// UnmarshalFromJSON returns an instance created from the passed JSON.
func (p *ProxyEndpoint) UnmarshalFromJSON(data []byte) (Model, error) {
	wrapper := proxyEndpointInstanceWrapper{}
	err := json.Unmarshal(data, &wrapper)
	return wrapper.Endpoint, err
}

// UnmarshalFromJSONWithID returns an instance created from the passed JSON,
// with its ID overridden.
func (p *ProxyEndpoint) UnmarshalFromJSONWithID(data []byte, id interface{}) (Model, error) {
	wrapper := proxyEndpointInstanceWrapper{}
	err := json.Unmarshal(data, &wrapper)
	wrapper.Endpoint.IDField = id.(int64)
	return wrapper.Endpoint, err
}

type proxyEndpointCollectionWrapper struct {
	Endpoints []*ProxyEndpoint `json:"proxyEndpoints"`
}

type proxyEndpointInstanceWrapper struct {
	Endpoint *ProxyEndpoint `json:"proxyEndpoint"`
}
