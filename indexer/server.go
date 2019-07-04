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
	"fmt"

	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/http"
	"go.uber.org/zap"
)

type Server struct {
	managerAddr string
	clusterId   string

	id              string
	metadata        map[string]interface{}
	raftStorageType string
	peerAddr        string

	indexConfig map[string]interface{}

	raftServer  *RaftServer
	grpcService *GRPCService
	grpcServer  *grpc.Server
	httpRouter  *http.Router
	httpServer  *http.Server

	logger     *zap.Logger
	httpLogger accesslog.Logger
}

func NewServer(managerAddr string, clusterId string, id string, metadata map[string]interface{}, raftStorageType, peerAddr string, indexConfig map[string]interface{}, logger *zap.Logger, httpLogger accesslog.Logger) (*Server, error) {
	return &Server{
		managerAddr: managerAddr,
		clusterId:   clusterId,

		id:              id,
		metadata:        metadata,
		raftStorageType: raftStorageType,
		peerAddr:        peerAddr,

		indexConfig: indexConfig,

		logger:     logger,
		httpLogger: httpLogger,
	}, nil
}

func (s *Server) Start() {
	// get peer from manager
	if s.managerAddr != "" {
		s.logger.Info("connect to master", zap.String("master_addr", s.managerAddr))

		mc, err := grpc.NewClient(s.managerAddr)
		defer func() {
			s.logger.Debug("close client", zap.String("address", mc.GetAddress()))
			err = mc.Close()
			if err != nil {
				s.logger.Error(err.Error())
				return
			}
		}()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		s.logger.Info("get nodes in cluster from master", zap.String("master_addr", mc.GetAddress()), zap.String("cluster", s.clusterId))
		clusterIntr, err := mc.GetState(fmt.Sprintf("cluster_config/clusters/%s/nodes", s.clusterId))
		if err != nil && err != errors.ErrNotFound {
			s.logger.Fatal(err.Error())
			return
		}
		if clusterIntr != nil {
			cluster := *clusterIntr.(*map[string]interface{})
			for nodeId, nodeIntr := range cluster {
				if nodeId == s.id {
					s.logger.Debug("skip own node id", zap.String("id", nodeId))
					continue
				}

				// get the peer node address
				metadata, ok := nodeIntr.(map[string]interface{})["metadata"].(map[string]interface{})
				if !ok {
					s.logger.Error("missing metadata", zap.String("id", nodeId), zap.Any("metadata", metadata))
					continue
				}

				grpcAddr, ok := metadata["grpc_addr"].(string)
				if !ok {
					s.logger.Error("missing gRPC address", zap.String("id", nodeId), zap.String("grpc_addr", grpcAddr))
					continue
				}

				s.peerAddr = grpcAddr

				s.logger.Info("peer node detected", zap.String("peer_addr", s.peerAddr))

				break
			}
		}
	}

	// bootstrap node?
	bootstrap := s.peerAddr == ""
	s.logger.Info("bootstrap", zap.Bool("bootstrap", bootstrap))

	// get index config from manager or peer
	if s.managerAddr != "" {
		mc, err := grpc.NewClient(s.managerAddr)
		defer func() {
			s.logger.Debug("close client", zap.String("address", mc.GetAddress()))
			err = mc.Close()
			if err != nil {
				s.logger.Error(err.Error())
				return
			}
		}()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		s.logger.Debug("pull index config from master", zap.String("address", mc.GetAddress()))
		value, err := mc.GetState("index_config")
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		if value != nil {
			s.indexConfig = *value.(*map[string]interface{})
		}
	} else if s.peerAddr != "" {
		pc, err := grpc.NewClient(s.peerAddr)
		defer func() {
			s.logger.Debug("close client", zap.String("address", pc.GetAddress()))
			err = pc.Close()
			if err != nil {
				s.logger.Fatal(err.Error())
				return
			}
		}()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		s.logger.Debug("pull index config from cluster peer", zap.String("address", pc.GetAddress()))
		s.indexConfig, err = pc.GetIndexConfig()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}
	}

	var err error

	// create raft server
	s.raftServer, err = NewRaftServer(s.id, s.metadata, s.raftStorageType, bootstrap, s.indexConfig, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create gRPC service
	s.grpcService, err = NewGRPCService(s.managerAddr, s.clusterId, s.raftServer, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	grpcAddr, ok := s.metadata["grpc_addr"].(string)
	if !ok {
		s.logger.Fatal("missing gRPC address", zap.String("grpc_addr", grpcAddr))
		return
	}

	// create gRPC server
	s.grpcServer, err = grpc.NewServer(grpcAddr, s.grpcService, s.logger)
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

	httpAddr, ok := s.metadata["http_addr"].(string)
	if !ok {
		s.logger.Fatal("missing HTTP address", zap.String("http_addr", httpAddr))
		return
	}

	// create HTTP server
	s.httpServer, err = http.NewServer(httpAddr, s.httpRouter, s.logger, s.httpLogger)
	if err != nil {
		s.logger.Fatal(err.Error())
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
		client, err := grpc.NewClient(s.peerAddr)
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

		err = client.SetNode(s.id, s.metadata)
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
