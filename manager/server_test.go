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

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/config"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/testutils"
)

func TestServer_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)
}

func TestServer_LivenessProbe(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// liveness
	liveness, err := client.LivenessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness := liveness
	if expLiveness != actLiveness {
		t.Fatalf("expected content to see %v, saw %v", expLiveness, actLiveness)
	}
}

func TestServer_ReadinessProbe(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// readiness
	readiness, err := client.ReadinessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness := readiness
	if expReadiness != actReadiness {
		t.Fatalf("expected content to see %v, saw %v", expReadiness, actReadiness)
	}
}

func TestServer_GetNode(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	nodeId, err := server.nodeConfig.GetNodeId()
	if err != nil {
		t.Fatalf("%v", err)
	}
	bindAddr, err := server.nodeConfig.GetBindAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	httpAddr, err := server.nodeConfig.GetHttpAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	dataDir, err := server.nodeConfig.GetDataDir()
	if err != nil {
		t.Fatalf("%v", err)
	}
	raftStorageType, err := server.nodeConfig.GetRaftStorageType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get node
	node, err := client.GetNode(nodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode := map[string]interface{}{
		"node_config": map[string]interface{}{
			"node_id":           nodeId,
			"bind_addr":         bindAddr,
			"grpc_addr":         grpcAddr,
			"http_addr":         httpAddr,
			"data_dir":          dataDir,
			"raft_storage_type": raftStorageType,
		},
		"state": "Leader",
	}
	actNode := node
	if !reflect.DeepEqual(expNode, actNode) {
		t.Errorf("expected content to see %v, saw %v", expNode, actNode)
	}
}

func TestServer_GetCluster(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	nodeId, err := server.nodeConfig.GetNodeId()
	if err != nil {
		t.Fatalf("%v", err)
	}
	bindAddr, err := server.nodeConfig.GetBindAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	httpAddr, err := server.nodeConfig.GetHttpAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	dataDir, err := server.nodeConfig.GetDataDir()
	if err != nil {
		t.Fatalf("%v", err)
	}
	raftStorageType, err := server.nodeConfig.GetRaftStorageType()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get cluster
	cluster, err := client.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster := map[string]interface{}{
		nodeId: map[string]interface{}{
			"node_config": map[string]interface{}{
				"node_id":           nodeId,
				"bind_addr":         bindAddr,
				"grpc_addr":         grpcAddr,
				"http_addr":         httpAddr,
				"data_dir":          dataDir,
				"raft_storage_type": raftStorageType,
			},
			"state": "Leader",
		},
	}
	actCluster := cluster
	if !reflect.DeepEqual(expCluster, actCluster) {
		t.Errorf("expected content to see %v, saw %v", expCluster, actCluster)
	}
}

func TestServer_GetIndexMapping(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	expIndexMapping, err := server.indexConfig.GetIndexMapping()
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIntr, err := client.GetState("index_config/index_mapping")
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexMapping, err := indexutils.NewIndexMappingFromMap(*actIntr.(*map[string]interface{}))
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(expIndexMapping, actIndexMapping) {
		t.Errorf("expected content to see %v, saw %v", expIndexMapping, actIndexMapping)
	}
}

func TestServer_GetIndexType(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	expIndexType, err := server.indexConfig.GetIndexType()
	if err != nil {
		t.Errorf("%v", err)
	}

	actIndexType, err := client.GetState("index_config/index_type")
	if err != nil {
		t.Errorf("%v", err)
	}

	if expIndexType != *actIndexType.(*string) {
		t.Errorf("expected content to see %v, saw %v", expIndexType, *actIndexType.(*string))
	}
}

func TestServer_GetIndexStorageType(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	expIndexStorageType, err := server.indexConfig.GetIndexStorageType()
	if err != nil {
		t.Errorf("%v", err)
	}

	actIndexStorageType, err := client.GetState("index_config/index_storage_type")
	if err != nil {
		t.Errorf("%v", err)
	}

	if expIndexStorageType != *actIndexStorageType.(*string) {
		t.Errorf("expected content to see %v, saw %v", expIndexStorageType, *actIndexStorageType.(*string))
	}
}

func TestServer_SetState(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// set value
	err = client.SetState("test/key1", "val1")
	if err != nil {
		t.Errorf("%v", err)
	}

	// get value
	val1, err := client.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}

	expVal1 := "val1"

	actVal1 := *val1.(*string)

	if expVal1 != actVal1 {
		t.Errorf("expected content to see %v, saw %v", expVal1, actVal1)
	}
}

func TestServer_GetState(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// set value
	err = client.SetState("test/key1", "val1")
	if err != nil {
		t.Errorf("%v", err)
	}

	// get value
	val1, err := client.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}

	expVal1 := "val1"

	actVal1 := *val1.(*string)

	if expVal1 != actVal1 {
		t.Errorf("expected content to see %v, saw %v", expVal1, actVal1)
	}
}

