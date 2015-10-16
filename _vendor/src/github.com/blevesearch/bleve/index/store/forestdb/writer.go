//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

// +build forestdb

package forestdb

import (
	"github.com/blevesearch/bleve/index/store"
)

type Writer struct {
	store *Store
}

func (w *Writer) BytesSafeAfterClose() bool {
	return true
}

func newWriter(store *Store) (*Writer, error) {
	store.writer.Lock()
	return &Writer{
		store: store,
	}, nil
}

func (w *Writer) Set(key, val []byte) error {
	err := w.store.setlocked(key, val)
	if err != nil {
		return err
	}
	return w.store.commit()
}

func (w *Writer) Delete(key []byte) error {
	err := w.store.deletelocked(key)
	if err != nil {
		return err
	}
	return w.store.commit()
}

func (w *Writer) NewBatch() store.KVBatch {
	return &Batch{
		s:      w.store,
		ops:    make([]op, 0, 1000),
		merges: make(map[string][][]byte),
	}
}

func (w *Writer) Close() error {
	w.store.writer.Unlock()
	return nil
}

// these two methods can safely read using the regular
// methods without a read transaction, because we know
// that no one else is writing but us
func (w *Writer) Get(key []byte) ([]byte, error) {
	return w.store.get(key)
}

func (w *Writer) Iterator(key []byte) store.KVIterator {
	return w.store.iterator(key)
}
