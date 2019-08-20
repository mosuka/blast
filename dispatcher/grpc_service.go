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
	"encoding/json"
	"errors"
	"hash/fnv"
	"io"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/indexer"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/distribute"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/mosuka/blast/sortutils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	managerGrpcAddress string
	logger             *zap.Logger

	managers             *management.Cluster
	managerClients       map[string]*manager.GRPCClient
	updateManagersStopCh chan struct{}
	updateManagersDoneCh chan struct{}

	indexers             map[string]*index.Cluster
	indexerClients       map[string]map[string]*indexer.GRPCClient
	updateIndexersStopCh chan struct{}
	updateIndexersDoneCh chan struct{}
}

func NewGRPCService(managerGrpcAddress string, logger *zap.Logger) (*GRPCService, error) {
	return &GRPCService{
		managerGrpcAddress: managerGrpcAddress,
		logger:             logger,

		managers:       &management.Cluster{Nodes: make(map[string]*management.Node, 0)},
		managerClients: make(map[string]*manager.GRPCClient, 0),

		indexers:       make(map[string]*index.Cluster, 0),
		indexerClients: make(map[string]map[string]*indexer.GRPCClient, 0),
	}, nil
}

func (s *GRPCService) Start() error {
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

	for id, node := range s.managers.Nodes {
		if node.Metadata == nil {
			s.logger.Warn("missing metadata", zap.String("id", id))
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
			//if err == io.EOF {
			//	s.logger.Info(err.Error())
			//	continue
			//}
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

func (s *GRPCService) startUpdateIndexers(checkInterval time.Duration) {
	s.updateIndexersStopCh = make(chan struct{})
	s.updateIndexersDoneCh = make(chan struct{})

	defer func() {
		close(s.updateIndexersDoneCh)
	}()

	// get active client for manager
	client, err := s.getManagerClient()
	if err != nil {
		s.logger.Error(err.Error())
	}

	// get initial indexers
	shards, err := client.Get("/cluster/shards")
	if err != nil {
		s.logger.Fatal(err.Error())
		return
	}
	if shards == nil {
		s.logger.Error("/cluster/shards is nil")
	}

	for shardId, shard := range *shards.(*map[string]interface{}) {
		shardBytes, err := json.Marshal(shard)
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}

		var cluster *index.Cluster
		err = json.Unmarshal(shardBytes, &cluster)
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}

		s.indexers[shardId] = cluster

		for nodeId, node := range cluster.Nodes {
			if node.Metadata.GrpcAddress == "" {
				s.logger.Warn("missing gRPC address", zap.String("id", node.Id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
				continue
			}
			newClient, err := indexer.NewGRPCClient(node.Metadata.GrpcAddress)
			if err != nil {
				s.logger.Error(err.Error(), zap.String("id", nodeId), zap.String("grpc_addr", node.Metadata.GrpcAddress))
				continue
			}
			if _, exist := s.indexerClients[shardId]; !exist {
				s.indexerClients[shardId] = make(map[string]*indexer.GRPCClient)
			}
			s.indexerClients[shardId][nodeId] = newClient
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

			stream, err := client.Watch("/cluster/shards/")
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
			s.logger.Debug("data has changed", zap.Any("command", resp.Command), zap.String("key", resp.Key), zap.Any("value", resp.Value))

			shards, err := client.Get("/cluster/shards/")
			if err != nil {
				s.logger.Error(err.Error())
				continue
			}
			if shards == nil {
				s.logger.Error("/cluster/shards is nil")
				continue
			}

			for shardId, shard := range *shards.(*map[string]interface{}) {
				shardBytes, err := json.Marshal(shard)
				if err != nil {
					s.logger.Error(err.Error())
					continue
				}

				var cluster *index.Cluster
				err = json.Unmarshal(shardBytes, &cluster)
				if err != nil {
					s.logger.Error(err.Error())
					continue
				}

				s.indexers[shardId] = cluster

				if _, exist := s.indexerClients[shardId]; !exist {
					s.indexerClients[shardId] = make(map[string]*indexer.GRPCClient)
				}

				// open clients for indexer nodes
				for nodeId, node := range cluster.Nodes {
					if node.Metadata.GrpcAddress == "" {
						s.logger.Warn("missing gRPC address", zap.String("id", node.Id), zap.String("grpc_addr", node.Metadata.GrpcAddress))
						continue
					}

					// check client that already exist in the client list
					if client, exist := s.indexerClients[shardId][node.Id]; !exist {
						// create new client
						newClient, err := indexer.NewGRPCClient(node.Metadata.GrpcAddress)
						if err != nil {
							s.logger.Error(err.Error(), zap.String("id", nodeId), zap.String("grpc_addr", node.Metadata.GrpcAddress))
							continue
						}
						s.indexerClients[shardId][nodeId] = newClient
					} else {
						if client.GetAddress() != node.Metadata.GrpcAddress {
							// close client
							s.logger.Info("close gRPC client", zap.String("id", node.Id), zap.String("grpc_addr", client.GetAddress()))
							err = client.Close()
							if err != nil {
								s.logger.Warn(err.Error(), zap.String("id", node.Id))
							}
							delete(s.indexerClients[shardId], node.Id)

							// re-create new client
							newClient, err := indexer.NewGRPCClient(node.Metadata.GrpcAddress)
							if err != nil {
								s.logger.Error(err.Error(), zap.String("id", nodeId), zap.String("grpc_addr", node.Metadata.GrpcAddress))
								continue
							}
							s.indexerClients[shardId][nodeId] = newClient
						}
					}
				}

				// close clients for non-existent indexer nodes
				for id, client := range s.indexerClients[shardId] {
					if _, exist := s.indexers[shardId].Nodes[id]; !exist {
						s.logger.Info("close gRPC client", zap.String("id", id), zap.String("grpc_addr", client.GetAddress()))
						err = client.Close()
						if err != nil {
							s.logger.Warn(err.Error(), zap.String("id", id), zap.String("grpc_addr", client.GetAddress()))
						}
						delete(s.indexerClients[shardId], id)
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

func (s *GRPCService) NodeHealthCheck(ctx context.Context, req *distribute.NodeHealthCheckRequest) (*distribute.NodeHealthCheckResponse, error) {
	resp := &distribute.NodeHealthCheckResponse{}

	switch req.Probe {
	case distribute.NodeHealthCheckRequest_HEALTHINESS:
		resp.State = distribute.NodeHealthCheckResponse_HEALTHY
	case distribute.NodeHealthCheckRequest_LIVENESS:
		resp.State = distribute.NodeHealthCheckResponse_ALIVE
	case distribute.NodeHealthCheckRequest_READINESS:
		resp.State = distribute.NodeHealthCheckResponse_READY
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
		doc       *index.Document
		err       error
	}

	// create response channel
	respChan := make(chan respVal, len(clusterIds))

	wg := &sync.WaitGroup{}
	for clusterId, client := range indexerClients {
		wg.Add(1)
		go func(clusterId string, client *indexer.GRPCClient, id string, respChan chan respVal) {
			// index documents
			doc, err := client.GetDocument(id)
			wg.Done()
			respChan <- respVal{
				clusterId: clusterId,
				doc:       doc,
				err:       err,
			}
		}(clusterId, client, req.Id, respChan)
	}
	wg.Wait()

	// close response channel
	close(respChan)

	// summarize responses
	var doc *index.Document
	for r := range respChan {
		if r.doc != nil {
			doc = r.doc
		}
		if r.err != nil {
			s.logger.Error(r.err.Error(), zap.String("cluster_id", r.clusterId))
		}
	}

	resp := &distribute.GetDocumentResponse{}

	// response
	resp.Document = doc

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
	docSet := make(map[string][]*index.Document, 0)
	for _, clusterId := range clusterIds {
		docSet[clusterId] = make([]*index.Document, 0)
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

		// distribute documents to each cluster based on document id
		docIdHash := s.docIdHash(req.Document.Id)
		clusterNum := uint64(len(indexerClients))
		clusterId := clusterIds[int(docIdHash%clusterNum)]
		docSet[clusterId] = append(docSet[clusterId], req.Document)
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
		go func(clusterId string, docs []*index.Document, respChan chan respVal) {
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
