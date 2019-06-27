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
	"log"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	*grpc.Service

	raftServer *RaftServer
	logger     *log.Logger

	watchClusterStopCh chan struct{}
	watchClusterDoneCh chan struct{}
	peers              map[string]interface{}
	peerClients        map[string]*grpc.Client
	cluster            map[string]interface{}
	clusterChans       map[chan protobuf.GetClusterResponse]struct{}
	clusterMutex       sync.RWMutex

	stateChans map[chan protobuf.WatchStateResponse]struct{}
	stateMutex sync.RWMutex
}

func NewGRPCService(raftServer *RaftServer, logger *log.Logger) (*GRPCService, error) {
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
	s.logger.Print("[INFO] start update cluster")
	go s.startUpdateCluster(500 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Print("[INFO] stop update cluster")
	s.stopUpdateCluster()

	return nil
}

func (s *GRPCService) getLeaderClient() (*grpc.Client, error) {
	var client *grpc.Client

	for id, node := range s.cluster {
		state := node.(map[string]interface{})["state"].(string)
		if state != raft.Shutdown.String() {

			if _, exist := s.peerClients[id]; exist {
				client = s.peerClients[id]
				break
			} else {
				s.logger.Printf("[DEBUG] %v does not exist", id)
			}
		}
	}

	if client == nil {
		return nil, errors.New("client does not exist")
	}

	return client, nil
}

func (s *GRPCService) startUpdateCluster(checkInterval time.Duration) {
	s.watchClusterStopCh = make(chan struct{})
	s.watchClusterDoneCh = make(chan struct{})

	s.logger.Printf("[INFO] start watching a cluster")

	defer func() {
		close(s.watchClusterDoneCh)
	}()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.watchClusterStopCh:
			s.logger.Print("[DEBUG] receive request that stop watching a cluster")
			return
		case <-ticker.C:
			// get servers
			servers, err := s.raftServer.GetServers()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				return
			}

			// create peer node list with out self node
			peers := make(map[string]interface{}, 0)
			for id, metadata := range servers {
				if id != s.raftServer.id {
					peers[id] = metadata
				}
			}

			if !reflect.DeepEqual(s.peers, peers) {
				// open clients
				for id, metadata := range peers {
					grpcAddr := metadata.(map[string]interface{})["grpc_addr"].(string)

					if _, clientExists := s.peerClients[id]; clientExists {
						client := s.peerClients[id]
						if client.GetAddress() != grpcAddr {
							s.logger.Printf("[DEBUG] close client for %s", client.GetAddress())
							err = client.Close()
							if err != nil {
								s.logger.Printf("[ERR] %v", err)
							}

							s.logger.Printf("[DEBUG] create client for %s", grpcAddr)
							newClient, err := grpc.NewClient(grpcAddr)
							if err != nil {
								s.logger.Printf("[ERR] %v", err)
							}

							if client != nil {
								s.logger.Printf("[DEBUG] create client for %s", newClient.GetAddress())
								s.peerClients[id] = newClient
							}
						}
					} else {
						s.logger.Printf("[DEBUG] create client for %s", grpcAddr)
						newClient, err := grpc.NewClient(grpcAddr)
						if err != nil {
							s.logger.Printf("[ERR] %v", err)
						}

						s.peerClients[id] = newClient
					}
				}

				// close nonexistent clients
				for id := range s.peers {
					if _, peerExists := peers[id]; !peerExists {
						if _, clientExists := s.peerClients[id]; clientExists {
							client := s.peerClients[id]

							s.logger.Printf("[DEBUG] close client for %s", client.GetAddress())
							err = client.Close()
							if err != nil {
								s.logger.Printf("[ERR] %v", err)
							}

							delete(s.peerClients, id)
						}
					}
				}

				// keep current peer nodes
				s.peers = peers
			}

			// get cluster
			ctx, _ := grpc.NewContext()
			resp, err := s.GetCluster(ctx, &empty.Empty{})
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
			clusterIntr, err := protobuf.MarshalAny(resp.Cluster)
			cluster := *clusterIntr.(*map[string]interface{})

			// notify current cluster
			if !reflect.DeepEqual(s.cluster, cluster) {
				for c := range s.clusterChans {
					c <- *resp
				}

				// keep current cluster
				s.cluster = cluster
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *GRPCService) stopUpdateCluster() {
	// close clients
	s.logger.Printf("[INFO] close peer clients")
	for _, client := range s.peerClients {
		s.logger.Printf("[DEBUG] close peer client for %s", client.GetAddress())
		err := client.Close()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
	}

	// stop watching a cluster
	if s.watchClusterStopCh != nil {
		s.logger.Printf("[INFO] stop watching a cluster")
		close(s.watchClusterStopCh)
	}

	// wait for stop watching a cluster has done
	s.logger.Printf("[INFO] wait for stop watching a cluster has done")
	<-s.watchClusterDoneCh
}

func (s *GRPCService) getSelfNode() (map[string]interface{}, error) {
	metadata, err := s.raftServer.GetMetadata(s.raftServer.id)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	node := map[string]interface{}{
		"metadata": metadata,
		"state":    s.raftServer.State(),
	}

	return node, nil
}

func (s *GRPCService) getPeerNode(id string) (map[string]interface{}, error) {
	var node map[string]interface{}
	var err error

	if client, exist := s.peerClients[id]; exist {
		node, err = client.GetNode(id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			node = map[string]interface{}{
				"metadata": map[string]interface{}{},
				"state":    raft.Shutdown.String(),
			}
		}
	} else {
		s.logger.Printf("[ERR] %v does not exist", id)
		node = map[string]interface{}{
			"metadata": map[string]interface{}{},
			"state":    "Gone",
		}
	}

	return node, nil
}

func (s *GRPCService) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	resp := &protobuf.GetNodeResponse{}

	var node map[string]interface{}
	var err error
	if req.Id == "" || req.Id == s.raftServer.id {
		node, err = s.getSelfNode()
	} else {
		node, err = s.getPeerNode(req.Id)
	}
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	metadataAny := &any.Any{}
	state := "Gone"
	if node != nil {
		if _, exist := node["metadata"]; exist {
			if node["metadata"] != nil {
				err = protobuf.UnmarshalAny(node["metadata"].(map[string]interface{}), metadataAny)
				if err != nil {
					return resp, status.Error(codes.Internal, err.Error())
				}
			}
		}

		if _, exist := node["state"]; exist {
			if node["state"] != nil {
				state = node["state"].(string)
			}
		}
	}

	resp.Metadata = metadataAny
	resp.State = state

	return resp, nil
}

