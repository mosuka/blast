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

package kvs

import (
	"context"
	"log"
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mosuka/blast/protobuf/kvs"
	"github.com/mosuka/blast/protobuf/raft"
	"google.golang.org/grpc"
)

type Client struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client kvs.KVSClient

	logger *log.Logger
}

func NewClient(address string) (*Client, error) {
	baseCtx := context.TODO()
	ctx, cancel := context.WithCancel(baseCtx)

	dialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(math.MaxInt32),
			grpc.MaxCallRecvMsgSize(math.MaxInt32),
		),
	}

	conn, err := grpc.DialContext(ctx, address, dialOpts...)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Client{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		client: kvs.NewKVSClient(conn),
	}, nil
}

func (c *Client) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}

	return c.ctx.Err()
}

func (c *Client) Join(req *raft.Node, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.Join(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Leave(req *raft.Node, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.Leave(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Snapshot(opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.Snapshot(c.ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Get(req *kvs.GetRequest, opts ...grpc.CallOption) (*kvs.GetResponse, error) {
	resp, err := c.client.Get(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) Put(req *kvs.PutRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.Put(c.ctx, req, opts...)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *Client) Delete(req *kvs.DeleteRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.Delete(c.ctx, req, opts...)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
