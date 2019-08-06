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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/management"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	managerGrpcAddress string
	shardId            string
	raftServer         *RaftServer
	logger             *zap.Logger

	updateClusterStopCh chan struct{}
	updateClusterDoneCh chan struct{}
	peers               *index.Cluster
	peerClients         map[string]*GRPCClient
	cluster             *index.Cluster
	clusterChans        map[chan index.ClusterWatchResponse]struct{}
	clusterMutex        sync.RWMutex

	managers             *management.Cluster
	managerClients       map[string]*manager.GRPCClient
	updateManagersStopCh chan struct{}
	updateManagersDoneCh chan struct{}
}

func NewGRPCService(managerGrpcAddress string, shardId string, raftServer *RaftServer, logger *zap.Logger) (*GRPCService, error) {
	return &GRPCService{
		managerGrpcAddress: managerGrpcAddress,
		shardId:            shardId,
		raftServer:         raftServer,
		logger:             logger,

		peers:        &index.Cluster{Nodes: make(map[string]*index.Node, 0)},
		peerClients:  make(map[string]*GRPCClient, 0),
		cluster:      &index.Cluster{Nodes: make(map[string]*index.Node, 0)},
		clusterChans: make(map[chan index.ClusterWatchResponse]struct{}),

		managers:       &management.Cluster{Nodes: make(map[string]*management.Node, 0)},
		managerClients: make(map[string]*manager.GRPCClient, 0),
	}, nil
}

