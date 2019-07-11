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

package config

import (
	"encoding/json"
	"fmt"

	"github.com/mosuka/blast/strutils"
)

type NodeConfig struct {
	NodeId          string `json:"node_id,omitempty"`
	BindAddr        string `json:"bind_addr,omitempty"`
	GRPCAddr        string `json:"grpc_addr,omitempty"`
	HTTPAddr        string `json:"http_addr,omitempty"`
	DataDir         string `json:"data_dir,omitempty"`
	RaftStorageType string `json:"raft_storage_type,omitempty"`
}

func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		NodeId:          fmt.Sprintf("node-%s", strutils.RandStr(5)),
		BindAddr:        ":2000",
		GRPCAddr:        ":5000",
		HTTPAddr:        ":8000",
		DataDir:         "/tmp/blast",
		RaftStorageType: "boltdb",
	}
}

func (c *NodeConfig) ToMap() map[string]interface{} {
	b, err := json.Marshal(c)
	if err != nil {
		return map[string]interface{}{}
	}

	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return map[string]interface{}{}
	}

	return m
}
