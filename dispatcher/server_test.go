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

package dispatcher

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/mosuka/blast/strutils"
	"github.com/mosuka/blast/testutils"
)

func TestServer_Start(t *testing.T) {
	curDir, _ := os.Getwd()

	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	grpcLogger := logutils.NewLogger("WARN", "", 500, 3, 30, false)
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

	managerPeerGrpcAddress1 := ""
	managerGrpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerHttpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerNodeId1 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	managerBindAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerDataDir1 := testutils.TmpDir()
	managerRaftStorageType1 := "boltdb"

	managerNode1 := &management.Node{
		BindAddress: managerBindAddress1,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: managerGrpcAddress1,
			HttpAddress: managerHttpAddress1,
		},
	}

	managerIndexConfig1, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	managerServer1, err := manager.NewServer(managerPeerGrpcAddress1, managerNodeId1, managerNode1, managerDataDir1, managerRaftStorageType1, managerIndexConfig1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if managerServer1 != nil {
			managerServer1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	managerServer1.Start()

	managerPeerGrpcAddress2 := managerGrpcAddress1
	managerGrpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerHttpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerNodeId2 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	managerBindAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerDataDir2 := testutils.TmpDir()
	managerRaftStorageType2 := "boltdb"

	managerNode2 := &management.Node{
		BindAddress: managerBindAddress2,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: managerGrpcAddress2,
			HttpAddress: managerHttpAddress2,
		},
	}

	managerIndexConfig2, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	managerServer2, err := manager.NewServer(managerPeerGrpcAddress2, managerNodeId2, managerNode2, managerDataDir2, managerRaftStorageType2, managerIndexConfig2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if managerServer2 != nil {
			managerServer2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	managerServer2.Start()

	managerPeerGrpcAddress3 := managerGrpcAddress1
	managerGrpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerHttpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerNodeId3 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	managerBindAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	managerDataDir3 := testutils.TmpDir()
	managerRaftStorageType3 := "boltdb"

	managerNode3 := &management.Node{
		BindAddress: managerBindAddress3,
		State:       management.Node_UNKNOWN,
		Metadata: &management.Metadata{
			GrpcAddress: managerGrpcAddress3,
			HttpAddress: managerHttpAddress3,
		},
	}

	managerIndexConfig3, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	// create server
	managerServer3, err := manager.NewServer(managerPeerGrpcAddress3, managerNodeId3, managerNode3, managerDataDir3, managerRaftStorageType3, managerIndexConfig3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		if managerServer3 != nil {
			managerServer3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// start server
	managerServer3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	managerClient1, err := manager.NewGRPCClient(managerNode1.Metadata.GrpcAddress)
	defer func() {
		_ = managerClient1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// get cluster info from manager1
	managerCluster1, err := managerClient1.ClusterInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expManagerCluster1 := &management.Cluster{
		Nodes: map[string]*management.Node{
			managerNodeId1: {
				BindAddress: managerBindAddress1,
				State:       management.Node_LEADER,
				Metadata: &management.Metadata{
					GrpcAddress: managerGrpcAddress1,
					HttpAddress: managerHttpAddress1,
				},
			},
			managerNodeId2: {
				BindAddress: managerBindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress: managerGrpcAddress2,
					HttpAddress: managerHttpAddress2,
				},
			},
			managerNodeId3: {
				BindAddress: managerBindAddress3,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress: managerGrpcAddress3,
					HttpAddress: managerHttpAddress3,
				},
			},
		},
	}
	actManagerCluster1 := managerCluster1
	if !reflect.DeepEqual(expManagerCluster1, actManagerCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expManagerCluster1, actManagerCluster1)
	}

	//
	// indexer cluster1
	//
	// create cluster config
	indexerClusterConfig1 := config.DefaultClusterConfig()
	indexerClusterConfig1.ManagerAddr = managerGrpcAddress1
	indexerClusterConfig1.ClusterId = "cluster1"
	// create node config
	indexerNodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(indexerNodeConfig1.DataDir)
	}()
	indexer1, err := indexer.NewServer(indexerClusterConfig1, indexerNodeConfig1, config.DefaultIndexConfig(), logger.Named("indexer1"), grpcLogger.Named("indexer1"), httpAccessLogger)
	defer func() {
		if indexer1 != nil {
			indexer1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server
	indexer1.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// create node config
	indexerNodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(indexerNodeConfig2.DataDir)
	}()
	indexer2, err := indexer.NewServer(indexerClusterConfig1, indexerNodeConfig2, config.DefaultIndexConfig(), logger.Named("indexer2"), grpcLogger.Named("indexer2"), httpAccessLogger)
	defer func() {
		if indexer2 != nil {
			indexer2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server
	indexer2.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// create node config
	indexerNodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(indexerNodeConfig3.DataDir)
	}()
	indexer3, err := indexer.NewServer(indexerClusterConfig1, indexerNodeConfig3, config.DefaultIndexConfig(), logger.Named("indexer3"), grpcLogger.Named("indexer3"), httpAccessLogger)
	defer func() {
		if indexer3 != nil {
			indexer3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server
	indexer3.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	indexerClient1, err := indexer.NewGRPCClient(indexerNodeConfig1.GRPCAddr)
	defer func() {
		_ = indexerClient1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// get cluster info from manager1
	indexerCluster1, err := indexerClient1.GetCluster()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expIndexerCluster1 := map[string]interface{}{
		indexerNodeConfig1.NodeId: map[string]interface{}{
			"node_config": indexerNodeConfig1.ToMap(),
			"state":       raft.Leader.String(),
		},
		indexerNodeConfig2.NodeId: map[string]interface{}{
			"node_config": indexerNodeConfig2.ToMap(),
			"state":       raft.Follower.String(),
		},
		indexerNodeConfig3.NodeId: map[string]interface{}{
			"node_config": indexerNodeConfig3.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actIndexerCluster1 := indexerCluster1
	expIndexerNodeConfig1 := expIndexerCluster1[indexerNodeConfig1.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actIndexerNodeConfig1 := actIndexerCluster1[indexerNodeConfig1.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expIndexerNodeConfig1, actIndexerNodeConfig1) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerNodeConfig1, actIndexerNodeConfig1)
	}
	actIndexerState1 := actIndexerCluster1[indexerNodeConfig1.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actIndexerState1 && raft.Follower.String() != actIndexerState1 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actIndexerState1)
	}
	expIndexerNodeConfig2 := expIndexerCluster1[indexerNodeConfig2.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actIndexerNodeConfig2 := actIndexerCluster1[indexerNodeConfig2.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expIndexerNodeConfig2, actIndexerNodeConfig2) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerNodeConfig2, actIndexerNodeConfig2)
	}
	actIndexerState2 := actIndexerCluster1[indexerNodeConfig2.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actIndexerState2 && raft.Follower.String() != actIndexerState2 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actIndexerState2)
	}
	expIndexerNodeConfig3 := expIndexerCluster1[indexerNodeConfig3.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actIndexerNodeConfig3 := actIndexerCluster1[indexerNodeConfig3.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expIndexerNodeConfig3, actIndexerNodeConfig3) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerNodeConfig3, actIndexerNodeConfig3)
	}
	actIndexerState3 := actIndexerCluster1[indexerNodeConfig3.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actIndexerState3 && raft.Follower.String() != actIndexerState3 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actIndexerState3)
	}

	//
	// indexer cluster2
	//
	// create cluster config
	indexerClusterConfig2 := config.DefaultClusterConfig()
	indexerClusterConfig2.ManagerAddr = managerGrpcAddress1
	indexerClusterConfig2.ClusterId = "cluster2"
	// create node config
	indexerNodeConfig4 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(indexerNodeConfig4.DataDir)
	}()
	indexer4, err := indexer.NewServer(indexerClusterConfig2, indexerNodeConfig4, config.DefaultIndexConfig(), logger.Named("indexer4"), grpcLogger.Named("indexer4"), httpAccessLogger)
	defer func() {
		if indexer4 != nil {
			indexer4.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server
	indexer4.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// create node config
	indexerNodeConfig5 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(indexerNodeConfig5.DataDir)
	}()
	indexer5, err := indexer.NewServer(indexerClusterConfig2, indexerNodeConfig5, config.DefaultIndexConfig(), logger.Named("indexer5"), grpcLogger.Named("indexer5"), httpAccessLogger)
	defer func() {
		if indexer5 != nil {
			indexer5.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server
	indexer5.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// create node config
	indexerNodeConfig6 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(indexerNodeConfig6.DataDir)
	}()
	indexer6, err := indexer.NewServer(indexerClusterConfig2, indexerNodeConfig6, config.DefaultIndexConfig(), logger.Named("indexer6"), grpcLogger.Named("indexer6"), httpAccessLogger)
	defer func() {
		if indexer6 != nil {
			indexer6.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server
	indexer6.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	indexerClient2, err := indexer.NewGRPCClient(indexerNodeConfig4.GRPCAddr)
	defer func() {
		_ = indexerClient1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// get cluster info from manager1
	indexerCluster2, err := indexerClient2.GetCluster()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expIndexerCluster2 := map[string]interface{}{
		indexerNodeConfig4.NodeId: map[string]interface{}{
			"node_config": indexerNodeConfig4.ToMap(),
			"state":       raft.Leader.String(),
		},
		indexerNodeConfig5.NodeId: map[string]interface{}{
			"node_config": indexerNodeConfig5.ToMap(),
			"state":       raft.Follower.String(),
		},
		indexerNodeConfig6.NodeId: map[string]interface{}{
			"node_config": indexerNodeConfig6.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actIndexerCluster2 := indexerCluster2
	expIndexerNodeConfig4 := expIndexerCluster2[indexerNodeConfig4.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actIndexerNodeConfig4 := actIndexerCluster2[indexerNodeConfig4.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expIndexerNodeConfig4, actIndexerNodeConfig4) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerNodeConfig4, actIndexerNodeConfig4)
	}
	actIndexerState4 := actIndexerCluster2[indexerNodeConfig4.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actIndexerState4 && raft.Follower.String() != actIndexerState4 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actIndexerState4)
	}
	expIndexerNodeConfig5 := expIndexerCluster2[indexerNodeConfig5.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actIndexerNodeConfig5 := actIndexerCluster2[indexerNodeConfig5.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expIndexerNodeConfig5, actIndexerNodeConfig5) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerNodeConfig5, actIndexerNodeConfig5)
	}
	actIndexerState5 := actIndexerCluster2[indexerNodeConfig5.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actIndexerState5 && raft.Follower.String() != actIndexerState5 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actIndexerState5)
	}
	expIndexerNodeConfig6 := expIndexerCluster2[indexerNodeConfig6.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actIndexerNodeConfig6 := actIndexerCluster2[indexerNodeConfig6.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expIndexerNodeConfig6, actIndexerNodeConfig6) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerNodeConfig6, actIndexerNodeConfig6)
	}
	actIndexerState6 := actIndexerCluster2[indexerNodeConfig6.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actIndexerState6 && raft.Follower.String() != actIndexerState6 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actIndexerState6)
	}

	//
	// dispatcher
	//
	// create cluster config
	dispatcherClusterConfig1 := config.DefaultClusterConfig()
	dispatcherClusterConfig1.ManagerAddr = managerGrpcAddress1
	// create node config
	dispatcherNodeConfig := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(dispatcherNodeConfig.DataDir)
	}()
	dispatcher1, err := NewServer(dispatcherClusterConfig1, dispatcherNodeConfig, logger.Named("dispatcher1"), grpcLogger.Named("dispatcher1"), httpAccessLogger)
	defer func() {
		dispatcher1.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start server
	dispatcher1.Start()

	// sleep
	time.Sleep(5 * time.Second)
}
