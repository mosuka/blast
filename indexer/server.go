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

	"github.com/mosuka/blast/protobuf"

	"github.com/mosuka/blast/protobuf/management"

	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf/raft"
)

type ServerConfig struct {
}

type Server struct {
	managerAddr string
	clusterId   string

	node      *raft.Node
	bootstrap bool
	peerAddr  string

	indexConfig map[string]interface{}

	raftServer *RaftServer

	grpcService *GRPCService
	grpcServer  *GRPCServer
	grpcClient  *GRPCClient

	httpServer *HTTPServer

	logger     *log.Logger
	httpLogger accesslog.Logger
}

func NewServer(managerAddr string, clusterId string, node *raft.Node, peerAddr string, indexConfig map[string]interface{}, logger *log.Logger, httpLogger accesslog.Logger) (*Server, error) {
	//var err error

	server := &Server{
		node:        node,
		bootstrap:   peerAddr == "",
		peerAddr:    peerAddr,
		indexConfig: indexConfig,
		managerAddr: managerAddr,
		clusterId:   clusterId,
		logger:      logger,
		httpLogger:  httpLogger,
	}

	return server, nil
}

func (s *Server) Start() {
	// get index config from leader node
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
			s.logger.Printf("[ERR] %v", err)
			return
		}
		kvp, err := mc.Get(&management.KeyValuePair{Key: "/index_config"})
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
		ins, err := protobuf.MarshalAny(kvp.Value)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
		if ins != nil {
			s.indexConfig = *ins.(*map[string]interface{})
		}
	}

	var err error

	// create raft server
	s.raftServer, err = NewRaftServer(s.node, s.bootstrap, s.indexConfig, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create gRPC service
	s.grpcService, err = NewGRPCService(s.raftServer, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create gRPC server
	s.grpcServer, err = NewGRPCServer(s.node.Metadata.GrpcAddr, s.grpcService, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create gRPC client for HTTP server
	s.grpcClient, err = NewGRPCClient(s.node.Metadata.GrpcAddr)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create HTTP server
	s.httpServer, err = NewHTTPServer(s.node.Metadata.HttpAddr, s.grpcClient, s.logger, s.httpLogger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// start Raft server
	go func() {
		var err error

		// start raft server
		err = s.raftServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		s.logger.Print("[INFO] Raft server started")
	}()

	// start gRPC server
	go func() {
		var err error

		err = s.grpcServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		s.logger.Print("[INFO] gRPC server started")
	}()

	// start HTTP server
	go func() {
		var err error

		err = s.httpServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		s.logger.Print("[INFO] HTTP server started")
	}()

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
