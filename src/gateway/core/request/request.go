package request

// Request defines the interface for all requests proxy code can make.
type Request interface {
	Perform() Response
	Log(bool) string
	JSON() ([]byte, error)
}

// Response defines the interface for the results of Requests.
type Response interface {
	JSON() ([]byte, error)
	Log() string
}
