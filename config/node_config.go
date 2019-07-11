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
	"fmt"

	"github.com/mosuka/blast/maputils"
	"github.com/mosuka/blast/strutils"
)

type NodeConfig struct {
	maputils.Map
}

func DefaultNodeConfig() NodeConfig {
	c := NodeConfig{Map: maputils.New()}
	_ = c.SetNodeId(fmt.Sprintf("node-%s", strutils.RandStr(5)))
	_ = c.SetBindAddr(":2000")
	_ = c.SetGrpcAddr(":5000")
	_ = c.SetHttpAddr(":8000")
	_ = c.SetDataDir("/tmp/blast")
	_ = c.SetRaftStorageType("boltdb")

	return c
}

func NewNodeConfigFromMap(src map[string]interface{}) NodeConfig {
	return NodeConfig{Map: maputils.FromMap(src)}
}

func NewNodeConfigFromYAML(src []byte) (NodeConfig, error) {
	m, err := maputils.FromYAML(src)
	if err != nil {
		return DefaultNodeConfig(), err
	}

	return NodeConfig{Map: m}, nil
}

func NewNodeConfigFromJSON(src []byte) (NodeConfig, error) {
	m, err := maputils.FromJSON(src)
	if err != nil {
		return DefaultNodeConfig(), err
	}

	return NodeConfig{Map: m}, nil
}

func (c NodeConfig) SetNodeId(nodeId string) error {
	err := c.Set("/node_id", nodeId)
	if err != nil {
		return err
	}
	return nil
}

func (c NodeConfig) GetNodeId() (string, error) {
	nodeId, err := c.Get("/node_id")
	if err != nil {
		return "", err
	}
	return nodeId.(string), nil
}

func (c NodeConfig) SetBindAddr(bindAddr string) error {
	err := c.Set("/bind_addr", bindAddr)
	if err != nil {
		return err
	}
	return nil
}

func (c NodeConfig) GetBindAddr() (string, error) {
	bindAddr, err := c.Get("/bind_addr")
	if err != nil {
		return "", err
	}
	return bindAddr.(string), nil
}

func (c NodeConfig) SetGrpcAddr(grpcAddr string) error {
	err := c.Set("/grpc_addr", grpcAddr)
	if err != nil {
		return err
	}
	return nil
}

func (c NodeConfig) GetGrpcAddr() (string, error) {
	grpcAddr, err := c.Get("/grpc_addr")
	if err != nil {
		return "", err
	}
	return grpcAddr.(string), nil
}

func (c NodeConfig) SetHttpAddr(httpAddr string) error {
	err := c.Set("/http_addr", httpAddr)
	if err != nil {
		return err
	}
	return nil
}

func (c NodeConfig) GetHttpAddr() (string, error) {
	httpAddr, err := c.Get("/http_addr")
	if err != nil {
		return "", err
	}
	return httpAddr.(string), nil
}

func (c NodeConfig) SetDataDir(dataDir string) error {
	err := c.Set("/data_dir", dataDir)
	if err != nil {
		return err
	}
	return nil
}

func (c NodeConfig) GetDataDir() (string, error) {
	dataDir, err := c.Get("/data_dir")
	if err != nil {
		return "", err
	}
	return dataDir.(string), nil
}

func (c NodeConfig) SetRaftStorageType(raftStorageType string) error {
	err := c.Set("/raft_storage_type", raftStorageType)
	if err != nil {
		return err
	}
	return nil
}

func (c NodeConfig) GetRaftStorageType() (string, error) {
	raftStorageType, err := c.Get("/raft_storage_type")
	if err != nil {
		return "", err
	}
	return raftStorageType.(string), nil
}
