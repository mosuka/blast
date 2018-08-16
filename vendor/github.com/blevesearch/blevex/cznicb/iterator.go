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

import (
	"bytes"
	"errors"

	"github.com/cznic/b"
)

var iteratorDoneErr = errors.New("iteratorDoneErr") // A sentinel value.

type Iterator struct {
	s *Store
	e *b.Enumerator

	currK   interface{}
	currV   interface{}
	currErr error

	prefix []byte
	start  []byte
	end    []byte
}

func (i *Iterator) Seek(k []byte) {
	if i.start != nil && bytes.Compare(k, i.start) < 0 {
		k = i.start
	}
	if i.prefix != nil && !bytes.HasPrefix(k, i.prefix) {
		if bytes.Compare(k, i.prefix) < 0 {
			k = i.prefix
		} else {
			var end []byte
			for x := len(i.prefix) - 1; x >= 0; x-- {
				c := i.prefix[x]
				if c < 0xff {
					end = make([]byte, x+1)
					copy(end, i.prefix)
					end[x] = c + 1
					break
				}
			}
			k = end
		}
	}

	i.currK = nil
	i.currV = nil
	i.currErr = nil

	i.s.m.RLock()
	i.e, _ = i.s.t.Seek(k)
	i.s.m.RUnlock() // cannot defer, must unlock before Next

	i.Next()
}

func (i *Iterator) Next() {
	if i.currErr != nil {
		i.currK = nil
		i.currV = nil
		i.currErr = iteratorDoneErr
		return
	}

	i.s.m.RLock()
	defer i.s.m.RUnlock()
	i.currK, i.currV, i.currErr = i.e.Next()
}

func (i *Iterator) Current() ([]byte, []byte, bool) {
	if i.currErr == iteratorDoneErr ||
		i.currK == nil ||
		i.currV == nil {
		return nil, nil, false
	}
	if i.prefix != nil && !bytes.HasPrefix(i.currK.([]byte), i.prefix) {
		return nil, nil, false
	} else if i.end != nil && bytes.Compare(i.currK.([]byte), i.end) >= 0 {
		return nil, nil, false
	}

	return i.currK.([]byte), i.currV.([]byte), true
}

func (i *Iterator) Key() []byte {
	k, _, ok := i.Current()
	if !ok {
		return nil
	}
	return k
}

func (i *Iterator) Value() []byte {
	_, v, ok := i.Current()
	if !ok {
		return nil
	}
	return v
}

func (i *Iterator) Valid() bool {
	_, _, ok := i.Current()
	return ok
}

func (i *Iterator) Close() error {
	if i.e != nil {
		i.e.Close()
	}
	i.e = nil
	return nil
}
