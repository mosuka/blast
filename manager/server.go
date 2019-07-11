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
	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/http"
	"github.com/mosuka/blast/maputils"
	"go.uber.org/zap"
)

type Server struct {
	clusterConfig config.ClusterConfig
	nodeConfig    config.NodeConfig
	indexConfig   config.IndexConfig

	logger     *zap.Logger
	grpcLogger *zap.Logger
	httpLogger accesslog.Logger

	raftServer  *RaftServer
	grpcService *GRPCService
	grpcServer  *grpc.Server
	httpRouter  *http.Router
	httpServer  *http.Server
}

func NewServer(clusterConfig config.ClusterConfig, nodeConfig config.NodeConfig, indexConfig config.IndexConfig, logger *zap.Logger, grpcLogger *zap.Logger, httpLogger accesslog.Logger) (*Server, error) {
	return &Server{
		clusterConfig: clusterConfig,
		nodeConfig:    nodeConfig,
		indexConfig:   indexConfig,
		logger:        logger,
		grpcLogger:    grpcLogger,
		httpLogger:    httpLogger,
	}, nil
}

func (s *Server) Start() {
	var err error

	peerAddr, err := s.clusterConfig.GetPeerAddr()
	if err != nil && err != maputils.ErrNotFound {
		s.logger.Fatal(err.Error())
		return
	}

	// bootstrap node?
	bootstrap := peerAddr == ""
	s.logger.Info("bootstrap", zap.Bool("bootstrap", bootstrap))

	// create raft server
	s.raftServer, err = NewRaftServer(s.nodeConfig, s.indexConfig, bootstrap, s.logger)
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

	grpcAddr, err := s.nodeConfig.GetGrpcAddr()
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create gRPC server
	s.grpcServer, err = grpc.NewServer(grpcAddr, s.grpcService, s.grpcLogger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create HTTP router
	s.httpRouter, err = NewRouter(grpcAddr, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	httpAddr, err := s.nodeConfig.GetHttpAddr()
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create HTTP server
	s.httpServer, err = http.NewServer(httpAddr, s.httpRouter, s.logger, s.httpLogger)
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

	// start HTTP server
	s.logger.Info("start HTTP server")
	go func() {
		_ = s.httpServer.Start()
	}()

	// join to the existing cluster
	if !bootstrap {
		client, err := grpc.NewClient(peerAddr)
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

		nodeId, err := s.nodeConfig.GetNodeId()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		err = client.SetNode(nodeId, s.nodeConfig.ToMap())
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}
	}
}

func (s *Server) Stop() {
	s.logger.Info("stop HTTP server")
	err := s.httpServer.Stop()
	if err != nil {
		s.logger.Error(err.Error())
	}

	err = s.httpRouter.Close()
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
