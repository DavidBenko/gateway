package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	sql "gateway/sql"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// RouteSessions routes all the endpoints for logging in and out
func RouteSessions(router aphttp.Router, db *sql.DB) {
	router.Handle("/sessions",
		handlers.MethodHandler{
			"POST":   NewSessionHandler(db),
			"DELETE": DeleteSessionHandler(db),
		})
}

// NewSessionHandler creates a new session based on the passed in credentials.
// Credentials should look like:
//     { "email": "email@sample.com", "password": "superS3cure" }
func NewSessionHandler(db *sql.DB) http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("%s Error reading new session body: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}

			var auth struct {
				Email    string `json:"email"`
				Password string `json:"password"`
			}
			err = json.Unmarshal(body, &auth)
			if err != nil {
				log.Printf("%s Error unmarshalling new session body: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}

			badAuthRequest := aphttp.NewError(errors.New("The email and password "+
				"you provided are invalid."),
				http.StatusBadRequest)

			var user model.User
			err := db.Get(&user, "SELECT * FROM `users` WHERE `email` = ?;",
				auth.Email)
			if err != nil {
				return badAuthRequest
			}

			accountsJSON, err := json.MarshalIndent(wrapped, "", "    ")
			if err != nil {
				log.Printf("%s Error marshaling accounts: %v", config.System, err)
				return aphttp.DefaultServerError()
			}

			fmt.Fprintf(w, "%s\n", string(accountsJSON))
			return nil
		})
}

// DeleteSessionHandler returns an http.Handler that shows the account.
func DeleteSessionHandler(db *sql.DB) http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			id := mux.Vars(r)["id"]

			account := model.Account{}
			err := db.Get(&account, "SELECT * FROM `accounts` WHERE `id` = ?;", id)
			if err != nil {
				return aphttp.NewError(fmt.Errorf("No account with id %s", id), 404)
			}

			wrapped := struct {
				Account model.Account `json:"account"`
			}{account}

			accountJSON, err := json.MarshalIndent(wrapped, "", "    ")
			if err != nil {
				log.Printf("%s Error marshaling accounts: %v", config.System, err)
				return aphttp.DefaultServerError()
			}

			fmt.Fprintf(w, "%s\n", string(accountJSON))
			return nil
		})
}
