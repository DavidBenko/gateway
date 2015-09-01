//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

// +build icu full

package fa

import (
	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/registry"

	"github.com/blevesearch/bleve/analysis/char_filters/zero_width_non_joiner"
	"github.com/blevesearch/bleve/analysis/language/ar"
	"github.com/blevesearch/bleve/analysis/token_filters/lower_case_filter"
	"github.com/blevesearch/bleve/analysis/tokenizers/icu"
)

const AnalyzerName = "fa"

func AnalyzerConstructor(config map[string]interface{}, cache *registry.Cache) (*analysis.Analyzer, error) {
	zFilter, err := cache.CharFilterNamed(zero_width_non_joiner.Name)
	if err != nil {
		return nil, err
	}
	icuTokenizer, err := cache.TokenizerNamed(icu.Name)
	if err != nil {
		return nil, err
	}
	normArFilter, err := cache.TokenFilterNamed(ar.NormalizeName)
	if err != nil {
		return nil, err
	}
	normFaFilter, err := cache.TokenFilterNamed(NormalizeName)
	if err != nil {
		return nil, err
	}
	toLowerFilter, err := cache.TokenFilterNamed(lower_case_filter.Name)
	if err != nil {
		return nil, err
	}
	stopFaFilter, err := cache.TokenFilterNamed(StopName)
	if err != nil {
		return nil, err
	}
	rv := analysis.Analyzer{
		CharFilters: []analysis.CharFilter{
			zFilter,
		},
		Tokenizer: icuTokenizer,
		TokenFilters: []analysis.TokenFilter{
			toLowerFilter,
			normArFilter,
			normFaFilter,
			stopFaFilter,
		},
	}
	return &rv, nil
}

func init() {
	registry.RegisterAnalyzer(AnalyzerName, AnalyzerConstructor)
}
