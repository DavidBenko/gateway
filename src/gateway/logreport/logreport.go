package logreport

import (
	"gateway/errors/report"
	"log"
	"net/http"
)

var (
	// Print reports errors via Airbrake and then delegates to log.Print.
	Print = wrap(log.Print)
	// Printf reports errors via Airbrake and then delegates to log.Printf.
	Printf = wrapf(log.Printf)
	// Println reports errors via Airbrake and then delegates to log.Println.
	Println = wrap(log.Println)

	// Fatal reports errors via Airbrake and then delegates to log.Fatal.
	Fatal = wrap(log.Fatal)
	// Fatalf reports errors via Airbrake and then delegates to log.Fatalf.
	Fatalf = wrapf(log.Fatalf)
	// Fatalln reports errors via Airbrake and then delegates to log.Fatalln.
	Fatalln = wrap(log.Fatalln)

	// Panic reports errors via Airbrake and then delegates to log.Panic
	Panic = wrap(log.Panic)
	// Panicf reports errors via Airbrake and then delegates to log.Panicf.
	Panicf = wrapf(log.Panicf)
	// Panicln reports errors via Airbrake and then delegates to log.Panicln.
	Panicln = wrap(log.Panicln)
)

func wrap(f func(v ...interface{})) func(v ...interface{}) {
	return func(v ...interface{}) {
		reportErrors(v...)
		f(v...)
	}
}

func wrapf(f func(fmt string, v ...interface{})) func(fmt string, v ...interface{}) {
	return func(fmt string, v ...interface{}) {
		reportErrors(v...)
		f(fmt, v...)
	}
}

func reportErrors(v ...interface{}) {
	var (
		errs []error
		req  *http.Request
	)

	for _, item := range v {
		switch t := item.(type) {
		case error:
			errs = append(errs, t)
		case *http.Request:
			req = t
		}
	}

	for _, err := range errs {
		report.Error(err, req)
	}
}
