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
	"fmt"
	"log"
	"net"
	"path/filepath"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	_ "github.com/mosuka/blast/config"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/management"
	blastraft "github.com/mosuka/blast/protobuf/raft"
)

type RaftServer struct {
	managerAddr string
	clusterId   string
	node        *blastraft.Node
	bootstrap   bool

	raft *raft.Raft
	fsm  *RaftFSM

	indexConfig map[string]interface{}

	logger *log.Logger
}

func NewRaftServer(managerAddr string, clusterId string, node *blastraft.Node, bootstrap bool, indexConfig map[string]interface{}, logger *log.Logger) (*RaftServer, error) {
	fsm, err := NewRaftFSM(filepath.Join(node.Metadata.DataDir, "index"), indexConfig, logger)
	if err != nil {
		return nil, err
	}

	return &RaftServer{
		managerAddr: managerAddr,
		clusterId:   clusterId,
		node:        node,
		bootstrap:   bootstrap,
		fsm:         fsm,
		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (s *RaftServer) Start() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.node.Id)
	config.SnapshotThreshold = 1024
	config.Logger = s.logger

	addr, err := net.ResolveTCPAddr("tcp", s.node.Metadata.BindAddr)
	if err != nil {
		return err
	}

	// create transport
	transport, err := raft.NewTCPTransportWithLogger(s.node.Metadata.BindAddr, addr, 3, 10*time.Second, s.logger)
	if err != nil {
		return err
	}

	// create snapshot store
	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger(s.node.Metadata.DataDir, 2, s.logger)
	if err != nil {
		return err
	}

	// create raft log store
	raftLogStore, err := raftboltdb.NewBoltStore(filepath.Join(s.node.Metadata.DataDir, "raft.db"))
	if err != nil {
		return err
	}

	// create raft
	s.raft, err = raft.NewRaft(config, s.fsm, raftLogStore, raftLogStore, snapshotStore, transport)
	if err != nil {
		return err
	}

	if s.bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		s.raft.BootstrapCluster(configuration)

		// wait for detect a leader
		err = s.WaitForDetectLeader(60 * time.Second)
		if err != nil {
			if err == errors.ErrTimeout {
				s.logger.Printf("[WARN] %v", err)
			} else {
				s.logger.Printf("[ERR] %v", err)
				return err
			}
		}

		// set node
		err = s.setNode(s.node)
		if err != nil {
			return err
		}

		// register node to manager
		err = s.updateCluster()
		if err != nil {
			return err
		}
	}

	err = s.fsm.Start()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) Stop() error {
	err := s.fsm.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) WaitForDetectLeader(timeout time.Duration) error {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			leaderAddr := s.raft.Leader()
			if leaderAddr != "" {
				s.logger.Printf("[INFO] detected %v as a leader", leaderAddr)
				return nil
			} else {
				s.logger.Printf("[WARN] %v", errors.ErrNotFoundLeader)
			}
		case <-timer.C:
			return errors.ErrTimeout
		}
	}
}

func (s *RaftServer) LeaderAddress(timeout time.Duration) (raft.ServerAddress, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			leaderAddr := s.raft.Leader()
			if leaderAddr != "" {
				return leaderAddr, nil
			}
		case <-timer.C:
			return "", errors.ErrTimeout
		}
	}
}

func (s *RaftServer) LeaderID(timeout time.Duration) (raft.ServerID, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return "", err
	}

	leaderAddr, err := s.LeaderAddress(timeout)
	if err != nil {
		return "", err
	}

	for _, server := range cf.Configuration().Servers {
		if server.Address == leaderAddr {
			return server.ID, nil
		}
	}

	return "", errors.ErrNotFoundLeader
}

