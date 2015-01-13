package admin

import (
	"gateway/config"
	aphttp "gateway/http"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
)

// DB defines the subset of database methods we want to use from
// a tracing DB.
type DB interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

type tracingDB struct {
	r          *http.Request
	underlying DB
}

func newTracingDB(r *http.Request, db DB) *tracingDB {
	return &tracingDB{r, db}
}

func (db *tracingDB) trace(start time.Time, query string) {
	reqID := context.Get(db.r, aphttp.ContextRequestIDKey)
	total := time.Since(start)
	log.Printf("%s [req %s] [sql] %s (%v)", config.Admin, reqID, query, total)
}

func (db *tracingDB) Get(dest interface{}, query string, args ...interface{}) error {
	defer db.trace(time.Now(), query)
	return db.underlying.Get(dest, query, args...)
}

func (db *tracingDB) Select(dest interface{}, query string, args ...interface{}) error {
	defer db.trace(time.Now(), query)
	return db.underlying.Select(dest, query, args...)
}
