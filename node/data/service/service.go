// Copyright (c) 2018 Minoru Osuka
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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/blevesearch/bleve/config"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"github.com/mosuka/blast/index"
	blastlog "github.com/mosuka/blast/log"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	blastraft "github.com/mosuka/blast/raft"
	"github.com/mosuka/blast/store"
	"github.com/pkg/errors"
)

type Service struct {
	peerAddress string
	raftConfig  *blastraft.RaftConfig
	bootstrap   bool
	raft        *raft.Raft

	storeConfig *store.StoreConfig
	store       *store.Store

	indexConfig *index.IndexConfig
	index       *index.Index

	mutex sync.Mutex

	metadata map[string]*protobuf.Metadata

	logger *log.Logger
}

func NewService(peerAddress string, raftConfig *blastraft.RaftConfig, bootstrap bool, storeConfig *store.StoreConfig, indexConfig *index.IndexConfig) (*Service, error) {
	return &Service{
		peerAddress: peerAddress,
		raftConfig:  raftConfig,
		bootstrap:   bootstrap,

		storeConfig: storeConfig,
		indexConfig: indexConfig,

		metadata: make(map[string]*protobuf.Metadata, 0),

		logger: blastlog.DefaultLogger(),
	}, nil
}

func (s *Service) SetLogger(logger *log.Logger) {
	s.logger = logger
	return
}

func (s *Service) Start() error {
	var err error

	// Open store
	s.store, err = store.NewStore(s.storeConfig)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}
	s.store.SetLogger(s.logger)

	// Open index
	s.index, err = index.NewIndex(s.indexConfig)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}
	s.index.SetLogger(s.logger)

	// Setup Raft configuration.
	config := s.raftConfig.Config
	config.Logger = s.logger

	// Setup Raft communication.
	peerAddr, err := net.ResolveTCPAddr("tcp", s.peerAddress)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}
	transport, err := raft.NewTCPTransportWithLogger(s.peerAddress, peerAddr, 3, s.raftConfig.Timeout, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStoreWithLogger(s.raftConfig.Dir, s.raftConfig.SnapshotCount, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	// Create the log store and stable store.
	raftLogStore := filepath.Join(s.raftConfig.Dir, "raft_log.db")
	logStore, err := raftboltdb.NewBoltStore(raftLogStore)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	// Instantiate the Raft systems.
	s.raft, err = raft.NewRaft(config, (*fsm)(s), logStore, logStore, snapshots, transport)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
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
		future := s.raft.BootstrapCluster(configuration)
		err = future.Error()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return err
		}
	}

	return nil
}

func (s *Service) Stop() error {
	err := s.index.Close()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	err = s.store.Close()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	return nil
}

func (s *Service) LeaderAddress(timeout time.Duration) (raft.ServerAddress, error) {
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
			return "", fmt.Errorf("timeout expired")
		}
	}
}

func (s *Service) LeaderID(timeout time.Duration) (raft.ServerID, error) {
	leaderAddr, err := s.LeaderAddress(timeout)
	if err != nil {
		return "", err
	}

	configFuture := s.raft.GetConfiguration()
	if err = configFuture.Error(); err != nil {
		return "", err
	}

	var leaderID raft.ServerID
	for _, srv := range configFuture.Configuration().Servers {
		if srv.Address == leaderAddr {
			leaderID = srv.ID
			break
		}
	}

	return leaderID, nil
}

func (s *Service) WaitDetectLeader(timeout time.Duration) error {
	_, err := s.LeaderAddress(timeout)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetMetadata(id string) (*protobuf.Metadata, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exist := s.metadata[id]
	if !exist {
		err := fmt.Errorf("%s does not exist in metadata", id)
		s.logger.Printf("[ERR] %v", err)
		return nil, err
	}

	return s.metadata[id], nil
}

func (s *Service) PutMetadata(req *protobuf.JoinRequest) error {
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_JOIN,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	return nil
}

func (s *Service) DeleteMetadata(req *protobuf.LeaveRequest) error {
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_LEAVE,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	f := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = f.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	return nil
}

func (s *Service) Join(ctx context.Context, req *protobuf.JoinRequest) (*protobuf.JoinResponse, error) {
	start := time.Now()
	defer Metrics(start, "Join")

	resp := &protobuf.JoinResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] Not a leader")

		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderGRPCAddress := leaderMetadata.GrpcAddress

		grpcClient, err := client.NewGRPCClient(leaderGRPCAddress)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.Join(req)
	}

	peerAddr, err := net.ResolveTCPAddr("tcp", req.Address)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	configFuture := s.raft.GetConfiguration()
	err = configFuture.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(req.NodeId) || srv.Address == raft.ServerAddress(peerAddr.String()) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.ID == raft.ServerID(req.NodeId) && srv.Address == raft.ServerAddress(peerAddr.String()) {
				message := fmt.Sprintf("%s (%s) already exists in cluster, ignoring join request", req.NodeId, srv.Address)
				s.logger.Printf("[DEBUG] %s", message)
				resp.Success = true
				resp.Message = message
				return resp, nil
			}

			// If either the ID or the address matches, need to remove the corresponding node by the ID.
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			err = future.Error()
			if err != nil {
				s.logger.Print(fmt.Sprintf("[ERR] %v", err))
				resp.Success = false
				resp.Message = err.Error()
				return resp, err
			}
		}
	}

	future := s.raft.AddVoter(raft.ServerID(req.NodeId), raft.ServerAddress(peerAddr.String()), 0, 0)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	err = s.PutMetadata(req)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp.Success = true

	return resp, nil
}