func (s *RaftServer) getNode(node *blastraft.Node) (*blastraft.Node, error) {
	node, err := s.fsm.GetNode(node)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *RaftServer) setNode(node *blastraft.Node) error {
	// Node -> Any
	nodeAny := &any.Any{}
	err := protobuf.UnmarshalAny(node, nodeAny)
	if err != nil {
		return err
	}

	c := &index.IndexCommand{
		Type: index.IndexCommand_SET_NODE,
		Data: nodeAny,
	}

	msg, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msg, 10*time.Second)
	err = f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) deleteNode(nodeId string) error {
	node := &blastraft.Node{
		Id: nodeId,
	}

	// Node -> Any
	nodeAny := &any.Any{}
	err := protobuf.UnmarshalAny(node, nodeAny)
	if err != nil {
		return err
	}

	c := &index.IndexCommand{
		Type: index.IndexCommand_DELETE_NODE,
		Data: nodeAny,
	}

	msg, err := proto.Marshal(c)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msg, 10*time.Second)
	err = f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) updateCluster() error {
	if s.managerAddr != "" {
		mc, err := manager.NewGRPCClient(s.managerAddr)
		defer func() {
			err = mc.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return
			}
		}()
		if err != nil {
			return err
		}

		cluster, err := s.GetCluster()
		if err != nil {
			return err
		}

		clusterAny := &any.Any{}
		err = protobuf.UnmarshalAny(cluster, clusterAny)
		if err != nil {
			return err
		}

		err = mc.Set(
			&management.KeyValuePair{
				Key:   fmt.Sprintf("/cluster_config/clusters/%s", s.clusterId),
				Value: clusterAny,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *RaftServer) Join(node *blastraft.Node) error {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		if err != nil {
			return err
		}

		client, err := NewGRPCClient(leaderNode.Metadata.GrpcAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			return err
		}

		err = client.Join(node)
		if err != nil {
			return err
		}

		return nil
	}

	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(node.Id) {
			s.logger.Printf("[INFO] node %v already joined the cluster", node)
			return nil
		}
	}

	f := s.raft.AddVoter(raft.ServerID(node.Id), raft.ServerAddress(node.Metadata.BindAddr), 0, 0)
	err = f.Error()
	if err != nil {
		return err
	}

	// set node
	err = s.setNode(node)
	if err != nil {
		return err
	}

	// update cluster to manager
	err = s.updateCluster()
	if err != nil {
		return err
	}

	s.logger.Printf("[INFO] node %v joined successfully", node)
	return nil
}

func (s *RaftServer) Leave(node *blastraft.Node) error {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		client, err := NewGRPCClient(leaderNode.Metadata.GrpcAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		err = client.Leave(node)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		return nil
	}

	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(node.Id) {
			f := s.raft.RemoveServer(server.ID, 0, 0)
			err = f.Error()
			if err != nil {
				return err
			}

			s.logger.Printf("[INFO] node %v leaved successfully", node)
			return nil
		}
	}

	// delete metadata
	err = s.deleteNode(node.Id)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
	}

	// unregister node from manager
	err = s.updateCluster()
	if err != nil {
		return err
	}

	s.logger.Printf("[INFO] node %v does not exists in the cluster", node)
	return nil
}

func (s *RaftServer) GetNode() (*blastraft.Node, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return nil, err
	}

	leaderAddr, err := s.LeaderAddress(60 * time.Second)
	if err != nil {
		return nil, err
	}

	node := &blastraft.Node{
		Metadata: &blastraft.Metadata{},
	}
	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(s.node.Id) {
			node.Id = string(server.ID)

			nodeInfo, err := s.getNode(&blastraft.Node{Id: node.Id})
			if err != nil {
				s.logger.Printf("[WARN] %v", err)
				break
			}
			node.Metadata = nodeInfo.Metadata
			node.Metadata.Leader = server.Address == leaderAddr

			break
		}
	}

	return node, nil
}

func (s *RaftServer) GetCluster() (*blastraft.Cluster, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return nil, err
	}

	leaderAddr, err := s.LeaderAddress(60 * time.Second)
	if err != nil {
		return nil, err
	}

	cluster := &blastraft.Cluster{
		Nodes: make(map[string]*blastraft.Metadata, 0),
	}

	for _, server := range cf.Configuration().Servers {
		node := &blastraft.Node{}
		node.Id = string(server.ID)

		nodeInfo, err := s.getNode(&blastraft.Node{Id: node.Id})
		if err != nil {
			s.logger.Printf("[WARN] %v", err)
			continue
		}
		node.Metadata = nodeInfo.Metadata
		node.Metadata.Leader = server.Address == leaderAddr

		cluster.Nodes[node.Id] = node.Metadata
	}

	return cluster, nil
}

