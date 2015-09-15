package queue

import (
	"errors"
	"fmt"
	"io"
	"runtime"

	"gateway/queue/util"
)

// Client must be implemented by a queue client which can connect to an
// address.  When it goes out of scope, it will be cleaned up using its Close
// method.  Close can also be called at any time to immediately clean up.  If
// the Server it is connected to is Closed, the Client will also be Closed.
type Client interface {
	io.Closer

	// Connect attaches the Client to a Server on the given URI.  It should
	// save a reference to the Client interface in the implementation, so
	// Close can be called if the Server is Closed.
	Connect(string) error
}

// Subscriber must be implemented by a queue client which can connect to an
// address and receive messages.
type Subscriber interface {
	Client

	// Channel returns a channel which messages can be received on.
	Channel() chan []byte
}

// SubChannel wraps a channel with its Subscriber.  Close will be called when
// it is garbage collected, if it has not already been called.
type SubChannel struct {
	s Subscriber

	closer chan struct{}

	// C is the channel which subscribed messages will arrive on.
	C <-chan []byte
}

// Close closes and drains the channel, and calls Close on its Subscriber.
func (s *SubChannel) Close() error {
	return safeClose(s.closer)
}

func teardownSubChan(sc *SubChannel) func() error {
	return func() error {
		close(sc.s.Channel())
		util.Drain(sc.s.Channel())
		return sc.s.Close()
	}
}

// Subscribe sets up a Subscriber and returns a channel yielded by its
// Subscribe() method.  SubBindings are methods which take a Subscriber and
// return a Subscriber.  SubBindings should be implemented by sub-packages.
// Configuration options should include filters.
func Subscribe(path string, bindings ...SubBinding) (*SubChannel, error) {
	if path == "" {
		return nil, errors.New("no path provided")
	}

	if len(bindings) < 1 {
		return nil, errors.New("no bindings provided")
	}

	s, err := newSubscriber(bindings...)
	if err != nil {
		return nil, fmt.Errorf("bad subscriber binding: %s", err.Error())
	}

	closeChan := make(chan struct{})
	sCloser := &subCloser{closeChan, s}
	sChan := &SubChannel{sCloser, closeChan, s.Channel()}
	go waitFunc(closeChan, closeFunc(s))
	go waitFunc(closeChan, teardownSubChan(sChan))
	runtime.SetFinalizer(sCloser, closeCloser(sCloser))
	runtime.SetFinalizer(sChan, closeCloser(sChan))

	err = s.Connect(path)
	if err != nil {
		return nil, fmt.Errorf("subscriber failed to connect: %s", err.Error())
	}
	return sChan, nil
}

func newSubscriber(bindings ...SubBinding) (Subscriber, error) {
	var s Subscriber
	var err error
	for _, binding := range bindings {
		s, err = binding(s)
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

type SubBinding func(Subscriber) (Subscriber, error)
