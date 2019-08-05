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

	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf/index"
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
		Id:          managerNodeId1,
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
	managerServer1, err := manager.NewServer(managerPeerGrpcAddress1, managerNode1, managerDataDir1, managerRaftStorageType1, managerIndexConfig1, logger, grpcLogger, httpAccessLogger)
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
		Id:          managerNodeId2,
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
	managerServer2, err := manager.NewServer(managerPeerGrpcAddress2, managerNode2, managerDataDir2, managerRaftStorageType2, managerIndexConfig2, logger, grpcLogger, httpAccessLogger)
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
		Id:          managerNodeId3,
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
	managerServer3, err := manager.NewServer(managerPeerGrpcAddress3, managerNode3, managerDataDir3, managerRaftStorageType3, managerIndexConfig3, logger, grpcLogger, httpAccessLogger)
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
				Id:          managerNodeId1,
				BindAddress: managerBindAddress1,
				State:       management.Node_LEADER,
				Metadata: &management.Metadata{
					GrpcAddress: managerGrpcAddress1,
					HttpAddress: managerHttpAddress1,
				},
			},
			managerNodeId2: {
				Id:          managerNodeId2,
				BindAddress: managerBindAddress2,
				State:       management.Node_FOLLOWER,
				Metadata: &management.Metadata{
					GrpcAddress: managerGrpcAddress2,
					HttpAddress: managerHttpAddress2,
				},
			},
			managerNodeId3: {
				Id:          managerNodeId3,
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
	indexerManagerGrpcAddress1 := managerGrpcAddress1
	indexerShardId1 := "shard-1"
	indexerPeerGrpcAddress1 := ""
	indexerGrpcAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerHttpAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerNodeId1 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	indexerBindAddress1 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerDataDir1 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(indexerDataDir1)
	}()
	indexerRaftStorageType1 := "boltdb"

	indexerNode1 := &index.Node{
		Id:          indexerNodeId1,
		BindAddress: indexerBindAddress1,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: indexerGrpcAddress1,
			HttpAddress: indexerHttpAddress1,
		},
	}
	indexConfig1, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer1, err := indexer.NewServer(indexerManagerGrpcAddress1, indexerShardId1, indexerPeerGrpcAddress1, indexerNode1, indexerDataDir1, indexerRaftStorageType1, indexConfig1, logger, grpcLogger, httpAccessLogger)
	defer func() {
		indexerServer1.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer1.Start()

	// sleep
	time.Sleep(5 * time.Second)

	indexerManagerGrpcAddress2 := managerGrpcAddress1
	indexerShardId2 := "shard-1"
	indexerPeerGrpcAddress2 := ""
	indexerGrpcAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerHttpAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerNodeId2 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	indexerBindAddress2 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerDataDir2 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(indexerDataDir2)
	}()
	indexerRaftStorageType2 := "boltdb"

	indexerNode2 := &index.Node{
		Id:          indexerNodeId2,
		BindAddress: indexerBindAddress2,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: indexerGrpcAddress2,
			HttpAddress: indexerHttpAddress2,
		},
	}
	indexConfig2, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer2, err := indexer.NewServer(indexerManagerGrpcAddress2, indexerShardId2, indexerPeerGrpcAddress2, indexerNode2, indexerDataDir2, indexerRaftStorageType2, indexConfig2, logger, grpcLogger, httpAccessLogger)
	defer func() {
		indexerServer2.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer2.Start()

	// sleep
	time.Sleep(5 * time.Second)

	indexerManagerGrpcAddress3 := managerGrpcAddress1
	indexerShardId3 := "shard-1"
	indexerPeerGrpcAddress3 := ""
	indexerGrpcAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerHttpAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerNodeId3 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	indexerBindAddress3 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerDataDir3 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(indexerDataDir3)
	}()
	indexerRaftStorageType3 := "boltdb"

	indexerNode3 := &index.Node{
		Id:          indexerNodeId3,
		BindAddress: indexerBindAddress3,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: indexerGrpcAddress3,
			HttpAddress: indexerHttpAddress3,
		},
	}
	indexConfig3, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer3, err := indexer.NewServer(indexerManagerGrpcAddress3, indexerShardId3, indexerPeerGrpcAddress3, indexerNode3, indexerDataDir3, indexerRaftStorageType3, indexConfig3, logger, grpcLogger, httpAccessLogger)
	defer func() {
		indexerServer3.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer3.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	indexerClient1, err := indexer.NewGRPCClient(indexerNode1.Metadata.GrpcAddress)
	defer func() {
		_ = indexerClient1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// get cluster info from manager1
	indexerCluster1, err := indexerClient1.ClusterInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expIndexerCluster1 := &index.Cluster{
		Nodes: map[string]*index.Node{
			indexerNodeId1: {
				Id:          indexerNodeId1,
				BindAddress: indexerBindAddress1,
				State:       index.Node_LEADER,
				Metadata: &index.Metadata{
					GrpcAddress: indexerGrpcAddress1,
					HttpAddress: indexerHttpAddress1,
				},
			},
			indexerNodeId2: {
				Id:          indexerNodeId2,
				BindAddress: indexerBindAddress2,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: indexerGrpcAddress2,
					HttpAddress: indexerHttpAddress2,
				},
			},
			indexerNodeId3: {
				Id:          indexerNodeId3,
				BindAddress: indexerBindAddress3,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: indexerGrpcAddress3,
					HttpAddress: indexerHttpAddress3,
				},
			},
		},
	}
	actIndexerCluster1 := indexerCluster1
	if !reflect.DeepEqual(expIndexerCluster1, actIndexerCluster1) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerCluster1, actIndexerCluster1)
	}

	//
	// indexer cluster2
	//
	indexerManagerGrpcAddress4 := managerGrpcAddress1
	indexerShardId4 := "shard-2"
	indexerPeerGrpcAddress4 := ""
	indexerGrpcAddress4 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerHttpAddress4 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerNodeId4 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	indexerBindAddress4 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerDataDir4 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(indexerDataDir4)
	}()
	indexerRaftStorageType4 := "boltdb"

	indexerNode4 := &index.Node{
		Id:          indexerNodeId4,
		BindAddress: indexerBindAddress4,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: indexerGrpcAddress4,
			HttpAddress: indexerHttpAddress4,
		},
	}
	indexConfig4, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer4, err := indexer.NewServer(indexerManagerGrpcAddress4, indexerShardId4, indexerPeerGrpcAddress4, indexerNode4, indexerDataDir4, indexerRaftStorageType4, indexConfig4, logger, grpcLogger, httpAccessLogger)
	defer func() {
		indexerServer4.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer4.Start()

	// sleep
	time.Sleep(5 * time.Second)

	indexerManagerGrpcAddress5 := managerGrpcAddress1
	indexerShardId5 := "shard-2"
	indexerPeerGrpcAddress5 := ""
	indexerGrpcAddress5 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerHttpAddress5 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerNodeId5 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	indexerBindAddress5 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerDataDir5 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(indexerDataDir5)
	}()
	indexerRaftStorageType5 := "boltdb"

	indexerNode5 := &index.Node{
		Id:          indexerNodeId5,
		BindAddress: indexerBindAddress5,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: indexerGrpcAddress5,
			HttpAddress: indexerHttpAddress5,
		},
	}
	indexConfig5, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer5, err := indexer.NewServer(indexerManagerGrpcAddress5, indexerShardId5, indexerPeerGrpcAddress5, indexerNode5, indexerDataDir5, indexerRaftStorageType5, indexConfig5, logger, grpcLogger, httpAccessLogger)
	defer func() {
		indexerServer5.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer5.Start()

	// sleep
	time.Sleep(5 * time.Second)

	indexerManagerGrpcAddress6 := managerGrpcAddress1
	indexerShardId6 := "shard-2"
	indexerPeerGrpcAddress6 := ""
	indexerGrpcAddress6 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerHttpAddress6 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerNodeId6 := fmt.Sprintf("node-%s", strutils.RandStr(5))
	indexerBindAddress6 := fmt.Sprintf(":%d", testutils.TmpPort())
	indexerDataDir6 := testutils.TmpDir()
	defer func() {
		_ = os.RemoveAll(indexerDataDir6)
	}()
	indexerRaftStorageType6 := "boltdb"

	indexerNode6 := &index.Node{
		Id:          indexerNodeId6,
		BindAddress: indexerBindAddress6,
		State:       index.Node_UNKNOWN,
		Metadata: &index.Metadata{
			GrpcAddress: indexerGrpcAddress6,
			HttpAddress: indexerHttpAddress6,
		},
	}
	indexConfig6, err := testutils.TmpIndexConfig(filepath.Join(curDir, "../example/wiki_index_mapping.json"), "upside_down", "boltdb")
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer6, err := indexer.NewServer(indexerManagerGrpcAddress6, indexerShardId6, indexerPeerGrpcAddress6, indexerNode6, indexerDataDir6, indexerRaftStorageType6, indexConfig6, logger, grpcLogger, httpAccessLogger)
	defer func() {
		indexerServer6.Stop()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	indexerServer6.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// gRPC client for manager1
	indexerClient2, err := indexer.NewGRPCClient(indexerNode4.Metadata.GrpcAddress)
	defer func() {
		_ = indexerClient1.Close()
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}
	// get cluster info from manager1
	indexerCluster2, err := indexerClient2.ClusterInfo()
	if err != nil {
		t.Fatalf("%v", err)
	}
	expIndexerCluster2 := &index.Cluster{
		Nodes: map[string]*index.Node{
			indexerNodeId4: {
				Id:          indexerNodeId4,
				BindAddress: indexerBindAddress4,
				State:       index.Node_LEADER,
				Metadata: &index.Metadata{
					GrpcAddress: indexerGrpcAddress4,
					HttpAddress: indexerHttpAddress4,
				},
			},
			indexerNodeId5: {
				Id:          indexerNodeId5,
				BindAddress: indexerBindAddress5,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: indexerGrpcAddress5,
					HttpAddress: indexerHttpAddress5,
				},
			},
			indexerNodeId6: {
				Id:          indexerNodeId6,
				BindAddress: indexerBindAddress6,
				State:       index.Node_FOLLOWER,
				Metadata: &index.Metadata{
					GrpcAddress: indexerGrpcAddress6,
					HttpAddress: indexerHttpAddress6,
				},
			},
		},
	}
	actIndexerCluster2 := indexerCluster2
	if !reflect.DeepEqual(expIndexerCluster2, actIndexerCluster2) {
		t.Fatalf("expected content to see %v, saw %v", expIndexerCluster2, actIndexerCluster2)
	}

	////
	//// dispatcher
	////
	//dispatcherManagerGrpcAddress := managerGrpcAddress1
	//dispatcherGrpcAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	//dispatcherHttpAddress := fmt.Sprintf(":%d", testutils.TmpPort())
	//
	//dispatcher1, err := NewServer(dispatcherManagerGrpcAddress, dispatcherGrpcAddress, dispatcherHttpAddress, logger.Named("dispatcher1"), grpcLogger.Named("dispatcher1"), httpAccessLogger)
	//defer func() {
	//	dispatcher1.Stop()
	//}()
	//if err != nil {
	//	t.Fatalf("%v", err)
	//}
	//// start server
	//dispatcher1.Start()
	//
	//// sleep
	//time.Sleep(5 * time.Second)
}
