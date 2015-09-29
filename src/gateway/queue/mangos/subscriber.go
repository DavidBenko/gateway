package mangos

import (
	"errors"
	"fmt"
	"gateway/queue"
	"runtime"

	"github.com/gdamore/mangos"
	"github.com/gdamore/mangos/protocol/sub"
	"github.com/gdamore/mangos/transport/ipc"
	"github.com/gdamore/mangos/transport/tcp"
)

var _ = queue.Subscriber(&SubSocket{})

// SubSocket implements queue.Subscriber.
type SubSocket struct {
	s       mangos.Socket
	filter  []byte
	c       chan []byte
	e       chan error
	control chan signal
}

// Connect implements queue.Subscriber.Connect for *SubSocket.
func (s *SubSocket) Connect(path string) error {
	if s.s == nil {
		return fmt.Errorf("mangos Subscriber couldn't Connect to %s: nil socket", path)
	}

	sock := s.s

	if err := sock.SetOption(mangos.OptionSubscribe, s.filter); err != nil {
		return err
	}

	control := make(chan signal)
	c := make(chan []byte, numChannels)
	e := make(chan error, numChannels)

	go func() {
		var msg []byte
		var err error
		for {
			select {
			case <-control:
				close(c)
				close(e)
				return
			default:
				msg, err = sock.Recv()
				switch {
				case err != nil:
					e <- err
				default:
					c <- msg
				}
			}
		}
	}()

	s.c = c
	s.e = e
	s.control = control

	return sock.Dial(path)
}

// Close implements io.Closer for *SubSocket.
func (s *SubSocket) Close() error {
	select {
	case s.control <- struct{}{}:
		// control was not yet closed, so we can safely close it.
		close(s.control)
	default:
	}

	if s.s != nil {
		return s.s.Close()
	}

	return nil
}

// Channel returns a handle to the SubSocket's underlying channel.
func (s *SubSocket) Channel() <-chan []byte {
	return s.c
}

// Sub is a queue.SubBinding which creates a new mangos SubSocket.
func Sub(s queue.Subscriber) (queue.Subscriber, error) {
	if s != nil {
		return nil, fmt.Errorf("mangos.Sub expects nil Subscriber, got %T", s)
	}

	sock, err := sub.NewSocket()
	if err != nil {
		return nil, fmt.Errorf("mangos.Sub failed to make Mangos Socket: %s", err.Error())
	}

	return &SubSocket{s: sock, filter: []byte("")}, nil
}

// SubTCP is a queue.SubBinding which adds a TCP binding to the SubSocket.
func SubTCP(s queue.Subscriber) (queue.Subscriber, error) {
	sock, err := GetSubSocket(s)
	switch {
	case err != nil:
		return nil, fmt.Errorf("SubTCP failed: %s", err)
	case sock == nil:
		return nil, errors.New("SubTCP requires a non-nil Socket, use Sub first")
	}

	sock.AddTransport(tcp.NewTransport())

	return s, nil
}

// SubIPC is a queue.SubBinding which adds a IPC binding to the SubSocket.
func SubIPC(s queue.Subscriber) (queue.Subscriber, error) {
	// https://github.com/gdamore/mangos/issues/2
	switch runtime.GOOS {
	case "linux", "darwin":
		// Unix domain sockets are supported on Linux and Darwin
	default:
		return nil, fmt.Errorf("SubIPC failed: mangos IPC transport not supported on OS %q", runtime.GOOS)
	}

	sock, err := GetSubSocket(s)
	switch {
	case err != nil:
		return nil, fmt.Errorf("SubIPC failed: %s", err)
	case sock == nil:
		return nil, errors.New("SubIPC requires a non-nil Socket, use Sub first")
	}

	sock.AddTransport(ipc.NewTransport())

	return s, nil
}

func Filter(filter string) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		switch tS := s.(type) {
		case *SubSocket:
			tS.filter = []byte(filter)
			return tS, nil
		default:
			return nil, fmt.Errorf("Filter expected *mangos.SubSocket, got %T", s)
		}
	}
}

// Gets a Mangos sub.Socket from a queue.Subscriber containing a Mangos Socket.
func GetSubSocket(s queue.Subscriber) (mangos.Socket, error) {
	if tS, ok := s.(*SubSocket); ok {
		return tS.s, nil
	}

	return nil, fmt.Errorf("GetSubSocket expected *mangos.SubSocket, got %T", s)
}
