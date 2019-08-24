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
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/protobuf/distribute"
	"google.golang.org/grpc"
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

func (c *GRPCClient) NodeHealthCheck(req *distribute.NodeHealthCheckRequest, opts ...grpc.CallOption) (*distribute.NodeHealthCheckResponse, error) {
	return c.client.NodeHealthCheck(c.ctx, req, opts...)
}

func (c *GRPCClient) Get(req *distribute.GetRequest, opts ...grpc.CallOption) (*distribute.GetResponse, error) {
	return c.client.Get(c.ctx, req, opts...)
}

func (c *GRPCClient) Index(req *distribute.IndexRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.Index(c.ctx, req, opts...)
}

func (c *GRPCClient) Delete(req *distribute.DeleteRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.Delete(c.ctx, req, opts...)
}

func (c *GRPCClient) BulkIndex(req *distribute.BulkIndexRequest, opts ...grpc.CallOption) (*distribute.BulkIndexResponse, error) {
	return c.client.BulkIndex(c.ctx, req, opts...)
}

func (c *GRPCClient) BulkDelete(req *distribute.BulkDeleteRequest, opts ...grpc.CallOption) (*distribute.BulkDeleteResponse, error) {
	return c.client.BulkDelete(c.ctx, req, opts...)
}

func (c *GRPCClient) Search(req *distribute.SearchRequest, opts ...grpc.CallOption) (*distribute.SearchResponse, error) {
	return c.client.Search(c.ctx, req, opts...)
}
