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
	"log"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	_ "github.com/mosuka/blast/config"
	"github.com/mosuka/blast/errors"
)

type RaftServer struct {
	id       string
	metadata map[string]interface{}

	bootstrap bool

	raft *raft.Raft
	fsm  *RaftFSM

	indexConfig map[string]interface{}

	logger *log.Logger
	mu     sync.RWMutex
}

func NewRaftServer(id string, metadata map[string]interface{}, bootstrap bool, indexConfig map[string]interface{}, logger *log.Logger) (*RaftServer, error) {
	return &RaftServer{
		id:       id,
		metadata: metadata,

		bootstrap: bootstrap,

		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (s *RaftServer) Start() error {
	var err error

	s.logger.Print("[INFO] create finite state machine")
	s.fsm, err = NewRaftFSM(filepath.Join(s.metadata["data_dir"].(string), "store"), s.logger)
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
	config.LocalID = raft.ServerID(s.id)
	config.SnapshotThreshold = 1024
	config.Logger = s.logger

	addr, err := net.ResolveTCPAddr("tcp", s.metadata["bind_addr"].(string))
	if err != nil {
		return err
	}

	// create transport
	transport, err := raft.NewTCPTransportWithLogger(s.metadata["bind_addr"].(string), addr, 3, 10*time.Second, s.logger)
	if err != nil {
		return err
	}

	// create snapshot store
	snapshotStore, err := raft.NewFileSnapshotStoreWithLogger(s.metadata["data_dir"].(string), 2, s.logger)
	if err != nil {
		return err
	}

	// create raft log store
	raftLogStore, err := raftboltdb.NewBoltStore(filepath.Join(s.metadata["data_dir"].(string), "raft.db"))
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
		err = s.setMetadata(s.id, s.metadata)
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

func (s *RaftServer) Stats() map[string]string {
	return s.raft.Stats()
}

func (s *RaftServer) State() string {
	return s.raft.State().String()
}

func (s *RaftServer) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

func (s *RaftServer) WaitForDetectLeader(timeout time.Duration) error {
	_, err := s.LeaderAddress(timeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) getMetadata(id string) (map[string]interface{}, error) {
	metadata, err := s.fsm.GetMetadata(id)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (s *RaftServer) setMetadata(id string, metadata map[string]interface{}) error {
	msg, err := newMessage(
		setNode,
		map[string]interface{}{
			"id":       id,
			"metadata": metadata,
		},
	)
	if err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) deleteMetadata(id string) error {
	msg, err := newMessage(
		deleteNode,
		map[string]interface{}{
			"id": id,
		},
	)
	if err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) setIndexConfig(indexConfig map[string]interface{}) error {
	err := s.Set("index_config", indexConfig)
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) GetMetadata(id string) (map[string]interface{}, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(id) {
			metadata, err = s.getMetadata(id)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	return metadata, nil
}

func (s *RaftServer) SetMetadata(id string, metadata map[string]interface{}) error {
	if !s.IsLeader() {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		leaderMetadata, err := s.getMetadata(string(leaderId))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		client, err := NewGRPCClient(leaderMetadata["grpc_addr"].(string))
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

		err = client.SetNode(id, metadata)
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
		if server.ID == raft.ServerID(id) {
			s.logger.Printf("[INFO] node %v already joined the cluster", id)
			return nil
		}
	}

	f := s.raft.AddVoter(raft.ServerID(id), raft.ServerAddress(metadata["bind_addr"].(string)), 0, 0)
	err = f.Error()
	if err != nil {
		return err
	}

	// set metadata
	err = s.setMetadata(id, metadata)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
	}

	s.logger.Printf("[INFO] node %v joined successfully", id)
	return nil
}

func (s *RaftServer) DeleteMetadata(id string) error {
	if !s.IsLeader() {
		// forward to leader node
		leaderId, err := s.LeaderID(60 * time.Second)
		if err != nil {
			return err
		}

		leaderMetadata, err := s.getMetadata(string(leaderId))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return nil
		}

		client, err := NewGRPCClient(leaderMetadata["grpc_addr"].(string))
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

		err = client.DeleteNode(id)
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
		if server.ID == raft.ServerID(id) {
			f := s.raft.RemoveServer(server.ID, 0, 0)
			err = f.Error()
			if err != nil {
				return err
			}

			s.logger.Printf("[INFO] node %v leaved successfully", id)
			return nil
		}
	}

	// delete metadata
	err = s.deleteMetadata(id)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
	}

	s.logger.Printf("[INFO] node %v does not exists in the cluster", id)
	return nil
}

func (s *RaftServer) GetServers() (map[string]interface{}, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return nil, err
	}

	servers := map[string]interface{}{}
	for _, server := range cf.Configuration().Servers {
		metadata, err := s.GetMetadata(string(server.ID))
		if err != nil {
			// could not get metadata
			continue
		}

		servers[string(server.ID)] = metadata
	}

	return servers, nil
}

func (s *RaftServer) Snapshot() error {
	f := s.raft.Snapshot()
	err := f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) Get(key string) (interface{}, error) {
	value, err := s.fsm.Get(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (s *RaftServer) Set(key string, value interface{}) error {
	if !s.IsLeader() {
		//// forward to leader node
		//leaderId, err := s.LeaderID(60 * time.Second)
		//if err != nil {
		//	return err
		//}
		//
		//leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return err
		//}
		//
		//client, err := NewGRPCClient(leaderNode.Metadata.GrpcAddr)
		//defer func() {
		//	err := client.Close()
		//	if err != nil {
		//		s.logger.Printf("[ERR] %v", err)
		//	}
		//}()
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return err
		//}
		//
		//err = client.Set(kvp)
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return err
		//}
		//
		//return nil

		return raft.ErrNotLeader
	}

	msg, err := newMessage(
		setKeyValue,
		map[string]interface{}{
			"key":   key,
			"value": value,
		},
	)
	if err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) Delete(key string) error {
	if !s.IsLeader() {
		//// forward to leader node
		//leaderId, err := s.LeaderID(60 * time.Second)
		//if err != nil {
		//	return err
		//}
		//
		//leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return err
		//}
		//
		//client, err := NewGRPCClient(leaderNode.Metadata.GrpcAddr)
		//defer func() {
		//	err := client.Close()
		//	if err != nil {
		//		s.logger.Printf("[ERR] %v", err)
		//	}
		//}()
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return err
		//}
		//
		//err = client.Delete(kvp)
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return err
		//}
		//
		//return nil

		return raft.ErrNotLeader
	}

	msg, err := newMessage(
		deleteKeyValue,
		map[string]interface{}{
			"key": key,
		},
	)
	if err != nil {
		return err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	f := s.raft.Apply(msgBytes, 10*time.Second)
	err = f.Error()
	if err != nil {
		return err
	}

	return nil
}
