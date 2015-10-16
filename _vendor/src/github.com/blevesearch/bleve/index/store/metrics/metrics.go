//  Copyright (c) 2015 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

// Package metrics provides a bleve.store.KVStore implementation that
// wraps another, real KVStore implementation, and uses go-metrics to
// track runtime performance metrics.
package metrics

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/blevesearch/bleve/index/store"
	"github.com/blevesearch/bleve/registry"

	"github.com/rcrowley/go-metrics"
)

const Name = "metrics"
const MaxErrors = 100

func init() {
	registry.RegisterKVStore(Name, StoreConstructor)
}

func StoreConstructor(config map[string]interface{}) (store.KVStore, error) {
	name, ok := config["kvStoreName_actual"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("metrics: missing kvStoreName_actual,"+
			" config: %#v", config)
	}

	if name == Name {
		return nil, fmt.Errorf("metrics: circular kvStoreName_actual")
	}

	ctr := registry.KVStoreConstructorByName(name)
	if ctr == nil {
		return nil, fmt.Errorf("metrics: no kv store constructor,"+
			" kvStoreName_actual: %s", name)
	}

	kvs, err := ctr(config)
	if err != nil {
		return nil, err
	}

	return NewBleveMetricsStore(kvs), nil
}

func NewBleveMetricsStore(o store.KVStore) *Store {
	return &Store{
		o: o,

		TimerReaderGet:         metrics.NewTimer(),
		TimerReaderIterator:    metrics.NewTimer(),
		TimerWriterGet:         metrics.NewTimer(),
		TimerWriterIterator:    metrics.NewTimer(),
		TimerWriterSet:         metrics.NewTimer(),
		TimerWriterDelete:      metrics.NewTimer(),
		TimerIteratorSeekFirst: metrics.NewTimer(),
		TimerIteratorSeek:      metrics.NewTimer(),
		TimerIteratorNext:      metrics.NewTimer(),
		TimerBatchMerge:        metrics.NewTimer(),
		TimerBatchExecute:      metrics.NewTimer(),

		errors: list.New(),
	}
}

// The following structs are wrappers around "real" bleve kvstore
// implementations.

type Store struct {
	o store.KVStore

	TimerReaderGet         metrics.Timer
	TimerReaderIterator    metrics.Timer
	TimerWriterGet         metrics.Timer
	TimerWriterIterator    metrics.Timer
	TimerWriterSet         metrics.Timer
	TimerWriterDelete      metrics.Timer
	TimerIteratorSeekFirst metrics.Timer
	TimerIteratorSeek      metrics.Timer
	TimerIteratorNext      metrics.Timer
	TimerBatchMerge        metrics.Timer
	TimerBatchExecute      metrics.Timer

	m      sync.Mutex // Protects the fields that follow.
	errors *list.List // Capped list of StoreError's.
}

type StoreError struct {
	Time string
	Op   string
	Err  string
	Key  string
}

type Reader struct {
	s *Store
	o store.KVReader
}

type Writer struct {
	s *Store
	o store.KVWriter
}

type Iterator struct {
	s *Store
	o store.KVIterator
}

type Batch struct {
	s *Store
	o store.KVBatch
}

func (s *Store) Open() error {
	return s.o.Open()
}

func (s *Store) Close() error {
	return s.o.Close()
}

func (s *Store) SetMergeOperator(mo store.MergeOperator) {
	s.o.SetMergeOperator(mo)
}

func (s *Store) Reader() (store.KVReader, error) {
	o, err := s.o.Reader()
	if err != nil {
		s.AddError("Reader", err, nil)
		return nil, err
	}
	return &Reader{s: s, o: o}, nil
}

func (s *Store) Writer() (store.KVWriter, error) {
	o, err := s.o.Writer()
	if err != nil {
		s.AddError("Writer", err, nil)
		return nil, err
	}
	return &Writer{s: s, o: o}, nil
}

