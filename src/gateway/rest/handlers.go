package rest

import (
	"io/ioutil"
	"net/http"

	aphttp "gateway/http"

	"github.com/gorilla/mux"
)

func bodyHandler(handler func(w http.ResponseWriter, b []byte) aphttp.Error) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return aphttp.NewServerError(err)
		}
		return handler(w, body)
	}
}

func bodyAndIDHandler(handler func(w http.ResponseWriter, b []byte, id string) aphttp.Error) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return aphttp.NewServerError(err)
		}
		return handler(w, body, mux.Vars(r)["id"])
	}
}
