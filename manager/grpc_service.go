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
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/mosuka/blast/protobuf/raft"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	raftServer *RaftServer

	logger *log.Logger
}

func NewGRPCService(raftServer *RaftServer, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: raftServer,
		logger:     logger,
	}, nil
}

func (s *GRPCService) Join(ctx context.Context, req *raft.Node) (*empty.Empty, error) {
	s.logger.Printf("[INFO] %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Join(req)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) Leave(ctx context.Context, req *raft.Node) (*empty.Empty, error) {
	s.logger.Printf("[INFO] leave %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Leave(req)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetNode(ctx context.Context, req *empty.Empty) (*raft.Node, error) {
	s.logger.Printf("[INFO] get node %v", req)

	resp := &raft.Node{}

	var err error

	resp, err = s.raftServer.GetNode()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*raft.Cluster, error) {
	s.logger.Printf("[INFO] get cluster %v", req)

	resp := &raft.Cluster{}

	var err error

	resp, err = s.raftServer.GetCluster()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	s.logger.Printf("[INFO] %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Snapshot()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) Get(ctx context.Context, req *management.KeyValuePair) (*management.KeyValuePair, error) {
	start := time.Now()
	defer RecordMetrics(start, "get")

	s.logger.Printf("[INFO] get %v", req)

	resp := &management.KeyValuePair{}

	var err error

	resp, err = s.raftServer.Get(req)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (s *GRPCService) Set(ctx context.Context, req *management.KeyValuePair) (*empty.Empty, error) {
	start := time.Now()
	defer RecordMetrics(start, "set")

	s.logger.Printf("[INFO] set %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Set(req)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (s *GRPCService) Delete(ctx context.Context, req *management.KeyValuePair) (*empty.Empty, error) {
	start := time.Now()
	defer RecordMetrics(start, "set")

	s.logger.Printf("[INFO] set %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Delete(req)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}
