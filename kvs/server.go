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
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	accesslog "github.com/mash/go-accesslog"
	blasthttp "github.com/mosuka/blast/http"
	"github.com/mosuka/blast/protobuf/kvs"
	"github.com/mosuka/blast/protobuf/raft"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type Server struct {
	store *RaftServer

	// gRPC
	grpcListener net.Listener
	grpcService  *Service
	grpcServer   *grpc.Server

	// HTTP
	httpListener net.Listener
	router       *mux.Router

	logger     *log.Logger
	httpLogger *log.Logger
}

func NewServer(nodeId string, bindAddr string, grpcAddr string, httpAddr string, dataDir string, joinAddr string, logger *log.Logger, httpLogger *log.Logger) *Server {
	var err error

	server := &Server{
		logger:     logger,
		httpLogger: httpLogger,
	}

	// store
	server.store, err = NewRaftServer(bindAddr, dataDir, server.logger)
	if err != nil {
		server.logger.Printf("[ERR] %v", err)
		return nil
	}

	bootstrap := joinAddr == ""
	err = server.store.Open(bootstrap, nodeId)
	if err != nil {
		server.logger.Printf("[ERR] %v", err)
		return nil
	}

	if !bootstrap {
		// send join request to node already exists
		client, err := NewClient(joinAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				server.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			server.logger.Printf("[ERR] %v", err)
			return nil
		}
		server.logger.Printf("[INFO] join request send to %s", joinAddr)

		node := &raft.Node{
			Id:       nodeId,
			BindAddr: bindAddr,
		}

		joinReq := &raft.JoinRequest{
			Node: node,
		}
		_, err = client.Join(joinReq)
		if err != nil {
			server.logger.Printf("[ERR] %v", err)
			return nil
		}
	}

	// gRPC service
	server.grpcService, err = NewService(server.store, server.logger)
	if err != nil {
		server.logger.Printf("[ERR] %v", err)
		return nil
	}

	// gRPC server
	server.grpcServer = grpc.NewServer()
	kvs.RegisterKVSServer(server.grpcServer, server.grpcService)

	// gRPC listener
	server.grpcListener, err = net.Listen("tcp", grpcAddr)
	if err != nil {
		server.logger.Printf("[ERR] %v", err)
		return nil
	}
	server.logger.Printf("[INFO] gRPC server listen in %s", grpcAddr)

	// HTTP listener
	server.httpListener, err = net.Listen("tcp", httpAddr)
	if err != nil {
		server.logger.Printf("[ERR] %v", err)
		return nil
	}
	server.logger.Printf("[INFO] HTTP server listen in %s", httpAddr)

	// HTTP router
	server.router = mux.NewRouter()
	server.router.StrictSlash(true)

	// HTTP handlers
	server.router.Handle("/", NewRootHandler(server.logger)).Methods("GET")
	server.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	return server
}

func (s *Server) Start() {
	// gRPC
	go func() {
		err := s.grpcServer.Serve(s.grpcListener)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
	}()
	s.logger.Print("[INFO] gRPC server started")

	// HTTP
	go func() {
		err := http.Serve(
			s.httpListener,
			accesslog.NewLoggingHandler(
				s.router,
				blasthttp.ApacheCombinedLogger{
					Logger: s.httpLogger,
				},
			),
		)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
	}()
	s.logger.Print("[INFO] HTTP server started")
}

func (s *Server) Stop() {
	// HTTP
	err := s.httpListener.Close()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
	s.logger.Print("[INFO] HTTP server stopped")

	// gRPC
	s.grpcServer.GracefulStop()
	s.logger.Print("[INFO] gRPC server stopped")

	err = s.store.Close()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
}