func TestServer_DeleteState(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server != nil {
			server.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	grpcAddr, err := server.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		if client != nil {
			err = client.Close()
			if err != nil {
				t.Fatalf("%v", err)
			}
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// set value
	err = client.SetState("test/key1", "val1")
	if err != nil {
		t.Errorf("%v", err)
	}

	// get value
	val1, err := client.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}

	expVal1 := "val1"

	actVal1 := *val1.(*string)

	if expVal1 != actVal1 {
		t.Errorf("expected content to see %v, saw %v", expVal1, actVal1)
	}

	// delete value
	err = client.DeleteState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}

	val1, err = client.GetState("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}

	if val1 != nil {
		t.Errorf("%v", err)
	}

	// delete non-existing data
	err = client.DeleteState("test/non-existing")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
}

func TestCluster_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)
}

func TestCluster_LivenessProbe(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	grpcAddr1, err := server1.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := grpc.NewClient(grpcAddr1)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr2, err := server2.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(grpcAddr2)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr3, err := server3.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(grpcAddr3)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// liveness check for manager1
	liveness1, err := client1.LivenessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expLiveness1 := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness1 := liveness1
	if expLiveness1 != actLiveness1 {
		t.Errorf("expected content to see %v, saw %v", expLiveness1, actLiveness1)
	}

	// liveness check for manager2
	liveness2, err := client2.LivenessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expLiveness2 := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness2 := liveness2
	if expLiveness2 != actLiveness2 {
		t.Errorf("expected content to see %v, saw %v", expLiveness2, actLiveness2)
	}

	// liveness check for manager3
	liveness3, err := client3.LivenessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expLiveness3 := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness3 := liveness3
	if expLiveness3 != actLiveness3 {
		t.Errorf("expected content to see %v, saw %v", expLiveness3, actLiveness3)
	}
}

func TestCluster_ReadinessProbe(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	grpcAddr1, err := server1.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := grpc.NewClient(grpcAddr1)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr2, err := server2.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(grpcAddr2)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr3, err := server3.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(grpcAddr3)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// readiness check for manager1
	readiness1, err := client1.ReadinessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expReadiness1 := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness1 := readiness1
	if expReadiness1 != actReadiness1 {
		t.Errorf("expected content to see %v, saw %v", expReadiness1, actReadiness1)
	}

	// readiness check for manager2
	readiness2, err := client2.ReadinessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expReadiness2 := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness2 := readiness2
	if expReadiness2 != actReadiness2 {
		t.Errorf("expected content to see %v, saw %v", expReadiness2, actReadiness2)
	}

	// readiness check for manager2
	readiness3, err := client3.ReadinessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expReadiness3 := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness3 := readiness3
	if expReadiness3 != actReadiness3 {
		t.Errorf("expected content to see %v, saw %v", expReadiness3, actReadiness3)
	}
}