func (s *Store) Actual() store.KVStore {
	return s.o
}

func (w *Reader) BytesSafeAfterClose() bool {
	return w.o.BytesSafeAfterClose()
}

func (w *Reader) Get(key []byte) (v []byte, err error) {
	w.s.TimerReaderGet.Time(func() {
		v, err = w.o.Get(key)
		if err != nil {
			w.s.AddError("Reader.Get", err, key)
		}
	})
	return
}

func (w *Reader) Iterator(key []byte) (i store.KVIterator) {
	w.s.TimerReaderIterator.Time(func() {
		i = &Iterator{s: w.s, o: w.o.Iterator(key)}
	})
	return
}

func (w *Reader) Close() error {
	err := w.o.Close()
	if err != nil {
		w.s.AddError("Reader.Close", err, nil)
	}
	return err
}

func (w *Writer) BytesSafeAfterClose() bool {
	return w.o.BytesSafeAfterClose()
}

func (w *Writer) Get(key []byte) (v []byte, err error) {
	w.s.TimerWriterGet.Time(func() {
		v, err = w.o.Get(key)
		if err != nil {
			w.s.AddError("Writer.Get", err, key)
		}
	})
	return
}

func (w *Writer) Iterator(key []byte) (i store.KVIterator) {
	w.s.TimerWriterIterator.Time(func() {
		i = &Iterator{s: w.s, o: w.o.Iterator(key)}
	})
	return
}

func (w *Writer) Close() error {
	err := w.o.Close()
	if err != nil {
		w.s.AddError("Writer.Close", err, nil)
	}
	return err
}

func (w *Writer) Set(key, val []byte) (err error) {
	w.s.TimerWriterSet.Time(func() {
		err = w.o.Set(key, val)
		if err != nil {
			w.s.AddError("Writer.Set", err, key)
		}
	})
	return
}

func (w *Writer) Delete(key []byte) (err error) {
	w.s.TimerWriterDelete.Time(func() {
		err = w.o.Delete(key)
		if err != nil {
			w.s.AddError("Writer.Delete", err, key)
		}
	})
	return
}

func (w *Writer) NewBatch() store.KVBatch {
	return &Batch{s: w.s, o: w.o.NewBatch()}
}

func (w *Iterator) SeekFirst() {
	w.s.TimerIteratorSeekFirst.Time(func() {
		w.o.SeekFirst()
	})
}

func (w *Iterator) Seek(x []byte) {
	w.s.TimerIteratorSeek.Time(func() {
		w.o.Seek(x)
	})
}

func (w *Iterator) Next() {
	w.s.TimerIteratorNext.Time(func() {
		w.o.Next()
	})
}

func (w *Iterator) Current() ([]byte, []byte, bool) {
	return w.o.Current()
}

func (w *Iterator) Key() []byte {
	return w.o.Key()
}

func (w *Iterator) Value() []byte {
	return w.o.Value()
}

func (w *Iterator) Valid() bool {
	return w.o.Valid()
}

func (w *Iterator) Close() error {
	err := w.o.Close()
	if err != nil {
		w.s.AddError("Iterator.Close", err, nil)
	}
	return err
}

func (w *Batch) Set(key, val []byte) {
	w.o.Set(key, val)
}

func (w *Batch) Delete(key []byte) {
	w.o.Delete(key)
}

func (w *Batch) Merge(key, val []byte) {
	w.s.TimerBatchMerge.Time(func() {
		w.o.Merge(key, val)
	})
}

func (w *Batch) Execute() (err error) {
	w.s.TimerBatchExecute.Time(func() {
		err = w.o.Execute()
		if err != nil {
			w.s.AddError("Batch.Execute", err, nil)
		}
	})
	return
}

func (w *Batch) Close() error {
	err := w.o.Close()
	if err != nil {
		w.s.AddError("Batch.Close", err, nil)
	}
	return err
}

// --------------------------------------------------------

