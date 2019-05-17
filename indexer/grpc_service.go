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
	"github.com/mosuka/blast/protobuf/index"
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

func (s *GRPCService) GetNode(ctx context.Context, req *index.GetNodeRequest) (*index.GetNodeResponse, error) {
	s.logger.Printf("[INFO] get node %v", req)

	resp := &index.GetNodeResponse{}

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

func (s *GRPCService) SetNode(ctx context.Context, req *index.SetNodeRequest) (*empty.Empty, error) {
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

func (s *GRPCService) DeleteNode(ctx context.Context, req *index.DeleteNodeRequest) (*empty.Empty, error) {
	s.logger.Printf("[INFO] leave %v", req)

	resp := &empty.Empty{}

	err := s.raftServer.DeleteNode(req.Id)
	if err != nil {
		return resp, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *GRPCService) GetCluster(ctx context.Context, req *empty.Empty) (*index.GetClusterResponse, error) {
	s.logger.Printf("[INFO] get cluster %v", req)

	resp := &index.GetClusterResponse{}

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

func (s *GRPCService) LivenessProbe(ctx context.Context, req *empty.Empty) (*index.LivenessProbeResponse, error) {
	resp := &index.LivenessProbeResponse{
		State: index.LivenessProbeResponse_ALIVE,
	}

	return resp, nil
}

func (s *GRPCService) ReadinessProbe(ctx context.Context, req *empty.Empty) (*index.ReadinessProbeResponse, error) {
	resp := &index.ReadinessProbeResponse{
		State: index.ReadinessProbeResponse_READY,
	}

	return resp, nil
}

func (s *GRPCService) GetDocument(ctx context.Context, req *index.GetDocumentRequest) (*index.GetDocumentResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "get")

	s.logger.Printf("[INFO] get %v", req)

	resp := &index.GetDocumentResponse{}

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

func (s *GRPCService) Search(ctx context.Context, req *index.SearchRequest) (*index.SearchResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "search")

	s.logger.Printf("[INFO] search %v", req)

	resp := &index.SearchResponse{}

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

func (s *GRPCService) IndexDocument(stream index.Index_IndexDocumentServer) error {
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

	resp := &index.IndexDocumentResponse{
		Count: int32(count),
	}

	return stream.SendAndClose(resp)
}

func (s *GRPCService) DeleteDocument(stream index.Index_DeleteDocumentServer) error {
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

	resp := &index.DeleteDocumentResponse{
		Count: int32(count),
	}

	return stream.SendAndClose(resp)
}

func (s *GRPCService) GetIndexConfig(ctx context.Context, req *empty.Empty) (*index.GetIndexConfigResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "indexconfig")

	resp := &index.GetIndexConfigResponse{}

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

func (s *GRPCService) GetIndexStats(ctx context.Context, req *empty.Empty) (*index.GetIndexStatsResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "indexstats")

	resp := &index.GetIndexStatsResponse{}

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
