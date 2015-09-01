//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package upside_down

import (
	"os"
	"testing"

	"github.com/blevesearch/bleve/index/store"
	"github.com/blevesearch/bleve/index/store/goleveldb"
)

var goLevelDBTestOptions = map[string]interface{}{
	"create_if_missing": true,
}

func CreateGoLevelDB() (store.KVStore, error) {
	return goleveldb.New("test", goLevelDBTestOptions)
}

func DestroyGoLevelDB() error {
	return os.RemoveAll("test")
}

func BenchmarkGoLevelDBIndexing1Workers(b *testing.B) {
	CommonBenchmarkIndex(b, CreateGoLevelDB, DestroyGoLevelDB, 1)
}

func BenchmarkGoLevelDBIndexing2Workers(b *testing.B) {
	CommonBenchmarkIndex(b, CreateGoLevelDB, DestroyGoLevelDB, 2)
}

func BenchmarkGoLevelDBIndexing4Workers(b *testing.B) {
	CommonBenchmarkIndex(b, CreateGoLevelDB, DestroyGoLevelDB, 4)
}

// batches

func BenchmarkGoLevelDBIndexing1Workers10Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 1, 10)
}

func BenchmarkGoLevelDBIndexing2Workers10Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 2, 10)
}

func BenchmarkGoLevelDBIndexing4Workers10Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 4, 10)
}

func BenchmarkGoLevelDBIndexing1Workers100Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 1, 100)
}

func BenchmarkGoLevelDBIndexing2Workers100Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 2, 100)
}

func BenchmarkGoLevelDBIndexing4Workers100Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 4, 100)
}

func BenchmarkGoLevelDBIndexing1Workers1000Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 1, 1000)
}

func BenchmarkGoLevelDBIndexing2Workers1000Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 2, 1000)
}

func BenchmarkGoLevelDBIndexing4Workers1000Batch(b *testing.B) {
	CommonBenchmarkIndexBatch(b, CreateGoLevelDB, DestroyGoLevelDB, 4, 1000)
}