func (s *GRPCService) Start() error {
	if s.managerGrpcAddress != "" {
		var err error
		s.managers, err = s.getManagerCluster(s.managerGrpcAddress)
		if err != nil {
			s.logger.Fatal(err.Error())
			return err
		}

		for id, node := range s.managers.Nodes {
			client, err := manager.NewGRPCClient(node.Metadata.GrpcAddress)
			if err != nil {
				s.logger.Fatal(err.Error(), zap.String("id", id), zap.String("grpc_address", s.managerGrpcAddress))
			}
			s.managerClients[node.Id] = client
		}

		s.logger.Info("start to update manager cluster info")
		go s.startUpdateManagers(500 * time.Millisecond)
	}

	s.logger.Info("start to update cluster info")
	go s.startUpdateCluster(500 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Info("stop to update cluster info")
	s.stopUpdateCluster()

	if s.managerGrpcAddress != "" {
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

		if node.State == management.Node_FOLLOWER || node.State == management.Node_LEADER {
			var ok bool
			client, ok = s.managerClients[id]
			if ok {
				return client, nil
			} else {
				s.logger.Error("node does not exist", zap.String("id", id))
			}
		} else {
			s.logger.Debug("node has not available", zap.String("id", id), zap.String("state", node.State.String()))
		}
	}

	err := errors.New("available client does not exist")
	s.logger.Error(err.Error())

	return nil, err
}

func (s *GRPCService) getManagerCluster(managerAddr string) (*management.Cluster, error) {
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

func (s *GRPCService) cloneManagerCluster(cluster *management.Cluster) (*management.Cluster, error) {
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

func (s *GRPCService) startUpdateManagers(checkInterval time.Duration) {
	s.updateManagersStopCh = make(chan struct{})
	s.updateManagersDoneCh = make(chan struct{})

	defer func() {
		close(s.updateManagersDoneCh)
	}()

	for {
		select {
		case <-s.updateManagersStopCh:
			s.logger.Info("received a request to stop updating a manager cluster")
			return
		default:
			// get client for manager from the list
			client, err := s.getManagerClient()
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}

			// create stream for watching cluster changes
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
			s.logger.Info("cluster has changed", zap.Any("resp", resp))
			switch resp.Event {
			case management.ClusterWatchResponse_JOIN, management.ClusterWatchResponse_UPDATE:
				// add to cluster nodes
				s.managers.Nodes[resp.Node.Id] = resp.Node

				// check node state
				switch resp.Node.State {
				case management.Node_UNKNOWN, management.Node_SHUTDOWN:
					// close client
					if client, exist := s.managerClients[resp.Node.Id]; exist {
						s.logger.Info("close gRPC client", zap.String("id", resp.Node.Id), zap.String("grpc_addr", client.GetAddress()))
						err = client.Close()
						if err != nil {
							s.logger.Warn(err.Error(), zap.String("id", resp.Node.Id))
						}
						delete(s.managerClients, resp.Node.Id)
					}
				default: // management.Node_FOLLOWER, management.Node_CANDIDATE, management.Node_LEADER
					if resp.Node.Metadata.GrpcAddress == "" {
						s.logger.Warn("missing gRPC address", zap.String("id", resp.Node.Id), zap.String("grpc_addr", resp.Node.Metadata.GrpcAddress))
						continue
					}

					// check client that already exist in the client list
					if client, exist := s.managerClients[resp.Node.Id]; !exist {
						// create new client
						s.logger.Info("create gRPC client", zap.String("id", resp.Node.Id), zap.String("grpc_addr", resp.Node.Metadata.GrpcAddress))
						newClient, err := manager.NewGRPCClient(resp.Node.Metadata.GrpcAddress)
						if err != nil {
							s.logger.Warn(err.Error(), zap.String("id", resp.Node.Id), zap.String("grpc_addr", resp.Node.Metadata.GrpcAddress))
							continue
						}
						s.managerClients[resp.Node.Id] = newClient
					} else {
						if client.GetAddress() != resp.Node.Metadata.GrpcAddress {
							// close client
							s.logger.Info("close gRPC client", zap.String("id", resp.Node.Id), zap.String("grpc_addr", client.GetAddress()))
							err = client.Close()
							if err != nil {
								s.logger.Warn(err.Error(), zap.String("id", resp.Node.Id))
							}
							delete(s.managerClients, resp.Node.Id)

							// re-create new client
							s.logger.Info("re-create gRPC client", zap.String("id", resp.Node.Id), zap.String("grpc_addr", resp.Node.Metadata.GrpcAddress))
							newClient, err := manager.NewGRPCClient(resp.Node.Metadata.GrpcAddress)
							if err != nil {
								s.logger.Warn(err.Error(), zap.String("id", resp.Node.Id), zap.String("grpc_addr", resp.Node.Metadata.GrpcAddress))
								continue
							}
							s.managerClients[resp.Node.Id] = newClient
						}
					}
				}
			case management.ClusterWatchResponse_LEAVE:
				if client, exist := s.managerClients[resp.Node.Id]; exist {
					s.logger.Info("close gRPC client", zap.String("id", resp.Node.Id), zap.String("grpc_addr", client.GetAddress()))
					err = client.Close()
					if err != nil {
						s.logger.Warn(err.Error(), zap.String("id", resp.Node.Id), zap.String("grpc_addr", client.GetAddress()))
					}
					delete(s.managerClients, resp.Node.Id)
				}

				if _, exist := s.managers.Nodes[resp.Node.Id]; exist {
					delete(s.managers.Nodes, resp.Node.Id)
				}
			default:
				s.logger.Debug("unknown event", zap.Any("event", resp.Event))
				continue
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
	for id, node := range s.cluster.Nodes {
		switch node.State {
		case index.Node_LEADER:
			if client, exist := s.peerClients[id]; exist {
				return client, nil
			}
		}
	}

	err := errors.New("there is no leader")
	s.logger.Error(err.Error())
	return nil, err
}

func (s *GRPCService) cloneCluster(cluster *index.Cluster) (*index.Cluster, error) {
	b, err := json.Marshal(cluster)
	if err != nil {
		return nil, err
	}

	var clone *index.Cluster
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
						continue
					}
					n2, err := json.Marshal(nodeSnapshot)
					if err != nil {
						s.logger.Warn(err.Error(), zap.String("id", id), zap.Any("node", nodeSnapshot))
						continue
					}
					if !cmp.Equal(n1, n2) {
						// node updated
						// notify the cluster changes
						clusterResp := &index.ClusterWatchResponse{
							Event:   index.ClusterWatchResponse_UPDATE,
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
					clusterResp := &index.ClusterWatchResponse{
						Event:   index.ClusterWatchResponse_JOIN,
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
					clusterResp := &index.ClusterWatchResponse{
						Event:   index.ClusterWatchResponse_LEAVE,
						Node:    node,
						Cluster: snapshotCluster,
					}
					for c := range s.clusterChans {
						c <- *clusterResp
					}
				}
			}

			// set cluster state to manager
			if !cmp.Equal(savedCluster, snapshotCluster) && s.managerGrpcAddress != "" && s.raftServer.IsLeader() {
				snapshotClusterBytes, err := json.Marshal(snapshotCluster)
				if err != nil {
					s.logger.Error(err.Error())
					continue
				}
				var snapshotClusterMap map[string]interface{}
				err = json.Unmarshal(snapshotClusterBytes, &snapshotClusterMap)
				if err != nil {
					s.logger.Error(err.Error())
					continue
				}

				client, err := s.getManagerClient()
				if err != nil {
					s.logger.Error(err.Error())
					continue
				}
				s.logger.Info("update shards", zap.Any("shards", snapshotClusterMap))
				err = client.Set(fmt.Sprintf("cluster/shards/%s", s.shardId), snapshotClusterMap)
				if err != nil {
					s.logger.Error(err.Error())
					continue
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

func (s *GRPCService) NodeHealthCheck(ctx context.Context, req *index.NodeHealthCheckRequest) (*index.NodeHealthCheckResponse, error) {
	resp := &index.NodeHealthCheckResponse{}

	switch req.Probe {
	case index.NodeHealthCheckRequest_HEALTHINESS:
		resp.State = index.NodeHealthCheckResponse_HEALTHY
	case index.NodeHealthCheckRequest_LIVENESS:
		resp.State = index.NodeHealthCheckResponse_ALIVE
	case index.NodeHealthCheckRequest_READINESS:
		resp.State = index.NodeHealthCheckResponse_READY
	}

	return resp, nil
}

func (s *GRPCService) NodeID() string {
	return s.raftServer.NodeID()
}

func (s *GRPCService) getSelfNode() *index.Node {
	node := s.raftServer.node

	switch s.raftServer.State() {
	case raft.Follower:
		node.State = index.Node_FOLLOWER
	case raft.Candidate:
		node.State = index.Node_CANDIDATE
	case raft.Leader:
		node.State = index.Node_LEADER
	case raft.Shutdown:
		node.State = index.Node_SHUTDOWN
	default:
		node.State = index.Node_UNKNOWN
	}

	return node
}

func (s *GRPCService) getPeerNode(id string) (*index.Node, error) {
	if _, exist := s.peerClients[id]; !exist {
		err := errors.New("node does not exist in peers")
		s.logger.Debug(err.Error(), zap.String("id", id))
		return nil, err
	}

	node, err := s.peerClients[id].NodeInfo()
	if err != nil {
		s.logger.Debug(err.Error(), zap.String("id", id))
		return &index.Node{
			BindAddress: "",
			State:       index.Node_UNKNOWN,
			Metadata: &index.Metadata{
				GrpcAddress: "",
				HttpAddress: "",
			},
		}, nil
	}

	return node, nil
}

func (s *GRPCService) getNode(id string) (*index.Node, error) {
	if id == "" || id == s.NodeID() {
		return s.getSelfNode(), nil
	} else {
		return s.getPeerNode(id)
	}
}

func (s *GRPCService) NodeInfo(ctx context.Context, req *empty.Empty) (*index.NodeInfoResponse, error) {
	resp := &index.NodeInfoResponse{}

	node, err := s.getNode(s.NodeID())
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return &index.NodeInfoResponse{
		Node: node,
	}, nil
}

func (s *GRPCService) setNode(node *index.Node) error {
	if s.raftServer.IsLeader() {
		err := s.raftServer.SetNode(node)
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
		err = client.ClusterJoin(node)
		if err != nil {
			s.logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func (s *GRPCService) ClusterJoin(ctx context.Context, req *index.ClusterJoinRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	err := s.setNode(req.Node)
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

func (s *GRPCService) ClusterLeave(ctx context.Context, req *index.ClusterLeaveRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	err := s.deleteNode(req.Id)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) getCluster() (*index.Cluster, error) {
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

func (s *GRPCService) ClusterInfo(ctx context.Context, req *empty.Empty) (*index.ClusterInfoResponse, error) {
	resp := &index.ClusterInfoResponse{}

	cluster, err := s.getCluster()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Cluster = cluster

	return resp, nil
}

func (s *GRPCService) ClusterWatch(req *empty.Empty, server index.Index_ClusterWatchServer) error {
	chans := make(chan index.ClusterWatchResponse)

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
		switch err {
		case blasterrors.ErrNotFound:
			s.logger.Debug(err.Error(), zap.String("id", req.Id))
			return resp, status.Error(codes.NotFound, err.Error())
		default:
			s.logger.Error(err.Error(), zap.String("id", req.Id))
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	docMap := map[string]interface{}{
		"id":     req.Id,
		"fields": fields,
	}

	docBytes, err := json.Marshal(docMap)
	if err != nil {
		s.logger.Error(err.Error(), zap.String("id", req.Id))
		return resp, status.Error(codes.Internal, err.Error())
	}

	doc := &index.Document{}
	err = index.UnmarshalDocument(docBytes, doc)
	if err != nil {
		s.logger.Error(err.Error(), zap.String("id", req.Id))
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Document = doc

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
	docs := make([]*index.Document, 0)

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

		docs = append(docs, req.Document)
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
	resp := &index.GetIndexConfigResponse{
		IndexConfig: &index.IndexConfig{},
	}

	indexConfig, err := s.raftServer.GetIndexConfig()
	if err != nil {
		s.logger.Error(err.Error())
		return resp, status.Error(codes.Internal, err.Error())
	}

	if indexMapping, ok := indexConfig["index_mapping"]; ok {
		indexMappingAny := &any.Any{}
		err = protobuf.UnmarshalAny(indexMapping.(*mapping.IndexMappingImpl), indexMappingAny)
		if err != nil {
			s.logger.Error(err.Error())
			return resp, status.Error(codes.Internal, err.Error())
		}
		resp.IndexConfig.IndexMapping = indexMappingAny
	}

	if indexType, ok := indexConfig["index_type"]; ok {
		resp.IndexConfig.IndexType = indexType.(string)
	}

	if indexStorageType, ok := indexConfig["index_storage_type"]; ok {
		resp.IndexConfig.IndexStorageType = indexStorageType.(string)
	}

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
