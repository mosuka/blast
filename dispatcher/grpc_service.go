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

package dispatcher

import (
	"context"
	"errors"
	"hash/fnv"
	"io"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/distribute"
	"github.com/mosuka/blast/sortutils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	managerAddr string
	logger      *zap.Logger

	managers             map[string]interface{}
	managerClients       map[string]*manager.GRPCClient
	updateManagersStopCh chan struct{}
	updateManagersDoneCh chan struct{}

	indexers             map[string]interface{}
	indexerClients       map[string]map[string]*indexer.GRPCClient
	updateIndexersStopCh chan struct{}
	updateIndexersDoneCh chan struct{}
}

func NewGRPCService(managerAddr string, logger *zap.Logger) (*GRPCService, error) {
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
	s.logger.Info("start to update manager cluster info")
	go s.startUpdateManagers(500 * time.Millisecond)

	s.logger.Info("start to update indexer cluster info")
	go s.startUpdateIndexers(500 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Info("stop to update manager cluster info")
	s.stopUpdateManagers()

	s.logger.Info("stop to update indexer cluster info")
	s.stopUpdateIndexers()

	return nil
}

func (s *GRPCService) getManagerClient() (*manager.GRPCClient, error) {
	var client *manager.GRPCClient

	for id, node := range s.managers {
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

		if state == raft.Leader.String() || state == raft.Follower.String() {
			client, ok = s.managerClients[id]
			if ok {
				return client, nil
			} else {
				s.logger.Error("node does not exist", zap.String("id", id))
			}
		} else {
			s.logger.Debug("node has not available", zap.String("id", id), zap.String("state", state))
		}
	}

	err := errors.New("available client does not exist")
	s.logger.Error(err.Error())

	return nil, err
}

func (s *GRPCService) getInitialManagers(managerAddr string) (map[string]interface{}, error) {
	client, err := manager.NewGRPCClient(s.managerAddr)
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

	managers, err := client.GetCluster()
	if err != nil {
		s.logger.Error(err.Error())
		return nil, err
	}

	return managers, nil
}

func (s *GRPCService) startUpdateManagers(checkInterval time.Duration) {
	s.updateManagersStopCh = make(chan struct{})
	s.updateManagersDoneCh = make(chan struct{})

	defer func() {
		close(s.updateManagersDoneCh)
	}()

	var err error

	// get initial managers
	s.managers, err = s.getInitialManagers(s.managerAddr)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.logger.Debug("initialize manager list", zap.Any("managers", s.managers))

	// create clients for managers
	for nodeId, node := range s.managers {
		nm, ok := node.(map[string]interface{})
		if !ok {
			s.logger.Warn("assertion failed", zap.String("node_id", nodeId))
			continue
		}

		nodeConfig, ok := nm["node_config"].(map[string]interface{})
		if !ok {
			s.logger.Warn("missing metadata", zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))
			continue
		}

		grpcAddr, ok := nodeConfig["grpc_addr"].(string)
		if !ok {
			s.logger.Warn("missing gRPC address", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
			continue
		}

		s.logger.Debug("create gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
		client, err := manager.NewGRPCClient(grpcAddr)
		if err != nil {
			s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
		}
		if client != nil {
			s.managerClients[nodeId] = client
		}
	}

	for {
		select {
		case <-s.updateManagersStopCh:
			s.logger.Info("received a request to stop updating a manager cluster")
			return
		default:
			client, err := s.getManagerClient()
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}

			// create stream
			stream, err := client.WatchCluster()
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

			// get current manager cluster
			managersIntr, err := protobuf.MarshalAny(resp.Cluster)
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}
			if managersIntr == nil {
				s.logger.Error(err.Error())
				continue
			}
			managers := *managersIntr.(*map[string]interface{})

			if !reflect.DeepEqual(s.managers, managers) {
				// open clients
				for nodeId, metadata := range managers {
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

					client, exist := s.managerClients[nodeId]
					if exist {
						s.logger.Debug("client has already exist in manager list", zap.String("node_id", nodeId))

						if client.GetAddress() != grpcAddr {
							s.logger.Debug("gRPC address has been changed", zap.String("node_id", nodeId), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", grpcAddr))
							s.logger.Debug("recreate gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))

							delete(s.managerClients, nodeId)

							err = client.Close()
							if err != nil {
								s.logger.Error(err.Error(), zap.String("node_id", nodeId))
							}

							newClient, err := manager.NewGRPCClient(grpcAddr)
							if err != nil {
								s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
							}

							if newClient != nil {
								s.managerClients[nodeId] = newClient
							}
						} else {
							s.logger.Debug("gRPC address has not changed", zap.String("node_id", nodeId), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", grpcAddr))
						}
					} else {
						s.logger.Debug("client does not exist in peer list", zap.String("node_id", nodeId))

						s.logger.Debug("create gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
						newClient, err := manager.NewGRPCClient(grpcAddr)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
						}
						if newClient != nil {
							s.managerClients[nodeId] = newClient
						}
					}
				}

				// close nonexistent clients
				for nodeId, client := range s.managerClients {
					if nodeConfig, exist := managers[nodeId]; !exist {
						s.logger.Info("this client is no longer in use", zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))

						s.logger.Debug("close client", zap.String("node_id", nodeId), zap.String("grpc_addr", client.GetAddress()))
						err = client.Close()
						if err != nil {
							s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", client.GetAddress()))
						}

						s.logger.Debug("delete client", zap.String("node_id", nodeId))
						delete(s.managerClients, nodeId)
					}
				}

				// keep current manager cluster
				s.managers = managers
				s.logger.Debug("managers", zap.Any("managers", s.managers))
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

func (s *GRPCService) startUpdateIndexers(checkInterval time.Duration) {
	s.updateIndexersStopCh = make(chan struct{})
	s.updateIndexersDoneCh = make(chan struct{})

	defer func() {
		close(s.updateIndexersDoneCh)
	}()

	// wait for manager available
	s.logger.Info("wait for manager clients are available")
	for {
		if len(s.managerClients) > 0 {
			s.logger.Info("manager clients are available")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// get active client for manager
	client, err := s.getManagerClient()
	if err != nil {
		s.logger.Error(err.Error())
	}

	// get initial indexers
	clusters, err := client.GetValue("/cluster_config/clusters/")
	if err != nil {
		s.logger.Error(err.Error())
	}
	if clusters == nil {
		s.logger.Error("nil")
	}
	s.indexers = *clusters.(*map[string]interface{})

	// create clients for indexer
	for clusterId, cluster := range s.indexers {
		cm, ok := cluster.(map[string]interface{})
		if !ok {
			s.logger.Warn("assertion failed", zap.String("cluster_id", clusterId), zap.Any("cluster", cm))
			continue
		}

		nodes, ok := cm["nodes"].(map[string]interface{})
		if !ok {
			s.logger.Warn("assertion failed", zap.String("cluster_id", clusterId), zap.Any("nodes", nodes))
			continue
		}

		for nodeId, node := range nodes {
			nm, ok := node.(map[string]interface{})
			if !ok {
				s.logger.Warn("assertion failed", zap.String("node_id", nodeId))
				continue
			}

			metadata, ok := nm["node_config"].(map[string]interface{})
			if !ok {
				s.logger.Warn("missing metadata", zap.String("node_id", nodeId), zap.Any("node_config", metadata))
				continue
			}

			grpcAddr, ok := metadata["grpc_addr"].(string)
			if !ok {
				s.logger.Warn("missing gRPC address", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
				continue
			}

			s.logger.Debug("create gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
			client, err := indexer.NewGRPCClient(metadata["grpc_addr"].(string))
			if err != nil {
				s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
			}
			if _, exist := s.indexerClients[clusterId]; !exist {
				s.indexerClients[clusterId] = make(map[string]*indexer.GRPCClient)
			}
			s.indexerClients[clusterId][nodeId] = client
		}
	}

	for {
		select {
		case <-s.updateIndexersStopCh:
			s.logger.Info("received a request to stop updating a indexer cluster")
			return
		default:
			client, err = s.getManagerClient()
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}

			stream, err := client.WatchStore("/cluster_config/clusters/")
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}

			s.logger.Info("wait for receive a indexer cluster updates from stream")
			resp, err := stream.Recv()
			if err == io.EOF {
				continue
			}
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}
			s.logger.Debug("data has changed", zap.String("key", resp.Key))

			cluster, err := client.GetValue("/cluster_config/clusters/")
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}
			if cluster == nil {
				s.logger.Error("nil")
				continue
			}
			indexers := *cluster.(*map[string]interface{})

			// compare previous manager with current manager
			if !reflect.DeepEqual(s.indexers, indexers) {
				// create clients for indexer
				for clusterId, cluster := range s.indexers {
					cm, ok := cluster.(map[string]interface{})
					if !ok {
						s.logger.Warn("assertion failed", zap.String("cluster_id", clusterId), zap.Any("cluster", cm))
						continue
					}

					nodes, ok := cm["nodes"].(map[string]interface{})
					if !ok {
						s.logger.Warn("assertion failed", zap.String("cluster_id", clusterId), zap.Any("nodes", nodes))
						continue
					}

					for nodeId, node := range nodes {
						nm, ok := node.(map[string]interface{})
						if !ok {
							s.logger.Warn("assertion failed", zap.String("node_id", nodeId))
							continue
						}

						nodeConfig, ok := nm["node_config"].(map[string]interface{})
						if !ok {
							s.logger.Warn("missing metadata", zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))
							continue
						}

						grpcAddr, ok := nodeConfig["grpc_addr"].(string)
						if !ok {
							s.logger.Warn("missing gRPC address", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
							continue
						}

						client, exist := s.indexerClients[clusterId][nodeId]
						if exist {
							s.logger.Debug("client has already exist in manager list", zap.String("node_id", nodeId))

							if client.GetAddress() != grpcAddr {
								s.logger.Debug("gRPC address has been changed", zap.String("node_id", nodeId), zap.String("client_grpc_addr", client.GetAddress()), zap.String("grpc_addr", grpcAddr))
								s.logger.Debug("recreate gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))

								delete(s.indexerClients[clusterId], nodeId)

								err = client.Close()
								if err != nil {
									s.logger.Error(err.Error(), zap.String("node_id", nodeId))
								}

								newClient, err := indexer.NewGRPCClient(grpcAddr)
								if err != nil {
									s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
								}

								if newClient != nil {
									s.indexerClients[clusterId][nodeId] = newClient
								}
							}

						} else {
							s.logger.Debug("client does not exist in peer list", zap.String("node_id", nodeId))

							s.logger.Debug("create gRPC client", zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
							newClient, err := indexer.NewGRPCClient(nodeConfig["grpc_addr"].(string))
							if err != nil {
								s.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.String("grpc_addr", grpcAddr))
							}
							if _, exist := s.indexerClients[clusterId]; !exist {
								s.indexerClients[clusterId] = make(map[string]*indexer.GRPCClient)
							}
							s.indexerClients[clusterId][nodeId] = newClient
						}
					}
				}

			}
		}
	}
}

func (s *GRPCService) stopUpdateIndexers() {
	s.logger.Info("close all indexer clients")
	for clusterId, cluster := range s.indexerClients {
		for id, client := range cluster {
			s.logger.Debug("close indexer client", zap.String("cluster_id", clusterId), zap.String("id", id), zap.String("address", client.GetAddress()))
			err := client.Close()
			if err != nil {
				s.logger.Error(err.Error())
			}
		}
	}

	if s.updateIndexersStopCh != nil {
		s.logger.Info("send a request to stop updating a index cluster")
		close(s.updateIndexersStopCh)
	}

	s.logger.Info("wait for the indexer cluster update to stop")
	<-s.updateIndexersDoneCh
	s.logger.Info("the indexer cluster update has been stopped")
}

func (s *GRPCService) getIndexerClients() map[string]*indexer.GRPCClient {
	indexerClients := make(map[string]*indexer.GRPCClient, 0)

	for clusterId, cluster := range s.indexerClients {
		nodeIds := make([]string, 0)
		for nodeId := range cluster {
			nodeIds = append(nodeIds, nodeId)
		}

		// pick a client at random
		nodeId := nodeIds[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(nodeIds))]

		indexerClients[clusterId] = s.indexerClients[clusterId][nodeId]
	}

	return indexerClients
}

func (s *GRPCService) LivenessProbe(ctx context.Context, req *empty.Empty) (*distribute.LivenessProbeResponse, error) {
	resp := &distribute.LivenessProbeResponse{
		State: distribute.LivenessProbeResponse_ALIVE,
	}

	return resp, nil
}

func (s *GRPCService) ReadinessProbe(ctx context.Context, req *empty.Empty) (*distribute.ReadinessProbeResponse, error) {
	resp := &distribute.ReadinessProbeResponse{
		State: distribute.ReadinessProbeResponse_READY,
	}

	return resp, nil
}

func (s *GRPCService) GetDocument(ctx context.Context, req *distribute.GetDocumentRequest) (*distribute.GetDocumentResponse, error) {
	indexerClients := s.getIndexerClients()

	// cluster id list sorted by cluster id
	clusterIds := make([]string, 0)
	for clusterId := range indexerClients {
		clusterIds = append(clusterIds, clusterId)
		sort.Strings(clusterIds)
	}

	type respVal struct {
		clusterId string
		fields    map[string]interface{}
		err       error
	}

	// create response channel
	respChan := make(chan respVal, len(clusterIds))

	wg := &sync.WaitGroup{}
	for clusterId, client := range indexerClients {
		wg.Add(1)
		go func(clusterId string, client *indexer.GRPCClient, id string, respChan chan respVal) {
			// index documents
			fields, err := client.GetDocument(id)
			wg.Done()
			respChan <- respVal{
				clusterId: clusterId,
				fields:    fields,
				err:       err,
			}
		}(clusterId, client, req.Id, respChan)
	}
	wg.Wait()

	// close response channel
	close(respChan)

	// summarize responses
	var fields map[string]interface{}
	for r := range respChan {
		if r.fields != nil {
			fields = r.fields
		}
		if r.err != nil {
			s.logger.Error(r.err.Error(), zap.String("cluster_id", r.clusterId))
		}
	}

	resp := &distribute.GetDocumentResponse{}

	fieldsAny := &any.Any{}
	err := protobuf.UnmarshalAny(fields, fieldsAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, err
	}

	// response
	resp.Fields = fieldsAny

	return resp, nil
}

func (s *GRPCService) Search(ctx context.Context, req *distribute.SearchRequest) (*distribute.SearchResponse, error) {
	start := time.Now()

	resp := &distribute.SearchResponse{}

	indexerClients := s.getIndexerClients()

	// cluster id list sorted by cluster id
	clusterIds := make([]string, 0)
	for clusterId := range indexerClients {
		clusterIds = append(clusterIds, clusterId)
		sort.Strings(clusterIds)
	}

	type respVal struct {
		clusterId    string
		searchResult *bleve.SearchResult
		err          error
	}

	// create response channel
	respChan := make(chan respVal, len(clusterIds))

	// create search request
	ins, err := protobuf.MarshalAny(req.SearchRequest)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, err
	}
	searchRequest := ins.(*bleve.SearchRequest)

	// change to distributed search request
	from := searchRequest.From
	size := searchRequest.Size
	searchRequest.From = 0
	searchRequest.Size = from + size

	wg := &sync.WaitGroup{}
	for clusterId, client := range indexerClients {
		wg.Add(1)
		go func(clusterId string, client *indexer.GRPCClient, searchRequest *bleve.SearchRequest, respChan chan respVal) {
			searchResult, err := client.Search(searchRequest)
			wg.Done()
			respChan <- respVal{
				clusterId:    clusterId,
				searchResult: searchResult,
				err:          err,
			}
		}(clusterId, client, searchRequest, respChan)
	}
	wg.Wait()

	// close response channel
	close(respChan)

	// revert to original search request
	searchRequest.From = from
	searchRequest.Size = size

	// summarize responses
	var searchResult *bleve.SearchResult
	for r := range respChan {
		if r.searchResult != nil {
			if searchResult == nil {
				searchResult = r.searchResult
			} else {
				searchResult.Merge(r.searchResult)
			}
		}
		if r.err != nil {
			s.logger.Error(r.err.Error(), zap.String("cluster_id", r.clusterId))
		}
	}

	// handle case where no results were successful
	if searchResult == nil {
		searchResult = &bleve.SearchResult{
			Status: &bleve.SearchStatus{
				Errors: make(map[string]error),
			},
		}
	}

	// sort all hits with the requested order
	if len(searchRequest.Sort) > 0 {
		sorter := sortutils.NewMultiSearchHitSorter(searchRequest.Sort, searchResult.Hits)
		sort.Sort(sorter)
	}

	// now skip over the correct From
	if searchRequest.From > 0 && len(searchResult.Hits) > searchRequest.From {
		searchResult.Hits = searchResult.Hits[searchRequest.From:]
	} else if searchRequest.From > 0 {
		searchResult.Hits = search.DocumentMatchCollection{}
	}

	// now trim to the correct size
	if searchRequest.Size > 0 && len(searchResult.Hits) > searchRequest.Size {
		searchResult.Hits = searchResult.Hits[0:searchRequest.Size]
	}

	// fix up facets
	for name, fr := range searchRequest.Facets {
		searchResult.Facets.Fixup(name, fr.Size)
	}

	// fix up original request
	searchResult.Request = searchRequest
	searchDuration := time.Since(start)
	searchResult.Took = searchDuration

	searchResultAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchResult, searchResultAny)
	if err != nil {
		s.logger.Error(err.Error())
		return resp, err
	}

	// response
	resp.SearchResult = searchResultAny

	return resp, nil
}

func (s *GRPCService) docIdHash(docId string) uint64 {
	hash := fnv.New64()
	_, err := hash.Write([]byte(docId))
	if err != nil {
		return 0
	}

	return hash.Sum64()
}

func (s *GRPCService) IndexDocument(stream distribute.Distribute_IndexDocumentServer) error {
	indexerClients := s.getIndexerClients()

	// cluster id list sorted by cluster id
	clusterIds := make([]string, 0)
	for clusterId := range indexerClients {
		clusterIds = append(clusterIds, clusterId)
		sort.Strings(clusterIds)
	}

	// initialize document list for each cluster
	docSet := make(map[string][]*indexutils.Document, 0)
	for _, clusterId := range clusterIds {
		docSet[clusterId] = make([]*indexutils.Document, 0)
	}

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

		// fields
		ins, err := protobuf.MarshalAny(req.Fields)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}
		fields := *ins.(*map[string]interface{})

		// document
		doc, err := indexutils.NewDocument(req.Id, fields)
		if err != nil {
			s.logger.Error(err.Error())
			return status.Error(codes.Internal, err.Error())
		}

		// distribute documents to each cluster based on document id
		docIdHash := s.docIdHash(req.Id)
		clusterNum := uint64(len(indexerClients))
		clusterId := clusterIds[int(docIdHash%clusterNum)]
		docSet[clusterId] = append(docSet[clusterId], doc)
	}

	type respVal struct {
		clusterId string
		count     int
		err       error
	}

	// create response channel
	respChan := make(chan respVal, len(clusterIds))

	wg := &sync.WaitGroup{}
	for clusterId, docs := range docSet {
		wg.Add(1)
		go func(clusterId string, docs []*indexutils.Document, respChan chan respVal) {
			count, err := indexerClients[clusterId].IndexDocument(docs)
			wg.Done()
			respChan <- respVal{
				clusterId: clusterId,
				count:     count,
				err:       err,
			}
		}(clusterId, docs, respChan)
	}
	wg.Wait()

	// close response channel
	close(respChan)

	// summarize responses
	totalCount := 0
	for r := range respChan {
		if r.count >= 0 {
			totalCount += r.count
		}
		if r.err != nil {
			s.logger.Error(r.err.Error(), zap.String("cluster_id", r.clusterId))
		}
	}

	// response
	resp := &distribute.IndexDocumentResponse{
		Count: int32(totalCount),
	}

	return stream.SendAndClose(resp)
}

func (s *GRPCService) DeleteDocument(stream distribute.Distribute_DeleteDocumentServer) error {
	indexerClients := s.getIndexerClients()

	// cluster id list sorted by cluster id
	clusterIds := make([]string, 0)
	for clusterId := range indexerClients {
		clusterIds = append(clusterIds, clusterId)
		sort.Strings(clusterIds)
	}

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

	type respVal struct {
		clusterId string
		count     int
		err       error
	}

	// create response channel
	respChan := make(chan respVal, len(clusterIds))

	wg := &sync.WaitGroup{}
	for clusterId, client := range indexerClients {
		wg.Add(1)
		go func(clusterId string, client *indexer.GRPCClient, ids []string, respChan chan respVal) {
			// index documents
			count, err := client.DeleteDocument(ids)
			wg.Done()
			respChan <- respVal{
				clusterId: clusterId,
				count:     count,
				err:       err,
			}
		}(clusterId, client, ids, respChan)
	}
	wg.Wait()

	// close response channel
	close(respChan)

	// summarize responses
	totalCount := len(ids)
	for r := range respChan {
		if r.err != nil {
			s.logger.Error(r.err.Error(), zap.String("cluster_id", r.clusterId))
		}
	}

	// response
	resp := &distribute.DeleteDocumentResponse{
		Count: int32(totalCount),
	}

	return stream.SendAndClose(resp)
}
