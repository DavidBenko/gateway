package queue

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

// Server must be implemented by a queue server which can bind to an address.
// Clients can connect to a Server on the URI which Bind is called on.  When it
// is garbage collected, it and its clients will be cleaned up using their Close
// method.  Close can also be called at any time to immediately clean up.
type Server interface {
	io.Closer

	// Bind starts the server listening on the given URI.
	Bind(string) error
}

// Publisher must be implemented by a queue server which can bind to an
// address, accept clients, and send messages.  When it is garbage collected,
// it will be cleaned up using its Close method and its channel will be
// drained.
type Publisher interface {
	Server

	// Channel returns a handle for the user to send messages on.
	Channel() chan []byte
}

// PubChannel wraps a channel with its Publisher.  Close will be called when
// it is garbage collected, if it has not already been called.
type PubChannel struct {
	c io.Closer

	closer chan struct{}

	// C is the channel which messages can be published on.
	C chan<- []byte
}

// Close triggers the PubChannel's teardown, calling Close on its Publisher.
func (p *PubChannel) Close() error {
	return safeClose(p.closer)
}

func teardownPubChan(pc *PubChannel) func() error {
	channel := pc.C
	c := pc.c
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					err = e
				}
				err = fmt.Errorf("panicked on Publisher teardown: %#v", r)
			}
		}()
		close(channel)
		return c.Close()
	}
}

// Publish sets up a Publisher with the given PubBindings, Binds it with the
// given path, and returns a *PubChannel made using its Channel() method.  The
// PubChannel keeps the Publisher alive until Close is called; at that time,
// Close will be called on the Publisher.
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
		e := pCloser.Close()
		if e != nil {
			err = fmt.Errorf("failed to close publisher on error, %s: %s", e, err)
		}
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

// PubBinding is a method which takes a Publisher and returns a Publisher and
// any setup error.  PubBindings should be implemented by sub-packages.
type PubBinding func(Publisher) (Publisher, error)
