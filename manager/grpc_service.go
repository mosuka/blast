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
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/hashutils"
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
	clusterChans        map[chan management.ClusterInfoResponse]struct{}
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

	for id, node := range s.cluster.Nodes {
		state := node.State
		if node.State == "" {
			s.logger.Warn("missing state", zap.String("id", id), zap.String("state", state))
			continue
		}

		if state == raft.Leader.String() {
			var ok bool
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

	// create initial cluster hash
	clusterHash, err := hashutils.Hash(s.cluster)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	peersHash, err := hashutils.Hash(s.peers)
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

			// create latest cluster hash
			newClusterHash, err := hashutils.Hash(s.cluster)
			if err != nil {
				s.logger.Error(err.Error())
				return
			}

			// create peer node list with out self node
			for id, node := range s.cluster.Nodes {
				if id != s.NodeID() {
					s.peers.Nodes[id] = node
				}
			}

			// create latest peers hash
			newPeersHash, err := hashutils.Hash(s.peers)
			if err != nil {
				s.logger.Error(err.Error())
				return
			}

			// compare peers hash
			if !cmp.Equal(peersHash, newPeersHash) {
				// open clients
				for id, node := range s.peers.Nodes {
					if node.Metadata.GrpcAddress == "" {
						s.logger.Warn("missing gRPC address", zap.String("id", id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
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

				// close client for non-existent node
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

				// update peers hash
				peersHash = newPeersHash
			}

			// compare cluster hash
			if !cmp.Equal(clusterHash, newClusterHash) {
				clusterResp := &management.ClusterInfoResponse{
					Cluster: s.cluster,
				}

				// output to channel
				for c := range s.clusterChans {
					c <- *clusterResp
				}

				// update cluster hash
				clusterHash = newClusterHash
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

func (s *GRPCService) getSelfNode() (*management.Node, error) {
	node := s.raftServer.node
	node.State = s.raftServer.State().String()

	return node, nil
}

func (s *GRPCService) getPeerNode(id string) (*management.Node, error) {
	var nodeInfo *management.Node
	var err error

	if peerClient, exist := s.peerClients[id]; exist {
		nodeInfo, err = peerClient.NodeInfo()
		if err != nil {
			s.logger.Debug(err.Error())
			nodeInfo = &management.Node{
				BindAddress: "",
				State:       raft.Shutdown.String(),
				Metadata:    &management.Metadata{},
			}
		}
	} else {
		nodeInfo = &management.Node{
			BindAddress: "",
			State:       raft.Shutdown.String(),
			Metadata:    &management.Metadata{},
		}
	}

	return nodeInfo, nil
}

func (s *GRPCService) getNode(id string) (*management.Node, error) {
	var nodeInfo *management.Node
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

func (s *GRPCService) NodeInfo(ctx context.Context, req *empty.Empty) (*management.NodeInfoResponse, error) {
	resp := &management.NodeInfoResponse{}

	nodeInfo, err := s.getNode(s.NodeID())
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Node = nodeInfo

	return resp, nil
}

func (s *GRPCService) setNode(id string, nodeConfig *management.Node) error {
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
	for nodeId := range cluster.Nodes {
		node, err := s.getNode(nodeId)
		if err != nil {
			s.logger.Warn(err.Error())
			continue
		}
		cluster.Nodes[nodeId].State = node.State
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
