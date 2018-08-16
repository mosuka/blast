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

// ConnectionTable represents a connection matrix of morphs.
type ConnectionTable struct {
	Row, Col int64
	Vec      []int16
}

// At returns the connection cost of matrix[row, col].
func (t *ConnectionTable) At(row, col int) int16 {
	return t.Vec[t.Col*int64(row)+int64(col)]
}

// WriteTo implements the io.WriterTo interface
func (t ConnectionTable) WriteTo(w io.Writer) (n int64, err error) {
	if err = binary.Write(w, binary.LittleEndian, t.Row); err != nil {
		return
	}
	n += int64(binary.Size(t.Row))
	if err = binary.Write(w, binary.LittleEndian, t.Col); err != nil {
		return
	}
	n += int64(binary.Size(t.Col))
	for i := range t.Vec {
		if err = binary.Write(w, binary.LittleEndian, t.Vec[i]); err != nil {
			return n, err
		}
		n += int64(binary.Size(t.Vec[i]))
	}
	return
}

// LoadConnectionTable loads ConnectionTable from io.Reader.
func LoadConnectionTable(r io.Reader) (t ConnectionTable, err error) {
	if err = binary.Read(r, binary.LittleEndian, &t.Row); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &t.Col); err != nil {
		return
	}
	t.Vec = make([]int16, t.Row*t.Col)
	for i := range t.Vec {
		if err = binary.Read(r, binary.LittleEndian, &t.Vec[i]); err != nil {
			return
		}
	}
	return
}
