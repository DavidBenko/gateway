package admin

import (
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	apsql "gateway/sql"
	"gateway/version"
	"github.com/gorilla/handlers"
)

type InfoController struct {
	BaseController
}

type Info struct {
	Id      string `json:"id"`
	Version string `json:"version"`
}

func RouteInfo(controller *InfoController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"GET": write(db, controller.Info),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

func (c *InfoController) Info(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
	infos := make([]Info, 1)
	infos[0] = Info{Id: "app", Version: version.Name()}
	wrapped := struct {
		InfoWrapper interface{} `json:"info"`
	}{infos}
	return serialize(wrapped, w)
}
