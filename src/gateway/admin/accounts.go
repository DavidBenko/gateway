package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	aphttp "gateway/http"
	"gateway/model"
	sql "gateway/sql"

	"github.com/gorilla/handlers"
)

// RouteAccounts routes all the endpoints for account management
func RouteAccounts(router aphttp.Router, db *sql.DB) {
	router.Handle("/accounts",
		handlers.MethodHandler{
			"GET": ListAccountsHandler(db),
			// "POST": CreateAccountHandler(),
		})
	// router.Handle("/accounts/{id}",
	// 	handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
	// 		"GET":    ShowAccountHandler(),
	// 		"PUT":    UpdateAccountHandler(),
	// 		"DELETE": DeleteAccountHandler(),
	// 	}))
}

// ListAccountsHandler returns an http.Handler that lists the accounts.
func ListAccountsHandler(db *sql.DB) http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			accounts := []model.Account{}
			err := db.Select(&accounts, "SELECT * FROM `accounts` ORDER BY `id` ASC;")
			if err != nil {
				return aphttp.NewServerError(err)
			}

			wrapped := struct {
				Accounts []model.Account `json:"accounts"`
			}{accounts}

			accountsJSON, err := json.MarshalIndent(wrapped, "", "    ")
			if err != nil {
				return aphttp.NewServerError(err)
			}

			fmt.Fprintf(w, "%s\n", string(accountsJSON))
			return nil
		})
}

//
// // CreateAccountHandler returns an http.Handler that creates the account.
// func CreateAccountHandler() http.Handler {
// 	return aphttp.ErrorCatchingHandler(
// 		bodyHandler(func(w http.ResponseWriter, body []byte) aphttp.Error {
// 			resource, err := h.Resource.Create(body)
// 			if err != nil {
// 				return aphttp.NewError(err, http.StatusBadRequest)
// 			}
//
// 			fmt.Fprintf(w, "%s\n", resource)
// 			return nil
// 		}))
// }
//
// // ShowAccountHandler returns an http.Handler that shows the account.
// func ShowAccountHandler() http.Handler {
// 	return aphttp.ErrorCatchingHandler(func(w http.ResponseWriter, r *http.Request) aphttp.Error {
// 		resource, err := h.Resource.Show(mux.Vars(r)["id"])
// 		if err != nil {
// 			return aphttp.NewError(err, http.StatusNotFound)
// 		}
//
// 		fmt.Fprintf(w, "%s\n", resource)
// 		return nil
// 	})
// }
//
// // UpdateAccountHandler returns an http.Handler that updates the account.
// func UpdateAccountHandler() http.Handler {
// 	return aphttp.ErrorCatchingHandler(
// 		bodyAndIDHandler(func(w http.ResponseWriter, body []byte, id string) aphttp.Error {
// 			resource, err := h.Resource.Update(id, body)
// 			if err != nil {
// 				return aphttp.NewError(err, http.StatusBadRequest)
// 			}
//
// 			fmt.Fprintf(w, "%s\n", resource)
// 			return nil
// 		}))
// }
//
// // DeleteAccountHandler returns an http.Handler that deletes the account.
// func DeleteAccountHandler() http.Handler {
// 	return aphttp.ErrorCatchingHandler(func(w http.ResponseWriter, r *http.Request) aphttp.Error {
// 		if err := h.Resource.Delete(mux.Vars(r)["id"]); err != nil {
// 			return aphttp.NewError(err, http.StatusBadRequest)
// 		}
//
// 		w.WriteHeader(http.StatusOK)
// 		return nil
// 	})
// }
