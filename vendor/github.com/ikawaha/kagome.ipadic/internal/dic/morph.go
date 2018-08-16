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
	"encoding/binary"
	"io"
)

// Morph represents part of speeches and an occurrence cost.
type Morph struct {
	LeftID, RightID, Weight int16
}

// MorphSlice represents a slice of morphs.
type MorphSlice []Morph

// WriteTo implements the io.WriterTo interface
func (m MorphSlice) WriteTo(w io.Writer) (n int64, err error) {
	l := int64(len(m))
	if err = binary.Write(w, binary.LittleEndian, l); err != nil {
		return
	}
	n += int64(binary.Size(l))
	for i := range m {
		if err = binary.Write(w, binary.LittleEndian, m[i].LeftID); err != nil {
			return n, err
		}
		n += int64(binary.Size(m[i].LeftID))
		if err = binary.Write(w, binary.LittleEndian, m[i].RightID); err != nil {
			return
		}
		n += int64(binary.Size(m[i].RightID))
		if err = binary.Write(w, binary.LittleEndian, m[i].Weight); err != nil {
			return
		}
		n += int64(binary.Size(m[i].Weight))
	}
	return
}

// LoadMorphSlice loads morph data from io.Reader
func LoadMorphSlice(r io.Reader) ([]Morph, error) {
	var l int64
	if e := binary.Read(r, binary.LittleEndian, &l); e != nil {
		return nil, e
	}
	m := make([]Morph, l)
	for i := range m {
		if e := binary.Read(r, binary.LittleEndian, &m[i].LeftID); e != nil {
			return m, e
		}
		if e := binary.Read(r, binary.LittleEndian, &m[i].RightID); e != nil {
			return m, e
		}
		if e := binary.Read(r, binary.LittleEndian, &m[i].Weight); e != nil {
			return m, e
		}
	}
	return m, nil
}
