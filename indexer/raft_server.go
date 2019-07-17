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
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	raftbadgerdb "github.com/markthethomas/raft-badger"
	_ "github.com/mosuka/blast/builtins"
	"github.com/mosuka/blast/config"
	blasterrors "github.com/mosuka/blast/errors"
	"go.uber.org/zap"
	//raftmdb "github.com/hashicorp/raft-mdb"
)

type RaftServer struct {
	nodeConfig  *config.NodeConfig
	indexConfig *config.IndexConfig
	bootstrap   bool
	logger      *zap.Logger

	raft *raft.Raft
	fsm  *RaftFSM
}

func NewRaftServer(nodeConfig *config.NodeConfig, indexConfig *config.IndexConfig, bootstrap bool, logger *zap.Logger) (*RaftServer, error) {
	return &RaftServer{
		nodeConfig:  nodeConfig,
		indexConfig: indexConfig,
		bootstrap:   bootstrap,
		logger:      logger,
	}, nil
}

func (s *RaftServer) Start() error {
	var err error

	fsmPath := filepath.Join(s.nodeConfig.DataDir, "index")
	s.logger.Info("create finite state machine", zap.String("path", fsmPath))
	s.fsm, err = NewRaftFSM(fsmPath, s.indexConfig, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	s.logger.Info("start finite state machine")
	err = s.fsm.Start()
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	s.logger.Info("create Raft config", zap.String("node_id", s.nodeConfig.NodeId))
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(s.nodeConfig.NodeId)
	raftConfig.SnapshotThreshold = 1024
	raftConfig.LogOutput = ioutil.Discard

	s.logger.Info("resolve TCP address", zap.String("bind_addr", s.nodeConfig.BindAddr))
	addr, err := net.ResolveTCPAddr("tcp", s.nodeConfig.BindAddr)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	s.logger.Info("create TCP transport", zap.String("bind_addr", s.nodeConfig.BindAddr))
	transport, err := raft.NewTCPTransport(s.nodeConfig.BindAddr, addr, 3, 10*time.Second, ioutil.Discard)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	snapshotPath := s.nodeConfig.DataDir
	s.logger.Info("create snapshot store", zap.String("path", snapshotPath))
	snapshotStore, err := raft.NewFileSnapshotStore(snapshotPath, 2, ioutil.Discard)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	s.logger.Info("create Raft machine")
	var logStore raft.LogStore
	var stableStore raft.StableStore
	switch s.nodeConfig.RaftStorageType {
	case "boltdb":
		logStorePath := filepath.Join(s.nodeConfig.DataDir, "raft", "log", "boltdb.db")
		s.logger.Info("create raft log store", zap.String("path", logStorePath), zap.String("raft_storage_type", s.nodeConfig.RaftStorageType))
		err = os.MkdirAll(filepath.Dir(logStorePath), 0755)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
		logStore, err = raftboltdb.NewBoltStore(logStorePath)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
		stableStorePath := filepath.Join(s.nodeConfig.DataDir, "raft", "stable", "boltdb.db")
		s.logger.Info("create raft stable store", zap.String("path", stableStorePath), zap.String("raft_storage_type", s.nodeConfig.RaftStorageType))
		err = os.MkdirAll(filepath.Dir(stableStorePath), 0755)
		stableStore, err = raftboltdb.NewBoltStore(stableStorePath)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
	case "badger":
		logStorePath := filepath.Join(s.nodeConfig.DataDir, "raft", "log")
		s.logger.Info("create raft log store", zap.String("path", logStorePath), zap.String("raft_storage_type", s.nodeConfig.RaftStorageType))
		err = os.MkdirAll(filepath.Join(logStorePath, "badger"), 0755)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
		logStore, err = raftbadgerdb.NewBadgerStore(logStorePath)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
		stableStorePath := filepath.Join(s.nodeConfig.DataDir, "raft", "stable")
		s.logger.Info("create raft stable store", zap.String("path", stableStorePath), zap.String("raft_storage_type", s.nodeConfig.RaftStorageType))
		err = os.MkdirAll(filepath.Join(stableStorePath, "badger"), 0755)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
		stableStore, err = raftbadgerdb.NewBadgerStore(stableStorePath)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
	default:
		logStorePath := filepath.Join(s.nodeConfig.DataDir, "raft", "log", "boltdb.db")
		s.logger.Info("create raft log store", zap.String("path", logStorePath), zap.String("raft_storage_type", s.nodeConfig.RaftStorageType))
		err = os.MkdirAll(filepath.Dir(logStorePath), 0755)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
		logStore, err = raftboltdb.NewBoltStore(logStorePath)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
		stableStorePath := filepath.Join(s.nodeConfig.DataDir, "raft", "stable", "boltdb.db")
		s.logger.Info("create raft stable store", zap.String("path", stableStorePath), zap.String("raft_storage_type", s.nodeConfig.RaftStorageType))
		err = os.MkdirAll(filepath.Dir(stableStorePath), 0755)
		stableStore, err = raftboltdb.NewBoltStore(stableStorePath)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}
	}

	s.logger.Info("create Raft machine")
	s.raft, err = raft.NewRaft(raftConfig, s.fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		s.logger.Fatal(err.Error())
		return err
	}

	if s.bootstrap {
		s.logger.Info("configure Raft machine as bootstrap")
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raftConfig.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		s.raft.BootstrapCluster(configuration)

		s.logger.Info("wait for become a leader")
		err = s.WaitForDetectLeader(60 * time.Second)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}

		// set node config
		s.logger.Info("register its own information", zap.String("node_id", s.nodeConfig.NodeId), zap.Any("node_config", s.nodeConfig))
		err = s.setNodeConfig(s.nodeConfig.NodeId, s.nodeConfig.ToMap())
		if err != nil {
			s.logger.Fatal(err.Error())
			return nil
		}
	}

	return nil
}

