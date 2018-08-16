// Copyright 2017 ikawaha
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
)

type POSTable struct {
	POSs     []POS
	NameList []string
}

// POSID represents a ID of part of speech.
type POSID int16

// POS represents a vector of part of speech.
type POS []POSID

// String returns a vector of part of speech name.
func (p POSTable) GetPOSName(pos POS) []string {
	ret := make([]string, 0, len(pos))
	for _, id := range pos {
		ret = append(ret, p.NameList[id])
	}
	return ret
}

// WriteTo saves a POS table.
func (p POSTable) WriteTo(w io.Writer) (int64, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	if err := enc.Encode(p.POSs); err != nil {
		return 0, err
	}
	if err := enc.Encode(p.NameList); err != nil {
		return 0, err
	}
	return b.WriteTo(w)
}

// ReadPOSTable loads a POS table.
func ReadPOSTable(r io.Reader) (POSTable, error) {
	t := POSTable{}
	dec := gob.NewDecoder(r)
	if err := dec.Decode(&t.POSs); err != nil {
		return t, fmt.Errorf("POSs read error, %v", err)
	}
	if err := dec.Decode(&t.NameList); err != nil {
		return t, fmt.Errorf("name list read error, %v", err)
	}
	return t, nil
}

// POSMap represents a part of speech control table.
type POSMap map[string]POSID

// Add adds part of speech item to the POS control table and returns it's id.
func (p POSMap) Add(pos []string) POS {
	ret := make(POS, 0, len(pos))
	for _, name := range pos {
		id := p.add(name)
		ret = append(ret, id)
	}
	return ret
}

func (p POSMap) add(pos string) POSID {
	id, ok := p[pos]
	if !ok {
		id = POSID(len(p)) + 1
		p[pos] = id
	}
	return id
}

// List returns a list whose index is POS ID and value is its name.
func (p POSMap) List() []string {
	ret := make([]string, len(p)+1)
	for k, v := range p {
		ret[v] = k
	}
	return ret
}
