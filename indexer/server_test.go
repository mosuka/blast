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

	"github.com/hashicorp/raft"

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/testutils"
)

func TestServer_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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
	node, err := client.GetNode(nodeConfig.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode := map[string]interface{}{
		"node_config": nodeConfig.ToMap(),
		"state":       "Leader",
	}
	actNode := node
	if !reflect.DeepEqual(expNode, actNode) {
		t.Fatalf("expected content to see %v, saw %v", expNode, actNode)
	}
}

func TestServer_GetCluster(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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
		t.Fatalf("%v", err)
	}
	expCluster := map[string]interface{}{
		nodeConfig.NodeId: map[string]interface{}{
			"node_config": nodeConfig.ToMap(),
			"state":       "Leader",
		},
	}
	actCluster := cluster
	if !reflect.DeepEqual(expCluster, actCluster) {
		t.Fatalf("expected content to see %v, saw %v", expCluster, actCluster)
	}
}

func TestServer_GetIndexMapping(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create gRPC logger
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()

	// create node config
	nodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig.DataDir)
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
		t.Fatalf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(nodeConfig.GRPCAddr)
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

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

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
		_ = os.RemoveAll(nodeConfig1.DataDir)
	}()
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named("server1"), grpcLogger, httpAccessLogger)
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

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	clusterConfig2.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig2.DataDir)
	}()
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named("server2"), grpcLogger, httpAccessLogger)
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
	clusterConfig3.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig3.DataDir)
	}()
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named("server3"), grpcLogger, httpAccessLogger)
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
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

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
		_ = os.RemoveAll(nodeConfig1.DataDir)
	}()
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named("server1"), grpcLogger, httpAccessLogger)
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

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	clusterConfig2.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig2.DataDir)
	}()
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named("server2"), grpcLogger, httpAccessLogger)
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
	clusterConfig3.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig3.DataDir)
	}()
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named("server3"), grpcLogger, httpAccessLogger)
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

	// gRPC client for all servers
	client1, err := grpc.NewClient(nodeConfig1.GRPCAddr)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(nodeConfig2.GRPCAddr)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(nodeConfig3.GRPCAddr)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// liveness check for server1
	liveness1, err := client1.LivenessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness1 := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness1 := liveness1
	if expLiveness1 != actLiveness1 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness1, actLiveness1)
	}

	// liveness check for server2
	liveness2, err := client2.LivenessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness2 := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness2 := liveness2
	if expLiveness2 != actLiveness2 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness2, actLiveness2)
	}

	// liveness check for server3
	liveness3, err := client3.LivenessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness3 := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness3 := liveness3
	if expLiveness3 != actLiveness3 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness3, actLiveness3)
	}
}

func TestCluster_ReadinessProbe(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

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
		_ = os.RemoveAll(nodeConfig1.DataDir)
	}()
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named("server1"), grpcLogger, httpAccessLogger)
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

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	clusterConfig2.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig2.DataDir)
	}()
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named("server2"), grpcLogger, httpAccessLogger)
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
	clusterConfig3.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig3.DataDir)
	}()
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named("server3"), grpcLogger, httpAccessLogger)
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

	// gRPC client for all servers
	client1, err := grpc.NewClient(nodeConfig1.GRPCAddr)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(nodeConfig2.GRPCAddr)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(nodeConfig3.GRPCAddr)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// readiness check for server1
	readiness1, err := client1.ReadinessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness1 := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness1 := readiness1
	if expReadiness1 != actReadiness1 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness1, actReadiness1)
	}

	// readiness check for server2
	readiness2, err := client2.ReadinessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness2 := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness2 := readiness2
	if expReadiness2 != actReadiness2 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness2, actReadiness2)
	}

	// readiness check for server3
	readiness3, err := client3.ReadinessProbe()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness3 := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness3 := readiness3
	if expReadiness3 != actReadiness3 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness3, actReadiness3)
	}
}

