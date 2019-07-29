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
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	//*grpc.Service

	raftServer *RaftServer
	logger     *zap.Logger

	updateClusterStopCh chan struct{}
	updateClusterDoneCh chan struct{}
	peers               map[string]interface{}
	peerClients         map[string]*GRPCClient
	cluster             map[string]interface{}
	clusterChans        map[chan management.ClusterInfoResponse]struct{}
	clusterMutex        sync.RWMutex

	stateChans map[chan management.WatchResponse]struct{}
	stateMutex sync.RWMutex
}

func NewGRPCService(raftServer *RaftServer, logger *zap.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: raftServer,
		logger:     logger,

		peers:        make(map[string]interface{}, 0),
		peerClients:  make(map[string]*GRPCClient, 0),
		cluster:      make(map[string]interface{}, 0),
		clusterChans: make(map[chan management.ClusterInfoResponse]struct{}),

		stateChans: make(map[chan management.WatchResponse]struct{}),
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
			}

			// notify current cluster
			if !reflect.DeepEqual(s.cluster, cluster) {
				// convert to GetClusterResponse for channel output
				clusterResp := &management.ClusterInfoResponse{}
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

				// keep current cluster
				s.logger.Debug("current cluster", zap.Any("cluster", cluster))
				s.cluster = cluster
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

func (s *GRPCService) NodeHealthCheck(ctx context.Context, req *management.NodeHealthCheckRequest) (*management.NodeHealthCheckResponse, error) {
	resp := &management.NodeHealthCheckResponse{}

	switch req.Probe {
	case management.NodeHealthCheckRequest_HEALTHINESS:
		resp.State = management.NodeHealthCheckResponse_HEALTHY
	case management.NodeHealthCheckRequest_LIVENESS:
		resp.State = management.NodeHealthCheckResponse_ALIVE
	case management.NodeHealthCheckRequest_READINESS:
		resp.State = management.NodeHealthCheckResponse_READY
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
		nodeInfo, err = peerClient.NodeInfo(id)
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

func (s *GRPCService) NodeInfo(ctx context.Context, req *management.NodeInfoRequest) (*management.NodeInfoResponse, error) {
	resp := &management.NodeInfoResponse{}

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
		err = client.ClusterJoin(id, nodeConfig)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func (s *GRPCService) ClusterJoin(ctx context.Context, req *management.ClusterJoinRequest) (*empty.Empty, error) {
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
		err = client.ClusterLeave(id)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func (s *GRPCService) ClusterLeave(ctx context.Context, req *management.ClusterLeaveRequest) (*empty.Empty, error) {
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

func (s *GRPCService) ClusterInfo(ctx context.Context, req *empty.Empty) (*management.ClusterInfoResponse, error) {
	resp := &management.ClusterInfoResponse{}

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

func (s *GRPCService) ClusterWatch(req *empty.Empty, server management.Management_ClusterWatchServer) error {
	chans := make(chan management.ClusterInfoResponse)

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

func (s *GRPCService) Get(ctx context.Context, req *management.GetRequest) (*management.GetResponse, error) {
	s.stateMutex.RLock()
	defer func() {
		s.stateMutex.RUnlock()
	}()

	resp := &management.GetResponse{}

	value, err := s.raftServer.GetValue(req.Key)
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

func (s *GRPCService) Set(ctx context.Context, req *management.SetRequest) (*empty.Empty, error) {
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
		err = s.raftServer.SetValue(req.Key, value)
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
		err = client.Set(req.Key, value)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	// notify
	for c := range s.stateChans {
		c <- management.WatchResponse{
			Command: management.WatchResponse_SET,
			Key:     req.Key,
			Value:   req.Value,
		}
	}

	return resp, nil
}

func (s *GRPCService) Delete(ctx context.Context, req *management.DeleteRequest) (*empty.Empty, error) {
	s.stateMutex.Lock()
	defer func() {
		s.stateMutex.Unlock()
	}()

	resp := &empty.Empty{}

	if s.raftServer.IsLeader() {
		err := s.raftServer.DeleteValue(req.Key)
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
		err = client.Delete(req.Key)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	// notify
	for c := range s.stateChans {
		c <- management.WatchResponse{
			Command: management.WatchResponse_DELETE,
			Key:     req.Key,
		}
	}

	return resp, nil
}

func (s *GRPCService) Watch(req *management.WatchRequest, server management.Management_WatchServer) error {
	chans := make(chan management.WatchResponse)

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
