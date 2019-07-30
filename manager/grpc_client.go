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
	"errors"
	"math"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
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

func (c *GRPCClient) NodeHealthCheck(probe string, opts ...grpc.CallOption) (string, error) {
	req := &management.NodeHealthCheckRequest{}

	switch probe {
	case management.NodeHealthCheckRequest_HEALTHINESS.String():
		req.Probe = management.NodeHealthCheckRequest_HEALTHINESS
	case management.NodeHealthCheckRequest_LIVENESS.String():
		req.Probe = management.NodeHealthCheckRequest_LIVENESS
	case management.NodeHealthCheckRequest_READINESS.String():
		req.Probe = management.NodeHealthCheckRequest_READINESS
	default:
		req.Probe = management.NodeHealthCheckRequest_HEALTHINESS
	}

	resp, err := c.client.NodeHealthCheck(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return management.NodeHealthCheckResponse_UNHEALTHY.String(), errors.New(st.Message())
	}

	return resp.State.String(), nil
}

func (c *GRPCClient) NodeInfo(opts ...grpc.CallOption) (*management.Node, error) {
	resp, err := c.client.NodeInfo(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	return resp.Node, nil
}

func (c *GRPCClient) ClusterJoin(id string, node *management.Node, opts ...grpc.CallOption) error {
	req := &management.ClusterJoinRequest{
		Id:   id,
		Node: node,
	}

	_, err := c.client.ClusterJoin(c.ctx, req, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) ClusterLeave(id string, opts ...grpc.CallOption) error {
	req := &management.ClusterLeaveRequest{
		Id: id,
	}

	_, err := c.client.ClusterLeave(c.ctx, req, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) ClusterInfo(opts ...grpc.CallOption) (*management.Cluster, error) {
	resp, err := c.client.ClusterInfo(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	return resp.Cluster, nil
}

func (c *GRPCClient) ClusterWatch(opts ...grpc.CallOption) (management.Management_ClusterWatchClient, error) {
	req := &empty.Empty{}

	watchClient, err := c.client.ClusterWatch(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	return watchClient, nil
}

func (c *GRPCClient) Get(key string, opts ...grpc.CallOption) (interface{}, error) {
	req := &management.GetRequest{
		Key: key,
	}

	resp, err := c.client.Get(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		switch st.Code() {
		case codes.NotFound:
			return nil, blasterrors.ErrNotFound
		default:
			return nil, errors.New(st.Message())
		}
	}

	value, err := protobuf.MarshalAny(resp.Value)

	return value, nil
}

func (c *GRPCClient) Set(key string, value interface{}, opts ...grpc.CallOption) error {
	valueAny := &any.Any{}
	err := protobuf.UnmarshalAny(value, valueAny)
	if err != nil {
		return err
	}

	req := &management.SetRequest{
		Key:   key,
		Value: valueAny,
	}

	_, err = c.client.Set(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		switch st.Code() {
		case codes.NotFound:
			return blasterrors.ErrNotFound
		default:
			return errors.New(st.Message())
		}
	}

	return nil
}

func (c *GRPCClient) Delete(key string, opts ...grpc.CallOption) error {
	req := &management.DeleteRequest{
		Key: key,
	}

	_, err := c.client.Delete(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		switch st.Code() {
		case codes.NotFound:
			return blasterrors.ErrNotFound
		default:
			return errors.New(st.Message())
		}
	}

	return nil
}

func (c *GRPCClient) Watch(key string, opts ...grpc.CallOption) (management.Management_WatchClient, error) {
	req := &management.WatchRequest{
		Key: key,
	}

	watchClient, err := c.client.Watch(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	return watchClient, nil
}

func (c *GRPCClient) Snapshot(opts ...grpc.CallOption) error {
	_, err := c.client.Snapshot(c.ctx, &empty.Empty{})
	if err != nil {
		st, _ := status.FromError(err)

		return errors.New(st.Message())
	}

	return nil
}
