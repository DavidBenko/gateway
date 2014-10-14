package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

type errorReturningHandler func(w http.ResponseWriter, r *http.Request) HTTPError

func errorCatchingHandler(handler errorReturningHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			if err != nil {
				http.Error(w, fmt.Sprintf("%v\n", err.Error()), err.Code())
			}
		}
	})
}

func bodyHandler(handler func(w http.ResponseWriter, b []byte) HTTPError) errorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) HTTPError {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &httpError{err: err}
		}
		return handler(w, body)
	}
}

func bodyAndIDHandler(handler func(w http.ResponseWriter, b []byte, id string) HTTPError) errorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) HTTPError {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return &httpError{err: err}
		}
		return handler(w, body, mux.Vars(r)["id"])
	}
}
