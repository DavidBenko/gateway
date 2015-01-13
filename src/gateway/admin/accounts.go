package admin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	sql "gateway/sql"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// RouteAccounts routes all the endpoints for account management
func RouteAccounts(router aphttp.Router, db *sql.DB) {
	router.Handle("/accounts",
		handlers.MethodHandler{
			"GET":  ListAccountsHandler(db),
			"POST": CreateAccountHandler(db),
		})
	router.Handle("/accounts/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    ShowAccountHandler(db),
			"PUT":    UpdateAccountHandler(db),
			"DELETE": DeleteAccountHandler(db),
		}))
}

// ListAccountsHandler returns an http.Handler that lists the accounts.
func ListAccountsHandler(db *sql.DB) http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			accounts := []model.Account{}
			err := db.Select(&accounts, "SELECT * FROM `accounts` ORDER BY `id` ASC;")
			if err != nil {
				log.Printf("%s Error listing accounts: %v", config.System, err)
				return aphttp.DefaultServerError()
			}

			wrapped := struct {
				Accounts []model.Account `json:"accounts"`
			}{accounts}

			accountsJSON, err := json.MarshalIndent(wrapped, "", "    ")
			if err != nil {
				log.Printf("%s Error marshaling accounts: %v", config.System, err)
				return aphttp.DefaultServerError()
			}

			fmt.Fprintf(w, "%s\n", string(accountsJSON))
			return nil
		})
}

// CreateAccountHandler returns an http.Handler that creates the account.
func CreateAccountHandler(db *sql.DB) http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("%s Error reading create account body: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}

			var wrapped struct {
				Account model.Account `json:"account"`
			}
			err = json.Unmarshal(body, &wrapped)
			if err != nil {
				log.Printf("%s Error unmarshalling account body: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}
			account := &wrapped.Account

			validationErrors := account.Validate()
			if !validationErrors.Empty() {
				errorsJSON, err := validationErrors.JSON()
				if err != nil {
					log.Printf("%s Error marshaling account: %v", config.System, err)
					return aphttp.DefaultServerError()
				}
				fmt.Fprintf(w, "%s\n", errorsJSON)
				return nil
			}

			tx := db.MustBegin()
			result, err := tx.Exec("INSERT INTO `accounts` (`name`) VALUES (?);",
				account.Name)
			if err != nil {
				log.Printf("%s Error inserting account: %v", config.System, err)
				tx.Rollback()
				return aphttp.DefaultServerError()
			}
			err = tx.Commit()
			if err != nil {
				log.Printf("%s Error committing insert account: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}
			account.ID, err = result.LastInsertId()
			if err != nil {
				log.Printf("%s Error getting last insert id for account: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}

			accountJSON, err := json.MarshalIndent(wrapped, "", "    ")
			if err != nil {
				log.Printf("%s Error marshaling account: %v", config.System, err)
				return aphttp.DefaultServerError()
			}

			fmt.Fprintf(w, "%s\n", accountJSON)
			return nil
		})
}

// ShowAccountHandler returns an http.Handler that shows the account.
func ShowAccountHandler(db *sql.DB) http.Handler {
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

// UpdateAccountHandler returns an http.Handler that updates the account.
func UpdateAccountHandler(db *sql.DB) http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			id := mux.Vars(r)["id"]

			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Printf("%s Error reading create account body: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}

			var wrapped struct {
				Account model.Account `json:"account"`
			}
			err = json.Unmarshal(body, &wrapped)
			if err != nil {
				log.Printf("%s Error unmarshalling account body: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}
			account := wrapped.Account

			tx := db.MustBegin()
			result, err := tx.Exec("UPDATE `accounts` SET `name` = ? WHERE `id` = ?;",
				account.Name, id)
			if err != nil {
				log.Printf("%s Error updating account: %v", config.System, err)
				tx.Rollback()
				return aphttp.DefaultServerError()
			}

			numRows, err := result.RowsAffected()
			if err != nil || numRows != 1 {
				log.Printf("%s Error updating account: %v", config.System, err)
				tx.Rollback()
				return aphttp.DefaultServerError()
			}

			err = tx.Commit()
			if err != nil {
				log.Printf("%s Error committing update account: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}

			accountJSON, err := json.MarshalIndent(wrapped, "", "    ")
			if err != nil {
				log.Printf("%s Error marshaling account: %v", config.System, err)
				return aphttp.DefaultServerError()
			}

			fmt.Fprintf(w, "%s\n", accountJSON)
			return nil
		})
}

// DeleteAccountHandler returns an http.Handler that deletes the account.
func DeleteAccountHandler(db *sql.DB) http.Handler {
	return aphttp.ErrorCatchingHandler(
		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
			id := mux.Vars(r)["id"]

			tx := db.MustBegin()
			result, err := tx.Exec("DELETE FROM `accounts` WHERE `id` = ?;", id)
			if err != nil {
				log.Printf("%s Error deleting account: %v", config.System, err)
				tx.Rollback()
				return aphttp.DefaultServerError()
			}

			numRows, err := result.RowsAffected()
			if err != nil || numRows != 1 {
				log.Printf("%s Error deleting account: %v", config.System, err)
				tx.Rollback()
				return aphttp.DefaultServerError()
			}

			err = tx.Commit()
			if err != nil {
				log.Printf("%s Error committing delete account: %v",
					config.System, err)
				return aphttp.DefaultServerError()
			}

			w.WriteHeader(http.StatusOK)
			return nil
		})
}
