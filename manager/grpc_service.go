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
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	raftServer *RaftServer
	logger     *log.Logger

	peers            map[string]interface{}
	peerClients      map[string]*GRPCClient
	watchPeersStopCh chan struct{}
	watchPeersDoneCh chan struct{}

	cluster map[string]interface{}

	clusterChans map[chan protobuf.GetClusterResponse]struct{}
	clusterMutex sync.RWMutex

	stateChans map[chan protobuf.WatchStateResponse]struct{}
	stateMutex sync.RWMutex
}

func NewGRPCService(raftServer *RaftServer, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: raftServer,
		logger:     logger,

		cluster: make(map[string]interface{}, 0),

		peers:       make(map[string]interface{}, 0),
		peerClients: make(map[string]*GRPCClient, 0),

		clusterChans: make(map[chan protobuf.GetClusterResponse]struct{}),

		stateChans: make(map[chan protobuf.WatchStateResponse]struct{}),
	}, nil
}

func (s *GRPCService) Start() error {
	s.logger.Print("[INFO] start watching cluster")
	go s.startWatchPeers(500 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Print("[INFO] stop watching cluster")
	s.stopWatchPeers()

	return nil
}

func (s *GRPCService) startWatchPeers(checkInterval time.Duration) {
	s.watchPeersStopCh = make(chan struct{})
	s.watchPeersDoneCh = make(chan struct{})

	s.logger.Printf("[INFO] start watching a cluster")

	defer func() {
		close(s.watchPeersDoneCh)
	}()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.watchPeersStopCh:
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
							newClient, err := NewGRPCClient(grpcAddr)
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
						newClient, err := NewGRPCClient(grpcAddr)
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
			ctx, _ := NewGRPCClientCtx()
			resp, err := s.GetCluster(ctx, &empty.Empty{})
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
			ins, err := protobuf.MarshalAny(resp.Cluster)
			cluster := *ins.(*map[string]interface{})

			// notify cluster state
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

func (s *GRPCService) stopWatchPeers() {
	s.logger.Printf("[INFO] stop watching a cluster")
	close(s.watchPeersStopCh)
	s.logger.Printf("[INFO] wait for stop watching a cluster has done")
	<-s.watchPeersDoneCh
	s.logger.Printf("[INFO] cluster watching has been stopped")

	// close clients
	s.logger.Printf("[INFO] close clients")
	for _, client := range s.peerClients {
		s.logger.Printf("[DEBUG] close client for %s", client.GetAddress())
		err := client.Close()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
	}
}

func (s *GRPCService) getMetadata(id string) (map[string]interface{}, error) {
	metadata, err := s.raftServer.GetMetadata(id)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (s *GRPCService) GetMetadata(ctx context.Context, req *protobuf.GetMetadataRequest) (*protobuf.GetMetadataResponse, error) {
	resp := &protobuf.GetMetadataResponse{}

	var metadata map[string]interface{}
	var err error

	if req.Id == s.raftServer.id {
		metadata, err = s.getMetadata(s.raftServer.id)
		if err != nil {
			s.logger.Printf("[ERR] %v: %v", err, s.raftServer.id)
		}
	} else {
		if client, exist := s.peerClients[req.Id]; exist {
			metadata, err = client.GetMetadata(req.Id)
			if err != nil {
				s.logger.Printf("[ERR] %v: %v", err, req.Id)
			}
		}
	}

	metadataAny := &any.Any{}
	err = protobuf.UnmarshalAny(metadata, metadataAny)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Metadata = metadataAny

	return resp, nil
}

func (s *GRPCService) getNodeState() string {
	return s.raftServer.State()
}

func (s *GRPCService) GetNodeState(ctx context.Context, req *protobuf.GetNodeStateRequest) (*protobuf.GetNodeStateResponse, error) {
	resp := &protobuf.GetNodeStateResponse{}

	state := ""
	var err error

	if req.Id == s.raftServer.id {
		state = s.getNodeState()
	} else {
		if client, exist := s.peerClients[req.Id]; exist {
			state, err = client.GetNodeState(req.Id)
			if err != nil {
				return resp, status.Error(codes.Internal, err.Error())
			}
		}
	}

	resp.State = state

	return resp, nil
}

func (s *GRPCService) getNode(id string) (map[string]interface{}, error) {
	metadata, err := s.getMetadata(id)
	if err != nil {
		return nil, err
	}

	state := s.getNodeState()

	node := map[string]interface{}{
		"metadata": metadata,
		"state":    state,
	}

	return node, nil
}

func (s *GRPCService) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	resp := &protobuf.GetNodeResponse{}

	var metadata map[string]interface{}
	var state string
	var err error

	if req.Id == s.raftServer.id {
		node, err := s.getNode(req.Id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}

		metadata = node["metadata"].(map[string]interface{})

		state = node["state"].(string)
	} else {
		if client, exist := s.peerClients[req.Id]; exist {
			metadata, err = client.GetMetadata(req.Id)
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}

			state, err = client.GetNodeState(req.Id)
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				state = raft.Shutdown.String()
			}
		}
	}

	metadataAny := &any.Any{}
	err = protobuf.UnmarshalAny(metadata, metadataAny)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Metadata = metadataAny
	resp.State = state

	return resp, nil
}

