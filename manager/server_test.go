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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/blevesearch/bleve/mapping"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/testutils"
)

func TestManagerStandalone(t *testing.T) {
	curDir, _ := os.Getwd()

	// index config
	indexMappingFile := filepath.Join(curDir, "../example/wiki_index_mapping.json")
	indexMapping := mapping.NewIndexMapping()
	if indexMappingFile != "" {
		_, err := os.Stat(indexMappingFile)
		if err == nil {
			// read index mapping file
			f, err := os.Open(indexMappingFile)
			if err != nil {
				t.Errorf("%v", err)
			}
			defer func() {
				_ = f.Close()
			}()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf("%v", err)
			}

			err = json.Unmarshal(b, indexMapping)
			if err != nil {
				t.Errorf("%v", err)
			}
		} else if os.IsNotExist(err) {
			t.Errorf("%v", err)
		}
	}
	err := indexMapping.Validate()
	if err != nil {
		t.Errorf("%v", err)
	}
	indexMappingJSON, err := json.Marshal(indexMapping)
	if err != nil {
		t.Errorf("%v", err)
	}
	var indexMappingMap map[string]interface{}
	err = json.Unmarshal(indexMappingJSON, &indexMappingMap)
	if err != nil {
		t.Errorf("%v", err)
	}

	indexType := "upside_down"
	indexStorageType := "boltdb"

	indexConfig := map[string]interface{}{
		"index_mapping":      indexMappingMap,
		"index_type":         indexType,
		"index_storage_type": indexStorageType,
	}

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// node
	nodeId := "manager1"

	bindPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	bindAddr := fmt.Sprintf(":%d", bindPort)

	grpcPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)

	httpPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	httpAddr := fmt.Sprintf(":%d", httpPort)

	dataDir, err := testutils.TmpDir()
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(dataDir)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	// peer address
	peerAddr := ""

	// metadata
	metadata := map[string]interface{}{
		"bind_addr": bindAddr,
		"grpc_addr": grpcAddr,
		"http_addr": httpAddr,
		"data_dir":  dataDir,
	}

	server, err := NewServer(nodeId, metadata, peerAddr, indexConfig, logger.Named(nodeId), httpAccessLogger)
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

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		_ = client.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// liveness
	liveness, err := client.LivenessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expLiveness := protobuf.LivenessProbeResponse_ALIVE.String()
	actLiveness := liveness
	if expLiveness != actLiveness {
		t.Errorf("expected content to see %v, saw %v", expLiveness, actLiveness)
	}

	// readiness
	readiness, err := client.ReadinessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	expReadiness := protobuf.ReadinessProbeResponse_READY.String()
	actReadiness := readiness
	if expLiveness != actLiveness {
		t.Errorf("expected content to see %v, saw %v", expReadiness, actReadiness)
	}

	// get node
	node, err := client.GetNode(nodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": bindAddr,
			"grpc_addr": grpcAddr,
			"http_addr": httpAddr,
			"data_dir":  dataDir,
		},
		"state": "Leader",
	}
	actNode := node
	if !reflect.DeepEqual(expNode, actNode) {
		t.Errorf("expected content to see %v, saw %v", expNode, actNode)
	}

	// get cluster
	cluster, err := client.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster := map[string]interface{}{
		nodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": bindAddr,
				"grpc_addr": grpcAddr,
				"http_addr": httpAddr,
				"data_dir":  dataDir,
			},
			"state": "Leader",
		},
	}
	actCluster := cluster
	if !reflect.DeepEqual(expCluster, actCluster) {
		t.Errorf("expected content to see %v, saw %v", expCluster, actCluster)
	}

	// get index mapping
	indexMapping1, err := client.GetState("index_config/index_mapping")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexMapping := indexMappingMap
	actIndexMapping := *indexMapping1.(*map[string]interface{})
	if !reflect.DeepEqual(expIndexMapping, actIndexMapping) {
		t.Errorf("expected content to see %v, saw %v", expIndexMapping, actIndexMapping)
	}

	// get index type
	indexType1, err := client.GetState("index_config/index_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexType := indexType
	actIndexType := *indexType1.(*string)
	if expIndexType != actIndexType {
		t.Errorf("expected content to see %v, saw %v", expIndexType, actIndexType)
	}

	// get index storage type
	indexStorageType1, err := client.GetState("index_config/index_storage_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexStorageType := indexStorageType
	actIndexStorageType := *indexStorageType1.(*string)
	if expIndexStorageType != actIndexStorageType {
		t.Errorf("expected content to see %v, saw %v", expIndexStorageType, actIndexStorageType)
	}

	// set value
	err = client.SetState("test/key1", "val1")
	if err != nil {
		t.Errorf("%v", err)
	}
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
	if err == nil {
		t.Errorf("%v", err)
	}
}

func TestManagerCluster(t *testing.T) {
	curDir, _ := os.Getwd()

	// index config
	indexMappingFile := filepath.Join(curDir, "../example/wiki_index_mapping.json")
	indexType := "upside_down"
	indexStorageType := "boltdb"
	indexMapping := mapping.NewIndexMapping()
	if indexMappingFile != "" {
		_, err := os.Stat(indexMappingFile)
		if err == nil {
			// read index mapping file
			f, err := os.Open(indexMappingFile)
			if err != nil {
				t.Errorf("%v", err)
			}
			defer func() {
				_ = f.Close()
			}()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf("%v", err)
			}

			err = json.Unmarshal(b, indexMapping)
			if err != nil {
				t.Errorf("%v", err)
			}
		} else if os.IsNotExist(err) {
			t.Errorf("%v", err)
		}
	}
	err := indexMapping.Validate()
	if err != nil {
		t.Errorf("%v", err)
	}
	indexMappingJSON, err := json.Marshal(indexMapping)
	if err != nil {
		t.Errorf("%v", err)
	}
	var indexMappingMap map[string]interface{}
	err = json.Unmarshal(indexMappingJSON, &indexMappingMap)
	if err != nil {
		t.Errorf("%v", err)
	}
	indexConfig := map[string]interface{}{
		"index_mapping":      indexMappingMap,
		"index_type":         indexType,
		"index_storage_type": indexStorageType,
	}

	// create logger
	logger := logutils.NewLogger("DEBUG", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	// manager1
	manager1NodeId := "manager1"
	manager1BindPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager1BindAddr := fmt.Sprintf(":%d", manager1BindPort)
	manager1GrpcPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager1GrpcAddr := fmt.Sprintf(":%d", manager1GrpcPort)
	manager1HttpPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager1HttpAddr := fmt.Sprintf(":%d", manager1HttpPort)
	manager1DataDir, err := testutils.TmpDir()
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(manager1DataDir)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	manager1PeerAddr := ""
	manager1Metadata := map[string]interface{}{
		"bind_addr": manager1BindAddr,
		"grpc_addr": manager1GrpcAddr,
		"http_addr": manager1HttpAddr,
		"data_dir":  manager1DataDir,
	}
	manager1, err := NewServer(manager1NodeId, manager1Metadata, manager1PeerAddr, indexConfig, logger.Named(manager1NodeId), httpAccessLogger)
	defer func() {
		manager1.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// manager2
	manager2NodeId := "manager2"
	manager2BindPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager2BindAddr := fmt.Sprintf(":%d", manager2BindPort)
	manager2GrpcPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager2GrpcAddr := fmt.Sprintf(":%d", manager2GrpcPort)
	manager2HttpPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager2HttpAddr := fmt.Sprintf(":%d", manager2HttpPort)
	manager2DataDir, err := testutils.TmpDir()
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(manager2DataDir)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	manager2PeerAddr := manager1GrpcAddr
	manager2Metadata := map[string]interface{}{
		"bind_addr": manager2BindAddr,
		"grpc_addr": manager2GrpcAddr,
		"http_addr": manager2HttpAddr,
		"data_dir":  manager2DataDir,
	}
	manager2, err := NewServer(manager2NodeId, manager2Metadata, manager2PeerAddr, nil, logger.Named(manager2NodeId), httpAccessLogger)
	defer func() {
		manager2.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// manager3
	manager3NodeId := "manager3"
	manager3BindPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager3BindAddr := fmt.Sprintf(":%d", manager3BindPort)
	manager3GrpcPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager3GrpcAddr := fmt.Sprintf(":%d", manager3GrpcPort)
	manager3HttpPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	manager3HttpAddr := fmt.Sprintf(":%d", manager3HttpPort)
	manager3DataDir, err := testutils.TmpDir()
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(manager3DataDir)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	manager3PeerAddr := manager1GrpcAddr
	manager3Metadata := map[string]interface{}{
		"bind_addr": manager3BindAddr,
		"grpc_addr": manager3GrpcAddr,
		"http_addr": manager3HttpAddr,
		"data_dir":  manager3DataDir,
	}
	manager3, err := NewServer(manager3NodeId, manager3Metadata, manager3PeerAddr, nil, logger.Named(manager3NodeId), httpAccessLogger)
	defer func() {
		manager3.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// start managers
	manager1.Start()

	time.Sleep(5 * time.Second)

	manager2.Start()

	time.Sleep(5 * time.Second)

	manager3.Start()

	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	client1, err := grpc.NewClient(manager1GrpcAddr)
	defer func() {
		_ = client1.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// gRPC client for manager2
	client2, err := grpc.NewClient(manager2GrpcAddr)
	defer func() {
		_ = client2.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// gRPC client for manager3
	client3, err := grpc.NewClient(manager3GrpcAddr)
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

	// get manager1's node info from manager1
	node1_1, err := client1.GetNode(manager1NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode1_1 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager1BindAddr,
			"grpc_addr": manager1GrpcAddr,
			"http_addr": manager1HttpAddr,
			"data_dir":  manager1DataDir,
		},
		"state": "Leader",
	}
	actNode1_1 := node1_1
	if !reflect.DeepEqual(expNode1_1, actNode1_1) {
		t.Errorf("expected content to see %v, saw %v", expNode1_1, actNode1_1)
	}

	// get manager2's node info from manager1
	node2_1, err := client1.GetNode(manager2NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode2_1 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager2BindAddr,
			"grpc_addr": manager2GrpcAddr,
			"http_addr": manager2HttpAddr,
			"data_dir":  manager2DataDir,
		},
		"state": "Follower",
	}
	actNode2_1 := node2_1
	if !reflect.DeepEqual(expNode2_1, actNode2_1) {
		t.Errorf("expected content to see %v, saw %v", expNode2_1, actNode2_1)
	}

	// get manager3's node info from manager1
	node3_1, err := client1.GetNode(manager3NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode3_1 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager3BindAddr,
			"grpc_addr": manager3GrpcAddr,
			"http_addr": manager3HttpAddr,
			"data_dir":  manager3DataDir,
		},
		"state": "Follower",
	}
	actNode3_1 := node3_1
	if !reflect.DeepEqual(expNode3_1, actNode3_1) {
		t.Errorf("expected content to see %v, saw %v", expNode3_1, actNode3_1)
	}

	// get manager1's node info from manager2
	node1_2, err := client2.GetNode(manager1NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode1_2 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager1BindAddr,
			"grpc_addr": manager1GrpcAddr,
			"http_addr": manager1HttpAddr,
			"data_dir":  manager1DataDir,
		},
		"state": "Leader",
	}
	actNode1_2 := node1_2
	if !reflect.DeepEqual(expNode1_2, actNode1_2) {
		t.Errorf("expected content to see %v, saw %v", expNode1_2, actNode1_2)
	}

	// get manager2's node info from manager2
	node2_2, err := client2.GetNode(manager2NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode2_2 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager2BindAddr,
			"grpc_addr": manager2GrpcAddr,
			"http_addr": manager2HttpAddr,
			"data_dir":  manager2DataDir,
		},
		"state": "Follower",
	}
	actNode2_2 := node2_2
	if !reflect.DeepEqual(expNode2_2, actNode2_2) {
		t.Errorf("expected content to see %v, saw %v", expNode2_2, actNode2_2)
	}

	// get manager3's node info from manager2
	node3_2, err := client2.GetNode(manager3NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode3_2 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager3BindAddr,
			"grpc_addr": manager3GrpcAddr,
			"http_addr": manager3HttpAddr,
			"data_dir":  manager3DataDir,
		},
		"state": "Follower",
	}
	actNode3_2 := node3_2
	if !reflect.DeepEqual(expNode3_2, actNode3_2) {
		t.Errorf("expected content to see %v, saw %v", expNode3_2, actNode3_2)
	}

	// get manager1's node info from manager3
	nodeNode1_3, err := client3.GetNode(manager1NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode1_3 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager1BindAddr,
			"grpc_addr": manager1GrpcAddr,
			"http_addr": manager1HttpAddr,
			"data_dir":  manager1DataDir,
		},
		"state": "Leader",
	}
	actNode1_3 := nodeNode1_3
	if !reflect.DeepEqual(expNode1_3, actNode1_3) {
		t.Errorf("expected content to see %v, saw %v", expNode1_3, actNode1_3)
	}

	// get manager2's node info from manager3
	node2_3, err := client3.GetNode(manager2NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode2_3 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager2BindAddr,
			"grpc_addr": manager2GrpcAddr,
			"http_addr": manager2HttpAddr,
			"data_dir":  manager2DataDir,
		},
		"state": "Follower",
	}
	actNode2_3 := node2_3
	if !reflect.DeepEqual(expNode2_3, actNode2_3) {
		t.Errorf("expected content to see %v, saw %v", expNode2_3, actNode2_3)
	}

	// get manager3's node info from manager3
	node3_3, err := client3.GetNode(manager3NodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	expNode3_3 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": manager3BindAddr,
			"grpc_addr": manager3GrpcAddr,
			"http_addr": manager3HttpAddr,
			"data_dir":  manager3DataDir,
		},
		"state": "Follower",
	}
	actNode3_3 := node3_3
	if !reflect.DeepEqual(expNode3_3, actNode3_3) {
		t.Errorf("expected content to see %v, saw %v", expNode3_3, actNode3_3)
	}

	// get cluster info from manager1
	cluster1, err := client1.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster1 := map[string]interface{}{
		manager1NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager1BindAddr,
				"grpc_addr": manager1GrpcAddr,
				"http_addr": manager1HttpAddr,
				"data_dir":  manager1DataDir,
			},
			"state": "Leader",
		},
		manager2NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager2BindAddr,
				"grpc_addr": manager2GrpcAddr,
				"http_addr": manager2HttpAddr,
				"data_dir":  manager2DataDir,
			},
			"state": "Follower",
		},
		manager3NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager3BindAddr,
				"grpc_addr": manager3GrpcAddr,
				"http_addr": manager3HttpAddr,
				"data_dir":  manager3DataDir,
			},
			"state": "Follower",
		},
	}
	actCluster1 := cluster1
	if !reflect.DeepEqual(expCluster1, actCluster1) {
		t.Errorf("expected content to see %v, saw %v", expCluster1, actCluster1)
	}

	// get cluster info from manager2
	cluster2, err := client2.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster2 := map[string]interface{}{
		manager1NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager1BindAddr,
				"grpc_addr": manager1GrpcAddr,
				"http_addr": manager1HttpAddr,
				"data_dir":  manager1DataDir,
			},
			"state": "Leader",
		},
		manager2NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager2BindAddr,
				"grpc_addr": manager2GrpcAddr,
				"http_addr": manager2HttpAddr,
				"data_dir":  manager2DataDir,
			},
			"state": "Follower",
		},
		manager3NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager3BindAddr,
				"grpc_addr": manager3GrpcAddr,
				"http_addr": manager3HttpAddr,
				"data_dir":  manager3DataDir,
			},
			"state": "Follower",
		},
	}
	actCluster2 := cluster2
	if !reflect.DeepEqual(expCluster2, actCluster2) {
		t.Errorf("expected content to see %v, saw %v", expCluster2, actCluster2)
	}

	// get cluster info from manager3
	cluster3, err := client3.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster3 := map[string]interface{}{
		manager1NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager1BindAddr,
				"grpc_addr": manager1GrpcAddr,
				"http_addr": manager1HttpAddr,
				"data_dir":  manager1DataDir,
			},
			"state": "Leader",
		},
		manager2NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager2BindAddr,
				"grpc_addr": manager2GrpcAddr,
				"http_addr": manager2HttpAddr,
				"data_dir":  manager2DataDir,
			},
			"state": "Follower",
		},
		manager3NodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": manager3BindAddr,
				"grpc_addr": manager3GrpcAddr,
				"http_addr": manager3HttpAddr,
				"data_dir":  manager3DataDir,
			},
			"state": "Follower",
		},
	}
	actCluster3 := cluster3
	if !reflect.DeepEqual(expCluster3, actCluster3) {
		t.Errorf("expected content to see %v, saw %v", expCluster3, actCluster3)
	}

	// get index mapping from manager1
	indexMapping1, err := client1.GetState("index_config/index_mapping")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexMapping1 := indexMappingMap
	actIndexMapping1 := *indexMapping1.(*map[string]interface{})
	if !reflect.DeepEqual(expIndexMapping1, actIndexMapping1) {
		t.Errorf("expected content to see %v, saw %v", expIndexMapping1, actIndexMapping1)
	}

	// get index mapping from manager2
	indexMapping2, err := client2.GetState("index_config/index_mapping")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexMapping2 := indexMappingMap
	actIndexMapping2 := *indexMapping2.(*map[string]interface{})
	if !reflect.DeepEqual(expIndexMapping2, actIndexMapping2) {
		t.Errorf("expected content to see %v, saw %v", expIndexMapping2, actIndexMapping2)
	}

	// get index mapping from manager3
	indexMapping3, err := client3.GetState("index_config/index_mapping")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexMapping3 := indexMappingMap
	actIndexMapping3 := *indexMapping3.(*map[string]interface{})
	if !reflect.DeepEqual(expIndexMapping3, actIndexMapping3) {
		t.Errorf("expected content to see %v, saw %v", expIndexMapping3, actIndexMapping3)
	}

	// get index type from manager1
	indexType1, err := client1.GetState("index_config/index_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexType1 := indexType
	actIndexType1 := *indexType1.(*string)
	if expIndexType1 != actIndexType1 {
		t.Errorf("expected content to see %v, saw %v", expIndexType1, actIndexType1)
	}

	// get index type from manager2
	indexType2, err := client2.GetState("index_config/index_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexType2 := indexType
	actIndexType2 := *indexType2.(*string)
	if expIndexType2 != actIndexType2 {
		t.Errorf("expected content to see %v, saw %v", expIndexType2, actIndexType2)
	}

	// get index type from manager3
	indexType3, err := client2.GetState("index_config/index_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexType3 := indexType
	actIndexType3 := *indexType3.(*string)
	if expIndexType3 != actIndexType3 {
		t.Errorf("expected content to see %v, saw %v", expIndexType3, actIndexType3)
	}

	// get index storage type from manager1
	indexStorageType1, err := client1.GetState("index_config/index_storage_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexStorageType1 := indexStorageType
	actIndexStorageType1 := *indexStorageType1.(*string)
	if expIndexStorageType1 != actIndexStorageType1 {
		t.Errorf("expected content to see %v, saw %v", expIndexStorageType1, actIndexStorageType1)
	}

	// get index storage type from manager2
	indexStorageType2, err := client2.GetState("index_config/index_storage_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexStorageType2 := indexStorageType
	actIndexStorageType2 := *indexStorageType2.(*string)
	if expIndexStorageType2 != actIndexStorageType2 {
		t.Errorf("expected content to see %v, saw %v", expIndexStorageType2, actIndexStorageType2)
	}

	// get index storage type from manager3
	indexStorageType3, err := client3.GetState("index_config/index_storage_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexStorageType3 := indexStorageType
	actIndexStorageType3 := *indexStorageType3.(*string)
	if expIndexStorageType3 != actIndexStorageType3 {
		t.Errorf("expected content to see %v, saw %v", expIndexStorageType3, actIndexStorageType3)
	}

	// set value to manager1
	err = client1.SetState("test/key1", "val1")
	if err != nil {
		t.Errorf("%v", err)
	}

	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from manager1
	val1_1, err := client1.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal1_1 := "val1"
	actVal1_1 := *val1_1.(*string)
	if expVal1_1 != actVal1_1 {
		t.Errorf("expected content to see %v, saw %v", expVal1_1, actVal1_1)
	}

	// get value from manager2
	val1_2, err := client2.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal1_2 := "val1"
	actVal1_2 := *val1_2.(*string)
	if expVal1_2 != actVal1_2 {
		t.Errorf("expected content to see %v, saw %v", expVal1_2, actVal1_2)
	}

	// get value from manager3
	val1_3, err := client3.GetState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal1_3 := "val1"
	actVal1_3 := *val1_3.(*string)
	if expVal1_3 != actVal1_3 {
		t.Errorf("expected content to see %v, saw %v", expVal1_3, actVal1_3)
	}

	// set value to manager2
	err = client2.SetState("test/key2", "val2")
	if err != nil {
		t.Errorf("%v", err)
	}

	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from manager1
	val2_1, err := client1.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal2_1 := "val2"
	actVal2_1 := *val2_1.(*string)
	if expVal2_1 != actVal2_1 {
		t.Errorf("expected content to see %v, saw %v", expVal2_1, actVal2_1)
	}

	// get value from manager2
	val2_2, err := client2.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal2_2 := "val2"
	actVal2_2 := *val2_2.(*string)
	if expVal2_2 != actVal2_2 {
		t.Errorf("expected content to see %v, saw %v", expVal2_2, actVal2_2)
	}

	// get value from manager3
	val2_3, err := client3.GetState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal2_3 := "val2"
	actVal2_3 := *val2_3.(*string)
	if expVal2_3 != actVal2_3 {
		t.Errorf("expected content to see %v, saw %v", expVal2_3, actVal2_3)
	}

	// set value to manager3
	err = client3.SetState("test/key3", "val3")
	if err != nil {
		t.Errorf("%v", err)
	}

	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from manager1
	val3_1, err := client1.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal3_1 := "val3"
	actVal3_1 := *val3_1.(*string)
	if expVal3_1 != actVal3_1 {
		t.Errorf("expected content to see %v, saw %v", expVal3_1, actVal3_1)
	}

	// get value from manager2
	val3_2, err := client2.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal3_2 := "val3"
	actVal3_2 := *val3_2.(*string)
	if expVal3_2 != actVal3_2 {
		t.Errorf("expected content to see %v, saw %v", expVal3_2, actVal3_2)
	}

	// get value from manager3
	val3_3, err := client3.GetState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}
	expVal3_3 := "val3"
	actVal3_3 := *val3_3.(*string)
	if expVal3_3 != actVal3_3 {
		t.Errorf("expected content to see %v, saw %v", expVal3_3, actVal3_3)
	}

	// delete value from manager1
	err = client1.DeleteState("test/key1")
	if err != nil {
		t.Errorf("%v", err)
	}

	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from manager1
	val1_1, err = client1.GetState("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val1_1 != nil {
		t.Errorf("%v", err)
	}

	// get value from manager2
	val1_2, err = client2.GetState("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val1_2 != nil {
		t.Errorf("%v", err)
	}

	// get value from manager3
	val1_3, err = client3.GetState("test/key1")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val1_3 != nil {
		t.Errorf("%v", err)
	}

	// delete value from manager2
	err = client2.DeleteState("test/key2")
	if err != nil {
		t.Errorf("%v", err)
	}

	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from manager1
	val2_1, err = client1.GetState("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val2_1 != nil {
		t.Errorf("%v", err)
	}

	// get value from manager2
	val2_2, err = client2.GetState("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val2_2 != nil {
		t.Errorf("%v", err)
	}

	// get value from manager2
	val2_3, err = client3.GetState("test/key2")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val2_3 != nil {
		t.Errorf("%v", err)
	}

	// delete value from manager3
	err = client3.DeleteState("test/key3")
	if err != nil {
		t.Errorf("%v", err)
	}

	time.Sleep(2 * time.Second) // wait for data to propagate

	// get value from manager1
	val3_1, err = client1.GetState("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val3_1 != nil {
		t.Errorf("%v", err)
	}

	// get value from manager2
	val3_2, err = client2.GetState("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val3_2 != nil {
		t.Errorf("%v", err)
	}

	// get value from manager3
	val3_3, err = client3.GetState("test/key3")
	if err != blasterrors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val3_3 != nil {
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
