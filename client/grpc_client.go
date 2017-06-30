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
)

type BlastClientWrapper struct {
	Conn   *grpc.ClientConn
	Client proto.BlastClient
}

func NewBlastClientWrapper(server string) (*BlastClientWrapper, error) {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return &BlastClientWrapper{}, err
	}

	ic := proto.NewBlastClient(conn)

	return &BlastClientWrapper{
		Conn:   conn,
		Client: ic,
	}, nil
}

func (c *BlastClientWrapper) GetIndex(ctx context.Context, includeIndexMapping bool, includeIndexType bool, includeKvstore bool, includeKvconfig bool, opts ...grpc.CallOption) (interface{}, error) {
	req := &proto.GetIndexRequest{
		IncludeIndexMapping: includeIndexMapping,
		IncludeIndexType:    includeIndexType,
		IncludeKvstore:      includeKvstore,
		IncludeKvconfig:     includeKvconfig,
	}

	resp, err := c.Client.GetIndex(context.Background(), req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		Path         string                    `json:"path,omitempty"`
		IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType    string                    `json:"index_type,omitempty"`
		Kvstore      string                    `json:"kvstore,omitempty"`
		Kvconfig     interface{}               `json:"kvconfig,omitempty"`
	}{
		Path:         resp.Path,
		IndexMapping: resp.GetIndexMappingActual(),
		IndexType:    resp.IndexType,
		Kvstore:      resp.Kvstore,
		Kvconfig:     resp.GetKvconfigActual(),
	}

	return r, nil
}

func (c *BlastClientWrapper) PutDocument(ctx context.Context, id string, fields map[string]interface{}, opts ...grpc.CallOption) (interface{}, error) {
	d := &proto.Document{
		Id: id,
	}
	d.SetFieldsActual(fields)

	req := &proto.PutDocumentRequest{
		Document: d,
	}

	resp, err := c.Client.PutDocument(context.Background(), req, opts...)
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

func (c *BlastClientWrapper) GetDocument(ctx context.Context, id string, opts ...grpc.CallOption) (interface{}, error) {
	req := &proto.GetDocumentRequest{
		Id: id,
	}

	resp, err := c.Client.GetDocument(context.Background(), req, opts...)
	if err != nil {
		return nil, err
	}

	d := struct {
		Id     string                 `json:"id"`
		Fields map[string]interface{} `json:"fields"`
	}{
		Id:     resp.Document.Id,
		Fields: resp.Document.GetFieldsActual(),
	}

	r := struct {
		Document interface{} `json:"document"`
	}{
		Document: d,
	}

	return r, err
}

func (c *BlastClientWrapper) DeleteDocument(ctx context.Context, id string, opts ...grpc.CallOption) (interface{}, error) {
	req := &proto.DeleteDocumentRequest{
		Id: id,
	}

	resp, err := c.Client.DeleteDocument(context.Background(), req, opts...)
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

func (c *BlastClientWrapper) Bulk(ctx context.Context, requests []map[string]interface{}, batchSize int32, opts ...grpc.CallOption) (interface{}, error) {
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

	resp, err := c.Client.Bulk(context.Background(), req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		PutCount      int32 `json:"put_count,omitempty"`
		PutErrorCount int32 `json:"put_error_count,omitempty"`
		DeleteCount   int32 `json:"delete_count,omitempty"`
	}{
		PutCount:      resp.PutCount,
		PutErrorCount: resp.PutErrorCount,
		DeleteCount:   resp.DeleteCount,
	}

	return r, nil
}

func (c *BlastClientWrapper) Search(ctx context.Context, seardhRequests *bleve.SearchRequest, opts ...grpc.CallOption) (interface{}, error) {
	sr, err := proto.MarshalAny(seardhRequests)
	if err != nil {
		return nil, err
	}

	req := &proto.SearchRequest{
		SearchRequest: &sr,
	}

	resp, err := c.Client.Search(context.Background(), req, opts...)
	if err != nil {
		return nil, err
	}

	r := struct {
		SearchResult *bleve.SearchResult `json:"search_result"`
	}{
		SearchResult: resp.GetSearchResultActual(),
	}

	return r, err
}

func (c *BlastClientWrapper) Close() error {
	return c.Conn.Close()
}
