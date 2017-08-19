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

func (c *BlastClient) GetIndex(includeIndexMapping bool, includeIndexType bool, includeKvstore bool, includeKvconfig bool, opts ...grpc.CallOption) (*GetIndexResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	protoReq := &proto.GetIndexRequest{
		IncludeIndexMapping: includeIndexMapping,
		IncludeIndexType:    includeIndexType,
		IncludeKvstore:      includeKvstore,
		IncludeKvconfig:     includeKvconfig,
	}

	protoResp, err := c.client.GetIndex(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	resp := &GetIndexResponse{
		IndexPath:    protoResp.IndexPath,
		IndexMapping: protoResp.GetIndexMappingActual(),
		IndexType:    protoResp.IndexType,
		Kvstore:      protoResp.Kvstore,
		Kvconfig:     protoResp.GetKvconfigActual(),
	}

	return resp, nil
}

func (c *BlastClient) PutDocument(id string, fields map[string]interface{}, opts ...grpc.CallOption) (*PutDocumentResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	doc := &proto.Document{}
	doc.Id = id
	doc.SetFieldsActual(fields)

	protoReq := &proto.PutDocumentRequest{
		Document: doc,
	}

	protoResp, err := c.client.PutDocument(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	resp := &PutDocumentResponse{
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

	var doc *Document
	if protoResp.Document != nil {
		doc = &Document{
			Id:     protoResp.Document.Id,
			Fields: protoResp.Document.GetFieldsActual(),
		}
	}

	resp := &GetDocumentResponse{
		Succeeded: protoResp.Succeeded,
		Document:  doc,
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

	protoResp, err := c.client.DeleteDocument(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	resp := &DeleteDocumentResponse{
		Succeeded: protoResp.Succeeded,
		Message:   protoResp.Message,
	}

	return resp, err
}

func (c *BlastClient) Bulk(requests []map[string]interface{}, batchSize int32, opts ...grpc.CallOption) (*BulkResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	updateRequests := make([]*proto.UpdateRequest, 0)
	for _, updateRequest := range requests {
		r := &proto.UpdateRequest{}

		// check method
		if _, ok := updateRequest["method"]; ok {
			r.Method = updateRequest["method"].(string)
		}

		// check document
		var document map[string]interface{}
		if _, ok := updateRequest["document"]; ok {
			d := &proto.Document{}

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

func (c *BlastClient) Search(searchRequests *bleve.SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	sr, err := proto.MarshalAny(searchRequests)
	if err != nil {
		return nil, err
	}

	protoReq := &proto.SearchRequest{
		SearchRequest: &sr,
	}

	protoResp, err := c.client.Search(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	resp := &SearchResponse{
		SearchResult: protoResp.GetSearchResultActual(),
		Succeeded:    protoResp.Succeeded,
		Message:      protoResp.Message,
	}

	return resp, err
}

func (c *BlastClient) Close() error {
	return c.conn.Close()
}
