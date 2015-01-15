package admin

import (
	aphttp "gateway/http"
	"gateway/model"
	"net/http"
)

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
//
// 	"gateway/config"
// 	aphttp "gateway/http"
// 	"gateway/model"
// 	sql "gateway/sql"
//
// 	"github.com/gorilla/handlers"
// 	"github.com/gorilla/mux"
// )
//
// // RouteUsers routes all the endpoints for user management
// func RouteUsers(router aphttp.Router, db *sql.DB) {
// 	router.Handle("/users",
// 		handlers.MethodHandler{
// 			"GET":  ListUsersHandler(db),
// 			"POST": CreateUserHandler(db),
// 		})
// 	router.Handle("/users/{id}",
// 		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
// 			"GET":    ShowUserHandler(db),
// 			"PUT":    UpdateUserHandler(db),
// 			"DELETE": DeleteUserHandler(db),
// 		}))
// }
//
// // ListUsersHandler returns an http.Handler that lists the users.
// func ListUsersHandler(db *sql.DB) http.Handler {
// 	return aphttp.ErrorCatchingHandler(
// 		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
// 			sql := newTracingDB(r, db)
//
// 			users := []model.User{}
// 			err := sql.Select(&users, "SELECT * FROM `users` ORDER BY `name` ASC;")
// 			if err != nil {
// 				log.Printf("%s Error listing users: %v", config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			wrapped := struct {
// 				Users []model.User `json:"users"`
// 			}{users}
//
// 			usersJSON, err := json.MarshalIndent(wrapped, "", "    ")
// 			if err != nil {
// 				log.Printf("%s Error marshaling users: %v", config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			fmt.Fprintf(w, "%s\n", string(usersJSON))
// 			return nil
// 		})
// }
//
// // CreateUserHandler returns an http.Handler that creates the user.
// func CreateUserHandler(db *sql.DB) http.Handler {
// 	return aphttp.ErrorCatchingHandler(
// 		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
// 			body, err := ioutil.ReadAll(r.Body)
// 			if err != nil {
// 				log.Printf("%s Error reading create user body: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			var wrapped struct {
// 				User model.User `json:"user"`
// 			}
// 			err = json.Unmarshal(body, &wrapped)
// 			if err != nil {
// 				log.Printf("%s Error unmarshalling user body: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
// 			user := &wrapped.User
//
// 			validationErrors := user.Validate()
// 			if !validationErrors.Empty() {
// 				errorsJSON, err := validationErrors.JSON()
// 				if err != nil {
// 					log.Printf("%s Error marshaling user: %v", config.System, err)
// 					return aphttp.DefaultServerError()
// 				}
// 				fmt.Fprintf(w, "%s\n", errorsJSON)
// 				return nil
// 			}
//
// 			tx := db.MustBegin()
// 			result, err := tx.Exec("INSERT INTO `users` (`name`) VALUES (?);",
// 				user.Name)
// 			if err != nil {
// 				log.Printf("%s Error inserting user: %v", config.System, err)
// 				tx.Rollback()
// 				return aphttp.DefaultServerError()
// 			}
// 			err = tx.Commit()
// 			if err != nil {
// 				log.Printf("%s Error committing insert user: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
// 			user.ID, err = result.LastInsertId()
// 			if err != nil {
// 				log.Printf("%s Error getting last insert id for user: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			userJSON, err := json.MarshalIndent(wrapped, "", "    ")
// 			if err != nil {
// 				log.Printf("%s Error marshaling user: %v", config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			fmt.Fprintf(w, "%s\n", userJSON)
// 			return nil
// 		})
// }
//
// // ShowUserHandler returns an http.Handler that shows the user.
// func ShowUserHandler(db *sql.DB) http.Handler {
// 	return aphttp.ErrorCatchingHandler(
// 		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
// 			id := mux.Vars(r)["id"]
//
// 			user := model.User{}
// 			err := db.Get(&user, "SELECT * FROM `users` WHERE `id` = ?;", id)
// 			if err != nil {
// 				return aphttp.NewError(fmt.Errorf("No user with id %s", id), 404)
// 			}
//
// 			wrapped := struct {
// 				User model.User `json:"user"`
// 			}{user}
//
// 			userJSON, err := json.MarshalIndent(wrapped, "", "    ")
// 			if err != nil {
// 				log.Printf("%s Error marshaling users: %v", config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			fmt.Fprintf(w, "%s\n", string(userJSON))
// 			return nil
// 		})
// }
//
// // UpdateUserHandler returns an http.Handler that updates the user.
// func UpdateUserHandler(db *sql.DB) http.Handler {
// 	return aphttp.ErrorCatchingHandler(
// 		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
// 			id := mux.Vars(r)["id"]
//
// 			body, err := ioutil.ReadAll(r.Body)
// 			if err != nil {
// 				log.Printf("%s Error reading create user body: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			var wrapped struct {
// 				User model.User `json:"user"`
// 			}
// 			err = json.Unmarshal(body, &wrapped)
// 			if err != nil {
// 				log.Printf("%s Error unmarshalling user body: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
// 			user := wrapped.User
//
// 			tx := db.MustBegin()
// 			result, err := tx.Exec("UPDATE `users` SET `name` = ? WHERE `id` = ?;",
// 				user.Name, id)
// 			if err != nil {
// 				log.Printf("%s Error updating user: %v", config.System, err)
// 				tx.Rollback()
// 				return aphttp.DefaultServerError()
// 			}
//
// 			numRows, err := result.RowsAffected()
// 			if err != nil || numRows != 1 {
// 				log.Printf("%s Error updating user: %v", config.System, err)
// 				tx.Rollback()
// 				return aphttp.DefaultServerError()
// 			}
//
// 			err = tx.Commit()
// 			if err != nil {
// 				log.Printf("%s Error committing update user: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			userJSON, err := json.MarshalIndent(wrapped, "", "    ")
// 			if err != nil {
// 				log.Printf("%s Error marshaling user: %v", config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			fmt.Fprintf(w, "%s\n", userJSON)
// 			return nil
// 		})
// }
//
// // DeleteUserHandler returns an http.Handler that deletes the user.
// func DeleteUserHandler(db *sql.DB) http.Handler {
// 	return aphttp.ErrorCatchingHandler(
// 		func(w http.ResponseWriter, r *http.Request) aphttp.Error {
// 			id := mux.Vars(r)["id"]
//
// 			tx := db.MustBegin()
// 			result, err := tx.Exec("DELETE FROM `users` WHERE `id` = ?;", id)
// 			if err != nil {
// 				log.Printf("%s Error deleting user: %v", config.System, err)
// 				tx.Rollback()
// 				return aphttp.DefaultServerError()
// 			}
//
// 			numRows, err := result.RowsAffected()
// 			if err != nil || numRows != 1 {
// 				log.Printf("%s Error deleting user: %v", config.System, err)
// 				tx.Rollback()
// 				return aphttp.DefaultServerError()
// 			}
//
// 			err = tx.Commit()
// 			if err != nil {
// 				log.Printf("%s Error committing delete user: %v",
// 					config.System, err)
// 				return aphttp.DefaultServerError()
// 			}
//
// 			w.WriteHeader(http.StatusOK)
// 			return nil
// 		})
// }

type wrappedUser struct {
	User *model.User `json:"user"`
}

type sanitizedUser struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type wrappedSanitizedUser struct {
	User *sanitizedUser `json:"user"`
}

func sanitizeUser(user *model.User) *sanitizedUser {
	return &sanitizedUser{user.ID, user.Name, user.Email}
}

func readUser(r *http.Request) (*model.User, error) {
	var wrapped wrappedUser
	if err := deserialize(&wrapped, r); err != nil {
		return nil, err
	}
	return wrapped.User, nil
}

func serializeUsers(users []*model.User, w http.ResponseWriter) aphttp.Error {
	wrappedUsers := struct {
		Users []*sanitizedUser `json:"users"`
	}{}
	for _, user := range users {
		wrappedUsers.Users = append(wrappedUsers.Users, sanitizeUser(user))
	}
	return serialize(wrappedUsers, w)
}
