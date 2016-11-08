package mangos

import (
	"errors"
	"fmt"
	"gateway/queue"
	"runtime"
	"time"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

var _ = queue.Publisher(&PubSocket{})

// PubSocket implements queue.Publisher.
type PubSocket struct {
	s       mangos.Socket
	control chan signal
	done    chan signal
	c       chan []byte
	e       chan error

	// PubBinding options
	raw       bool
	useBroker bool
}

// Bind is part of queue.Server.
func (p *PubSocket) Bind(path string) error {
	s := p.s

	if s == nil {
		return fmt.Errorf("mangos PubSocket couldn't Bind to %s: nil socket", path)
	}
	if err := s.SetOption(mangos.OptionBestEffort, true); err != nil {
		return fmt.Errorf("mangos PubSocket couldn't set BestEffort option: %s", err.Error())
	}

	switch {
	case p.raw:
		// In case of a raw channel, don't start the send loop.
		return s.Listen(path)
	case p.useBroker:
		if err := s.SetOption(mangos.OptionMaxReconnectTime, 5*time.Minute); err != nil {
			return fmt.Errorf("mangos broker PubSocket dial couldn't set MaxReconnectTime option: %s", err.Error())
		}
		if err := s.Dial(path); err != nil {
			return err
		}
	default:
		if err := s.Listen(path); err != nil {
			return err
		}
	}

	control := make(chan signal)
	done := make(chan signal)
	c := make(chan []byte, channelSize)
	e := make(chan error, channelSize)

	p.c = c
	p.e = e
	p.control = control
	p.done = done

	go sendLoop(s, c, e, control, done)
	return nil
}

func sendLoop(
	s mangos.Socket,
	c chan []byte,
	e chan error,
	control chan signal,
	done chan signal,
) {
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
}

// Close is part of queue.Server.
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

// Channels is part of queue.Publisher.  When used with XPub, these will be nil.
func (p *PubSocket) Channels() (chan<- []byte, <-chan error) {
	return p.c, p.e
}

// Pub is a queue.PubBinding which creates a new mangos PubSocket.  Use XPub
// instead to create an XPUB endpoint.
//
// To use Pub with an XPUB/XSUB broker, use Pub(true).  Otherwise, use
// Pub(false).
func Pub(brokered bool) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		if p != nil {
			return nil, fmt.Errorf("Pub expects nil Publisher, got %T", p)
		}

		s, err := pub.NewSocket()
		if err != nil {
			return nil, fmt.Errorf("Pub failed to make Mangos Socket: %s", err.Error())
		}

		return &PubSocket{s: s, useBroker: brokered}, nil
	}
}

// XPub creates a mangos pub socket for use as an XPUB endpoint.  Use instead
// of Pub if you want to use XPUB.
// https://raw.githubusercontent.com/imatix/zguide/master/images/fig14.png
func XPub(p queue.Publisher) (queue.Publisher, error) {
	if p != nil {
		return nil, fmt.Errorf("XPub expects nil Publisher, got %T", p)
	}

	s, err := pub.NewSocket()
	if err != nil {
		return nil, fmt.Errorf("XPub failed to make Mangos Socket: %s", err.Error())
	}

	if err := s.SetOption(mangos.OptionRaw, true); err != nil {
		return nil, err
	}

	return &PubSocket{s: s, raw: true}, nil
}

// PubTCP is a queue.PubBinding which adds a TCP binding to the PubSocket.
func PubTCP(p queue.Publisher) (queue.Publisher, error) {
	s, err := getPubSocket(p)
	switch {
	case err != nil:
		return nil, fmt.Errorf("PubTCP failed: %s", err)
	case s == nil:
		return nil, errors.New("PubTCP requires a non-nil Socket, use Pub or XPub first")
	}

	s.AddTransport(tcp.NewTransport())

	return p, nil
}

// PubIPC is a queue.PubBinding which adds a IPC binding to the PubSocket.
func PubIPC(p queue.Publisher) (queue.Publisher, error) {
	// https://github.com/go-mangos/mangos/issues/2
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
		return nil, errors.New("PubIPC requires a non-nil Socket, use Pub or XPub first")
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

		s, err := getPubSocket(p)
		if err != nil {
			return nil, err
		}

		if s == nil {
			return nil, fmt.Errorf("PubBuffer expects non-nil socket, use Pub or XPub first")
		}

		if err := s.SetOption(mangos.OptionWriteQLen, size); err != nil {
			return nil, err
		}

		return p, nil
	}
}

// getPubSocket gets a Mangos pub.Socket from a queue.Publisher containing a
// Mangos Socket.
func getPubSocket(p queue.Publisher) (mangos.Socket, error) {
	if tP, ok := p.(*PubSocket); ok {
		return tP.s, nil
	}

	return nil, fmt.Errorf("getPubSocket expects *mangos.PubSocket, got %T", p)
}
