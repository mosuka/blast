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

package tokenizer

import (
	"fmt"
	"strings"

	"github.com/ikawaha/kagome.ipadic/internal/dic"
	"github.com/ikawaha/kagome.ipadic/internal/lattice"
)

// TokenClass represents the token type.
type TokenClass lattice.NodeClass

const (
	// DUMMY represents the dummy token.
	DUMMY = TokenClass(lattice.DUMMY)
	// KNOWN represents the token in the dictionary.
	KNOWN = TokenClass(lattice.KNOWN)
	// UNKNOWN represents the token which is not in the dictionary.
	UNKNOWN = TokenClass(lattice.UNKNOWN)
	// USER represents the token in the user dictionary.
	USER = TokenClass(lattice.USER)
)

func (c TokenClass) String() string {
	ret := ""
	switch c {
	case DUMMY:
		ret = "DUMMY"
	case KNOWN:
		ret = "KNOWN"
	case UNKNOWN:
		ret = "UNKNOWN"
	case USER:
		ret = "USER"
	}
	return ret
}

// Token represents a morph of a sentence.
type Token struct {
	ID      int
	Class   TokenClass
	Start   int
	End     int
	Surface string
	dic     *dic.Dic
	udic    *dic.UserDic
}

// Features returns contents of a token.
func (t Token) Features() (features []string) {
	switch lattice.NodeClass(t.Class) {
	case lattice.DUMMY:
		return
	case lattice.KNOWN:
		features = t.dic.POSTable.GetPOSName(t.dic.POSTable.POSs[t.ID])
		if t.dic.Contents != nil {
			features = append(features, t.dic.Contents[t.ID]...)
		}
	case lattice.UNKNOWN:
		features = t.dic.UnkContents[t.ID]
	case lattice.USER:
		pos := t.udic.Contents[t.ID].Pos
		tokens := strings.Join(t.udic.Contents[t.ID].Tokens, "/")
		yomi := strings.Join(t.udic.Contents[t.ID].Yomi, "/")
		features = append(features, pos, tokens, yomi)
	}
	return
}

// Pos returns the first element of features.
func (t Token) Pos() string {
	f := t.Features()
	if len(f) < 1 {
		return ""
	}
	return f[0]
}

// String returns a string representation of a token.
func (t Token) String() string {
	return fmt.Sprintf("%v(%v, %v)%v[%v]", t.Surface, t.Start, t.End, t.Class, t.ID)
}