func TestCluster_GetNode(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	grpcAddr1, err := server1.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := grpc.NewClient(grpcAddr1)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr2, err := server2.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(grpcAddr2)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr3, err := server3.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(grpcAddr3)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// get all node info from all nodes
	node11, err := client1.GetNode(nodeId1)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode11 := map[string]interface{}{
		"node_config": server1.nodeConfig.ToMap(),
		"state":       raft.Leader.String(),
	}
	actNode11 := node11
	if !reflect.DeepEqual(expNode11, actNode11) {
		t.Errorf("expected content to see %v, saw %v", expNode11, actNode11)
	}

	node12, err := client1.GetNode(nodeId2)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode12 := map[string]interface{}{
		"node_config": server2.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode12 := node12
	if !reflect.DeepEqual(expNode12, actNode12) {
		t.Errorf("expected content to see %v, saw %v", expNode12, actNode12)
	}

	node13, err := client1.GetNode(nodeId3)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode13 := map[string]interface{}{
		"node_config": server3.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode13 := node13
	if !reflect.DeepEqual(expNode13, actNode13) {
		t.Errorf("expected content to see %v, saw %v", expNode13, actNode13)
	}

	node21, err := client2.GetNode(nodeId1)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode21 := map[string]interface{}{
		"node_config": server1.nodeConfig.ToMap(),
		"state":       raft.Leader.String(),
	}
	actNode21 := node21
	if !reflect.DeepEqual(expNode21, actNode21) {
		t.Errorf("expected content to see %v, saw %v", expNode21, actNode21)
	}

	node22, err := client2.GetNode(nodeId2)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode22 := map[string]interface{}{
		"node_config": server2.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode22 := node22
	if !reflect.DeepEqual(expNode22, actNode22) {
		t.Errorf("expected content to see %v, saw %v", expNode22, actNode22)
	}

	node23, err := client2.GetNode(nodeId3)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode23 := map[string]interface{}{
		"node_config": server3.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode23 := node23
	if !reflect.DeepEqual(expNode23, actNode23) {
		t.Errorf("expected content to see %v, saw %v", expNode23, actNode23)
	}

	node31, err := client3.GetNode(nodeId1)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode31 := map[string]interface{}{
		"node_config": server1.nodeConfig.ToMap(),
		"state":       raft.Leader.String(),
	}
	actNode31 := node31
	if !reflect.DeepEqual(expNode31, actNode31) {
		t.Errorf("expected content to see %v, saw %v", expNode31, actNode31)
	}

	node32, err := client3.GetNode(nodeId2)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode32 := map[string]interface{}{
		"node_config": server2.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode32 := node32
	if !reflect.DeepEqual(expNode32, actNode32) {
		t.Errorf("expected content to see %v, saw %v", expNode32, actNode32)
	}

	node33, err := client3.GetNode(nodeId3)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode33 := map[string]interface{}{
		"node_config": server3.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode33 := node33
	if !reflect.DeepEqual(expNode33, actNode33) {
		t.Errorf("expected content to see %v, saw %v", expNode33, actNode33)
	}
}

func TestCluster_GetCluster(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	grpcAddr1, err := server1.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := grpc.NewClient(grpcAddr1)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr2, err := server2.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(grpcAddr2)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr3, err := server3.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(grpcAddr3)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// get cluster info from manager1
	cluster1, err := client1.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster1 := map[string]interface{}{
		nodeId1: map[string]interface{}{
			"node_config": server1.nodeConfig.ToMap(),
			"state":       raft.Leader.String(),
		},
		nodeId2: map[string]interface{}{
			"node_config": server2.nodeConfig.ToMap(),
			"state":       raft.Follower.String(),
		},
		nodeId3: map[string]interface{}{
			"node_config": server3.nodeConfig.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actCluster1 := cluster1
	if !reflect.DeepEqual(expCluster1, actCluster1) {
		t.Errorf("expected content to see %v, saw %v", expCluster1, actCluster1)
	}

	cluster2, err := client2.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster2 := map[string]interface{}{
		nodeId1: map[string]interface{}{
			"node_config": server1.nodeConfig.ToMap(),
			"state":       raft.Leader.String(),
		},
		nodeId2: map[string]interface{}{
			"node_config": server2.nodeConfig.ToMap(),
			"state":       raft.Follower.String(),
		},
		nodeId3: map[string]interface{}{
			"node_config": server3.nodeConfig.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actCluster2 := cluster2
	if !reflect.DeepEqual(expCluster2, actCluster2) {
		t.Errorf("expected content to see %v, saw %v", expCluster2, actCluster2)
	}

	cluster3, err := client3.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster3 := map[string]interface{}{
		nodeId1: map[string]interface{}{
			"node_config": server1.nodeConfig.ToMap(),
			"state":       raft.Leader.String(),
		},
		nodeId2: map[string]interface{}{
			"node_config": server2.nodeConfig.ToMap(),
			"state":       raft.Follower.String(),
		},
		nodeId3: map[string]interface{}{
			"node_config": server3.nodeConfig.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actCluster3 := cluster3
	if !reflect.DeepEqual(expCluster3, actCluster3) {
		t.Errorf("expected content to see %v, saw %v", expCluster3, actCluster3)
	}
}

func TestCluster_GetState(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	grpcAddr1, err := server1.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := grpc.NewClient(grpcAddr1)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr2, err := server2.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(grpcAddr2)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr3, err := server3.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(grpcAddr3)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// get index mapping from all nodes
	indexConfig1, err := client1.GetState("index_config")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexConfig1 := indexConfig.ToMap()
	actIndexConfig1 := *indexConfig1.(*map[string]interface{})
	if !reflect.DeepEqual(expIndexConfig1, actIndexConfig1) {
		t.Errorf("expected content to see %v, saw %v", expIndexConfig1, actIndexConfig1)
	}

	indexConfig2, err := client2.GetState("index_config")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexConfig2 := indexConfig.ToMap()
	actIndexConfig2 := *indexConfig2.(*map[string]interface{})
	if !reflect.DeepEqual(expIndexConfig2, actIndexConfig2) {
		t.Errorf("expected content to see %v, saw %v", expIndexConfig2, actIndexConfig2)
	}

	indexConfig3, err := client3.GetState("index_config")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexConfig3 := indexConfig.ToMap()
	actIndexConfig3 := *indexConfig3.(*map[string]interface{})
	if !reflect.DeepEqual(expIndexConfig3, actIndexConfig3) {
		t.Errorf("expected content to see %v, saw %v", expIndexConfig3, actIndexConfig3)
	}
}

func TestCluster_SetState(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	grpcAddr1, err := server1.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := grpc.NewClient(grpcAddr1)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr2, err := server2.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(grpcAddr2)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr3, err := server3.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(grpcAddr3)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	err = client1.SetState("test/key1", "val1")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val11, err := client1.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if expVal11 != actVal11 {
		t.Errorf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	val21, err := client2.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if expVal21 != actVal21 {
		t.Errorf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	val31, err := client3.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if expVal31 != actVal31 {
		t.Errorf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	err = client2.SetState("test/key2", "val2")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val12, err := client1.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if expVal12 != actVal12 {
		t.Errorf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	val22, err := client2.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if expVal22 != actVal22 {
		t.Errorf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	val32, err := client3.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if expVal32 != actVal32 {
		t.Errorf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	err = client3.SetState("test/key3", "val3")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val13, err := client1.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if expVal13 != actVal13 {
		t.Errorf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	val23, err := client2.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if expVal23 != actVal23 {
		t.Errorf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	val33, err := client3.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if expVal33 != actVal33 {
		t.Errorf("expected content to see %v, saw %v", expVal33, actVal33)
	}
}

func TestCluster_DeleteState(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server1
	clusterConfig1 := config.DefaultClusterConfig()
	nodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig1.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId1, err := nodeConfig1.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named(nodeId1), grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server1
	server1.Start()

	// set peer address
	peerAddr, err := nodeConfig1.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	err = clusterConfig2.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig2.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId2, err := nodeConfig2.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named(nodeId2), grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server2
	server2.Start()

	// create configs for server3
	clusterConfig3 := config.DefaultClusterConfig()
	err = clusterConfig3.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		if dataDir, err := nodeConfig3.GetDataDir(); err != nil {
			_ = os.RemoveAll(dataDir)
		}
	}()
	nodeId3, err := nodeConfig3.GetNodeId()
	if err != nil {
		t.Errorf("%v", err)
	}
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named(nodeId3), grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server3
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	grpcAddr1, err := server1.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client1, err := grpc.NewClient(grpcAddr1)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr2, err := server2.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(grpcAddr2)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr3, err := server3.nodeConfig.GetGrpcAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(grpcAddr3)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// set test data before delete

	err = client1.SetState("test/key1", "val1")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val11, err := client1.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if expVal11 != actVal11 {
		t.Errorf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	val21, err := client2.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if expVal21 != actVal21 {
		t.Errorf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	val31, err := client3.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if expVal31 != actVal31 {
		t.Errorf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	err = client2.SetState("test/key2", "val2")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val12, err := client1.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if expVal12 != actVal12 {
		t.Errorf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	val22, err := client2.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if expVal22 != actVal22 {
		t.Errorf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	val32, err := client3.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if expVal32 != actVal32 {
		t.Errorf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	err = client3.SetState("test/key3", "val3")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val13, err := client1.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if expVal13 != actVal13 {
		t.Errorf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	val23, err := client2.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if expVal23 != actVal23 {
		t.Errorf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	val33, err := client3.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if expVal33 != actVal33 {
		t.Errorf("expected content to see %v, saw %v", expVal33, actVal33)
	}

	// delete

	err = client1.DeleteState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val11, err = client1.GetState("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val11 != nil {
		t.Errorf("%v", err)
	}
	val21, err = client2.GetState("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val21 != nil {
		t.Errorf("%v", err)
	}
	val31, err = client3.GetState("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val31 != nil {
		t.Errorf("%v", err)
	}

	err = client2.DeleteState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val12, err = client1.GetState("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val12 != nil {
		t.Errorf("%v", err)
	}
	val22, err = client2.GetState("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val22 != nil {
		t.Errorf("%v", err)
	}
	val32, err = client3.GetState("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val32 != nil {
		t.Errorf("%v", err)
	}

	err = client3.DeleteState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val13, err = client1.GetState("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val13 != nil {
		t.Errorf("%v", err)
	}
	val23, err = client2.GetState("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val23 != nil {
		t.Errorf("%v", err)
	}
	val33, err = client3.GetState("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val33 != nil {
		t.Errorf("%v", err)
	}

	// delete non-existing data from manager1
	err = client1.DeleteState("test/non-existing")
	if err == nil {
		t.Errorf("%v", err)
	}

	// delete non-existing data from manager2
	err = client2.DeleteState("test/non-existing")
	if err == nil {
		t.Errorf("%v", err)
	}

	// delete non-existing data from manager3
	err = client3.DeleteState("test/non-existing")
	if err == nil {
		t.Errorf("%v", err)
	}
}
