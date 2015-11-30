package report

import (
	"log"
	"net/http"
)

type logReporter struct{}

// Error reports an error.  If the error occurred within the context of an
// http request, additional details can be reported if the http.Request object
// is provided
func (a *logReporter) Error(err error, request *http.Request) {
	if request == nil {
		log.Printf("Error encountered: %v", err)
	} else {
		log.Printf("Error encountered: %v\nHTTP Request: %v", err, request)
	}
}

// NewLogReporter configures a new logReporter and returns it
func NewLogReporter() Reporter {
	return &logReporter{}
}
