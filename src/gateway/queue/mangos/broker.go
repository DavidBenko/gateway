package mangos

import (
	"fmt"
	"io"

	"gateway/errors"
	"gateway/queue"

	mg "github.com/go-mangos/mangos"
)

// Kind represents the different available Scalable Protocols messaging brokers.
type Kind int

const (
	// XPubXSub is an SP protocol for forwarding Publisher messages to a set
	// of Subscribers.
	// Note that Publishers to the Broker must be set up with Pub(true).
	XPubXSub Kind = iota
)

// FilterFunc is a transform which will be applied between the Server and
// Client.  It should mutate the received Mangos Message and return a handle to
// the new Message, or nil if it should not be passed on.
type FilterFunc func(*mg.Message) *mg.Message

// Broker has a Server and a Client for proxying published Mangos messages.
type Broker struct {
	Server queue.Server
	Client queue.Client
}

// NewBroker sets up a new Mangos Broker with the given Kind, Transport, paths,
// and FilterFuncs.  If any error occurs, it will try to clean up gracefully.
func NewBroker(
	k Kind,
	t Transport,
	pubPath, subPath string,
	// TODO: filters ...FilterFunc,
) (*Broker, error) {
	switch t {
	case TCP, IPC:
	default:
		return nil, fmt.Errorf("unknown Broker Transport %d", t)
	}

	switch k {
	case XPubXSub:
		return newXPubXSubBroker(t, pubPath, subPath)
	default:
		return nil, fmt.Errorf("unknown Broker Kind %d", k)
	}
}

func newXPubXSubBroker(
	trans Transport,
	pubPath, subPath string,
) (*Broker, error) {
	var (
		b        = new(Broker)
		pubSetup = []queue.PubBinding{XPub}
		subSetup = []queue.SubBinding{XSub}
	)

	switch trans {
	case IPC:
		pubSetup = append(pubSetup, PubIPC)
		subSetup = append(subSetup, SubIPC)
	case TCP:
		pubSetup = append(pubSetup, PubTCP)
		subSetup = append(subSetup, SubTCP)
	}

	pubSetup = append(pubSetup, PubBuffer(2048))
	subSetup = append(subSetup, SubBuffer(2048))

	ps, err := queue.Publish(
		pubPath,
		pubSetup...,
	)
	if err != nil {
		return nil, err
	}

	ss, err := queue.Subscribe(
		subPath,
		subSetup...,
	)
	if err != nil {
		return nil, tryClose(
			"failed to clean up after broker setup error", ps, err,
		)
	}

	b.Client, b.Server = ss, ps

	sSock, err := getSubSocket(ss)
	if err != nil {
		return nil, tryClose(
			"failed to clean up after broker setup error", b, err,
		)
	}

	pSock, err := getPubSocket(ps)
	if err != nil {
		return nil, tryClose(
			"failed to clean up after broker setup error", b, err,
		)
	}

	// go-mangos/mangos.Device implements a naive forwarder.  TODO: expand on
	// this if we have FilterFuncs.  Then we can have, for example, an
	// XPUB/XSUB broker subscribed only to a subset of messages.
	if err := mg.Device(sSock, pSock); err != nil {
		return nil, err
	}

	return b, nil
}

// Close gracefully closes the Broker's connections.
func (b *Broker) Close() error {
	var clientErr error
	var serverErr error

	if c := b.Client; c != nil {
		clientErr = c.Close()
	}
	if s := b.Server; s != nil {
		serverErr = s.Close()
	}

	switch {
	case clientErr != nil && serverErr != nil:
		return errors.WrapErrors(
			"failed to close Server after Client Close error",
			serverErr, clientErr,
		)
	case clientErr != nil:
		return clientErr
	case serverErr != nil:
		return serverErr
	default:
		return nil
	}
}

// tryClose tries to Close the Closer and wraps the given error with any Close
// error.
func tryClose(reason string, c io.Closer, err error) error {
	if closeErr := c.Close(); closeErr != nil {
		err = errors.WrapErrors(reason, err, closeErr)
	}

	return err
}
