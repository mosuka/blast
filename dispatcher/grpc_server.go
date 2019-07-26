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
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mosuka/blast/protobuf/distribute"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	//grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	//grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	//grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	//grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
)

type GRPCServer struct {
	service  distribute.DistributeServer
	server   *grpc.Server
	listener net.Listener

	logger *zap.Logger
}

func NewGRPCServer(grpcAddr string, service distribute.DistributeServer, logger *zap.Logger) (*GRPCServer, error) {
	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			//grpc_ctxtags.StreamServerInterceptor(),
			//grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logger),
			//grpc_auth.StreamServerInterceptor(myAuthFunction),
			//grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			//grpc_ctxtags.UnaryServerInterceptor(),
			//grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logger),
			//grpc_auth.UnaryServerInterceptor(myAuthFunction),
			//grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	distribute.RegisterDistributeServer(server, service)

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(server)

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, err
	}

	return &GRPCServer{
		service:  service,
		server:   server,
		listener: listener,
		logger:   logger,
	}, nil
}

func (s *GRPCServer) Start() error {
	s.logger.Info("start server")
	err := s.server.Serve(s.listener)
	if err != nil {
		return err
	}

	return nil
}

func (s *GRPCServer) Stop() error {
	s.logger.Info("stop server")
	s.server.Stop()
	//s.server.GracefulStop()

	return nil
}