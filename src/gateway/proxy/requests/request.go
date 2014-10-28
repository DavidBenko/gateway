package requests

import (
	"fmt"
	"gateway/config"
	"log"
	"time"
)

// Request defines the interface for all requests proxy code can make.
type Request interface {
	Perform(c chan<- responsePayload, index int)
	Log() string
}

// Response defines the interface for the results of Requests.
type Response interface {
	JSON() ([]byte, error)
	Log() string
}

type responsePayload struct {
	index    int
	response Response
}

// RequestFromData unpacks a request specified by the data.
func RequestFromData(requestData []string) (Request, error) {
	if len(requestData) != 2 {
		return nil, fmt.Errorf("Request data must have type and JSON data")
	}

	requestType := requestData[0]
	requestJSON := requestData[1]

	switch requestType {
	case "HTTP":
		return NewHTTPRequest(requestJSON)
	default:
		return nil, fmt.Errorf("The request type '%s' is not supported", requestType)
	}
}

// MakeRequests makes the requests and returns all responses.
func MakeRequests(requests []Request, reqID string) ([]Response, error) {
	start := time.Now()

	n := len(requests)
	requestDurations := make([]time.Duration, n)
	responses := make([]Response, n)

	c := make(chan responsePayload)
	for i, request := range requests {
		go request.Perform(c, i)
	}

	for i := 0; i < n; i++ {
		select {
		case r := <-c:
			requestDurations[r.index] = time.Since(start)
			responses[r.index] = r.response
		}
	}

	for i, request := range requests {
		log.Printf("%s [req %s] [request] %s %s (%v)", config.Proxy,
			reqID, request.Log(), responses[i].Log(), requestDurations[i])
	}

	return responses, nil
}
