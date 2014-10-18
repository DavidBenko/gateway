package admin

import (
	"encoding/json"
	"fmt"
	"gateway/db"
	"gateway/model"
	"gateway/rest"
)

type adminRoutes struct {
	rest.HTTPResource
	db db.DB
}

func (r *adminRoutes) Name() string {
	return "routes"
}

func (r *adminRoutes) Index() (resources interface{}, err error) {
	return nil, fmt.Errorf("Routes have no index")
}

func (r *adminRoutes) Create(data interface{}) (resource interface{}, err error) {
	return nil, fmt.Errorf("Routes have no create")
}

func (r *adminRoutes) Show(id interface{}) (resource interface{}, err error) {
	return json.MarshalIndent(r.db.Router(), "", "    ")
}

func (r *adminRoutes) Update(id interface{}, data interface{}) (resource interface{}, err error) {
	var tmpRouter model.Router
	if err := json.Unmarshal(data.([]byte), &tmpRouter); err != nil {
		return nil, err
	}

	router, err := r.db.UpdateRouter(tmpRouter.Script)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(router, "", "    ")
}

func (r *adminRoutes) Delete(id interface{}) error {
	return fmt.Errorf("Routes have no delete")
}
