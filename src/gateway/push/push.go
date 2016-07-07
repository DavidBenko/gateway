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

	pushChannelMessagePayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	pushChannelMessage := &model.PushChannelMessage{
		AccountID:     accountID,
		PushChannelID: channel.ID,
		Stamp:         time.Now().Unix(),
		Data:          pushChannelMessagePayload,
	}

	err = pushChannelMessage.Insert(tx)

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
			payload := struct{ Error string }{err.Error()}
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
			AccountID:            accountID,
			PushChannelID:        channel.ID,
			PushDeviceID:         device.ID,
			PushChannelMessageID: pushChannelMessage.ID,
			Stamp:                time.Now().Unix(),
			Data:                 data,
		}
		err = pushMessage.Insert(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PushPool) Subscribe(platforms *re.Push, tx *apsql.Tx, accountID, apiID, remoteEndpointID int64, platformName string, channelName string, period int64, deviceName string, token string) error {
	endpoint, err := model.FindRemoteEndpointForAPIIDAndAccountID(tx.DB, remoteEndpointID, apiID, accountID)
	if err != nil {
		return err
	}
	found := false
	for _, platform := range platforms.PushPlatforms {
		if platform.Codename == platformName {
			found = true
		}
	}
	for _, environment := range endpoint.EnvironmentData {
		push := &re.Push{}
		err = json.Unmarshal(environment.Data, push)
		if err != nil {
			return err
		}
		for _, platform := range push.PushPlatforms {
			if platform.Codename == platformName {
				found = true
			}
		}
	}
	if !found {
		return fmt.Errorf("%v is not a valid platform", platformName)
	}

	channel := &model.PushChannel{
		AccountID:        accountID,
		APIID:            apiID,
		RemoteEndpointID: endpoint.ID,
		Name:             channelName,
	}
	_channel, err := channel.Find(tx.DB)
	expires := time.Now().Unix() + period
	if err != nil {
		channel.Expires = expires
		err := channel.Insert(tx)
		if err != nil {
			return err
		}
	} else {
		channel = _channel
		if channel.Expires < expires {
			channel.Expires = expires
			err := channel.Update(tx)
			if err != nil {
				return err
			}
		}
	}

	device := &model.PushDevice{
		AccountID:        accountID,
		PushChannelID:    channel.ID,
		Token:            token,
		Name:             deviceName,
		RemoteEndpointID: endpoint.ID,
	}
	dev, err := device.Find(tx.DB)
	update := false
	if err != nil {
		device.Name = deviceName
		device.Type = platformName
		device.Expires = expires
		err = device.Insert(tx)
		if err != nil {
			return err
		}
	} else {
		update = true
	}
	if update {
		dev.PushChannelID = channel.ID
		dev.Name = deviceName
		dev.Expires = expires
		err := dev.Update(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PushPool) Unsubscribe(platforms *re.Push, tx *apsql.Tx, accountID, apiID, remoteEndpointID int64, platformName string, channelName string, token string) error {
	endpoint, err := model.FindRemoteEndpointForAPIIDAndAccountID(tx.DB, remoteEndpointID, apiID, accountID)
	if err != nil {
		return err
	}

	channel := &model.PushChannel{
		AccountID:        accountID,
		APIID:            apiID,
		RemoteEndpointID: endpoint.ID,
		Name:             channelName,
	}
	channel, err = channel.Find(tx.DB)
	if err != nil {
		return err
	}

	device := &model.PushDevice{
		AccountID:        accountID,
		PushChannelID:    channel.ID,
		Token:            token,
		RemoteEndpointID: endpoint.ID,
	}
	dev, err := device.Find(tx.DB)
	if err != nil {
		return err
	}
	err = dev.DeleteFromChannel(tx)
	if err != nil {
		return err
	}

	return nil
}
