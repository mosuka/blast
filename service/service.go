//  Copyright (c) 2018 Minoru Osuka
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
	"math"
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
	"github.com/mosuka/blast/grpc/client"
	"github.com/mosuka/blast/index/bleve"
	"github.com/mosuka/blast/logging"
	"github.com/mosuka/blast/protobuf"
	braft "github.com/mosuka/blast/raft"
	"github.com/mosuka/blast/store/boltdb"
	"github.com/pkg/errors"
)

type KVSService struct {
	peerAddress string
	raftConfig  *braft.RaftConfig
	bootstrap   bool
	raft        *raft.Raft

	storeConfig *boltdb.StoreConfig
	store       *boltdb.Store

	indexConfig *bleve.IndexConfig
	index       *bleve.Index

	maxSendMessageSize    int
	maxReceiveMessageSize int

	mutex sync.Mutex

	metadata map[string]*protobuf.Metadata

	logger *log.Logger
}

func NewKVSService(peerAddress string, raftConfig *braft.RaftConfig, bootstrap bool, storeConfig *boltdb.StoreConfig, indexConfig *bleve.IndexConfig) (*KVSService, error) {
	return &KVSService{
		peerAddress: peerAddress,
		raftConfig:  raftConfig,
		bootstrap:   bootstrap,

		storeConfig: storeConfig,
		indexConfig: indexConfig,

		maxSendMessageSize:    math.MaxInt32,
		maxReceiveMessageSize: math.MaxInt32,

		metadata: make(map[string]*protobuf.Metadata, 0),

		logger: logging.DefaultLogger(),
	}, nil
}

func (s *KVSService) SetLogger(logger *log.Logger) {
	s.logger = logger
	return
}

func (s *KVSService) Start() error {
	var err error

	// Open store
	if s.store, err = boltdb.NewStore(s.storeConfig); err != nil {
		s.logger.Printf("[ERR] service: Failed to open store: %v: %v", s.storeConfig, err)
		return err
	}
	s.store.SetLogger(s.logger)
	s.logger.Printf("[INFO] service: Store has been opened %v", s.storeConfig)

	// Open index
	if s.index, err = bleve.NewIndex(s.indexConfig); err != nil {
		s.logger.Printf("[ERR] service: Failed to open index: %v, %v", s.indexConfig, err)
		return err
	}
	s.index.SetLogger(s.logger)
	s.logger.Printf("[INFO] service: Index has been opened %v", s.indexConfig)

	// Setup Raft configuration.
	config := s.raftConfig.Config
	config.Logger = s.logger

	// Setup Raft communication.
	var peerAddr *net.TCPAddr
	if peerAddr, err = net.ResolveTCPAddr("tcp", s.peerAddress); err != nil {
		s.logger.Printf("[ERR] service: Failed to resolve TCP address for Raft at %s: %v", s.peerAddress, err)
		return err
	}
	s.logger.Printf("[INFO] service: TCP address for Raft at %s has been resolved", s.peerAddress)
	var transport *raft.NetworkTransport
	if transport, err = raft.NewTCPTransportWithLogger(s.peerAddress, peerAddr, 3, s.raftConfig.Timeout, s.logger); err != nil {
		s.logger.Printf("[ERR] service: Failed to create TCP transport for Raft at %s: %v", s.peerAddress, err)
		return err
	}
	s.logger.Printf("[INFO] service: TCP transport for Raft at %s has been created", s.peerAddress)

	// Create the snapshot store. This allows the Raft to truncate the log.
	var snapshots *raft.FileSnapshotStore
	if snapshots, err = raft.NewFileSnapshotStoreWithLogger(s.raftConfig.Path, s.raftConfig.RetainSnapshotCount, s.logger); err != nil {
		s.logger.Printf("[ERR] service: Failed to create snapshot store at %s: %v", s.raftConfig.Path, err)
		return err
	}
	s.logger.Printf("[INFO] service: Snapshot store has been created at %s", s.raftConfig.Path)

	// Create the log store and stable store.
	raftLogStore := filepath.Join(s.raftConfig.Path, "raft_log.db")
	var logStore *raftboltdb.BoltStore
	if logStore, err = raftboltdb.NewBoltStore(raftLogStore); err != nil {
		s.logger.Printf("[ERR] service: Failed to create log store at %s: %v", raftLogStore, err)
		return err
	}
	s.logger.Printf("[INFO] service: Log store has been created at %s", raftLogStore)

	// Instantiate the Raft systems.
	if s.raft, err = raft.NewRaft(config, (*fsm)(s), logStore, logStore, snapshots, transport); err != nil {
		s.logger.Printf("[ERR] service: Failed to instantiate Raft: %v", err)
		return err
	}
	s.logger.Print("[INFO] service: Raft has been instantiated")

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
		s.logger.Print("[INFO] service: Raft has been started as bootstrap")
	}

	s.logger.Print("[INFO] server: Service has been started")

	return nil
}

