package push

import (
	"encoding/json"
	"sync"

	"gateway/logreport"
	re "gateway/model/remote_endpoint"
)

type Pusher interface {
	Push(token string, data interface{}) error
}

type PushPool struct {
	sync.RWMutex
	pool map[string]Pusher
}

func NewPushPool() *PushPool {
	pool := &PushPool{
		pool: make(map[string]Pusher),
	}
	return pool
}

func (p *PushPool) Connection(platform *re.PushPlatform) Pusher {
	spec, err := json.Marshal(platform)
	if err != nil {
		logreport.Fatal(err)
	}
	p.RLock()
	pusher := p.pool[string(spec)]
	p.RUnlock()
	if pusher != nil {
		return pusher
	}

	switch platform.Type {
	case re.PushTypeOSX:
		fallthrough
	case re.PushTypeIOS:
		pusher = NewApplePusher(platform)
	case re.PushTypeGCM:
		pusher = NewGooglePusher(platform)
	}
	p.Lock()
	defer p.Unlock()
	p.pool[string(spec)] = pusher
	return pusher
}
