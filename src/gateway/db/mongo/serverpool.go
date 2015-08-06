package mongo

import (
	"gateway/db"
  "runtime"
	"sync"
)

type ServerPool struct {
	dbs map[string]*DB
	sync.RWMutex
}

func dbCloser(d *DB) {
  d.Close()
}

func (s *ServerPool) Get(spec db.Specifier) (db.DB, bool) {
	if d, ok := s.dbs[spec.UniqueServer()]; ok {
    dCopy := &DB{d.Session.Copy(), d.conf}
    runtime.SetFinalizer(dCopy, dbCloser)
	  return dCopy, ok
  }
  return nil, false
}

func (s *ServerPool) Put(spec db.Specifier, d db.DB) {
	s.dbs[spec.UniqueServer()] = d.(*DB)
}

func (s *ServerPool) Delete(spec db.Specifier) {
  s.dbs[spec.UniqueServer()].Close()
	delete(s.dbs, spec.UniqueServer())
}

func MakeServerPool() *ServerPool {
	return &ServerPool{dbs: make(map[string]*DB)}
}
