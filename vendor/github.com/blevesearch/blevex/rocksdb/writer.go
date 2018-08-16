//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package rocksdb

import (
	"fmt"

	"github.com/blevesearch/bleve/index/store"
	"github.com/tecbot/gorocksdb"
)

type Writer struct {
	store   *Store
	options *gorocksdb.WriteOptions
}

func (w *Writer) NewBatch() store.KVBatch {
	rv := Batch{
		batch: gorocksdb.NewWriteBatch(),
	}
	return &rv
}

func (w *Writer) NewBatchEx(options store.KVBatchOptions) ([]byte, store.KVBatch, error) {
	// Disabled due to https://github.com/blevesearch/blevex/issues/22
	// rv := newBatchEx(options)
	// return rv.buf, rv, nil

	return make([]byte, options.TotalBytes), w.NewBatch(), nil
}

func (w *Writer) ExecuteBatch(b store.KVBatch) error {
	batchex, ok := b.(*BatchEx)
	if ok {
		return batchex.execute(w)
	}
	batch, ok := b.(*Batch)
	if ok {
		return w.store.db.Write(w.options, batch.batch)
	}
	return fmt.Errorf("wrong type of batch")
}

func (w *Writer) Close() error {
	w.options.Destroy()
	return nil
}
