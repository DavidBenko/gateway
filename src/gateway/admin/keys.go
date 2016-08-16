package admin

import (
	"gateway/config"
	aphttp "gateway/http"
	apsql "gateway/sql"
	"net/http"

	"github.com/gorilla/handlers"
)

type KeysController struct {
	BaseController
}

func RouteKeys(controller *KeysController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET": read(db, controller.List),
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

func (c *KeysController) List(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
	return nil
}
