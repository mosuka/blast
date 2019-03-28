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
	"log"
	"math"

	"github.com/golang/protobuf/ptypes/empty"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf/management"
	"github.com/mosuka/blast/protobuf/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client management.ManagementClient

	logger *log.Logger
}

func NewGRPCClient(address string) (*GRPCClient, error) {
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

	return &GRPCClient{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		client: management.NewManagementClient(conn),
	}, nil
}

func (c *GRPCClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}

	return c.ctx.Err()
}

func (c *GRPCClient) Join(req *raft.Node, opts ...grpc.CallOption) error {
	_, err := c.client.Join(c.ctx, req, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) Leave(req *raft.Node, opts ...grpc.CallOption) error {
	_, err := c.client.Leave(c.ctx, req, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) GetNode(opts ...grpc.CallOption) (*raft.Node, error) {
	node, err := c.client.GetNode(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	return node, nil
}

func (c *GRPCClient) GetCluster(opts ...grpc.CallOption) (*raft.Cluster, error) {
	cluster, err := c.client.GetCluster(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	return cluster, nil
}

func (c *GRPCClient) Snapshot(opts ...grpc.CallOption) error {
	_, err := c.client.Snapshot(c.ctx, &empty.Empty{})
	if err != nil {
		st, _ := status.FromError(err)

		return errors.New(st.Message())
	}

	return nil
}

func (c *GRPCClient) Get(req *management.KeyValuePair, opts ...grpc.CallOption) (*management.KeyValuePair, error) {
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

	return resp, nil
}

func (c *GRPCClient) Set(req *management.KeyValuePair, opts ...grpc.CallOption) error {
	_, err := c.client.Set(c.ctx, req, opts...)
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

func (c *GRPCClient) Delete(req *management.KeyValuePair, opts ...grpc.CallOption) error {
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
