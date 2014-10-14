package admin

import (
	"encoding/json"
	"strconv"

	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
	"github.com/AnyPresence/gateway/rest"
)

type proxyEndpoint struct {
	rest.HTTPResource

	db db.DB
}

func (p *proxyEndpoint) Name() string {
	return (model.ProxyEndpoint{}).CollectionName()
}

func (p *proxyEndpoint) Index() (resources interface{}, err error) {
	list, err := p.db.List(model.ProxyEndpoint{})
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

	if err := p.db.Insert(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Show(id interface{}) (resource interface{}, err error) {
	int64id, err := strconv.ParseInt(id.(string), 10, 64)
	if err != nil {
		return nil, err
	}

	instance, err := p.db.Get(model.ProxyEndpoint{}, int64id)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Update(id interface{}, data interface{}) (resource interface{}, err error) {
	var instance model.ProxyEndpoint
	if err := json.Unmarshal(data.([]byte), &instance); err != nil {
		return nil, err
	}

	if err := p.db.Update(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Delete(id interface{}) error {
	int64id, err := strconv.ParseInt(id.(string), 10, 64)
	if err != nil {
		return err
	}

	return p.db.Delete(model.ProxyEndpoint{}, int64id)

}
