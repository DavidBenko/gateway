package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// HTTPResource is a Resource that knows how to communicate over HTTP.
type HTTPResource struct {
	Resource Resource
}

// Route adds the RESTful routes for this resource to the provided mux.Router.
func (h *HTTPResource) Route(router *mux.Router) {
	router.Handle(fmt.Sprintf("/%s", h.Resource.Name()), handlers.MethodHandler{
		"GET":  h.IndexHandler(),
		"POST": h.CreateHandler()},
	)
	router.Handle(fmt.Sprintf("/%s/{id}", h.Resource.Name()),
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    h.ShowHandler(),
			"PATCH":  h.UpdateHandler(),
			"PUT":    h.UpdateHandler(),
			"DELETE": h.DeleteHandler(),
		}))
}

// IndexHandler returns an http.Handler that lists the resource.
func (h *HTTPResource) IndexHandler() http.Handler {
	return errorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) error {
			resources, err := h.Resource.Index()
			if err != nil {
				return err
			}
			fmt.Fprintf(w, "%s\n", resources)
			return nil
		})
}

// CreateHandler returns an http.Handler that creates the resource.
func (h *HTTPResource) CreateHandler() http.Handler {
	return errorCatchingHandler(
		bodyHandler(func(w http.ResponseWriter, body []byte) error {
			resource, err := h.Resource.Create(body)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "%s\n", resource)
			return nil
		}))
}

// ShowHandler returns an http.Handler that shows the resource.
func (h *HTTPResource) ShowHandler() http.Handler {
	return errorCatchingHandler(func(w http.ResponseWriter, r *http.Request) error {
		resource, err := h.Resource.Show(mux.Vars(r)["id"])
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "%s\n", resource)
		return nil
	})
}

// UpdateHandler returns an http.Handler that updates the resource.
func (h *HTTPResource) UpdateHandler() http.Handler {
	return errorCatchingHandler(
		bodyAndIDHandler(func(w http.ResponseWriter, body []byte, id string) error {
			resource, err := h.Resource.Update(id, body)
			if err != nil {
				return err
			}

			fmt.Fprintf(w, "%s\n", resource)
			return nil
		}))
}

// DeleteHandler returns an http.Handler that deletes the resource.
func (h *HTTPResource) DeleteHandler() http.Handler {
	return errorCatchingHandler(func(w http.ResponseWriter, r *http.Request) error {
		if err := h.Resource.Delete(mux.Vars(r)["id"]); err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		return nil
	})
}
