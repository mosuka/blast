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
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/mosuka/blast/protobuf/raft"
)

type Server struct {
	managerAddr string
	clusterId   string

	node     *raft.Node
	peerAddr string

	indexConfig map[string]interface{}

	raftServer *RaftServer
	grpcServer *GRPCServer
	httpServer *HTTPServer

	logger     *log.Logger
	httpLogger accesslog.Logger
}

func NewServer(managerAddr string, clusterId string, node *raft.Node, peerAddr string, indexConfig map[string]interface{}, logger *log.Logger, httpLogger accesslog.Logger) (*Server, error) {
	return &Server{
		managerAddr: managerAddr,
		clusterId:   clusterId,

		node:     node,
		peerAddr: peerAddr,

		indexConfig: indexConfig,

		logger:     logger,
		httpLogger: httpLogger,
	}, nil
}

func (s *Server) Start() {
	// get peer from manager
	if s.managerAddr != "" {
		mc, err := manager.NewGRPCClient(s.managerAddr)
		defer func() {
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

		kvp, err := mc.Get(
			&management.KeyValuePair{
				Key: fmt.Sprintf("/cluster_config/clusters/%s/nodes", s.clusterId),
			},
		)
		if err == errors.ErrNotFound {
			// not found
			s.logger.Printf("[WARN] %v", err)
		} else if err != nil {
			// error
			s.logger.Printf("[ERR] %v", err)
			return
		} else {
			ins, err := protobuf.MarshalAny(kvp.Value)
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return
			}

			if ins != nil {
				m := *ins.(*map[string]interface{})
				for k, v := range m {
					// skip if it is own node id
					if k == s.node.Id {
						continue
					}

					// get the peer node address
					metadata := v.(map[string]interface{})
					s.peerAddr = metadata["grpc_addr"].(string)
					break
				}
			}
		}
	}

	// bootstrap node?
	bootstrap := s.peerAddr == ""

	// get index config from manager or peer
	if s.managerAddr != "" {
		mc, err := manager.NewGRPCClient(s.managerAddr)
		defer func() {
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

		kvp, err := mc.Get(&management.KeyValuePair{Key: "/index_config"})
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		ins, err := protobuf.MarshalAny(kvp.Value)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		if ins != nil {
			s.indexConfig = *ins.(*map[string]interface{})
		}
	} else if s.peerAddr != "" {
		pc, err := NewGRPCClient(s.peerAddr)
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

		indexConfig, err := pc.GetIndexConfig()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		ins, err := protobuf.MarshalAny(indexConfig.IndexMapping)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}

		s.indexConfig = map[string]interface{}{
			"index_mapping":      *ins.(*map[string]interface{}),
			"index_type":         indexConfig.IndexType,
			"index_storage_type": indexConfig.IndexStorageType,
		}
	}

	var err error

	// create raft server
	s.raftServer, err = NewRaftServer(s.managerAddr, s.clusterId, s.node, bootstrap, s.indexConfig, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create gRPC server
	s.grpcServer, err = NewGRPCServer(s.node.Metadata.GrpcAddr, s.raftServer, s.logger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// create HTTP server
	s.httpServer, err = NewHTTPServer(s.node.Metadata.HttpAddr, s.node.Metadata.GrpcAddr, s.logger, s.httpLogger)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}

	// start Raft server
	go func() {
		// start raft server
		err := s.raftServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
		s.logger.Print("[INFO] Raft server started")
	}()

	// start gRPC server
	go func() {
		err := s.grpcServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
		s.logger.Print("[INFO] gRPC server started")
	}()

	// start HTTP server
	go func() {
		err := s.httpServer.Start()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
		s.logger.Print("[INFO] HTTP server started")
	}()

	// join to the existing cluster
	if !bootstrap {
		client, err := NewGRPCClient(s.peerAddr)
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

		err = client.Join(s.node)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			return
		}
	}
}

func (s *Server) Stop() {
	// stop HTTP server
	err := s.httpServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop gRPC server
	err = s.grpcServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// stop Raft server
	err = s.raftServer.Stop()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
}
