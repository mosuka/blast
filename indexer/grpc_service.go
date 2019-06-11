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
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
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
	managerAddr string
	clusterId   string

	raftServer *RaftServer
	logger     *log.Logger

	watchClusterStopCh chan struct{}
	watchClusterDoneCh chan struct{}
	peers              map[string]interface{}
	peerClients        map[string]*GRPCClient
	cluster            map[string]interface{}
	clusterChans       map[chan protobuf.GetClusterResponse]struct{}
	clusterMutex       sync.RWMutex

	managers            map[string]interface{}
	managerClients      map[string]*grpc.Client
	watchManagersStopCh chan struct{}
	watchManagersDoneCh chan struct{}
}

func NewGRPCService(managerAddr string, clusterId string, raftServer *RaftServer, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		managerAddr: managerAddr,
		clusterId:   clusterId,

		raftServer: raftServer,
		logger:     logger,

		peers:        make(map[string]interface{}, 0),
		peerClients:  make(map[string]*GRPCClient, 0),
		cluster:      make(map[string]interface{}, 0),
		clusterChans: make(map[chan protobuf.GetClusterResponse]struct{}),

		managers:       make(map[string]interface{}, 0),
		managerClients: make(map[string]*grpc.Client, 0),
	}, nil
}

func (s *GRPCService) Start() error {
	s.logger.Print("[INFO] start watching cluster")
	go s.startWatchCluster(500 * time.Millisecond)

	if s.managerAddr != "" {
		s.logger.Print("[INFO] start watching managers")
		go s.startWatchManagers(500 * time.Millisecond)
	}

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Print("[INFO] stop watching cluster")
	s.stopWatchCluster()

	if s.managerAddr != "" {
		s.logger.Print("[INFO] stop watching managers")
		s.stopWatchManagers()
	}

	return nil
}

func (s *GRPCService) getManagerClient() (*grpc.Client, error) {
	var client *grpc.Client

	for id, node := range s.managers {
		state := node.(map[string]interface{})["state"].(string)
		if state != raft.Shutdown.String() {

			if _, exist := s.managerClients[id]; exist {
				client = s.managerClients[id]
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

func (s *GRPCService) getInitialManagers(managerAddr string) (map[string]interface{}, error) {
	client, err := grpc.NewClient(s.managerAddr)
	defer func() {
		err := client.Close()
		s.logger.Printf("[ERR] %v", err)
		return
	}()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil, err
	}

	managers, err := client.GetCluster()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return nil, err
	}

	return managers, nil
}

func (s *GRPCService) startWatchManagers(checkInterval time.Duration) {
	s.logger.Printf("[INFO] start watching a cluster")

	s.watchManagersStopCh = make(chan struct{})
	s.watchManagersDoneCh = make(chan struct{})

	defer func() {
		close(s.watchManagersDoneCh)
	}()

	var err error

	// get initial managers
	s.managers, err = s.getInitialManagers(s.managerAddr)
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
		return
	}
	s.logger.Printf("[DEBUG] %v", s.managers)

	// create clients for managers
	for nodeId, node := range s.managers {
		metadata := node.(map[string]interface{})["metadata"].(map[string]interface{})

		s.logger.Printf("[DEBUG] create client for %s", metadata["grpc_addr"].(string))

		client, err := grpc.NewClient(metadata["grpc_addr"].(string))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			continue
		}
		s.managerClients[nodeId] = client
	}

	for {
		select {
		case <-s.watchManagersStopCh:
			s.logger.Print("[DEBUG] receive request that stop watching a cluster")
			return
		default:
			// get active client for manager
			client, err := s.getManagerClient()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				continue
			}

			// create stream
			stream, err := client.WatchCluster()
			if err != nil {
				st, _ := status.FromError(err)
				switch st.Code() {
				case codes.Canceled:
					s.logger.Printf("[DEBUG] %v", err)
				default:
					s.logger.Printf("[ERR] %v", err)
				}
				continue
			}

			// wait for receive cluster updates from stream
			s.logger.Print("[DEBUG] wait for receive cluster updates from stream")
			resp, err := stream.Recv()
			if err == io.EOF {
				continue
			}
			if err != nil {
				st, _ := status.FromError(err)
				switch st.Code() {
				case codes.Canceled:
					s.logger.Printf("[DEBUG] %v", err)
				default:
					s.logger.Printf("[ERR] %v", err)
				}
				continue
			}

			// get current manager cluster
			cluster, err := protobuf.MarshalAny(resp.Cluster)
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				continue
			}
			if cluster == nil {
				s.logger.Print("[ERR] nil")
				continue
			}
			managers := *cluster.(*map[string]interface{})

			// compare previous manager with current manager
			if !reflect.DeepEqual(s.managers, managers) {
				s.logger.Printf("[INFO] %v", managers)

				// close the client for left manager node
				for id := range s.managers {
					if _, managerExists := managers[id]; !managerExists {
						if _, clientExists := s.managerClients[id]; clientExists {
							client := s.managerClients[id]

							s.logger.Printf("[DEBUG] close client for %s", client.GetAddress())
							err = client.Close()
							if err != nil {
								s.logger.Printf("[ERR] %v", err)
							}

							delete(s.managerClients, id)
						}
					}
				}

				// keep current manager cluster
				s.managers = managers
			}
		}
	}
}

func (s *GRPCService) stopWatchManagers() {
	// close clients
	s.logger.Printf("[INFO] close manager clients")
	for _, client := range s.managerClients {
		s.logger.Printf("[DEBUG] close manager client for %s", client.GetAddress())
		err := client.Close()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
	}

	// stop watching managers
	if s.watchManagersStopCh != nil {
		s.logger.Printf("[INFO] stop watching managers")
		close(s.watchManagersStopCh)
	}

	// wait for stop watching managers has done
	s.logger.Printf("[INFO] wait for stop watching managers has done")
	<-s.watchManagersDoneCh
}

func (s *GRPCService) startWatchCluster(checkInterval time.Duration) {
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

			// create and close clients for manager
			if !reflect.DeepEqual(s.peers, peers) {
				// create clients
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

			if !reflect.DeepEqual(s.cluster, cluster) {
				// notify cluster state
				for c := range s.clusterChans {
					c <- *resp
				}

				// update cluster to manager
				if s.raftServer.IsLeader() && s.managerAddr != "" {
					// get active client for manager
					client, err := s.getManagerClient()
					if err != nil {
						s.logger.Printf("[ERR] %v", err)
						continue
					}
					// TODO: SetState -> SetClusterNode
					err = client.SetState(fmt.Sprintf("/cluster_config/clusters/%s", s.clusterId), cluster)
					if err != nil {
						continue
					}
				}

				// keep current cluster
				s.cluster = cluster
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *GRPCService) stopWatchCluster() {
	// close clients
	s.logger.Printf("[INFO] close peer clients")
	for _, client := range s.peerClients {
		s.logger.Printf("[DEBUG] close peer client for %s", client.GetAddress())
		err := client.Close()
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
	}

	// stop watching peers
	if s.watchClusterStopCh != nil {
		s.logger.Printf("[INFO] stop watching peers")
		close(s.watchClusterStopCh)
	}

	// wait for stop watching peers has done
	s.logger.Printf("[INFO] wait for stop watching peers has done")
	<-s.watchClusterDoneCh
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
		metadata, err = s.getMetadata(req.Id)
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}

		state = s.getNodeState()
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
