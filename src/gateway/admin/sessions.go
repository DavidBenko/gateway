package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/config"
	aphttp "gateway/http"
	"gateway/model"
	apsql "gateway/sql"
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
func RouteSessions(path string, router aphttp.Router, db *apsql.DB,
	conf config.ProxyAdmin) {

	routes := map[string]http.Handler{
		"POST":   read(db, NewSessionHandler),
		"DELETE": read(db, DeleteSessionHandler),
	}
	if conf.CORSEnabled {
		routes["OPTIONS"] = aphttp.CORSOptionsHandler([]string{"POST", "DELETE", "OPTIONS"})
	}

	router.Handle(path, handlers.MethodHandler(routes))
}

// NewSessionHandler returns a hndler that adds authenticating information
// to the session if the credentials are valid.
func NewSessionHandler(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {

	// If you're trying to authenticate again, we're logging you out
	_deleteSession(w, r)

	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := deserialize(&credentials, r.Body); err != nil {
		log.Printf("%s Error reading credentials: %v", config.System, err)
		return aphttp.DefaultServerError()
	}

	user, err := model.FindUserByEmail(db, credentials.Email)
	if err != nil {
		return aphttp.NewError(errors.New("No user with that email."), http.StatusBadRequest)
	}
	if !user.ValidPassword(credentials.Password) {
		return aphttp.NewError(errors.New("Invalid password."), http.StatusBadRequest)
	}

	session := requestSession(r)
	session.Values[userIDKey] = user.ID
	session.Values[accountIDKey] = user.AccountID
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
	return nil
}

// DeleteSessionHandler returns a hndler that removes authenticating information
// from the session.
func DeleteSessionHandler(w http.ResponseWriter, r *http.Request,
	db *apsql.DB) aphttp.Error {
	_deleteSession(w, r)
	w.WriteHeader(http.StatusOK)
	return nil
}

func _deleteSession(w http.ResponseWriter, r *http.Request) {
	session := requestSession(r)
	delete(session.Values, userIDKey)
	delete(session.Values, accountIDKey)
	session.Save(r, w)
}

// NewSessionAuthRouter wraps a router with session checking behavior.
func NewSessionAuthRouter(router aphttp.Router, whitelist []string) aphttp.Router {
	return &SessionAuthRouter{router, whitelist}
}

// SessionAuthRouter wraps all Handle calls in an HTTP Basic check.
type SessionAuthRouter struct {
	router           aphttp.Router
	whitelistMethods []string
}

// Handle wraps the handler in the auth check.
func (s *SessionAuthRouter) Handle(pattern string, handler http.Handler) {
	s.router.Handle(pattern, s.Wrap(handler))
}

// Wrap provides the wrapped handling functionality.
func (s *SessionAuthRouter) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.checkAuth(w, r) {
			handler.ServeHTTP(w, r)
			return
		}

		var body string
		data, err := json.MarshalIndent(struct {
			Error string `json:"error"`
		}{"Unauthorized"}, "", "    ")
		if err == nil {
			body = string(data)
		} else {
			// Fall back to non-JSON body
			body = fmt.Sprintf("%s\n", "Unauthorized\n")
		}

		http.Error(w, body, http.StatusUnauthorized)
	})
}

func (s *SessionAuthRouter) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if s.isWhitelisted(r) {
		return true
	}

	session := requestSession(r)
	userID := session.Values[userIDKey]
	accountID := session.Values[accountIDKey]
	if userID == nil || accountID == nil {
		return false
	}
	return userID.(int64) > 0 && accountID.(int64) > 0
}

func (s *SessionAuthRouter) isWhitelisted(r *http.Request) bool {
	for _, method := range s.whitelistMethods {
		if r.Method == method {
			return true
		}
	}
	return false
}
