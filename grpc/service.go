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

package grpc

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct{}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Stop() error {
	return nil
}

func (s *Service) LivenessProbe(ctx context.Context, req *empty.Empty) (*protobuf.LivenessProbeResponse, error) {
	resp := &protobuf.LivenessProbeResponse{
		State: protobuf.LivenessProbeResponse_ALIVE,
	}

	return resp, nil
}

func (s *Service) ReadinessProbe(ctx context.Context, req *empty.Empty) (*protobuf.ReadinessProbeResponse, error) {
	resp := &protobuf.ReadinessProbeResponse{
		State: protobuf.ReadinessProbeResponse_READY,
	}

	return resp, nil
}

func (s *Service) GetNode(ctx context.Context, req *protobuf.GetNodeRequest) (*protobuf.GetNodeResponse, error) {
	return &protobuf.GetNodeResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) SetNode(ctx context.Context, req *protobuf.SetNodeRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) DeleteNode(ctx context.Context, req *protobuf.DeleteNodeRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) GetCluster(ctx context.Context, req *empty.Empty) (*protobuf.GetClusterResponse, error) {
	return &protobuf.GetClusterResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) WatchCluster(req *empty.Empty, server protobuf.Blast_WatchClusterServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *Service) Snapshot(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) GetState(ctx context.Context, req *protobuf.GetStateRequest) (*protobuf.GetStateResponse, error) {
	return &protobuf.GetStateResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) SetState(ctx context.Context, req *protobuf.SetStateRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) DeleteState(ctx context.Context, req *protobuf.DeleteStateRequest) (*empty.Empty, error) {
	return &empty.Empty{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) WatchState(req *protobuf.WatchStateRequest, server protobuf.Blast_WatchStateServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *Service) GetDocument(ctx context.Context, req *protobuf.GetDocumentRequest) (*protobuf.GetDocumentResponse, error) {
	return &protobuf.GetDocumentResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) Search(ctx context.Context, req *protobuf.SearchRequest) (*protobuf.SearchResponse, error) {
	return &protobuf.SearchResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) IndexDocument(stream protobuf.Blast_IndexDocumentServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *Service) DeleteDocument(stream protobuf.Blast_DeleteDocumentServer) error {
	return status.Error(codes.Unavailable, "not implement")
}

func (s *Service) GetIndexConfig(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexConfigResponse, error) {
	return &protobuf.GetIndexConfigResponse{}, status.Error(codes.Unavailable, "not implement")
}

func (s *Service) GetIndexStats(ctx context.Context, req *empty.Empty) (*protobuf.GetIndexStatsResponse, error) {
	return &protobuf.GetIndexStatsResponse{}, status.Error(codes.Unavailable, "not implement")
}
