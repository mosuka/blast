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

package grpc

import (
	"log"
	"net"

	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc"
)

type Server struct {
	service  protobuf.BlastServer
	server   *grpc.Server
	listener net.Listener

	logger *log.Logger
}

func NewServer(grpcAddr string, service protobuf.BlastServer, logger *log.Logger) (*Server, error) {
	server := grpc.NewServer()

	protobuf.RegisterBlastServer(server, service)

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		service:  service,
		server:   server,
		listener: listener,
		logger:   logger,
	}, nil
}

func (s *Server) Start() error {
	s.logger.Print("[INFO] start server")
	err := s.server.Serve(s.listener)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	s.logger.Print("[INFO] stop server")
	s.server.Stop() // TODO: graceful stop

	return nil
}