func (s *RaftServer) Stop() error {
	s.logger.Info("shutdown Raft machine")
	f := s.raft.Shutdown()
	err := f.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	s.logger.Info("stop finite state machine")
	err = s.fsm.Stop()
	if err != nil {
		s.logger.Error(err.Error())
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
				s.logger.Debug("detect a leader", zap.String("address", string(leaderAddr)))
				return leaderAddr, nil
			}
		case <-timer.C:
			s.logger.Error("timeout exceeded")
			return "", blasterrors.ErrTimeout
		}
	}
}

func (s *RaftServer) LeaderID(timeout time.Duration) (raft.ServerID, error) {
	leaderAddr, err := s.LeaderAddress(timeout)
	if err != nil {
		s.logger.Error(err.Error())
		return "", err
	}

	cf := s.raft.GetConfiguration()
	err = cf.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return "", err
	}

	for _, server := range cf.Configuration().Servers {
		if server.Address == leaderAddr {
			return server.ID, nil
		}
	}

	s.logger.Error(blasterrors.ErrNotFoundLeader.Error())
	return "", blasterrors.ErrNotFoundLeader
}

func (s *RaftServer) NodeID() string {
	return s.nodeConfig.NodeId
}

func (s *RaftServer) Stats() map[string]string {
	return s.raft.Stats()
}

func (s *RaftServer) State() raft.RaftState {
	return s.raft.State()
}

func (s *RaftServer) IsLeader() bool {
	return s.State() == raft.Leader
}

