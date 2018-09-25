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

	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/config"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	_ "github.com/mosuka/blast/encoding"
	"github.com/mosuka/blast/index"
	blastlog "github.com/mosuka/blast/log"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	blastraft "github.com/mosuka/blast/raft"
	"github.com/mosuka/blast/store"
	"github.com/pkg/errors"
)

type Service struct {
	raftAddress string
	raftConfig  *blastraft.RaftConfig
	bootstrap   bool
	raft        *raft.Raft

	storeConfig *store.StoreConfig
	store       *store.Store

	indexConfig *index.IndexConfig
	index       *index.Index

	mutex sync.Mutex

	metadataMap map[string]*protobuf.Metadata
	//metadata map[string]map[string]interface{}
	//cluster *protobuf.Cluster

	logger *log.Logger
}

func NewService(raftAddress string, raftConfig *blastraft.RaftConfig, bootstrap bool, storeConfig *store.StoreConfig, indexConfig *index.IndexConfig) (*Service, error) {
	return &Service{
		raftAddress: raftAddress,
		raftConfig:  raftConfig,
		bootstrap:   bootstrap,

		storeConfig: storeConfig,
		indexConfig: indexConfig,

		metadataMap: make(map[string]*protobuf.Metadata, 0),
		//metadata: make(map[string]map[string]interface{}, 0),
		//cluster: &protobuf.Cluster{
		//	Nodes: make([]*protobuf.Node, 0),
		//},

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
	peerAddr, err := net.ResolveTCPAddr("tcp", s.raftAddress)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}
	transport, err := raft.NewTCPTransportWithLogger(s.raftAddress, peerAddr, 3, s.raftConfig.Timeout, s.logger)
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
		s.raft.BootstrapCluster(configuration)
	}

	// TODO
	// start connection monitoring

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

