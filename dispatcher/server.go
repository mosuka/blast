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

package dispatcher

import (
	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/config"
	"go.uber.org/zap"
)

type Server struct {
	clusterConfig *config.ClusterConfig
	nodeConfig    *config.NodeConfig
	logger        *zap.Logger
	grpcLogger    *zap.Logger
	httpLogger    accesslog.Logger

	grpcService *GRPCService
	grpcServer  *GRPCServer
	httpRouter  *Router
	httpServer  *HTTPServer
}

func NewServer(clusterConfig *config.ClusterConfig, nodeConfig *config.NodeConfig, logger *zap.Logger, grpcLogger *zap.Logger, httpLogger accesslog.Logger) (*Server, error) {
	return &Server{
		clusterConfig: clusterConfig,
		nodeConfig:    nodeConfig,
		logger:        logger,
		grpcLogger:    grpcLogger,
		httpLogger:    httpLogger,
	}, nil
}

func (s *Server) Start() {
	var err error

	// create gRPC service
	s.grpcService, err = NewGRPCService(s.clusterConfig.ManagerAddr, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create gRPC server
	s.grpcServer, err = NewGRPCServer(s.nodeConfig.GRPCAddr, s.grpcService, s.grpcLogger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create HTTP router
	s.httpRouter, err = NewRouter(s.nodeConfig.GRPCAddr, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create HTTP server
	s.httpServer, err = NewHTTPServer(s.nodeConfig.HTTPAddr, s.httpRouter, s.logger, s.httpLogger)
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
}
