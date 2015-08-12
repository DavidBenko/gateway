package pools

import (
	"gateway/db"
	"sync"
)

// sqlPool implements pools.ServerPool for SQL databases.
type sqlPool struct {
	dbs map[string]db.DB
	sync.RWMutex
}

// Get implements ServerPool.Get.
func (s *sqlPool) Get(spec db.Specifier) (db.DB, bool) {
	d, ok := s.dbs[spec.UniqueServer()]
	return d, ok
}

// Put implements ServerPool.Put.
func (s *sqlPool) Put(spec db.Specifier, d db.DB) {
	s.dbs[spec.UniqueServer()] = d
}

// Delete implements ServerPool.Delete.
func (s *sqlPool) Delete(spec db.Specifier) {
	delete(s.dbs, spec.UniqueServer())
}

// Iterator implements ServerPool.Iterator.
func (s *sqlPool) Iterator() <-chan db.Specifier {
	s.RLock()
	iter := make(chan db.Specifier, len(s.dbs))
	for _, d := range s.dbs {
		iter <- d.Spec()
	}
	close(iter)
	s.RUnlock()
	return iter
}

// makeSqlPool makes a sqlPool with a prepared connection map.
func makeSqlPool() *sqlPool {
	return &sqlPool{dbs: make(map[string]db.DB)}
}
