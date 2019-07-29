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
	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/manager"
	"go.uber.org/zap"
)

type Server struct {
	clusterConfig *config.ClusterConfig
	nodeConfig    *config.NodeConfig
	indexConfig   *config.IndexConfig
	logger        *zap.Logger
	grpcLogger    *zap.Logger
	httpLogger    accesslog.Logger

	raftServer  *RaftServer
	grpcService *GRPCService
	grpcServer  *GRPCServer
	httpRouter  *Router
	httpServer  *HTTPServer
}

func NewServer(clusterConfig *config.ClusterConfig, nodeConfig *config.NodeConfig, indexConfig *config.IndexConfig, logger *zap.Logger, grpcLogger *zap.Logger, httpLogger accesslog.Logger) (*Server, error) {
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
	// get peer from manager
	if s.clusterConfig.ManagerAddr != "" {
		s.logger.Info("connect to manager", zap.String("manager_addr", s.clusterConfig.ManagerAddr))

		mc, err := manager.NewGRPCClient(s.clusterConfig.ManagerAddr)
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

		clusterIntr, err := mc.Get(fmt.Sprintf("cluster_config/clusters/%s/nodes", s.clusterConfig.ClusterId))
		if err != nil && err != errors.ErrNotFound {
			s.logger.Fatal(err.Error())
			return
		}
		if clusterIntr != nil {
			cluster := *clusterIntr.(*map[string]interface{})
			for nodeId, nodeInfoIntr := range cluster {
				if nodeId == s.nodeConfig.NodeId {
					s.logger.Debug("skip own node id", zap.String("node_id", nodeId))
					continue
				}

				nodeInfo := nodeInfoIntr.(map[string]interface{})

				// get the peer node config
				nodeConfig, ok := nodeInfo["node_config"].(map[string]interface{})
				if !ok {
					s.logger.Error("missing node config", zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))
					continue
				}

				// get the peer node gRPC address
				grpcAddr, ok := nodeConfig["grpc_addr"].(string)
				if !ok {
					s.logger.Error("missing gRPC address", zap.String("id", nodeId), zap.String("grpc_addr", grpcAddr))
					continue
				}

				s.logger.Info("peer node detected", zap.String("peer_addr", grpcAddr))
				s.clusterConfig.PeerAddr = grpcAddr
				break
			}
		}
	}

	//get index config from manager or peer
	if s.clusterConfig.ManagerAddr != "" {
		mc, err := manager.NewGRPCClient(s.clusterConfig.ManagerAddr)
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

		s.logger.Debug("pull index config from manager", zap.String("address", mc.GetAddress()))
		value, err := mc.Get("/index_config")
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		if value != nil {
			s.indexConfig = config.NewIndexConfigFromMap(*value.(*map[string]interface{}))
		}
	} else if s.clusterConfig.PeerAddr != "" {
		pc, err := NewGRPCClient(s.clusterConfig.PeerAddr)
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
		value, err := pc.GetIndexConfig()
		if err != nil {
			s.logger.Fatal(err.Error())
			return
		}

		if value != nil {
			s.indexConfig = config.NewIndexConfigFromMap(value)
		}
	}

	// bootstrap node?
	bootstrap := s.clusterConfig.PeerAddr == ""
	s.logger.Info("bootstrap", zap.Bool("bootstrap", bootstrap))

	var err error

	// create raft server
	s.raftServer, err = NewRaftServer(s.nodeConfig, s.indexConfig, bootstrap, s.logger)
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}

	// create gRPC service
	s.grpcService, err = NewGRPCService(s.clusterConfig, s.raftServer, s.logger)
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
		client, err := NewGRPCClient(s.clusterConfig.PeerAddr)
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

		err = client.SetNode(s.nodeConfig.NodeId, s.nodeConfig.ToMap())
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
