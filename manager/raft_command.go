// Copyright (c) 2019 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import "encoding/json"

type command int

const (
	unknown command = iota
	setNode
	deleteNode
	setKeyValue
	deleteKeyValue
)

type message struct {
	Command command         `json:"command,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func newMessage(cmd command, data interface{}) (*message, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &message{
		Command: cmd,
		Data:    b,
	}, nil
}
