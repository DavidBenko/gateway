package request

import "io"

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

// ReusableConnection allows the reuse of a connection across multiple proxy
// endpoint components (i.e. does not terminate when the component is finished
// executing)
type ReusableConnection interface {
	// CreateOrReuse does one of two things:
	//   1. If the io.Closer passed in is nil, it creates a new connection, which
	//      implements the io.Closer interface, and returns it
	//   2. If the io.Closer passed in is not nil, it is free to utilize the
	//      io.Closer that was passed in to connect to the underlying data source.
	//      Ultimately, it should return the io.Closer for which Close needs
	//      to be invoked in order to clean up the connection at a later point.
	CreateOrReuse(io.Closer) (io.Closer, error)
}
