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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/mosuka/blast/strutils"

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/testutils"
)

func TestServer_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
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

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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

	// healthiness
	healthiness, err := client.NodeHealthCheck(index.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness := index.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness := healthiness
	if expHealthiness != actHealthiness {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness, actHealthiness)
	}

	// liveness
	liveness, err := client.NodeHealthCheck(index.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness := index.NodeHealthCheckResponse_ALIVE.String()
	actLiveness := liveness
	if expLiveness != actLiveness {
		t.Fatalf("expected content to see %v, saw %v", expLiveness, actLiveness)
	}

	// readiness
	readiness, err := client.NodeHealthCheck(index.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness := index.NodeHealthCheckResponse_READY.String()
	actReadiness := readiness
	if expReadiness != actReadiness {
		t.Fatalf("expected content to see %v, saw %v", expReadiness, actReadiness)
	}
}

func TestServer_GetNode(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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
	nodeInfo, err := client.NodeInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNodeInfo := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_LEADER,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}
	actNodeInfo := nodeInfo
	if !reflect.DeepEqual(expNodeInfo, actNodeInfo) {
		t.Fatalf("expected content to see %v, saw %v", expNodeInfo, actNodeInfo)
	}
}

func TestServer_GetCluster(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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
	cluster, err := client.ClusterInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCluster := &index.Cluster{
		Nodes: map[string]*index.Node{
			nodeId: {
				Id:          nodeId,
				BindAddress: bindAddress,
				State:       index.Node_LEADER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress,
					HttpAddress: httpAddress,
				},
			},
		},
	}
	actCluster := cluster
	if !reflect.DeepEqual(expCluster, actCluster) {
		t.Fatalf("expected content to see %v, saw %v", expCluster, actCluster)
	}
}

func TestServer_GetIndexMapping(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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

	expIndexMapping := indexConfig.IndexMapping
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
		t.Fatalf("expected content to see %v, saw %v", expIndexMapping, actIndexMapping)
	}
}

func TestServer_GetIndexType(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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

	expIndexType := indexConfig.IndexType
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexConfigMap, err := client.GetIndexConfig()
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexType := actIndexConfigMap["index_type"].(string)

	if !reflect.DeepEqual(expIndexType, actIndexType) {
		t.Fatalf("expected content to see %v, saw %v", expIndexType, actIndexType)
	}
}

func TestServer_GetIndexStorageType(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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

	expIndexStorageType := indexConfig.IndexStorageType
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexConfigMap, err := client.GetIndexConfig()
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexStorageType := actIndexConfigMap["index_storage_type"].(string)

	if !reflect.DeepEqual(expIndexStorageType, actIndexStorageType) {
		t.Fatalf("expected content to see %v, saw %v", expIndexStorageType, actIndexStorageType)
	}
}

func TestServer_GetIndexStats(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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
		t.Fatalf("expected content to see %v, saw %v", expIndexStats, actIndexStats)
	}
}

