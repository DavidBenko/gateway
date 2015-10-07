package queue

import (
	"errors"
	"fmt"
	"io"
)

// Server must be implemented by a queue server which can bind to an address.
// Clients can connect to a Server on the URI which Bind is called on.
type Server interface {
	io.Closer

	// Bind starts the server listening on the given URI.
	Bind(string) error
}

// Publisher must be implemented by a queue server which can bind to an
// address, accept clients, and send messages.  When it is garbage collected,
// it will be cleaned up using its Close method.
type Publisher interface {
	Server

	// Channels returns a channel to send messages on, and a channel to receive
	// any errors on.
	//
	// Usage:
	// s, e := pub.Channels()
	//
	// select {
	// case msg := <-s:
	//     // ...
	// case err := <-e:
	//     // ...
	// }
	Channels() (chan<- []byte, <-chan error)
}

// Publish sets up a Publisher with the given PubBindings, Binds it with the
// given path, and returns it.
func Publish(path string, bindings ...PubBinding) (Publisher, error) {
	if path == "" {
		return nil, errors.New("no path provided")
	}

	if len(bindings) < 1 {
		return nil, errors.New("no bindings provided")
	}

	p, err := newPublisher(bindings...)
	if err != nil {
		return nil, fmt.Errorf("bad publisher binding: %s", err.Error())
	}

	err = p.Bind(path)
	if err != nil {
		return nil, fmt.Errorf("publisher failed to bind: %s", err.Error())
	}

	return p, nil
}

func newPublisher(bindings ...PubBinding) (Publisher, error) {
	var p Publisher
	var err error
	for _, binding := range bindings {
		p, err = binding(p)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

// PubBinding is a method which takes a Publisher and returns a Publisher and
// any setup error.  PubBindings should be implemented by sub-packages.
type PubBinding func(Publisher) (Publisher, error)
