package admin

import (
	"gateway/config"
	aphttp "gateway/http"
	apsql "gateway/sql"
	"net/http"

	"github.com/gorilla/handlers"
)

// ResourceController defines what we expect a controller to do to route
// a RESTful resource
type ResourceController interface {
	List(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error
	Create(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
	Show(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error
	Update(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
	Delete(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error
}

func RouteResource(controller ResourceController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	collectionRoutes := map[string]http.Handler{
		"GET":  read(db, controller.List),
		"POST": write(db, controller.Create),
	}
	instanceRoutes := map[string]http.Handler{
		"GET":    read(db, controller.Show),
		"PUT":    write(db, controller.Update),
		"DELETE": write(db, controller.Delete),
	}

	if conf.CORSEnabled {
		collectionRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "POST", "OPTIONS"})
		instanceRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "PUT", "DELETE", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(collectionRoutes))
	router.Handle(path+"/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler(instanceRoutes)))
}

func read(db *apsql.DB, handler DatabaseAwareHandler) http.Handler {
	return aphttp.JSONErrorCatchingHandler(DatabaseWrappedHandler(db, handler))
}

func write(db *apsql.DB, handler TransactionAwareHandler) http.Handler {
	return aphttp.JSONErrorCatchingHandler(TransactionWrappedHandler(db, handler))
}
