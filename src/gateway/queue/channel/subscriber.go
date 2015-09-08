package channel

import (
	"fmt"
	"gateway/queue"
)

var _ = queue.Client(&Client{})
var _ = queue.Subscriber(&Subscriber{})

type Client struct {
	path string
	comm chan []byte
}

func (c *Client) Connect(path string) error {
	cmd := command{
		cmd:  COMMAND_SUB,
		path: path,
		comm: make(chan []byte, 8),
		err:  make(chan error, 1),
	}
	cmdChan <- cmd
	if err := <-cmd.err; err != nil {
		return err
	}
	c.path = path
	c.comm = cmd.comm
	return nil
}

type Subscriber struct {
	Client
}

func (s *Subscriber) Channel() chan []byte {
	return s.comm
}

func (s *Subscriber) Close() error {
	if s.path == "" {
		return fmt.Errorf("the subscriber isn't connected")
	}

	cmd := command{
		cmd:  COMMAND_CLOSE_SUB,
		path: s.path,
		comm: s.comm,
		err:  make(chan error, 1),
	}
	cmdChan <- cmd
	return <-cmd.err
}

func Subscribe() queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		return &Subscriber{}, nil
	}
}
