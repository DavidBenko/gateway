package admin

import (
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type RootEndpointGroupsController struct {
	BaseController
}

func RouteRootEndpointGroups(controller *RootEndpointGroupsController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	collectionRoutes := map[string]http.Handler{
		"GET": read(db, controller.List),
	}

	if conf.CORSEnabled {
		collectionRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(collectionRoutes))
}

func (c *RootEndpointGroupsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	groups, err := model.AllEndpointGroupsForAPIIDAndAccountID(db, 0, c.accountID(r))

	if err != nil {
		logreport.Printf("%s Error listing endpoint group: %v\n%v", config.System, err, r)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(groups, w)
}

func (c *RootEndpointGroupsController) serializeCollection(collection []*model.EndpointGroup,
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		EndpointGroups []*model.EndpointGroup `json:"endpoint_groups"`
	}{collection}
	return serialize(wrapped, w)
}