func (s *Service) Leave(ctx context.Context, req *protobuf.LeaveRequest) (*protobuf.LeaveResponse, error) {
	start := time.Now()
	defer Metrics(start, "Leave")

	resp := &protobuf.LeaveResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] Not a leader")

		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderGRPCAddress := leaderMetadata.GrpcAddress

		grpcClient, err := client.NewGRPCClient(leaderGRPCAddress)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.Leave(req)
	}

	peerAddr, err := net.ResolveTCPAddr("tcp", req.Address)
	if err != nil {
		s.logger.Print(fmt.Sprintf("[ERR] %v", err))
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	configFuture := s.raft.GetConfiguration()
	err = configFuture.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(req.NodeId) && srv.Address == raft.ServerAddress(peerAddr.String()) {
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			err = future.Error()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				resp.Success = false
				resp.Message = err.Error()
				return resp, err
			}
		}
	}

	future := s.raft.DemoteVoter(raft.ServerID(req.NodeId), 0, 0)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err

	}

	err = s.DeleteMetadata(req)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp.Success = true

	return resp, nil
}

func (s *Service) Peers(ctx context.Context, req *empty.Empty) (*protobuf.PeersResponse, error) {
	start := time.Now()
	defer Metrics(start, "Peers")

	resp := &protobuf.PeersResponse{}

	configFuture := s.raft.GetConfiguration()
	err := configFuture.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	leaderAddr, err := s.LeaderAddress(60 * time.Second)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp.Peers = make([]*protobuf.Peer, 0)
	for _, srv := range configFuture.Configuration().Servers {
		peer := &protobuf.Peer{}
		peer.NodeId = string(srv.ID)
		peer.Address = string(srv.Address)
		peer.Leader = srv.Address == leaderAddr
		peer.Metadata, err = s.GetMetadata(string(srv.ID))
		if err != nil {
			s.logger.Printf("[WARN] %v", err)
		}
		resp.Peers = append(resp.Peers, peer)
	}

	resp.Success = true

	return resp, nil
}

func (s *Service) Snapshot(ctx context.Context, req *empty.Empty) (*protobuf.SnapshotResponse, error) {
	start := time.Now()
	defer Metrics(start, "Snapshot")

	resp := &protobuf.SnapshotResponse{}

	future := s.raft.Snapshot()
	err := future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp.Success = true

	return resp, nil
}

func (s *Service) Get(ctx context.Context, req *protobuf.GetRequest) (*protobuf.GetResponse, error) {
	start := time.Now()
	defer Metrics(start, "Get")

	resp := &protobuf.GetResponse{}

	reader, err := s.store.Reader()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer reader.Close()

	fieldsBytes, err := reader.Get([]byte(req.Id))
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	if fieldsBytes != nil {
		var fieldsMap map[string]interface{}
		err = json.Unmarshal(fieldsBytes, &fieldsMap)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		fieldsAny := &any.Any{}
		err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		resp.Fields = fieldsAny
	}

	resp.Id = req.Id
	resp.Success = true

	return resp, nil
}

func (s *Service) Put(ctx context.Context, req *protobuf.PutRequest) (*protobuf.PutResponse, error) {
	start := time.Now()
	defer Metrics(start, "Put")

	resp := &protobuf.PutResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] Not a leader")

		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderGRPCAddress := leaderMetadata.GrpcAddress

		grpcClient, err := client.NewGRPCClient(leaderGRPCAddress)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.Put(req)
	}

	// PutRequest to Any
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_PUT,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp = future.Response().(*protobuf.PutResponse)

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, req *protobuf.DeleteRequest) (*protobuf.DeleteResponse, error) {
	start := time.Now()
	defer Metrics(start, "Delete")

	resp := &protobuf.DeleteResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] Not a leader")

		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderGRPCAddress := leaderMetadata.GrpcAddress

		grpcClient, err := client.NewGRPCClient(leaderGRPCAddress)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.Delete(req)
	}

	// DeleteRequest to Any
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_DELETE,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp = future.Response().(*protobuf.DeleteResponse)

	return resp, nil
}

