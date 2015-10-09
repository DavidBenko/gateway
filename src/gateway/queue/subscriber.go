package queue

import (
	"errors"
	"fmt"
	"io"
)

// Client must be implemented by a queue client which can connect to an
// address.
type Client interface {
	io.Closer

	// Connect attaches the Client to a Server on the given URI.
	Connect(string) error
}

// Subscriber must be implemented by a queue client which can connect to an
// address and receive messages.
type Subscriber interface {
	Client

	// Channel returns a channel which messages can be received on.
	// Channels returns a channel to receive messages on, and a channel to receive
	// any errors on.
	//
	// Usage:
	// s, e := sub.Channels()
	//
	// select {
	// case msg := <-s:
	//     // ...
	// case err := <-e:
	//     // ...
	// }
	Channels() (<-chan []byte, <-chan error)
}

// Subscribe sets up a Subscriber with the given SubBindings, Connects it with
// the given path, and returns it.
func Subscribe(path string, bindings ...SubBinding) (Subscriber, error) {
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

	err = s.Connect(path)
	if err != nil {
		return nil, fmt.Errorf("subscriber failed to connect: %s", err.Error())
	}

	return s, nil
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

// SubBinding is a method which takes a Subscriber and returns a Subscriber and
// any setup error.  SubBindings should be implemented by sub-packages. Options
// should include filters if desired.
type SubBinding func(Subscriber) (Subscriber, error)
