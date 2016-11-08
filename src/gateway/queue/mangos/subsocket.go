package mangos

import (
	"errors"
	"fmt"
	"gateway/queue"
	"runtime"
	"time"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

var _ = queue.Subscriber(&SubSocket{})

// SubSocket implements queue.Subscriber.
type SubSocket struct {
	s       mangos.Socket
	c       chan []byte
	e       chan error
	control chan signal
	done    chan signal

	// SubBinding options
	filter []byte
	raw    bool
}

// Connect is part of queue.Client.
func (s *SubSocket) Connect(path string) error {
	sock := s.s

	if sock == nil {
		return fmt.Errorf("SubSocket couldn't Connect to %s: nil socket", path)
	}

	if err := sock.SetOption(mangos.OptionSubscribe, s.filter); err != nil {
		return err
	}

	switch {
	case s.raw:
		// If we're in raw mode, don't start the receive loop
		return sock.Listen(path)
	default:
		if err := sock.SetOption(mangos.OptionMaxReconnectTime, 5*time.Minute); err != nil {
			return fmt.Errorf("mangos SubSocket dial couldn't set MaxReconnectTime option: %s", err.Error())
		}
		if err := sock.Dial(path); err != nil {
			return err
		}
	}

	control := make(chan signal)
	done := make(chan signal)
	c := make(chan []byte, channelSize)
	e := make(chan error, channelSize)

	s.c = c
	s.e = e
	s.control = control
	s.done = done

	go receiveLoop(sock, c, e, control, done)

	return nil
}

func receiveLoop(
	sock mangos.Socket,
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
			close(c)
			close(e)
			close(done)
			return
		default:
			if msg, err = sock.Recv(); err != nil {
				e <- err
			} else {
				c <- msg
			}
		}
	}
}

// Close is part of queue.Client.
func (s *SubSocket) Close() (e error) {
	if s.control != nil {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					e = err
				}
				e = fmt.Errorf("unexpected panic: %v", e)
			}
		}()
		close(s.control)
	}

	if s.s != nil {
		if err := s.s.Close(); err != nil {
			return err
		}
	}

	// If nil, Connect was never called and we can cleanly close.
	// Otherwise, wait for it to clean up.
	if s.done != nil {
		select {
		case <-s.done:
			// Everything is fine
		case <-time.After(Timeout):
			msg := "SubSocket failed to clean up"
			if s.s != nil {
				if err := s.s.Close(); err != nil {
					msg += fmt.Sprintf(", Socket close error: %v", err)
				}
			}
			return errors.New(msg)
		}
	}

	return nil
}

// Channels is part of queue.Subscriber.  When used with XSub, these will be
// nil.
func (s *SubSocket) Channels() (<-chan []byte, <-chan error) {
	return s.c, s.e
}

// Sub is a queue.SubBinding which creates a new mangos SubSocket.  Use XSub
// instead to create an XSUB endpoint.
func Sub(s queue.Subscriber) (queue.Subscriber, error) {
	if s != nil {
		return nil, fmt.Errorf("Sub expects nil Subscriber, got %T", s)
	}

	sock, err := sub.NewSocket()
	if err != nil {
		return nil, fmt.Errorf("Sub failed to make Mangos Socket: %s", err.Error())
	}

	return &SubSocket{s: sock, filter: []byte("")}, nil
}

// XSub is a SubBinding which creates a mangos pub socket for use as an XSUB
// endpoint.  Use instead of Sub if you want to use XSUB.
// https://raw.githubusercontent.com/imatix/zguide/master/images/fig14.png
func XSub(s queue.Subscriber) (queue.Subscriber, error) {
	if s != nil {
		return nil, fmt.Errorf("XSub expects nil Subscriber, got %T", s)
	}

	sock, err := sub.NewSocket()
	if err != nil {
		return nil, fmt.Errorf("XSub failed to make Mangos Socket: %s", err.Error())
	}

	if err = sock.SetOption(mangos.OptionRaw, true); err != nil {
		return nil, fmt.Errorf("XSub failed to set OptionRaw: %s", err.Error())
	}

	return &SubSocket{s: sock, filter: []byte(""), raw: true}, nil
}

// SubTCP is a queue.SubBinding which adds a TCP binding to the SubSocket.
func SubTCP(s queue.Subscriber) (queue.Subscriber, error) {
	sock, err := getSubSocket(s)
	switch {
	case err != nil:
		return nil, fmt.Errorf("SubTCP failed: %s", err)
	case sock == nil:
		return nil, errors.New("SubTCP requires a non-nil Socket, use Sub or XSub first")
	}

	sock.AddTransport(tcp.NewTransport())

	return s, nil
}

// SubIPC is a queue.SubBinding which adds a IPC binding to the SubSocket.
func SubIPC(s queue.Subscriber) (queue.Subscriber, error) {
	// https://github.com/go-mangos/mangos/issues/2
	switch runtime.GOOS {
	case "linux", "darwin":
		// Unix domain sockets are supported on Linux and Darwin
	default:
		return nil, fmt.Errorf("SubIPC failed: IPC transport not supported on OS %q", runtime.GOOS)
	}

	sock, err := getSubSocket(s)
	switch {
	case err != nil:
		return nil, fmt.Errorf("SubIPC failed: %s", err)
	case sock == nil:
		return nil, errors.New("SubIPC requires a non-nil Socket, use Sub or XSub first")
	}

	sock.AddTransport(ipc.NewTransport())

	return s, nil
}

// SubBuffer sets the size of the socket buffer for a mangos Subscriber.  This
// should almost never be necessary.
func SubBuffer(size int) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		if size <= 0 {
			return nil, fmt.Errorf("SubBuffer expects positive size, got %d", size)
		}

		sS, err := getSubSocket(s)
		if err != nil {
			return nil, err
		}

		if sS == nil {
			return nil, fmt.Errorf("SubBuffer expects non-nil socket, use Sub or XSub first")
		}

		if err := sS.SetOption(mangos.OptionReadQLen, size); err != nil {
			return nil, err
		}

		return s, nil
	}
}

// Filter sets a string for the SubSocket to subscribe to, i.e. it will ignore
// messages except ones that begin with the string.
func Filter(filter string) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		if s == nil {
			return nil, errors.New("Filter got nil Subscriber, use Pub first")
		}
		if tS, ok := s.(*SubSocket); ok {
			if tS == nil {
				return nil, errors.New("Filter got nil Subscriber, use Pub first")
			}

			if tS.raw {
				return nil, errors.New("raw SubSocket doesn't support Filter")
			}

			tS.filter = []byte(filter)
			return tS, nil
		}

		return nil, fmt.Errorf("Filter expected *SubSocket, got %T", s)
	}
}

// Gets a Mangos sub.Socket from a queue.Subscriber containing a Mangos Socket.
func getSubSocket(s queue.Subscriber) (mangos.Socket, error) {
	if tS, ok := s.(*SubSocket); ok {
		return tS.s, nil
	}

	return nil, fmt.Errorf("getSubSocket expected *SubSocket, got %T", s)
}
