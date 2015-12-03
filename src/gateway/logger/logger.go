package logger

import (
	"gateway/errors/report"
	"log"
)

// Fatal reports errors via Airbrake and then delegates to log.Fatal
func Fatal(v ...interface{}) {
	reportErrors(v...)
	log.Fatal(v...)
}

// Fatalf reports errors via Airbrake and then delegates to log.Fatalf
func Fatalf(format string, v ...interface{}) {
	reportErrors(v...)
	log.Fatalf(format, v...)
}

// Fatalln reports errors via Airbrake and then delegates to log.Fatalln
func Fatalln(v ...interface{}) {
	reportErrors(v...)
	log.Fatalln(v...)
}

// Panic reports errors via Airbrake and then delegates to log.Panic
func Panic(v ...interface{}) {
	reportErrors(v...)
	log.Panic(v...)
}

// Panicf reports errors via Airbrake and then delegates to log.Panicf
func Panicf(format string, v ...interface{}) {
	reportErrors(v...)
	log.Panicf(format, v...)
}

// Panicln reports errors via Airbrake and then delegates to log.Panicln
func Panicln(v ...interface{}) {
	reportErrors(v...)
	log.Panicln(v...)
}

// Print reports errors via Airbrake and then delegates to log.Print
func Print(v ...interface{}) {
	reportErrors(v...)
	log.Print(v...)
}

// Printf reports errors via Airbrake and then delegates to log.Printf
func Printf(format string, v ...interface{}) {
	reportErrors(v...)
	log.Printf(format, v...)
}

// Println reports errors via Airbrake and then delegates to log.Println
func Println(v ...interface{}) {
	reportErrors(v...)
	log.Println(v...)
}

func reportErrors(v ...interface{}) {
	for _, item := range v {
		switch t := item.(type) {
		case error:
			report.Error(t, nil)
		}
	}
}
