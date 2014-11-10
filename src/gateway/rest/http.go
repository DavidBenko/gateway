package rest

import (
	"fmt"
	"net/http"

	aphttp "gateway/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// HTTPResource is a Resource that knows how to communicate over HTTP.
type HTTPResource struct {
	Resource Resource
}

// Route adds the RESTful routes for this resource to the provided mux.Router.
func (h *HTTPResource) Route(router aphttp.Router) {
	router.Handle(fmt.Sprintf("/%s", h.Resource.Name()), handlers.MethodHandler{
		"GET":  h.IndexHandler(),
		"POST": h.CreateHandler()},
	)
	router.Handle(fmt.Sprintf("/%s/{id}", h.Resource.Name()),
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    h.ShowHandler(),
			"PUT":    h.UpdateHandler(),
			"DELETE": h.DeleteHandler(),
		}))
}

// RouteSingleton adds show and update routes to the provided mux.Router.
func (h *HTTPResource) RouteSingleton(router aphttp.Router) {
	router.Handle(fmt.Sprintf("/%s", h.Resource.Name()),
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET": h.ShowHandler(),
			"PUT": h.UpdateHandler(),
		}))
}

// IndexHandler returns an http.Handler that lists the resource.
func (h *HTTPResource) IndexHandler() http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			resources, err := h.Resource.Index()
			if err != nil {
				return aphttp.NewServerError(err)
			}
			fmt.Fprintf(w, "%s\n", resources)
			return nil
		})
}

// CreateHandler returns an http.Handler that creates the resource.
func (h *HTTPResource) CreateHandler() http.Handler {
	return aphttp.ErrorCatchingHandler(
		bodyHandler(func(w http.ResponseWriter, body []byte) aphttp.Error {
			resource, err := h.Resource.Create(body)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}

			fmt.Fprintf(w, "%s\n", resource)
			return nil
		}))
}

// ShowHandler returns an http.Handler that shows the resource.
func (h *HTTPResource) ShowHandler() http.Handler {
	return aphttp.ErrorCatchingHandler(func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		resource, err := h.Resource.Show(mux.Vars(r)["id"])
		if err != nil {
			return aphttp.NewError(err, http.StatusNotFound)
		}

		fmt.Fprintf(w, "%s\n", resource)
		return nil
	})
}

// UpdateHandler returns an http.Handler that updates the resource.
func (h *HTTPResource) UpdateHandler() http.Handler {
	return aphttp.ErrorCatchingHandler(
		bodyAndIDHandler(func(w http.ResponseWriter, body []byte, id string) aphttp.Error {
			resource, err := h.Resource.Update(id, body)
			if err != nil {
				return aphttp.NewError(err, http.StatusBadRequest)
			}

			fmt.Fprintf(w, "%s\n", resource)
			return nil
		}))
}

// DeleteHandler returns an http.Handler that deletes the resource.
func (h *HTTPResource) DeleteHandler() http.Handler {
	return aphttp.ErrorCatchingHandler(func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		if err := h.Resource.Delete(mux.Vars(r)["id"]); err != nil {
			return aphttp.NewError(err, http.StatusBadRequest)
		}

		w.WriteHeader(http.StatusOK)
		return nil
	})
}
