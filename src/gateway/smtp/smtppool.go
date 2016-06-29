package smtp

import (
	"sync"
)

type SmtpPool struct {
	sync.RWMutex
	pool map[string]Mailer
}

func NewSmtpPool() *SmtpPool {
	return &SmtpPool{pool: make(map[string]Mailer)}
}

func (p *SmtpPool) Connection(spec *Spec) (Mailer, error) {
	p.RLock()
	connection := p.pool[spec.ConnectionString()]
	p.RUnlock()

	if connection != nil {
		return connection, nil
	}

	p.Lock()
	defer p.Unlock()
	p.pool[spec.ConnectionString()] = spec
	return spec, nil
}
