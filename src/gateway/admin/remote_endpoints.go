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

type RootRemoteEndpointsController struct {
	BaseController
}

func RouteRootRemoteEndpoints(controller *RootRemoteEndpointsController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	collectionRoutes := map[string]http.Handler{
		"GET": read(db, controller.List),
	}

	if conf.CORSEnabled {
		collectionRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(collectionRoutes))
}

func (c *RootRemoteEndpointsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	remoteEndpoints, err := model.AllRemoteEndpointsForAPIIDAndAccountID(db,
		0, c.accountID(r))

	if err != nil {
		logreport.Printf("%s Error listing remote endpoint: %v\n%v", config.System, err, r)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(remoteEndpoints, w)
}

func (c *RootRemoteEndpointsController) serializeCollection(collection []*model.RemoteEndpoint,
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		RemoteEndpoints []*model.RemoteEndpoint `json:"remote_endpoints"`
	}{collection}
	return serialize(wrapped, w)
}