func (s *Service) PutMetadata(metadata *protobuf.Metadata) error {
	data := &any.Any{}
	err := protobuf.UnmarshalAny(metadata, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_PUT_METADATA,
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

func (s *Service) GetMetadata(id string) (*protobuf.Metadata, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	values, exist := s.metadataMap[id]
	if !exist {
		err := fmt.Errorf("%s does not exist in metadata", id)
		s.logger.Printf("[ERR] %v", err)
		return nil, err
	}

	return values, nil
}

func (s *Service) DeleteMetadata(metadata *protobuf.Metadata) error {
	data := &any.Any{}
	err := protobuf.UnmarshalAny(metadata, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_DELETE_METADATA,
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

func (s *Service) PutNode(ctx context.Context, req *protobuf.PutNodeRequest) (*empty.Empty, error) {
	start := time.Now()
	defer Metrics(start, "PutNode")

	resp := &empty.Empty{}

	if s.raft.State() != raft.Leader {
		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		grpcClient, err := client.NewGRPCClient(leaderMetadata.GrpcAddr)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.PutNode(req)
	}

	peerAddr, err := net.ResolveTCPAddr("tcp", req.RaftAddr)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	configFuture := s.raft.GetConfiguration()
	err = configFuture.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(req.Id) || srv.Address == raft.ServerAddress(peerAddr.String()) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.ID == raft.ServerID(req.Id) && srv.Address == raft.ServerAddress(peerAddr.String()) {
				message := fmt.Sprintf("%s (%s) already exists in cluster, ignoring join request", req.Id, srv.Address)
				s.logger.Printf("[DEBUG] %s", message)
				return &empty.Empty{}, nil
			}

			// If either the ID or the address matches, need to remove the corresponding node by the ID.
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			err = future.Error()
			if err != nil {
				s.logger.Print(fmt.Sprintf("[ERR] %v", err))
				return &empty.Empty{}, err
			}
		}
	}

	future := s.raft.AddVoter(raft.ServerID(req.Id), raft.ServerAddress(peerAddr.String()), 0, 0)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	err = s.PutMetadata(
		&protobuf.Metadata{
			Id:       req.Id,
			RaftAddr: req.RaftAddr,
			GrpcAddr: req.GrpcAddr,
			HttpAddr: req.HttpAddr,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	return resp, nil
}

func (s *Service) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	start := time.Now()
	defer Metrics(start, "GetNode")

	resp := &protobuf.GetNodeResponse{}

	configFuture := s.raft.GetConfiguration()
	err := configFuture.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	leaderAddr, err := s.LeaderAddress(60 * time.Second)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	node := &protobuf.Node{}
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(req.Id) {
			metadata, err := s.GetMetadata(string(srv.ID))
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return resp, err
			}

			node.Id = string(srv.ID)
			node.RaftAddr = string(srv.Address)
			node.GrpcAddr = metadata.GrpcAddr
			node.HttpAddr = metadata.HttpAddr
			node.Leader = srv.Address == leaderAddr
			break
		}
	}

	resp.Node = node

	return resp, nil
}

func (s *Service) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	start := time.Now()
	defer Metrics(start, "DeleteNode")

	resp := &empty.Empty{}

	if s.raft.State() != raft.Leader {
		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		grpcClient, err := client.NewGRPCClient(leaderMetadata.GrpcAddr)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.DeleteNode(req)
	}

	configFuture := s.raft.GetConfiguration()
	err := configFuture.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	peerAddr, err := net.ResolveTCPAddr("tcp", req.RaftAddr)
	if err != nil {
		s.logger.Print(fmt.Sprintf("[ERR] %v", err))
		return resp, err
	}

	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == raft.ServerID(req.Id) || srv.Address == raft.ServerAddress(peerAddr.String()) {
			future := s.raft.RemoveServer(srv.ID, 0, 0)
			err = future.Error()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return resp, err
			}
		}
	}

	future := s.raft.DemoteVoter(raft.ServerID(req.Id), 0, 0)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err

	}

	err = s.DeleteMetadata(
		&protobuf.Metadata{
			Id:       req.Id,
			RaftAddr: req.RaftAddr,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	return resp, nil
}

func (s *Service) GetCluster(ctx context.Context, req *empty.Empty) (*protobuf.GetClusterResponse, error) {
	start := time.Now()
	defer Metrics(start, "GetCluster")

	resp := &protobuf.GetClusterResponse{}

	configFuture := s.raft.GetConfiguration()
	err := configFuture.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	leaderAddr, err := s.LeaderAddress(60 * time.Second)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	cluster := &protobuf.Cluster{
		Nodes:  make([]*protobuf.Node, 0),
		Health: protobuf.Cluster_UNKNOWN_HEALTH,
	}

	for _, srv := range configFuture.Configuration().Servers {
		metadata, err := s.GetMetadata(string(srv.ID))
		if err != nil {
			return resp, err
		}

		node := &protobuf.Node{
			Id:       string(srv.ID),
			RaftAddr: string(srv.Address),
			GrpcAddr: metadata.GrpcAddr,
			HttpAddr: metadata.HttpAddr,
			Leader:   srv.Address == leaderAddr,
		}
		cluster.Nodes = append(cluster.Nodes, node)
	}

	resp.Cluster = cluster

	return resp, nil
}

func (s *Service) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	start := time.Now()
	defer Metrics(start, "Snapshot")

	resp := &empty.Empty{}

	future := s.raft.Snapshot()
	err := future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	return resp, nil
}

func (s *Service) PutDocument(ctx context.Context, req *protobuf.PutDocumentRequest) (*empty.Empty, error) {
	start := time.Now()
	defer Metrics(start, "PutDocument")

	resp := &empty.Empty{}

	if s.raft.State() != raft.Leader {
		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		grpcClient, err := client.NewGRPCClient(leaderMetadata.GrpcAddr)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.PutDocument(req)
	}

	// PutRequest to Any
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_PUT_DOCUMENT,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	resp = future.Response().(*empty.Empty)

	return resp, nil
}

func (s *Service) GetDocument(ctx context.Context, req *protobuf.GetDocumentRequest) (*protobuf.GetDocumentResponse, error) {
	start := time.Now()
	defer Metrics(start, "Get")

	resp := &protobuf.GetDocumentResponse{}

	reader, err := s.store.Reader()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}
	defer reader.Close()

	fieldsBytes, err := reader.Get([]byte(req.Id))
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	// document does not exist
	if fieldsBytes == nil {
		return resp, nil
	}

	var fieldsMap map[string]interface{}
	err = json.Unmarshal(fieldsBytes, &fieldsMap)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	fieldsAny := &any.Any{}
	err = protobuf.UnmarshalAny(fieldsMap, fieldsAny)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	resp.Document = &protobuf.Document{
		Id:     req.Id,
		Fields: fieldsAny,
	}

	return resp, nil
}