func (s *Store) AddError(op string, err error, key []byte) {
	e := &StoreError{
		Time: time.Now().Format(time.RFC3339Nano),
		Op:   op,
		Err:  fmt.Sprintf("%v", err),
		Key:  string(key),
	}

	s.m.Lock()
	for s.errors.Len() >= MaxErrors {
		s.errors.Remove(s.errors.Front())
	}
	s.errors.PushBack(e)
	s.m.Unlock()
}

// --------------------------------------------------------

func (s *Store) WriteJSON(w io.Writer) {
	w.Write([]byte(`{"TimerReaderGet":`))
	WriteTimerJSON(w, s.TimerReaderGet)
	w.Write([]byte(`,"TimerReaderIterator":`))
	WriteTimerJSON(w, s.TimerReaderIterator)
	w.Write([]byte(`,"TimerWriterGet":`))
	WriteTimerJSON(w, s.TimerWriterGet)
	w.Write([]byte(`,"TimerWriterIterator":`))
	WriteTimerJSON(w, s.TimerWriterIterator)
	w.Write([]byte(`,"TimerWriterSet":`))
	WriteTimerJSON(w, s.TimerWriterSet)
	w.Write([]byte(`,"TimerWriterDelete":`))
	WriteTimerJSON(w, s.TimerWriterDelete)
	w.Write([]byte(`,"TimerIteratorSeekFirst":`))
	WriteTimerJSON(w, s.TimerIteratorSeekFirst)
	w.Write([]byte(`,"TimerIteratorSeek":`))
	WriteTimerJSON(w, s.TimerIteratorSeek)
	w.Write([]byte(`,"TimerIteratorNext":`))
	WriteTimerJSON(w, s.TimerIteratorNext)
	w.Write([]byte(`,"TimerBatchMerge":`))
	WriteTimerJSON(w, s.TimerBatchMerge)
	w.Write([]byte(`,"TimerBatchExecute":`))
	WriteTimerJSON(w, s.TimerBatchExecute)

	w.Write([]byte(`,"Errors":[`))
	s.m.Lock()
	e := s.errors.Front()
	i := 0
	for e != nil {
		se, ok := e.Value.(*StoreError)
		if ok && se != nil {
			if i > 0 {
				w.Write([]byte(","))
			}
			buf, err := json.Marshal(se)
			if err == nil {
				w.Write(buf)
			}
		}
		e = e.Next()
		i = i + 1
	}
	s.m.Unlock()
	w.Write([]byte(`]`))

	w.Write([]byte(`}`))
}

func (s *Store) WriteCSVHeader(w io.Writer) {
	WriteTimerCSVHeader(w, "TimerReaderGet")
	WriteTimerCSVHeader(w, "TimerReaderIterator")
	WriteTimerCSVHeader(w, "TimerWriterGet")
	WriteTimerCSVHeader(w, "TimerWriterIterator")
	WriteTimerCSVHeader(w, "TimerWriterSet")
	WriteTimerCSVHeader(w, "TimerWriterDelete")
	WriteTimerCSVHeader(w, "TimerIteratorSeekFirst")
	WriteTimerCSVHeader(w, "TimerIteratorSeek")
	WriteTimerCSVHeader(w, "TimerIteratorNext")
	WriteTimerCSVHeader(w, "TimerBatchMerge")
	WriteTimerCSVHeader(w, "TimerBatchExecute")
}

func (s *Store) WriteCSV(w io.Writer) {
	WriteTimerCSV(w, s.TimerReaderGet)
	WriteTimerCSV(w, s.TimerReaderIterator)
	WriteTimerCSV(w, s.TimerWriterGet)
	WriteTimerCSV(w, s.TimerWriterIterator)
	WriteTimerCSV(w, s.TimerWriterSet)
	WriteTimerCSV(w, s.TimerWriterDelete)
	WriteTimerCSV(w, s.TimerIteratorSeekFirst)
	WriteTimerCSV(w, s.TimerIteratorSeek)
	WriteTimerCSV(w, s.TimerIteratorNext)
	WriteTimerCSV(w, s.TimerBatchMerge)
	WriteTimerCSV(w, s.TimerBatchExecute)
}

