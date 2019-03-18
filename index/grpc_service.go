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

package index

import (
	"context"
	"log"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/raft"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	raftServer *RaftServer

	logger *log.Logger
}

func NewGRPCService(store *RaftServer, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: store,
		logger:     logger,
	}, nil
}

func (s *GRPCService) Join(ctx context.Context, req *raft.JoinRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] join %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Join(req.Node)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) Leave(ctx context.Context, req *raft.LeaveRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] leave %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Leave(req.Node)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetNode(ctx context.Context, req *empty.Empty) (*raft.GetNodeResponse, error) {
	s.logger.Printf("[INFO] get node %v", req)

	resp := &raft.GetNodeResponse{}

	var err error

	resp.Node, err = s.raftServer.GetNode()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*raft.GetClusterResponse, error) {
	s.logger.Printf("[INFO] get cluster %v", req)

	resp := &raft.GetClusterResponse{}

	var err error

	resp.Cluster, err = s.raftServer.GetCluster()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	resp := &empty.Empty{}

	s.logger.Printf("[INFO] %v", req)

	err := s.raftServer.Snapshot()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) Get(ctx context.Context, req *index.GetRequest) (*index.GetResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "get")

	resp := &index.GetResponse{}

	s.logger.Printf("[INFO] get %v", req)

	doc, err := s.raftServer.Get(req.Document)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	resp.Document = doc

	return resp, nil
}

func (s *GRPCService) Search(ctx context.Context, req *index.SearchRequest) (*index.SearchResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "search")

	resp := &index.SearchResponse{}

	s.logger.Printf("[INFO] search %v", req)

	// Any -> bleve.SearchRequest
	searchRequest, err := protobuf.MarshalAny(req.SearchRequest)
	if err != nil {
		return resp, status.Error(codes.InvalidArgument, err.Error())
	}

	searchResult, err := s.raftServer.Search(searchRequest.(*bleve.SearchRequest))
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	// bleve.SearchResult -> Any
	searchResultAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchResult, searchResultAny)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.SearchResult = searchResultAny

	return resp, nil
}

func (s *GRPCService) Index(ctx context.Context, req *index.IndexRequest) (*empty.Empty, error) {
	start := time.Now()
	defer RecordMetrics(start, "index")

	s.logger.Printf("[DEBUG] index %v", req)

	resp := &empty.Empty{}

	// index
	err := s.raftServer.Index(req.Document)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) Delete(ctx context.Context, req *index.DeleteRequest) (*empty.Empty, error) {
	start := time.Now()
	defer RecordMetrics(start, "delete")

	resp := &empty.Empty{}

	s.logger.Printf("[INFO] delete %v", req)

	err := s.raftServer.Delete(req.Document)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetStats(ctx context.Context, req *empty.Empty) (*index.GetStatsResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "stats")

	resp := &index.GetStatsResponse{}

	s.logger.Printf("[INFO] stats %v", req)

	stats, err := s.raftServer.Stats()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Stats = stats

	return resp, nil
}
