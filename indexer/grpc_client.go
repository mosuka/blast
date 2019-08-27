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
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/protobuf/index"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client index.IndexClient
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
		client: index.NewIndexClient(conn),
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

func (c *GRPCClient) NodeHealthCheck(req *index.NodeHealthCheckRequest, opts ...grpc.CallOption) (*index.NodeHealthCheckResponse, error) {
	return c.client.NodeHealthCheck(c.ctx, req, opts...)
}

func (c *GRPCClient) NodeInfo(req *empty.Empty, opts ...grpc.CallOption) (*index.NodeInfoResponse, error) {
	return c.client.NodeInfo(c.ctx, req, opts...)
}

func (c *GRPCClient) ClusterJoin(req *index.ClusterJoinRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.ClusterJoin(c.ctx, req, opts...)
}

func (c *GRPCClient) ClusterLeave(req *index.ClusterLeaveRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.ClusterLeave(c.ctx, req, opts...)
}

func (c *GRPCClient) ClusterInfo(req *empty.Empty, opts ...grpc.CallOption) (*index.ClusterInfoResponse, error) {
	return c.client.ClusterInfo(c.ctx, &empty.Empty{}, opts...)
}

func (c *GRPCClient) ClusterWatch(req *empty.Empty, opts ...grpc.CallOption) (index.Index_ClusterWatchClient, error) {
	return c.client.ClusterWatch(c.ctx, req, opts...)
}

func (c *GRPCClient) Get(req *index.GetRequest, opts ...grpc.CallOption) (*index.GetResponse, error) {
	return c.client.Get(c.ctx, req, opts...)
}

func (c *GRPCClient) Index(req *index.IndexRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.Index(c.ctx, req, opts...)
}

func (c *GRPCClient) Delete(req *index.DeleteRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.Delete(c.ctx, req, opts...)
}

func (c *GRPCClient) BulkIndex(req *index.BulkIndexRequest, opts ...grpc.CallOption) (*index.BulkIndexResponse, error) {
	return c.client.BulkIndex(c.ctx, req, opts...)
}

func (c *GRPCClient) BulkDelete(req *index.BulkDeleteRequest, opts ...grpc.CallOption) (*index.BulkDeleteResponse, error) {
	return c.client.BulkDelete(c.ctx, req, opts...)
}

func (c *GRPCClient) Search(req *index.SearchRequest, opts ...grpc.CallOption) (*index.SearchResponse, error) {
	return c.client.Search(c.ctx, req, opts...)
}

func (c *GRPCClient) GetIndexConfig(req *empty.Empty, opts ...grpc.CallOption) (*index.GetIndexConfigResponse, error) {
	return c.client.GetIndexConfig(c.ctx, &empty.Empty{}, opts...)
}

func (c *GRPCClient) GetIndexStats(req *empty.Empty, opts ...grpc.CallOption) (*index.GetIndexStatsResponse, error) {
	return c.client.GetIndexStats(c.ctx, &empty.Empty{}, opts...)
}

func (c *GRPCClient) Snapshot(req *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.Snapshot(c.ctx, &empty.Empty{})
}