func (s *Service) Bulk(ctx context.Context, req *protobuf.BulkRequest) (*protobuf.BulkResponse, error) {
	start := time.Now()
	defer Metrics(start, "Bulk")

	resp := &protobuf.BulkResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] Not a leader")

		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}

		leaderGRPCAddress := leaderMetadata.GrpcAddress

		grpcClient, err := client.NewGRPCClient(leaderGRPCAddress)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			resp.Success = false
			resp.Message = err.Error()
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.Bulk(req)
	}

	// BulkRequest to Any
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_BULK,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp = future.Response().(*protobuf.BulkResponse)

	return resp, nil
}

func (s *Service) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	start := time.Now()
	defer Metrics(start, "Search")

	resp := &protobuf.SearchResponse{}

	searcher, err := s.index.Searcher()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer searcher.Close()

	searchRequest, err := protobuf.MarshalAny(req.SearchRequest)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	searchRequestBytes, err := json.Marshal(searchRequest.(*map[string]interface{}))
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	searchResultBytes, err := searcher.Search(searchRequestBytes)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	var searchResult map[string]interface{}
	err = json.Unmarshal(searchResultBytes, &searchResult)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	searchResultAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchResult, searchResultAny)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp.SearchResult = searchResultAny
	resp.Success = true

	return resp, nil
}

// ----------------------------------------

type fsm Service

// Apply applies a Raft log entry to the key-value store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	start := time.Now()
	defer Metrics(start, "Apply")

	proposal := &protobuf.Proposal{}
	err := proto.Unmarshal(l.Data, proposal)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	data, err := protobuf.MarshalAny(proposal.Data)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	switch proposal.Type {
	case protobuf.Proposal_PUT:
		var resp interface{}
		resp, err = f.applyPut(data.(*protobuf.PutRequest))
		return resp.(*protobuf.PutResponse)
	case protobuf.Proposal_DELETE:
		var resp interface{}
		resp, err = f.applyDelete(data.(*protobuf.DeleteRequest))
		return resp.(*protobuf.DeleteResponse)
	case protobuf.Proposal_BULK:
		var resp interface{}
		resp, err = f.applyBulk(data.(*protobuf.BulkRequest))
		return resp.(*protobuf.BulkResponse)
	case protobuf.Proposal_JOIN:
		var resp interface{}
		resp, err = f.applyJoin(data.(*protobuf.JoinRequest))
		return resp.(*protobuf.JoinResponse)
	case protobuf.Proposal_LEAVE:
		var resp interface{}
		resp, err = f.applyLeave(data.(*protobuf.LeaveRequest))
		return resp.(*protobuf.LeaveResponse)
	default:
		err = errors.New("unsupported proposal type")
		f.logger.Printf("[ERR] %v", err)
		return err
	}
}

// Snapshot returns a snapshot of the key-value store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	start := time.Now()
	defer Metrics(start, "Snapshot")

	f.mutex.Lock()
	defer f.mutex.Unlock()

	iterator, err := f.store.Iterator()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return nil, err
	}
	defer iterator.Close()

	// Clone the map.
	data := make(map[string][]byte)
	for iterator.Next() {
		data[string(iterator.Key())] = iterator.Value()
	}

	return &fsmSnapshot{
		logger: f.logger,
		data:   data,
	}, nil
}

// Restore stores the key-value store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	start := time.Now()
	defer Metrics(start, "Restore")

	data := make(map[string][]byte)
	err := json.NewDecoder(rc).Decode(&data)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	batchSize := 1000

	// Restore store
	bulkWriter, err := f.store.Bulker(batchSize)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}
	defer bulkWriter.Close()

	bulkIndexer, err := f.index.Bulker(batchSize)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}
	defer bulkIndexer.Close()

	for key, value := range data {
		// Restore store
		err = bulkWriter.Put([]byte(key), value)
		if err != nil {
			f.logger.Printf("[WARN] %v", err)
		}

		// Restore index
		var fields map[string]interface{}
		err = json.Unmarshal(value, &fields)
		if err != nil {
			f.logger.Printf("[WARN] %v", err)
		}
		err = bulkIndexer.Index(key, fields)
		if err != nil {
			f.logger.Printf("[WARN] %v", err)
		}
	}

	return nil
}

