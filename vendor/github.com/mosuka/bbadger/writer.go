package bbadger

import (
	"fmt"

	"github.com/blevesearch/bleve/index/store"
)

// Writer bleve.search/store/Writer implementation
// I (alash3al) adopted it from bleve/store/boltdb
type Writer struct {
	s *Store
}

// NewBatch implements NewBatch
func (w *Writer) NewBatch() store.KVBatch {
	return store.NewEmulatedBatch(w.s.mo)
}

// NewBatchEx implements bleve NewBatchEx
func (w *Writer) NewBatchEx(options store.KVBatchOptions) ([]byte, store.KVBatch, error) {
	return make([]byte, options.TotalBytes), w.NewBatch(), nil
}

// ExecuteBatch implements bleve ExecuteBatch
func (w *Writer) ExecuteBatch(batch store.KVBatch) (err error) {
	emulatedBatch, ok := batch.(*store.EmulatedBatch)
	if !ok {
		return fmt.Errorf("wrong type of batch")
	}

	txn := w.s.db.NewTransaction(true)

	defer (func() {
		txn.Commit(nil)
	})()

	for k, mergeOps := range emulatedBatch.Merger.Merges {
		kb := []byte(k)
		item, err := txn.Get(kb)
		existingVal := []byte{}
		if err == nil {
			existingVal, _ = item.ValueCopy(nil)
		}
		mergedVal, fullMergeOk := w.s.mo.FullMerge(kb, existingVal, mergeOps)
		if !fullMergeOk {
			return fmt.Errorf("merge operator returned failure")
		}
		err = txn.Set(kb, mergedVal)
		if err != nil {
			return err
		}
	}

	for _, op := range emulatedBatch.Ops {
		if op.V != nil {
			err = txn.Set(op.K, op.V)
			if err != nil {
				return err
			}
		} else {
			err = txn.Delete(op.K)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Close closes the current writer
func (w *Writer) Close() error {
	return nil
}
