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
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/config"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	*grpc.Service

	clusterConfig *config.ClusterConfig
	raftServer    *RaftServer
	logger        *zap.Logger

	updateClusterStopCh chan struct{}
	updateClusterDoneCh chan struct{}
	peers               map[string]interface{}
	peerClients         map[string]*grpc.Client
	cluster             map[string]interface{}
	clusterChans        map[chan protobuf.GetClusterResponse]struct{}
	clusterMutex        sync.RWMutex

	managers             map[string]interface{}
	managerClients       map[string]*grpc.Client
	updateManagersStopCh chan struct{}
	updateManagersDoneCh chan struct{}
}

func NewGRPCService(clusterConfig *config.ClusterConfig, raftServer *RaftServer, logger *zap.Logger) (*GRPCService, error) {
	return &GRPCService{
		clusterConfig: clusterConfig,
		raftServer:    raftServer,
		logger:        logger,

		peers:        make(map[string]interface{}, 0),
		peerClients:  make(map[string]*grpc.Client, 0),
		cluster:      make(map[string]interface{}, 0),
		clusterChans: make(map[chan protobuf.GetClusterResponse]struct{}),

		managers:       make(map[string]interface{}, 0),
		managerClients: make(map[string]*grpc.Client, 0),
	}, nil
}

func (s *GRPCService) Start() error {
	s.logger.Info("start to update cluster info")
	go s.startUpdateCluster(500 * time.Millisecond)

	//managerAddr, err := s.clusterConfig.ManagerAddr
	//if err != nil && err != maputils.ErrNotFound {
	//	s.logger.Fatal(err.Error(), zap.String("manager_addr", managerAddr))
	//}
	if s.clusterConfig.ManagerAddr != "" {
		s.logger.Info("start to update manager cluster info")
		go s.startUpdateManagers(500 * time.Millisecond)
	}

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Info("stop to update cluster info")
	s.stopUpdateCluster()

	//managerAddr, err := s.clusterConfig.GetManagerAddr()
	//if err != nil && err != maputils.ErrNotFound {
	//	s.logger.Fatal(err.Error(), zap.String("manager_addr", managerAddr))
	//}
	if s.clusterConfig.ManagerAddr != "" {
		s.logger.Info("stop to update manager cluster info")
		s.stopUpdateManagers()
	}

	return nil
}

func (s *GRPCService) getManagerClient() (*grpc.Client, error) {
	var client *grpc.Client

	for id, node := range s.managers {
		nm, ok := node.(map[string]interface{})
		if !ok {
			s.logger.Warn("assertion failed", zap.String("id", id))
			continue
		}

		state, ok := nm["state"].(string)
		if !ok {
			s.logger.Warn("missing state", zap.String("id", id), zap.String("state", state))
			continue
		}

		if state == raft.Leader.String() || state == raft.Follower.String() {
			client, ok = s.managerClients[id]
			if ok {
				return client, nil
			} else {
				s.logger.Error("node does not exist", zap.String("id", id))
			}
		} else {
			s.logger.Debug("node has not available", zap.String("id", id), zap.String("state", state))
		}
	}

	err := errors.New("available client does not exist")
	s.logger.Error(err.Error())

	return nil, err
}

func (s *GRPCService) getInitialManagers(managerAddr string) (map[string]interface{}, error) {
	client, err := grpc.NewClient(managerAddr)
	defer func() {
		err := client.Close()
		if err != nil {
			s.logger.Error(err.Error())
		}
		return
	}()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	managers, err := client.GetCluster()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return managers, nil
}

