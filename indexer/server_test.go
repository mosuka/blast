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

package indexer

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/blevesearch/bleve"

	"github.com/mosuka/blast/errors"

	"github.com/mosuka/blast/indexutils"

	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/protobuf"

	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/logutils"
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	actIndexConfigMap, err := client.GetIndexConfig()
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexMapping, err := indexutils.NewIndexMappingFromMap(actIndexConfigMap["index_mapping"].(map[string]interface{}))
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	actIndexConfigMap, err := client.GetIndexConfig()
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexType := actIndexConfigMap["index_type"].(string)

	if !reflect.DeepEqual(expIndexType, actIndexType) {
		t.Errorf("expected content to see %v, saw %v", expIndexType, actIndexType)
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	actIndexConfigMap, err := client.GetIndexConfig()
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexStorageType := actIndexConfigMap["index_storage_type"].(string)

	if !reflect.DeepEqual(expIndexStorageType, actIndexStorageType) {
		t.Errorf("expected content to see %v, saw %v", expIndexStorageType, actIndexStorageType)
	}
}

func TestServer_GetIndexStats(t *testing.T) {
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	expIndexStats := map[string]interface{}{
		"index": map[string]interface{}{
			"analysis_time":                float64(0),
			"batches":                      float64(0),
			"deletes":                      float64(0),
			"errors":                       float64(0),
			"index_time":                   float64(0),
			"num_plain_text_bytes_indexed": float64(0),
			"term_searchers_finished":      float64(0),
			"term_searchers_started":       float64(0),
			"updates":                      float64(0),
		},
		"search_time": float64(0),
		"searches":    float64(0),
	}

	actIndexStats, err := client.GetIndexStats()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(expIndexStats, actIndexStats) {
		t.Errorf("expected content to see %v, saw %v", expIndexStats, actIndexStats)
	}
}

func TestServer_PutDocument(t *testing.T) {
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	// put document
	docs := make([]map[string]interface{}, 0)
	docPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	docFile1, err := os.Open(docPath1)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		_ = docFile1.Close()
	}()
	docBytes1, err := ioutil.ReadAll(docFile1)
	if err != nil {
		t.Errorf("%v", err)
	}
	var docFields1 map[string]interface{}
	err = json.Unmarshal(docBytes1, &docFields1)
	if err != nil {
		t.Errorf("%v", err)
	}
	doc1 := map[string]interface{}{
		"id":     "doc1",
		"fields": docFields1,
	}
	docs = append(docs, doc1)
	count, err := client.IndexDocument(docs)
	if err != nil {
		t.Errorf("%v", err)
	}

	expCount := 1
	actCount := count

	if expCount != actCount {
		t.Errorf("expected content to see %v, saw %v", expCount, actCount)
	}
}

func TestServer_GetDocument(t *testing.T) {
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	// put document
	putDocs := make([]map[string]interface{}, 0)
	putDocPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	putDocFile1, err := os.Open(putDocPath1)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		_ = putDocFile1.Close()
	}()
	putDocBytes1, err := ioutil.ReadAll(putDocFile1)
	if err != nil {
		t.Errorf("%v", err)
	}
	var putDocFields1 map[string]interface{}
	err = json.Unmarshal(putDocBytes1, &putDocFields1)
	if err != nil {
		t.Errorf("%v", err)
	}
	putDoc1 := map[string]interface{}{
		"id":     "doc1",
		"fields": putDocFields1,
	}
	putDocs = append(putDocs, putDoc1)
	putCount, err := client.IndexDocument(putDocs)
	if err != nil {
		t.Errorf("%v", err)
	}

	expPutCount := 1
	actPutCount := putCount

	if expPutCount != actPutCount {
		t.Errorf("expected content to see %v, saw %v", expPutCount, actPutCount)
	}

	// get document
	getDocFields1, err := client.GetDocument("doc1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expGetDocFields1 := putDocFields1
	actGetDocFields1 := getDocFields1
	if !reflect.DeepEqual(expGetDocFields1, actGetDocFields1) {
		t.Errorf("expected content to see %v, saw %v", expGetDocFields1, actGetDocFields1)
	}

	// get non-existing document
	getDocFields2, err := client.GetDocument("doc2")
	if err != errors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if getDocFields2 != nil {
		t.Errorf("expected content to see nil, saw %v", getDocFields2)
	}
}

