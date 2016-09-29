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

type RootEnvironmentsController struct {
	BaseController
}

func RouteRootEnvironments(controller *RootEnvironmentsController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	collectionRoutes := map[string]http.Handler{
		"GET": read(db, controller.List),
	}

	if conf.CORSEnabled {
		collectionRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(collectionRoutes))
}

func (c *RootEnvironmentsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	environments, err := model.AllEnvironmentsForAPIIDAndAccountID(db, 0, c.accountID(r))

	if err != nil {
		logreport.Printf("%s Error listing environment: %v\n%v", config.System, err, r)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(environments, w)
}

func (c *RootEnvironmentsController) serializeCollection(collection []*model.Environment,
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		Environments []*model.Environment `json:"environments"`
	}{collection}
	return serialize(wrapped, w)
}
