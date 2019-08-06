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

package dispatcher

import (
	"context"
	"errors"
	"math"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/distribute"
	"github.com/mosuka/blast/protobuf/index"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client distribute.DistributeClient
}

func NewGRPCContext() (context.Context, context.CancelFunc) {
	baseCtx := context.TODO()
	//return context.WithTimeout(baseCtx, 60*time.Second)
	return context.WithCancel(baseCtx)
}

func NewGRPCClient(address string) (*GRPCClient, error) {
	ctx, cancel := NewGRPCContext()

	//streamRetryOpts := []grpc_retry.CallOption{
	//	grpc_retry.Disable(),
	//}

	//unaryRetryOpts := []grpc_retry.CallOption{
	//	grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
	//	grpc_retry.WithCodes(codes.Unavailable),
	//	grpc_retry.WithMax(100),
	//}

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(math.MaxInt32),
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
		),
		//grpc.WithStreamInterceptor(grpc_retry.StreamClientInterceptor(streamRetryOpts...)),
		//grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(unaryRetryOpts...)),
	}

	conn, err := grpc.DialContext(ctx, address, dialOpts...)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		client: distribute.NewDistributeClient(conn),
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

func (c *GRPCClient) GetAddress() string {
	return c.conn.Target()
}

func (c *GRPCClient) NodeHealthCheck(probe string, opts ...grpc.CallOption) (string, error) {
	req := &distribute.NodeHealthCheckRequest{}

	switch probe {
	case distribute.NodeHealthCheckRequest_HEALTHINESS.String():
		req.Probe = distribute.NodeHealthCheckRequest_HEALTHINESS
	case distribute.NodeHealthCheckRequest_LIVENESS.String():
		req.Probe = distribute.NodeHealthCheckRequest_LIVENESS
	case distribute.NodeHealthCheckRequest_READINESS.String():
		req.Probe = distribute.NodeHealthCheckRequest_READINESS
	default:
		req.Probe = distribute.NodeHealthCheckRequest_HEALTHINESS
	}

	resp, err := c.client.NodeHealthCheck(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return distribute.NodeHealthCheckResponse_UNHEALTHY.String(), errors.New(st.Message())
	}

	return resp.State.String(), nil
}

func (c *GRPCClient) GetDocument(id string, opts ...grpc.CallOption) (*index.Document, error) {
	req := &distribute.GetDocumentRequest{
		Id: id,
	}

	resp, err := c.client.GetDocument(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		switch st.Code() {
		case codes.NotFound:
			return nil, blasterrors.ErrNotFound
		default:
			return nil, errors.New(st.Message())
		}
	}

	return resp.Document, nil
}

func (c *GRPCClient) Search(searchRequest *bleve.SearchRequest, opts ...grpc.CallOption) (*bleve.SearchResult, error) {
	// bleve.SearchRequest -> Any
	searchRequestAny := &any.Any{}
	err := protobuf.UnmarshalAny(searchRequest, searchRequestAny)
	if err != nil {
		return nil, err
	}

	req := &distribute.SearchRequest{
		SearchRequest: searchRequestAny,
	}

	resp, err := c.client.Search(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	// Any -> bleve.SearchResult
	searchResultInstance, err := protobuf.MarshalAny(resp.SearchResult)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}
	if searchResultInstance == nil {
		return nil, errors.New("nil")
	}
	searchResult := searchResultInstance.(*bleve.SearchResult)

	return searchResult, nil
}

func (c *GRPCClient) IndexDocument(docs []*index.Document, opts ...grpc.CallOption) (int, error) {
	stream, err := c.client.IndexDocument(c.ctx, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return -1, errors.New(st.Message())
	}

	for _, doc := range docs {
		req := &distribute.IndexDocumentRequest{
			Document: doc,
		}

		err = stream.Send(req)
		if err != nil {
			return -1, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return -1, err
	}

	return int(resp.Count), nil
}

func (c *GRPCClient) DeleteDocument(ids []string, opts ...grpc.CallOption) (int, error) {
	stream, err := c.client.DeleteDocument(c.ctx, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return -1, errors.New(st.Message())
	}

	for _, id := range ids {
		req := &distribute.DeleteDocumentRequest{
			Id: id,
		}

		err := stream.Send(req)
		if err != nil {
			return -1, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return -1, err
	}

	return int(resp.Count), nil
}
