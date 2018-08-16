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
	"fmt"
	"io"
	"strings"
)

const (
	rowDelimiter = "\n"
	colDelimiter = "\a"
)

// Contents represents dictionary contents.
type Contents [][]string

// WriteTo implements the io.WriterTo interface
func (c Contents) WriteTo(w io.Writer) (n int64, err error) {
	for i := 0; i < len(c)-1; i++ {
		x, e := fmt.Fprintf(w, "%s%s", strings.Join(c[i], colDelimiter), rowDelimiter)
		if e != nil {
			return
		}
		n += int64(x)
	}
	if i := len(c) - 1; i > 0 {
		x, e := fmt.Fprintf(w, "%s", strings.Join(c[i], colDelimiter))
		if e != nil {
			return n, e
		}
		n += int64(x)
	}
	return
}

// NewContents creates dictionary contents from byte slice
func NewContents(b []byte) [][]string {
	str := string(b)
	rows := strings.Split(str, rowDelimiter)
	m := make([][]string, len(rows))
	for i, r := range rows {
		m[i] = strings.Split(r, colDelimiter)
	}
	return m
}
