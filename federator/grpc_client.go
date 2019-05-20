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

package federator

import (
	"context"
	"errors"
	"math"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/blevesearch/bleve"
	"github.com/mosuka/blast/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client protobuf.BlastClient
}

func NewGRPCClient(grpcAddr string) (*GRPCClient, error) {
	baseCtx := context.TODO()
	ctx, cancel := context.WithCancel(baseCtx)

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(math.MaxInt32),
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
		),
	}

	conn, err := grpc.DialContext(ctx, grpcAddr, dialOpts...)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		client: protobuf.NewBlastClient(conn),
	}, nil
}

func (c *GRPCClient) Cancel() {
	c.cancel()
}

func (c *GRPCClient) Close() error {
	c.Cancel()
	if c.conn != nil {
		return c.conn.Close()
	}

	return c.ctx.Err()
}

func (c *GRPCClient) LivenessProbe(opts ...grpc.CallOption) (string, error) {
	livenessStatus, err := c.client.LivenessProbe(c.ctx, &empty.Empty{})
	if err != nil {
		st, _ := status.FromError(err)

		return protobuf.LivenessProbeResponse_UNKNOWN.String(), errors.New(st.Message())
	}

	return livenessStatus.State.String(), nil
}

func (c *GRPCClient) ReadinessProbe(opts ...grpc.CallOption) (string, error) {
	readinessProbe, err := c.client.ReadinessProbe(c.ctx, &empty.Empty{})
	if err != nil {
		st, _ := status.FromError(err)

		return protobuf.ReadinessProbeResponse_UNKNOWN.String(), errors.New(st.Message())
	}

	return readinessProbe.State.String(), nil
}

func (c *GRPCClient) GetDocument(id string, opts ...grpc.CallOption) (map[string]interface{}, error) {
	//retDoc, err := c.client.GetDocument(c.ctx, id, opts...)
	//if err != nil {
	//	st, _ := status.FromError(err)
	//
	//	switch st.Code() {
	//	case codes.NotFound:
	//		return nil, blasterrors.ErrNotFound
	//	default:
	//		return nil, errors.New(st.Message())
	//	}
	//}

	return nil, nil
}

func (c *GRPCClient) Search(searchRequest *bleve.SearchRequest, opts ...grpc.CallOption) (*bleve.SearchResult, error) {
	//// bleve.SearchRequest -> Any
	//searchRequestAny := &any.Any{}
	//err := protobuf.UnmarshalAny(searchRequest, searchRequestAny)
	//if err != nil {
	//	return nil, err
	//}
	//
	//req := &federation.SearchRequest{
	//	SearchRequest: searchRequestAny,
	//}
	//
	//resp, err := c.client.Search(c.ctx, req, opts...)
	//if err != nil {
	//	st, _ := status.FromError(err)
	//
	//	return nil, errors.New(st.Message())
	//}
	//
	//// Any -> bleve.SearchResult
	//searchResultInstance, err := protobuf.MarshalAny(resp.SearchResult)
	//if err != nil {
	//	st, _ := status.FromError(err)
	//
	//	return nil, errors.New(st.Message())
	//}
	//if searchResultInstance == nil {
	//	return nil, errors.New("nil")
	//}
	//searchResult := searchResultInstance.(*bleve.SearchResult)

	return nil, nil
}

func (c *GRPCClient) IndexDocument(docs []map[string]interface{}, opts ...grpc.CallOption) (int, error) {
	//stream, err := c.client.Index(c.ctx, opts...)
	//if err != nil {
	//	st, _ := status.FromError(err)
	//
	//	return nil, errors.New(st.Message())
	//}
	//
	//for _, doc := range docs {
	//	err := stream.Send(doc)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//
	//rep, err := stream.CloseAndRecv()
	//if err != nil {
	//	return nil, err
	//}

	return -1, nil
}

func (c *GRPCClient) DeleteDocument(ids []string, opts ...grpc.CallOption) (int, error) {
	//stream, err := c.client.Delete(c.ctx, opts...)
	//if err != nil {
	//	st, _ := status.FromError(err)
	//
	//	return nil, errors.New(st.Message())
	//}
	//
	//for _, doc := range docs {
	//	err := stream.Send(doc)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//
	//rep, err := stream.CloseAndRecv()
	//if err != nil {
	//	return nil, err
	//}

	return -1, nil
}
