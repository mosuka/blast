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

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf"
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
	reqHealthiness := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_HEALTHINESS}
	resHealthiness, err := client.NodeHealthCheck(reqHealthiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness := management.NodeHealthCheckResponse_HEALTHY
	actHealthiness := resHealthiness.State
	if expHealthiness != actHealthiness {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness, actHealthiness)
	}

	// liveness
	reqLiveness := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_LIVENESS}
	resLiveness, err := client.NodeHealthCheck(reqLiveness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness := management.NodeHealthCheckResponse_ALIVE
	actLiveness := resLiveness.State
	if expLiveness != actLiveness {
		t.Fatalf("expected content to see %v, saw %v", expLiveness, actLiveness)
	}

	// readiness
	reqReadiness := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_READINESS}
	resReadiness, err := client.NodeHealthCheck(reqReadiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness := management.NodeHealthCheckResponse_READY
	actReadiness := resReadiness.State
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
	grpcGatewawyAddress := fmt.Sprintf(":%d", testutils.TmpPort())
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
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewawyAddress,
			HttpAddress:        httpAddress,
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
	res, err := client.NodeInfo(&empty.Empty{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNodeInfo := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_LEADER,
		Metadata: &management.Metadata{
			GrpcAddress:        grpcAddress,
			GrpcGatewayAddress: grpcGatewawyAddress,
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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
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
	res, err := client.ClusterInfo(&empty.Empty{})
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

func TestServer_Set(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
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
	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny("val1", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq := &management.SetRequest{
		Key:   "test/key1",
		Value: valueAny,
	}
	_, err = client.Set(setReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get value
	getReq := &management.GetRequest{
		Key: "test/key1",
	}
	getRes, err := client.Get(getReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expVal1 := "val1"

	val1, err := protobuf.MarshalAny(getRes.Value)
	actVal1 := *val1.(*string)

	if !cmp.Equal(expVal1, actVal1) {
		t.Fatalf("expected content to see %v, saw %v", expVal1, actVal1)
	}
}

func TestServer_Get(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
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
	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny("val1", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq := &management.SetRequest{
		Key:   "test/key1",
		Value: valueAny,
	}
	_, err = client.Set(setReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get value
	getReq := &management.GetRequest{Key: "test/key1"}
	getRes, err := client.Get(getReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expVal1 := "val1"

	val1, err := protobuf.MarshalAny(getRes.Value)
	actVal1 := *val1.(*string)

	if !cmp.Equal(expVal1, actVal1) {
		t.Fatalf("expected content to see %v, saw %v", expVal1, actVal1)
	}
}

func TestServer_Delete(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	node := &management.Node{
		Id:          nodeId,
		BindAddress: bindAddress,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
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
	valueAny := &any.Any{}
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = protobuf.UnmarshalAny("val1", valueAny)
	setReq := &management.SetRequest{
		Key:   "test/key1",
		Value: valueAny,
	}
	_, err = client.Set(setReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// get value
	getReq := &management.GetRequest{
		Key: "test/key1",
	}
	res, err := client.Get(getReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	expVal1 := "val1"

	val1, err := protobuf.MarshalAny(res.Value)
	actVal1 := *val1.(*string)

	if !cmp.Equal(expVal1, actVal1) {
		t.Fatalf("expected content to see %v, saw %v", expVal1, actVal1)
	}

	// delete value
	deleteReq := &management.DeleteRequest{
		Key: "test/key1",
	}
	_, err = client.Delete(deleteReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// delete non-existing data
	deleteNonExistingReq := &management.DeleteRequest{
		Key: "test/non-existing",
	}
	_, err = client.Delete(deleteNonExistingReq)
	if err != nil {
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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

	reqHealtiness := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_HEALTHINESS}
	reqLiveness := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_LIVENESS}
	reqReadiness := &management.NodeHealthCheckRequest{Probe: management.NodeHealthCheckRequest_READINESS}

	// healthiness
	resHealthiness1, err := client1.NodeHealthCheck(reqHealtiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness1 := management.NodeHealthCheckResponse_HEALTHY
	actHealthiness1 := resHealthiness1.State
	if expHealthiness1 != actHealthiness1 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness1, actHealthiness1)
	}

	// liveness
	resLiveness1, err := client1.NodeHealthCheck(reqLiveness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness1 := management.NodeHealthCheckResponse_ALIVE
	actLiveness1 := resLiveness1.State
	if expLiveness1 != actLiveness1 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness1, actLiveness1)
	}

	// readiness
	resReadiness1, err := client1.NodeHealthCheck(reqReadiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness1 := management.NodeHealthCheckResponse_READY
	actReadiness1 := resReadiness1.State
	if expReadiness1 != actReadiness1 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness1, actReadiness1)
	}

	// healthiness
	resHealthiness2, err := client2.NodeHealthCheck(reqHealtiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness2 := management.NodeHealthCheckResponse_HEALTHY
	actHealthiness2 := resHealthiness2.State
	if expHealthiness2 != actHealthiness2 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness2, actHealthiness2)
	}

	// liveness
	resLiveness2, err := client2.NodeHealthCheck(reqLiveness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness2 := management.NodeHealthCheckResponse_ALIVE
	actLiveness2 := resLiveness2.State
	if expLiveness2 != actLiveness2 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness2, actLiveness2)
	}

	// readiness
	resReadiness2, err := client2.NodeHealthCheck(reqReadiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness2 := management.NodeHealthCheckResponse_READY
	actReadiness2 := resReadiness2.State
	if expReadiness2 != actReadiness2 {
		t.Fatalf("expected content to see %v, saw %v", expReadiness2, actReadiness2)
	}

	// healthiness
	resHealthiness3, err := client3.NodeHealthCheck(reqHealtiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expHealthiness3 := management.NodeHealthCheckResponse_HEALTHY
	actHealthiness3 := resHealthiness3.State
	if expHealthiness3 != actHealthiness3 {
		t.Fatalf("expected content to see %v, saw %v", expHealthiness3, actHealthiness3)
	}

	// liveness
	resLiveness3, err := client3.NodeHealthCheck(reqLiveness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expLiveness3 := management.NodeHealthCheckResponse_ALIVE
	actLiveness3 := resLiveness3.State
	if expLiveness3 != actLiveness3 {
		t.Fatalf("expected content to see %v, saw %v", expLiveness3, actLiveness3)
	}

	// readiness
	resReadiness3, err := client3.NodeHealthCheck(reqReadiness)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expReadiness3 := management.NodeHealthCheckResponse_READY
	actReadiness3 := resReadiness3.State
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	req := &empty.Empty{}
	resNodeInfo11, err := client1.NodeInfo(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode11 := &management.Node{
		Id:          nodeId1,
		BindAddress: bindAddress1,
		State:       management.Node_LEADER,
		Metadata: &management.Metadata{
			GrpcAddress:        grpcAddress1,
			GrpcGatewayAddress: grpcGatewayAddress1,
			HttpAddress:        httpAddress1,
		},
	}
	actNode11 := resNodeInfo11.Node
	if !reflect.DeepEqual(expNode11, actNode11) {
		t.Fatalf("expected content to see %v, saw %v", expNode11, actNode11)
	}

	resNodeInfo21, err := client2.NodeInfo(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode21 := &management.Node{
		Id:          nodeId2,
		BindAddress: bindAddress2,
		State:       management.Node_FOLLOWER,
		Metadata: &management.Metadata{
			GrpcAddress:        grpcAddress2,
			GrpcGatewayAddress: grpcGatewayAddress2,
			HttpAddress:        httpAddress2,
		},
	}
	actNode21 := resNodeInfo21.Node
	if !reflect.DeepEqual(expNode21, actNode21) {
		t.Fatalf("expected content to see %v, saw %v", expNode21, actNode21)
	}

	resNodeInfo31, err := client3.NodeInfo(req)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expNode31 := &management.Node{
		Id:          nodeId3,
		BindAddress: bindAddress3,
		State:       management.Node_FOLLOWER,
		Metadata: &management.Metadata{
			GrpcAddress:        grpcAddress3,
			GrpcGatewayAddress: grpcGatewayAddress3,
			HttpAddress:        httpAddress3,
		},
	}
	actNode31 := resNodeInfo31.Node
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
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
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
	req := &empty.Empty{}
	resClusterInfo1, err := client1.ClusterInfo(req)
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
					GrpcAddress:        grpcAddress1,
					GrpcGatewayAddress: grpcGatewayAddress1,
					HttpAddress:        httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress:        grpcAddress2,
					GrpcGatewayAddress: grpcGatewayAddress2,
					HttpAddress:        httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress:        grpcAddress3,
					GrpcGatewayAddress: grpcGatewayAddress3,
					HttpAddress:        httpAddress3,
				},
			},
		},
	}
	actCluster1 := resClusterInfo1.Cluster
	if !reflect.DeepEqual(expCluster1, actCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expCluster1, actCluster1)
	}

	resClusterInfo2, err := client2.ClusterInfo(req)
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
					GrpcAddress:        grpcAddress1,
					GrpcGatewayAddress: grpcGatewayAddress1,
					HttpAddress:        httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress:        grpcAddress2,
					GrpcGatewayAddress: grpcGatewayAddress2,
					HttpAddress:        httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress:        grpcAddress3,
					GrpcGatewayAddress: grpcGatewayAddress3,
					HttpAddress:        httpAddress3,
				},
			},
		},
	}
	actCluster2 := resClusterInfo2.Cluster
	if !reflect.DeepEqual(expCluster2, actCluster2) {
		t.Fatalf("expected content to see %v, saw %v", expCluster2, actCluster2)
	}

	resClusterInfo3, err := client3.ClusterInfo(req)
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
					GrpcAddress:        grpcAddress1,
					GrpcGatewayAddress: grpcGatewayAddress1,
					HttpAddress:        httpAddress1,
				},
			},
			nodeId2: {
				Id:          nodeId2,
				BindAddress: bindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress:        grpcAddress2,
					GrpcGatewayAddress: grpcGatewayAddress2,
					HttpAddress:        httpAddress2,
				},
			},
			nodeId3: {
				Id:          nodeId3,
				BindAddress: bindAddress3,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress:        grpcAddress3,
					GrpcGatewayAddress: grpcGatewayAddress3,
					HttpAddress:        httpAddress3,
				},
			},
		},
	}
	actCluster3 := resClusterInfo3.Cluster
	if !reflect.DeepEqual(expCluster3, actCluster3) {
		t.Fatalf("expected content to see %v, saw %v", expCluster3, actCluster3)
	}
}

