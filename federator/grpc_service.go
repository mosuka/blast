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
	"hash/fnv"
	"io"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	*grpc.Service

	managerAddr string

	logger *log.Logger

	managers            map[string]interface{}
	managerClients      map[string]*grpc.Client
	watchManagersStopCh chan struct{}
	watchManagersDoneCh chan struct{}

	indexers            map[string]interface{}
	indexerClients      map[string]map[string]*grpc.Client
	watchIndexersStopCh chan struct{}
	watchIndexersDoneCh chan struct{}
}

func NewGRPCService(managerAddr string, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		managerAddr: managerAddr,
		logger:      logger,

		managers:       make(map[string]interface{}, 0),
		managerClients: make(map[string]*grpc.Client, 0),

		indexers:       make(map[string]interface{}, 0),
		indexerClients: make(map[string]map[string]*grpc.Client, 0),
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
		if err != nil {
			s.logger.Printf("[ERR] %v", err)
		}
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

		client, err := grpc.NewClient(metadata["grpc_addr"].(string))
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
		for nodeId, node := range cluster["nodes"].(map[string]interface{}) {
			metadata := node.(map[string]interface{})["metadata"].(map[string]interface{})

			s.logger.Printf("[DEBUG] create indexer client for %s at %s", metadata["grpc_addr"].(string), clusterId)

			client, err := grpc.NewClient(metadata["grpc_addr"].(string))
			if err != nil {
				s.logger.Printf("[ERR] %v", err)
				continue
			}

			if _, exist := s.indexerClients[clusterId]; !exist {
				s.indexerClients[clusterId] = make(map[string]*grpc.Client)
			}

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
			stream, err := client.WatchState("/cluster_config/clusters/")
			if err != nil {
				st, _ := status.FromError(err)
				switch st.Code() {
				case codes.Canceled:
					s.logger.Printf("[DEBUG] %s: %v", client.GetAddress(), err)
				default:
					s.logger.Printf("[ERR] %s: %v", client.GetAddress(), err)
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

func (s *GRPCService) getIndexerClients() map[string]*grpc.Client {
	indexerClients := make(map[string]*grpc.Client, 0)

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

func (s *GRPCService) GetDocument(ctx context.Context, req *protobuf.GetDocumentRequest) (*protobuf.GetDocumentResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "get")

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
		go func(clusterId string, client *grpc.Client, id string, respChan chan respVal) {
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
			s.logger.Printf("[ERR] %s %v", r.clusterId, r.err)
		}
	}

	resp := &protobuf.GetDocumentResponse{}

	fieldsAny := &any.Any{}
	err := protobuf.UnmarshalAny(fields, fieldsAny)
	if err != nil {
		return resp, err
	}

	// response
	resp.Fields = fieldsAny

	return resp, nil
}

func (s *GRPCService) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "search")

	resp := &protobuf.SearchResponse{}

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

	searchRequest, err := protobuf.MarshalAny(req.SearchRequest)
	if err != nil {
		return resp, err
	}

	// create response channel
	respChan := make(chan respVal, len(clusterIds))

	wg := &sync.WaitGroup{}
	for clusterId, client := range indexerClients {
		wg.Add(1)
		go func(clusterId string, client *grpc.Client, searchRequest *bleve.SearchRequest, respChan chan respVal) {
			// index documents
			searchResult, err := client.Search(searchRequest)
			wg.Done()
			respChan <- respVal{
				clusterId:    clusterId,
				searchResult: searchResult,
				err:          err,
			}
		}(clusterId, client, searchRequest.(*bleve.SearchRequest), respChan)
	}
	wg.Wait()

	// close response channel
	close(respChan)

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
			s.logger.Printf("[ERR] %s %v", r.clusterId, r.err)
		}
	}

	searchResultAny := &any.Any{}
	err = protobuf.UnmarshalAny(searchResult, searchResultAny)
	if err != nil {
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

func (s *GRPCService) IndexDocument(stream protobuf.Blast_IndexDocumentServer) error {
	indexerClients := s.getIndexerClients()

	// cluster id list sorted by cluster id
	clusterIds := make([]string, 0)
	for clusterId := range indexerClients {
		clusterIds = append(clusterIds, clusterId)
		sort.Strings(clusterIds)
	}

	// initialize document list for each cluster
	docSet := make(map[string][]map[string]interface{}, 0)
	for _, clusterId := range clusterIds {
		docSet[clusterId] = make([]map[string]interface{}, 0)
	}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// fields
		ins, err := protobuf.MarshalAny(req.Fields)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		fields := *ins.(*map[string]interface{})

		// document
		doc := map[string]interface{}{
			"id":     req.Id,
			"fields": fields,
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
		go func(clusterId string, docs []map[string]interface{}, respChan chan respVal) {
			// index documents
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
			s.logger.Printf("[ERR] %s %v", r.clusterId, r.err)
		}
	}

	// response
	resp := &protobuf.IndexDocumentResponse{
		Count: int32(totalCount),
	}

	return stream.SendAndClose(resp)
}

func (s *GRPCService) DeleteDocument(stream protobuf.Blast_DeleteDocumentServer) error {
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
		if err == io.EOF {
			break
		}
		if err != nil {
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
		go func(clusterId string, client *grpc.Client, ids []string, respChan chan respVal) {
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
	totalCount := 0
	for r := range respChan {
		if r.count >= 0 {
			totalCount += r.count
		}
		if r.err != nil {
			s.logger.Printf("[ERR] %s %v", r.clusterId, r.err)
		}
	}

	// response
	resp := &protobuf.DeleteDocumentResponse{
		Count: int32(totalCount),
	}

	return stream.SendAndClose(resp)
}
