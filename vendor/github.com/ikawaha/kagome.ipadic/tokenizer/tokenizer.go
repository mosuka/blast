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
	"io"
	"unicode/utf8"

	"github.com/ikawaha/kagome.ipadic/internal/dic"
	"github.com/ikawaha/kagome.ipadic/internal/lattice"
)

// TokenizeMode represents a mode of tokenize.
type TokenizeMode int

const (
	// Normal is the normal tokenize mode.
	Normal TokenizeMode = iota + 1
	// Search is the tokenize mode for search.
	Search
	// Extended is the experimental tokenize mode.
	Extended
	// BosEosID means the beginning a sentence or the end of a sentence.
	BosEosID = lattice.BosEosID
)

// Tokenizer represents morphological analyzer.
type Tokenizer struct {
	dic  *dic.Dic     // system dictionary
	udic *dic.UserDic // user dictionary
}

// New create a default tokenize.
func New() (t Tokenizer) {
	return Tokenizer{dic: dic.SysDic()}
}

// NewWithDic create a tokenizer with specified dictionary.
func NewWithDic(d Dic) (t Tokenizer) {
	return Tokenizer{dic: d.dic}
}

// SetDic sets dictionary to dic.
func (t *Tokenizer) SetDic(d Dic) {
	t.dic = d.dic
}

// SetUserDic sets user dictionary to udic.
func (t *Tokenizer) SetUserDic(d UserDic) {
	t.udic = d.dic
}

// Tokenize analyze a sentence in standard tokenize mode.
func (t Tokenizer) Tokenize(input string) []Token {
	return t.Analyze(input, Normal)
}

// Analyze tokenizes a sentence in the specified mode.
func (t Tokenizer) Analyze(input string, mode TokenizeMode) (tokens []Token) {
	la := lattice.New(t.dic, t.udic)
	defer la.Free()
	la.Build(input)
	m := lattice.Normal
	switch mode {
	case Normal:
		m = lattice.Normal
	case Search:
		m = lattice.Search
	case Extended:
		m = lattice.Extended
	}
	la.Forward(m)
	la.Backward(m)
	size := len(la.Output)
	tokens = make([]Token, 0, size)
	for i := range la.Output {
		n := la.Output[size-1-i]
		tok := Token{
			ID:      n.ID,
			Class:   TokenClass(n.Class),
			Start:   n.Start,
			End:     n.Start + utf8.RuneCountInString(n.Surface),
			Surface: n.Surface,
			dic:     t.dic,
			udic:    t.udic,
		}
		if tok.ID == lattice.BosEosID {
			if i == 0 {
				tok.Surface = "BOS"
			} else {
				tok.Surface = "EOS"
			}
		}
		tokens = append(tokens, tok)
	}
	return
}

// Dot returns morphs of a sentense and exports a lattice graph to dot format in standard tokenize mode.
func (t Tokenizer) Dot(input string, w io.Writer) (tokens []Token) {
	return t.AnalyzeGraph(input, Normal, w)
}

// AnalyzeGraph returns morphs of a sentense and exports a lattice graph to dot format.
func (t Tokenizer) AnalyzeGraph(input string, mode TokenizeMode, w io.Writer) (tokens []Token) {
	la := lattice.New(t.dic, t.udic)
	defer la.Free()
	la.Build(input)
	m := lattice.Normal
	switch mode {
	case Normal:
		m = lattice.Normal
	case Search:
		m = lattice.Search
	case Extended:
		m = lattice.Extended
	}
	la.Forward(m)
	la.Backward(m)
	size := len(la.Output)
	tokens = make([]Token, 0, size)
	for i := range la.Output {
		n := la.Output[size-1-i]
		tok := Token{
			ID:      n.ID,
			Class:   TokenClass(n.Class),
			Start:   n.Start,
			End:     n.Start + utf8.RuneCountInString(n.Surface),
			Surface: n.Surface,
			dic:     t.dic,
			udic:    t.udic,
		}
		if tok.ID == lattice.BosEosID {
			if i == 0 {
				tok.Surface = "BOS"
			} else {
				tok.Surface = "EOS"
			}
		}
		tokens = append(tokens, tok)
	}
	la.Dot(w)
	return
}
