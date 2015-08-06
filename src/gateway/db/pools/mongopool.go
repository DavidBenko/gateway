package pools

import (
	"gateway/db"
	"gateway/db/mongo"
	"sync"
)

type mongoPool struct {
	dbs map[string]*mongo.DB
	sync.RWMutex
}

func (s *mongoPool) Get(spec db.Specifier) (db.DB, bool) {
	if d, ok := s.dbs[spec.UniqueServer()]; ok {
	  return d.Copy(), ok
  }
  return nil, false
}

func (s *mongoPool) Put(spec db.Specifier, d db.DB) {
	s.dbs[spec.UniqueServer()] = d.(*mongo.DB)
}

func (s *mongoPool) Delete(spec db.Specifier) {
  s.dbs[spec.UniqueServer()].Close()
	delete(s.dbs, spec.UniqueServer())
}

func makeMongoPool() *mongoPool {
	return &mongoPool{dbs: make(map[string]*mongo.DB)}
}
