package testing

import (
	"errors"
	"gateway/queue"
)

var _ = queue.Server(&Server{})
var _ = queue.Publisher(&Publisher{})

type Server struct {
	bindErr error
}

func (s *Server) Bind(path string) error {
	return s.bindErr
}

func (s *Server) Close() error {
	return nil
}

type Publisher struct {
	Server
}

func (p *Publisher) Channels() (chan<- []byte, <-chan error) {
	return nil, nil
}

func PubBindingOk(p queue.Publisher) (queue.Publisher, error) {
	if p != nil {
		return nil, errors.New("tried to bind to non-nil Publisher")
	}

	return &Publisher{}, nil
}

func PubBindingErr(err string) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		return nil, errors.New(err)
	}
}

func PubBindingBindErr(err string) queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		tP := p.(*Publisher)
		tP.bindErr = errors.New(err)
		return tP, nil
	}
}
