// Copyright 2015 ikawaha
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dic

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"sort"

	"github.com/ikawaha/kagome.ipadic/internal/da"
)

// IndexTable represents a dictionary index.
type IndexTable struct {
	Da  da.DoubleArray
	Dup map[int32]int32
}

// BuildIndexTable constructs a index table from keywords.
func BuildIndexTable(sortedKeywords []string) (IndexTable, error) {
	idx := IndexTable{Dup: map[int32]int32{}}
	if !sort.StringsAreSorted(sortedKeywords) {
		return idx, fmt.Errorf("unsorted keywords")
	}
	keys := make([]string, 0, len(sortedKeywords))
	ids := make([]int, 0, len(sortedKeywords))
	prev := struct {
		no   int
		word string
	}{}
	for i, key := range sortedKeywords {
		if key == prev.word {
			idx.Dup[int32(prev.no)]++
			continue
		}
		prev.no = i
		prev.word = key
		keys = append(keys, key)
		ids = append(ids, i)
	}
	d, err := da.BuildWithIDs(keys, ids)
	if err != nil {
		return idx, fmt.Errorf("build error, %v", err)
	}
	idx.Da = d
	return idx, nil
}

// CommonPrefixSearch finds keywords sharing common prefix in an input
// and returns the ids and it's lengths if found.
func (idx IndexTable) CommonPrefixSearch(input string) (lens []int, ids [][]int) {
	seeds, lens := idx.Da.CommonPrefixSearch(input)
	for _, id := range seeds {
		dup, _ := idx.Dup[int32(id)]
		list := make([]int, 1+dup, 1+dup)
		for i := 0; i < len(list); i++ {
			list[i] = id + i
		}
		ids = append(ids, list)
	}
	return
}

// CommonPrefixSearchCallback finds keywords sharing common prefix in an input
// and callback with id and length.
func (idx IndexTable) CommonPrefixSearchCallback(input string, callback func(id, l int)) {
	idx.Da.CommonPrefixSearchCallback(input, func(x, y int) {
		dup := idx.Dup[int32(x)]
		for i := x; i < x+int(dup)+1; i++ {
			callback(i, y)
		}
	})
	return
}

// Search finds the given keyword and returns the id if found.
func (idx IndexTable) Search(input string) []int {
	id, ok := idx.Da.Find(input)
	if !ok {
		return nil
	}
	dup, _ := idx.Dup[int32(id)]
	list := make([]int, 1+dup, 1+dup)
	for i := 0; i < len(list); i++ {
		list[i] = id + i
	}
	return list
}

// WriteTo saves a index table.
func (idx IndexTable) WriteTo(w io.Writer) (n int64, err error) {
	n, err = idx.Da.WriteTo(w)
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err = enc.Encode(idx.Dup); err != nil {
		return
	}
	x, err := b.WriteTo(w)
	if err != nil {
		return
	}
	n += x
	return
}

// ReadIndexTable loads a index table.
func ReadIndexTable(r io.Reader) (IndexTable, error) {
	idx := IndexTable{}
	d, err := da.Read(r)
	if err != nil {
		return idx, fmt.Errorf("read index error, %v", err)
	}
	idx.Da = d

	dec := gob.NewDecoder(r)
	if e := dec.Decode(&idx.Dup); e != nil {
		return idx, fmt.Errorf("read index dup table error, %v", e)
	}

	return idx, nil
}
