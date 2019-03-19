package bbadger

import (
	"github.com/blevesearch/bleve/index/store"
	"github.com/dgraph-io/badger"
)

// Reader implements bleve/Store/Reader interface
type Reader struct {
	s   *Store
	txn *badger.Txn
}

// Get fetch the value of the specified key from the store
func (r *Reader) Get(k []byte) ([]byte, error) {
	item, err := r.txn.Get(k)
	if err != nil {
		return nil, nil
	}
	return item.ValueCopy(nil)
}

// MultiGet returns multiple values for the specified keys
func (r *Reader) MultiGet(keys [][]byte) ([][]byte, error) {
	return store.MultiGet(r, keys)
}

// PrefixIterator initialize a new prefix iterator
func (r *Reader) PrefixIterator(k []byte) store.KVIterator {
	txn := r.s.db.NewTransaction(false)
	rv := &PrefixIterator{
		txn:      txn,
		iterator: txn.NewIterator(badger.DefaultIteratorOptions),
		prefix:   k,
	}
	rv.iterator.Seek(k)
	return rv
}

// RangeIterator initialize a new range iterator
func (r *Reader) RangeIterator(start, end []byte) store.KVIterator {
	txn := r.s.db.NewTransaction(false)
	rv := &RangeIterator{
		txn:      txn,
		iterator: txn.NewIterator(badger.DefaultIteratorOptions),
		start:    start,
		stop:     end,
	}
	rv.iterator.Seek(start)
	return rv
}

// Close closes the current reader and do some cleanup
func (r *Reader) Close() error {
	return r.txn.Commit(nil)
}
