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

package index

import (
	"time"
)

type Indexer struct {
	index *Index
}

func NewIndexer(index *Index) (*Indexer, error) {
	return &Indexer{
		index: index,
	}, nil
}

func (i *Indexer) Index(id string, fields map[string]interface{}) error {
	start := time.Now()
	defer Metrics(start, "Indexer", "Index")

	var err error

	if err = i.index.index.Index(id, fields); err != nil {
		i.index.logger.Printf("[ERR] bleve: Failed to index document: %s: %v: %v", id, fields, err)
		return err
	}

	i.index.logger.Printf("[DEBUG] bleve: Document has been indexed: %s: %v", id, fields)
	return nil
}

func (i *Indexer) Delete(id string) error {
	start := time.Now()
	defer Metrics(start, "Indexer", "Delete")

	var err error

	if err = i.index.index.Delete(id); err != nil {
		i.index.logger.Printf("[ERR] bleve: Failed to delete document: %s: %v", id, err)
		return err
	}

	i.index.logger.Printf("[DEBUG] bleve: Document has been deleted: %s", id)
	return nil
}

func (i *Indexer) Close() error {
	return nil
}
