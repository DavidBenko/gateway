package db

import (
	"testing"

	"github.com/AnyPresence/gateway/model"
)

var (
	foo = model.ProxyEndpoint{
		Name:   "foo",
		Path:   "foo",
		Method: "GET",
		Script: "script",
	}
	bar = model.ProxyEndpoint{
		Name:   "bar",
		Path:   "bar",
		Method: "POST",
		Script: "script2",
	}
)

func TestListProxyEndpoints(t *testing.T) {
	db := NewMemoryStore()
	db.CreateProxyEndpoint(foo)
	db.CreateProxyEndpoint(bar)

	list, _ := db.ListProxyEndpoints()
	if !endpointInList(foo, list) {
		t.Error("Expected foo to be in the list of all endpoints")
	}
	if !endpointInList(bar, list) {
		t.Error("Expected bar to be in the list of all endpoints")
	}
}

func endpointInList(a model.ProxyEndpoint, list []model.ProxyEndpoint) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
