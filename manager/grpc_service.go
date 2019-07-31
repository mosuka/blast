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
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	raftServer *RaftServer
	logger     *zap.Logger

	updateClusterStopCh chan struct{}
	updateClusterDoneCh chan struct{}
	peers               *management.Cluster
	peerClients         map[string]*GRPCClient
	cluster             *management.Cluster
	clusterChans        map[chan management.ClusterWatchResponse]struct{}
	clusterMutex        sync.RWMutex

	stateChans map[chan management.WatchResponse]struct{}
	stateMutex sync.RWMutex
}

func NewGRPCService(raftServer *RaftServer, logger *zap.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: raftServer,
		logger:     logger,

		peers:        &management.Cluster{Nodes: make(map[string]*management.Node, 0)},
		peerClients:  make(map[string]*GRPCClient, 0),
		cluster:      &management.Cluster{Nodes: make(map[string]*management.Node, 0)},
		clusterChans: make(map[chan management.ClusterWatchResponse]struct{}),

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
	for id, node := range s.cluster.Nodes {
		switch node.State {
		case management.Node_LEADER:
			if client, exist := s.peerClients[id]; exist {
				return client, nil
			}
		}
	}

	err := errors.New("there is no leader")
	s.logger.Error(err.Error())
	return nil, err
}

func (s *GRPCService) cloneCluster(cluster *management.Cluster) (*management.Cluster, error) {
	b, err := json.Marshal(cluster)
	if err != nil {
		return nil, err
	}

	var clone *management.Cluster
	err = json.Unmarshal(b, &clone)
	if err != nil {
		return nil, err
	}

	return clone, nil
}

