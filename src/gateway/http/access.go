package http

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

// AccessLoggingHandler logs general access notes about a request, plus
// sets up an ID in the context for other methods to use for logging.
func AccessLoggingHandler(prefix string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		uuid, err := newUUID()
		if err != nil {
			log.Printf("%s Could not generate request UUID", prefix)
			uuid = "x"
		}

		context.Set(r, ContextRequestIDKey, uuid)

		l := &responseLogger{w: w}
		handler.ServeHTTP(l, r)

		clf := buildCommonLogLine(r, *r.URL, t, l.Status(), l.Size())
		log.Printf("%s [req %s] [access] %s", prefix, uuid, clf)
	})
}

// AccessLoggingRouter wraps all Handle calls in an AccessLoggingHandler.
type AccessLoggingRouter struct {
	prefix string
	router *mux.Router
}

// Handle wraps the handler in an AccessLoggingHandler for the router.
func (l *AccessLoggingRouter) Handle(pattern string, handler http.Handler) {
	l.router.Handle(pattern, AccessLoggingHandler(l.prefix, handler))
}

// NewAccessLoggingRouter wraps the router.
func NewAccessLoggingRouter(prefix string, router *mux.Router) *AccessLoggingRouter {
	return &AccessLoggingRouter{prefix: prefix, router: router}
}

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}

	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8],
		uuid[8:10], uuid[10:]), nil
}