func TestServer_PutDocument(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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
	docs := make([]*indexutils.Document, 0)
	docPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	docFile1, err := os.Open(docPath1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		_ = docFile1.Close()
	}()
	docBytes1, err := ioutil.ReadAll(docFile1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var docFields1 map[string]interface{}
	err = json.Unmarshal(docBytes1, &docFields1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	doc1, err := indexutils.NewDocument("doc1", docFields1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	docs = append(docs, doc1)
	count, err := client.IndexDocument(docs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expCount := 1
	actCount := count

	if expCount != actCount {
		t.Fatalf("expected content to see %v, saw %v", expCount, actCount)
	}
}

func TestServer_GetDocument(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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
	putDocs := make([]*indexutils.Document, 0)
	putDocPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	putDocFile1, err := os.Open(putDocPath1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		_ = putDocFile1.Close()
	}()
	putDocBytes1, err := ioutil.ReadAll(putDocFile1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	putDoc1, err := indexutils.NewDocumentFromBytes(putDocBytes1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	putDocs = append(putDocs, putDoc1)
	putCount, err := client.IndexDocument(putDocs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expPutCount := 1
	actPutCount := putCount

	if expPutCount != actPutCount {
		t.Fatalf("expected content to see %v, saw %v", expPutCount, actPutCount)
	}

	// get document
	getDocFields1, err := client.GetDocument("enwiki_1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expGetDocFields1 := putDoc1.Fields
	actGetDocFields1 := getDocFields1
	if !reflect.DeepEqual(expGetDocFields1, actGetDocFields1) {
		t.Fatalf("expected content to see %v, saw %v", expGetDocFields1, actGetDocFields1)
	}

	// get non-existing document
	getDocFields2, err := client.GetDocument("doc2")
	if err != errors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if getDocFields2 != nil {
		t.Fatalf("expected content to see nil, saw %v", getDocFields2)
	}
}

func TestServer_DeleteDocument(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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
	putDocs := make([]*indexutils.Document, 0)
	putDocPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	putDocFile1, err := os.Open(putDocPath1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		_ = putDocFile1.Close()
	}()
	putDocBytes1, err := ioutil.ReadAll(putDocFile1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	putDoc1, err := indexutils.NewDocumentFromBytes(putDocBytes1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	putDocs = append(putDocs, putDoc1)
	putCount, err := client.IndexDocument(putDocs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expPutCount := 1
	actPutCount := putCount

	if expPutCount != actPutCount {
		t.Fatalf("expected content to see %v, saw %v", expPutCount, actPutCount)
	}

	// get document
	getDocFields1, err := client.GetDocument("enwiki_1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expGetDocFields1 := putDoc1.Fields
	actGetDocFields1 := getDocFields1
	if !reflect.DeepEqual(expGetDocFields1, actGetDocFields1) {
		t.Fatalf("expected content to see %v, saw %v", expGetDocFields1, actGetDocFields1)
	}

	// get non-existing document
	getDocFields2, err := client.GetDocument("non-existing")
	if err != errors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if getDocFields2 != nil {
		t.Fatalf("expected content to see nil, saw %v", getDocFields2)
	}

	// delete document
	delCount, err := client.DeleteDocument([]string{"enwiki_1"})
	if err != nil {
		t.Fatalf("%v", err)
	}
	expDelCount := 1
	actDelCount := delCount
	if expDelCount != actDelCount {
		t.Fatalf("expected content to see %v, saw %v", expDelCount, actDelCount)
	}

	// get document
	getDocFields1, err = client.GetDocument("enwiki_1")
	if err != errors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if getDocFields1 != nil {
		t.Fatalf("expected content to see nil, saw %v", getDocFields1)
	}

	// delete non-existing document
	getDocFields1, err = client.GetDocument("non-existing")
	if err != errors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if getDocFields1 != nil {
		t.Fatalf("expected content to see nil, saw %v", getDocFields1)
	}
}

func TestServer_Search(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir)
	}()
	raftStorageType := "boltdb"

	node := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexConfig, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := NewGRPCClient(node.Metadata.GrpcAddress)
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
	putDocs := make([]*indexutils.Document, 0)
	putDocPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	putDocFile1, err := os.Open(putDocPath1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		_ = putDocFile1.Close()
	}()
	putDocBytes1, err := ioutil.ReadAll(putDocFile1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	var putDocFields1 map[string]interface{}
	err = json.Unmarshal(putDocBytes1, &putDocFields1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	putDoc1, err := indexutils.NewDocument("doc1", putDocFields1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	putDocs = append(putDocs, putDoc1)
	putCount, err := client.IndexDocument(putDocs)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expPutCount := 1
	actPutCount := putCount

	if expPutCount != actPutCount {
		t.Fatalf("expected content to see %v, saw %v", expPutCount, actPutCount)
	}

	// search
	searchRequestPath := filepath.Join(curDir, "../example/wiki_search_request.json")

	searchRequestFile, err := os.Open(searchRequestPath)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		_ = searchRequestFile.Close()
	}()

	searchRequestByte, err := ioutil.ReadAll(searchRequestFile)
	if err != nil {
		t.Fatalf("%v", err)
	}

	searchRequest := bleve.NewSearchRequest(nil)
	err = json.Unmarshal(searchRequestByte, searchRequest)
	if err != nil {
		t.Fatalf("%v", err)
	}

	searchResult1, err := client.Search(searchRequest)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expTotal := uint64(1)
	actTotal := searchResult1.Total
	if expTotal != actTotal {
		t.Fatalf("expected content to see %v, saw %v", expTotal, actTotal)
	}
}

func TestCluster_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress1 := ""
	shardId1 := ""
	peerGrpcAddress1 := ""
	grpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId1 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir1 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir1)
	}()
	raftStorageType1 := "boltdb"

	node1 := &index.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexConfig1, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexConfig1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server1.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

	managerGrpcAddress2 := ""
	shardId2 := ""
	peerGrpcAddress2 := grpcAddress1
	grpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId2 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir2 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir2)
	}()
	raftStorageType2 := "boltdb"

	node2 := &index.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexConfig2, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexConfig2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server2.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

	managerGrpcAddress3 := ""
	shardId3 := ""
	peerGrpcAddress3 := grpcAddress1
	grpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId3 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir3 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir3)
	}()
	raftStorageType3 := "boltdb"

	node3 := &index.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexConfig3, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexConfig3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server3.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)
}

func TestCluster_LivenessProbe(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress1 := ""
	shardId1 := ""
	peerGrpcAddress1 := ""
	grpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId1 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir1 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir1)
	}()
	raftStorageType1 := "boltdb"

	node1 := &index.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexConfig1, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexConfig1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server1.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

	managerGrpcAddress2 := ""
	shardId2 := ""
	peerGrpcAddress2 := grpcAddress1
	grpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId2 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir2 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir2)
	}()
	raftStorageType2 := "boltdb"

	node2 := &index.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexConfig2, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexConfig2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server2.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

	managerGrpcAddress3 := ""
	shardId3 := ""
	peerGrpcAddress3 := grpcAddress1
	grpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId3 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir3 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir3)
	}()
	raftStorageType3 := "boltdb"

	node3 := &index.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexConfig3, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexConfig3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server3.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for all servers
	client1, err := NewGRPCClient(node1.Metadata.GrpcAddress)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := NewGRPCClient(node2.Metadata.GrpcAddress)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := NewGRPCClient(node3.Metadata.GrpcAddress)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// healthiness
	healthiness1, err := client1.NodeHealthCheck(index.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness1 := index.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness1 := healthiness1
	if expHealthiness1 != actHealthiness1 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness1, actHealthiness1)
	}

	// liveness
	liveness1, err := client1.NodeHealthCheck(index.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness1 := index.NodeHealthCheckResponse_ALIVE.String()
	actLiveness1 := liveness1
	if expLiveness1 != actLiveness1 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness1, actLiveness1)
	}

	// readiness
	readiness1, err := client1.NodeHealthCheck(index.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness1 := index.NodeHealthCheckResponse_READY.String()
	actReadiness1 := readiness1
	if expReadiness1 != actReadiness1 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness1, actReadiness1)
	}

	// healthiness
	healthiness2, err := client2.NodeHealthCheck(index.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness2 := index.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness2 := healthiness2
	if expHealthiness2 != actHealthiness2 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness2, actHealthiness2)
	}

	// liveness
	liveness2, err := client2.NodeHealthCheck(index.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness2 := index.NodeHealthCheckResponse_ALIVE.String()
	actLiveness2 := liveness2
	if expLiveness2 != actLiveness2 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness2, actLiveness2)
	}

	// readiness
	readiness2, err := client2.NodeHealthCheck(index.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness2 := index.NodeHealthCheckResponse_READY.String()
	actReadiness2 := readiness2
	if expReadiness2 != actReadiness2 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness2, actReadiness2)
	}

	// healthiness
	healthiness3, err := client3.NodeHealthCheck(index.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness3 := index.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness3 := healthiness3
	if expHealthiness3 != actHealthiness3 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness3, actHealthiness3)
	}

	// liveness
	liveness3, err := client3.NodeHealthCheck(index.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness3 := index.NodeHealthCheckResponse_ALIVE.String()
	actLiveness3 := liveness3
	if expLiveness3 != actLiveness3 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness3, actLiveness3)
	}

	// readiness
	readiness3, err := client3.NodeHealthCheck(index.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness3 := index.NodeHealthCheckResponse_READY.String()
	actReadiness3 := readiness3
	if expReadiness3 != actReadiness3 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness3, actReadiness3)
	}
}

