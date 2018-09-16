// Copyright (c) 2018 Minoru Osuka
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

package client

import (
	"context"
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/node/data/protobuf"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client protobuf.DataNodeClient
}

func NewGRPCClient(address string) (*GRPCClient, error) {
	var err error

	// Connect context
	baseCtx := context.TODO()
	ctx, cancel := context.WithCancel(baseCtx)

	// Create dial options
	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(math.MaxInt32),
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
		),
	}

	// Create client connection
	var conn *grpc.ClientConn
	if conn, err = grpc.DialContext(ctx, address, dialOpts...); err != nil {
		cancel()
		return nil, err
	}

	return &GRPCClient{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		client: protobuf.NewDataNodeClient(conn),
	}, nil
}

func (c *GRPCClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}

	return c.ctx.Err()
}

func (c *GRPCClient) Get(req *protobuf.GetRequest, opts ...grpc.CallOption) (*protobuf.GetResponse, error) {
	var err error

	var resp *protobuf.GetResponse
	if resp, err = c.client.Get(c.ctx, req, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) Put(req *protobuf.PutRequest, opts ...grpc.CallOption) (*protobuf.PutResponse, error) {
	var err error

	var resp *protobuf.PutResponse
	if resp, err = c.client.Put(c.ctx, req, opts...); err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *GRPCClient) Delete(req *protobuf.DeleteRequest, opts ...grpc.CallOption) (*protobuf.DeleteResponse, error) {
	var err error

	var resp *protobuf.DeleteResponse
	if resp, err = c.client.Delete(c.ctx, req, opts...); err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *GRPCClient) Bulk(req *protobuf.BulkRequest, opts ...grpc.CallOption) (*protobuf.BulkResponse, error) {
	var err error

	var resp *protobuf.BulkResponse
	if resp, err = c.client.Bulk(c.ctx, req, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) Search(req *protobuf.SearchRequest, opts ...grpc.CallOption) (*protobuf.SearchResponse, error) {
	var err error

	var resp *protobuf.SearchResponse
	if resp, err = c.client.Search(c.ctx, req, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) Join(req *protobuf.JoinRequest, opts ...grpc.CallOption) (*protobuf.JoinResponse, error) {
	var err error

	var resp *protobuf.JoinResponse
	if resp, err = c.client.Join(c.ctx, req, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) Leave(req *protobuf.LeaveRequest, opts ...grpc.CallOption) (*protobuf.LeaveResponse, error) {
	var err error

	var resp *protobuf.LeaveResponse
	if resp, err = c.client.Leave(c.ctx, req, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) Peers(opts ...grpc.CallOption) (*protobuf.PeersResponse, error) {
	var err error

	var resp *protobuf.PeersResponse
	if resp, err = c.client.Peers(c.ctx, &empty.Empty{}, opts...); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) Snapshot(opts ...grpc.CallOption) (*protobuf.SnapshotResponse, error) {
	var err error

	var resp *protobuf.SnapshotResponse
	if resp, err = c.client.Snapshot(c.ctx, &empty.Empty{}); err != nil {
		return nil, err
	}

	return resp, nil
}
