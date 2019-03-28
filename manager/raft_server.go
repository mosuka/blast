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
	"log"
	"net"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	blastraft "github.com/mosuka/blast/protobuf/raft"
)

type RaftServer struct {
	Node      *blastraft.Node
	bootstrap bool

	BindAddr string
	DataDir  string

	raft *raft.Raft
	fsm  *RaftFSM

	logger *log.Logger
}

func NewRaftServer(node *blastraft.Node, bootstrap bool, logger *log.Logger) (*RaftServer, error) {
	fsm, err := NewRaftFSM(filepath.Join(node.DataDir, "kvs"), logger)
	if err != nil {
		return nil, err
	}

	return &RaftServer{
		Node:      node,
		bootstrap: bootstrap,
		fsm:       fsm,
		logger:    logger,
	}, nil
}

func (s *RaftServer) Start() error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.Node.Id)
	config.SnapshotThreshold = 1024
	config.Logger = s.logger

	addr, err := net.ResolveTCPAddr("tcp", s.Node.BindAddr)
	if err != nil {
		return err
	}

	// create transport
	transport, err := raft.NewTCPTransportWithLogger(s.Node.BindAddr, addr, 3, 10*time.Second, s.logger)
	if err != nil {
		return err
	}

	// create snapshot store
	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger(s.Node.DataDir, 2, s.logger)
	if err != nil {
		return err
	}

	// create raft log store
	raftLogStore, err := raftboltdb.NewBoltStore(filepath.Join(s.Node.DataDir, "raft.db"))
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
				return nil
			}
		}

		// set metadata
		err = s.setMetadata(s.Node.Id, s.Node)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}
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

func (s *RaftServer) getMetadata(nodeId string) (*blastraft.Node, error) {
	node, err := s.fsm.GetMetadata(nodeId)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *RaftServer) setMetadata(nodeId string, node *blastraft.Node) error {
	// Node -> Any
	nodeAny := &any.Any{}
	err := protobuf.UnmarshalAny(node, nodeAny)
	if err != nil {
		return err
	}

	c := &management.ManagementCommand{
		Type: management.ManagementCommand_SET_METADATA,
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

func (s *RaftServer) deleteMetadata(nodeId string) error {
	node := &blastraft.Node{
		Id: nodeId,
	}

	// Node -> Any
	nodeAny := &any.Any{}
	err := protobuf.UnmarshalAny(node, nodeAny)
	if err != nil {
		return err
	}

	c := &management.ManagementCommand{
		Type: management.ManagementCommand_DELETE_METADATA,
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

func (s *RaftServer) Join(node *blastraft.Node) error {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		node, err := s.getMetadata(string(leaderId))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		client, err := NewGRPCClient(string(node.GrpcAddr))
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

		err = client.Join(node)
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
			s.logger.Printf("[INFO] node %s already joined the cluster", node.Id)
			return nil
		}
	}

	f := s.raft.AddVoter(raft.ServerID(node.Id), raft.ServerAddress(node.BindAddr), 0, 0)
	err = f.Error()
	if err != nil {
		return err
	}

	// set metadata
	err = s.setMetadata(node.Id, node)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
	}

	s.logger.Printf("[INFO] node %s at %s joined successfully", node.Id, node.BindAddr)
	return nil
}

func (s *RaftServer) Leave(node *blastraft.Node) error {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		node, err := s.getMetadata(string(leaderId))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		client, err := NewGRPCClient(string(node.GrpcAddr))
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

			s.logger.Printf("[INFO] node %s leaved successfully", node.Id)
			return nil
		}
	}

	// delete metadata
	err = s.deleteMetadata(node.Id)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
	}

	s.logger.Printf("[INFO] node %s does not exists in the cluster", node.Id)
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

	node := &blastraft.Node{}
	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(s.Node.Id) {
			node.Id = string(server.ID)
			node.BindAddr = string(server.Address)
			node.Leader = server.Address == leaderAddr

			nodeInfo, err := s.getMetadata(node.Id)
			if err != nil {
				s.logger.Printf("[WARN] %v", err)
				break
			}
			node.GrpcAddr = nodeInfo.GrpcAddr
			node.HttpAddr = nodeInfo.HttpAddr
			node.DataDir = nodeInfo.DataDir
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

	nodes := make([]*blastraft.Node, 0)
	for _, server := range cf.Configuration().Servers {
		node := &blastraft.Node{}
		node.Id = string(server.ID)
		node.BindAddr = string(server.Address)
		node.Leader = server.Address == leaderAddr

		nodeInfo, err := s.getMetadata(node.Id)
		if err != nil {
			s.logger.Printf("[WARN] %v", err)
			continue
		}
		node.GrpcAddr = nodeInfo.GrpcAddr
		node.HttpAddr = nodeInfo.HttpAddr
		node.DataDir = nodeInfo.DataDir

		nodes = append(nodes, node)
	}

	return &blastraft.Cluster{
		Nodes: nodes,
	}, nil
}

func (s *RaftServer) Snapshot() error {
	f := s.raft.Snapshot()
	err := f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) Get(kvp *management.KeyValuePair) (*management.KeyValuePair, error) {
	value, err := s.fsm.Get(kvp.Key)
	if err != nil {
		return nil, err
	}

	s.logger.Printf("[DEBUG] %v", value)

	// map[string]interface{} -> Any
	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny(value, valueAny)
	if err != nil {
		return nil, err
	}

	retKVP := &management.KeyValuePair{
		Key:   kvp.Key,
		Value: valueAny,
	}

	return retKVP, nil
}

func (s *RaftServer) Set(kvp *management.KeyValuePair) error {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		node, err := s.getMetadata(string(leaderId))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return err
		}

		client, err := NewGRPCClient(string(node.GrpcAddr))
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return err
		}

		err = client.Set(kvp)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return err
		}

		return nil
	}

	// KeyValuePair -> Any
	kvpAny := &any.Any{}
	err := protobuf.UnmarshalAny(kvp, kvpAny)
	if err != nil {
		return err
	}

	c := &management.ManagementCommand{
		Type: management.ManagementCommand_PUT_KEY_VALUE_PAIR,
		Data: kvpAny,
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

func (s *RaftServer) Delete(kvp *management.KeyValuePair) error {
	if s.raft.State() != raft.Leader {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		node, err := s.getMetadata(string(leaderId))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return err
		}

		client, err := NewGRPCClient(string(node.GrpcAddr))
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return err
		}

		err = client.Delete(kvp)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return err
		}

		return nil
	}

	// KeyValuePair -> Any
	kvpAny := &any.Any{}
	err := protobuf.UnmarshalAny(kvp, kvpAny)
	if err != nil {
		return err
	}

	c := &management.ManagementCommand{
		Type: management.ManagementCommand_DELETE_KEY_VALUE_PAIR,
		Data: kvpAny,
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