func (s *KVSService) Stop() error {
	var err error

	// Close index
	if err = s.index.Close(); err != nil {
		s.logger.Printf("[ERR] service: Failed to close index: %v", err)
	}
	s.logger.Print("[INFO] service: Index has been closed")

	// Close store
	if err = s.store.Close(); err != nil {
		s.logger.Printf("[ERR] service: Failed to close store: %s", err.Error())
	}
	s.logger.Print("[INFO] service: Store has been closed")

	s.logger.Print("[INFO] server: Service has been stopped")

	return nil
}

func (s *KVSService) WaitForLeader(timeout time.Duration) (string, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ticker.C:
			leader := string(s.raft.Leader())
			if leader != "" {
				s.logger.Printf("[INFO] server: Leader has been detected at %s", leader)
				return leader, nil
			}
		case <-timer.C:
			s.logger.Print("[ERR] server: Detecting leader timeout expired")
			return "", fmt.Errorf("timeout expired")
		}
	}
}

func (s *KVSService) LeaderID() (string, error) {
	var err error

	configFuture := s.raft.GetConfiguration()
	if err = configFuture.Error(); err != nil {
		return "", err
	}

	var leaderID string
	for _, srv := range configFuture.Configuration().Servers {
		if srv.Address == s.raft.Leader() {
			leaderID = string(srv.ID)
			break
		}
	}

	return leaderID, nil
}

func (s *KVSService) GetMetadata(id string) (*protobuf.Metadata, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.metadata[id]; !ok {
		return nil, errors.New("Node does not exist in metadata")
	}

	return s.metadata[id], nil
}