func (s *GRPCService) startUpdateManagers(checkInterval time.Duration) {
	s.updateManagersStopCh = make(chan struct{})
	s.updateManagersDoneCh = make(chan struct{})

	defer func() {
		close(s.updateManagersDoneCh)
	}()

	//managerAddr, err := s.clusterConfig.GetManagerAddr()
	//if err != nil && err != maputils.ErrNotFound {
	//	s.logger.Fatal(err.Error(), zap.String("manager_addr", managerAddr))
	//}

	var err error

	// get initial managers
	s.managers, err = s.getInitialManagers(s.clusterConfig.ManagerAddr)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.logger.Debug("initialize manager list", zap.Any("managers", s.managers))

	// create clients for managers
	for nodeId, node := range s.managers {
		nm, ok := node.(map[string]interface{})
		if !ok {
			s.logger.Warn("assertion failed", zap.String("id", nodeId))
			continue
		}

		metadata, ok := nm["metadata"].(map[string]interface{})
		if !ok {
			s.logger.Warn("missing metadata", zap.String("id", nodeId), zap.Any("metadata", metadata))
			continue
		}

		grpcAddr, ok := metadata["grpc_addr"].(string)
		if !ok {
			s.logger.Warn("missing gRPC address", zap.String("id", nodeId), zap.String("grpc_addr", grpcAddr))
			continue
		}

		s.logger.Debug("create gRPC client", zap.String("id", nodeId), zap.String("grpc_addr", grpcAddr))
		client, err := grpc.NewClient(grpcAddr)
		if err != nil {
			s.logger.Error(err.Error(), zap.String("id", nodeId), zap.String("grpc_addr", grpcAddr))
		}
		if client != nil {
			s.managerClients[nodeId] = client
		}
	}

	for {
		select {
		case <-s.updateManagersStopCh:
			s.logger.Info("received a request to stop updating a manager cluster")
			return
		default:
			client, err := s.getManagerClient()
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}

			stream, err := client.WatchCluster()
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}

			s.logger.Info("wait for receive a manager cluster updates from stream")
			resp, err := stream.Recv()
			if err == io.EOF {
				s.logger.Info(err.Error())
				continue
			}
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}

			// get current manager cluster
			managersIntr, err := protobuf.MarshalAny(resp.Cluster)
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}
			if managersIntr == nil {
				s.logger.Error(err.Error())
				continue
			}
			managers := *managersIntr.(*map[string]interface{})

			if !reflect.DeepEqual(s.managers, managers) {
				// open clients
				for id, metadata := range managers {
					mm, ok := metadata.(map[string]interface{})
					if !ok {
						s.logger.Warn("assertion failed", zap.String("id", id))
						continue
					}

					grpcAddr, ok := mm["grpc_addr"].(string)
					if !ok {
						s.logger.Warn("missing metadata", zap.String("id", id), zap.String("grpc_addr", grpcAddr))
						continue
					}

					client, exist := s.managerClients[id]
					if exist {
						s.logger.Debug("client has already exist in manager list", zap.String("id", id))

						if client.GetAddress() != grpcAddr {
							s.logger.Debug("gRPC address has been changed", zap.String("id", id), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", grpcAddr))
							s.logger.Debug("recreate gRPC client", zap.String("id", id), zap.String("grpc_addr", grpcAddr))

							delete(s.managerClients, id)

							err = client.Close()
							if err != nil {
								s.logger.Error(err.Error(), zap.String("id", id))
							}

							newClient, err := grpc.NewClient(grpcAddr)
							if err != nil {
								s.logger.Error(err.Error(), zap.String("id", id), zap.String("grpc_addr", grpcAddr))
							}

							if newClient != nil {
								s.managerClients[id] = newClient
							}
						} else {
							s.logger.Debug("gRPC address has not changed", zap.String("id", id), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", grpcAddr))
						}
					} else {
						s.logger.Debug("client does not exist in peer list", zap.String("id", id))

						s.logger.Debug("create gRPC client", zap.String("id", id), zap.String("grpc_addr", grpcAddr))
						newClient, err := grpc.NewClient(grpcAddr)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("id", id), zap.String("grpc_addr", grpcAddr))
						}
						if newClient != nil {
							s.managerClients[id] = newClient
						}
					}
				}

				// close nonexistent clients
				for id, client := range s.managerClients {
					if metadata, exist := managers[id]; !exist {
						s.logger.Info("this client is no longer in use", zap.String("id", id), zap.Any("metadata", metadata))

						s.logger.Debug("close client", zap.String("id", id), zap.String("address", client.GetAddress()))
						err = client.Close()
						if err != nil {
							s.logger.Error(err.Error(), zap.String("id", id), zap.String("address", client.GetAddress()))
						}

						s.logger.Debug("delete client", zap.String("id", id))
						delete(s.managerClients, id)
					}
				}

				// keep current manager cluster
				s.managers = managers
				s.logger.Debug("managers", zap.Any("managers", s.managers))
			}
		}
	}
}

