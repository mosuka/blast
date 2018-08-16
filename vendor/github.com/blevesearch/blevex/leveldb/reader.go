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
	"github.com/blevesearch/bleve/index/store"
	"github.com/jmhodges/levigo"
)

type Reader struct {
	store    *Store
	snapshot *levigo.Snapshot
	options  *levigo.ReadOptions
}

func (r *Reader) Get(key []byte) ([]byte, error) {
	b, err := r.store.db.Get(r.options, key)
	return b, err
}

func (r *Reader) MultiGet(keys [][]byte) ([][]byte, error) {
	return store.MultiGet(r, keys)
}

func (r *Reader) PrefixIterator(prefix []byte) store.KVIterator {
	rv := Iterator{
		store:    r.store,
		iterator: r.store.db.NewIterator(r.options),
		prefix:   prefix,
	}
	rv.Seek(prefix)
	return &rv
}

func (r *Reader) RangeIterator(start, end []byte) store.KVIterator {
	rv := Iterator{
		store:    r.store,
		iterator: r.store.db.NewIterator(r.options),
		start:    start,
		end:      end,
	}
	rv.Seek(start)
	return &rv
}

func (r *Reader) Close() error {
	r.options.Close()
	r.store.db.ReleaseSnapshot(r.snapshot)
	return nil
}
