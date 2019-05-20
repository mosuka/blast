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

package federator

import (
	"log"

	accesslog "github.com/mash/go-accesslog"
)

type Server struct {
	managerAddr string
	clusterId   string

	grpcAddr string
	httpAddr string

	grpcServer *GRPCServer
	httpServer *HTTPServer

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

	// create gRPC server
	s.grpcServer, err = NewGRPCServer(s.managerAddr, s.grpcAddr, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create HTTP server
	s.httpServer, err = NewHTTPServer(s.httpAddr, s.grpcAddr, s.logger, s.httpLogger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

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
	s.logger.Printf("[INFO] stop coordinator")

	// stop HTTP server
	s.logger.Printf("[INFO] stop HTTP server")
	err := s.httpServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop gRPC server
	s.logger.Printf("[INFO] stop gRPC server")
	err = s.grpcServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
}
