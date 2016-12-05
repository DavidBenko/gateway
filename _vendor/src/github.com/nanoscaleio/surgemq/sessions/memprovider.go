// Copyright (c) 2014 The SurgeMQ Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sessions

import (
	"fmt"
	"sync"

	"github.com/nanoscaleio/surgemq/message"
)

var _ SessionsProvider = (*memProvider)(nil)

func init() {
	Register("mem", NewMemProvider)
}

type memSessionTopics struct {
	// topics stores all the topis for this session/client
	topics map[string]byte

	// Initialized?
	initted bool

	// Serialize access to this session
	mu sync.Mutex
}

func (this *memSessionTopics) InitTopics(msg *message.ConnectMessage) error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.initted {
		return fmt.Errorf("SessionTopics already initialized")
	}

	this.topics = make(map[string]byte, 1)
	this.initted = true

	return nil
}

func (this *memSessionTopics) AddTopic(topic string, qos byte) error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if !this.initted {
		return fmt.Errorf("Session not yet initialized")
	}

	this.topics[topic] = qos

	return nil
}

func (this *memSessionTopics) RemoveTopic(topic string) error {
	this.mu.Lock()
	defer this.mu.Unlock()

	if !this.initted {
		return fmt.Errorf("Session not yet initialized")
	}

	delete(this.topics, topic)

	return nil
}

func (this *memSessionTopics) Topics() ([]string, []byte, error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	if !this.initted {
		return nil, nil, fmt.Errorf("Session not yet initialized")
	}

	var (
		topics []string
		qoss   []byte
	)

	for k, v := range this.topics {
		topics = append(topics, k)
		qoss = append(qoss, v)
	}

	return topics, qoss, nil
}

type memProvider struct {
	st      map[string]*Session
	context fmt.Stringer
	mu      sync.RWMutex
}

func NewMemProvider(context fmt.Stringer) SessionsProvider {
	return &memProvider{
		st:      make(map[string]*Session),
		context: context,
	}
}

func (this *memProvider) New(id string) (*Session, error) {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.st[id] = &Session{Id: id, SessionTopics: &memSessionTopics{}}
	return this.st[id], nil
}

func (this *memProvider) Get(id string) (*Session, error) {
	this.mu.RLock()
	defer this.mu.RUnlock()

	sess, ok := this.st[id]
	if !ok {
		return nil, fmt.Errorf("store/Get: No session found for key %s", id)
	}

	return sess, nil
}

func (this *memProvider) Del(id string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.st, id)
}

func (this *memProvider) Save(id string, session *Session) error {
	return nil
}

func (this *memProvider) Count() int {
	return len(this.st)
}

func (this *memProvider) Close() error {
	this.st = make(map[string]*Session)
	return nil
}