func (s *GRPCService) stopUpdateManagers() {
	s.logger.Info("close all manager clients")
	for id, client := range s.managerClients {
		s.logger.Debug("close manager client", zap.String("id", id), zap.String("address", client.GetAddress()))
		err := client.Close()
		if err != nil {
			s.logger.Error(err.Error())
		}
	}

	if s.updateManagersStopCh != nil {
		s.logger.Info("send a request to stop updating a manager cluster")
		close(s.updateManagersStopCh)
	}

	s.logger.Info("wait for the manager cluster update to stop")
	<-s.updateManagersDoneCh
	s.logger.Info("the manager cluster update has been stopped")
}

func (s *GRPCService) getLeaderClient() (*grpc.Client, error) {
	var client *grpc.Client

	for id, node := range s.cluster {
		nm, ok := node.(map[string]interface{})
		if !ok {
			s.logger.Warn("assertion failed", zap.String("id", id))
			continue
		}

		state, ok := nm["state"].(string)
		if !ok {
			s.logger.Warn("missing state", zap.String("id", id), zap.String("state", state))
			continue
		}

		if state == raft.Leader.String() {
			client, ok = s.peerClients[id]
			if ok {
				return client, nil
			} else {
				s.logger.Error("node does not exist", zap.String("id", id))
			}
		} else {
			s.logger.Debug("not a leader", zap.String("id", id))
		}
	}

	err := errors.New("available client does not exist")
	s.logger.Error(err.Error())

	return nil, err
}

