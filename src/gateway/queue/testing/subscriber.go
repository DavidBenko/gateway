package testing

import (
	"errors"
	"gateway/queue"
)

var _ = queue.Client(&Client{})
var _ = queue.Subscriber(&Subscriber{})

type Client struct {
	connectFn  func(string) error
	closeFn    func() error
	closeError error
}

func (c *Client) Connect(path string) error {
	return c.connectFn(path)
}

func (c *Client) Close() error {
	return c.closeFn()
}

type Subscriber struct {
	Client
	c chan []byte
}

func (s *Subscriber) Channel() chan []byte {
	return s.c
}

func SubBindingOk(s queue.Subscriber) (queue.Subscriber, error) {
	s = &Subscriber{
		Client: Client{connectFn: func(s string) error { return nil }},
		c:      make(chan []byte),
	}
	return s, nil
}

func SubBindingErr(err string) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		return nil, errors.New(err)
	}
}

func SubBindingConnectErr(err string) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		tS := s.(*Subscriber)
		tS.connectFn = func(path string) error {
			return errors.New(err)
		}
		return tS, nil
	}
}

func SubBindingCloseChan(reply chan struct{}) queue.SubBinding {
	return func(s queue.Subscriber) (queue.Subscriber, error) {
		tS := s.(*Subscriber)
		tS.closeFn = func() error {
			reply <- struct{}{}
			return tS.closeError
		}
		return tS, nil
	}
}