func (s *GRPCService) SetNode(ctx context.Context, req *protobuf.SetNodeRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	ins, err := protobuf.MarshalAny(req.Metadata)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	metadata := *ins.(*map[string]interface{})

	if s.raftServer.IsLeader() {
		err = s.raftServer.SetMetadata(req.Id, metadata)
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.SetNode(req.Id, metadata)
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (s *GRPCService) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	resp := &empty.Empty{}

	if s.raftServer.IsLeader() {
		err := s.raftServer.DeleteMetadata(req.Id)
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
	} else {
		// forward to leader
		client, err := s.getLeaderClient()
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.DeleteNode(req.Id)
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*protobuf.GetClusterResponse, error) {
	resp := &protobuf.GetClusterResponse{}

	servers, err := s.raftServer.GetServers()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	cluster := map[string]interface{}{}
	for id := range servers {
		nodeResp, err := s.GetNode(ctx, &protobuf.GetNodeRequest{Id: id})
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}

		metadataIntr, err := protobuf.MarshalAny(nodeResp.Metadata)
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
		metadata := *metadataIntr.(*map[string]interface{})

		node := map[string]interface{}{
			"metadata": metadata,
			"state":    nodeResp.State,
		}

		cluster[id] = node
	}

	clusterAny := &any.Any{}
	err = protobuf.UnmarshalAny(cluster, clusterAny)
	if err != nil {
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
		return resp, status.Error(codes.Internal, err.Error())
	}

	if s.raftServer.IsLeader() {
		err = s.raftServer.SetState(req.Key, value)
		if err != nil {
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
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.SetState(req.Key, value)
		if err != nil {
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

	s.logger.Printf("[INFO] set %v", req)

	resp := &empty.Empty{}

	if s.raftServer.IsLeader() {
		err := s.raftServer.DeleteState(req.Key)
		if err != nil {
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
			return resp, status.Error(codes.Internal, err.Error())
		}
		err = client.DeleteState(req.Key)
		if err != nil {
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
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}
