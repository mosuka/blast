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
	"log"

	"github.com/blevesearch/bleve/mapping"
	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/protobuf/raft"
)

type ServerConfig struct {
}

type Server struct {
	node      *raft.Node
	bootstrap bool
	peerAddr  string

	managerAddr string
	clusterId   string

	raftServer *RaftServer

	grpcService *GRPCService
	grpcServer  *GRPCServer
	grpcClient  *GRPCClient

	httpServer *HTTPServer

	logger     *log.Logger
	httpLogger accesslog.Logger
}

func NewServer(managerAddr string, clusterId string, node *raft.Node, peerAddr string, indexMapping *mapping.IndexMappingImpl, indexType string, indexStorageType string, logger *log.Logger, httpLogger accesslog.Logger) (*Server, error) {
	var err error

	server := &Server{
		node:        node,
		bootstrap:   peerAddr == "",
		peerAddr:    peerAddr,
		managerAddr: managerAddr,
		clusterId:   clusterId,
		logger:      logger,
		httpLogger:  httpLogger,
	}

	// create raft server
	server.raftServer, err = NewRaftServer(server.node, server.bootstrap, indexMapping, indexStorageType, server.logger)
	if err != nil {
		return nil, err
	}

	// create gRPC service
	server.grpcService, err = NewGRPCService(server.raftServer, server.logger)
	if err != nil {
		return nil, err
	}

	// create gRPC server
	server.grpcServer, err = NewGRPCServer(node.Metadata.GrpcAddr, server.grpcService, server.logger)
	if err != nil {
		return nil, err
	}

	// create gRPC client for HTTP server
	server.grpcClient, err = NewGRPCClient(node.Metadata.GrpcAddr)
	if err != nil {
		return nil, err
	}

	// create HTTP server
	server.httpServer, err = NewHTTPServer(node.Metadata.HttpAddr, server.grpcClient, server.logger, server.httpLogger)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) Start() {
	// start Raft server
	go func() {
		err := s.raftServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}()
	s.logger.Print("[INFO] Raft server started")

	// start gRPC server
	go func() {
		err := s.grpcServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}()
	s.logger.Print("[INFO] gRPC server started")

	// start HTTP server
	go func() {
		err := s.httpServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}()
	s.logger.Print("[INFO] HTTP server started")

	// join to the existing cluster
	if !s.bootstrap {
		// create gRPC client
		client, err := NewGRPCClient(s.peerAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		// join to the existing cluster
		err = client.Join(s.node)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}
}

func (s *Server) Stop() {
	// stop HTTP server
	err := s.httpServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// close gRPC client
	err = s.grpcClient.Close()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop gRPC server
	err = s.grpcServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop Raft server
	err = s.raftServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
}
