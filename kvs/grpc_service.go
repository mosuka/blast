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
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/protobuf/kvs"
	"github.com/mosuka/blast/protobuf/raft"
)

type Service struct {
	// wrapper and manager for db instance
	store *RaftServer

	logger *log.Logger
}

func NewService(store *RaftServer, logger *log.Logger) (*Service, error) {
	return &Service{
		store:  store,
		logger: logger,
	}, nil
}

func (s *Service) Join(ctx context.Context, req *raft.JoinRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] %v", req)

	// join node to the cluster
	err := s.store.Join(req.Node.Id, req.Node.BindAddr)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Service) Leave(ctx context.Context, req *raft.LeaveRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] %v", req)

	// leave node from the cluster
	err := s.store.Leave(req.Node.Id)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Service) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	s.logger.Printf("[INFO] %v", req)

	// snapshot
	err := s.store.Snapshot()
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Service) Get(ctx context.Context, req *kvs.GetRequest) (*kvs.GetResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "get")

	resp := &kvs.GetResponse{}

	s.logger.Printf("[INFO] get %v", req)

	// get value by key
	value, err := s.store.Get(req.Key)
	if err != nil {
		return resp, err
	}
	if value == nil {
		// key does not exist
		return resp, nil
	}

	resp.Value = value

	return resp, nil
}

func (s *Service) Put(ctx context.Context, req *kvs.PutRequest) (*empty.Empty, error) {
	start := time.Now()
	defer RecordMetrics(start, "put")

	resp := &empty.Empty{}

	s.logger.Printf("[INFO] put %v", req)

	// put value by key
	err := s.store.Set(req.Key, req.Value)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, req *kvs.DeleteRequest) (*empty.Empty, error) {
	start := time.Now()
	defer RecordMetrics(start, "delete")

	resp := &empty.Empty{}

	s.logger.Printf("[INFO] delete %v", req)

	// delete value by key
	err := s.store.Delete(req.Key)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
