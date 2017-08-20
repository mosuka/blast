//  Copyright (c) 2017 Minoru Osuka
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
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/proto"
	"google.golang.org/grpc"
	"time"
)

type BlastClient struct {
	conn           *grpc.ClientConn
	client         proto.IndexClient
	requestTimeout int
}

func NewBlastClient(server string, dialTimeout int, requestTimeout int) (*BlastClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dialTimeout)*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	ic := proto.NewIndexClient(conn)

	return &BlastClient{
		conn:           conn,
		client:         ic,
		requestTimeout: requestTimeout,
	}, nil
}

func (c *BlastClient) GetIndex(indexMapping bool, indexType bool, kvstore bool, kvconfig bool, opts ...grpc.CallOption) (*GetIndexResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	protoReq := &proto.GetIndexRequest{
		IndexMapping: indexMapping,
		IndexType:    indexType,
		Kvstore:      kvstore,
		Kvconfig:     kvconfig,
	}

	protoResp, _ := c.client.GetIndex(ctx, protoReq, opts...)

	im, err := proto.UnmarshalAny(protoResp.IndexMapping)
	if err != nil {
		return nil, err
	}

	kvc, err := proto.UnmarshalAny(protoResp.Kvconfig)
	if err != nil {
		return nil, err
	}

	resp := &GetIndexResponse{
		IndexPath:    protoResp.IndexPath,
		IndexMapping: im.(*mapping.IndexMappingImpl),
		IndexType:    protoResp.IndexType,
		Kvstore:      protoResp.Kvstore,
		Kvconfig:     *kvc.(*map[string]interface{}),
	}

	return resp, nil
}

func (c *BlastClient) PutDocument(id string, fields map[string]interface{}, opts ...grpc.CallOption) (*PutDocumentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	fieldAny, err := proto.MarshalAny(fields)
	if err != nil {
		return nil, err
	}

	protoReq := &proto.PutDocumentRequest{
		Id:     id,
		Fields: &fieldAny,
	}

	protoResp, _ := c.client.PutDocument(ctx, protoReq, opts...)

	fieldsTmp, err := proto.UnmarshalAny(protoResp.Fields)
	if err != nil {
		return nil, err
	}

	var fieldsPut map[string]interface{}
	if fieldsTmp != nil {
		fieldsPut = *fieldsTmp.(*map[string]interface{})
	}

	resp := &PutDocumentResponse{
		Id:        id,
		Fields:    fieldsPut,
		Succeeded: protoResp.Succeeded,
		Message:   protoResp.Message,
	}

	return resp, nil
}

func (c *BlastClient) GetDocument(id string, opts ...grpc.CallOption) (*GetDocumentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	protoReq := &proto.GetDocumentRequest{
		Id: id,
	}

	protoResp, _ := c.client.GetDocument(ctx, protoReq, opts...)

	fieldsTmp, err := proto.UnmarshalAny(protoResp.Fields)
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	if fieldsTmp != nil {
		fields = *fieldsTmp.(*map[string]interface{})
	}

	resp := &GetDocumentResponse{
		Id:        id,
		Fields:    fields,
		Succeeded: protoResp.Succeeded,
		Message:   protoResp.Message,
	}

	return resp, nil
}

func (c *BlastClient) DeleteDocument(id string, opts ...grpc.CallOption) (*DeleteDocumentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	protoReq := &proto.DeleteDocumentRequest{
		Id: id,
	}

	protoResp, _ := c.client.DeleteDocument(ctx, protoReq, opts...)

	resp := &DeleteDocumentResponse{
		Id:        id,
		Succeeded: protoResp.Succeeded,
		Message:   protoResp.Message,
	}

	return resp, nil
}

func (c *BlastClient) Bulk(requests []map[string]interface{}, batchSize int32, opts ...grpc.CallOption) (*BulkResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	updateRequests := make([]*proto.BulkRequest_UpdateRequest, 0)
	for _, updateRequest := range requests {
		r := &proto.BulkRequest_UpdateRequest{}

		// check method
		if _, ok := updateRequest["method"]; ok {
			r.Method = updateRequest["method"].(string)
		}

		// check document
		var document map[string]interface{}
		if _, ok := updateRequest["document"]; ok {
			d := &proto.BulkRequest_UpdateRequest_Document{}

			document = updateRequest["document"].(map[string]interface{})

			// check document.id
			if _, ok := document["id"]; ok {
				d.Id = document["id"].(string)
			}

			// check document.fields
			if _, ok := document["fields"]; ok {
				fields, err := proto.MarshalAny(document["fields"].(map[string]interface{}))
				if err != nil {
					return nil, err
				}
				d.Fields = &fields
			}

			r.Document = d
		}

		updateRequests = append(updateRequests, r)
	}

	protoReq := &proto.BulkRequest{
		BatchSize:      batchSize,
		UpdateRequests: updateRequests,
	}

	protoResp, err := c.client.Bulk(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	resp := &BulkResponse{
		PutCount:      protoResp.PutCount,
		PutErrorCount: protoResp.PutErrorCount,
		DeleteCount:   protoResp.DeleteCount,
		Succeeded:     protoResp.Succeeded,
		Message:       protoResp.Message,
	}

	return resp, nil
}

func (c *BlastClient) Search(searchRequest *bleve.SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	searchResultAny, err := proto.MarshalAny(searchRequest)
	if err != nil {
		return nil, err
	}

	protoReq := &proto.SearchRequest{
		SearchRequest: &searchResultAny,
	}

	protoResp, err := c.client.Search(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	searchResult, err := proto.UnmarshalAny(protoResp.SearchResult)

	resp := &SearchResponse{
		SearchResult: searchResult.(*bleve.SearchResult),
		Succeeded:    protoResp.Succeeded,
		Message:      protoResp.Message,
	}

	return resp, err
}

func (c *BlastClient) Close() error {
	return c.conn.Close()
}
