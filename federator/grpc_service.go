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
	"github.com/mosuka/blast/protobuf/federation"
	"github.com/mosuka/blast/protobuf/management"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	managerAddr   string
	managerClient *manager.GRPCClient
	watchClient   management.Management_WatchClient
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

func (s *GRPCService) LivenessProbe(ctx context.Context, req *empty.Empty) (*federation.LivenessStatus, error) {
	resp := &federation.LivenessStatus{
		State: federation.LivenessStatus_ALIVE,
	}

	return resp, nil
}

func (s *GRPCService) ReadinessProbe(ctx context.Context, req *empty.Empty) (*federation.ReadinessStatus, error) {
	resp := &federation.ReadinessStatus{
		State: federation.ReadinessStatus_READY,
	}

	return resp, nil
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

func (s *GRPCService) Get(ctx context.Context, req *federation.Document) (*federation.Document, error) {
	start := time.Now()
	defer RecordMetrics(start, "get")

	s.logger.Printf("[INFO] get %v", req)

	resp := &federation.Document{}

	return resp, nil
}

func (s *GRPCService) Search(ctx context.Context, req *federation.SearchRequest) (*federation.SearchResponse, error) {
	start := time.Now()
	defer RecordMetrics(start, "search")

	s.logger.Printf("[INFO] search %v", req)

	resp := &federation.SearchResponse{}

	return resp, nil
}

func (s *GRPCService) Index(stream federation.Federation_IndexServer) error {
	docs := make([]*federation.Document, 0)

	for {
		doc, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		docs = append(docs, doc)
	}

	//// index
	//result, err := s.raftServer.Index(docs)
	//if err != nil {
	//	return status.Error(codes.Internal, err.Error())
	//}

	//return stream.SendAndClose(result)

	return stream.SendAndClose(nil)
}

func (s *GRPCService) Delete(stream federation.Federation_DeleteServer) error {
	docs := make([]*federation.Document, 0)

	for {
		doc, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		docs = append(docs, doc)
	}

	//// delete
	//result, err := s.raftServer.Delete(docs)
	//if err != nil {
	//	return status.Error(codes.Internal, err.Error())
	//}

	//return stream.SendAndClose(result)

	return stream.SendAndClose(nil)
}
