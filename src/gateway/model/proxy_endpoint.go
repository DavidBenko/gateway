package model

import (
	"encoding/json"
	"fmt"
	"gateway/proxy/vm"
)

// ProxyEndpoint represents an endpoint that the Gateway should handle.
type ProxyEndpoint struct {
	IDField int64  `json:"id" autoincrement:"true"`
	Name    string `json:"name" index:"true" unique:"true"`
	Script  string `json:"script"`
}

// CollectionName returns a system-unique name that can be used to reference
// collections of this model, e.g. for URLs or database table names.
func (p ProxyEndpoint) CollectionName() string {
	return "proxy_endpoints"
}

// ID returns the id of the model.
func (p ProxyEndpoint) ID() interface{} {
	return p.IDField
}

// EmptyInstance returns an empty instance of the model.
func (p ProxyEndpoint) EmptyInstance() Model {
	return ProxyEndpoint{}
}

// Valid identifies whether or not the instance can be persisted.
func (p ProxyEndpoint) Valid() (bool, error) {
	vm, err := vm.NewVM()
	if err != nil {
		return false, fmt.Errorf("Error setting up VM")
	}
	if _, err = vm.Run(p.Script); err != nil {
		return false, err
	}
	return true, nil
}

// UnmarshalFromJSON returns an instance created from the passed JSON.
func (p ProxyEndpoint) UnmarshalFromJSON(data []byte) (Model, error) {
	instance := ProxyEndpoint{}
	err := json.Unmarshal(data, &instance)
	return instance, err
}
