package report

import (
	"fmt"
	"net/http"
	"sync"

	"gopkg.in/airbrake/gobrake.v2"
)

type airbrakeReporter struct {
	ab *gobrake.Notifier
}

var once sync.Once
var airbrake *airbrakeReporter

// Error reports an error.  If the error occurred within the context of an
// http request, additional details can be reported if the http.Request object
// is provided
func (a *airbrakeReporter) Error(err error, request *http.Request) {
	a.ab.Notify(err, request)
}

// ConfigureAirbrake configures airbrake and returns a new reporter that is
// capable of reporting errors via airbrake.io
func ConfigureAirbrake(apiKey string, projectID int64, environment string) Reporter {
	if airbrake != nil {
		panic("Airbrake is already configured!")
	}

	once.Do(func() {
		ab := gobrake.NewNotifier(projectID, apiKey)
		// Override Printf function for airbrake notifier.  By default, it uses
		// log.Printf, which will try to send the logs to elastic search, which will
		// try to send an airbrake failure if it fails, and we could end up in a
		// feedback loop.  Using fmt.Printf will log the message to stdout,
		// while subverting the rest of the logging infrastructure
		ab.Printf = func(format string, v ...interface{}) {
			withNewline := fmt.Sprintf("%s\n", format)
			fmt.Printf(withNewline, v...)
		}
		airbrake = &airbrakeReporter{ab: ab}
		ab.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
			notice.Context["environment"] = environment
			return notice
		})
	})

	return airbrake
}
