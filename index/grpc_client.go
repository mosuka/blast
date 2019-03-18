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

package index

import (
	"context"
	"errors"
	"log"
	"math"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client index.IndexClient

	logger *log.Logger
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
		cancel()
		return nil, err
	}

	return &GRPCClient{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
		client: index.NewIndexClient(conn),
	}, nil
}

func (c *GRPCClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}

	return c.ctx.Err()
}

func (c *GRPCClient) Join(node *raft.Node, opts ...grpc.CallOption) error {
	req := &raft.JoinRequest{
		Node: node,
	}

	_, err := c.client.Join(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return errors.New(st.Message())
	}

	return nil
}

func (c *GRPCClient) Leave(node *raft.Node, opts ...grpc.CallOption) error {
	req := &raft.LeaveRequest{
		Node: node,
	}

	_, err := c.client.Leave(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return errors.New(st.Message())
	}

	return nil
}

func (c *GRPCClient) GetNode(opts ...grpc.CallOption) (*raft.Node, error) {
	resp, err := c.client.GetNode(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	return resp.Node, nil
}

func (c *GRPCClient) GetCluster(opts ...grpc.CallOption) (*raft.Cluster, error) {
	resp, err := c.client.GetCluster(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	return resp.Cluster, nil
}

func (c *GRPCClient) Snapshot(opts ...grpc.CallOption) error {
	_, err := c.client.Snapshot(c.ctx, &empty.Empty{})
	if err != nil {
		st, _ := status.FromError(err)

		return errors.New(st.Message())
	}

	return nil
}

func (c *GRPCClient) Get(doc *index.Document, opts ...grpc.CallOption) (*index.Document, error) {
	req := &index.GetRequest{
		Document: doc,
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

	return resp.Document, nil
}

func (c *GRPCClient) Search(searchRequest *bleve.SearchRequest, opts ...grpc.CallOption) (*bleve.SearchResult, error) {
	// bleve.SearchRequest -> Any
	searchRequestAny := &any.Any{}
	err := protobuf.UnmarshalAny(searchRequest, searchRequestAny)
	if err != nil {
		return nil, err
	}

	req := &index.SearchRequest{
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

func (c *GRPCClient) Index(doc *index.Document, opts ...grpc.CallOption) error {
	req := &index.IndexRequest{
		Document: doc,
	}

	_, err := c.client.Index(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return errors.New(st.Message())
	}

	return nil
}

func (c *GRPCClient) Delete(doc *index.Document, opts ...grpc.CallOption) error {
	req := &index.DeleteRequest{
		Document: doc,
	}

	_, err := c.client.Delete(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return errors.New(st.Message())
	}

	return nil
}

func (c *GRPCClient) GetIndexStats(opts ...grpc.CallOption) (*index.Stats, error) {
	resp, err := c.client.GetStats(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)

		return nil, errors.New(st.Message())
	}

	return resp.Stats, nil
}
