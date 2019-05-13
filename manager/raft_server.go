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
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	_ "github.com/mosuka/blast/config"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	blastraft "github.com/mosuka/blast/protobuf/raft"
)

type RaftServer struct {
	node      *blastraft.Node
	bootstrap bool

	path string

	raft *raft.Raft
	fsm  *RaftFSM

	indexConfig map[string]interface{}

	logger *log.Logger
	mu     sync.RWMutex
}

func NewRaftServer(node *blastraft.Node, bootstrap bool, indexConfig map[string]interface{}, logger *log.Logger) (*RaftServer, error) {
	return &RaftServer{
		node:        node,
		bootstrap:   bootstrap,
		path:        node.Metadata.DataDir,
		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (s *RaftServer) Start() error {
	var err error

	s.logger.Print("[INFO] create finite state machine")
	s.fsm, err = NewRaftFSM(filepath.Join(s.path, "store"), s.logger)
	if err != nil {
		return err
	}

	s.logger.Print("[INFO] start finite state machine")
	err = s.fsm.Start()
	if err != nil {
		return err
	}

	s.logger.Print("[INFO] initialize Raft")
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
	s.logger.Print("[INFO] start Raft")
	s.raft, err = raft.NewRaft(config, s.fsm, raftLogStore, raftLogStore, snapshotStore, transport)
	if err != nil {
		return err
	}

	if s.bootstrap {
		s.logger.Print("[INFO] configure Raft as bootstrap")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		s.raft.BootstrapCluster(configuration)

		// wait for become a leader
		s.logger.Print("[INFO] wait for become a leader")
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
		s.logger.Print("[INFO] register itself in a cluster")
		err = s.setNode(s.node)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		// set index config
		s.logger.Print("[INFO] register index config")
		err = s.setIndexConfig(s.indexConfig)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}
	}

	return nil
}

func (s *RaftServer) Stop() error {
	s.logger.Print("[INFO] shutdown Raft")
	f := s.raft.Shutdown()
	err := f.Error()
	if err != nil {
		return err
	}

	s.logger.Print("[INFO] stop finite state machine")
	err = s.fsm.Stop()
	if err != nil {
		return err
	}

	return nil
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
				s.logger.Printf("[DEBUG] leader address detected: %v", leaderAddr)
				return leaderAddr, nil
			}
		case <-timer.C:
			return "", errors.ErrTimeout
		}
	}
}

func (s *RaftServer) LeaderID(timeout time.Duration) (raft.ServerID, error) {
	leaderAddr, err := s.LeaderAddress(timeout)
	if err != nil {
		return "", err
	}

	cf := s.raft.GetConfiguration()
	err = cf.Error()
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

func (s *RaftServer) WaitForDetectLeader(timeout time.Duration) error {
	_, err := s.LeaderAddress(timeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) getNode(node *blastraft.Node) (*blastraft.Node, error) {
	meta, err := s.fsm.GetNode(node)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func (s *RaftServer) setNode(node *blastraft.Node) error {
	// Node -> Any
	nodeAny := &any.Any{}
	err := protobuf.UnmarshalAny(node, nodeAny)
	if err != nil {
		return err
	}

	c := &management.ManagementCommand{
		Type: management.ManagementCommand_SET_NODE,
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

func (s *RaftServer) deleteNode(node *blastraft.Node) error {
	// Node -> Any
	nodeAny := &any.Any{}
	err := protobuf.UnmarshalAny(node, nodeAny)
	if err != nil {
		return err
	}

	c := &management.ManagementCommand{
		Type: management.ManagementCommand_DELETE_NODE,
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

func (s *RaftServer) setIndexConfig(indexConfig map[string]interface{}) error {
	indexConfigAny := &any.Any{}
	err := protobuf.UnmarshalAny(indexConfig, indexConfigAny)
	if err != nil {
		return err
	}

	kvp := &management.KeyValuePair{
		Key:   "/index_config",
		Value: indexConfigAny,
	}

	err = s.Set(kvp)
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
			s.logger.Printf("[INFO] node %v already joined the cluster", node)
			return nil
		}
	}

	f := s.raft.AddVoter(raft.ServerID(node.Id), raft.ServerAddress(node.Metadata.BindAddr), 0, 0)
	err = f.Error()
	if err != nil {
		return err
	}

	// set metadata
	err = s.setNode(node)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
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
	err = s.deleteNode(node)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
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

func (s *RaftServer) Get(kvp *management.KeyValuePair) (*management.KeyValuePair, error) {
	value, err := s.fsm.Get(kvp.Key)
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, errors.ErrNotFound
	}

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

		leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
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

		leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
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