func (s *GRPCService) startUpdateCluster(checkInterval time.Duration) {
	s.updateClusterStopCh = make(chan struct{})
	s.updateClusterDoneCh = make(chan struct{})

	defer func() {
		close(s.updateClusterDoneCh)
	}()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	savedCluster, err := s.cloneCluster(s.cluster)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	for {
		select {
		case <-s.updateClusterStopCh:
			s.logger.Info("received a request to stop updating a cluster")
			return
		case <-ticker.C:
			s.cluster, err = s.getCluster()
			if err != nil {
				s.logger.Error(err.Error())
				return
			}

			snapshotCluster, err := s.cloneCluster(s.cluster)
			if err != nil {
				s.logger.Error(err.Error())
				return
			}

			// create peer node list with out self node
			for id, node := range snapshotCluster.Nodes {
				if id != s.NodeID() {
					s.peers.Nodes[id] = node
				}
			}

			// open clients for peer nodes
			for id, node := range s.peers.Nodes {
				if node.Metadata.GrpcAddress == "" {
					s.logger.Debug("missing gRPC address", zap.String("id", id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
					continue
				}

				client, exist := s.peerClients[id]
				if exist {
					if client.GetAddress() != node.Metadata.GrpcAddress {
						s.logger.Info("recreate gRPC client", zap.String("id", id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
						delete(s.peerClients, id)
						err = client.Close()
						if err != nil {
							s.logger.Warn(err.Error(), zap.String("id", id))
						}
						newClient, err := NewGRPCClient(node.Metadata.GrpcAddress)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("id", id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
							continue
						}
						s.peerClients[id] = newClient
					}
				} else {
					s.logger.Info("create gRPC client", zap.String("id", id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
					newClient, err := NewGRPCClient(node.Metadata.GrpcAddress)
					if err != nil {
						s.logger.Warn(err.Error(), zap.String("id", id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
						continue
					}
					s.peerClients[id] = newClient
				}
			}

			// close clients for non-existent peer nodes
			for id, client := range s.peerClients {
				if _, exist := s.peers.Nodes[id]; !exist {
					s.logger.Info("close gRPC client", zap.String("id", id), zap.String("grpc_addr", client.GetAddress()))
					err = client.Close()
					if err != nil {
						s.logger.Warn(err.Error(), zap.String("id", id), zap.String("grpc_addr", client.GetAddress()))
					}
					delete(s.peerClients, id)
				}
			}

			// check joined and updated nodes
			for id, node := range snapshotCluster.Nodes {
				nodeSnapshot, exist := savedCluster.Nodes[id]
				if exist {
					// node exists in the cluster
					n1, err := json.Marshal(node)
					if err != nil {
						s.logger.Warn(err.Error(), zap.String("id", id), zap.Any("node", node))
					}
					n2, err := json.Marshal(nodeSnapshot)
					if err != nil {
						s.logger.Warn(err.Error(), zap.String("id", id), zap.Any("node", nodeSnapshot))
					}
					if !cmp.Equal(n1, n2) {
						// node updated
						// notify the cluster changes
						clusterResp := &management.ClusterWatchResponse{
							Event:   management.ClusterWatchResponse_UPDATE,
							Id:      id,
							Node:    node,
							Cluster: snapshotCluster,
						}
						for c := range s.clusterChans {
							c <- *clusterResp
						}
					}
				} else {
					// node joined
					// notify the cluster changes
					clusterResp := &management.ClusterWatchResponse{
						Event:   management.ClusterWatchResponse_JOIN,
						Id:      id,
						Node:    node,
						Cluster: snapshotCluster,
					}
					for c := range s.clusterChans {
						c <- *clusterResp
					}
				}
			}

			// check left nodes
			for id, node := range savedCluster.Nodes {
				if _, exist := snapshotCluster.Nodes[id]; !exist {
					// node left
					// notify the cluster changes
					clusterResp := &management.ClusterWatchResponse{
						Event:   management.ClusterWatchResponse_LEAVE,
						Id:      id,
						Node:    node,
						Cluster: snapshotCluster,
					}
					for c := range s.clusterChans {
						c <- *clusterResp
					}
				}
			}

			savedCluster = snapshotCluster
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

func (s *GRPCService) getSelfNode() *management.Node {
	node := s.raftServer.node

	switch s.raftServer.State() {
	case raft.Follower:
		node.State = management.Node_FOLLOWER
	case raft.Candidate:
		node.State = management.Node_CANDIDATE
	case raft.Leader:
		node.State = management.Node_LEADER
	case raft.Shutdown:
		node.State = management.Node_SHUTDOWN
	default:
		node.State = management.Node_UNKNOWN
	}

	return node
}

func (s *GRPCService) getPeerNode(id string) (*management.Node, error) {
	if _, exist := s.peerClients[id]; !exist {
		err := errors.New("node does not exist in peers")
		s.logger.Debug(err.Error(), zap.String("id", id))
		return nil, err
	}

	node, err := s.peerClients[id].NodeInfo()
	if err != nil {
		s.logger.Debug(err.Error(), zap.String("id", id))
		return &management.Node{
			BindAddress: "",
			State:       management.Node_UNKNOWN,
			Metadata: &management.Metadata{
				GrpcAddress: "",
				HttpAddress: "",
			},
		}, nil
	}

	return node, nil
}

func (s *GRPCService) getNode(id string) (*management.Node, error) {
	if id == "" || id == s.NodeID() {
		return s.getSelfNode(), nil
	} else {
		return s.getPeerNode(id)
	}
}

func (s *GRPCService) NodeInfo(ctx context.Context, req *empty.Empty) (*management.NodeInfoResponse, error) {
	resp := &management.NodeInfoResponse{}

	node, err := s.getNode(s.NodeID())
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return &management.NodeInfoResponse{
		Node: node,
	}, nil
}

func (s *GRPCService) setNode(id string, node *management.Node) error {
	if s.raftServer.IsLeader() {
		err := s.raftServer.SetNode(id, node)
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
		err = client.ClusterJoin(id, node)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func (s *GRPCService) ClusterJoin(ctx context.Context, req *management.ClusterJoinRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	err := s.setNode(req.Id, req.Node)
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

func (s *GRPCService) getCluster() (*management.Cluster, error) {
	cluster, err := s.raftServer.GetCluster()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	// update latest node state
	for id := range cluster.Nodes {
		node, err := s.getNode(id)
		if err != nil {
			s.logger.Debug(err.Error())
			continue
		}
		cluster.Nodes[id] = node
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

	resp.Cluster = cluster

	return resp, nil
}

func (s *GRPCService) ClusterWatch(req *empty.Empty, server management.Management_ClusterWatchServer) error {
	chans := make(chan management.ClusterWatchResponse)

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
