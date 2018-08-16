//  Copyright (c) 2015 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package cznicb

import "github.com/blevesearch/bleve/index/store"

type Reader struct {
	s *Store
}

func (r *Reader) Get(key []byte) ([]byte, error) {
	r.s.m.RLock()
	defer r.s.m.RUnlock()
	v, ok := r.s.t.Get(key)
	if !ok || v == nil {
		return nil, nil
	}
	rv := make([]byte, len(v.([]byte)))
	copy(rv, v.([]byte))
	return rv, nil
}

func (r *Reader) MultiGet(keys [][]byte) ([][]byte, error) {
	return store.MultiGet(r, keys)
}

func (r *Reader) PrefixIterator(prefix []byte) store.KVIterator {
	e, _ := r.s.t.SeekFirst()
	rv := Iterator{
		s:      r.s,
		e:      e,
		prefix: prefix,
	}
	rv.Seek(prefix)
	return &rv
}

func (r *Reader) RangeIterator(start, end []byte) store.KVIterator {
	e, _ := r.s.t.SeekFirst()
	rv := Iterator{
		s:     r.s,
		e:     e,
		start: start,
		end:   end,
	}
	rv.Seek(start)
	return &rv
}

func (r *Reader) Close() error {
	return nil
}
