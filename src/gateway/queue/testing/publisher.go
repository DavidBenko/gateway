package testing

import (
	"errors"
	"gateway/queue"
)

var _ = queue.Server(&Server{})
var _ = queue.Publisher(&Publisher{})

type Server struct {
	bindFn     func(string) error
	closeFn    func() error
	closeError error
}

func (s *Server) Bind(path string) error {
	return s.bindFn(path)
}

func (s *Server) Close() error {
	return s.closeFn()
}

type Publisher struct {
	Server
	c chan []byte
}

func (p *Publisher) Channel() chan []byte {
	return p.c
}

func PubBindingOk(p queue.Publisher) (queue.Publisher, error) {
	if p != nil {
		return nil, errors.New("tried to bind to non-nil Publisher")
	}

	p = &Publisher{
		Server: Server{bindFn: func(s string) error { return nil }},
		c:      make(chan []byte),
	}

	return p, nil
}

func PubBindingErr(err string) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		return nil, errors.New(err)
	}
}

func PubBindingBindMessages(msgs *[][]byte) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		tP := p.(*Publisher)
		go func() {
			var m []byte
			for m = range tP.c {
				*msgs = append(*msgs, m)
			}
		}()

		return tP, nil
	}
}

func PubBindingBindErr(err string) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		tP := p.(*Publisher)
		tP.bindFn = func(path string) error {
			return errors.New(err)
		}
		return tP, nil
	}
}

func PubBindingCloseError(err string) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		tP := p.(*Publisher)
		tP.closeError = errors.New(err)
		return tP, nil
	}
}

func PubBindingCloseChan(reply *chan struct{}) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		tP := p.(*Publisher)

		tP.closeFn = func() error {
			*reply <- struct{}{}
			return tP.closeError
		}
		return tP, nil
	}
}