func (s *KVSService) PutMetadata(req *protobuf.JoinRequest) error {
	var err error

	// JoinRequest to Any
	data := &any.Any{}
	if err = protobuf.UnmarshalAny(req, data); err != nil {
		message := fmt.Sprintf("Failed to unmarshalling data to any: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		return err
	}

	var proposalBytes []byte
	if proposalBytes, err = proto.Marshal(&protobuf.Proposal{
		Type: protobuf.Proposal_JOIN,
		Data: data,
	}); err != nil {
		message := fmt.Sprintf("Failed to marshalling proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		return err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	if err = future.Error(); err != nil {
		message := fmt.Sprintf("Failed to apply join proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		return err
	}

	return nil
}

func (s *KVSService) DeleteMetadata(req *protobuf.LeaveRequest) error {
	var err error

	data := &any.Any{}
	if err = protobuf.UnmarshalAny(req, data); err != nil {
		message := fmt.Sprintf("Failed to unmarshalling data to any: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		return err
	}

	var proposalBytes []byte
	if proposalBytes, err = proto.Marshal(&protobuf.Proposal{
		Type: protobuf.Proposal_LEAVE,
		Data: data,
	}); err != nil {
		message := fmt.Sprintf("Failed to marshalling proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		return err
	}

	f := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	if err = f.Error(); err != nil {
		message := fmt.Sprintf("Failed to apply leave proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		return err
	}

	return nil
}

func (s *KVSService) Join(ctx context.Context, req *protobuf.JoinRequest) (*protobuf.JoinResponse, error) {
	start := time.Now()
	defer Metrics(start, "Join")

	var err error

	resp := &protobuf.JoinResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] service: Not a leader")

		// Wait for leader detected
		if _, err = s.WaitForLeader(60 * time.Second); err != nil {
			message := "Failed to detect leader"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderID string
		if leaderID, err = s.LeaderID(); err != nil {
			message := "Failed to get leader ID"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderGRPCAddress string
		leaderGRPCAddress = s.metadata[leaderID].GrpcAddress

		var grpcClient *client.GRPCClient
		if grpcClient, err = client.NewGRPCClient(leaderGRPCAddress, s.maxSendMessageSize, s.maxReceiveMessageSize); err != nil {
			message := fmt.Sprintf("Failed to create gRPC client for %s", leaderGRPCAddress)
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}
		defer grpcClient.Close()

		if resp, err = grpcClient.Join(req); err != nil {
			s.logger.Printf("[ERR] service: Node could not join to cluster: %v: %v", req, err)
			return resp, err
		}
		return resp, nil
	}

	var peerAddr *net.TCPAddr
	if peerAddr, err = net.ResolveTCPAddr("tcp", req.Address); err != nil {
		message := fmt.Sprintf("Failed to resolve TCP address for Raft at %s", req.Address)
		s.logger.Print(fmt.Sprintf("[ERR] service: %s: %v", message, err))
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	configFuture := s.raft.GetConfiguration()
	if err = configFuture.Error(); err != nil {
		message := "Failed to get Raft configuration"
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(req.NodeId) || srv.Address == raft.ServerAddress(peerAddr.String()) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.ID == raft.ServerID(req.NodeId) && srv.Address == raft.ServerAddress(peerAddr.String()) {
				message := fmt.Sprintf("Node already member of cluster, ignoring join request: %v", req)
				s.logger.Printf("[INFO] service: %s", message)
				resp.Success = true
				resp.Message = message
				return resp, nil
			}

			// If either the ID or the address matches, need to remove the corresponding node by the ID.
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err = future.Error(); err != nil {
				message := fmt.Sprintf("Failed to remove existing node: %v", req)
				s.logger.Print(fmt.Sprintf("[ERR] service: %s: %v", message, err))
				resp.Success = false
				resp.Message = message
				return resp, err
			}
		}
	}

	future := s.raft.AddVoter(raft.ServerID(req.NodeId), raft.ServerAddress(peerAddr.String()), 0, 0)
	if err = future.Error(); err != nil {
		message := fmt.Sprintf("Failed to add node to cluster: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	if err = s.PutMetadata(req); err != nil {
		message := "Failed to put metadata"
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	s.logger.Printf("[INFO] service: Node has been joined to cluster: %v", req)

	resp.Success = true

	return resp, nil
}

func (s *KVSService) Leave(ctx context.Context, req *protobuf.LeaveRequest) (*protobuf.LeaveResponse, error) {
	start := time.Now()
	defer Metrics(start, "Leave")

	var err error

	resp := &protobuf.LeaveResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] service: Not a leader")

		// Wait for leader detected
		if _, err = s.WaitForLeader(60 * time.Second); err != nil {
			message := "Failed to detect leader"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderID string
		if leaderID, err = s.LeaderID(); err != nil {
			message := "Failed to get leader ID"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderGRPCAddress string
		leaderGRPCAddress = s.metadata[leaderID].GrpcAddress

		var grpcClient *client.GRPCClient
		if grpcClient, err = client.NewGRPCClient(leaderGRPCAddress, s.maxSendMessageSize, s.maxReceiveMessageSize); err != nil {
			message := "Failed to create gRPC client"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}
		defer grpcClient.Close()

		if resp, err = grpcClient.Leave(req); err != nil {
			s.logger.Printf("[ERR] service: Node could not leave from cluster: %v: %v", req, err)
			return resp, err
		}
		return resp, nil
	}

	var peerAddr *net.TCPAddr
	if peerAddr, err = net.ResolveTCPAddr("tcp", req.Address); err != nil {
		message := fmt.Sprintf("Failed to resolve TCP address for Raft at %s", req.Address)
		s.logger.Print(fmt.Sprintf("[ERR] service: %s: %v", message, err))
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	configFuture := s.raft.GetConfiguration()
	if err = configFuture.Error(); err != nil {
		message := "Failed to get Raft configuration"
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(req.NodeId) && srv.Address == raft.ServerAddress(peerAddr.String()) {
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			if err = future.Error(); err != nil {
				message := fmt.Sprintf("Failed to remove existing node: %v", req)
				s.logger.Printf("[ERR] service: %s: %v", message, err)
				resp.Success = false
				resp.Message = message
				return resp, err
			}
		}
	}

	future := s.raft.DemoteVoter(raft.ServerID(req.NodeId), 0, 0)
	if err = future.Error(); err != nil {
		message := fmt.Sprintf("Failed to remove node from cluster: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err

	}

	if err = s.DeleteMetadata(req); err != nil {
		message := "Failed to delete metadata"
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	s.logger.Printf("[INFO] service: Node has been left from cluster: %v", req)

	resp.Success = true

	return resp, nil
}

func (s *KVSService) Peers(ctx context.Context, req *empty.Empty) (*protobuf.PeersResponse, error) {
	start := time.Now()
	defer Metrics(start, "Peers")

	var err error

	resp := &protobuf.PeersResponse{}

	configFuture := s.raft.GetConfiguration()
	if err = configFuture.Error(); err != nil {
		message := "Failed to get Raft configuration"
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	peers := make([]*protobuf.Peer, 0)
	for _, srv := range configFuture.Configuration().Servers {
		peer := &protobuf.Peer{}
		peer.NodeId = string(srv.ID)
		peer.Address = string(srv.Address)
		peer.Leader = srv.Address == s.raft.Leader()
		var metadata *protobuf.Metadata
		if metadata, err = s.GetMetadata(string(srv.ID)); err != nil {
			message := "Failed to get metadata"
			s.logger.Printf("[WARN] service: %s: %v", message, err)
		}
		peer.Metadata = metadata
		peers = append(peers, peer)
	}

	s.logger.Print("[INFO] service: Raft peers info has been got")

	resp.Peers = peers
	resp.Success = true

	return resp, nil
}

func (s *KVSService) Snapshot(ctx context.Context, req *empty.Empty) (*protobuf.SnapshotResponse, error) {
	start := time.Now()
	defer Metrics(start, "Snapshot")

	var err error

	resp := &protobuf.SnapshotResponse{}

	future := s.raft.Snapshot()
	if err = future.Error(); err != nil {
		message := "Failed to create snapshot"
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	s.logger.Print("[INFO] service: Snapshot has been done")

	resp.Success = true
	return resp, nil
}

func (s *KVSService) Get(ctx context.Context, req *protobuf.GetRequest) (*protobuf.GetResponse, error) {
	start := time.Now()
	defer Metrics(start, "Get")

	var err error

	resp := &protobuf.GetResponse{}

	var reader *boltdb.Reader
	if reader, err = s.store.Reader(); err != nil {
		message := "Failed to get reader"
		s.logger.Printf("[ERR] service: %s, %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	defer reader.Close()
	s.logger.Print("[DEBUG] service: Store reader has been got")

	var fieldsBytes []byte
	if fieldsBytes, err = reader.Get([]byte(req.Id)); err != nil {
		message := fmt.Sprintf("Failed to get fields: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	if fieldsBytes != nil {
		s.logger.Printf("[INFO] service: Document has been got: %v", req)

		var fieldsMap map[string]interface{}
		if err = json.Unmarshal(fieldsBytes, &fieldsMap); err != nil {
			message := fmt.Sprintf("Failed to unmarshal fields to map: %v", fieldsBytes)
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		fieldsAny := &any.Any{}
		if err = protobuf.UnmarshalAny(fieldsMap, fieldsAny); err != nil {
			message := "Failed to unmarshal fields to any"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		resp.Fields = fieldsAny
	} else {
		resp.Message = ""
		s.logger.Printf("[INFO] service: No such document: %v", req)
	}
	resp.Id = req.Id

	resp.Success = true
	return resp, nil
}

func (s *KVSService) Put(ctx context.Context, req *protobuf.PutRequest) (*protobuf.PutResponse, error) {
	start := time.Now()
	defer Metrics(start, "Put")

	var err error

	resp := &protobuf.PutResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] service: Not a leader")

		// Wait for leader detected
		if _, err = s.WaitForLeader(60 * time.Second); err != nil {
			message := "Failed to detect leader"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderID string
		if leaderID, err = s.LeaderID(); err != nil {
			message := "Failed to get leader ID"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderGRPCAddress string
		leaderGRPCAddress = s.metadata[leaderID].GrpcAddress

		var grpcClient *client.GRPCClient
		if grpcClient, err = client.NewGRPCClient(leaderGRPCAddress, s.maxSendMessageSize, s.maxReceiveMessageSize); err != nil {
			message := "Failed to create gRPC client"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}
		defer grpcClient.Close()

		if resp, err = grpcClient.Put(req); err != nil {
			s.logger.Printf("[ERR] service: Failed to put document: %v: %v", req, err)
			return resp, err
		}
		return resp, nil
	}

	// PutRequest to Any
	data := &any.Any{}
	if err = protobuf.UnmarshalAny(req, data); err != nil {
		message := fmt.Sprintf("Failed to unmarshalling put request to any: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	var proposalBytes []byte
	if proposalBytes, err = proto.Marshal(&protobuf.Proposal{
		Type: protobuf.Proposal_PUT,
		Data: data,
	}); err != nil {
		message := fmt.Sprintf("Failed to marshalling proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	if err = future.Error(); err != nil {
		message := fmt.Sprintf("Failed to apply put proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	s.logger.Printf("[INFO] service: Put proposal data has been applyed: %v", req)

	resp = future.Response().(*protobuf.PutResponse)

	return resp, nil
}

func (s *KVSService) Delete(ctx context.Context, req *protobuf.DeleteRequest) (*protobuf.DeleteResponse, error) {
	start := time.Now()
	defer Metrics(start, "Delete")

	var err error

	resp := &protobuf.DeleteResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] service: Not a leader")

		// Wait for leader detected
		if _, err = s.WaitForLeader(60 * time.Second); err != nil {
			message := "Failed to detect leader"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderID string
		if leaderID, err = s.LeaderID(); err != nil {
			message := "Failed to get leader ID"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderGRPCAddress string
		leaderGRPCAddress = s.metadata[leaderID].GrpcAddress

		var grpcClient *client.GRPCClient
		if grpcClient, err = client.NewGRPCClient(leaderGRPCAddress, s.maxSendMessageSize, s.maxReceiveMessageSize); err != nil {
			message := "Failed to create gRPC client"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}
		defer grpcClient.Close()

		if resp, err = grpcClient.Delete(req); err != nil {
			s.logger.Printf("[ERR] service: Failed to delete document: %v: %v", req, err)
			return resp, err
		}
		return resp, nil
	}

	// DeleteRequest to Any
	data := &any.Any{}
	if err = protobuf.UnmarshalAny(req, data); err != nil {
		message := fmt.Sprintf("Failed to unmarshalling data to any: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	var proposalBytes []byte
	if proposalBytes, err = proto.Marshal(&protobuf.Proposal{
		Type: protobuf.Proposal_DELETE,
		Data: data,
	}); err != nil {
		message := fmt.Sprintf("Failed to marshalling proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	if err = future.Error(); err != nil {
		message := fmt.Sprintf("Failed to apply delete proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	s.logger.Printf("[INFO] service: Delete proposal data has been applyed: %v", req)

	resp = future.Response().(*protobuf.DeleteResponse)

	return resp, nil
}

func (s *KVSService) Bulk(ctx context.Context, req *protobuf.BulkRequest) (*protobuf.BulkResponse, error) {
	start := time.Now()
	defer Metrics(start, "Bulk")

	var err error

	resp := &protobuf.BulkResponse{}

	if s.raft.State() != raft.Leader {
		s.logger.Print("[DEBUG] service: Not a leader")

		// Wait for leader detected
		if _, err = s.WaitForLeader(60 * time.Second); err != nil {
			message := "Failed to detect leader"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderID string
		if leaderID, err = s.LeaderID(); err != nil {
			message := "Failed to get leader ID"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}

		var leaderGRPCAddress string
		leaderGRPCAddress = s.metadata[leaderID].GrpcAddress

		var grpcClient *client.GRPCClient
		if grpcClient, err = client.NewGRPCClient(leaderGRPCAddress, s.maxSendMessageSize, s.maxReceiveMessageSize); err != nil {
			message := "Failed to create gRPC client"
			s.logger.Printf("[ERR] service: %s: %v", message, err)
			resp.Success = false
			resp.Message = message
			return resp, err
		}
		defer grpcClient.Close()

		if resp, err = grpcClient.Bulk(req); err != nil {
			s.logger.Printf("[ERR] service: Failed to update document in bulk: %v: %v", req, err)
			return resp, err
		}
		return resp, nil
	}

	// BulkRequest to Any
	data := &any.Any{}
	if err = protobuf.UnmarshalAny(req, data); err != nil {
		message := fmt.Sprintf("Failed to unmarshalling data to any: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	var proposalBytes []byte
	if proposalBytes, err = proto.Marshal(&protobuf.Proposal{
		Type: protobuf.Proposal_BULK,
		Data: data,
	}); err != nil {
		message := fmt.Sprintf("Failed to marshalling proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	if err = future.Error(); err != nil {
		message := fmt.Sprintf("Failed to apply bulk proposal data: %v", req)
		s.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	s.logger.Printf("[INFO] service: Bulk proposal data has been applyed: %v", req)

	resp = future.Response().(*protobuf.BulkResponse)

	return resp, nil
}

func (s *KVSService) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	start := time.Now()
	defer Metrics(start, "Search")

	var err error

	resp := &protobuf.SearchResponse{}

	var searcher *bleve.Searcher
	if searcher, err = s.index.Searcher(); err != nil {
		message := "Failed to get searcher"
		s.logger.Printf("[ERR] service: %s, %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	defer searcher.Close()

	var searchRequest interface{}
	if searchRequest, err = protobuf.MarshalAny(req.SearchRequest); err != nil {
		message := "Failed to marshal search request to any"
		s.logger.Printf("[ERR] service: %s, %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	var searchRequestBytes []byte
	if searchRequestBytes, err = json.Marshal(searchRequest.(*map[string]interface{})); err != nil {
		message := "Failed to marshal search request"
		s.logger.Printf("[ERR] service: %s, %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	var searchResultBytes []byte
	if searchResultBytes, err = searcher.Search(searchRequestBytes); err != nil {
		message := "Failed to search documents"
		s.logger.Printf("[ERR] service: %s, %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	var searchResult map[string]interface{}
	if err = json.Unmarshal(searchResultBytes, &searchResult); err != nil {
		message := "Failed to unmarshal search result to map"
		s.logger.Printf("[ERR] service: %s, %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	searchResultAny := &any.Any{}
	if err = protobuf.UnmarshalAny(searchResult, searchResultAny); err != nil {
		message := "Failed to unmarshal search result to any"
		s.logger.Printf("[ERR] service: %s, %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	s.logger.Print("[INFO] service: Documents has been searched")

	resp.SearchResult = searchResultAny
	resp.Success = true

	return resp, nil
}

// ----------------------------------------

type fsm KVSService

// Apply applies a Raft log entry to the key-value store.
func (f *fsm) Apply(l *raft.Log) interface{} {
	start := time.Now()
	defer Metrics(start, "Apply")

	var err error

	proposal := &protobuf.Proposal{}
	if err = proto.Unmarshal(l.Data, proposal); err != nil {
		f.logger.Printf("[ERR] service: Failed to unmarshalling proposal: %v: %v", l.Data, err)
		return err
	}

	var data interface{}
	if data, err = protobuf.MarshalAny(proposal.Data); err != nil {
		f.logger.Printf("[ERR] service: Failed to marshalling proposal data: %v: %v", l.Data, err)
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
		f.logger.Printf("[ERR] service: Failed to applying proposal data: %v: %v", proposal, err)
		return err
	}
}

// Snapshot returns a snapshot of the key-value store.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	start := time.Now()
	defer Metrics(start, "Snapshot")

	var err error

	f.logger.Print("[INFO] service: Start making a snapshot of the store")

	f.mutex.Lock()
	defer f.mutex.Unlock()

	var iterator *boltdb.Iterator
	if iterator, err = f.store.Iterator(); err != nil {
		f.logger.Printf("[ERR] service: Failed to get iterator: %s", err.Error())
		return nil, err
	}
	defer iterator.Close()

	// Clone the map.
	data := make(map[string][]byte)
	for iterator.Next() {
		data[string(iterator.Key())] = iterator.Value()
	}

	f.logger.Print("[INFO] service: Finished making a snapshot of the store")
	return &fsmSnapshot{
		logger: f.logger,
		data:   data,
	}, nil
}

// Restore stores the key-value store to a previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	start := time.Now()
	defer Metrics(start, "Restore")

	var err error

	f.logger.Print("[INFO] service: Start restoring store from a previous state")

	data := make(map[string][]byte)
	if err = json.NewDecoder(rc).Decode(&data); err != nil {
		return err
	}

	batchSize := 1000

	// Restore store
	var writer *boltdb.Bulker
	if writer, err = f.store.Bulker(batchSize); err != nil {
		f.logger.Printf("[ERR] service: Failed to get store bulker: %v", err)
		return err
	}
	defer writer.Close()

	var indexer *bleve.Bulker
	if indexer, err = f.index.Bulker(batchSize); err != nil {
		f.logger.Printf("[ERR] service: Failed to get index bulker: %v", err)
		return err
	}
	defer indexer.Close()

	for key, value := range data {
		// Restore store
		if err = writer.Put([]byte(key), value); err != nil {
			f.logger.Printf("[ERR] service: Failed to put document: %s: %v: %v", key, value, err)
		}

		// Restore index
		var fields map[string]interface{}
		if err = json.Unmarshal(value, &fields); err != nil {
			f.logger.Printf("[ERR] service: Failed to unmarshal fields to map: %s: %v: %v", key, value, err)
		}
		if err = indexer.Index(key, fields); err != nil {
			f.logger.Printf("[ERR] service: Failed to index document: %s: %v: %v", key, value, err)
		}
	}

	f.logger.Print("[INFO] service: Finished restoring store from a previous state")
	return nil
}

func (f *fsm) applyPut(req *protobuf.PutRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyPut")

	var err error

	resp := &protobuf.PutResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Put into store
	var writer *boltdb.Writer
	if writer, err = f.store.Writer(); err != nil {
		message := "Failed to get writer"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	defer writer.Close()

	var fieldsBytes []byte
	if fieldsBytes, err = req.GetFieldsBytes(); err != nil {
		message := "Failed to get fields as bytes"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	if err = writer.Put([]byte(req.Id), fieldsBytes); err != nil {
		message := "Failed to put data into store"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	// Index into index
	var indexer *bleve.Indexer
	if indexer, err = f.index.Indexer(); err != nil {
		message := "Failed to get indexer"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	defer indexer.Close()

	var fieldsMap map[string]interface{}
	if fieldsMap, err = req.GetFieldsMap(); err != nil {
		message := "Failed to get fields as map"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	if err = indexer.Index(req.Id, fieldsMap); err != nil {
		message := "Failed to index data into index"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	resp.Success = true
	return resp, nil
}

func (f *fsm) applyDelete(req *protobuf.DeleteRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyDelete")

	var err error

	resp := &protobuf.DeleteResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Delete from store
	var writer *boltdb.Writer
	if writer, err = f.store.Writer(); err != nil {
		message := "Failed to get writer"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	defer writer.Close()

	if err = writer.Delete([]byte(req.Id)); err != nil {
		message := "Failed to delete data from store"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	// Delete from index
	var indexer *bleve.Indexer
	if indexer, err = f.index.Indexer(); err != nil {
		message := "Failed to get indexer"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	defer indexer.Close()

	if err = indexer.Delete(req.Id); err != nil {
		message := "Failed to delete data from index"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
		return resp, err
	}

	resp.Success = true
	return resp, nil
}

func (f *fsm) applyBulk(req *protobuf.BulkRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyBulk")

	var err error

	resp := &protobuf.BulkResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Store bulker
	var writer *boltdb.Bulker
	if writer, err = f.store.Bulker(int(req.BatchSize)); err != nil {
		message := "Failed to get writer"
		f.logger.Print(fmt.Sprintf("[ERR] service: %s: %v", message, err))
		resp.Success = false
		resp.Message = message
		return resp, err
	}
	defer writer.Close()

	// Index bulker
	var indexer *bleve.Bulker
	if indexer, err = f.index.Bulker(int(req.BatchSize)); err != nil {
		message := "Failed to index indexer"
		f.logger.Printf("[ERR] service: %s: %v", message, err)
		resp.Success = false
		resp.Message = message
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
			var fieldsBytes []byte
			if fieldsBytes, err = updateRequest.Document.GetFieldsBytes(); err != nil {
				f.logger.Printf("[WARN] service: Failed to get fields as bytes: %v", err)
				continue
			}

			if err = writer.Put([]byte(updateRequest.Document.Id), fieldsBytes); err != nil {
				f.logger.Printf("[WARN] service: Failed to put document into store: %v", err)
				continue
			}

			var fieldsMap map[string]interface{}
			if fieldsMap, err = updateRequest.Document.GetFieldsMap(); err != nil {
				f.logger.Printf("[WARN] service: Failed to get fields as map: %v", err)
				continue
			}

			if err = indexer.Index(updateRequest.Document.Id, fieldsMap); err != nil {
				f.logger.Printf("[WARN] service: Failed to index document into index: %v", err)
				continue
			}

			putCnt++
		case protobuf.UpdateRequest_DELETE:
			if err = writer.Delete([]byte(updateRequest.Document.Id)); err != nil {
				f.logger.Printf("[WARN] service: Failed to delete document from store: %v", err)
				continue
			}

			if err = indexer.Delete(updateRequest.Document.Id); err != nil {
				f.logger.Printf("[WARN] service: Failed to delete document from index: %v", err)
				continue
			}

			deleteCnt++
		default:
			f.logger.Printf("[WARN] service: Skip bulk operation: %s", updateRequest.Type.String())
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

	if _, ok := f.metadata[req.NodeId]; ok {
		message := fmt.Sprintf("Node already exist in metadata, ignoring join request: %v", req)
		f.logger.Printf("[INFO] service: %s", message)
		resp.Success = true
		resp.Message = message

		return resp, nil
	}

	f.metadata[req.NodeId] = req.Metadata

	resp.Success = true

	return resp, nil
}

func (f *fsm) applyLeave(req *protobuf.LeaveRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyLeave")

	resp := &protobuf.LeaveResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if _, ok := f.metadata[req.NodeId]; ok {
		message := fmt.Sprintf("Node does not exist in metadata, ignoring leave request: %v", req)
		f.logger.Printf("[INFO] service: %s", message)
		resp.Success = true
		resp.Message = message

		return resp, nil
	}

	delete(f.metadata, req.NodeId)

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

	var err error

	f.logger.Print("[INFO] service: Start to persist snapshot")

	if err = func() error {
		// Encode data.
		b, err := json.Marshal(f.data)
		if err != nil {
			return err
		}

		// Write data to sink.
		if _, err := sink.Write(b); err != nil {
			return err
		}

		// Close the sink.
		return sink.Close()
	}(); err != nil {
		sink.Cancel()
		return err
	}

	f.logger.Print("[INFO] service: Snapshot has been persisted")

	return nil
}

func (f *fsmSnapshot) Release() {}