func TestCluster_GetNode(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress1 := ""
	shardId1 := ""
	peerGrpcAddress1 := ""
	grpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId1 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir1 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir1)
	}()
	raftStorageType1 := "boltdb"

	node1 := &index.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexConfig1, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexConfig1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server1.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

	managerGrpcAddress2 := ""
	shardId2 := ""
	peerGrpcAddress2 := grpcAddress1
	grpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId2 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir2 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir2)
	}()
	raftStorageType2 := "boltdb"

	node2 := &index.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexConfig2, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexConfig2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server2.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

	managerGrpcAddress3 := ""
	shardId3 := ""
	peerGrpcAddress3 := grpcAddress1
	grpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId3 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir3 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir3)
	}()
	raftStorageType3 := "boltdb"

	node3 := &index.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexConfig3, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexConfig3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server3.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for all servers
	client1, err := NewGRPCClient(node1.Metadata.GrpcAddress)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := NewGRPCClient(node2.Metadata.GrpcAddress)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := NewGRPCClient(node3.Metadata.GrpcAddress)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get all node info from all nodes
	node11, err := client1.NodeInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode11 := &index.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       index.Node_LEADER,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}
	actNode11 := node11
	if !reflect.DeepEqual(expNode11, actNode11) {
		t.Fatalf("expected content to see %v, saw %v", expNode11, actNode11)
	}

	node21, err := client2.NodeInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode21 := &index.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       index.Node_FOLLOWER,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}
	actNode21 := node21
	if !reflect.DeepEqual(expNode21, actNode21) {
		t.Fatalf("expected content to see %v, saw %v", expNode21, actNode21)
	}

	node31, err := client3.NodeInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode31 := &index.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       index.Node_FOLLOWER,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}
	actNode31 := node31
	if !reflect.DeepEqual(expNode31, actNode31) {
		t.Fatalf("expected content to see %v, saw %v", expNode31, actNode31)
	}
}

