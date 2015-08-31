package testing

import (
	"fmt"
	"gateway/queue"
)

var _ = queue.Client(&Client{})
var _ = queue.Subscriber(&Subscriber{})

type Client struct {
}

func (c *Client) Connect(path string) error {
	return fmt.Errorf("implement me!")
}

type Subscriber struct {
	Client
}

func (s *Subscriber) Channel() chan []byte {
	return nil
}

func (s *Subscriber) Close() error {
	return fmt.Errorf("implement me!")
}

func Subscribe() queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		return nil, fmt.Errorf("implement me!")
	}
}
