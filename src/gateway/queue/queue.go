package queue

import (
	"runtime"

	"gateway/queue/util"
)

type Filter string

// Server must be implemented by a queue server which can bind to an address.
type Server interface {
	Bind(string) error
}

// Client must be implemented by a queue client which can connect to an
// address.
type Client interface {
	Connect(string) error
}

// Subscriber must be implemented by a queue client which can connect to an
// address and receive messages.  When it goes out of scope, it will be cleaned
// up using its Close method and its channel will be drained.
type Subscriber interface {
	Client
	Channel() chan []byte
	Close() error
}

// Publisher must be implemented by a queue server which can bind to an
// address, accept clients, and send messages.  When it goes out of scope, it
// will be cleaned up using its Close method and its channel will be drained.
type Publisher interface {
	Server
	Channel() chan []byte
	Close() error
}

// PubChannel sets up a Publisher and returns a channel yielded by its Channel()
// method.  PubBindings are methods which take a Publisher and return a
// Publisher.  PubBindings should be implemented by sub-packages.
func PubChannel(path string, bindings ...PubBinding) (chan<- []byte, error) {
	p, err := newPublisher(bindings...)
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(p, pubTeardown)
	err = p.Bind(path)
	if err != nil {
		return nil, err
	}
	return p.Channel(), nil
}

// SubChannel sets up a Subscriber and returns a channel yielded by its
// Channel() method.  SubBindings are methods which take a Subscriber and
// return a Subscriber.  SubBindings should be implemented by sub-packages.
func SubChannel(path string, bindings ...SubBinding) (<-chan []byte, error) {
	s, err := newSubscriber(bindings...)
	if err != nil {
		return nil, err
	}

	runtime.SetFinalizer(s, subTeardown)
	err = s.Connect(path)
	if err != nil {
		return nil, err
	}
	return s.Channel(), nil
}

func pubTeardown(p Publisher) {
	p.Close()
	util.Drain(p.Channel())
}

func subTeardown(s Subscriber) {
	s.Close()
	util.Drain(s.Channel())
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
type PubBinding func(Publisher) (Publisher, error)
