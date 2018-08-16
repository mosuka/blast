// Copyright 2016 ikawaha
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
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/ikawaha/kagome.ipadic/internal/dic"
)

// UserDic represents a user dictionary.
type UserDic struct {
	dic *dic.UserDic
}

// NewUserDic build a user dictionary from a file.
func NewUserDic(path string) (UserDic, error) {
	f, err := os.Open(path)
	if err != nil {
		return UserDic{}, err
	}
	defer f.Close()

	r, err := NewUserDicRecords(f)
	if err != nil {
		return UserDic{}, err
	}
	return r.NewUserDic()
}

// UserDicRecord represents a record of the user dictionary file format.
type UserDicRecord struct {
	Text   string   `json:"text"`
	Tokens []string `json:"tokens"`
	Yomi   []string `json:"yomi"`
	Pos    string   `json:"pos"`
}

// UserDicRecords represents user dictionary data.
type UserDicRecords []UserDicRecord

func (u UserDicRecords) Len() int           { return len(u) }
func (u UserDicRecords) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u UserDicRecords) Less(i, j int) bool { return u[i].Text < u[j].Text }

// NewUserDicRecords loads user dictionary data from io.Reader.
func NewUserDicRecords(r io.Reader) (UserDicRecords, error) {
	var text []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		text = append(text, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var records UserDicRecords
	for _, line := range text {
		vec := strings.Split(line, ",")
		if len(vec) != dic.UserDicColumnSize {
			return nil, fmt.Errorf("invalid format: %s", line)
		}
		tokens := strings.Split(vec[1], " ")
		yomi := strings.Split(vec[2], " ")
		if len(tokens) == 0 || len(tokens) != len(yomi) {
			return nil, fmt.Errorf("invalid format: %s", line)
		}
		r := UserDicRecord{
			Text:   vec[0],
			Tokens: tokens,
			Yomi:   yomi,
			Pos:    vec[3],
		}
		records = append(records, r)
	}
	return records, nil
}

// NewUserDic builds a user dictionary.
func (u UserDicRecords) NewUserDic() (UserDic, error) {
	udic := new(dic.UserDic)
	sort.Sort(u)

	prev := ""
	keys := make([]string, 0, len(u))
	for _, r := range u {
		k := strings.TrimSpace(r.Text)
		if prev == k {
			return UserDic{}, fmt.Errorf("duplicated error, %+v", r)
		}
		prev = k
		keys = append(keys, k)
		if len(r.Tokens) == 0 || len(r.Tokens) != len(r.Yomi) {
			return UserDic{}, fmt.Errorf("invalid format, %+v", r)
		}
		c := dic.UserDicContent{
			Tokens: r.Tokens,
			Yomi:   r.Yomi,
			Pos:    r.Pos,
		}
		udic.Contents = append(udic.Contents, c)
	}
	idx, err := dic.BuildIndexTable(keys)
	udic.Index = idx
	return UserDic{dic: udic}, err
}
