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
	"log"

	accesslog "github.com/mash/go-accesslog"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/http"
	"github.com/mosuka/blast/protobuf"
)

type Server struct {
	managerAddr string
	clusterId   string

	id       string
	metadata map[string]interface{}

	peerAddr string

	indexConfig map[string]interface{}

	raftServer  *RaftServer
	grpcService *GRPCService
	grpcServer  *grpc.Server
	httpRouter  *http.Router
	httpServer  *http.Server

	logger     *log.Logger
	httpLogger accesslog.Logger
}

func NewServer(managerAddr string, clusterId string, id string, metadata map[string]interface{}, peerAddr string, indexConfig map[string]interface{}, logger *log.Logger, httpLogger accesslog.Logger) (*Server, error) {
	return &Server{
		managerAddr: managerAddr,
		clusterId:   clusterId,

		id:       id,
		metadata: metadata,

		peerAddr: peerAddr,

		indexConfig: indexConfig,

		logger:     logger,
		httpLogger: httpLogger,
	}, nil
}

func (s *Server) Start() {
	// get peer from manager
	if s.managerAddr != "" {
		s.logger.Printf("[INFO] connect to master %s", s.managerAddr)

		mc, err := grpc.NewClient(s.managerAddr)
		defer func() {
			s.logger.Printf("[DEBUG] close client for %v", mc.GetAddress())
			err = mc.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		s.logger.Printf("[INFO] get nodes in cluster: %s", s.clusterId)
		clusterIntr, err := mc.GetState(fmt.Sprintf("cluster_config/clusters/%s/nodes", s.clusterId))
		if err == errors.ErrNotFound {
			// cluster does not found
			s.logger.Printf("[INFO] cluster does not found: %s", s.clusterId)
		} else if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		} else {
			if clusterIntr == nil {
				s.logger.Print("[INFO] value is nil")
			} else {
				cluster := *clusterIntr.(*map[string]interface{})
				for nodeId, nodeIntr := range cluster {
					// skip if it is own node id
					if nodeId == s.id {
						continue
					}

					// get the peer node address
					metadata := nodeIntr.(map[string]interface{})["metadata"].(map[string]interface{})
					s.peerAddr = metadata["grpc_addr"].(string)

					s.logger.Printf("[INFO] peer node detected: %s", s.peerAddr)

					break
				}
			}
		}
	}

	// bootstrap node?
	bootstrap := s.peerAddr == ""
	s.logger.Printf("[INFO] bootstrap: %v", bootstrap)

	// get index config from manager or peer
	if s.managerAddr != "" {
		mc, err := grpc.NewClient(s.managerAddr)
		defer func() {
			s.logger.Printf("[DEBUG] close client for %v", mc.GetAddress())
			err = mc.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		value, err := mc.GetState("index_config")
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		if value != nil {
			s.indexConfig = *value.(*map[string]interface{})
		}
	} else if s.peerAddr != "" {
		pc, err := grpc.NewClient(s.peerAddr)
		defer func() {
			err = pc.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		resp, err := pc.GetIndexConfig()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		ins, err := protobuf.MarshalAny(resp.IndexConfig)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		s.indexConfig = *ins.(*map[string]interface{})
	}

	var err error

	// create raft server
	s.raftServer, err = NewRaftServer(s.id, s.metadata, bootstrap, s.indexConfig, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create gRPC service
	s.grpcService, err = NewGRPCService(s.managerAddr, s.clusterId, s.raftServer, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create gRPC server
	s.grpcServer, err = grpc.NewServer(s.metadata["grpc_addr"].(string), s.grpcService, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create HTTP router
	s.httpRouter, err = NewRouter(s.metadata["grpc_addr"].(string), s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create HTTP server
	s.httpServer, err = http.NewServer(s.metadata["http_addr"].(string), s.httpRouter, s.logger, s.httpLogger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// start Raft server
	s.logger.Print("[INFO] start Raft server")
	err = s.raftServer.Start()
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

	// join to the existing cluster
	if !bootstrap {
		client, err := grpc.NewClient(s.peerAddr)
		defer func() {
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		err = client.SetNode(s.id, s.metadata)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}
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

	// stop Raft server
	s.logger.Printf("[INFO] stop Raft server")
	err = s.raftServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
}
