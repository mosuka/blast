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
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/management"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	clusterConfig *config.ClusterConfig
	raftServer    *RaftServer
	logger        *zap.Logger

	updateClusterStopCh chan struct{}
	updateClusterDoneCh chan struct{}
	peers               map[string]interface{}
	peerClients         map[string]*GRPCClient
	cluster             map[string]interface{}
	clusterChans        map[chan index.GetClusterResponse]struct{}
	clusterMutex        sync.RWMutex

	managers             *management.Cluster
	managerClients       map[string]*manager.GRPCClient
	updateManagersStopCh chan struct{}
	updateManagersDoneCh chan struct{}
}

func NewGRPCService(clusterConfig *config.ClusterConfig, raftServer *RaftServer, logger *zap.Logger) (*GRPCService, error) {
	return &GRPCService{
		clusterConfig: clusterConfig,
		raftServer:    raftServer,
		logger:        logger,

		peers:        make(map[string]interface{}, 0),
		peerClients:  make(map[string]*GRPCClient, 0),
		cluster:      make(map[string]interface{}, 0),
		clusterChans: make(map[chan index.GetClusterResponse]struct{}),

		managers:       &management.Cluster{Nodes: make(map[string]*management.Node, 0)},
		managerClients: make(map[string]*manager.GRPCClient, 0),
	}, nil
}

func (s *GRPCService) Start() error {
	s.logger.Info("start to update cluster info")
	go s.startUpdateCluster(500 * time.Millisecond)

	if s.clusterConfig.ManagerAddr != "" {
		s.logger.Info("start to update manager cluster info")
		go s.startUpdateManagers(500 * time.Millisecond)
	}

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Info("stop to update cluster info")
	s.stopUpdateCluster()

	if s.clusterConfig.ManagerAddr != "" {
		s.logger.Info("stop to update manager cluster info")
		s.stopUpdateManagers()
	}

	return nil
}

func (s *GRPCService) getManagerClient() (*manager.GRPCClient, error) {
	var client *manager.GRPCClient

	for id, node := range s.managers.Nodes {
		if node.Metadata == nil {
			s.logger.Warn("assertion failed", zap.String("id", id))
			continue
		}

		if node.Status == raft.Leader.String() || node.Status == raft.Follower.String() {
			var ok bool
			client, ok = s.managerClients[id]
			if ok {
				return client, nil
			} else {
				s.logger.Error("node does not exist", zap.String("id", id))
			}
		} else {
			s.logger.Debug("node has not available", zap.String("id", id), zap.String("state", node.Status))
		}
	}

	err := errors.New("available client does not exist")
	s.logger.Error(err.Error())

	return nil, err
}