// --------------------------------------------------------

// NOTE: This is copy & pasted from cbft as otherwise there
// would be an import cycle.

var timerPercentiles = []float64{0.5, 0.75, 0.95, 0.99, 0.999}

func WriteTimerJSON(w io.Writer, timer metrics.Timer) {
	t := timer.Snapshot()
	p := t.Percentiles(timerPercentiles)

	fmt.Fprintf(w, `{"count":%9d,`, t.Count())
	fmt.Fprintf(w, `"min":%9d,`, t.Min())
	fmt.Fprintf(w, `"max":%9d,`, t.Max())
	fmt.Fprintf(w, `"mean":%12.2f,`, t.Mean())
	fmt.Fprintf(w, `"stddev":%12.2f,`, t.StdDev())
	fmt.Fprintf(w, `"percentiles":{`)
	fmt.Fprintf(w, `"median":%12.2f,`, p[0])
	fmt.Fprintf(w, `"75%%":%12.2f,`, p[1])
	fmt.Fprintf(w, `"95%%":%12.2f,`, p[2])
	fmt.Fprintf(w, `"99%%":%12.2f,`, p[3])
	fmt.Fprintf(w, `"99.9%%":%12.2f},`, p[4])
	fmt.Fprintf(w, `"rates":{`)
	fmt.Fprintf(w, `"1-min":%12.2f,`, t.Rate1())
	fmt.Fprintf(w, `"5-min":%12.2f,`, t.Rate5())
	fmt.Fprintf(w, `"15-min":%12.2f,`, t.Rate15())
	fmt.Fprintf(w, `"mean":%12.2f}}`, t.RateMean())
}

func WriteTimerCSVHeader(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s-count,", prefix)
	fmt.Fprintf(w, "%s-min,", prefix)
	fmt.Fprintf(w, "%s-max,", prefix)
	fmt.Fprintf(w, "%s-mean,", prefix)
	fmt.Fprintf(w, "%s-stddev,", prefix)
	fmt.Fprintf(w, "%s-percentile-50%%,", prefix)
	fmt.Fprintf(w, "%s-percentile-75%%,", prefix)
	fmt.Fprintf(w, "%s-percentile-95%%,", prefix)
	fmt.Fprintf(w, "%s-percentile-99%%,", prefix)
	fmt.Fprintf(w, "%s-percentile-99.9%%,", prefix)
	fmt.Fprintf(w, "%s-rate-1-min,", prefix)
	fmt.Fprintf(w, "%s-rate-5-min,", prefix)
	fmt.Fprintf(w, "%s-rate-15-min,", prefix)
	fmt.Fprintf(w, "%s-rate-mean", prefix)
}

func WriteTimerCSV(w io.Writer, timer metrics.Timer) {
	t := timer.Snapshot()
	p := t.Percentiles(timerPercentiles)

	fmt.Fprintf(w, `%d,`, t.Count())
	fmt.Fprintf(w, `%d,`, t.Min())
	fmt.Fprintf(w, `%d,`, t.Max())
	fmt.Fprintf(w, `%f,`, t.Mean())
	fmt.Fprintf(w, `%f,`, t.StdDev())
	fmt.Fprintf(w, `%f,`, p[0])
	fmt.Fprintf(w, `%f,`, p[1])
	fmt.Fprintf(w, `%f,`, p[2])
	fmt.Fprintf(w, `%f,`, p[3])
	fmt.Fprintf(w, `%f,`, p[4])
	fmt.Fprintf(w, `%f,`, t.Rate1())
	fmt.Fprintf(w, `%f,`, t.Rate5())
	fmt.Fprintf(w, `%f,`, t.Rate15())
	fmt.Fprintf(w, `%f`, t.RateMean())
}
