//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package inmem

import (
	"sync"

	"github.com/blevesearch/bleve/index/store"
	"github.com/blevesearch/bleve/registry"
	"github.com/ryszard/goskiplist/skiplist"
)

const Name = "mem"

type Store struct {
	list   *skiplist.SkipList
	writer sync.Mutex
	mo     store.MergeOperator
}

func New() (*Store, error) {
	rv := Store{
		list: skiplist.NewStringMap(),
	}

	return &rv, nil
}

func MustOpen() *Store {
	rv := Store{
		list: skiplist.NewStringMap(),
	}

	return &rv
}

func (i *Store) Open() error {
	return nil
}

func (i *Store) SetMergeOperator(mo store.MergeOperator) {
	i.mo = mo
}

func (i *Store) get(key []byte) ([]byte, error) {
	val, ok := i.list.Get(string(key))
	if ok {
		return []byte(val.(string)), nil
	}
	return nil, nil
}

func (i *Store) set(key, val []byte) error {
	i.writer.Lock()
	defer i.writer.Unlock()
	return i.setlocked(key, val)
}

func (i *Store) setlocked(key, val []byte) error {
	i.list.Set(string(key), string(val))
	return nil
}

func (i *Store) delete(key []byte) error {
	i.writer.Lock()
	defer i.writer.Unlock()
	return i.deletelocked(key)
}

func (i *Store) deletelocked(key []byte) error {
	i.list.Delete(string(key))
	return nil
}

func (i *Store) Close() error {
	return nil
}

func (i *Store) iterator(key []byte) store.KVIterator {
	rv := newIterator(i)
	rv.Seek(key)
	return rv
}

func (i *Store) Reader() (store.KVReader, error) {
	return newReader(i)
}

func (i *Store) Writer() (store.KVWriter, error) {
	return newWriter(i)
}

func StoreConstructor(config map[string]interface{}) (store.KVStore, error) {
	return New()
}

func init() {
	registry.RegisterKVStore(Name, StoreConstructor)
}