func TestServer_DeleteDocument(t *testing.T) {
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	// put document
	putDocs := make([]map[string]interface{}, 0)
	putDocPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	putDocFile1, err := os.Open(putDocPath1)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		_ = putDocFile1.Close()
	}()
	putDocBytes1, err := ioutil.ReadAll(putDocFile1)
	if err != nil {
		t.Errorf("%v", err)
	}
	var putDocFields1 map[string]interface{}
	err = json.Unmarshal(putDocBytes1, &putDocFields1)
	if err != nil {
		t.Errorf("%v", err)
	}
	putDoc1 := map[string]interface{}{
		"id":     "doc1",
		"fields": putDocFields1,
	}
	putDocs = append(putDocs, putDoc1)
	putCount, err := client.IndexDocument(putDocs)
	if err != nil {
		t.Errorf("%v", err)
	}

	expPutCount := 1
	actPutCount := putCount

	if expPutCount != actPutCount {
		t.Errorf("expected content to see %v, saw %v", expPutCount, actPutCount)
	}

	// get document
	getDocFields1, err := client.GetDocument("doc1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expGetDocFields1 := putDocFields1
	actGetDocFields1 := getDocFields1
	if !reflect.DeepEqual(expGetDocFields1, actGetDocFields1) {
		t.Errorf("expected content to see %v, saw %v", expGetDocFields1, actGetDocFields1)
	}

	// get non-existing document
	getDocFields2, err := client.GetDocument("non-existing")
	if err != errors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if getDocFields2 != nil {
		t.Errorf("expected content to see nil, saw %v", getDocFields2)
	}

	// delete document
	delCount, err := client.DeleteDocument([]string{"doc1"})
	if err != nil {
		t.Errorf("%v", err)
	}
	expDelCount := 1
	actDelCount := delCount
	if expDelCount != actDelCount {
		t.Errorf("expected content to see %v, saw %v", expDelCount, actDelCount)
	}

	// get document
	getDocFields1, err = client.GetDocument("doc1")
	if err != errors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if getDocFields1 != nil {
		t.Errorf("expected content to see nil, saw %v", getDocFields1)
	}

	// delete non-existing document
	getDocFields1, err = client.GetDocument("non-existing")
	if err != errors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if getDocFields1 != nil {
		t.Errorf("expected content to see nil, saw %v", getDocFields1)
	}
}

func TestServer_Search(t *testing.T) {
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

	server, err := NewServer(clusterConfig, nodeConfig, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
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

	// put document
	putDocs := make([]map[string]interface{}, 0)
	putDocPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	putDocFile1, err := os.Open(putDocPath1)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		_ = putDocFile1.Close()
	}()
	putDocBytes1, err := ioutil.ReadAll(putDocFile1)
	if err != nil {
		t.Errorf("%v", err)
	}
	var putDocFields1 map[string]interface{}
	err = json.Unmarshal(putDocBytes1, &putDocFields1)
	if err != nil {
		t.Errorf("%v", err)
	}
	putDoc1 := map[string]interface{}{
		"id":     "doc1",
		"fields": putDocFields1,
	}
	putDocs = append(putDocs, putDoc1)
	putCount, err := client.IndexDocument(putDocs)
	if err != nil {
		t.Errorf("%v", err)
	}

	expPutCount := 1
	actPutCount := putCount

	if expPutCount != actPutCount {
		t.Errorf("expected content to see %v, saw %v", expPutCount, actPutCount)
	}

	// search
	searchRequestPath := filepath.Join(curDir, "../example/wiki_search_request.json")

	searchRequestFile, err := os.Open(searchRequestPath)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		_ = searchRequestFile.Close()
	}()

	searchRequestByte, err := ioutil.ReadAll(searchRequestFile)
	if err != nil {
		t.Errorf("%v", err)
	}

	searchRequest := bleve.NewSearchRequest(nil)
	err = json.Unmarshal(searchRequestByte, searchRequest)
	if err != nil {
		t.Errorf("%v", err)
	}

	searchResult1, err := client.Search(searchRequest)
	if err != nil {
		t.Errorf("%v", err)
	}
	expTotal := uint64(1)
	actTotal := searchResult1.Total
	if expTotal != actTotal {
		t.Errorf("expected content to see %v, saw %v", expTotal, actTotal)
	}
}
