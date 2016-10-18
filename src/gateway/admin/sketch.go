package admin

import (
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/handlers"
)

type SketchController struct {
	BaseController
}

func RouteSketch(controller *SketchController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	collectionRoutes := map[string]http.Handler{
		"GET": read(db, controller.List),
	}

	if conf.CORSEnabled {
		collectionRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(collectionRoutes))
}

func (c *SketchController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	sketch := model.NewSketch()

	return c.serializeInstance(sketch, w)
}

func (c *SketchController) serializeInstance(instance interface{},
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		Sketch interface{} `json:"sketch"`
	}{instance}
	return serialize(wrapped, w)
}