func (s *Service) DeleteDocument(ctx context.Context, req *protobuf.DeleteDocumentRequest) (*empty.Empty, error) {
	start := time.Now()
	defer Metrics(start, "DeleteDocument")

	resp := &empty.Empty{}

	if s.raft.State() != raft.Leader {
		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		grpcClient, err := client.NewGRPCClient(leaderMetadata.GrpcAddr)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.DeleteDocument(req)
	}

	// DeleteRequest to Any
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_DELETE_DOCUMENT,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	resp = future.Response().(*empty.Empty)

	return resp, nil
}

func (s *Service) BulkUpdate(ctx context.Context, req *protobuf.BulkUpdateRequest) (*protobuf.BulkUpdateResponse, error) {
	start := time.Now()
	defer Metrics(start, "BulkUpdate")

	resp := &protobuf.BulkUpdateResponse{}

	if s.raft.State() != raft.Leader {
		leaderID, err := s.LeaderID(60 * time.Second)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		leaderMetadata, err := s.GetMetadata(string(leaderID))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}

		grpcClient, err := client.NewGRPCClient(leaderMetadata.GrpcAddr)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return resp, err
		}
		defer grpcClient.Close()

		return grpcClient.BulkUpdate(req)
	}

	// BulkRequest to Any
	data := &any.Any{}
	err := protobuf.UnmarshalAny(req, data)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	proposalBytes, err := proto.Marshal(
		&protobuf.Proposal{
			Type: protobuf.Proposal_BULK_UPDATE,
			Data: data,
		},
	)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	future := s.raft.Apply(proposalBytes, s.raftConfig.Timeout)
	err = future.Error()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	resp = future.Response().(*protobuf.BulkUpdateResponse)

	return resp, nil
}

func (s *Service) SearchDocuments(ctx context.Context, req *protobuf.SearchDocumentsRequest) (*protobuf.SearchDocumentsResponse, error) {
	start := time.Now()
	defer Metrics(start, "SearchDocuments")

	resp := &protobuf.SearchDocumentsResponse{}

	searcher, err := s.index.Searcher()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}
	defer searcher.Close()

	searchRequest, err := protobuf.MarshalAny(req.SearchRequest)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	searchResult, err := searcher.Search(searchRequest.(*bleve.SearchRequest))
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	searchResultAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchResult, searchResultAny)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	resp.SearchResult = searchResultAny

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
	case protobuf.Proposal_PUT_METADATA:
		resp, err := f.applyPutMetadata(data.(*protobuf.Metadata))
		if err != nil {
			return err
		}
		return resp.(*empty.Empty)
	case protobuf.Proposal_DELETE_METADATA:
		resp, err := f.applyDeleteMetadata(data.(*protobuf.Metadata))
		if err != nil {
			return err
		}
		return resp.(*empty.Empty)
	case protobuf.Proposal_PUT_DOCUMENT:
		resp, err := f.applyPutDocument(data.(*protobuf.PutDocumentRequest))
		if err != nil {
			return err
		}
		return resp.(*empty.Empty)
	case protobuf.Proposal_DELETE_DOCUMENT:
		resp, err := f.applyDeleteDocument(data.(*protobuf.DeleteDocumentRequest))
		if err != nil {
			return err
		}
		return resp.(*empty.Empty)
	case protobuf.Proposal_BULK_UPDATE:
		resp, err := f.applyBulkUpdate(data.(*protobuf.BulkUpdateRequest))
		if err != nil {
			return err
		}
		return resp.(*protobuf.BulkUpdateResponse)
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

func (f *fsm) applyPutMetadata(metadata *protobuf.Metadata) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyPutMetadata")

	f.mutex.Lock()
	defer f.mutex.Unlock()

	resp := &empty.Empty{}

	for i, v := range f.metadataMap {
		// If a node already exists with either the putting node's ID or address,
		// that node may need to be removed from the metadata first.
		if v.Id == metadata.Id || v.RaftAddr == metadata.RaftAddr {
			delete(f.metadataMap, i)
			continue
		}
	}

	f.metadataMap[metadata.Id] = metadata

	f.logger.Printf("[DEBUG] Metadata = %v", f.metadataMap)

	return resp, nil
}