func (s *RaftServer) Snapshot() error {
	f := s.raft.Snapshot()
	err := f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) Get(doc *index.Document) (*index.Document, error) {
	fieldsMap, err := s.fsm.Get(doc.Id)
	if err != nil {
		return nil, err
	}

	// map[string]interface{} -> Any
	fieldsAny := &any.Any{}
	err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
	if err != nil {
		return nil, err
	}

	retDoc := &index.Document{
		Id:     doc.Id,
		Fields: fieldsAny,
	}

	return retDoc, nil
}

func (s *RaftServer) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := s.fsm.Search(request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *RaftServer) Index(docs []*index.Document) (*index.UpdateResult, error) {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return nil, err
		}

		leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil, err
		}

		client, err := NewGRPCClient(leaderNode.Metadata.GrpcAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil, err
		}

		result, err := client.Index(docs)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil, err
		}
		s.logger.Printf("[DEBUG] %v", result)

		return result, nil
	}

	count := int32(0)
	for _, doc := range docs {
		// Document -> Any
		docAny := &any.Any{}
		err := protobuf.UnmarshalAny(doc, docAny)
		if err != nil {
			return nil, err
		}

		c := &index.IndexCommand{
			Type: index.IndexCommand_INDEX_DOCUMENT,
			Data: docAny,
		}

		msg, err := proto.Marshal(c)
		if err != nil {
			return nil, err
		}

		f := s.raft.Apply(msg, 10*time.Second)
		err = f.Error()
		if err != nil {
			return nil, err
		}

		count++
	}

	return &index.UpdateResult{
		Count: count,
	}, nil
}

func (s *RaftServer) Delete(docs []*index.Document) (*index.UpdateResult, error) {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return nil, err
		}

		leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil, err
		}

		client, err := NewGRPCClient(leaderNode.Metadata.GrpcAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil, err
		}

		result, err := client.Delete(docs)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil, err
		}
		s.logger.Printf("[DEBUG] %v", result)

		return result, nil
	}

	count := int32(0)
	for _, doc := range docs {
		// Document -> Any
		docAny := &any.Any{}
		err := protobuf.UnmarshalAny(doc, docAny)
		if err != nil {
			return nil, err
		}

		c := &index.IndexCommand{
			Type: index.IndexCommand_DELETE_DOCUMENT,
			Data: docAny,
		}

		msg, err := proto.Marshal(c)
		if err != nil {
			return nil, err
		}

		f := s.raft.Apply(msg, 10*time.Second)
		err = f.Error()
		if err != nil {
			return nil, err
		}

		count++
	}

	return &index.UpdateResult{
		Count: count,
	}, nil
}

func (s *RaftServer) IndexConfig() (*index.IndexConfig, error) {
	configMap, err := s.fsm.IndexConfig()
	if err != nil {
		return nil, err
	}

	// IndexMappingImpl{} -> Any
	indexMappingAny := &any.Any{}
	err = protobuf.UnmarshalAny(configMap["index_mapping"].(map[string]interface{}), indexMappingAny)
	if err != nil {
		return nil, err
	}

	indexConfig := &index.IndexConfig{
		IndexMapping:     indexMappingAny,
		IndexType:        configMap["index_type"].(string),
		IndexStorageType: configMap["index_storage_type"].(string),
	}

	return indexConfig, nil
}

func (s *RaftServer) IndexStats() (*index.IndexStats, error) {
	statsMap, err := s.fsm.IndexStats()
	if err != nil {
		return nil, err
	}

	// map[string]interface{} -> Any
	statsAny := &any.Any{}
	err = protobuf.UnmarshalAny(statsMap, statsAny)
	if err != nil {
		return nil, err
	}

	indexStats := &index.IndexStats{
		Stats: statsAny,
	}

	return indexStats, nil
}