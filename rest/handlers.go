package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type errorReturningHandler func(w http.ResponseWriter, r *http.Request) error

func errorCatchingHandler(handler errorReturningHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			if err != nil {
				http.Error(w, fmt.Sprintf("An error occurred: %v", err), http.StatusMethodNotAllowed)
			}
		}
	})
}

func bodyHandler(handler func(w http.ResponseWriter, b []byte) error) errorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		return handler(w, body)
	}
}