func (f *fsm) applyPut(req *protobuf.PutRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyPut")

	resp := &protobuf.PutResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Put into store
	writer, err := f.store.Writer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer writer.Close()

	fieldsBytes, err := req.GetFieldsBytes()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	err = writer.Put([]byte(req.Id), fieldsBytes)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	// Index into index
	indexer, err := f.index.Indexer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer indexer.Close()

	fieldsMap, err := req.GetFieldsMap()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	err = indexer.Index(req.Id, fieldsMap)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp.Success = true

	return resp, nil
}

func (f *fsm) applyDelete(req *protobuf.DeleteRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyDelete")

	resp := &protobuf.DeleteResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Delete from store
	writer, err := f.store.Writer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer writer.Close()

	err = writer.Delete([]byte(req.Id))
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	// Delete from index
	indexer, err := f.index.Indexer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer indexer.Close()

	err = indexer.Delete(req.Id)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}

	resp.Success = true

	return resp, nil
}

func (f *fsm) applyBulk(req *protobuf.BulkRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyBulk")

	resp := &protobuf.BulkResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Store bulker
	writer, err := f.store.Bulker(int(req.BatchSize))
	if err != nil {
		f.logger.Print(fmt.Sprintf("[ERR] %v", err))
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer writer.Close()

	// Index bulker
	indexer, err := f.index.Bulker(int(req.BatchSize))
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	defer indexer.Close()

	var (
		putCnt    int
		deleteCnt int
	)

	for _, updateRequest := range req.UpdateRequests {
		switch updateRequest.Type {
		case protobuf.UpdateRequest_PUT:
			fieldsBytes, err := updateRequest.Document.GetFieldsBytes()
			if err != nil {
				f.logger.Printf("[WARN] %v", err)
				continue
			}

			err = writer.Put([]byte(updateRequest.Document.Id), fieldsBytes)
			if err != nil {
				f.logger.Printf("[WARN] %v", err)
				continue
			}

			fieldsMap, err := updateRequest.Document.GetFieldsMap()
			if err != nil {
				f.logger.Printf("[WARN] %v", err)
				continue
			}

			err = indexer.Index(updateRequest.Document.Id, fieldsMap)
			if err != nil {
				f.logger.Printf("[WARN] %v", err)
				continue
			}

			putCnt++
		case protobuf.UpdateRequest_DELETE:
			err = writer.Delete([]byte(updateRequest.Document.Id))
			if err != nil {
				f.logger.Printf("[WARN] %v", err)
				continue
			}

			err = indexer.Delete(updateRequest.Document.Id)
			if err != nil {
				f.logger.Printf("[WARN] %v", err)
				continue
			}

			deleteCnt++
		default:
			f.logger.Printf("[WARN] Skip bulk operation: %s", updateRequest.Type.String())
			continue
		}
	}

	resp.PutCount = int32(putCnt)
	resp.DeleteCount = int32(deleteCnt)
	resp.Success = true

	return resp, nil
}

func (f *fsm) applyJoin(req *protobuf.JoinRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyJoin")

	resp := &protobuf.JoinResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	_, exist := f.metadata[req.NodeId]
	if exist {
		f.logger.Printf("[INFO] %s already exist in metadata", req.NodeId)
	} else {
		f.logger.Printf("[INFO] Put %s in metadata", req.NodeId)
		f.metadata[req.NodeId] = req.Metadata
	}

	resp.Success = true

	return resp, nil
}

func (f *fsm) applyLeave(req *protobuf.LeaveRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyLeave")

	resp := &protobuf.LeaveResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	_, exist := f.metadata[req.NodeId]
	if exist {
		f.logger.Printf("[INFO] Delete %s from metadata", req.NodeId)
		delete(f.metadata, req.NodeId)
	} else {
		f.logger.Printf("[INFO] %s does not exist in metadata", req.NodeId)
	}

	resp.Success = true

	return resp, nil
}

// ----------------------------------------

type fsmSnapshot struct {
	logger *log.Logger
	data   map[string][]byte
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	start := time.Now()
	defer Metrics(start, "Persist")

	err := func() error {
		// Encode data.
		b, err := json.Marshal(f.data)
		if err != nil {
			f.logger.Printf("[ERR] %v", err)
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			f.logger.Printf("[ERR] %v", err)
			return err
		}

		// Close the sink.
		return sink.Close()
	}()
	if err != nil {
		sink.Cancel()
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	return nil
}

func (f *fsmSnapshot) Release() {}
