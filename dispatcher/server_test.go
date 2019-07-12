package dispatcher

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/raft"

	"github.com/mosuka/blast/grpc"

	"github.com/mosuka/blast/indexer"

	"github.com/mosuka/blast/manager"

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
	if !reflect.DeepEqual(expManagerCluster1, actManagerCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expManagerCluster1, actManagerCluster1)
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
	if !reflect.DeepEqual(expIndexerCluster1, actIndexerCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerCluster1, actIndexerCluster1)
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
