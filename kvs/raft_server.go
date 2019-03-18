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

package kvs

import (
	"log"
	"net"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf/kvs"
)

type RaftServer struct {
	BindAddr string
	DataDir  string

	raft *raft.Raft
	fsm  *RaftFSM

	logger *log.Logger
}

func NewRaftServer(bindAddr string, dataDir string, logger *log.Logger) (*RaftServer, error) {
	fsm, err := NewRaftFSM(filepath.Join(dataDir, "kvs"), logger)
	if err != nil {
		return nil, err
	}

	return &RaftServer{
		BindAddr: bindAddr,
		DataDir:  dataDir,
		fsm:      fsm,
		logger:   logger,
	}, nil
}

func (s *RaftServer) Open(bootstrap bool, localID string) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)
	config.SnapshotThreshold = 1024
	config.Logger = s.logger

	addr, err := net.ResolveTCPAddr("tcp", s.BindAddr)
	if err != nil {
		return err
	}

	transport, err := raft.NewTCPTransportWithLogger(s.BindAddr, addr, 3, 10*time.Second, s.logger)
	if err != nil {
		return err
	}

	ss, err := raft.NewFileSnapshotStoreWithLogger(s.DataDir, 2, s.logger)
	if err != nil {
		return err
	}

	// boltDB implement log store and stable store interface
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(s.DataDir, "raft.db"))
	if err != nil {
		return err
	}

	// raft system
	s.raft, err = raft.NewRaft(config, s.fsm, boltDB, boltDB, ss, transport)
	if err != nil {
		return err
	}

	if bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		s.raft.BootstrapCluster(configuration)
	}

	return nil
}

func (s *RaftServer) Close() error {
	err := s.fsm.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) Join(nodeId string, addr string) error {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeId) {
			s.logger.Printf("[INFO] node %s already joined raft cluster", nodeId)
			return nil
		}
	}

	f := s.raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(addr), 0, 0)
	err = f.Error()
	if err != nil {
		return err
	}

	s.logger.Printf("[INFO] node %s at %s joined successfully", nodeId, addr)
	return nil
}

func (s *RaftServer) Leave(nodeId string) error {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return err
	}

	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(nodeId) {
			f := s.raft.RemoveServer(server.ID, 0, 0)
			err = f.Error()
			if err != nil {
				return err
			}

			s.logger.Printf("[INFO] node %s leaved successfully", nodeId)
			return nil
		}
	}

	s.logger.Printf("[INFO] node %s not exists in raft group", nodeId)
	return nil
}

func (s *RaftServer) Snapshot() error {
	f := s.raft.Snapshot()
	err := f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) Get(key []byte) ([]byte, error) {
	value, err := s.fsm.Get(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *RaftServer) Set(key []byte, value []byte) error {
	if s.raft.State() != raft.Leader {
		return errors.ErrNotLeader
	}

	c := &kvs.KVSCommand{
		Op:    "set",
		Key:   key,
		Value: value,
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

func (s *RaftServer) Delete(key []byte) error {
	if s.raft.State() != raft.Leader {
		return errors.ErrNotLeader
	}

	c := &kvs.KVSCommand{
		Op:  "delete",
		Key: key,
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
