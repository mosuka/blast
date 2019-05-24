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
	"fmt"
	"log"
	"net"
	"path/filepath"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	_ "github.com/mosuka/blast/config"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/manager"
)

type RaftServer struct {
	managerAddr string
	clusterId   string

	nodeId   string
	metadata map[string]interface{}

	bootstrap bool

	raft *raft.Raft
	fsm  *RaftFSM

	indexConfig map[string]interface{}

	logger *log.Logger
}

func NewRaftServer(managerAddr string, clusterId string, nodeId string, metadata map[string]interface{}, bootstrap bool, indexConfig map[string]interface{}, logger *log.Logger) (*RaftServer, error) {
	return &RaftServer{
		managerAddr: managerAddr,
		clusterId:   clusterId,
		nodeId:      nodeId,
		metadata:    metadata,
		bootstrap:   bootstrap,
		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (s *RaftServer) Start() error {
	var err error

	s.logger.Print("[INFO] create finite state machine")
	s.fsm, err = NewRaftFSM(filepath.Join(s.metadata["data_dir"].(string), "index"), s.indexConfig, s.logger)
	if err != nil {
		return err
	}

	s.logger.Print("[INFO] start finite state machine")
	err = s.fsm.Start()
	if err != nil {
		return err
	}

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.nodeId)
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
		err = s.setNode(s.nodeId, s.metadata)
		if err != nil {
			return err
		}

		// register node to manager
		err = s.updateCluster()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *RaftServer) Stop() error {
	err := s.fsm.Stop()
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

func (s *RaftServer) WaitForDetectLeader(timeout time.Duration) error {
	_, err := s.LeaderAddress(timeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *RaftServer) getNode(id string) (map[string]interface{}, error) {
	metadata, err := s.fsm.GetNode(id)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (s *RaftServer) setNode(id string, metadata map[string]interface{}) error {
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

func (s *RaftServer) deleteNode(id string) error {
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

		err = mc.Set(fmt.Sprintf("/cluster_config/clusters/%s", s.clusterId), cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *RaftServer) GetNode(id string) (map[string]interface{}, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return nil, err
	}

	leaderAddr, err := s.LeaderAddress(60 * time.Second)
	if err != nil {
		return nil, err
	}

	var metadata map[string]interface{}
	for _, server := range cf.Configuration().Servers {
		if server.ID == raft.ServerID(id) {
			metadata, err = s.getNode(id)
			if err != nil {
				return nil, err
			}
			metadata["leader"] = server.Address == leaderAddr
			break
		}
	}

	return metadata, nil
}

func (s *RaftServer) SetNode(id string, metadata map[string]interface{}) error {
	if s.raft.State() != raft.Leader {
		//// forward to leader node
		//leaderId, err := s.LeaderID(60 * time.Second)
		//if err != nil {
		//	return err
		//}
		//
		//leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		//if err != nil {
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
		//	return err
		//}
		//
		//err = client.Join(node)
		//if err != nil {
		//	return err
		//}
		//
		//return nil

		return raft.ErrNotLeader
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

	// set node
	err = s.setNode(id, metadata)
	if err != nil {
		return err
	}

	// update cluster to manager
	err = s.updateCluster()
	if err != nil {
		return err
	}

	s.logger.Printf("[INFO] node %v joined successfully", id)
	return nil
}

func (s *RaftServer) DeleteNode(id string) error {
	if s.raft.State() != raft.Leader {
		//// forward to leader node
		//leaderId, err := s.LeaderID(60 * time.Second)
		//if err != nil {
		//	return err
		//}
		//
		//leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return nil
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
		//	return nil
		//}
		//
		//err = client.Leave(node)
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return nil
		//}
		//
		//return nil

		return raft.ErrNotLeader
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
	err = s.deleteNode(id)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil
	}

	// unregister node from manager
	err = s.updateCluster()
	if err != nil {
		return err
	}

	s.logger.Printf("[INFO] node %v does not exists in the cluster", id)
	return nil
}

func (s *RaftServer) GetCluster() (map[string]interface{}, error) {
	cf := s.raft.GetConfiguration()
	err := cf.Error()
	if err != nil {
		return nil, err
	}

	leaderAddr, err := s.LeaderAddress(60 * time.Second)
	if err != nil {
		return nil, err
	}

	cluster := map[string]interface{}{}
	for _, server := range cf.Configuration().Servers {
		metadata, err := s.getNode(string(server.ID))
		if err != nil {
			s.logger.Printf("[WARN] %v", err)
			continue
		}
		metadata["leader"] = server.Address == leaderAddr

		cluster[string(server.ID)] = metadata
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

func (s *RaftServer) GetDocument(id string) (map[string]interface{}, error) {
	fields, err := s.fsm.GetDocument(id)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func (s *RaftServer) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := s.fsm.Search(request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *RaftServer) IndexDocument(docs []map[string]interface{}) (int, error) {
	if s.raft.State() != raft.Leader {
		//// forward to leader node
		//leaderId, err := s.LeaderID(60 * time.Second)
		//if err != nil {
		//	return nil, err
		//}
		//
		//leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return nil, err
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
		//	return nil, err
		//}
		//
		//result, err := client.Index(docs)
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return nil, err
		//}
		//s.logger.Printf("[DEBUG] %v", result)
		//
		//return result, nil

		return -1, raft.ErrNotLeader
	}

	count := 0
	for _, doc := range docs {
		msg, err := newMessage(
			indexDocument,
			doc,
		)
		if err != nil {
			return -1, err
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			return -1, err
		}

		f := s.raft.Apply(msgBytes, 10*time.Second)
		err = f.Error()
		if err != nil {
			return -1, err
		}

		count++
	}

	return count, nil
}

func (s *RaftServer) DeleteDocument(ids []string) (int, error) {
	if s.raft.State() != raft.Leader {
		//// forward to leader node
		//leaderId, err := s.LeaderID(60 * time.Second)
		//if err != nil {
		//	return nil, err
		//}
		//
		//leaderNode, err := s.getNode(&blastraft.Node{Id: string(leaderId)})
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return nil, err
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
		//	return nil, err
		//}
		//
		//result, err := client.Delete(docs)
		//if err != nil {
		//	s.logger.Printf("[ERR] %v", err)
		//	return nil, err
		//}
		//s.logger.Printf("[DEBUG] %v", result)
		//
		//return result, nil

		return -1, raft.ErrNotLeader
	}

	count := 0
	for _, id := range ids {
		msg, err := newMessage(
			deleteDocument,
			id,
		)
		if err != nil {
			return -1, err
		}

		msgBytes, err := json.Marshal(msg)
		if err != nil {
			return -1, err
		}

		f := s.raft.Apply(msgBytes, 10*time.Second)
		err = f.Error()
		if err != nil {
			return -1, err
		}

		count++
	}

	return count, nil
}

func (s *RaftServer) GetIndexConfig() (map[string]interface{}, error) {
	indexConfig, err := s.fsm.GetIndexConfig()
	if err != nil {
		return nil, err
	}

	return indexConfig, nil
}

func (s *RaftServer) GetIndexStats() (map[string]interface{}, error) {
	indexStats, err := s.fsm.GetIndexStats()
	if err != nil {
		return nil, err
	}

	return indexStats, nil
}
