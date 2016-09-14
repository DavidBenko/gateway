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

type RootJobsController struct {
	BaseController
}

func RouteRootJobs(controller *RootJobsController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	collectionRoutes := map[string]http.Handler{
		"GET": read(db, controller.List),
	}

	if conf.CORSEnabled {
		collectionRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(collectionRoutes))
}

func (c *RootJobsController) List(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	job := model.ProxyEndpoint{
		Type:      model.ProxyEndpointTypeJob,
		AccountID: c.accountID(r),
	}
	jobs, err := job.All(db)

	if err != nil {
		logreport.Printf("%s Error listing job: %v\n%v", config.System, err, r)
		return aphttp.DefaultServerError()
	}

	return c.serializeCollection(jobs, w)
}

func (c *RootJobsController) serializeCollection(collection []*model.ProxyEndpoint,
	w http.ResponseWriter) aphttp.Error {

	wrapped := struct {
		Jobs []*model.ProxyEndpoint `json:"jobs"`
	}{collection}
	return serialize(wrapped, w)
}
