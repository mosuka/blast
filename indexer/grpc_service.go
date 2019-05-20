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
	"time"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	raftServer *RaftServer

	logger *log.Logger
}

func NewGRPCService(raftServer *RaftServer, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		raftServer: raftServer,
		logger:     logger,
	}, nil
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

	return resp, nil
}

func (s *GRPCService) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] leave %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.DeleteNode(req.Id)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
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
		case errors.ErrNotFound:
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