func (s *RaftServer) WaitForDetectLeader(timeout time.Duration) error {
	_, err := s.LeaderAddress(timeout)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

func (s *RaftServer) getNodeConfig(nodeId string) (map[string]interface{}, error) {
	nodeConfig, err := s.fsm.GetNodeConfig(nodeId)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return nodeConfig, nil
}

func (s *RaftServer) setNodeConfig(nodeId string, nodeConfig map[string]interface{}) error {
	msg, err := newMessage(
		setNode,
		map[string]interface{}{
			"node_id":     nodeId,
			"node_config": nodeConfig,
		},
	)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	err = f.Response().(*fsmResponse).error
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

func (s *RaftServer) deleteNodeConfig(nodeId string) error {
	msg, err := newMessage(
		deleteNode,
		map[string]interface{}{
			"node_id": nodeId,
		},
	)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}
	err = f.Response().(*fsmResponse).error
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

func (s *RaftServer) GetNode(id string) (map[string]interface{}, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	node := make(map[string]interface{}, 0)
	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(id) {
			nodeConfig, err := s.getNodeConfig(id)
			if err != nil {
				s.logger.Error(err.Error())
				return nil, err
			}
			node["node_config"] = nodeConfig
			break
		}
	}

	return node, nil
}

func (s *RaftServer) SetNode(nodeId string, nodeConfig map[string]interface{}) error {
	if !s.IsLeader() {
		s.logger.Warn(raft.ErrNotLeader.Error(), zap.String("state", s.raft.State().String()))
		return raft.ErrNotLeader
	}

	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeId) {
			s.logger.Info("node already joined the cluster", zap.String("id", nodeId))
			return nil
		}
	}

	bindAddr, ok := nodeConfig["bind_addr"].(string)
	if !ok {
		s.logger.Error("missing metadata", zap.String("bind_addr", bindAddr))
		return errors.New("missing metadata")
	}

	// add node to Raft cluster
	s.logger.Info("add voter", zap.String("nodeId", nodeId), zap.String("address", bindAddr))
	f := s.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(bindAddr), 0, 0)
	err = f.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	// set node config
	err = s.setNodeConfig(nodeId, nodeConfig)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

func (s *RaftServer) DeleteNode(nodeId string) error {
	if !s.IsLeader() {
		s.logger.Warn(raft.ErrNotLeader.Error(), zap.String("state", s.raft.State().String()))
		return raft.ErrNotLeader
	}

	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	// delete node from Raft cluster
	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeId) {
			s.logger.Debug("remove server", zap.String("node_id", nodeId))
			f := s.raft.RemoveServer(server.ID, 0, 0)
			err = f.Error()
			if err != nil {
				s.logger.Error(err.Error())
				return err
			}
		}
	}

	// delete node config
	err = s.deleteNodeConfig(nodeId)
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

func (s *RaftServer) GetCluster() (map[string]interface{}, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	cluster := map[string]interface{}{}
	for _, server := range cf.Configuration().Servers {
		node, err := s.GetNode(string(server.ID))
		if err != nil {
			s.logger.Warn(err.Error())
			node = map[string]interface{}{}
		}
		cluster[string(server.ID)] = node
	}

	return cluster, nil
}

func (s *RaftServer) Snapshot() error {
	f := s.raft.Snapshot()
	err := f.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return err
	}

	return nil
}

func (s *RaftServer) GetDocument(id string) (map[string]interface{}, error) {
	fields, err := s.fsm.GetDocument(id)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return fields, nil
}

func (s *RaftServer) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := s.fsm.Search(request)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return result, nil
}

func (s *RaftServer) IndexDocument(docs []map[string]interface{}) (int, error) {
	if !s.IsLeader() {
		s.logger.Error(raft.ErrNotLeader.Error(), zap.String("state", s.raft.State().String()))
		return -1, raft.ErrNotLeader
	}

	msg, err := newMessage(
		indexDocument,
		docs,
	)
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}
	err = f.Response().(*fsmIndexDocumentResponse).error
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}

	return f.Response().(*fsmIndexDocumentResponse).count, nil
}

func (s *RaftServer) DeleteDocument(ids []string) (int, error) {
	if !s.IsLeader() {
		s.logger.Error(raft.ErrNotLeader.Error(), zap.String("state", s.raft.State().String()))
		return -1, raft.ErrNotLeader
	}

	msg, err := newMessage(
		deleteDocument,
		ids,
	)
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}
	err = f.Response().(*fsmDeleteDocumentResponse).error
	if err != nil {
		s.logger.Error(err.Error())
		return -1, err
	}

	return f.Response().(*fsmDeleteDocumentResponse).count, nil
}

func (s *RaftServer) GetIndexConfig() (map[string]interface{}, error) {
	indexConfig, err := s.fsm.GetIndexConfig()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return indexConfig, nil
}

func (s *RaftServer) GetIndexStats() (map[string]interface{}, error) {
	indexStats, err := s.fsm.GetIndexStats()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return indexStats, nil
}
