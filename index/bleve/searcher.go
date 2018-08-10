//  Copyright (c) 2018 Minoru Osuka
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

package bleve

import (
	"encoding/json"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/index/metrics"
)

type Searcher struct {
	index *Index
}

func NewSearcher(index *Index) (*Searcher, error) {
	return &Searcher{
		index: index,
	}, nil
}

func (s *Searcher) Search(request []byte) ([]byte, error) {
	var err error

	start := time.Now()
	defer metrics.Metrics(start, "Searcher", "Search")

	var searchRequest *bleve.SearchRequest
	if err = json.Unmarshal(request, &searchRequest); err != nil {
		s.index.logger.Printf("[ERR] bleve: Failed to unmarshaling request: %v: %v", request, err)
		return nil, err
	}

	var searchResult *bleve.SearchResult
	if searchResult, err = s.index.index.Search(searchRequest); err != nil {
		s.index.logger.Printf("[ERR] bleve: Failed to search index: %v: %v", searchRequest, err)
		return nil, err
	}

	var result []byte
	if result, err = json.Marshal(searchResult); err != nil {
		s.index.logger.Printf("[ERR] bleve: Failed to marshaling search result: %v: %v", searchResult, err)
		return nil, err
	}

	s.index.logger.Printf("[DEBUG] bleve: Documents has been searched: %v: %v", searchResult, searchResult)
	return result, nil
}

func (s *Searcher) Close() error {
	return nil
}
