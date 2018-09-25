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
	client protobuf.DataClient
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
		client: protobuf.NewDataClient(conn),
	}, nil
}

func (c *GRPCClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}

	return c.ctx.Err()
}

func (c *GRPCClient) PutDocument(req *protobuf.PutDocumentRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.PutDocument(c.ctx, req, opts...)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *GRPCClient) GetDocument(req *protobuf.GetDocumentRequest, opts ...grpc.CallOption) (*protobuf.GetDocumentResponse, error) {
	resp, err := c.client.GetDocument(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) DeleteDocument(req *protobuf.DeleteDocumentRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.DeleteDocument(c.ctx, req, opts...)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (c *GRPCClient) BulkUpdate(req *protobuf.BulkUpdateRequest, opts ...grpc.CallOption) (*protobuf.BulkUpdateResponse, error) {
	resp, err := c.client.BulkUpdate(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) SearchDocuments(req *protobuf.SearchDocumentsRequest, opts ...grpc.CallOption) (*protobuf.SearchDocumentsResponse, error) {
	resp, err := c.client.SearchDocuments(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) PutNode(req *protobuf.PutNodeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.PutNode(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) GetNode(req *protobuf.GetNodeRequest, opts ...grpc.CallOption) (*protobuf.GetNodeResponse, error) {
	resp, err := c.client.GetNode(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) DeleteNode(req *protobuf.DeleteNodeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.DeleteNode(c.ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) GetCluster(opts ...grpc.CallOption) (*protobuf.GetClusterResponse, error) {
	resp, err := c.client.GetCluster(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *GRPCClient) Snapshot(opts ...grpc.CallOption) (*empty.Empty, error) {
	resp, err := c.client.Snapshot(c.ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
