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
	"time"

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/index/metrics"
)

type Bulker struct {
	index     *Index
	batch     *bleve.Batch
	batchSize int
	count     int
}

func NewBulker(index *Index, batchSize int) (*Bulker, error) {

	batch := index.index.NewBatch()

	return &Bulker{
		index:     index,
		batch:     batch,
		batchSize: batchSize,
		count:     0,
	}, nil
}

func (b *Bulker) Index(id string, fields map[string]interface{}) error {
	start := time.Now()
	defer metrics.Metrics(start, "Bulker", "Index")

	var err error

	if err = b.batch.Index(id, fields); err != nil {
		b.index.logger.Printf("[ERR] bleve: Failed to index document in bulk: %s: %v: %v", id, fields, err)
		return err
	}

	b.count++

	if b.count >= b.batchSize {
		if err = b.index.index.Batch(b.batch); err != nil {
			b.index.logger.Printf("[ERR] bleve: Failed to execute batch: %v", err)
			return err
		}
		b.batch = b.index.index.NewBatch()
		b.count = 0
	}

	return nil
}

func (b *Bulker) Delete(id string) error {
	start := time.Now()
	defer metrics.Metrics(start, "Bulker", "Delete")

	var err error

	b.batch.Delete(id)

	b.count++

	if b.count >= b.batchSize {
		if err = b.index.index.Batch(b.batch); err != nil {
			b.index.logger.Printf("[ERR] bleve: Failed to execute batch: %v", err)
			return err
		}
		b.batch = b.index.index.NewBatch()
		b.count = 0
	}

	return nil
}

func (b *Bulker) Close() error {
	var err error

	if b.count > 0 {
		if err = b.index.index.Batch(b.batch); err != nil {
			b.index.logger.Printf("[ERR] bleve: Failed to execute batch: %v", err)
			return err
		}
	}

	return nil
}
