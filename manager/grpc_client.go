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
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/protobuf/management"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client management.ManagementClient
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
		client: management.NewManagementClient(conn),
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

func (c *GRPCClient) NodeHealthCheck(req *management.NodeHealthCheckRequest, opts ...grpc.CallOption) (*management.NodeHealthCheckResponse, error) {
	return c.client.NodeHealthCheck(c.ctx, req, opts...)
}

func (c *GRPCClient) NodeInfo(req *empty.Empty, opts ...grpc.CallOption) (*management.NodeInfoResponse, error) {
	return c.client.NodeInfo(c.ctx, req, opts...)
}

func (c *GRPCClient) ClusterJoin(req *management.ClusterJoinRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.ClusterJoin(c.ctx, req, opts...)
}

func (c *GRPCClient) ClusterLeave(req *management.ClusterLeaveRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.ClusterLeave(c.ctx, req, opts...)
}

func (c *GRPCClient) ClusterInfo(req *empty.Empty, opts ...grpc.CallOption) (*management.ClusterInfoResponse, error) {
	return c.client.ClusterInfo(c.ctx, &empty.Empty{}, opts...)
}

func (c *GRPCClient) ClusterWatch(req *empty.Empty, opts ...grpc.CallOption) (management.Management_ClusterWatchClient, error) {
	return c.client.ClusterWatch(c.ctx, req, opts...)
}

func (c *GRPCClient) Get(req *management.GetRequest, opts ...grpc.CallOption) (*management.GetResponse, error) {
	res, err := c.client.Get(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			return &management.GetResponse{}, nil
		default:
			return nil, err
		}
	}
	return res, nil
}

func (c *GRPCClient) Set(req *management.SetRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.Set(c.ctx, req, opts...)
}

func (c *GRPCClient) Delete(req *management.DeleteRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	res, err := c.client.Delete(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			return &empty.Empty{}, nil
		default:
			return nil, err
		}
	}
	return res, nil
}

func (c *GRPCClient) Watch(req *management.WatchRequest, opts ...grpc.CallOption) (management.Management_WatchClient, error) {
	return c.client.Watch(c.ctx, req, opts...)
}

func (c *GRPCClient) Snapshot(req *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	return c.client.Snapshot(c.ctx, &empty.Empty{})
}