func TestCluster_Set(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	peerGrpcAddress1 := ""
	grpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId1 := "node-1"
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

	// sleep
	time.Sleep(5 * time.Second)

	peerGrpcAddress2 := grpcAddress1
	grpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId2 := "node-2"
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

	// sleep
	time.Sleep(5 * time.Second)

	peerGrpcAddress3 := grpcAddress1
	grpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId3 := "node-3"
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

	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny("val1", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq1 := &management.SetRequest{
		Key:   "test/key1",
		Value: valueAny,
	}
	_, err = client1.Set(setReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq1 := &management.GetRequest{
		Key: "test/key1",
	}
	getRes11, err := client1.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val11, err := protobuf.MarshalAny(getRes11.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if !cmp.Equal(expVal11, actVal11) {
		t.Fatalf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	getRes21, err := client2.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val21, err := protobuf.MarshalAny(getRes21.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if !cmp.Equal(expVal21, actVal21) {
		t.Fatalf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	getRes31, err := client3.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val31, err := protobuf.MarshalAny(getRes31.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if !cmp.Equal(expVal31, actVal31) {
		t.Fatalf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	valueAny = &any.Any{}
	err = protobuf.UnmarshalAny("val2", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq2 := &management.SetRequest{
		Key:   "test/key2",
		Value: valueAny,
	}
	_, err = client2.Set(setReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq2 := &management.GetRequest{
		Key: "test/key2",
	}
	getRes12, err := client1.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val12, err := protobuf.MarshalAny(getRes12.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if !cmp.Equal(expVal12, actVal12) {
		t.Fatalf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	getRes22, err := client2.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val22, err := protobuf.MarshalAny(getRes22.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if !cmp.Equal(expVal22, actVal22) {
		t.Fatalf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	getRes32, err := client3.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val32, err := protobuf.MarshalAny(getRes32.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if !cmp.Equal(expVal32, actVal32) {
		t.Fatalf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	valueAny = &any.Any{}
	err = protobuf.UnmarshalAny("val3", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq3 := &management.SetRequest{
		Key:   "test/key3",
		Value: valueAny,
	}
	_, err = client3.Set(setReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq3 := &management.GetRequest{
		Key: "test/key3",
	}
	getRes13, err := client1.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val13, err := protobuf.MarshalAny(getRes13.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if !cmp.Equal(expVal13, actVal13) {
		t.Fatalf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	getRes23, err := client2.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val23, err := protobuf.MarshalAny(getRes23.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if !cmp.Equal(expVal23, actVal23) {
		t.Fatalf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	getRes33, err := client3.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val33, err := protobuf.MarshalAny(getRes33.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if !cmp.Equal(expVal33, actVal33) {
		t.Fatalf("expected content to see %v, saw %v", expVal33, actVal33)
	}
}

func TestCluster_Get(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	peerGrpcAddress1 := ""
	grpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId1 := "node-1"
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

	// sleep
	time.Sleep(5 * time.Second)

	peerGrpcAddress2 := grpcAddress1
	grpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId2 := "node-2"
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

	// sleep
	time.Sleep(5 * time.Second)

	peerGrpcAddress3 := grpcAddress1
	grpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId3 := "node-3"
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

	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny("val1", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq1 := &management.SetRequest{
		Key:   "test/key1",
		Value: valueAny,
	}
	_, err = client1.Set(setReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq1 := &management.GetRequest{
		Key: "test/key1",
	}
	getRes11, err := client1.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val11, err := protobuf.MarshalAny(getRes11.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if !cmp.Equal(expVal11, actVal11) {
		t.Fatalf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	getRes21, err := client2.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val21, err := protobuf.MarshalAny(getRes21.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if !cmp.Equal(expVal21, actVal21) {
		t.Fatalf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	getRes31, err := client3.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val31, err := protobuf.MarshalAny(getRes31.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if !cmp.Equal(expVal31, actVal31) {
		t.Fatalf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	valueAny = &any.Any{}
	err = protobuf.UnmarshalAny("val2", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq2 := &management.SetRequest{
		Key:   "test/key2",
		Value: valueAny,
	}
	_, err = client2.Set(setReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq2 := &management.GetRequest{
		Key: "test/key2",
	}
	getRes12, err := client1.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val12, err := protobuf.MarshalAny(getRes12.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if !cmp.Equal(expVal12, actVal12) {
		t.Fatalf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	getRes22, err := client2.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val22, err := protobuf.MarshalAny(getRes22.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if !cmp.Equal(expVal22, actVal22) {
		t.Fatalf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	getRes32, err := client3.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val32, err := protobuf.MarshalAny(getRes32.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if !cmp.Equal(expVal32, actVal32) {
		t.Fatalf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	valueAny = &any.Any{}
	err = protobuf.UnmarshalAny("val3", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq3 := &management.SetRequest{
		Key:   "test/key3",
		Value: valueAny,
	}
	_, err = client3.Set(setReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq3 := &management.GetRequest{
		Key: "test/key3",
	}
	getRes13, err := client1.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val13, err := protobuf.MarshalAny(getRes13.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if !cmp.Equal(expVal13, actVal13) {
		t.Fatalf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	getRes23, err := client2.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val23, err := protobuf.MarshalAny(getRes23.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if !cmp.Equal(expVal23, actVal23) {
		t.Fatalf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	getRes33, err := client3.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val33, err := protobuf.MarshalAny(getRes33.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if !cmp.Equal(expVal33, actVal33) {
		t.Fatalf("expected content to see %v, saw %v", expVal33, actVal33)
	}
}

func TestCluster_Delete(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	peerGrpcAddress1 := ""
	grpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId1 := "node-1"
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

	// sleep
	time.Sleep(5 * time.Second)

	peerGrpcAddress2 := grpcAddress1
	grpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId2 := "node-2"
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

	// sleep
	time.Sleep(5 * time.Second)

	peerGrpcAddress3 := grpcAddress1
	grpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	grpcGatewayAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	httpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	nodeId3 := "node-3"
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

	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny("val1", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq1 := &management.SetRequest{
		Key:   "test/key1",
		Value: valueAny,
	}
	_, err = client1.Set(setReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq1 := &management.GetRequest{
		Key: "test/key1",
	}
	getRes11, err := client1.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val11, err := protobuf.MarshalAny(getRes11.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal11 := "val1"
	actVal11 := *val11.(*string)
	if !cmp.Equal(expVal11, actVal11) {
		t.Fatalf("expected content to see %v, saw %v", expVal11, actVal11)
	}
	getRes21, err := client2.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val21, err := protobuf.MarshalAny(getRes21.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal21 := "val1"
	actVal21 := *val21.(*string)
	if !cmp.Equal(expVal21, actVal21) {
		t.Fatalf("expected content to see %v, saw %v", expVal21, actVal21)
	}
	getRes31, err := client3.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val31, err := protobuf.MarshalAny(getRes31.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal31 := "val1"
	actVal31 := *val31.(*string)
	if !cmp.Equal(expVal31, actVal31) {
		t.Fatalf("expected content to see %v, saw %v", expVal31, actVal31)
	}

	valueAny = &any.Any{}
	err = protobuf.UnmarshalAny("val2", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq2 := &management.SetRequest{
		Key:   "test/key2",
		Value: valueAny,
	}
	_, err = client2.Set(setReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq2 := &management.GetRequest{
		Key: "test/key2",
	}
	getRes12, err := client1.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val12, err := protobuf.MarshalAny(getRes12.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal12 := "val2"
	actVal12 := *val12.(*string)
	if !cmp.Equal(expVal12, actVal12) {
		t.Fatalf("expected content to see %v, saw %v", expVal12, actVal12)
	}
	getRes22, err := client2.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val22, err := protobuf.MarshalAny(getRes22.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal22 := "val2"
	actVal22 := *val22.(*string)
	if !cmp.Equal(expVal22, actVal22) {
		t.Fatalf("expected content to see %v, saw %v", expVal22, actVal22)
	}
	getRes32, err := client3.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val32, err := protobuf.MarshalAny(getRes32.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal32 := "val2"
	actVal32 := *val32.(*string)
	if !cmp.Equal(expVal32, actVal32) {
		t.Fatalf("expected content to see %v, saw %v", expVal32, actVal32)
	}

	valueAny = &any.Any{}
	err = protobuf.UnmarshalAny("val3", valueAny)
	if err != nil {
		t.Fatalf("%v", err)
	}
	setReq3 := &management.SetRequest{
		Key:   "test/key3",
		Value: valueAny,
	}
	_, err = client3.Set(setReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getReq3 := &management.GetRequest{
		Key: "test/key3",
	}
	getRes13, err := client1.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val13, err := protobuf.MarshalAny(getRes13.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal13 := "val3"
	actVal13 := *val13.(*string)
	if !cmp.Equal(expVal13, actVal13) {
		t.Fatalf("expected content to see %v, saw %v", expVal13, actVal13)
	}
	getRes23, err := client2.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val23, err := protobuf.MarshalAny(getRes23.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal23 := "val3"
	actVal23 := *val23.(*string)
	if !cmp.Equal(expVal23, actVal23) {
		t.Fatalf("expected content to see %v, saw %v", expVal23, actVal23)
	}
	getRes33, err := client3.Get(getReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val33, err := protobuf.MarshalAny(getRes33.Value)
	if err != nil {
		t.Fatalf("%v", err)
	}
	expVal33 := "val3"
	actVal33 := *val33.(*string)
	if !cmp.Equal(expVal33, actVal33) {
		t.Fatalf("expected content to see %v, saw %v", expVal33, actVal33)
	}

	// delete
	deleteReq1 := &management.DeleteRequest{
		Key: "test/key1",
	}
	_, err = client1.Delete(deleteReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getRes11, err = client1.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if getRes11.Value != nil {
		t.Fatalf("%v", err)
	}
	getRes21, err = client2.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if getRes21.Value != nil {
		t.Fatalf("%v", err)
	}
	getRes31, err = client3.Get(getReq1)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if getRes31.Value != nil {
		t.Fatalf("%v", err)
	}

	deleteReq2 := &management.DeleteRequest{
		Key: "test/key2",
	}
	_, err = client2.Delete(deleteReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from all nodes
	getRes12, err = client1.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if getRes12.Value != nil {
		t.Fatalf("%v", err)
	}
	getRes22, err = client2.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if getRes22.Value != nil {
		t.Fatalf("%v", err)
	}
	getRes32, err = client3.Get(getReq2)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if getRes32.Value != nil {
		t.Fatalf("%v", err)
	}

	deleteReq3 := &management.DeleteRequest{
		Key: "test/key2",
	}
	_, err = client3.Delete(deleteReq3)
	if err != nil {
		t.Fatalf("%v", err)
	}
	time.Sleep(2 * time.Second) // wait for data to propagate

	// delete non-existing data from manager1
	deleteNonExistingReq := &management.DeleteRequest{
		Key: "test/non-existing",
	}
	_, err = client1.Delete(deleteNonExistingReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// delete non-existing data from manager2
	_, err = client2.Delete(deleteNonExistingReq)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// delete non-existing data from manager3
	_, err = client3.Delete(deleteNonExistingReq)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
