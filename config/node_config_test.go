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
	"reflect"
	"testing"
)

func TestFromMap(t *testing.T) {
	src := map[string]interface{}{
		"node_id":           "node1",
		"bind_addr":         ":2000",
		"grpc_addr":         ":5000",
		"http_addr":         ":8000",
		"data_dir":          "/tmp/blast",
		"raft_storage_type": "boltdb",
	}

	cfg := NewNodeConfigFromMap(src)
	expCfg := DefaultNodeConfig()
	_ = expCfg.SetNodeId("node1")
	_ = expCfg.SetBindAddr(":2000")
	_ = expCfg.SetGrpcAddr(":5000")
	_ = expCfg.SetHttpAddr(":8000")
	_ = expCfg.SetDataDir("/tmp/blast")
	_ = expCfg.SetRaftStorageType("boltdb")
	actCfg := cfg

	if !reflect.DeepEqual(expCfg, actCfg) {
		t.Fatalf("expected content to see %v, saw %v", expCfg, actCfg)
	}
}

func TestNodeConfig_ToMap(t *testing.T) {
	src := map[string]interface{}{
		"node_id":           "node1",
		"bind_addr":         ":2000",
		"grpc_addr":         ":5000",
		"http_addr":         ":8000",
		"data_dir":          "/tmp/blast",
		"raft_storage_type": "boltdb",
	}

	cfg := NewNodeConfigFromMap(src)
	expMap := src
	actMap := cfg.ToMap()

	if !reflect.DeepEqual(expMap, actMap) {
		t.Fatalf("expected content to see %v, saw %v", expMap, actMap)
	}
}

func TestFromYAML(t *testing.T) {
	src := []byte(`bind_addr: :2000
data_dir: /tmp/blast
grpc_addr: :5000
http_addr: :8000
node_id: node1
raft_storage_type: boltdb
`)

	cfg, err := NewNodeConfigFromYAML(src)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCfg := DefaultNodeConfig()
	_ = expCfg.SetNodeId("node1")
	_ = expCfg.SetBindAddr(":2000")
	_ = expCfg.SetGrpcAddr(":5000")
	_ = expCfg.SetHttpAddr(":8000")
	_ = expCfg.SetDataDir("/tmp/blast")
	_ = expCfg.SetRaftStorageType("boltdb")
	actCfg := cfg

	if !reflect.DeepEqual(expCfg, actCfg) {
		t.Fatalf("expected content to see %v, saw %v", expCfg, actCfg)
	}
}

func TestNodeConfig_ToYAML(t *testing.T) {
	src := []byte(`bind_addr: :2000
data_dir: /tmp/blast
grpc_addr: :5000
http_addr: :8000
node_id: node1
raft_storage_type: boltdb
`)

	cfg, err := NewNodeConfigFromYAML(src)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expYAML := src
	actYAML, err := cfg.ToYAML()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(expYAML, actYAML) {
		t.Fatalf("expected content to see %v, saw %v", expYAML, actYAML)
	}
}

func TestFromJSON(t *testing.T) {
	src := []byte(`{"bind_addr":":2000","data_dir":"/tmp/blast","grpc_addr":":5000","http_addr":":8000","node_id":"node1","raft_storage_type":"boltdb"}`)

	cfg, err := NewNodeConfigFromJSON(src)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCfg := DefaultNodeConfig()
	_ = expCfg.SetNodeId("node1")
	_ = expCfg.SetBindAddr(":2000")
	_ = expCfg.SetGrpcAddr(":5000")
	_ = expCfg.SetHttpAddr(":8000")
	_ = expCfg.SetDataDir("/tmp/blast")
	_ = expCfg.SetRaftStorageType("boltdb")
	actCfg := cfg

	if !reflect.DeepEqual(expCfg, actCfg) {
		t.Fatalf("expected content to see %v, saw %v", expCfg, actCfg)
	}
}

func TestNodeConfig_ToJSON(t *testing.T) {
	src := []byte(`{"bind_addr":":2000","data_dir":"/tmp/blast","grpc_addr":":5000","http_addr":":8000","node_id":"node1","raft_storage_type":"boltdb"}`)

	cfg, err := NewNodeConfigFromJSON(src)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expJSON := src
	actJSON, err := cfg.ToJSON()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(expJSON, actJSON) {
		t.Fatalf("expected content to see %v, saw %v", expJSON, actJSON)
	}
}

