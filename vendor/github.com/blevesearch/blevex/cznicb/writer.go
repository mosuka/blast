//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package cznicb

import (
	"fmt"

	"github.com/blevesearch/bleve/index/store"
)

type Writer struct {
	s *Store
}

func (w *Writer) NewBatch() store.KVBatch {
	return store.NewEmulatedBatch(w.s.mo)
}

func (w *Writer) NewBatchEx(options store.KVBatchOptions) ([]byte, store.KVBatch, error) {
	return make([]byte, options.TotalBytes), w.NewBatch(), nil
}

func (w *Writer) ExecuteBatch(batch store.KVBatch) error {

	emulatedBatch, ok := batch.(*store.EmulatedBatch)
	if !ok {
		return fmt.Errorf("wrong type of batch")
	}

	w.s.m.Lock()
	defer w.s.m.Unlock()

	t := w.s.t
	for key, mergeOps := range emulatedBatch.Merger.Merges {
		k := []byte(key)
		t.Put(k, func(oldV interface{}, exists bool) (newV interface{}, write bool) {
			ob := []byte(nil)
			if exists && oldV != nil {
				ob = oldV.([]byte)
			}
			mergedVal, fullMergeOk := w.s.mo.FullMerge(k, ob, mergeOps)
			if !fullMergeOk {
				return nil, false
			}
			return mergedVal, true
		})
	}

	for _, op := range emulatedBatch.Ops {
		if op.V != nil {
			t.Set(op.K, op.V)
		} else {
			t.Delete(op.K)
		}
	}

	return nil
}

func (w *Writer) Close() error {
	w.s.m.Lock()
	defer w.s.m.Unlock()
	w.s = nil
	return nil
}
