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
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Index interface {
	GetIndexInfo(ctx context.Context, indexPath bool, indexMapping bool, indexType bool, kvstore bool, kvconfig bool, opts ...grpc.CallOption) (*GetIndexInfoResponse, error)
	PutDocument(ctx context.Context, id string, fields map[string]interface{}, opts ...grpc.CallOption) (*PutDocumentResponse, error)
	GetDocument(ctx context.Context, id string, opts ...grpc.CallOption) (*GetDocumentResponse, error)
	DeleteDocument(ctx context.Context, id string, opts ...grpc.CallOption) (*DeleteDocumentResponse, error)
	Bulk(ctx context.Context, requests []map[string]interface{}, batchSize int32, opts ...grpc.CallOption) (*BulkResponse, error)
	Search(ctx context.Context, searchRequest *bleve.SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error)
}

type index struct {
	client proto.IndexClient
}

func NewIndex(c *Client) Index {
	ic := proto.NewIndexClient(c.conn)

	return &index{
		client: ic,
	}
}

func (i *index) GetIndexInfo(ctx context.Context, indexPath bool, indexMapping bool, indexType bool, kvstore bool, kvconfig bool, opts ...grpc.CallOption) (*GetIndexInfoResponse, error) {
	protoReq := &proto.GetIndexInfoRequest{
		IndexMapping: indexMapping,
		IndexType:    indexType,
		Kvstore:      kvstore,
		Kvconfig:     kvconfig,
	}

	protoResp, _ := i.client.GetIndexInfo(ctx, protoReq, opts...)

	im, err := proto.UnmarshalAny(protoResp.IndexMapping)
	if err != nil {
		return nil, err
	}

	kvc, err := proto.UnmarshalAny(protoResp.Kvconfig)
	if err != nil {
		return nil, err
	}

	return &GetIndexInfoResponse{
		IndexPath:    protoResp.IndexPath,
		IndexMapping: im.(*mapping.IndexMappingImpl),
		IndexType:    protoResp.IndexType,
		Kvstore:      protoResp.Kvstore,
		Kvconfig:     *kvc.(*map[string]interface{}),
	}, nil
}

func (i *index) PutDocument(ctx context.Context, id string, fields map[string]interface{}, opts ...grpc.CallOption) (*PutDocumentResponse, error) {
	fieldAny, err := proto.MarshalAny(fields)
	if err != nil {
		return nil, err
	}

	protoReq := &proto.PutDocumentRequest{
		Id:     id,
		Fields: &fieldAny,
	}

	protoResp, _ := i.client.PutDocument(ctx, protoReq, opts...)

	fieldsTmp, err := proto.UnmarshalAny(protoResp.Fields)
	if err != nil {
		return nil, err
	}

	var fieldsPut map[string]interface{}
	if fieldsTmp != nil {
		fieldsPut = *fieldsTmp.(*map[string]interface{})
	}

	return &PutDocumentResponse{
		Id:        id,
		Fields:    fieldsPut,
		Succeeded: protoResp.Succeeded,
		Message:   protoResp.Message,
	}, nil
}

func (i *index) GetDocument(ctx context.Context, id string, opts ...grpc.CallOption) (*GetDocumentResponse, error) {
	protoReq := &proto.GetDocumentRequest{
		Id: id,
	}

	protoResp, _ := i.client.GetDocument(ctx, protoReq, opts...)

	fieldsTmp, err := proto.UnmarshalAny(protoResp.Fields)
	if err != nil {
		return nil, err
	}

	var fields map[string]interface{}
	if fieldsTmp != nil {
		fields = *fieldsTmp.(*map[string]interface{})
	}

	return &GetDocumentResponse{
		Id:        id,
		Fields:    fields,
		Succeeded: protoResp.Succeeded,
		Message:   protoResp.Message,
	}, nil
}

func (i *index) DeleteDocument(ctx context.Context, id string, opts ...grpc.CallOption) (*DeleteDocumentResponse, error) {
	protoReq := &proto.DeleteDocumentRequest{
		Id: id,
	}

	protoResp, _ := i.client.DeleteDocument(ctx, protoReq, opts...)

	return &DeleteDocumentResponse{
		Id:        id,
		Succeeded: protoResp.Succeeded,
		Message:   protoResp.Message,
	}, nil
}

func (i *index) Bulk(ctx context.Context, requests []map[string]interface{}, batchSize int32, opts ...grpc.CallOption) (*BulkResponse, error) {
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

	protoResp, err := i.client.Bulk(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	return &BulkResponse{
		PutCount:      protoResp.PutCount,
		PutErrorCount: protoResp.PutErrorCount,
		DeleteCount:   protoResp.DeleteCount,
		Succeeded:     protoResp.Succeeded,
		Message:       protoResp.Message,
	}, nil
}

func (i *index) Search(ctx context.Context, searchRequest *bleve.SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	searchResultAny, err := proto.MarshalAny(searchRequest)
	if err != nil {
		return nil, err
	}

	protoReq := &proto.SearchRequest{
		SearchRequest: &searchResultAny,
	}

	protoResp, err := i.client.Search(ctx, protoReq, opts...)
	if err != nil {
		return nil, err
	}

	searchResult, err := proto.UnmarshalAny(protoResp.SearchResult)

	return &SearchResponse{
		SearchResult: searchResult.(*bleve.SearchResult),
		Succeeded:    protoResp.Succeeded,
		Message:      protoResp.Message,
	}, err
}