func TestNodeConfig_SetNodeId(t *testing.T) {
	expNodeId := "node1"
	cfg := DefaultNodeConfig()
	err := cfg.SetNodeId(expNodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actNodeId, err := cfg.GetNodeId()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expNodeId != actNodeId {
		t.Fatalf("expected content to see %v, saw %v", expNodeId, actNodeId)
	}
}

func TestNodeConfig_GetNodeId(t *testing.T) {
	expNodeId := "node1"
	cfg := DefaultNodeConfig()
	err := cfg.SetNodeId(expNodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actNodeId, err := cfg.GetNodeId()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expNodeId != actNodeId {
		t.Fatalf("expected content to see %v, saw %v", expNodeId, actNodeId)
	}
}

func TestNodeConfig_SetBindAddr(t *testing.T) {
	expBindAddr := ":2000"
	cfg := DefaultNodeConfig()
	err := cfg.SetBindAddr(expBindAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actBindAddr, err := cfg.GetBindAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expBindAddr != actBindAddr {
		t.Fatalf("expected content to see %v, saw %v", expBindAddr, actBindAddr)
	}
}

func TestNodeConfig_GetBindAddr(t *testing.T) {
	expBindAddr := ":2000"
	cfg := DefaultNodeConfig()
	err := cfg.SetBindAddr(expBindAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actBindAddr, err := cfg.GetBindAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expBindAddr != actBindAddr {
		t.Fatalf("expected content to see %v, saw %v", expBindAddr, actBindAddr)
	}
}

func TestNodeConfig_SetGrpcAddr(t *testing.T) {
	expGrpcAddr := ":5000"
	cfg := DefaultNodeConfig()
	err := cfg.SetGrpcAddr(expGrpcAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actGrpcAddr, err := cfg.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expGrpcAddr != actGrpcAddr {
		t.Fatalf("expected content to see %v, saw %v", expGrpcAddr, actGrpcAddr)
	}
}

func TestNodeConfig_GetGrpcAddr(t *testing.T) {
	expGrpcAddr := ":5000"
	cfg := DefaultNodeConfig()
	err := cfg.SetGrpcAddr(expGrpcAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actGrpcAddr, err := cfg.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expGrpcAddr != actGrpcAddr {
		t.Fatalf("expected content to see %v, saw %v", expGrpcAddr, actGrpcAddr)
	}
}

func TestNodeConfig_SetHttpAddr(t *testing.T) {
	expHttpAddr := ":8000"
	cfg := DefaultNodeConfig()
	err := cfg.SetHttpAddr(expHttpAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actHttpAddr, err := cfg.GetHttpAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expHttpAddr != actHttpAddr {
		t.Fatalf("expected content to see %v, saw %v", expHttpAddr, actHttpAddr)
	}
}

func TestNodeConfig_GetHttpAddr(t *testing.T) {
	expHttpAddr := ":8000"
	cfg := DefaultNodeConfig()
	err := cfg.SetHttpAddr(expHttpAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actHttpAddr, err := cfg.GetHttpAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expHttpAddr != actHttpAddr {
		t.Fatalf("expected content to see %v, saw %v", expHttpAddr, actHttpAddr)
	}
}

func TestNodeConfig_SetDataDir(t *testing.T) {
	expDataDir := "/tmp/blast"
	cfg := DefaultNodeConfig()
	err := cfg.SetDataDir(expDataDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actDataDir, err := cfg.GetDataDir()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expDataDir != actDataDir {
		t.Fatalf("expected content to see %v, saw %v", expDataDir, actDataDir)
	}
}

func TestNodeConfig_GetDataDir(t *testing.T) {
	expDataDir := "/tmp/blast"
	cfg := DefaultNodeConfig()
	err := cfg.SetDataDir(expDataDir)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actDataDir, err := cfg.GetDataDir()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expDataDir != actDataDir {
		t.Fatalf("expected content to see %v, saw %v", expDataDir, actDataDir)
	}
}

func TestNodeConfig_SetRaftStorageType(t *testing.T) {
	expRaftStorageType := "boltdb"
	cfg := DefaultNodeConfig()
	err := cfg.SetRaftStorageType(expRaftStorageType)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actRaftStorageType, err := cfg.GetRaftStorageType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expRaftStorageType != actRaftStorageType {
		t.Fatalf("expected content to see %v, saw %v", expRaftStorageType, actRaftStorageType)
	}
}

func TestNodeConfig_GetRaftStorageType(t *testing.T) {
	expRaftStorageType := "boltdb"
	cfg := DefaultNodeConfig()
	err := cfg.SetRaftStorageType(expRaftStorageType)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actRaftStorageType, err := cfg.GetRaftStorageType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if expRaftStorageType != actRaftStorageType {
		t.Fatalf("expected content to see %v, saw %v", expRaftStorageType, actRaftStorageType)
	}
}
