package bbadger

import (
	"bytes"

	"github.com/dgraph-io/badger"
)

// PrefixIterator blevesearch prefix-iterator implementation
type PrefixIterator struct {
	txn      *badger.Txn
	iterator *badger.Iterator
	prefix   []byte
}

// Seek advance the iterator to the specified key
func (i *PrefixIterator) Seek(key []byte) {
	if bytes.Compare(key, i.prefix) < 0 {
		i.iterator.Seek(i.prefix)
		return
	}
	i.iterator.Seek(key)
}

// Next advance the iterator to the next step
func (i *PrefixIterator) Next() {
	i.iterator.Next()
}

// Current returns the key & value of the current step
func (i *PrefixIterator) Current() ([]byte, []byte, bool) {
	if i.Valid() {
		return i.Key(), i.Value(), true
	}
	return nil, nil, false
}

// Key return the key of the current step
func (i *PrefixIterator) Key() []byte {
	return i.iterator.Item().KeyCopy(nil)
}

// Value returns the value of the current step
func (i *PrefixIterator) Value() []byte {
	v, _ := i.iterator.Item().ValueCopy(nil)
	return v
}

// Valid whether the current iterator step is valid or not
func (i *PrefixIterator) Valid() bool {
	return i.iterator.ValidForPrefix(i.prefix)
}

// Close closes the current iterator and commit its transaction
func (i *PrefixIterator) Close() error {
	i.iterator.Close()
	err := i.txn.Commit(nil)
	if err != nil {
		return err
	}
	return nil
}
