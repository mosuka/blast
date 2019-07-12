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

package manager

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	*grpc.Service

	raftServer *RaftServer
	logger     *zap.Logger

	updateClusterStopCh chan struct{}
	updateClusterDoneCh chan struct{}
	peers               map[string]interface{}
	peerClients         map[string]*grpc.Client
	cluster             map[string]interface{}
	clusterChans        map[chan protobuf.GetClusterResponse]struct{}
	clusterMutex        sync.RWMutex

	stateChans map[chan protobuf.WatchStateResponse]struct{}
	stateMutex sync.RWMutex
}

func NewGRPCService(raftServer *RaftServer, logger *zap.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: raftServer,
		logger:     logger,

		peers:        make(map[string]interface{}, 0),
		peerClients:  make(map[string]*grpc.Client, 0),
		cluster:      make(map[string]interface{}, 0),
		clusterChans: make(map[chan protobuf.GetClusterResponse]struct{}),

		stateChans: make(map[chan protobuf.WatchStateResponse]struct{}),
	}, nil
}

func (s *GRPCService) Start() error {
	s.logger.Info("start to update cluster info")
	go s.startUpdateCluster(500 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Info("stop to update cluster info")
	s.stopUpdateCluster()

	return nil
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
			for id, metadata := range servers {
				if id != s.raftServer.nodeConfig.NodeId {
					peers[id] = metadata
				}
			}

			if !reflect.DeepEqual(s.peers, peers) {
				// open clients
				for nodeId, metadata := range peers {
					mm, ok := metadata.(map[string]interface{})
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
	s.stateMutex.Lock()
	defer func() {
		s.stateMutex.Unlock()
	}()

	resp := &empty.Empty{}

	err := s.raftServer.Snapshot()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetState(ctx context.Context, req *protobuf.GetStateRequest) (*protobuf.GetStateResponse, error) {
	s.stateMutex.RLock()
	defer func() {
		s.stateMutex.RUnlock()
	}()

	resp := &protobuf.GetStateResponse{}

	value, err := s.raftServer.GetState(req.Key)
	if err != nil {
		s.logger.Error(err.Error())
		switch err {
		case blasterrors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	valueAny := &any.Any{}
	err = protobuf.UnmarshalAny(value, valueAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Value = valueAny

	return resp, nil
}

func (s *GRPCService) SetState(ctx context.Context, req *protobuf.SetStateRequest) (*empty.Empty, error) {
	s.stateMutex.Lock()
	defer func() {
		s.stateMutex.Unlock()
	}()

	resp := &empty.Empty{}

	value, err := protobuf.MarshalAny(req.Value)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	if s.raftServer.IsLeader() {
		err = s.raftServer.SetState(req.Key, value)
		if err != nil {
			s.logger.Error(err.Error())
			switch err {
			case blasterrors.ErrNotFound:
				return resp, status.Error(codes.NotFound, err.Error())
			default:
				return resp, status.Error(codes.Internal, err.Error())
			}
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.SetState(req.Key, value)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	// notify
	for c := range s.stateChans {
		c <- protobuf.WatchStateResponse{
			Command: protobuf.WatchStateResponse_SET,
			Key:     req.Key,
			Value:   req.Value,
		}
	}

	return resp, nil
}

func (s *GRPCService) DeleteState(ctx context.Context, req *protobuf.DeleteStateRequest) (*empty.Empty, error) {
	s.stateMutex.Lock()
	defer func() {
		s.stateMutex.Unlock()
	}()

	resp := &empty.Empty{}

	if s.raftServer.IsLeader() {
		err := s.raftServer.DeleteState(req.Key)
		if err != nil {
			s.logger.Error(err.Error())
			switch err {
			case blasterrors.ErrNotFound:
				return resp, status.Error(codes.NotFound, err.Error())
			default:
				return resp, status.Error(codes.Internal, err.Error())
			}
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.DeleteState(req.Key)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	// notify
	for c := range s.stateChans {
		c <- protobuf.WatchStateResponse{
			Command: protobuf.WatchStateResponse_DELETE,
			Key:     req.Key,
		}
	}

	return resp, nil
}

func (s *GRPCService) WatchState(req *protobuf.WatchStateRequest, server protobuf.Blast_WatchStateServer) error {
	chans := make(chan protobuf.WatchStateResponse)

	s.stateMutex.Lock()
	s.stateChans[chans] = struct{}{}
	s.stateMutex.Unlock()

	defer func() {
		s.stateMutex.Lock()
		delete(s.stateChans, chans)
		s.stateMutex.Unlock()
		close(chans)
	}()

	for resp := range chans {
		if !strings.HasPrefix(resp.Key, req.Key) {
			continue
		}
		err := server.Send(&resp)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}
