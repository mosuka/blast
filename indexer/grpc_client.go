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
	"errors"
	"math"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (c *GRPCClient) NodeHealthCheck(probe string, opts ...grpc.CallOption) (string, error) {
	req := &index.NodeHealthCheckRequest{}

	switch probe {
	case index.NodeHealthCheckRequest_HEALTHINESS.String():
		req.Probe = index.NodeHealthCheckRequest_HEALTHINESS
	case index.NodeHealthCheckRequest_LIVENESS.String():
		req.Probe = index.NodeHealthCheckRequest_LIVENESS
	case index.NodeHealthCheckRequest_READINESS.String():
		req.Probe = index.NodeHealthCheckRequest_READINESS
	default:
		req.Probe = index.NodeHealthCheckRequest_HEALTHINESS
	}

	resp, err := c.client.NodeHealthCheck(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return index.NodeHealthCheckResponse_UNHEALTHY.String(), errors.New(st.Message())
	}

	return resp.State.String(), nil
}

func (c *GRPCClient) NodeInfo(opts ...grpc.CallOption) (*index.Node, error) {
	resp, err := c.client.NodeInfo(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	return resp.Node, nil
}

func (c *GRPCClient) ClusterJoin(node *index.Node, opts ...grpc.CallOption) error {
	req := &index.ClusterJoinRequest{
		Node: node,
	}

	_, err := c.client.ClusterJoin(c.ctx, req, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) ClusterLeave(id string, opts ...grpc.CallOption) error {
	req := &index.ClusterLeaveRequest{
		Id: id,
	}

	_, err := c.client.ClusterLeave(c.ctx, req, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) ClusterInfo(opts ...grpc.CallOption) (*index.Cluster, error) {
	resp, err := c.client.ClusterInfo(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	return resp.Cluster, nil
}

func (c *GRPCClient) ClusterWatch(opts ...grpc.CallOption) (index.Index_ClusterWatchClient, error) {
	req := &empty.Empty{}

	watchClient, err := c.client.ClusterWatch(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	return watchClient, nil
}

func (c *GRPCClient) Get(id string, opts ...grpc.CallOption) (*index.Document, error) {
	req := &index.GetRequest{
		Id: id,
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

func (c *GRPCClient) Delete(id string, opts ...grpc.CallOption) error {
	req := &index.DeleteRequest{
		Id: id,
	}

	_, err := c.client.Delete(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return errors.New(st.Message())
	}

	return nil
}

func (c *GRPCClient) BulkIndex(docs []*index.Document, opts ...grpc.CallOption) (int, error) {
	req := &index.BulkIndexRequest{
		Documents: docs,
	}

	resp, err := c.client.BulkIndex(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return -1, errors.New(st.Message())
	}

	return int(resp.Count), nil
}

func (c *GRPCClient) BulkDelete(ids []string, opts ...grpc.CallOption) (int, error) {
	req := &index.BulkDeleteRequest{
		Ids: ids,
	}

	resp, err := c.client.BulkDelete(c.ctx, req, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return -1, errors.New(st.Message())
	}

	return int(resp.Count), nil
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

func (c *GRPCClient) GetIndexConfig(opts ...grpc.CallOption) (map[string]interface{}, error) {
	resp, err := c.client.GetIndexConfig(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	indexMapping, err := protobuf.MarshalAny(resp.IndexConfig.IndexMapping)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	indexConfig := map[string]interface{}{
		"index_mapping":      indexMapping,
		"index_type":         resp.IndexConfig.IndexType,
		"index_storage_type": resp.IndexConfig.IndexStorageType,
	}

	return indexConfig, nil
}

func (c *GRPCClient) GetIndexStats(opts ...grpc.CallOption) (map[string]interface{}, error) {
	resp, err := c.client.GetIndexStats(c.ctx, &empty.Empty{}, opts...)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}

	indexStatsIntr, err := protobuf.MarshalAny(resp.IndexStats)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, errors.New(st.Message())
	}
	indexStats := *indexStatsIntr.(*map[string]interface{})

	return indexStats, nil
}

func (c *GRPCClient) Snapshot(opts ...grpc.CallOption) error {
	_, err := c.client.Snapshot(c.ctx, &empty.Empty{})
	if err != nil {
		st, _ := status.FromError(err)
		return errors.New(st.Message())
	}

	return nil
}
