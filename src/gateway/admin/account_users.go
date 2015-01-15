package admin

import (
	aphttp "gateway/http"
	sql "gateway/sql"

	"github.com/gorilla/handlers"
)

// RouteAccountUsers routes all the endpoints for account based user management
// by site admins.
func RouteAccountUsers(router aphttp.Router, db *sql.DB) {
	router.Handle("/accounts/{accountID}/users",
		handlers.MethodHandler{
			"GET":  aphttp.ErrorCatchingHandler(ListUsersHandler(db, accountIDFromPath)),
			"POST": aphttp.ErrorCatchingHandler(CreateUserHandler(db, accountIDFromPath)),
		})
	router.Handle("/accounts/{accountID}/users/{id}",
		handlers.HTTPMethodOverrideHandler(handlers.MethodHandler{
			"GET":    aphttp.ErrorCatchingHandler(ShowUserHandler(db, accountIDFromPath)),
			"PUT":    aphttp.ErrorCatchingHandler(UpdateUserHandler(db, accountIDFromPath)),
			"DELETE": aphttp.ErrorCatchingHandler(DeleteUserHandler(db, accountIDFromPath)),
		}))
}
