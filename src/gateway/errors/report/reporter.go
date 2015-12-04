package report

import "net/http"

// Reporter provides a general error reporting abstraction.
type Reporter interface {

	// Error reports an error.  If the error occurred within the context of an
	// http request, additional details can be reported if the http.Request object
	// is provided.
	Error(err error, request *http.Request)
}

var reporters []Reporter

// RegisterReporter registers a reporter.
func RegisterReporter(more ...Reporter) {
	reporters = append(reporters, more...)
}

// Error reports an error.  If the error occurred within the context of an
// http request, additional details can be reported if the http.Request object
// is provided.
func Error(err error, request *http.Request) {
	for _, rep := range reporters {
		rep.Error(err, request)
	}
}
