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

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/strutils"
	"github.com/mosuka/blast/testutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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
	reqHealthiness := &index.NodeHealthCheckRequest{Probe: index.NodeHealthCheckRequest_HEALTHINESS}
	resHealthiness, err := client.NodeHealthCheck(reqHealthiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthinessState := index.NodeHealthCheckResponse_HEALTHY
	actHealthinessState := resHealthiness.State
	if expHealthinessState != actHealthinessState {
		t.Fatalf("expected content to see %v, saw %v", expHealthinessState, actHealthinessState)
	}

	// liveness
	reqLiveness := &index.NodeHealthCheckRequest{Probe: index.NodeHealthCheckRequest_LIVENESS}
	resLiveness, err := client.NodeHealthCheck(reqLiveness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLivenessState := index.NodeHealthCheckResponse_ALIVE
	actLivenessState := resLiveness.State
	if expLivenessState != actLivenessState {
		t.Fatalf("expected content to see %v, saw %v", expLivenessState, actLivenessState)
	}

	// readiness
	reqReadiness := &index.NodeHealthCheckRequest{Probe: index.NodeHealthCheckRequest_READINESS}
	resReadiness, err := client.NodeHealthCheck(reqReadiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadinessState := index.NodeHealthCheckResponse_READY
	actReadinessState := resReadiness.State
	if expReadinessState != actReadinessState {
		t.Fatalf("expected content to see %v, saw %v", expReadinessState, actReadinessState)
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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
	req := &empty.Empty{}
	res, err := client.NodeInfo(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNodeInfo := &index.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       index.Node_LEADER,
		Metadata: &index.Metadata{
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}
	actNodeInfo := res.Node
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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
	req := &empty.Empty{}
	res, err := client.ClusterInfo(req)
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
					GrpcAddress:        grpcAddress,
					GrpcGatewayAddress: grpcGatewayAddress,
					HttpAddress:        httpAddress,
				},
			},
		},
	}
	actCluster := res.Cluster
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	expIndexMapping := indexMapping

	req := &empty.Empty{}
	res, err := client.GetIndexConfig(req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	im, err := protobuf.MarshalAny(res.IndexConfig.IndexMapping)
	if err != nil {
		t.Fatalf("%v", err)
	}
	actIndexMapping := im.(*mapping.IndexMappingImpl)

	exp, err := json.Marshal(expIndexMapping)
	if err != nil {
		t.Fatalf("%v", err)
	}
	act, err := json.Marshal(actIndexMapping)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(exp, act) {
		t.Fatalf("expected content to see %v, saw %v", exp, act)
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	expIndexType := indexType

	req := &empty.Empty{}
	res, err := client.GetIndexConfig(req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexType := res.IndexConfig.IndexType

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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	expIndexStorageType := indexStorageType

	req := &empty.Empty{}
	res, err := client.GetIndexConfig(req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actIndexStorageType := res.IndexConfig.IndexStorageType

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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	req := &empty.Empty{}
	res, err := client.GetIndexStats(req)
	if err != nil {
		t.Fatalf("%v", err)
	}

	is, err := protobuf.MarshalAny(res.IndexStats)
	if err != nil {
		t.Fatalf("%v", err)
	}
	actIndexStats := *is.(*map[string]interface{})

	if !reflect.DeepEqual(expIndexStats, actIndexStats) {
		t.Fatalf("expected content to see %v, saw %v", expIndexStats, actIndexStats)
	}
}

func TestServer_Index(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	// index document
	docPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
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
	doc1 := &index.Document{}
	err = index.UnmarshalDocument(docBytes1, doc1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	req := &index.IndexRequest{
		Id:     doc1.Id,
		Fields: doc1.Fields,
	}
	_, err = client.Index(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
}

func TestServer_Get(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerGrpcAddress := ""
	shardId := ""
	peerGrpcAddress := ""
	grpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	// index document
	docPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
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
	doc1 := &index.Document{}
	err = index.UnmarshalDocument(docBytes1, doc1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexReq := &index.IndexRequest{
		Id:     doc1.Id,
		Fields: doc1.Fields,
	}
	_, err = client.Index(indexReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get document
	getReq := &index.GetRequest{Id: "enwiki_1"}
	getRes, err := client.Get(getReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expFields, err := protobuf.MarshalAny(doc1.Fields)
	if err != nil {
		t.Fatalf("%v", err)
	}
	actFields, err := protobuf.MarshalAny(getRes.Fields)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !cmp.Equal(expFields, actFields) {
		t.Fatalf("expected content to see %v, saw %v", expFields, actFields)
	}

	// get non-existing document
	getReq2 := &index.GetRequest{Id: "non-existing"}
	getRes2, err := client.Get(getReq2)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			// noop
		default:
			t.Fatalf("%v", err)
		}
	}
	if getRes2 != nil {
		t.Fatalf("expected content to see nil, saw %v", getRes2)
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	// index document
	docPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
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
	doc1 := &index.Document{}
	err = index.UnmarshalDocument(docBytes1, doc1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexReq := &index.IndexRequest{
		Id:     doc1.Id,
		Fields: doc1.Fields,
	}
	_, err = client.Index(indexReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get document
	getReq := &index.GetRequest{Id: "enwiki_1"}
	getRes, err := client.Get(getReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expFields, err := protobuf.MarshalAny(doc1.Fields)
	if err != nil {
		t.Fatalf("%v", err)
	}
	actFields, err := protobuf.MarshalAny(getRes.Fields)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !cmp.Equal(expFields, actFields) {
		t.Fatalf("expected content to see %v, saw %v", expFields, actFields)
	}

	// delete document
	deleteReq := &index.DeleteRequest{Id: "enwiki_1"}
	_, err = client.Delete(deleteReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get document again
	getRes, err = client.Get(getReq)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			// noop
		default:
			t.Fatalf("%v", err)
		}
	}
	if getRes != nil {
		t.Fatalf("expected content to see nil, saw %v", getRes)
	}

	// delete non-existing document
	deleteReq2 := &index.DeleteRequest{Id: "non-existing"}
	_, err = client.Delete(deleteReq2)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			// noop
		default:
			t.Fatalf("%v", err)
		}
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewayAddress,
			HttpAddress:        httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	server, err := NewServer(managerGrpcAddress, shardId, peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	// index document
	docPath1 := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
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
	doc1 := &index.Document{}
	err = index.UnmarshalDocument(docBytes1, doc1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexReq := &index.IndexRequest{
		Id:     doc1.Id,
		Fields: doc1.Fields,
	}
	_, err = client.Index(indexReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get document
	getReq := &index.GetRequest{Id: "enwiki_1"}
	getRes, err := client.Get(getReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expFields, err := protobuf.MarshalAny(doc1.Fields)
	if err != nil {
		t.Fatalf("%v", err)
	}
	actFields, err := protobuf.MarshalAny(getRes.Fields)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if !cmp.Equal(expFields, actFields) {
		t.Fatalf("expected content to see %v, saw %v", expFields, actFields)
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

	searchReq := &index.SearchRequest{}
	marshaler := JsonMarshaler{}
	err = marshaler.Unmarshal(searchRequestByte, searchReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	searchRes, err := client.Search(searchReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	searchResult, err := protobuf.MarshalAny(searchRes.SearchResult)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expTotal := uint64(1)
	actTotal := searchResult.(*bleve.SearchResult).Total
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress1,
			GrpcGatewayAddress: grpcGatewayAddress1,
			HttpAddress:        httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress2,
			GrpcGatewayAddress: grpcGatewayAddress2,
			HttpAddress:        httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress3,
			GrpcGatewayAddress: grpcGatewayAddress3,
			HttpAddress:        httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress1,
			GrpcGatewayAddress: grpcGatewayAddress1,
			HttpAddress:        httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress2,
			GrpcGatewayAddress: grpcGatewayAddress2,
			HttpAddress:        httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress3,
			GrpcGatewayAddress: grpcGatewayAddress3,
			HttpAddress:        httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
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

	healthinessReq := &index.NodeHealthCheckRequest{Probe: index.NodeHealthCheckRequest_HEALTHINESS}
	livenessReq := &index.NodeHealthCheckRequest{Probe: index.NodeHealthCheckRequest_LIVENESS}
	readinessReq := &index.NodeHealthCheckRequest{Probe: index.NodeHealthCheckRequest_READINESS}

	// healthiness
	healthinessRes1, err := client1.NodeHealthCheck(healthinessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness1 := index.NodeHealthCheckResponse_HEALTHY
	actHealthiness1 := healthinessRes1.State
	if expHealthiness1 != actHealthiness1 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness1, actHealthiness1)
	}

	// liveness
	livenessRes1, err := client1.NodeHealthCheck(livenessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness1 := index.NodeHealthCheckResponse_ALIVE
	actLiveness1 := livenessRes1.State
	if expLiveness1 != actLiveness1 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness1, actLiveness1)
	}

	// readiness
	readinessRes1, err := client1.NodeHealthCheck(readinessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness1 := index.NodeHealthCheckResponse_READY
	actReadiness1 := readinessRes1.State
	if expReadiness1 != actReadiness1 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness1, actReadiness1)
	}

	// healthiness
	healthinessRes2, err := client2.NodeHealthCheck(healthinessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness2 := index.NodeHealthCheckResponse_HEALTHY
	actHealthiness2 := healthinessRes2.State
	if expHealthiness2 != actHealthiness2 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness2, actHealthiness2)
	}

	// liveness
	livenessRes2, err := client2.NodeHealthCheck(livenessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness2 := index.NodeHealthCheckResponse_ALIVE
	actLiveness2 := livenessRes2.State
	if expLiveness2 != actLiveness2 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness2, actLiveness2)
	}

	// readiness
	readinessRes2, err := client2.NodeHealthCheck(readinessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness2 := index.NodeHealthCheckResponse_READY
	actReadiness2 := readinessRes2.State
	if expReadiness2 != actReadiness2 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness2, actReadiness2)
	}

	// healthiness
	healthinessRes3, err := client3.NodeHealthCheck(healthinessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness3 := index.NodeHealthCheckResponse_HEALTHY
	actHealthiness3 := healthinessRes3.State
	if expHealthiness3 != actHealthiness3 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness3, actHealthiness3)
	}

	// liveness
	livenessRes3, err := client3.NodeHealthCheck(livenessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness3 := index.NodeHealthCheckResponse_ALIVE
	actLiveness3 := livenessRes3.State
	if expLiveness3 != actLiveness3 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness3, actLiveness3)
	}

	// readiness
	readinessRes3, err := client3.NodeHealthCheck(readinessReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness3 := index.NodeHealthCheckResponse_READY
	actReadiness3 := readinessRes3.State
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress1,
			GrpcGatewayAddress: grpcGatewayAddress1,
			HttpAddress:        httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress2,
			GrpcGatewayAddress: grpcGatewayAddress2,
			HttpAddress:        httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress3,
			GrpcGatewayAddress: grpcGatewayAddress3,
			HttpAddress:        httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
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
	node11, err := client1.NodeInfo(&empty.Empty{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode11 := &index.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       index.Node_LEADER,
		Metadata: &index.Metadata{
			GrpcAddress:        grpcAddress1,
			GrpcGatewayAddress: grpcGatewayAddress1,
			HttpAddress:        httpAddress1,
		},
	}
	actNode11 := node11.Node
	if !reflect.DeepEqual(expNode11, actNode11) {
		t.Fatalf("expected content to see %v, saw %v", expNode11, actNode11)
	}

	node21, err := client2.NodeInfo(&empty.Empty{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode21 := &index.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       index.Node_FOLLOWER,
		Metadata: &index.Metadata{
			GrpcAddress:        grpcAddress2,
			GrpcGatewayAddress: grpcGatewayAddress2,
			HttpAddress:        httpAddress2,
		},
	}
	actNode21 := node21.Node
	if !reflect.DeepEqual(expNode21, actNode21) {
		t.Fatalf("expected content to see %v, saw %v", expNode21, actNode21)
	}

	node31, err := client3.NodeInfo(&empty.Empty{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode31 := &index.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       index.Node_FOLLOWER,
		Metadata: &index.Metadata{
			GrpcAddress:        grpcAddress3,
			GrpcGatewayAddress: grpcGatewayAddress3,
			HttpAddress:        httpAddress3,
		},
	}
	actNode31 := node31.Node
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress1,
			GrpcGatewayAddress: grpcGatewayAddress1,
			HttpAddress:        httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	server1, err := NewServer(managerGrpcAddress1, shardId1, peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress2,
			GrpcGatewayAddress: grpcGatewayAddress2,
			HttpAddress:        httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	server2, err := NewServer(managerGrpcAddress2, shardId2, peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress3,
			GrpcGatewayAddress: grpcGatewayAddress3,
			HttpAddress:        httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	server3, err := NewServer(managerGrpcAddress3, shardId3, peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
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
	cluster1, err := client1.ClusterInfo(&empty.Empty{})
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
					GrpcAddress:        grpcAddress1,
					GrpcGatewayAddress: grpcGatewayAddress1,
					HttpAddress:        httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress:        grpcAddress2,
					GrpcGatewayAddress: grpcGatewayAddress2,
					HttpAddress:        httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress:        grpcAddress3,
					GrpcGatewayAddress: grpcGatewayAddress3,
					HttpAddress:        httpAddress3,
				},
			},
		},
	}
	actCluster1 := cluster1.Cluster
	if !reflect.DeepEqual(expCluster1, actCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expCluster1, actCluster1)
	}

	cluster2, err := client2.ClusterInfo(&empty.Empty{})
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
					GrpcAddress:        grpcAddress1,
					GrpcGatewayAddress: grpcGatewayAddress1,
					HttpAddress:        httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress:        grpcAddress2,
					GrpcGatewayAddress: grpcGatewayAddress2,
					HttpAddress:        httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress:        grpcAddress3,
					GrpcGatewayAddress: grpcGatewayAddress3,
					HttpAddress:        httpAddress3,
				},
			},
		},
	}
	actCluster2 := cluster2.Cluster
	if !reflect.DeepEqual(expCluster2, actCluster2) {
		t.Fatalf("expected content to see %v, saw %v", expCluster2, actCluster2)
	}

	cluster3, err := client3.ClusterInfo(&empty.Empty{})
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
					GrpcAddress:        grpcAddress1,
					GrpcGatewayAddress: grpcGatewayAddress1,
					HttpAddress:        httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress:        grpcAddress2,
					GrpcGatewayAddress: grpcGatewayAddress2,
					HttpAddress:        httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress:        grpcAddress3,
					GrpcGatewayAddress: grpcGatewayAddress3,
					HttpAddress:        httpAddress3,
				},
			},
		},
	}
	actCluster3 := cluster3.Cluster
	if !reflect.DeepEqual(expCluster3, actCluster3) {
		t.Fatalf("expected content to see %v, saw %v", expCluster3, actCluster3)
	}
}
