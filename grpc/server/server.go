//  Copyright (c) 2018 Minoru Osuka
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

package server

import (
	"log"
	"net"

	"github.com/mosuka/blast/logging"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/service"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	grpcAddress string

	logger *log.Logger

	server   *grpc.Server
	service  *service.KVSService
	listener net.Listener
}

func NewGRPCServer(gRPCAddress string, service *service.KVSService) (*GRPCServer, error) {
	return &GRPCServer{
		grpcAddress: gRPCAddress,
		service:     service,
		logger:      logging.DefaultLogger(),
	}, nil
}

func (s *GRPCServer) SetLogger(logger *log.Logger) {
	s.logger = logger
	return
}

func (s *GRPCServer) Start() error {
	var err error

	// Create gRPC server
	s.server = grpc.NewServer()
	s.logger.Print("[INFO] server: New gRPC server has been created")

	// Register service to gRPC server
	protobuf.RegisterKVSServer(s.server, s.service)
	s.logger.Print("[INFO] server: Service has been registered to gRPC server")

	// Create listener
	if s.listener, err = net.Listen("tcp", s.grpcAddress); err != nil {
		s.logger.Printf("[ERR] server: Failed to create listener for gRPC at %s: %v", s.grpcAddress, err)
		return err
	}
	s.logger.Printf("[INFO] server: Listener for gRPC has been created at %s", s.grpcAddress)

	go func() {
		// Start server
		s.logger.Print("[INFO] server: Start gRPC server")
		if err := s.server.Serve(s.listener); err != nil {
			s.logger.Printf("[ERR] server: Failed to start gRPC server: %v", err)
		}
		return
	}()

	return nil
}

func (s *GRPCServer) Stop() error {
	// Stop server
	s.server.GracefulStop()
	s.logger.Print("[INFO] server: gRPC server has been stopped")

	return nil
}
