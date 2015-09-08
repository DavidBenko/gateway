package channel

import (
	"fmt"
	"gateway/queue"
)

var _ = queue.Server(&Server{})
var _ = queue.Publisher(&Publisher{})

const (
	COMMAND_PUB       = iota
	COMMAND_SUB       = iota
	COMMAND_CLOSE_PUB = iota
	COMMAND_CLOSE_SUB = iota
)

type command struct {
	cmd  int
	path string
	comm chan []byte
	err  chan error
}

func publisher(cmdChan chan command, comm chan []byte) {
	subscribers := make([]chan []byte, 8)
	for {
		select {
		case cmd := <-cmdChan:
			switch cmd.cmd {
			case COMMAND_SUB:
				found := false
				for k, v := range subscribers {
					if v == nil {
						subscribers[k] = cmd.comm
						found = true
						break
					}
				}
				if !found {
					subscribers = append(subscribers, cmd.comm)
				}
			case COMMAND_CLOSE_PUB:
				close(comm)
				for _, subscriber := range subscribers {
					if subscriber != nil {
						close(subscriber)
					}
				}
				return
			case COMMAND_CLOSE_SUB:
				for k, v := range subscribers {
					if v == cmd.comm {
						subscribers[k] = nil
						close(v)
						break
					}
				}
			}
		case message := <-comm:
			for _, subscriber := range subscribers {
				if subscriber != nil {
					subscriber <- message
				}
			}
		}
	}
}

var (
	cmdChan = make(chan command, 8)
)

func init() {
	go func() {
		publishers := make(map[string]chan command)
		for cmd := range cmdChan {
			switch cmd.cmd {
			case COMMAND_PUB:
				if _, has := publishers[cmd.path]; has {
					cmd.err <- fmt.Errorf("publisher for path '%v' already exists", cmd.path)
					break
				}
				cmdChan := make(chan command, 8)
				publishers[cmd.path] = cmdChan
				go publisher(cmdChan, cmd.comm)
				cmd.err <- nil
			case COMMAND_SUB:
				if publisher, ok := publishers[cmd.path]; ok {
					publisher <- cmd
					cmd.err <- nil
				} else {
					cmd.err <- fmt.Errorf("publisher for path '%v' doesn't exist", cmd.path)
				}
			case COMMAND_CLOSE_PUB:
				if publisher, ok := publishers[cmd.path]; ok {
					publisher <- cmd
					delete(publishers, cmd.path)
					cmd.err <- nil
				} else {
					cmd.err <- fmt.Errorf("publisher for path '%v' doesn't exist", cmd.path)
				}
			case COMMAND_CLOSE_SUB:
				if publisher, ok := publishers[cmd.path]; ok {
					publisher <- cmd
					cmd.err <- nil
				} else {
					cmd.err <- fmt.Errorf("publisher for path '%v' doesn't exist", cmd.path)
				}
			}
		}
	}()
}

type Server struct {
	path string
	comm chan []byte
}

func (s *Server) Bind(path string) error {
	cmd := command{
		cmd:  COMMAND_PUB,
		path: path,
		comm: make(chan []byte, 8),
		err:  make(chan error, 1),
	}
	cmdChan <- cmd
	if err := <-cmd.err; err != nil {
		return err
	}

	s.path = path
	s.comm = cmd.comm
	return nil
}

type Publisher struct {
	Server
}

func (p *Publisher) Channel() chan []byte {
	return p.comm
}

func (p *Publisher) Close() error {
	if p.path == "" {
		return fmt.Errorf("the publisher isn't bound")
	}

	cmd := command{
		cmd:  COMMAND_CLOSE_PUB,
		path: p.path,
		err:  make(chan error, 1),
	}
	cmdChan <- cmd
	return <-cmd.err
}

func Publish() queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		return &Publisher{}, nil
	}
}
