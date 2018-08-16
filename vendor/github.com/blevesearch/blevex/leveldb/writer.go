//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package leveldb

import (
	"fmt"

	"github.com/blevesearch/bleve/index/store"
	"github.com/jmhodges/levigo"
)

type Writer struct {
	store   *Store
	options *levigo.WriteOptions
}

func (w *Writer) NewBatch() store.KVBatch {
	rv := Batch{
		w:     w,
		merge: store.NewEmulatedMerge(w.store.mo),
		batch: levigo.NewWriteBatch(),
	}
	return &rv
}

func (w *Writer) NewBatchEx(options store.KVBatchOptions) ([]byte, store.KVBatch, error) {
	return make([]byte, options.TotalBytes), w.NewBatch(), nil
}

func (w *Writer) ExecuteBatch(b store.KVBatch) error {

	batch, ok := b.(*Batch)
	if !ok {
		return fmt.Errorf("wrong type of batch")
	}

	// get a lock, because we can't allow
	// concurrent writes during the merge
	w.store.mergeMutex.Lock()
	defer w.store.mergeMutex.Unlock()

	// get a snapshot
	snapshot := w.store.db.NewSnapshot()
	ro := defaultReadOptions()
	ro.SetSnapshot(snapshot)
	defer w.store.db.ReleaseSnapshot(snapshot)
	defer ro.Close()

	for key, mergeOps := range batch.merge.Merges {
		k := []byte(key)
		orig, err := w.store.db.Get(ro, k)
		if err != nil {
			return err
		}
		mergedVal, fullMergeOk := w.store.mo.FullMerge(k, orig, mergeOps)
		if !fullMergeOk {
			return fmt.Errorf("unable to merge")
		}
		batch.Set(k, mergedVal)
	}

	err := w.store.db.Write(w.options, batch.batch)
	return err

}

func (w *Writer) Close() error {
	w.options.Close()
	return nil
}
