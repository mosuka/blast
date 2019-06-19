// Copyright (c) 2019 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sortutils

import (
	"github.com/blevesearch/bleve/search"
)

type MultiSearchHitSorter struct {
	hits          search.DocumentMatchCollection
	sort          search.SortOrder
	cachedScoring []bool
	cachedDesc    []bool
}

func NewMultiSearchHitSorter(sort search.SortOrder, hits search.DocumentMatchCollection) *MultiSearchHitSorter {
	return &MultiSearchHitSorter{
		sort:          sort,
		hits:          hits,
		cachedScoring: sort.CacheIsScore(),
		cachedDesc:    sort.CacheDescending(),
	}
}

func (m *MultiSearchHitSorter) Len() int      { return len(m.hits) }
func (m *MultiSearchHitSorter) Swap(i, j int) { m.hits[i], m.hits[j] = m.hits[j], m.hits[i] }
func (m *MultiSearchHitSorter) Less(i, j int) bool {
	c := m.sort.Compare(m.cachedScoring, m.cachedDesc, m.hits[i], m.hits[j])
	return c < 0
}
