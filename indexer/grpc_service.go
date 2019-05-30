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
	"io"
	"log"
	"reflect"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	managerAddr string
	clusterId   string

	raftServer *RaftServer
	logger     *log.Logger

	watchPeersStopCh chan struct{}
	watchPeersDoneCh chan struct{}

	peers       map[string]interface{}
	peerClients map[string]*GRPCClient

	managers       map[string]interface{}
	managerClients map[string]*GRPCClient
}

func NewGRPCService(managerAddr string, clusterId string, raftServer *RaftServer, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		managerAddr: managerAddr,
		clusterId:   clusterId,

		raftServer: raftServer,
		logger:     logger,

		peers:       make(map[string]interface{}, 0),
		peerClients: make(map[string]*GRPCClient, 0),

		managers:       make(map[string]interface{}, 0),
		managerClients: make(map[string]*GRPCClient, 0),
	}, nil
}

func (s *GRPCService) Start() error {
	go s.startWatchPeers(500 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.stopWatchPeers()

	return nil
}

func (s *GRPCService) startWatchManagers(checkInterval time.Duration) {

}

func (s *RaftServer) updateCluster() error {
	//if s.managerAddr != "" {
	//	mc, err := manager.NewGRPCClient(s.managerAddr)
	//	defer func() {
	//		err = mc.Close()
	//		if err != nil {
	//			s.logger.Printf("[ERR] %v", err)
	//			return
	//		}
	//	}()
	//	if err != nil {
	//		return err
	//	}
	//
	//	cluster, err := s.GetCluster()
	//	if err != nil {
	//		return err
	//	}
	//
	//	err = mc.Set(fmt.Sprintf("/cluster_config/clusters/%s", s.clusterId), cluster)
	//	if err != nil {
	//		return err
	//	}
	//}

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
			// get the cluster
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

			if reflect.DeepEqual(s.peers, peers) {
				// there is no change in the cluster
				continue
			}

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

			// update master cluster state
			// TODO
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *GRPCService) stopWatchPeers() {
	s.logger.Printf("[INFO] stop watching peers")
	close(s.watchPeersStopCh)
	s.logger.Printf("[INFO] wait for stop watching peers has done")
	<-s.watchPeersDoneCh
	s.logger.Printf("[INFO] peer watching has been stopped")

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
		metadata, err = s.getMetadata(req.Id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
	} else {
		metadata, err = s.peerClients[req.Id].GetMetadata(req.Id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
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
		state, err = s.peerClients[req.Id].GetNodeState(req.Id)
		if err != nil {
			return resp, status.Error(codes.Internal, err.Error())
		}
	}

	resp.State = state

	return resp, nil
}

func (s *GRPCService) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	resp := &protobuf.GetNodeResponse{}

	var metadata map[string]interface{}
	var state string
	var err error

	if req.Id == s.raftServer.id {
		metadata, err = s.getMetadata(req.Id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}

		state = s.getNodeState()
	} else {
		metadata, err = s.peerClients[req.Id].GetMetadata(req.Id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}

		state, err = s.peerClients[req.Id].GetNodeState(req.Id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			state = raft.Shutdown.String()
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
		r, err := s.GetNode(ctx, &protobuf.GetNodeRequest{Id: id})
		if err != nil {
			return resp, err
		}

		ins, err := protobuf.MarshalAny(r.Metadata)
		metadata := *ins.(*map[string]interface{})

		node := map[string]interface{}{
			"metadata": metadata,
			"state":    r.State,
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

func (s *GRPCService) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	s.logger.Printf("[INFO] %v", req)

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
	return &protobuf.GetStateResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) SetState(ctx context.Context, req *protobuf.SetStateRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) DeleteState(ctx context.Context, req *protobuf.DeleteStateRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) WatchState(req *protobuf.WatchStateRequest, server protobuf.Blast_WatchStateServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) GetDocument(ctx context.Context, req *protobuf.GetDocumentRequest) (*protobuf.GetDocumentResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "get")

	s.logger.Printf("[INFO] get %v", req)

	resp := &protobuf.GetDocumentResponse{}

	fields, err := s.raftServer.GetDocument(req.Id)
	if err != nil {
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
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Fields = fieldsAny

	return resp, nil
}

func (s *GRPCService) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "search")

	s.logger.Printf("[INFO] search %v", req)

	resp := &protobuf.SearchResponse{}

	// Any -> bleve.SearchRequest
	searchRequest, err := protobuf.MarshalAny(req.SearchRequest)
	if err != nil {
		return resp, status.Error(codes.InvalidArgument, err.Error())
	}

	searchResult, err := s.raftServer.Search(searchRequest.(*bleve.SearchRequest))
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	// bleve.SearchResult -> Any
	searchResultAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchResult, searchResultAny)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.SearchResult = searchResultAny

	return resp, nil
}

func (s *GRPCService) IndexDocument(stream protobuf.Blast_IndexDocumentServer) error {
	docs := make([]map[string]interface{}, 0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		fields, err := protobuf.MarshalAny(req.Fields)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		doc := map[string]interface{}{
			"id":     req.Id,
			"fields": fields,
		}

		docs = append(docs, doc)
	}

	// index
	count, err := s.raftServer.IndexDocument(docs)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	resp := &protobuf.IndexDocumentResponse{
		Count: int32(count),
	}

	return stream.SendAndClose(resp)
}

func (s *GRPCService) DeleteDocument(stream protobuf.Blast_DeleteDocumentServer) error {
	ids := make([]string, 0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		ids = append(ids, req.Id)
	}

	// delete
	count, err := s.raftServer.DeleteDocument(ids)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	resp := &protobuf.DeleteDocumentResponse{
		Count: int32(count),
	}

	return stream.SendAndClose(resp)
}

func (s *GRPCService) GetIndexConfig(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexConfigResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "indexconfig")

	resp := &protobuf.GetIndexConfigResponse{}

	s.logger.Printf("[INFO] stats %v", req)

	var err error

	indexConfig, err := s.raftServer.GetIndexConfig()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	indexConfigAny := &any.Any{}
	err = protobuf.UnmarshalAny(indexConfig, indexConfigAny)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.IndexConfig = indexConfigAny

	return resp, nil
}

func (s *GRPCService) GetIndexStats(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexStatsResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "indexstats")

	resp := &protobuf.GetIndexStatsResponse{}

	s.logger.Printf("[INFO] stats %v", req)

	var err error

	indexStats, err := s.raftServer.GetIndexStats()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	indexStatsAny := &any.Any{}
	err = protobuf.UnmarshalAny(indexStats, indexStatsAny)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.IndexStats = indexStatsAny

	return resp, nil
}
