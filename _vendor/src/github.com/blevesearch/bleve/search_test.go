//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package bleve

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/blevesearch/bleve/search"
)

func TestSearchResultString(t *testing.T) {

	tests := []struct {
		result *SearchResult
		str    string
	}{
		{
			result: &SearchResult{
				Request: &SearchRequest{
					Size: 10,
				},
				Total: 5,
				Took:  1 * time.Second,
				Hits: search.DocumentMatchCollection{
					&search.DocumentMatch{},
					&search.DocumentMatch{},
					&search.DocumentMatch{},
					&search.DocumentMatch{},
					&search.DocumentMatch{},
				},
			},
			str: "5 matches, showing 1 through 5, took 1s",
		},
		{
			result: &SearchResult{
				Request: &SearchRequest{
					Size: 0,
				},
				Total: 5,
				Hits:  search.DocumentMatchCollection{},
			},
			str: "5 matches",
		},
		{
			result: &SearchResult{
				Request: &SearchRequest{
					Size: 10,
				},
				Total: 0,
				Hits:  search.DocumentMatchCollection{},
			},
			str: "No matches",
		},
	}

	for _, test := range tests {
		srstring := test.result.String()
		if !strings.HasPrefix(srstring, test.str) {
			t.Errorf("expected to start %s, got %s", test.str, srstring)
		}
	}
}

func TestSearchResultMerge(t *testing.T) {
	l := &SearchResult{
		Total:    1,
		MaxScore: 1,
		Hits: search.DocumentMatchCollection{
			&search.DocumentMatch{
				ID:    "a",
				Score: 1,
			},
		},
	}

	r := &SearchResult{
		Total:    1,
		MaxScore: 2,
		Hits: search.DocumentMatchCollection{
			&search.DocumentMatch{
				ID:    "b",
				Score: 2,
			},
		},
	}

	expected := &SearchResult{
		Total:    2,
		MaxScore: 2,
		Hits: search.DocumentMatchCollection{
			&search.DocumentMatch{
				ID:    "a",
				Score: 1,
			},
			&search.DocumentMatch{
				ID:    "b",
				Score: 2,
			},
		},
	}

	l.Merge(r)

	if !reflect.DeepEqual(l, expected) {
		t.Errorf("expected %#v, got %#v", expected, l)
	}
}
