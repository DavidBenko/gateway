package mangos

import (
	"fmt"
	"gateway/queue"

	mg "github.com/gdamore/mangos"
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
) (bro *Broker, e error) {
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

	pubSetup = append(pubSetup, PubBuffer(1024))
	subSetup = append(subSetup, SubBuffer(1024))

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
		if closeErr := ps.Close(); closeErr != nil {
			err = wrapError("failed to close PubSocket after setup error", err, closeErr)
		}
		return nil, err
	}

	b.Client, b.Server = ss, ps

	// Now that the Broker is set up, it must be cleaned up if an error
	// occurs.  Defer the cleanup to keep it DRY.
	defer func() {
		if e != nil {
			if closeErr := bro.Close(); closeErr != nil {
				e = wrapError("failed to clean up after broker setup error", e, closeErr)
			}
		}
	}()

	sSock, err := getSubSocket(ss)
	if err != nil {
		return nil, err
	}

	pSock, err := getPubSocket(ps)
	if err != nil {
		return nil, err
	}

	// gdamore/mangos.Device implements a naive forwarder.  TODO: expand on
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
		return wrapError(
			"failed to close Broker Server after client close error: %s: %s",
			serverErr,
			clientErr,
		)
	case clientErr != nil:
		return clientErr
	case serverErr != nil:
		return serverErr
	default:
		return nil
	}
}
