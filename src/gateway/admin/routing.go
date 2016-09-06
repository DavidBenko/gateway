package admin

import (
	"encoding/json"
	"net/http"
	"sync"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/logreport"
	"gateway/model"
	apsql "gateway/sql"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func RouteSingularResource(controller SingularResourceController, path string,
	router aphttp.Router, db *apsql.DB, conf config.ProxyAdmin) {

	singularRoutes := map[string]http.Handler{
		"GET":    read(db, controller.Show),
		"PUT":    write(db, controller.Update),
		"POST":   write(db, controller.Create),
		"DELETE": write(db, controller.Delete),
	}

	if conf.CORSEnabled {
		singularRoutes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"GET", "PUT", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(singularRoutes))
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

type HostMatcher struct {
	router *mux.Router
	mutex  sync.RWMutex
	db     *apsql.DB
}

type HostMatch struct {
	AccountID int64
	APIID     int64
}

func newHostMatcher(db *apsql.DB) *HostMatcher {
	m := &HostMatcher{db: db}
	m.rebuild()
	db.RegisterListener(m)
	return m
}

func (m *HostMatcher) rebuild() {
	hosts, err := model.AllHosts(m.db)
	if err != nil {
		logreport.Printf("%s Error fetching hosts to route: %v", config.System, err)
		return
	}

	router := mux.NewRouter()
	for _, host := range hosts {
		route := router.NewRoute()
		match := &HostMatch{
			AccountID: host.AccountID,
			APIID:     host.APIID,
		}
		name, err := json.Marshal(match)
		if err != nil {
			logreport.Fatal(err)
		}
		route.Name(string(name))
		route.Host(host.Hostname)
	}

	defer m.mutex.Unlock()
	m.mutex.Lock()
	m.router = router
}

func (m *HostMatcher) isRouted(r *http.Request, rm *mux.RouteMatch) bool {
	var match mux.RouteMatch
	defer m.mutex.RUnlock()
	m.mutex.RLock()
	ok := m.router.Match(r, &match)
	if ok {
		hostMatch := &HostMatch{}
		err := json.Unmarshal([]byte(match.Route.GetName()), hostMatch)
		if err != nil {
			logreport.Fatal(err)
		}
		context.Set(r, aphttp.ContextAccountIDKey, hostMatch.AccountID)
		context.Set(r, aphttp.ContextAPIIDKey, hostMatch.APIID)
		context.Set(r, aphttp.ContextMatchKey, &match)
	}
	return ok
}

func (s *HostMatcher) Notify(n *apsql.Notification) {
	if n.Table == "hosts" {
		go s.rebuild()
	}
}

func (s *HostMatcher) Reconnect() {
	logreport.Printf("%s Admin notified of database reconnection", config.System)
	go s.rebuild()
}

type DatabaseHostAwareHandler func(w http.ResponseWriter, r *http.Request,
	db *apsql.DB, match *HostMatch) aphttp.Error

func DatabaseHostHandler(handler DatabaseHostAwareHandler) DatabaseAwareHandler {
	return func(w http.ResponseWriter, r *http.Request, db *apsql.DB) aphttp.Error {
		match := context.Get(r, aphttp.ContextMatchKey).(*mux.RouteMatch)
		hostMatch := &HostMatch{}
		err := json.Unmarshal([]byte(match.Route.GetName()), hostMatch)
		if err != nil {
			logreport.Fatal(err)
		}
		return handler(w, r, db, hostMatch)
	}
}

type TransactionHostAwareHandler func(w http.ResponseWriter, r *http.Request,
	tx *apsql.Tx, match *HostMatch) aphttp.Error

func TransactionHostHandler(handler TransactionHostAwareHandler) TransactionAwareHandler {
	return func(w http.ResponseWriter, r *http.Request, tx *apsql.Tx) aphttp.Error {
		match := context.Get(r, aphttp.ContextMatchKey).(*mux.RouteMatch)
		hostMatch := &HostMatch{}
		err := json.Unmarshal([]byte(match.Route.GetName()), hostMatch)
		if err != nil {
			logreport.Fatal(err)
		}
		return handler(w, r, tx, hostMatch)
	}
}

func readForHost(db *apsql.DB, handler DatabaseHostAwareHandler) http.Handler {
	return context.ClearHandler(aphttp.JSONErrorCatchingHandler(DatabaseWrappedHandler(db, DatabaseHostHandler(handler))))
}

func writeForHost(db *apsql.DB, handler TransactionHostAwareHandler) http.Handler {
	return context.ClearHandler(aphttp.JSONErrorCatchingHandler(TransactionWrappedHandler(db, TransactionHostHandler(handler))))
}
