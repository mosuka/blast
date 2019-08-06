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
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/mosuka/blast/strutils"
	"github.com/mosuka/blast/testutils"
)

func TestServer_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	// create server
	server, err := NewServer(peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

func TestServer_HealthCheck(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	// create server
	server, err := NewServer(peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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
	healthiness, err := client.NodeHealthCheck(management.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness := management.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness := healthiness
	if expHealthiness != actHealthiness {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness, actHealthiness)
	}

	// liveness
	liveness, err := client.NodeHealthCheck(management.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness := management.NodeHealthCheckResponse_ALIVE.String()
	actLiveness := liveness
	if expLiveness != actLiveness {
		t.Fatalf("expected content to see %v, saw %v", expLiveness, actLiveness)
	}

	// readiness
	readiness, err := client.NodeHealthCheck(management.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness := management.NodeHealthCheckResponse_READY.String()
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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	// create server
	server, err := NewServer(peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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
	expNodeInfo := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_LEADER,
		Metadata: &management.Metadata{
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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	// create server
	server, err := NewServer(peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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
	expCluster := &management.Cluster{
		Nodes: map[string]*management.Node{
			nodeId: {
				Id:          nodeId,
				BindAddress: bindAddress,
				State:       management.Node_LEADER,
				Metadata: &management.Metadata{
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

func TestServer_SetState(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	// create server
	server, err := NewServer(peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	// set value
	err = client.Set("test/key1", "val1")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get value
	val1, err := client.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}

	expVal1 := "val1"

	actVal1 := *val1.(*string)

	if expVal1 != actVal1 {
		t.Fatalf("expected content to see %v, saw %v", expVal1, actVal1)
	}
}

func TestServer_GetState(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	// create server
	server, err := NewServer(peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	// set value
	err = client.Set("test/key1", "val1")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get value
	val1, err := client.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}

	expVal1 := "val1"

	actVal1 := *val1.(*string)

	if expVal1 != actVal1 {
		t.Fatalf("expected content to see %v, saw %v", expVal1, actVal1)
	}
}

func TestServer_DeleteState(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress,
			HttpAddress: httpAddress,
		},
	}

	indexMapping, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType := "upside_down"
	indexStorageType := "boltdb"

	// create server
	server, err := NewServer(peerGrpcAddress, node, dataDir, raftStorageType, indexMapping, indexType, indexStorageType, logger, grpcLogger, httpAccessLogger)
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

	// set value
	err = client.Set("test/key1", "val1")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get value
	val1, err := client.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}

	expVal1 := "val1"

	actVal1 := *val1.(*string)

	if expVal1 != actVal1 {
		t.Fatalf("expected content to see %v, saw %v", expVal1, actVal1)
	}

	// delete value
	err = client.Delete("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}

	val1, err = client.Get("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}

	if val1 != nil {
		t.Fatalf("%v", err)
	}

	// delete non-existing data
	err = client.Delete("test/non-existing")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
}

func TestCluster_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node1 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	// create server
	server1, err := NewServer(peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

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

	node2 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	// create server
	server2, err := NewServer(peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

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

	node3 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	// create server
	server3, err := NewServer(peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server3.Start()

	// sleep
	time.Sleep(5 * time.Second)
}

func TestCluster_HealthCheck(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node1 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	// create server
	server1, err := NewServer(peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

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

	node2 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	// create server
	server2, err := NewServer(peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

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

	node3 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	// create server
	server3, err := NewServer(peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
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
	healthiness1, err := client1.NodeHealthCheck(management.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness1 := management.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness1 := healthiness1
	if expHealthiness1 != actHealthiness1 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness1, actHealthiness1)
	}

	// liveness
	liveness1, err := client1.NodeHealthCheck(management.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness1 := management.NodeHealthCheckResponse_ALIVE.String()
	actLiveness1 := liveness1
	if expLiveness1 != actLiveness1 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness1, actLiveness1)
	}

	// readiness
	readiness1, err := client1.NodeHealthCheck(management.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness1 := management.NodeHealthCheckResponse_READY.String()
	actReadiness1 := readiness1
	if expReadiness1 != actReadiness1 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness1, actReadiness1)
	}

	// healthiness
	healthiness2, err := client2.NodeHealthCheck(management.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness2 := management.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness2 := healthiness2
	if expHealthiness2 != actHealthiness2 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness2, actHealthiness2)
	}

	// liveness
	liveness2, err := client2.NodeHealthCheck(management.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness2 := management.NodeHealthCheckResponse_ALIVE.String()
	actLiveness2 := liveness2
	if expLiveness2 != actLiveness2 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness2, actLiveness2)
	}

	// readiness
	readiness2, err := client2.NodeHealthCheck(management.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness2 := management.NodeHealthCheckResponse_READY.String()
	actReadiness2 := readiness2
	if expReadiness2 != actReadiness2 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness2, actReadiness2)
	}

	// healthiness
	healthiness3, err := client3.NodeHealthCheck(management.NodeHealthCheckRequest_HEALTHINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness3 := management.NodeHealthCheckResponse_HEALTHY.String()
	actHealthiness3 := healthiness3
	if expHealthiness3 != actHealthiness3 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness3, actHealthiness3)
	}

	// liveness
	liveness3, err := client3.NodeHealthCheck(management.NodeHealthCheckRequest_LIVENESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness3 := management.NodeHealthCheckResponse_ALIVE.String()
	actLiveness3 := liveness3
	if expLiveness3 != actLiveness3 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness3, actLiveness3)
	}

	// readiness
	readiness3, err := client3.NodeHealthCheck(management.NodeHealthCheckRequest_READINESS.String())
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness3 := management.NodeHealthCheckResponse_READY.String()
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

	node1 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	// create server
	server1, err := NewServer(peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

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

	node2 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	// create server
	server2, err := NewServer(peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

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

	node3 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	// create server
	server3, err := NewServer(peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
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
	expNode11 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_LEADER,
		Metadata: &management.Metadata{
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
	expNode21 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_FOLLOWER,
		Metadata: &management.Metadata{
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
	expNode31 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_FOLLOWER,
		Metadata: &management.Metadata{
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

	node1 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	// create server
	server1, err := NewServer(peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

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

	node2 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	// create server
	server2, err := NewServer(peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

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

	node3 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	// create server
	server3, err := NewServer(peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
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
	expCluster1 := &management.Cluster{
		Nodes: map[string]*management.Node{
			nodeId1: {
				Id:          nodeId1,
				BindAddress: bindAddress1,
				State:       management.Node_LEADER,
				Metadata: &management.Metadata{
					GrpcAddress: grpcAddress1,
					HttpAddress: httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress: grpcAddress2,
					HttpAddress: httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
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
	expCluster2 := &management.Cluster{
		Nodes: map[string]*management.Node{
			nodeId1: {
				Id:          nodeId1,
				BindAddress: bindAddress1,
				State:       management.Node_LEADER,
				Metadata: &management.Metadata{
					GrpcAddress: grpcAddress1,
					HttpAddress: httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress: grpcAddress2,
					HttpAddress: httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
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
	expCluster3 := &management.Cluster{
		Nodes: map[string]*management.Node{
			nodeId1: {
				Id:          nodeId1,
				BindAddress: bindAddress1,
				State:       management.Node_LEADER,
				Metadata: &management.Metadata{
					GrpcAddress: grpcAddress1,
					HttpAddress: httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress: grpcAddress2,
					HttpAddress: httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
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

func TestCluster_SetState(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node1 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	// create server
	server1, err := NewServer(peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

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

	node2 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	// create server
	server2, err := NewServer(peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

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

	node3 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	// create server
	server3, err := NewServer(peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
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

	err = client1.Set("test/key1", "val1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val11, err := client1.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if expVal11 != actVal11 {
		t.Fatalf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	val21, err := client2.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if expVal21 != actVal21 {
		t.Fatalf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	val31, err := client3.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if expVal31 != actVal31 {
		t.Fatalf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	err = client2.Set("test/key2", "val2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val12, err := client1.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if expVal12 != actVal12 {
		t.Fatalf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	val22, err := client2.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if expVal22 != actVal22 {
		t.Fatalf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	val32, err := client3.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if expVal32 != actVal32 {
		t.Fatalf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	err = client3.Set("test/key3", "val3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val13, err := client1.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if expVal13 != actVal13 {
		t.Fatalf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	val23, err := client2.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if expVal23 != actVal23 {
		t.Fatalf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	val33, err := client3.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if expVal33 != actVal33 {
		t.Fatalf("expected content to see %v, saw %v", expVal33, actVal33)
	}
}

func TestCluster_GetState(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node1 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	// create server
	server1, err := NewServer(peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

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

	node2 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	// create server
	server2, err := NewServer(peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

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

	node3 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	// create server
	server3, err := NewServer(peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
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

	err = client1.Set("test/key1", "val1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val11, err := client1.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if expVal11 != actVal11 {
		t.Fatalf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	val21, err := client2.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if expVal21 != actVal21 {
		t.Fatalf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	val31, err := client3.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if expVal31 != actVal31 {
		t.Fatalf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	err = client2.Set("test/key2", "val2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val12, err := client1.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if expVal12 != actVal12 {
		t.Fatalf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	val22, err := client2.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if expVal22 != actVal22 {
		t.Fatalf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	val32, err := client3.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if expVal32 != actVal32 {
		t.Fatalf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	err = client3.Set("test/key3", "val3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val13, err := client1.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if expVal13 != actVal13 {
		t.Fatalf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	val23, err := client2.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if expVal23 != actVal23 {
		t.Fatalf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	val33, err := client3.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if expVal33 != actVal33 {
		t.Fatalf("expected content to see %v, saw %v", expVal33, actVal33)
	}
}

func TestCluster_DeleteState(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node1 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress1,
			HttpAddress: httpAddress1,
		},
	}

	indexMapping1, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType1 := "upside_down"
	indexStorageType1 := "boltdb"

	// create server
	server1, err := NewServer(peerGrpcAddress1, node1, dataDir1, raftStorageType1, indexMapping1, indexType1, indexStorageType1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server1 != nil {
			server1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server1.Start()

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

	node2 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress2,
			HttpAddress: httpAddress2,
		},
	}

	indexMapping2, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType2 := "upside_down"
	indexStorageType2 := "boltdb"

	// create server
	server2, err := NewServer(peerGrpcAddress2, node2, dataDir2, raftStorageType2, indexMapping2, indexType2, indexStorageType2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server2 != nil {
			server2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	server2.Start()

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

	node3 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: grpcAddress3,
			HttpAddress: httpAddress3,
		},
	}

	indexMapping3, err := indexutils.NewIndexMappingFromFile(filepath.Join(curDir, "../example/wiki_index_mapping.json"))
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexType3 := "upside_down"
	indexStorageType3 := "boltdb"

	// create server
	server3, err := NewServer(peerGrpcAddress3, node3, dataDir3, raftStorageType3, indexMapping3, indexType3, indexStorageType3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if server3 != nil {
			server3.Stop()
		}
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

	// set test data before delete
	err = client1.Set("test/key1", "val1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val11, err := client1.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if expVal11 != actVal11 {
		t.Fatalf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	val21, err := client2.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if expVal21 != actVal21 {
		t.Fatalf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	val31, err := client3.Get("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if expVal31 != actVal31 {
		t.Fatalf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	err = client2.Set("test/key2", "val2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val12, err := client1.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if expVal12 != actVal12 {
		t.Fatalf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	val22, err := client2.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if expVal22 != actVal22 {
		t.Fatalf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	val32, err := client3.Get("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if expVal32 != actVal32 {
		t.Fatalf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	err = client3.Set("test/key3", "val3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val13, err := client1.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if expVal13 != actVal13 {
		t.Fatalf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	val23, err := client2.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if expVal23 != actVal23 {
		t.Fatalf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	val33, err := client3.Get("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if expVal33 != actVal33 {
		t.Fatalf("expected content to see %v, saw %v", expVal33, actVal33)
	}

	// delete
	err = client1.Delete("test/key1")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val11, err = client1.Get("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val11 != nil {
		t.Fatalf("%v", err)
	}
	val21, err = client2.Get("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val21 != nil {
		t.Fatalf("%v", err)
	}
	val31, err = client3.Get("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val31 != nil {
		t.Fatalf("%v", err)
	}

	err = client2.Delete("test/key2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val12, err = client1.Get("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val12 != nil {
		t.Fatalf("%v", err)
	}
	val22, err = client2.Get("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val22 != nil {
		t.Fatalf("%v", err)
	}
	val32, err = client3.Get("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val32 != nil {
		t.Fatalf("%v", err)
	}

	err = client3.Delete("test/key3")
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	val13, err = client1.Get("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val13 != nil {
		t.Fatalf("%v", err)
	}
	val23, err = client2.Get("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val23 != nil {
		t.Fatalf("%v", err)
	}
	val33, err = client3.Get("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Fatalf("%v", err)
	}
	if val33 != nil {
		t.Fatalf("%v", err)
	}

	// delete non-existing data from manager1
	err = client1.Delete("test/non-existing")
	if err == nil {
		t.Fatalf("%v", err)
	}

	// delete non-existing data from manager2
	err = client2.Delete("test/non-existing")
	if err == nil {
		t.Fatalf("%v", err)
	}

	// delete non-existing data from manager3
	err = client3.Delete("test/non-existing")
	if err == nil {
		t.Fatalf("%v", err)
	}
}
