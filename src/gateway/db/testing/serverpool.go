package testing

import (
	"gateway/db"
	"sync"
)

type ServerPool struct {
	DBs map[string]db.DB
	sync.RWMutex
}

func (s *ServerPool) Get(spec db.Specifier) (db.DB, bool) {
	d, ok := s.DBs[spec.UniqueServer()]
	return d, ok
}

func (s *ServerPool) Put(spec db.Specifier, d db.DB) {
	s.DBs[spec.UniqueServer()] = d
}

func (s *ServerPool) Delete(spec db.Specifier) {
	delete(s.DBs, spec.UniqueServer())
}