func TestCluster_GetNode(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

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
		_ = os.RemoveAll(nodeConfig1.DataDir)
	}()
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named("server1"), grpcLogger, httpAccessLogger)
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

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	clusterConfig2.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig2.DataDir)
	}()
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named("server2"), grpcLogger, httpAccessLogger)
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
	clusterConfig3.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig3.DataDir)
	}()
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named("server3"), grpcLogger, httpAccessLogger)
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

	// gRPC client for all servers
	client1, err := grpc.NewClient(nodeConfig1.GRPCAddr)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(nodeConfig2.GRPCAddr)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(nodeConfig3.GRPCAddr)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get all node info from all nodes
	node11, err := client1.GetNode(nodeConfig1.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode11 := map[string]interface{}{
		"node_config": server1.nodeConfig.ToMap(),
		"state":       raft.Leader.String(),
	}
	actNode11 := node11
	if !reflect.DeepEqual(expNode11, actNode11) {
		t.Fatalf("expected content to see %v, saw %v", expNode11, actNode11)
	}

	node12, err := client1.GetNode(nodeConfig2.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode12 := map[string]interface{}{
		"node_config": server2.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode12 := node12
	if !reflect.DeepEqual(expNode12, actNode12) {
		t.Fatalf("expected content to see %v, saw %v", expNode12, actNode12)
	}

	node13, err := client1.GetNode(nodeConfig3.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode13 := map[string]interface{}{
		"node_config": server3.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode13 := node13
	if !reflect.DeepEqual(expNode13, actNode13) {
		t.Fatalf("expected content to see %v, saw %v", expNode13, actNode13)
	}

	node21, err := client2.GetNode(nodeConfig1.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode21 := map[string]interface{}{
		"node_config": server1.nodeConfig.ToMap(),
		"state":       raft.Leader.String(),
	}
	actNode21 := node21
	if !reflect.DeepEqual(expNode21, actNode21) {
		t.Fatalf("expected content to see %v, saw %v", expNode21, actNode21)
	}

	node22, err := client2.GetNode(nodeConfig2.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode22 := map[string]interface{}{
		"node_config": server2.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode22 := node22
	if !reflect.DeepEqual(expNode22, actNode22) {
		t.Fatalf("expected content to see %v, saw %v", expNode22, actNode22)
	}

	node23, err := client2.GetNode(nodeConfig3.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode23 := map[string]interface{}{
		"node_config": server3.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode23 := node23
	if !reflect.DeepEqual(expNode23, actNode23) {
		t.Fatalf("expected content to see %v, saw %v", expNode23, actNode23)
	}

	node31, err := client3.GetNode(nodeConfig1.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode31 := map[string]interface{}{
		"node_config": server1.nodeConfig.ToMap(),
		"state":       raft.Leader.String(),
	}
	actNode31 := node31
	if !reflect.DeepEqual(expNode31, actNode31) {
		t.Fatalf("expected content to see %v, saw %v", expNode31, actNode31)
	}

	node32, err := client3.GetNode(nodeConfig2.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode32 := map[string]interface{}{
		"node_config": server2.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode32 := node32
	if !reflect.DeepEqual(expNode32, actNode32) {
		t.Fatalf("expected content to see %v, saw %v", expNode32, actNode32)
	}

	node33, err := client3.GetNode(nodeConfig3.NodeId)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode33 := map[string]interface{}{
		"node_config": server3.nodeConfig.ToMap(),
		"state":       raft.Follower.String(),
	}
	actNode33 := node33
	if !reflect.DeepEqual(expNode33, actNode33) {
		t.Fatalf("expected content to see %v, saw %v", expNode33, actNode33)
	}
}

func TestCluster_GetCluster(t *testing.T) {
	curDir, _ := os.Getwd()

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

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
		_ = os.RemoveAll(nodeConfig1.DataDir)
	}()
	// create server1
	server1, err := NewServer(clusterConfig1, nodeConfig1, indexConfig, logger.Named("server1"), grpcLogger, httpAccessLogger)
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

	// create configs for server2
	clusterConfig2 := config.DefaultClusterConfig()
	clusterConfig2.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig2.DataDir)
	}()
	// create server2
	server2, err := NewServer(clusterConfig2, nodeConfig2, config.DefaultIndexConfig(), logger.Named("server2"), grpcLogger, httpAccessLogger)
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
	clusterConfig3.PeerAddr = nodeConfig1.GRPCAddr
	nodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(nodeConfig3.DataDir)
	}()
	// create server3
	server3, err := NewServer(clusterConfig3, nodeConfig3, config.DefaultIndexConfig(), logger.Named("server3"), grpcLogger, httpAccessLogger)
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
	client1, err := grpc.NewClient(nodeConfig1.GRPCAddr)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client2, err := grpc.NewClient(nodeConfig2.GRPCAddr)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	client3, err := grpc.NewClient(nodeConfig3.GRPCAddr)
	defer func() {
		_ = client3.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get cluster info from all servers
	cluster1, err := client1.GetCluster()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCluster1 := map[string]interface{}{
		nodeConfig1.NodeId: map[string]interface{}{
			"node_config": nodeConfig1.ToMap(),
			"state":       raft.Leader.String(),
		},
		nodeConfig2.NodeId: map[string]interface{}{
			"node_config": nodeConfig2.ToMap(),
			"state":       raft.Follower.String(),
		},
		nodeConfig3.NodeId: map[string]interface{}{
			"node_config": nodeConfig3.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actCluster1 := cluster1
	if !reflect.DeepEqual(expCluster1, actCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expCluster1, actCluster1)
	}

	cluster2, err := client2.GetCluster()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCluster2 := map[string]interface{}{
		nodeConfig1.NodeId: map[string]interface{}{
			"node_config": nodeConfig1.ToMap(),
			"state":       raft.Leader.String(),
		},
		nodeConfig2.NodeId: map[string]interface{}{
			"node_config": nodeConfig2.ToMap(),
			"state":       raft.Follower.String(),
		},
		nodeConfig3.NodeId: map[string]interface{}{
			"node_config": nodeConfig3.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actCluster2 := cluster2
	if !reflect.DeepEqual(expCluster2, actCluster2) {
		t.Fatalf("expected content to see %v, saw %v", expCluster2, actCluster2)
	}

	cluster3, err := client3.GetCluster()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expCluster3 := map[string]interface{}{
		nodeConfig1.NodeId: map[string]interface{}{
			"node_config": nodeConfig1.ToMap(),
			"state":       raft.Leader.String(),
		},
		nodeConfig2.NodeId: map[string]interface{}{
			"node_config": nodeConfig2.ToMap(),
			"state":       raft.Follower.String(),
		},
		nodeConfig3.NodeId: map[string]interface{}{
			"node_config": nodeConfig3.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actCluster3 := cluster3
	if !reflect.DeepEqual(expCluster3, actCluster3) {
		t.Fatalf("expected content to see %v, saw %v", expCluster3, actCluster3)
	}
}
