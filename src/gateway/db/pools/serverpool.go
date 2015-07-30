package pools

import (
	"gateway/db"
	"sync"
)

// serverPool implements pools.ServerPool
type serverPool struct {
	dbs map[string]db.DB
	sync.RWMutex
}

func (s *serverPool) Get(spec db.Specifier) (db.DB, bool) {
	d, ok := s.dbs[spec.UniqueServer()]
	return d, ok
}

func (s *serverPool) Put(spec db.Specifier, d db.DB) {
	s.dbs[spec.UniqueServer()] = d
}

func (s *serverPool) Delete(spec db.Specifier) {
	delete(s.dbs, spec.UniqueServer())
}

// makeServerPool makes a ServerPool with a prepared connection map.
func makeServerPool() *serverPool {
	return &serverPool{dbs: make(map[string]db.DB)}
}
