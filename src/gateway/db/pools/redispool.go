package pools

import (
	"gateway/db"
	"gateway/db/redis"
	"sync"
)

// redisPool implements pools.ServerPool for Redis.
type redisPool struct {
	dbs map[string]*redis.DB
	sync.RWMutex
}

func (p *redisPool) Get(spec db.Specifier) (db.DB, bool) {
	if d, ok := p.dbs[spec.UniqueServer()]; ok {
		return d, ok
	}
	return nil, false
}

func (p *redisPool) Put(spec db.Specifier, d db.DB) {
	p.dbs[spec.UniqueServer()] = d.(*redis.DB)
}

func (p *redisPool) Delete(spec db.Specifier) {
	delete(p.dbs, spec.UniqueServer())
}

// Iterator implements ServerPool.Iterator.
func (p *redisPool) Iterator() <-chan db.Specifier {
	iter := make(chan db.Specifier, len(p.dbs))
	for _, d := range p.dbs {
		iter <- d.Spec()
	}
	close(iter)
	return iter
}

func makeRedisPool() *redisPool {
	return &redisPool{dbs: make(map[string]*redis.DB)}
}
