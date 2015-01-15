package admin

import (
	"errors"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	sql "gateway/sql"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/sessions"
)

var requestSession func(r *http.Request) *sessions.Session
var (
	userIDKey    = "user_id"
	accountIDKey = "account_id"
)

func setupSessions(conf config.ProxyAdmin) {
	if conf.AuthKey == "" {
		log.Fatal("Admin session auth key is required.")
	}

	rotating := (conf.AuthKey2 != "")

	sessionConfig := [][]byte{[]byte(conf.AuthKey)}
	if conf.EncryptionKey != "" {
		sessionConfig = append(sessionConfig, []byte(conf.EncryptionKey))
	} else if rotating {
		sessionConfig = append(sessionConfig, nil)
	}
	if rotating {
		sessionConfig = append(sessionConfig, []byte(conf.AuthKey2))
		if conf.EncryptionKey2 != "" {
			sessionConfig = append(sessionConfig, []byte(conf.EncryptionKey2))
		}
	}

	store := sessions.NewCookieStore(sessionConfig...)
	requestSession = func(r *http.Request) *sessions.Session {
		s, _ := store.Get(r, conf.SessionName)
		return s
	}
}

// RouteSessions routes all the endpoints for logging in and out
func RouteSessions(router aphttp.Router, db *sql.DB) {
	router.Handle("/sessions",
		handlers.MethodHandler{
			"POST":   aphttp.ErrorCatchingHandler(NewSessionHandler(db)),
			"DELETE": aphttp.ErrorCatchingHandler(DeleteSessionHandler(db)),
		})
}

// NewSessionHandler returns a hndler that adds authenticating information
// to the session if the credentials are valid.
func NewSessionHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := deserialize(&credentials, r); err != nil {
			log.Printf("%s Error reading credentials: %v", config.System, err)
			return aphttp.DefaultServerError()
		}

		user, err := model.FindUserByEmail(db, credentials.Email)
		if err != nil {
			return aphttp.NewError(errors.New("No user with that email."), 400)
		}
		if !user.ValidPassword(credentials.Password) {
			return aphttp.NewError(errors.New("Invalid password."), 400)
		}

		session := requestSession(r)
		session.Values[userIDKey] = user.ID
		session.Values[accountIDKey] = user.AccountID
		session.Save(r, w)

		w.WriteHeader(http.StatusOK)
		return nil
	}
}

// DeleteSessionHandler returns a hndler that removes authenticating information
// from the session.
func DeleteSessionHandler(db *sql.DB) aphttp.ErrorReturningHandler {
	return func(w http.ResponseWriter, r *http.Request) aphttp.Error {
		session := requestSession(r)
		delete(session.Values, userIDKey)
		delete(session.Values, accountIDKey)
		session.Save(r, w)

		w.WriteHeader(http.StatusOK)
		return nil
	}
}
