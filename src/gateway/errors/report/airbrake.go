package report

import (
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
		airbrake = &airbrakeReporter{ab: ab}
		ab.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
			notice.Context["environment"] = environment
			return notice
		})
	})

	return airbrake
}