func (s *GRPCService) getInitialManagers(managerAddr string) (*management.Cluster, error) {
	client, err := manager.NewGRPCClient(managerAddr)
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

	managers, err := client.ClusterInfo()
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

	var err error

	// get initial managers
	s.managers, err = s.getInitialManagers(s.clusterConfig.ManagerAddr)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.logger.Debug("initialize manager list", zap.Any("managers", s.managers))

	// create clients for managers
	for nodeId, node := range s.managers.Nodes {
		if node.Metadata == nil {
			s.logger.Warn("missing metadata", zap.String("id", nodeId))
			continue
		}

		if node.Metadata.GrpcAddress == "" {
			s.logger.Warn("missing gRPC address", zap.String("id", nodeId), zap.String("grpc_addr", node.Metadata.GrpcAddress))
			continue
		}

		s.logger.Debug("create gRPC client", zap.String("id", nodeId), zap.String("grpc_addr", node.Metadata.GrpcAddress))
		client, err := manager.NewGRPCClient(node.Metadata.GrpcAddress)
		if err != nil {
			s.logger.Error(err.Error(), zap.String("id", nodeId), zap.String("grpc_addr", node.Metadata.GrpcAddress))
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

			stream, err := client.ClusterWatch()
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
			managers := resp.Cluster

			if !reflect.DeepEqual(s.managers, managers) {
				// open clients
				for nodeId, nodeConfig := range managers.Nodes {
					if nodeConfig.Metadata == nil {
						s.logger.Warn("missing metadata", zap.String("node_id", nodeId))
						continue
					}

					if nodeConfig.Metadata.GrpcAddress == "" {
						s.logger.Warn("missing metadata", zap.String("node_id", nodeId), zap.String("grpc_addr", nodeConfig.Metadata.GrpcAddress))
						continue
					}

					client, exist := s.managerClients[nodeId]
					if exist {
						s.logger.Debug("client has already exist in manager list", zap.String("id", nodeId))

						if client.GetAddress() != nodeConfig.Metadata.GrpcAddress {
							s.logger.Debug("gRPC address has been changed", zap.String("node_id", nodeId), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", nodeConfig.Metadata.GrpcAddress))
							s.logger.Debug("recreate gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", nodeConfig.Metadata.GrpcAddress))

							delete(s.managerClients, nodeId)

							err = client.Close()
							if err != nil {
								s.logger.Error(err.Error(), zap.String("node_id", nodeId))
							}

							newClient, err := manager.NewGRPCClient(nodeConfig.Metadata.GrpcAddress)
							if err != nil {
								s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", nodeConfig.Metadata.GrpcAddress))
							}

							if newClient != nil {
								s.managerClients[nodeId] = newClient
							}
						} else {
							s.logger.Debug("gRPC address has not changed", zap.String("node_id", nodeId), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", nodeConfig.Metadata.GrpcAddress))
						}
					} else {
						s.logger.Debug("client does not exist in peer list", zap.String("node_id", nodeId))

						s.logger.Debug("create gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", nodeConfig.Metadata.GrpcAddress))
						newClient, err := manager.NewGRPCClient(nodeConfig.Metadata.GrpcAddress)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", nodeConfig.Metadata.GrpcAddress))
						}
						if newClient != nil {
							s.managerClients[nodeId] = newClient
						}
					}
				}

				// close nonexistent clients
				for nodeId, client := range s.managerClients {
					if nodeConfig, exist := managers.Nodes[nodeId]; !exist {
						s.logger.Info("this client is no longer in use", zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))

						s.logger.Debug("close client", zap.String("node_id", nodeId), zap.String("address", client.GetAddress()))
						err = client.Close()
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("address", client.GetAddress()))
						}

						s.logger.Debug("delete client", zap.String("node_id", nodeId))
						delete(s.managerClients, nodeId)
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

func (s *GRPCService) getLeaderClient() (*GRPCClient, error) {
	var client *GRPCClient

	for id, node := range s.cluster {
		state, ok := node.(map[string]interface{})["state"].(string)
		if !ok {
			s.logger.Warn("missing state", zap.String("id", id), zap.String("state", state))
			continue
		}

		if state == raft.Leader.String() {
			client, ok = s.peerClients[id]
			if ok {
				break
			} else {
				s.logger.Error("node does not exist", zap.String("id", id))
			}
		} else {
			s.logger.Debug("not a leader", zap.String("id", id))
		}
	}

	if client == nil {
		err := errors.New("there is no leader")
		s.logger.Error(err.Error())
		return nil, err
	}

	return client, nil
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
			cluster, err := s.getCluster()
			if err != nil {
				s.logger.Error(err.Error())
				return
			}

			// create peer node list with out self node
			peers := make(map[string]interface{}, 0)
			for nodeId, node := range cluster {
				if nodeId != s.NodeID() {
					peers[nodeId] = node
				}
			}

			if !reflect.DeepEqual(s.peers, peers) {
				// open clients
				for nodeId, nodeInfo := range peers {
					nodeConfig, ok := nodeInfo.(map[string]interface{})["node_config"].(map[string]interface{})
					if !ok {
						s.logger.Warn("assertion failed", zap.String("node_id", nodeId), zap.Any("node_info", nodeInfo))
						continue
					}
					grpcAddr, ok := nodeConfig["grpc_addr"].(string)
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
								s.logger.Warn(err.Error(), zap.String("node_id", nodeId))
							}

							newClient, err := NewGRPCClient(grpcAddr)
							if err != nil {
								s.logger.Warn(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
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
						peerClient, err := NewGRPCClient(grpcAddr)
						if err != nil {
							s.logger.Warn(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
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
							s.logger.Warn(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", client.GetAddress()))
						}

						s.logger.Debug("delete client", zap.String("node_id", nodeId))
						delete(s.peerClients, nodeId)
					}
				}

				// keep current peer nodes
				s.logger.Debug("current peers", zap.Any("peers", peers))
				s.peers = peers
			} else {
				s.logger.Debug("there is no change in peers", zap.Any("peers", peers))
			}

			// notify current cluster
			if !reflect.DeepEqual(s.cluster, cluster) {
				// convert to GetClusterResponse for channel output
				clusterResp := &index.GetClusterResponse{}
				clusterAny := &any.Any{}
				err = protobuf.UnmarshalAny(cluster, clusterAny)
				if err != nil {
					s.logger.Warn(err.Error())
				}
				clusterResp.Cluster = clusterAny

				// output to channel
				for c := range s.clusterChans {
					c <- *clusterResp
				}

				// notify cluster config to manager
				if s.clusterConfig.ManagerAddr != "" && s.raftServer.IsLeader() {
					client, err := s.getManagerClient()
					if err != nil {
						s.logger.Error(err.Error())
					}
					err = client.Set(fmt.Sprintf("cluster_config/clusters/%s/nodes", s.clusterConfig.ClusterId), cluster)
					if err != nil {
						s.logger.Error(err.Error())
					}
				}

				// keep current cluster
				s.logger.Debug("current cluster", zap.Any("cluster", cluster))
				s.cluster = cluster
			} else {
				s.logger.Debug("there is no change in cluster", zap.Any("cluster", cluster))
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
			s.logger.Warn(err.Error())
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

func (s *GRPCService) LivenessProbe(ctx context.Context, req *empty.Empty) (*index.LivenessProbeResponse, error) {
	resp := &index.LivenessProbeResponse{
		State: index.LivenessProbeResponse_ALIVE,
	}

	return resp, nil
}

func (s *GRPCService) ReadinessProbe(ctx context.Context, req *empty.Empty) (*index.ReadinessProbeResponse, error) {
	resp := &index.ReadinessProbeResponse{
		State: index.ReadinessProbeResponse_READY,
	}

	return resp, nil
}

func (s *GRPCService) NodeID() string {
	return s.raftServer.NodeID()
}

func (s *GRPCService) getSelfNode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"node_config": s.raftServer.nodeConfig.ToMap(),
		"state":       s.raftServer.State().String(),
	}, nil
}

func (s *GRPCService) getPeerNode(id string) (map[string]interface{}, error) {
	var nodeInfo map[string]interface{}
	var err error

	if peerClient, exist := s.peerClients[id]; exist {
		nodeInfo, err = peerClient.GetNode(id)
		if err != nil {
			s.logger.Warn(err.Error())
			nodeInfo = map[string]interface{}{
				"node_config": map[string]interface{}{},
				"state":       raft.Shutdown.String(),
			}
		}
	} else {
		s.logger.Warn("node does not exist in peer list", zap.String("id", id))
		nodeInfo = map[string]interface{}{
			"node_config": map[string]interface{}{},
			"state":       raft.Shutdown.String(),
		}
	}

	return nodeInfo, nil
}

func (s *GRPCService) getNode(id string) (map[string]interface{}, error) {
	var nodeInfo map[string]interface{}
	var err error

	if id == "" || id == s.NodeID() {
		nodeInfo, err = s.getSelfNode()
	} else {
		nodeInfo, err = s.getPeerNode(id)
	}

	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return nodeInfo, nil
}

func (s *GRPCService) GetNode(ctx context.Context, req *index.GetNodeRequest) (*index.GetNodeResponse, error) {
	resp := &index.GetNodeResponse{}

	nodeInfo, err := s.getNode(req.Id)
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

	resp.NodeConfig = nodeConfigAny
	resp.State = state

	return resp, nil
}

func (s *GRPCService) setNode(id string, nodeConfig map[string]interface{}) error {
	if s.raftServer.IsLeader() {
		err := s.raftServer.SetNode(id, nodeConfig)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
		err = client.SetNode(id, nodeConfig)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func (s *GRPCService) SetNode(ctx context.Context, req *index.SetNodeRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	ins, err := protobuf.MarshalAny(req.NodeConfig)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	nodeConfig := *ins.(*map[string]interface{})

	err = s.setNode(req.Id, nodeConfig)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) deleteNode(id string) error {
	if s.raftServer.IsLeader() {
		err := s.raftServer.DeleteNode(id)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
		err = client.DeleteNode(id)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func (s *GRPCService) DeleteNode(ctx context.Context, req *index.DeleteNodeRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	err := s.deleteNode(req.Id)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) getCluster() (map[string]interface{}, error) {
	cluster, err := s.raftServer.GetCluster()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	// update node state
	for nodeId := range cluster {
		node, err := s.getNode(nodeId)
		if err != nil {
			s.logger.Error(err.Error())
		}
		state := node["state"].(string)

		if _, ok := cluster[nodeId]; !ok {
			cluster[nodeId] = map[string]interface{}{}
		}
		nodeInfo := cluster[nodeId].(map[string]interface{})
		nodeInfo["state"] = state
	}

	return cluster, nil
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*index.GetClusterResponse, error) {
	resp := &index.GetClusterResponse{}

	cluster, err := s.getCluster()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
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

func (s *GRPCService) WatchCluster(req *empty.Empty, server index.Index_WatchClusterServer) error {
	chans := make(chan index.GetClusterResponse)

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

func (s *GRPCService) GetDocument(ctx context.Context, req *index.GetDocumentRequest) (*index.GetDocumentResponse, error) {
	resp := &index.GetDocumentResponse{}

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

func (s *GRPCService) Search(ctx context.Context, req *index.SearchRequest) (*index.SearchResponse, error) {
	resp := &index.SearchResponse{}

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

func (s *GRPCService) IndexDocument(stream index.Index_IndexDocumentServer) error {
	docs := make([]*indexutils.Document, 0)

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
		doc, err := indexutils.NewDocument(req.Id, fields)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
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
		&index.IndexDocumentResponse{
			Count: int32(count),
		},
	)
}

func (s *GRPCService) DeleteDocument(stream index.Index_DeleteDocumentServer) error {
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
		&index.DeleteDocumentResponse{
			Count: int32(count),
		},
	)
}

func (s *GRPCService) GetIndexConfig(ctx context.Context, req *empty.Empty) (*index.GetIndexConfigResponse, error) {
	resp := &index.GetIndexConfigResponse{}

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

func (s *GRPCService) GetIndexStats(ctx context.Context, req *empty.Empty) (*index.GetIndexStatsResponse, error) {
	resp := &index.GetIndexStatsResponse{}

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

func (s *GRPCService) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	resp := &empty.Empty{}

	err := s.raftServer.Snapshot()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}
