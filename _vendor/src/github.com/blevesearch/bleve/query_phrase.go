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
	"encoding/json"
	"fmt"

	"github.com/blevesearch/bleve/index"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/searchers"
)

type phraseQuery struct {
	TermQueries []Query `json:"terms"`
	BoostVal    float64 `json:"boost,omitempty"`
	terms       []string
}

// NewPhraseQuery creates a new Query for finding
// exact term phrases in the index.
// The provided terms must exist in the correct
// order, at the correct index offsets, in the
// specified field.
func NewPhraseQuery(terms []string, field string) *phraseQuery {
	termQueries := make([]Query, 0)
	for _, term := range terms {
		if term != "" {
			termQueries = append(termQueries, NewTermQuery(term).SetField(field))
		}
	}
	return &phraseQuery{
		TermQueries: termQueries,
		BoostVal:    1.0,
		terms:       terms,
	}
}

func (q *phraseQuery) Boost() float64 {
	return q.BoostVal
}

func (q *phraseQuery) SetBoost(b float64) Query {
	q.BoostVal = b
	return q
}

func (q *phraseQuery) Searcher(i index.IndexReader, m *IndexMapping, explain bool) (search.Searcher, error) {

	conjunctionQuery := NewConjunctionQuery(q.TermQueries)
	conjunctionSearcher, err := conjunctionQuery.Searcher(i, m, explain)
	if err != nil {
		return nil, err
	}
	return searchers.NewPhraseSearcher(i, conjunctionSearcher.(*searchers.ConjunctionSearcher), q.terms)
}

func (q *phraseQuery) Validate() error {
	if len(q.TermQueries) < 1 {
		return ErrorPhraseQueryNoTerms
	}
	return nil
}

func (q *phraseQuery) UnmarshalJSON(data []byte) error {
	tmp := struct {
		Terms    []json.RawMessage `json:"terms"`
		BoostVal float64           `json:"boost,omitempty"`
	}{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	q.TermQueries = make([]Query, len(tmp.Terms))
	q.terms = make([]string, 0)
	for i, term := range tmp.Terms {
		query, err := ParseQuery(term)
		if err != nil {
			return err
		}
		q.TermQueries[i] = query
		tq, isTermQuery := query.(*termQuery)
		if !isTermQuery {
			return fmt.Errorf("phrase query can only contain term queries")
		}
		q.terms = append(q.terms, tq.Term)
	}
	q.BoostVal = tmp.BoostVal
	if q.BoostVal == 0 {
		q.BoostVal = 1
	}
	return nil
}

func (q *phraseQuery) Field() string {
	return ""
}

func (q *phraseQuery) SetField(f string) Query {
	return q
}
