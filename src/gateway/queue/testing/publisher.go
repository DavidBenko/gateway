package testing

import (
	"fmt"
	"gateway/queue"
)

var _ = queue.Server(&Server{})
var _ = queue.Publisher(&Publisher{})

type Server struct {
}

func (s *Server) Bind(path string) error {
	return fmt.Errorf("implement me!")
}

type Publisher struct {
	Server
}

func (p *Publisher) Channel() chan []byte {
	return nil
}

func (p *Publisher) Close() error {
	return fmt.Errorf("implement me!")
}

func Publish() queue.PubBinding {
	return func(p queue.Publisher) (queue.Publisher, error) {
		return nil, fmt.Errorf("implement me!")
	}
}
