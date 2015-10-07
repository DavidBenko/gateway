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
		comm: c.comm,
		err:  make(chan error, 1),
	}
	cmdChan <- cmd
	if err := <-cmd.err; err != nil {
		return err
	}
	c.path = path
	return nil
}

func (c *Client) Close() error {
	if c.path == "" {
		return fmt.Errorf("the subscriber isn't connected")
	}

	cmd := command{
		cmd:  COMMAND_CLOSE_SUB,
		path: c.path,
		comm: c.comm,
		err:  make(chan error, 1),
	}
	cmdChan <- cmd
	return <-cmd.err
}

type Subscriber struct {
	Client
}

func (s *Subscriber) Channels() (<-chan []byte, <-chan error) {
	return s.comm, make(chan error)
}

func Subscribe(s queue.Subscriber) (queue.Subscriber, error) {
	subscriber := &Subscriber{}
	subscriber.comm = make(chan []byte, 8)
	return subscriber, nil
}
