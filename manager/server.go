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
	"github.com/blevesearch/bleve/mapping"
	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/protobuf/management"
	"go.uber.org/zap"
)

type Server struct {
	peerGrpcAddr     string
	node             *management.Node
	dataDir          string
	raftStorageType  string
	indexMapping     *mapping.IndexMappingImpl
	indexType        string
	indexStorageType string
	logger           *zap.Logger
	grpcLogger       *zap.Logger
	httpLogger       accesslog.Logger

	raftServer  *RaftServer
	grpcService *GRPCService
	grpcServer  *GRPCServer
	grpcGateway *GRPCGateway
}

func NewServer(peerGrpcAddr string, node *management.Node, dataDir string, raftStorageType string, indexMapping *mapping.IndexMappingImpl, indexType string, indexStorageType string, logger *zap.Logger, grpcLogger *zap.Logger, httpLogger accesslog.Logger) (*Server, error) {
	return &Server{
		peerGrpcAddr:     peerGrpcAddr,
		node:             node,
		dataDir:          dataDir,
		raftStorageType:  raftStorageType,
		indexMapping:     indexMapping,
		indexType:        indexType,
		indexStorageType: indexStorageType,
		logger:           logger,
		grpcLogger:       grpcLogger,
		httpLogger:       httpLogger,
	}, nil
}

func (s *Server) Start() {
	var err error

	// bootstrap node?
	bootstrap := s.peerGrpcAddr == ""
	s.logger.Info("bootstrap", zap.Bool("bootstrap", bootstrap))

	// create raft server
	s.raftServer, err = NewRaftServer(s.node, s.dataDir, s.raftStorageType, s.indexMapping, s.indexType, s.indexStorageType, bootstrap, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create gRPC service
	s.grpcService, err = NewGRPCService(s.raftServer, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create gRPC server
	s.grpcServer, err = NewGRPCServer(s.node.Metadata.GrpcAddress, s.grpcService, s.grpcLogger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create HTTP server
	s.grpcGateway, err = NewGRPCGateway(s.node.Metadata.HttpAddress, s.node.Metadata.GrpcAddress, s.logger)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	// start Raft server
	s.logger.Info("start Raft server")
	err = s.raftServer.Start()
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// start gRPC service
	s.logger.Info("start gRPC service")
	go func() {
		err := s.grpcService.Start()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}
	}()

	// start gRPC server
	s.logger.Info("start gRPC server")
	go func() {
		err := s.grpcServer.Start()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}
	}()

	// start gRPC gateway
	s.logger.Info("start gRPC gateway")
	go func() {
		_ = s.grpcGateway.Start()
	}()

	// join to the existing cluster
	if !bootstrap {
		client, err := NewGRPCClient(s.peerGrpcAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Error(err.Error())
			}
		}()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		err = client.ClusterJoin(s.node)
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}
	}
}

func (s *Server) Stop() {
	s.logger.Info("stop gRPC gateway")
	err := s.grpcGateway.Stop()
	if err != nil {
		s.logger.Error(err.Error())
	}

	s.logger.Info("stop gRPC server")
	err = s.grpcServer.Stop()
	if err != nil {
		s.logger.Error(err.Error())
	}

	s.logger.Info("stop gRPC service")
	err = s.grpcService.Stop()
	if err != nil {
		s.logger.Error(err.Error())
	}

	s.logger.Info("stop Raft server")
	err = s.raftServer.Stop()
	if err != nil {
		s.logger.Error(err.Error())
	}
}

func (s *Server) BindAddress() string {
	return s.raftServer.NodeAddress()
}

func (s *Server) GrpcAddress() string {
	address, err := s.grpcServer.GetAddress()
	if err != nil {
		return ""
	}

	return address
}

func (s *Server) HttpAddress() string {
	address, err := s.grpcGateway.GetAddress()
	if err != nil {
		return ""
	}

	return address
}
