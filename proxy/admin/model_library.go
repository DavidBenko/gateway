package admin

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
	"github.com/AnyPresence/gateway/rest"
)

type library struct {
	rest.HTTPResource

	db db.DB
}

func (p *library) Name() string {
	return (model.Library{}).CollectionName()
}

func (p *library) Index() (resources interface{}, err error) {
	list, err := p.db.List(model.Library{})
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(list, "", "    ")
}

func (p *library) Create(data interface{}) (resource interface{}, err error) {
	var instance model.Library
	if err := json.Unmarshal(data.([]byte), &instance); err != nil {
		return nil, err
	}

	if err := p.db.Insert(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (p *library) Show(id interface{}) (resource interface{}, err error) {
	int64id, err := p.int64ID(id)
	if err != nil {
		return nil, err
	}

	instance, err := p.db.Get(model.Library{}, int64id)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(instance, "", "    ")
}

func (p *library) Update(id interface{}, data interface{}) (resource interface{}, err error) {
	int64id, err := p.int64ID(id)
	if err != nil {
		return nil, err
	}

	var instance model.Library
	if err := json.Unmarshal(data.([]byte), &instance); err != nil {
		return nil, err
	}

	if int64id != instance.IDField {
		return nil, fmt.Errorf("Cannot change id via update")
	}

	if err := p.db.Update(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (p *library) Delete(id interface{}) error {
	int64id, err := strconv.ParseInt(id.(string), 10, 64)
	if err != nil {
		return err
	}

	return p.db.Delete(model.Library{}, int64id)
}

func (p *library) int64ID(id interface{}) (int64, error) {
	return strconv.ParseInt(id.(string), 10, 64)
}