func TestCluster_GetCluster(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress1 := ""
	shardId1 := ""
	peerGrpcAddress1 := ""
	grpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId1 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir1 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir1)
	}()
	raftStorageType1 := "boltdb"

	node1 := &index.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexConfig1, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexConfig1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server1.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

	managerGrpcAddress2 := ""
	shardId2 := ""
	peerGrpcAddress2 := grpcAddress1
	grpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId2 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir2 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir2)
	}()
	raftStorageType2 := "boltdb"

	node2 := &index.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexConfig2, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexConfig2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server2.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

	managerGrpcAddress3 := ""
	shardId3 := ""
	peerGrpcAddress3 := grpcAddress1
	grpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId3 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	bindAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	dataDir3 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(dataDir3)
	}()
	raftStorageType3 := "boltdb"

	node3 := &index.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexConfig3, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexConfig3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		server3.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for all servers
	client1, err := NewGRPCClient(node1.Metadata.GrpcAddress)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := NewGRPCClient(node2.Metadata.GrpcAddress)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := NewGRPCClient(node3.Metadata.GrpcAddress)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get cluster info from manager1
	cluster1, err := client1.ClusterInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCluster1 := &index.Cluster{
		Nodes: map[string]*index.Node{
			nodeId1: {
				Id:          nodeId1,
				BindAddress: bindAddress1,
				State:       index.Node_LEADER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress1,
					HttpAddress: httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress2,
					HttpAddress: httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress3,
					HttpAddress: httpAddress3,
				},
			},
		},
	}
	actCluster1 := cluster1
	if !reflect.DeepEqual(expCluster1, actCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expCluster1, actCluster1)
	}

	cluster2, err := client2.ClusterInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCluster2 := &index.Cluster{
		Nodes: map[string]*index.Node{
			nodeId1: {
				Id:          nodeId1,
				BindAddress: bindAddress1,
				State:       index.Node_LEADER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress1,
					HttpAddress: httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress2,
					HttpAddress: httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress3,
					HttpAddress: httpAddress3,
				},
			},
		},
	}
	actCluster2 := cluster2
	if !reflect.DeepEqual(expCluster2, actCluster2) {
		t.Fatalf("expected content to see %v, saw %v", expCluster2, actCluster2)
	}

	cluster3, err := client3.ClusterInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCluster3 := &index.Cluster{
		Nodes: map[string]*index.Node{
			nodeId1: {
				Id:          nodeId1,
				BindAddress: bindAddress1,
				State:       index.Node_LEADER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress1,
					HttpAddress: httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress2,
					HttpAddress: httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: grpcAddress3,
					HttpAddress: httpAddress3,
				},
			},
		},
	}
	actCluster3 := cluster3
	if !reflect.DeepEqual(expCluster3, actCluster3) {
		t.Fatalf("expected content to see %v, saw %v", expCluster3, actCluster3)
	}
}
