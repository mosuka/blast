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
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/manager"
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

	// create index config
	indexConfig, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}

	//
	// manager
	//
	// create cluster config
	managerClusterConfig1 := config.DefaultClusterConfig()
	// create node config
	managerNodeConfig1 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(managerNodeConfig1.DataDir)
	}()
	// create manager
	manager1, err := manager.NewServer(managerClusterConfig1, managerNodeConfig1, indexConfig, logger.Named("manager1"), grpcLogger.Named("manager1"), httpAccessLogger)
	defer func() {
		if manager1 != nil {
			manager1.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start manager
	manager1.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// create cluster config
	managerClusterConfig2 := config.DefaultClusterConfig()
	managerClusterConfig2.PeerAddr = managerNodeConfig1.GRPCAddr
	// create node config
	managerNodeConfig2 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(managerNodeConfig2.DataDir)
	}()
	// create manager
	manager2, err := manager.NewServer(managerClusterConfig2, managerNodeConfig2, indexConfig, logger.Named("manager2"), grpcLogger.Named("manager2"), httpAccessLogger)
	defer func() {
		if manager2 != nil {
			manager2.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start manager
	manager2.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// create cluster config
	managerClusterConfig3 := config.DefaultClusterConfig()
	managerClusterConfig3.PeerAddr = managerNodeConfig1.GRPCAddr
	// create node config
	managerNodeConfig3 := testutils.TmpNodeConfig()
	defer func() {
		_ = os.RemoveAll(managerNodeConfig3.DataDir)
	}()
	// create manager
	manager3, err := manager.NewServer(managerClusterConfig3, managerNodeConfig3, indexConfig, logger.Named("manager3"), grpcLogger.Named("manager3"), httpAccessLogger)
	defer func() {
		if manager3 != nil {
			manager3.Stop()
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// start manager
	manager3.Start()
	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	managerClient1, err := grpc.NewClient(managerNodeConfig1.GRPCAddr)
	defer func() {
		_ = managerClient1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// get cluster info from manager1
	managerCluster1, err := managerClient1.GetCluster()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expManagerCluster1 := map[string]interface{}{
		managerNodeConfig1.NodeId: map[string]interface{}{
			"node_config": managerNodeConfig1.ToMap(),
			"state":       raft.Leader.String(),
		},
		managerNodeConfig2.NodeId: map[string]interface{}{
			"node_config": managerNodeConfig2.ToMap(),
			"state":       raft.Follower.String(),
		},
		managerNodeConfig3.NodeId: map[string]interface{}{
			"node_config": managerNodeConfig3.ToMap(),
			"state":       raft.Follower.String(),
		},
	}
	actManagerCluster1 := managerCluster1
	expManagerNodeConfig1 := expManagerCluster1[managerNodeConfig1.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actManagerNodeConfig1 := actManagerCluster1[managerNodeConfig1.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expManagerNodeConfig1, actManagerNodeConfig1) {
		t.Fatalf("expected content to see %v, saw %v", expManagerNodeConfig1, actManagerNodeConfig1)
	}
	actManagerState1 := actManagerCluster1[managerNodeConfig1.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actManagerState1 && raft.Follower.String() != actManagerState1 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actManagerState1)
	}
	expManagerNodeConfig2 := expManagerCluster1[managerNodeConfig2.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actManagerNodeConfig2 := actManagerCluster1[managerNodeConfig2.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expManagerNodeConfig2, actManagerNodeConfig2) {
		t.Fatalf("expected content to see %v, saw %v", expManagerNodeConfig2, actManagerNodeConfig2)
	}
	actManagerState2 := actManagerCluster1[managerNodeConfig2.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actManagerState2 && raft.Follower.String() != actManagerState2 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actManagerState2)
	}
	expManagerNodeConfig3 := expManagerCluster1[managerNodeConfig3.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	actManagerNodeConfig3 := actManagerCluster1[managerNodeConfig3.NodeId].(map[string]interface{})["node_config"].(map[string]interface{})
	if !reflect.DeepEqual(expManagerNodeConfig3, actManagerNodeConfig3) {
		t.Fatalf("expected content to see %v, saw %v", expManagerNodeConfig3, actManagerNodeConfig3)
	}
	actManagerState3 := actManagerCluster1[managerNodeConfig3.NodeId].(map[string]interface{})["state"].(string)
	if raft.Leader.String() != actManagerState3 && raft.Follower.String() != actManagerState3 {
		t.Fatalf("expected content to see %v or %v, saw %v", raft.Leader.String(), raft.Follower.String(), actManagerState3)
	}

	//
	// indexer cluster1
	//
	// create cluster config
	indexerClusterConfig1 := config.DefaultClusterConfig()
	indexerClusterConfig1.ManagerAddr = managerNodeConfig1.GRPCAddr
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
	indexerClient1, err := grpc.NewClient(indexerNodeConfig1.GRPCAddr)
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
	indexerClusterConfig2.ManagerAddr = managerNodeConfig1.GRPCAddr
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
	indexerClient2, err := grpc.NewClient(indexerNodeConfig4.GRPCAddr)
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
	dispatcherClusterConfig1.ManagerAddr = managerNodeConfig1.GRPCAddr
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
