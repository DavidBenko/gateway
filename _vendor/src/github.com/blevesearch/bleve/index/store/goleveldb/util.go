//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package goleveldb

import (
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func defaultWriteOptions() *opt.WriteOptions {
	wo := &opt.WriteOptions{}
	// request fsync on write for safety
	wo.Sync = true
	return wo
}

func defaultReadOptions() *opt.ReadOptions {
	ro := &opt.ReadOptions{}
	return ro
}
