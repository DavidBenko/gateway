package admin

import (
	"encoding/json"
	"fmt"

	"github.com/AnyPresence/gateway/command"
	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
	"github.com/AnyPresence/gateway/rest"
	"github.com/goraft/raft"
)

type proxyEndpoint struct {
	rest.HTTPResource

	raft raft.Server
}

func (p *proxyEndpoint) Index() (resources []interface{}, err error) {
	fmt.Print("Index of proxy endpoints\n")
	return nil, nil
}

func (p *proxyEndpoint) Create(data interface{}) (resource interface{}, err error) {
	var instance model.ProxyEndpoint
	if err := json.Unmarshal(data.([]byte), &instance); err != nil {
		return nil, err
	}

	if _, err := p.raft.Do(command.CreateProxyEndpoint(instance)); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Show(id interface{}) (resource interface{}, err error) {
	db := p.raft.Context().(db.DB)
	instance, err := db.GetProxyEndpointByName(id.(string))
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(instance, "", "    ")
}

func (p *proxyEndpoint) Update(id interface{}) (resource interface{}, err error) {
	fmt.Printf("Update proxy endpoint with name %s\n", id)
	return nil, nil
}

func (p *proxyEndpoint) Delete(id interface{}) error {
	fmt.Printf("Delete proxy endpoint with name %v\n", id)
	return nil
}
