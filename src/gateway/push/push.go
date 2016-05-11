package push

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gateway/logreport"
	"gateway/model"
	re "gateway/model/remote_endpoint"
	apsql "gateway/sql"
)

type Pusher interface {
	Push(token string, data interface{}) error
}

type PushPoolEntry struct {
	sync.RWMutex
	Pusher
	time.Time
}

type PushPool struct {
	sync.RWMutex
	pool map[string]*PushPoolEntry
}

func NewPushPool() *PushPool {
	pool := &PushPool{
		pool: make(map[string]*PushPoolEntry),
	}
	deleteTicker := time.NewTicker(time.Hour)
	go func() {
		for _ = range deleteTicker.C {
			now := time.Now()
			pool.Lock()
			for key, value := range pool.pool {
				value.RLock()
				if value.Before(now) {
					delete(pool.pool, key)
				}
				value.RUnlock()
			}
			pool.Unlock()
		}
	}()
	return pool
}

func (p *PushPool) Connection(platform *re.PushPlatform) Pusher {
	spec, err := json.Marshal(platform)
	if err != nil {
		logreport.Fatal(err)
	}
	p.RLock()
	entry := p.pool[string(spec)]
	p.RUnlock()
	if entry != nil {
		entry.Lock()
		defer entry.Unlock()
		entry.Time = time.Now().Add(time.Hour)
		return entry
	}

	var pusher Pusher
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
	p.pool[string(spec)] = &PushPoolEntry{Pusher: pusher, Time: time.Now().Add(time.Hour)}
	return pusher
}

func (p *PushPool) Push(platforms *re.Push, tx *apsql.Tx, accountID, apiID, remoteEndpointID int64, name string, payload map[string]interface{}) error {
	channel := &model.PushChannel{
		AccountID:        accountID,
		APIID:            apiID,
		RemoteEndpointID: remoteEndpointID,
		Name:             name,
	}
	channel, err := channel.Find(tx.DB)
	if err != nil {
		return err
	}

	device := &model.PushDevice{
		AccountID:     accountID,
		PushChannelID: channel.ID,
	}
	devices, err := device.All(tx.DB)
	if err != nil {
		return err
	}

	for _, device := range devices {
		err := fmt.Errorf("coulnd't find device platform %v", device.Name)
		var _payload interface{}
		for _, platform := range platforms.PushPlatforms {
			if device.Type == platform.Codename {
				err = nil
				pusher := p.Connection(&platform)
				var ok bool
				_payload, ok = payload[device.Type]
				if !ok {
					_payload = payload["default"]
				}
				err = pusher.Push(device.Token, _payload)
				break
			}
		}
		var data []byte
		if err != nil {
			payload := struct{ err string }{err.Error()}
			var _err error
			data, _err = json.Marshal(&payload)
			if _err != nil {
				return err
			}
		} else {
			var _err error
			data, _err = json.Marshal(_payload)
			if _err != nil {
				return err
			}
		}
		pushMessage := &model.PushMessage{
			AccountID:     accountID,
			PushChannelID: channel.ID,
			PushDeviceID:  device.ID,
			Stamp:         time.Now().Unix(),
			Data:          data,
		}
		err = pushMessage.Insert(tx)
		if err != nil {
			return err
		}
		err = pushMessage.DeleteOffset(tx)
		if err != nil {
			return err
		}
	}

	return nil
}