func (f *fsm) applyDeleteMetadata(metadata *protobuf.Metadata) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyDeleteMetadata")

	f.mutex.Lock()
	defer f.mutex.Unlock()

	resp := &empty.Empty{}

	for i, v := range f.metadataMap {
		// If a node already exists with either the putting node's ID or address,
		// that node may need to be removed from the metadata first.
		if v.Id == metadata.Id || v.RaftAddr == metadata.RaftAddr {
			delete(f.metadataMap, i)
			continue
		}
	}

	f.logger.Printf("[DEBUG] Metadata = %v", f.metadataMap)

	return resp, nil
}

func (f *fsm) applyPutDocument(req *protobuf.PutDocumentRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyPutDocument")

	f.mutex.Lock()
	defer f.mutex.Unlock()

	resp := &empty.Empty{}

	fieldsInstance, err := protobuf.MarshalAny(req.Fields)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	if fieldsInstance == nil {
		err := errors.New("Fields is nil")
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	fieldsMap := *fieldsInstance.(*map[string]interface{})

	fieldsBytes, err := json.Marshal(fieldsMap)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	// Put into store
	writer, err := f.store.Writer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}
	defer writer.Close()

	err = writer.Put([]byte(req.Id), fieldsBytes)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	// Index into index
	indexer, err := f.index.Indexer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}
	defer indexer.Close()

	err = indexer.Index(req.Id, fieldsMap)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	return resp, nil
}

func (f *fsm) applyDeleteDocument(req *protobuf.DeleteDocumentRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyDeleteDocument")

	f.mutex.Lock()
	defer f.mutex.Unlock()

	resp := &empty.Empty{}

	// Delete from store
	writer, err := f.store.Writer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}
	defer writer.Close()

	err = writer.Delete([]byte(req.Id))
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	// Delete from index
	indexer, err := f.index.Indexer()
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}
	defer indexer.Close()

	err = indexer.Delete(req.Id)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}

	return resp, nil
}

func (f *fsm) applyBulkUpdate(req *protobuf.BulkUpdateRequest) (interface{}, error) {
	start := time.Now()
	defer Metrics(start, "applyBulkUpdate")

	resp := &protobuf.BulkUpdateResponse{}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Store bulker
	writer, err := f.store.Bulker(int(req.BatchSize))
	if err != nil {
		f.logger.Print(fmt.Sprintf("[ERR] %v", err))
		return resp, err
	}
	defer writer.Close()

	// Index bulker
	indexer, err := f.index.Bulker(int(req.BatchSize))
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return resp, err
	}
	defer indexer.Close()

	var (
		putCnt    int
		deleteCnt int
	)

	for _, updateRequest := range req.Updates {
		switch updateRequest.Command {
		case protobuf.BulkUpdateRequest_Update_PUT:
			fieldsInstance, err := protobuf.MarshalAny(updateRequest.Document.Fields)
			if err != nil {
				f.logger.Printf("[ERR] %v", err)
			}

			if fieldsInstance == nil {
				err := errors.New("Fields is nil")
				f.logger.Printf("[ERR] %v", err)
				return resp, err
			}

			fieldsMap := *fieldsInstance.(*map[string]interface{})

			fieldsBytes, err := json.Marshal(fieldsMap)
			if err != nil {
				f.logger.Printf("[ERR] %v", err)
				return resp, err
			}

			err = writer.Put([]byte(updateRequest.Document.Id), fieldsBytes)
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
		case protobuf.BulkUpdateRequest_Update_DELETE:
			err = writer.Delete([]byte(updateRequest.Document.Id))
			if err != nil {
				f.logger.Printf("[ERR] %v", err)
				return resp, err
			}

			err = indexer.Delete(updateRequest.Document.Id)
			if err != nil {
				f.logger.Printf("[WARN] %v", err)
				continue
			}

			deleteCnt++
		default:
			f.logger.Printf("[WARN] Unknown command: %s", updateRequest.Command.String())
			continue
		}
	}

	resp.PutCount = int32(putCnt)
	resp.DeleteCount = int32(deleteCnt)

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
