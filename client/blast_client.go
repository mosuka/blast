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

type BlastClientWrapper struct {
	conn           *grpc.ClientConn
	client         proto.IndexClient
	requestTimeout int
}

func NewBlastClientWrapper(server string, requestTimeout int) (*BlastClientWrapper, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	ic := proto.NewIndexClient(conn)

	return &BlastClientWrapper{
		conn:           conn,
		client:         ic,
		requestTimeout: requestTimeout,
	}, nil
}

func (c *BlastClientWrapper) GetIndex(includeIndexMapping bool, includeIndexType bool, includeKvstore bool, includeKvconfig bool, opts ...grpc.CallOption) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	req := &proto.GetIndexRequest{
		IncludeIndexMapping: includeIndexMapping,
		IncludeIndexType:    includeIndexType,
		IncludeKvstore:      includeKvstore,
		IncludeKvconfig:     includeKvconfig,
	}

	resp, err := c.client.GetIndex(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		IndexPath    string                    `json:"index_path,omitempty"`
		IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType    string                    `json:"index_type,omitempty"`
		Kvstore      string                    `json:"kvstore,omitempty"`
		Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
	}{
		IndexPath:    resp.IndexPath,
		IndexMapping: resp.GetIndexMappingActual(),
		IndexType:    resp.IndexType,
		Kvstore:      resp.Kvstore,
		Kvconfig:     resp.GetKvconfigActual(),
	}

	return r, nil
}

func (c *BlastClientWrapper) PutDocument(id string, fields map[string]interface{}, opts ...grpc.CallOption) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	d := &proto.Document{
		Id: id,
	}
	d.SetFieldsActual(fields)

	req := &proto.PutDocumentRequest{
		Document: d,
	}

	resp, err := c.client.PutDocument(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		PutCount int32 `json:"put_count"`
	}{
		PutCount: resp.PutCount,
	}

	return r, nil
}

func (c *BlastClientWrapper) GetDocument(id string, opts ...grpc.CallOption) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	req := &proto.GetDocumentRequest{
		Id: id,
	}

	resp, err := c.client.GetDocument(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	d := struct {
		Id     string                 `json:"id,omitempty"`
		Fields map[string]interface{} `json:"fields,omitempty"`
	}{
		Id:     resp.Document.Id,
		Fields: resp.Document.GetFieldsActual(),
	}

	r := struct {
		Document interface{} `json:"document,omitempty"`
	}{
		Document: d,
	}

	return r, nil
}

func (c *BlastClientWrapper) DeleteDocument(id string, opts ...grpc.CallOption) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	req := &proto.DeleteDocumentRequest{
		Id: id,
	}

	resp, err := c.client.DeleteDocument(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		DeleteCount int32 `json:"delete_count"`
	}{
		DeleteCount: resp.DeleteCount,
	}

	return r, err
}

func (c *BlastClientWrapper) Bulk(requests []map[string]interface{}, batchSize int32, opts ...grpc.CallOption) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	rs := make([]*proto.UpdateRequest, 0)
	for _, request := range requests {
		r := &proto.UpdateRequest{}

		// check method
		if _, ok := request["method"]; ok {
			r.Method = request["method"].(string)
		}

		// check document
		var document map[string]interface{}
		if _, ok := request["document"]; ok {
			d := &proto.Document{}

			document = request["document"].(map[string]interface{})

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

		rs = append(rs, r)
	}

	req := &proto.BulkRequest{
		BatchSize: batchSize,
		Requests:  rs,
	}

	resp, err := c.client.Bulk(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		PutCount      int32 `json:"put_count"`
		PutErrorCount int32 `json:"put_error_count"`
		DeleteCount   int32 `json:"delete_count"`
	}{
		PutCount:      resp.PutCount,
		PutErrorCount: resp.PutErrorCount,
		DeleteCount:   resp.DeleteCount,
	}

	return r, nil
}

func (c *BlastClientWrapper) Search(searchRequests *bleve.SearchRequest, opts ...grpc.CallOption) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	sr, err := proto.MarshalAny(searchRequests)
	if err != nil {
		return nil, err
	}

	req := &proto.SearchRequest{
		SearchRequest: &sr,
	}

	resp, err := c.client.Search(ctx, req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		SearchResult *bleve.SearchResult `json:"search_result,omitempty"`
	}{
		SearchResult: resp.GetSearchResultActual(),
	}

	return r, err
}

func (c *BlastClientWrapper) Close() error {
	return c.conn.Close()
}