func (s *GRPCService) SetNode(ctx context.Context, req *protobuf.SetNodeRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] %v", req)

	resp := &empty.Empty{}

	ins, err := protobuf.MarshalAny(req.Metadata)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	metadata := *ins.(*map[string]interface{})

	err = s.raftServer.SetMetadata(req.Id, metadata)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] leave %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.DeleteMetadata(req.Id)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
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
		node := map[string]interface{}{}
		if id == s.raftServer.id {
			node, err = s.getNode(id)
			if err != nil {
				return resp, err
			}
		} else {
			r, err := s.GetNode(ctx, &protobuf.GetNodeRequest{Id: id})
			if err != nil {
				return resp, err
			}

			ins, err := protobuf.MarshalAny(r.Metadata)
			metadata := *ins.(*map[string]interface{})

			node = map[string]interface{}{
				"metadata": metadata,
				"state":    r.State,
			}
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
	start := time.Now()
	s.stateMutex.Lock()
	defer func() {
		s.stateMutex.Unlock()
		RecordMetrics(start, "snapshot")
	}()

	resp := &empty.Empty{}

	err := s.raftServer.Snapshot()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) LivenessProbe(ctx context.Context, req *empty.Empty) (*protobuf.LivenessProbeResponse, error) {
	resp := &protobuf.LivenessProbeResponse{
		State: protobuf.LivenessProbeResponse_ALIVE,
	}

	return resp, nil
}

func (s *GRPCService) ReadinessProbe(ctx context.Context, req *empty.Empty) (*protobuf.ReadinessProbeResponse, error) {
	resp := &protobuf.ReadinessProbeResponse{
		State: protobuf.ReadinessProbeResponse_READY,
	}

	return resp, nil
}

func (s *GRPCService) GetState(ctx context.Context, req *protobuf.GetStateRequest) (*protobuf.GetStateResponse, error) {
	start := time.Now()
	s.stateMutex.RLock()
	defer func() {
		s.stateMutex.RUnlock()
		RecordMetrics(start, "get")
	}()

	resp := &protobuf.GetStateResponse{}

	value, err := s.raftServer.Get(req.Key)
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
	start := time.Now()
	s.stateMutex.Lock()
	defer func() {
		s.stateMutex.Unlock()
		RecordMetrics(start, "set")
	}()

	resp := &empty.Empty{}

	value, err := protobuf.MarshalAny(req.Value)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	err = s.raftServer.Set(req.Key, value)
	if err != nil {
		switch err {
		case blasterrors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
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
	start := time.Now()
	s.stateMutex.Lock()
	defer func() {
		s.stateMutex.Unlock()
		RecordMetrics(start, "delete")
	}()

	s.logger.Printf("[INFO] set %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.Delete(req.Key)
	if err != nil {
		switch err {
		case blasterrors.ErrNotFound:
			return resp, status.Error(codes.NotFound, err.Error())
		default:
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

func (s *GRPCService) GetDocument(ctx context.Context, req *protobuf.GetDocumentRequest) (*protobuf.GetDocumentResponse, error) {
	return &protobuf.GetDocumentResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	return &protobuf.SearchResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) IndexDocument(stream protobuf.Blast_IndexDocumentServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) DeleteDocument(stream protobuf.Blast_DeleteDocumentServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) GetIndexConfig(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexConfigResponse, error) {
	return &protobuf.GetIndexConfigResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) GetIndexStats(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexStatsResponse, error) {
	return &protobuf.GetIndexStatsResponse{}, status.Error(codes.Unavailable, "not implement")
}
