package queue

import "io"

type subCloser struct {
	c chan struct{}
	Subscriber
}

func (s *subCloser) Close() error {
	return safeClose(s.c)
}

type pubCloser struct {
	c chan struct{}
	Publisher
}

func (p *pubCloser) Close() error {
	return safeClose(p.c)
}

func closeFunc(c io.Closer) func() error {
	return func() error {
		return c.Close()
	}
}

func closeCloser(c io.Closer) func(io.Closer) error {
	return func(cl io.Closer) error {
		return cl.Close()
	}
}

func waitFunc(ch chan struct{}, fn func() error) {
	for _ = range ch {
	}
	fn()
}

func safeClose(ch chan struct{}) error {
	select {
	case _, ok := <-ch:
		if ok {
			close(ch)
		}
	default:
		close(ch)
	}

	return nil
}