func (s *GRPCService) startUpdateCluster(checkInterval time.Duration) {
	s.updateClusterStopCh = make(chan struct{})
	s.updateClusterDoneCh = make(chan struct{})

	defer func() {
		close(s.updateClusterDoneCh)
	}()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.updateClusterStopCh:
			s.logger.Info("received a request to stop updating a cluster")
			return
		case <-ticker.C:
			// get servers
			servers, err := s.raftServer.GetServers()
			if err != nil {
				s.logger.Error(err.Error())
				return
			}

			// create peer node list with out self node
			peers := make(map[string]interface{}, 0)
			for nodeId, nodeConfig := range servers {
				if nodeId != s.raftServer.nodeConfig.NodeId {
					peers[nodeId] = nodeConfig
				}
			}

			if !reflect.DeepEqual(s.peers, peers) {
				// create clients
				for nodeId, nodeConfig := range peers {
					mm, ok := nodeConfig.(map[string]interface{})
					if !ok {
						s.logger.Warn("assertion failed", zap.String("node_id", nodeId))
						continue
					}

					grpcAddr, ok := mm["grpc_addr"].(string)
					if !ok {
						s.logger.Warn("missing metadata", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
						continue
					}

					client, exist := s.peerClients[nodeId]
					if exist {
						s.logger.Debug("client has already exist in peer list", zap.String("node_id", nodeId))

						if client.GetAddress() != grpcAddr {
							s.logger.Debug("gRPC address has been changed", zap.String("node_id", nodeId), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", grpcAddr))
							s.logger.Debug("recreate gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))

							delete(s.peerClients, nodeId)

							err = client.Close()
							if err != nil {
								s.logger.Error(err.Error(), zap.String("node_id", nodeId))
							}

							newClient, err := grpc.NewClient(grpcAddr)
							if err != nil {
								s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
							}

							if newClient != nil {
								s.peerClients[nodeId] = newClient
							}
						} else {
							s.logger.Debug("gRPC address has not changed", zap.String("node_id", nodeId), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", grpcAddr))
						}
					} else {
						s.logger.Debug("client does not exist in peer list", zap.String("node_id", nodeId))

						s.logger.Debug("create gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
						peerClient, err := grpc.NewClient(grpcAddr)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
						}
						if peerClient != nil {
							s.logger.Debug("append peer client to peer client list", zap.String("grpc_addr", peerClient.GetAddress()))
							s.peerClients[nodeId] = peerClient
						}
					}
				}

				// close nonexistent clients
				for nodeId, client := range s.peerClients {
					if nodeConfig, exist := peers[nodeId]; !exist {
						s.logger.Info("this client is no longer in use", zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))

						s.logger.Debug("close client", zap.String("node_id", nodeId), zap.String("grpc_addr", client.GetAddress()))
						err = client.Close()
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", client.GetAddress()))
						}

						s.logger.Debug("delete client", zap.String("node_id", nodeId))
						delete(s.peerClients, nodeId)
					}
				}

				// keep current peer nodes
				s.peers = peers
				s.logger.Debug("peers", zap.Any("peers", s.peers))
			}

			// get cluster
			cluster := make(map[string]interface{}, 0)
			ctx, _ := grpc.NewContext()
			resp, err := s.GetCluster(ctx, &empty.Empty{})
			if err != nil {
				s.logger.Error(err.Error())
			}
			clusterIntr, err := protobuf.MarshalAny(resp.Cluster)
			if err != nil {
				s.logger.Error(err.Error())
			}
			if clusterIntr == nil {
				s.logger.Error("unexpected value")
			}
			cluster = *clusterIntr.(*map[string]interface{})

			// notify current cluster
			if !reflect.DeepEqual(s.cluster, cluster) {
				for c := range s.clusterChans {
					s.logger.Debug("notify cluster changes to client", zap.Any("response", resp))
					c <- *resp
				}

				if s.clusterConfig.ManagerAddr != "" && s.raftServer.IsLeader() {
					client, err := s.getManagerClient()
					if err != nil {
						s.logger.Error(err.Error())
					}
					err = client.SetState(fmt.Sprintf("cluster_config/clusters/%s/nodes", s.clusterConfig.ClusterId), cluster)
					if err != nil {
						s.logger.Error(err.Error())
					}
				}

				// keep current cluster
				s.cluster = cluster
				s.logger.Debug("cluster", zap.Any("cluster", cluster))
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *GRPCService) stopUpdateCluster() {
	s.logger.Info("close all peer clients")
	for id, client := range s.peerClients {
		s.logger.Debug("close peer client", zap.String("id", id), zap.String("address", client.GetAddress()))
		err := client.Close()
		if err != nil {
			s.logger.Error(err.Error())
		}
	}

	if s.updateClusterStopCh != nil {
		s.logger.Info("send a request to stop updating a cluster")
		close(s.updateClusterStopCh)
	}

	s.logger.Info("wait for the cluster update to stop")
	<-s.updateClusterDoneCh
	s.logger.Info("the cluster update has been stopped")
}

func (s *GRPCService) getSelfNode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"node_config": s.raftServer.nodeConfig.ToMap(),
		"state":       s.raftServer.State(),
	}, nil
}

func (s *GRPCService) getPeerNode(id string) (map[string]interface{}, error) {
	var nodeInfo map[string]interface{}
	var err error

	if peerClient, exist := s.peerClients[id]; exist {
		nodeInfo, err = peerClient.GetNode(id)
		if err != nil {
			s.logger.Error(err.Error())
			nodeInfo = map[string]interface{}{
				"node_config": map[string]interface{}{},
				"state":       raft.Shutdown.String(),
			}
		}
	} else {
		s.logger.Error("node does not exist in peer list", zap.String("id", id))
		nodeInfo = map[string]interface{}{
			"node_config": map[string]interface{}{},
			"state":       raft.Shutdown.String(),
		}
	}

	return nodeInfo, nil
}

func (s *GRPCService) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	resp := &protobuf.GetNodeResponse{}

	var nodeInfo map[string]interface{}
	var err error

	if req.Id == "" || req.Id == s.raftServer.nodeConfig.NodeId {
		nodeInfo, err = s.getSelfNode()
	} else {
		nodeInfo, err = s.getPeerNode(req.Id)
	}
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	nodeConfigAny := &any.Any{}
	if nodeConfig, exist := nodeInfo["node_config"]; exist {
		err = protobuf.UnmarshalAny(nodeConfig.(map[string]interface{}), nodeConfigAny)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	} else {
		s.logger.Error("missing node_config", zap.Any("node_config", nodeConfig))
	}

	state, exist := nodeInfo["state"].(string)
	if !exist {
		s.logger.Error("missing node state", zap.String("state", state))
		state = raft.Shutdown.String()
	}

	resp.Metadata = nodeConfigAny
	resp.State = state

	return resp, nil
}

