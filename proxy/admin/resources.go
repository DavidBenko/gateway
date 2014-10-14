package admin

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/AnyPresence/gateway/db"
	"github.com/AnyPresence/gateway/model"
	"github.com/AnyPresence/gateway/rest"
)

type adminResource struct {
	rest.HTTPResource
	backingModel model.Model
	db           db.DB
}

func (r *adminResource) Name() string {
	return r.backingModel.CollectionName()
}

func (r *adminResource) Index() (resources interface{}, err error) {
	list, err := r.db.List(r.backingModel.EmptyInstance())
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(list, "", "    ")
}

func (r *adminResource) Create(data interface{}) (resource interface{}, err error) {
	instance, err := r.backingModel.UnmarshalFromJSON(data.([]byte))
	if err != nil {
		return nil, err
	}

	if err := r.db.Insert(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (r *adminResource) Show(id interface{}) (resource interface{}, err error) {
	int64id, err := r.int64ID(id)
	if err != nil {
		return nil, err
	}

	instance, err := r.db.Get(r.backingModel.EmptyInstance(), int64id)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(instance, "", "    ")
}

func (r *adminResource) Update(id interface{}, data interface{}) (resource interface{}, err error) {
	int64id, err := r.int64ID(id)
	if err != nil {
		return nil, err
	}

	instance, err := r.backingModel.UnmarshalFromJSON(data.([]byte))
	if err != nil {
		return nil, err
	}

	if int64id != instance.ID() {
		return nil, fmt.Errorf("Cannot change id via update")
	}

	if err := r.db.Update(instance); err != nil {
		return nil, err
	}

	return json.MarshalIndent(instance, "", "    ")
}

func (r *adminResource) Delete(id interface{}) error {
	int64id, err := strconv.ParseInt(id.(string), 10, 64)
	if err != nil {
		return err
	}

	return r.db.Delete(r.backingModel.EmptyInstance(), int64id)
}

func (r *adminResource) int64ID(id interface{}) (int64, error) {
	return strconv.ParseInt(id.(string), 10, 64)
}
