//  Copyright (c) 2018 Minoru Osuka
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
	"encoding/json"
	"errors"
	"io"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	managerAddr   string
	managerClient *manager.GRPCClient
	watchClient   protobuf.Blast_WatchStateClient
	watchStopCh   chan struct{}
	watchDoneCh   chan struct{}
	logger        *log.Logger
}

func NewGRPCService(managerAddr string, logger *log.Logger) (*GRPCService, error) {
	return &GRPCService{
		managerAddr: managerAddr,
		watchStopCh: make(chan struct{}),
		watchDoneCh: make(chan struct{}),
		logger:      logger,
	}, nil
}

func (s *GRPCService) watch() {
	// notify done
	defer func() { close(s.watchDoneCh) }()

	for {
		select {
		case <-s.watchStopCh:
			s.logger.Printf("[DEBUG] receive request that stop watching a cluster")
			return
		default:
			s.logger.Printf("[DEBUG] wait for status update")
			resp, err := s.watchClient.Recv()
			if err == io.EOF {
				s.logger.Printf("[DEBUG] receive EOF")
				return
			}
			if err != nil {
				st, _ := status.FromError(err)
				switch st.Code() {
				case codes.Canceled:
					s.logger.Printf("[DEBUG] %v", st.Message())
					return
				default:
					s.logger.Printf("[ERR] %v", err)
					continue
				}
			}

			value, err := protobuf.MarshalAny(resp.Value)
			if err != nil {
				s.logger.Printf("[ERR} %v", err)
				continue
			}
			if value == nil {
				s.logger.Printf("[ERR} %v", errors.New("nil"))
				continue
			}

			valueBytes, err := json.Marshal(value)
			if err != nil {
				s.logger.Printf("[ERR} %v", err)
				continue
			}

			s.logger.Printf("[DEBUG} %v", string(valueBytes))
		}
	}
}

func (s *GRPCService) StartWatchCluster() error {
	var err error

	s.managerClient, err = manager.NewGRPCClient(s.managerAddr)
	if err != nil {
		return err
	}

	s.watchClient, err = s.managerClient.Watch("/cluster_config/clusters/")
	if err != nil {
		return err
	}

	s.logger.Printf("[DEBUG] start watching a cluster")

	// start watching a cluster
	go s.watch()

	s.logger.Printf("[DEBUG] started")

	return nil
}

func (s *GRPCService) StopWatchCluster() error {
	s.logger.Printf("[DEBUG] stop watching a cluster")

	// cancel receiving stream
	s.logger.Printf("[DEBUG] cancel revieving stream")
	s.managerClient.Cancel()

	// stop watching a cluster
	close(s.watchStopCh)

	// wait for stop watching a cluster has done
	<-s.watchDoneCh

	err := s.managerClient.Close()
	if err != nil {
		return err
	}

	return nil
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
