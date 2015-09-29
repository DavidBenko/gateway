package mangos

import (
	"errors"
	"fmt"
	"gateway/queue"
	"runtime"

	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/pub"
	"github.com/gdamore/mangos/transport/ipc"
	"github.com/gdamore/mangos/transport/tcp"
)

var _ = queue.Publisher(&PubSocket{})

// PubSocket implements queue.Publisher.
type PubSocket struct {
	s       mangos.Socket
	control chan signal
	c       chan []byte
	e       chan error
}

// Bind implements queue.Publisher.Bind for *PubSocket.
func (p *PubSocket) Bind(path string) error {
	if p.s == nil {
		return fmt.Errorf("mangos Publisher couldn't Bind to %s: nil socket", path)
	}

	control := make(chan signal)
	c := make(chan []byte, numChannels)
	e := make(chan error, numChannels)

	s := p.s

	go func() {
		var msg []byte
		var err error
		for {
			select {
			case <-control:
				close(e)
				return
			case msg = <-c:
				if err = s.Send(msg); err != nil {
					e <- err
				}
			}
		}
	}()

	p.c = c
	p.e = e
	p.control = control

	return s.Listen(path)
}

// Close implements io.Closer for *PubSocket.
func (p *PubSocket) Close() error {
	select {
	case p.control <- struct{}{}:
		// control was not yet closed, so we can safely close it.
		close(p.control)
	default:
	}

	if p.s != nil {
		return p.s.Close()
	}

	return nil
}

// Socket returns the underlying gdamore/mangos.Socket.
func (p *PubSocket) Socket() mangos.Socket {
	return p.s
}

// Channel returns a handle to the PubSocket's underlying channel.
func (p *PubSocket) Channel() chan<- []byte {
	return p.c
}

// Pub is a queue.PubBinding which creates a new mangos PubSocket.
func Pub(p queue.Publisher) (queue.Publisher, error) {
	if p != nil {
		return nil, fmt.Errorf("mangos.Pub expects nil Publisher, got %T", p)
	}

	s, err := pub.NewSocket()
	if err != nil {
		return nil, fmt.Errorf("mangos.Pub failed to make Mangos Socket: %s", err.Error())
	}

	return &PubSocket{s: s}, nil
}

// PubTCP is a queue.PubBinding which adds a TCP binding to the PubSocket.
func PubTCP(p queue.Publisher) (queue.Publisher, error) {
	s, err := GetPubSocket(p)
	switch {
	case err != nil:
		return nil, fmt.Errorf("PubTCP failed: %s", err)
	case s == nil:
		return nil, errors.New("PubTCP requires a non-nil Socket, use Pub first")
	}

	s.AddTransport(tcp.NewTransport())

	return p, nil
}

// PubIPC is a queue.PubBinding which adds a IPC binding to the PubSocket.
func PubIPC(p queue.Publisher) (queue.Publisher, error) {
	// https://github.com/gdamore/mangos/issues/2
	switch runtime.GOOS {
	case "linux", "darwin":
		// Unix domain sockets are supported on Linux and Darwin
	default:
		return nil, fmt.Errorf("PubIPC failed: mangos IPC transport not supported on OS %q", runtime.GOOS)
	}

	s, err := GetPubSocket(p)
	switch {
	case err != nil:
		return nil, fmt.Errorf("PubIPC failed: %s", err)
	case s == nil:
		return nil, errors.New("PubIPC requires a non-nil Socket, use Pub first")
	}

	s.AddTransport(ipc.NewTransport())

	return p, nil
}

// Gets a Mangos pub.Socket from a queue.Publisher containing a Mangos Socket.
func GetPubSocket(p queue.Publisher) (mangos.Socket, error) {
	if pP, ok := p.(*PubSocket); ok {
		return pP.s, nil
	}

	return nil, fmt.Errorf("GetPubSocket expected *mangos.PubSocket, got %T", p)
}
