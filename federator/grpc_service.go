//  Copyright (c) 2019 Minoru Osuka
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

package federator

import (
	"context"
	"errors"
	"io"
	"log"
	"reflect"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	managerAddr string

	logger *log.Logger

	managers            map[string]interface{}
	managerClients      map[string]*manager.GRPCClient
	watchManagersStopCh chan struct{}
	watchManagersDoneCh chan struct{}

	indexers            map[string]interface{}
	indexerClients      map[string]map[string]*indexer.GRPCClient
	watchIndexersStopCh chan struct{}
	watchIndexersDoneCh chan struct{}
}

func NewGRPCService(managerAddr string, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		managerAddr: managerAddr,
		logger:      logger,

		managers:       make(map[string]interface{}, 0),
		managerClients: make(map[string]*manager.GRPCClient, 0),

		indexers:       make(map[string]interface{}, 0),
		indexerClients: make(map[string]map[string]*indexer.GRPCClient, 0),
	}, nil
}

func (s *GRPCService) Start() error {
	s.logger.Print("[INFO] start watching managers")
	go s.startWatchManagers(500 * time.Millisecond)

	s.logger.Print("[INFO] start watching indexers")
	go s.startWatchIndexers(500 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Print("[INFO] stop watching managers")
	s.stopWatchManagers()

	s.logger.Print("[INFO] stop watching indexers")
	s.stopWatchIndexers()

	return nil
}

func (s *GRPCService) getManagerClient() (*manager.GRPCClient, error) {
	var client *manager.GRPCClient

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
	client, err := manager.NewGRPCClient(s.managerAddr)
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
	for id, node := range s.managers {
		metadata := node.(map[string]interface{})["metadata"].(map[string]interface{})

		s.logger.Printf("[DEBUG] create client for %s", metadata["grpc_addr"].(string))

		client, err := manager.NewGRPCClient(metadata["grpc_addr"].(string))
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
			continue
		}
		s.managerClients[id] = client
	}

	for {
		select {
		case <-s.watchManagersStopCh:
			s.logger.Print("[DEBUG] receive request that stop watching managers")
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

func (s *GRPCService) startWatchIndexers(checkInterval time.Duration) {
	s.logger.Printf("[INFO] start watching a cluster")

	s.watchIndexersStopCh = make(chan struct{})
	s.watchIndexersDoneCh = make(chan struct{})

	defer func() {
		close(s.watchIndexersDoneCh)
	}()

	// wait for manager available
	s.logger.Print("[INFO] wait for manager clients are available")
	for {
		if len(s.managerClients) > 0 {
			s.logger.Print("[INFO] manager clients are available")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// get active client for manager
	client, err := s.getManagerClient()
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}

	// get initial indexers
	clusters, err := client.GetState("/cluster_config/clusters/")
	if err != nil {
		s.logger.Printf("[ERR] %v", err)
	}
	if clusters == nil {
		s.logger.Print("[ERR] nil")
	}
	s.indexers = *clusters.(*map[string]interface{})

	// create clients for indexer
	for clusterId, ins := range s.indexers {
		cluster := ins.(map[string]interface{})
		for nodeId, node := range cluster {
			metadata := node.(map[string]interface{})["metadata"].(map[string]interface{})

			s.logger.Printf("[DEBUG] create indexer client for %s at %s", metadata["grpc_addr"].(string), clusterId)

			client, err := indexer.NewGRPCClient(metadata["grpc_addr"].(string))
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				continue
			}

			s.indexerClients[clusterId] = make(map[string]*indexer.GRPCClient)
			s.indexerClients[clusterId][nodeId] = client
		}
	}

	for {
		select {
		case <-s.watchIndexersStopCh:
			s.logger.Print("[DEBUG] receive request that stop watching indexers")
			return
		default:
			// get active client for manager
			client, err = s.getManagerClient()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				continue
			}

			// create stream
			stream, err := client.Watch("/cluster_config/clusters/")
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
			log.Printf("[DEBUG] %v", resp)

			// get current indexer cluster
			cluster, err := client.GetState("/cluster_config/clusters/")
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				continue
			}
			if cluster == nil {
				s.logger.Print("[ERR] nil")
				continue
			}
			indexers := *cluster.(*map[string]interface{})

			// compare previous manager with current manager
			if !reflect.DeepEqual(s.indexers, indexers) {
				s.logger.Printf("[INFO] %v", indexers)

			}
		}
	}
}

func (s *GRPCService) stopWatchIndexers() {
	// close clients
	s.logger.Printf("[INFO] close indexer clients")
	for clusterId, cluster := range s.indexerClients {
		for _, client := range cluster {
			s.logger.Printf("[DEBUG] close indexer client for %s at %s", client.GetAddress(), clusterId)
			err := client.Close()
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
			}
		}
	}

	// stop watching managers
	if s.watchIndexersStopCh != nil {
		s.logger.Printf("[INFO] stop watching indexers")
		close(s.watchIndexersStopCh)
	}

	// wait for stop watching indexers has done
	s.logger.Printf("[INFO] wait for stop watching indexers has done")
	<-s.watchIndexersDoneCh
}

func (s *GRPCService) GetMetadata(ctx context.Context, req *protobuf.GetMetadataRequest) (*protobuf.GetMetadataResponse, error) {
	return &protobuf.GetMetadataResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) GetNodeState(ctx context.Context, req *protobuf.GetNodeStateRequest) (*protobuf.GetNodeStateResponse, error) {
	return &protobuf.GetNodeStateResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	return &protobuf.GetNodeResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) SetNode(ctx context.Context, req *protobuf.SetNodeRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*protobuf.GetClusterResponse, error) {
	return &protobuf.GetClusterResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) WatchCluster(req *empty.Empty, server protobuf.Blast_WatchClusterServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
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

	return resp, nil
}

func (s *GRPCService) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "search")

	s.logger.Printf("[INFO] search %v", req)

	resp := &protobuf.SearchResponse{}

	return resp, nil
}

func (s *GRPCService) IndexDocument(stream protobuf.Blast_IndexDocumentServer) error {
	//docs := make([]*federation.Document, 0)
	//
	//for {
	//	doc, err := stream.Recv()
	//	if err == io.EOF {
	//		break
	//	}
	//	if err != nil {
	//		return status.Error(codes.Internal, err.Error())
	//	}
	//
	//	docs = append(docs, doc)
	//}

	//// index
	//result, err := s.raftServer.Index(docs)
	//if err != nil {
	//	return status.Error(codes.Internal, err.Error())
	//}

	//return stream.SendAndClose(result)

	return stream.SendAndClose(nil)
}

func (s *GRPCService) DeleteDocument(stream protobuf.Blast_DeleteDocumentServer) error {
	//docs := make([]*federation.Document, 0)
	//
	//for {
	//	doc, err := stream.Recv()
	//	if err == io.EOF {
	//		break
	//	}
	//	if err != nil {
	//		return status.Error(codes.Internal, err.Error())
	//	}
	//
	//	docs = append(docs, doc)
	//}

	//// delete
	//result, err := s.raftServer.Delete(docs)
	//if err != nil {
	//	return status.Error(codes.Internal, err.Error())
	//}

	//return stream.SendAndClose(result)

	return stream.SendAndClose(nil)
}

func (s *GRPCService) GetIndexConfig(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexConfigResponse, error) {
	return &protobuf.GetIndexConfigResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *GRPCService) GetIndexStats(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexStatsResponse, error) {
	return &protobuf.GetIndexStatsResponse{}, status.Error(codes.Unavailable, "not implement")
}
