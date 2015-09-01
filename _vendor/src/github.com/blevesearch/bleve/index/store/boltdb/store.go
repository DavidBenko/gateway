//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package boltdb

import (
	"fmt"
	"sync"

	"github.com/blevesearch/bleve/index/store"
	"github.com/blevesearch/bleve/registry"
	"github.com/boltdb/bolt"
)

const Name = "boltdb"

type Store struct {
	path   string
	bucket string
	db     *bolt.DB
	writer sync.Mutex
	mo     store.MergeOperator
}

func New(path string, bucket string) *Store {
	rv := Store{
		path:   path,
		bucket: bucket,
	}
	return &rv
}

func (bs *Store) Open() error {

	var err error
	bs.db, err = bolt.Open(bs.path, 0600, nil)
	if err != nil {
		return err
	}

	err = bs.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bs.bucket))

		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (bs *Store) SetMergeOperator(mo store.MergeOperator) {
	bs.mo = mo
}

func (bs *Store) Close() error {
	return bs.db.Close()
}

func (bs *Store) Reader() (store.KVReader, error) {
	tx, err := bs.db.Begin(false)
	if err != nil {
		return nil, err
	}
	return &Reader{
		store: bs,
		tx:    tx,
	}, nil
}

func (bs *Store) Writer() (store.KVWriter, error) {
	bs.writer.Lock()
	tx, err := bs.db.Begin(true)
	if err != nil {
		bs.writer.Unlock()
		return nil, err
	}
	reader := &Reader{
		store: bs,
		tx:    tx,
	}
	return &Writer{
		store:  bs,
		tx:     tx,
		reader: reader,
	}, nil
}

func StoreConstructor(config map[string]interface{}) (store.KVStore, error) {
	path, ok := config["path"].(string)
	if !ok {
		return nil, fmt.Errorf("must specify path")
	}

	bucket, ok := config["bucket"].(string)
	if !ok {
		bucket = "bleve"
	}

	return New(path, bucket), nil
}

func init() {
	registry.RegisterKVStore(Name, StoreConstructor)
}
