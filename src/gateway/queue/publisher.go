package queue

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	"gateway/queue/util"
)

// Server must be implemented by a queue server which can bind to an address.
// Clients can connect to a Server on the URI which Bind is called on.  When it
// goes out of scope, it and its clients will be cleaned up using its Close
// method.  Close can also be called at any time to immediately clean up.
type Server interface {
	io.Closer

	// Bind starts the server listening on the given URI.  It should save a
	// reference to the Server interface in the implementation.
	Bind(string) error
}

// Publisher must be implemented by a queue server which can bind to an
// address, accept clients, and send messages.  When it goes out of scope, it
// will be cleaned up using its Close method and its channel will be drained.
//
// The channel it returns from Channel will be wrapped with a reference back to
// the Publisher.  When this channel goes out of scope, Close() will be called
// on the Publisher.
type Publisher interface {
	Server

	Channel() chan []byte
}

// PubChannel wraps a channel with its Subscriber.  Close will be called when
// it is garbage collected, if it has not already been called.
type PubChannel struct {
	p Publisher

	closer chan struct{}

	// C is the channel which messages can be published on.
	C chan<- []byte
}

// Close closes its channel, and calls Close on its Publisher.
func (p *PubChannel) Close() error {
	return safeClose(p.closer)
}

func teardownPubChan(pc *PubChannel) func() error {
	return func() error {
		close(pc.p.Channel())
		util.Drain(pc.p.Channel())
		return pc.p.Close()
	}
}

// Publish sets up a Publisher with the given PubBindings and returns a
// PubChannel from it.  The PubChannel keeps the Publisher alive until Close is
// called; at that time, the channel will be drained and Close will be called
// on the Publisher.
//
// PubBindings are methods which take a Publisher and return a Publisher.
// PubBindings should be implemented by sub-packages.
func Publish(path string, bindings ...PubBinding) (*PubChannel, error) {
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

	closeChan := make(chan struct{})
	pCloser := &pubCloser{closeChan, p}
	pChan := &PubChannel{pCloser, closeChan, p.Channel()}
	go waitFunc(closeChan, closeFunc(p))
	go waitFunc(closeChan, teardownPubChan(pChan))
	runtime.SetFinalizer(pCloser, closeCloser(pCloser))
	runtime.SetFinalizer(pChan, closeCloser(pChan))

	err = pCloser.Bind(path)
	if err != nil {
		return nil, fmt.Errorf("publisher failed to bind: %s", err.Error())
	}

	return pChan, nil
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

type PubBinding func(Publisher) (Publisher, error)
