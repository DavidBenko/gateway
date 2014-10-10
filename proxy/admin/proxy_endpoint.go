package admin

import (
	"encoding/json"
	"fmt"

	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
	"github.com/AnyPresence/gateway/rest"
)

type proxyEndpoint struct {
	rest.HTTPResource

	db db.DB
}

func (p *proxyEndpoint) Name() string {
	return "proxy_endpoints"
}

func (p *proxyEndpoint) Index() (resources interface{}, err error) {
	list, err := p.db.ListProxyEndpoints()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(list, "", "    ")
}

func (p *proxyEndpoint) Create(data interface{}) (resource interface{}, err error) {
	var instance model.ProxyEndpoint
	if err := json.Unmarshal(data.([]byte), &instance); err != nil {
		return nil, err
	}

	if err := p.db.CreateProxyEndpoint(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Show(id interface{}) (resource interface{}, err error) {
	instance, err := p.db.GetProxyEndpointByName(id.(string))
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Update(id interface{}, data interface{}) (resource interface{}, err error) {
	instance, err := p.db.GetProxyEndpointByName(id.(string))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data.([]byte), &instance); err != nil {
		return nil, err
	}
	if instance.Name != id.(string) {
		return nil, fmt.Errorf("TODO: Id is a string name right now, should not change. Prolly change to int64 id")
	}

	if err := p.db.UpdateProxyEndpoint(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Delete(id interface{}) error {
	return p.db.DeleteProxyEndpointByName(id.(string))
}
