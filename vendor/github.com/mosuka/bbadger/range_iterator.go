package bbadger

import (
	"bytes"

	"github.com/dgraph-io/badger"
)

// RangeIterator implements blevesearch store iterator
type RangeIterator struct {
	txn      *badger.Txn
	iterator *badger.Iterator
	start    []byte
	stop     []byte
}

// Seek advance the iterator to the specified key
func (i *RangeIterator) Seek(key []byte) {
	if bytes.Compare(key, i.start) < 0 {
		i.iterator.Seek(i.start)
		return
	}
	i.iterator.Seek(key)
}

// Next advance the iterator to the next step
func (i *RangeIterator) Next() {
	i.iterator.Next()
}

// Current returns the key & value of the current step
func (i *RangeIterator) Current() ([]byte, []byte, bool) {
	if i.Valid() {
		return i.Key(), i.Value(), true
	}
	return nil, nil, false
}

// Key return the key of the current step
func (i *RangeIterator) Key() []byte {
	return i.iterator.Item().KeyCopy(nil)
}

// Value returns the value of the current step
func (i *RangeIterator) Value() []byte {
	v, _ := i.iterator.Item().ValueCopy(nil)

	return v
}

// Valid whether the current iterator step is valid or not
func (i *RangeIterator) Valid() bool {
	if !i.iterator.Valid() {
		return false
	}

	if i.stop == nil || len(i.stop) == 0 {
		return true
	}

	if bytes.Compare(i.stop, i.iterator.Item().Key()) <= 0 {
		return false
	}
	return true
	//return i.iterator.Valid()
}

// Close closes the current iterator and commit its transaction
func (i *RangeIterator) Close() error {
	i.iterator.Close()
	err := i.txn.Commit(nil)
	if err != nil {
		return err
	}
	return nil
}
