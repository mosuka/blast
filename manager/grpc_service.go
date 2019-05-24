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
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	raftServer *RaftServer
	logger     *log.Logger

	clusterChans map[chan protobuf.GetClusterResponse]struct{}
	clusterMutex sync.RWMutex
	stateChans   map[chan protobuf.WatchStateResponse]struct{}
	stateMutex   sync.RWMutex

	watchClusterStopCh chan struct{}
	watchClusterDoneCh chan struct{}
}

func NewGRPCService(raftServer *RaftServer, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: raftServer,
		logger:     logger,

		clusterChans: make(map[chan protobuf.GetClusterResponse]struct{}),
		stateChans:   make(map[chan protobuf.WatchStateResponse]struct{}),

		watchClusterStopCh: make(chan struct{}),
		watchClusterDoneCh: make(chan struct{}),
	}, nil
}

func (s *GRPCService) Start() error {
	// start watching a cluster
	s.logger.Printf("[INFO] start watching a cluster")
	go s.watchCluster(1000 * time.Millisecond)

	return nil
}

func (s *GRPCService) Stop() error {
	s.logger.Printf("[INFO] stop watching a cluster")
	close(s.watchClusterStopCh)
	s.logger.Printf("[INFO] wait for stop watching a cluster has done")
	<-s.watchClusterDoneCh
	s.logger.Printf("[INFO] cluster watching has been stopped")

	return nil
}

func (s *GRPCService) watchCluster(checkInterval time.Duration) {
	defer func() {
		close(s.watchClusterDoneCh)
	}()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	savedCluster := map[string]interface{}{}

	for {
		select {
		case <-s.watchClusterStopCh:
			s.logger.Printf("[DEBUG] receive request that stop watching a cluster")
			//cancel()
			return
		case <-ticker.C:
			// only the leader node watches own cluster
			if s.raftServer.IsLeader() {
				// get the cluster
				cluster, err := s.raftServer.GetCluster() // TODO: if there is a problem, return an error and do not wait for the timeout
				if err != nil {
					s.logger.Printf("[ERR] %v", err)
					return
				}

				// TODO: update cluster health

				// notify when there is a change in the cluster
				if !reflect.DeepEqual(savedCluster, cluster) {
					s.logger.Printf("[INFO] cluster has chaged")
					s.logger.Printf("[DEBUG] %v", cluster)

					// save latest cluster
					savedCluster = cluster

					// create cluster response
					clusterAny := &any.Any{}
					err = protobuf.UnmarshalAny(cluster, clusterAny)
					if err != nil {
						s.logger.Printf("[ERR] %v", err)
						continue
					}
					getClusterResponse := &protobuf.GetClusterResponse{Cluster: clusterAny}

					// notify cluster changes
					for c := range s.clusterChans {
						c <- *getClusterResponse
					}
				}
			} else {
				s.logger.Printf("[INFO] not a leader")
			}
		default:
			// avoid being blocked
		}
	}
}

func (s *GRPCService) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	s.logger.Printf("[INFO] get node %v", req)

	resp := &protobuf.GetNodeResponse{}

	var err error

	metadata, err := s.raftServer.GetNode(req.Id)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	metadataAny := &any.Any{}
	err = protobuf.UnmarshalAny(metadata, metadataAny)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	resp.Metadata = metadataAny

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

	err = s.raftServer.SetNode(req.Id, metadata)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	getClusterResponse, err := s.GetCluster(ctx, &empty.Empty{})
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	// notify
	for c := range s.clusterChans {
		c <- *getClusterResponse
	}

	return resp, nil
}

func (s *GRPCService) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] leave %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.DeleteNode(req.Id)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	getClusterResponse, err := s.GetCluster(ctx, &empty.Empty{})
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	// notify
	for c := range s.clusterChans {
		c <- *getClusterResponse
	}

	return resp, nil
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*protobuf.GetClusterResponse, error) {
	s.logger.Printf("[INFO] get cluster %v", req)

	resp := &protobuf.GetClusterResponse{}

	cluster, err := s.raftServer.GetCluster()
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
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

	var err error

	value, err := s.raftServer.Get(req.Key)
	if err != nil {
		switch err {
		case errors.ErrNotFound:
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
		case errors.ErrNotFound:
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
		case errors.ErrNotFound:
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