func (s *GRPCService) SetNode(ctx context.Context, req *protobuf.SetNodeRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	ins, err := protobuf.MarshalAny(req.Metadata)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	nodeConfig := *ins.(*map[string]interface{})

	if s.raftServer.IsLeader() {
		err = s.raftServer.SetNodeConfig(req.Id, nodeConfig)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.SetNode(req.Id, nodeConfig)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (s *GRPCService) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	if s.raftServer.IsLeader() {
		err := s.raftServer.DeleteNodeConfig(req.Id)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.DeleteNode(req.Id)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*protobuf.GetClusterResponse, error) {
	resp := &protobuf.GetClusterResponse{}

	servers, err := s.raftServer.GetServers()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	cluster := map[string]interface{}{}
	for id := range servers {
		nodeResp, err := s.GetNode(ctx, &protobuf.GetNodeRequest{Id: id})
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}

		nodeConfigIntr, err := protobuf.MarshalAny(nodeResp.Metadata)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
		nodeConfig := *nodeConfigIntr.(*map[string]interface{})

		node := map[string]interface{}{
			"node_config": nodeConfig,
			"state":       nodeResp.State,
		}

		cluster[id] = node
	}

	clusterAny := &any.Any{}
	err = protobuf.UnmarshalAny(cluster, clusterAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Cluster = clusterAny

	return resp, nil
}

func (s *GRPCService) WatchCluster(req *empty.Empty, server protobuf.Blast_WatchClusterServer) error {
	chans := make(chan protobuf.GetClusterResponse)

	s.clusterMutex.Lock()
	s.clusterChans[chans] = struct{}{}
	s.clusterMutex.Unlock()

	defer func() {
		s.clusterMutex.Lock()
		delete(s.clusterChans, chans)
		s.clusterMutex.Unlock()
		close(chans)
	}()

	for resp := range chans {
		err := server.Send(&resp)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (s *GRPCService) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	resp := &empty.Empty{}

	err := s.raftServer.Snapshot()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetDocument(ctx context.Context, req *protobuf.GetDocumentRequest) (*protobuf.GetDocumentResponse, error) {
	resp := &protobuf.GetDocumentResponse{}

	fields, err := s.raftServer.GetDocument(req.Id)
	if err != nil {
		s.logger.Error(err.Error())
		switch err {
		case blasterrors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	fieldsAny := &any.Any{}
	err = protobuf.UnmarshalAny(fields, fieldsAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Fields = fieldsAny

	return resp, nil
}

func (s *GRPCService) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	resp := &protobuf.SearchResponse{}

	searchRequest, err := protobuf.MarshalAny(req.SearchRequest)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.InvalidArgument, err.Error())
	}

	searchResult, err := s.raftServer.Search(searchRequest.(*bleve.SearchRequest))
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	searchResultAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchResult, searchResultAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.SearchResult = searchResultAny

	return resp, nil
}

func (s *GRPCService) IndexDocument(stream protobuf.Blast_IndexDocumentServer) error {
	docs := make([]map[string]interface{}, 0)

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				s.logger.Debug(err.Error())
				break
			}
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}

		// fields
		ins, err := protobuf.MarshalAny(req.Fields)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
		fields := *ins.(*map[string]interface{})

		// document
		doc := map[string]interface{}{
			"id":     req.Id,
			"fields": fields,
		}

		docs = append(docs, doc)
	}

	// index
	count := -1
	var err error
	if s.raftServer.IsLeader() {
		count, err = s.raftServer.IndexDocument(docs)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
		count, err = client.IndexDocument(docs)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
	}

	return stream.SendAndClose(
		&protobuf.IndexDocumentResponse{
			Count: int32(count),
		},
	)
}

func (s *GRPCService) DeleteDocument(stream protobuf.Blast_DeleteDocumentServer) error {
	ids := make([]string, 0)

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				s.logger.Debug(err.Error())
				break
			}
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}

		ids = append(ids, req.Id)
	}

	// delete
	count := -1
	var err error
	if s.raftServer.IsLeader() {
		count, err = s.raftServer.DeleteDocument(ids)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
		count, err = client.DeleteDocument(ids)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
	}

	return stream.SendAndClose(
		&protobuf.DeleteDocumentResponse{
			Count: int32(count),
		},
	)
}

func (s *GRPCService) GetIndexConfig(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexConfigResponse, error) {
	resp := &protobuf.GetIndexConfigResponse{}

	indexConfig, err := s.raftServer.GetIndexConfig()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	indexConfigAny := &any.Any{}
	err = protobuf.UnmarshalAny(indexConfig, indexConfigAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.IndexConfig = indexConfigAny

	return resp, nil
}

func (s *GRPCService) GetIndexStats(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexStatsResponse, error) {
	resp := &protobuf.GetIndexStatsResponse{}

	indexStats, err := s.raftServer.GetIndexStats()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	indexStatsAny := &any.Any{}
	err = protobuf.UnmarshalAny(indexStats, indexStatsAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.IndexStats = indexStatsAny

	return resp, nil
}
