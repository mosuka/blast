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
	"log"

	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/http"
)

type Server struct {
	managerAddr string
	clusterId   string

	grpcAddr string
	httpAddr string

	grpcService *GRPCService
	grpcServer  *grpc.Server
	httpRouter  *http.Router
	httpServer  *http.Server

	logger     *log.Logger
	httpLogger accesslog.Logger
}

func NewServer(managerAddr string, grpcAddr string, httpAddr string, logger *log.Logger, httpLogger accesslog.Logger) (*Server, error) {
	return &Server{
		managerAddr: managerAddr,

		grpcAddr: grpcAddr,
		httpAddr: httpAddr,

		logger:     logger,
		httpLogger: httpLogger,
	}, nil
}

func (s *Server) Start() {
	s.logger.Printf("[INFO] start coordinator")

	var err error

	// create gRPC service
	s.grpcService, err = NewGRPCService(s.managerAddr, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create gRPC server
	s.grpcServer, err = grpc.NewServer(s.grpcAddr, s.grpcService, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create HTTP router
	s.httpRouter, err = NewRouter(s.grpcAddr, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create HTTP server
	s.httpServer, err = http.NewServer(s.httpAddr, s.httpRouter, s.logger, s.httpLogger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// start gRPC service
	s.logger.Print("[INFO] start gRPC service")
	go func() {
		err := s.grpcService.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}()

	// start gRPC server
	s.logger.Print("[INFO] start gRPC server")
	go func() {
		err := s.grpcServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}()

	// start HTTP server
	s.logger.Print("[INFO] start HTTP server")
	go func() {
		_ = s.httpServer.Start()
	}()
}

func (s *Server) Stop() {
	// stop HTTP server
	s.logger.Printf("[INFO] stop HTTP server")
	err := s.httpServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop HTTP router
	err = s.httpRouter.Close()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop gRPC server
	s.logger.Printf("[INFO] stop gRPC server")
	err = s.grpcServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop gRPC service
	s.logger.Print("[INFO] stop gRPC service")
	err = s.grpcService.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
}
