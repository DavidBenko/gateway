package mangos

import (
	"errors"
	"fmt"
	"gateway/queue"
	"runtime"
	"time"

	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/pub"
	"github.com/gdamore/mangos/transport/ipc"
	"github.com/gdamore/mangos/transport/tcp"
)

var _ = queue.Publisher(&PubSocket{})

// PubSocket implements queue.Publisher.
type PubSocket struct {
	s        mangos.Socket
	control  chan signal
	done     chan signal
	c        chan []byte
	e        chan error
	buffSize int
}

func (p *PubSocket) Bind(path string) error {
	s := p.s

	if s == nil {
		return fmt.Errorf("mangos Publisher couldn't Bind to %s: nil socket", path)
	}

	if p.buffSize != 0 {
		if err := s.SetOption(mangos.OptionWriteQLen, p.buffSize); err != nil {
			return err
		}
	}

	if err := s.Listen(path); err != nil {
		return err
	}

	control := make(chan signal)
	done := make(chan signal)
	c := make(chan []byte, channelSize)
	e := make(chan error, channelSize)

	p.c = c
	p.e = e
	p.control = control
	p.done = done

	go func() {
		var msg []byte
		var err error
		for {
			select {
			case <-control:
				close(e)
				close(done)
				return
			case msg = <-c:
				if err = s.Send(msg); err != nil {
					e <- err
				}
			}
		}
	}()

	return nil
}

func (p *PubSocket) Close() (e error) {
	if p.control != nil {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					e = err
				}
				e = fmt.Errorf("unexpected panic: %v", e)
			}
		}()

		close(p.control)
	}

	if p.s != nil {
		if err := p.s.Close(); err != nil {
			return err
		}
	}

	// If nil, Bind was never called and we can cleanly close.
	if p.done != nil {
		select {
		case <-p.done:
			// Everything is fine
		case <-time.After(Timeout):
			msg := "PubSocket failed to clean up"
			if p.s != nil {
				if err := p.s.Close(); err != nil {
					msg += fmt.Sprintf(", Socket close error: %v", err)
				}
			}
			return errors.New(msg)
		}
	}

	return nil
}

func (p *PubSocket) Channels() (chan<- []byte, <-chan error) {
	return p.c, p.e
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
	s, err := getPubSocket(p)
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

	s, err := getPubSocket(p)
	switch {
	case err != nil:
		return nil, fmt.Errorf("PubIPC failed: %s", err)
	case s == nil:
		return nil, errors.New("PubIPC requires a non-nil Socket, use Pub first")
	}

	s.AddTransport(ipc.NewTransport())

	return p, nil
}

// PubBuffer sets the size of the socket buffer for a mangos Publisher.  This
// should almost never be necessary.
func PubBuffer(size int) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		if size <= 0 {
			return nil, fmt.Errorf("PubBuffer expects positive size, got %d", size)
		}

		if tP, ok := p.(*PubSocket); ok {
			tP.buffSize = size
			return tP, nil
		}

		return nil, fmt.Errorf("PubBuffer expected *mangos.PubSocket, got %T", p)
	}
}

// getPubSocket gets a Mangos pub.Socket from a queue.Publisher containing a
// Mangos Socket.
func getPubSocket(p queue.Publisher) (mangos.Socket, error) {
	if tP, ok := p.(*PubSocket); ok {
		return tP.s, nil
	}

	return nil, fmt.Errorf("getPubSocket expected *mangos.PubSocket, got %T", p)
}
