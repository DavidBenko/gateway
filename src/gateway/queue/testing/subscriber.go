package testing

import (
	"errors"
	"gateway/queue"
)

var _ = queue.Client(&Client{})
var _ = queue.Subscriber(&Subscriber{})

type Client struct {
	connectErr error
}

func (c *Client) Connect(path string) error {
	return c.connectErr
}

func (c *Client) Close() error {
	return nil
}

type Subscriber struct {
	Client
}

func (s *Subscriber) Channel() <-chan []byte {
	return nil
}

func SubBindingOk(s queue.Subscriber) (queue.Subscriber, error) {
	return &Subscriber{}, nil
}

func SubBindingErr(err string) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		return nil, errors.New(err)
	}
}

func SubBindingConnectErr(err string) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		tS := s.(*Subscriber)
		tS.connectErr = errors.New(err)
		return tS, nil
	}
}
