package admin

import (
	"net/http"

	"gateway/config"
	"gateway/core"
	aphttp "gateway/http"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type MetaScratchPadsController struct {
	ScratchPadsController
	*core.Core
}

func RouteScratchPads(controller ResourceController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	RouteResource(controller, path, router, db, conf)

	routes := map[string]http.Handler{
		"GET": read(db, controller.(*MetaScratchPadsController).Test),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path+"/{id}/test", handlers.MethodHandler(routes))
}

func (c *MetaScratchPadsController) Test(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	return nil
}
