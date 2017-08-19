// Code generated by protoc-gen-go.
// source: proto/index_service.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	proto/index_service.proto

It has these top-level messages:
	Document
	UpdateRequest
	GetIndexRequest
	GetIndexResponse
	PutDocumentRequest
	PutDocumentResponse
	GetDocumentRequest
	GetDocumentResponse
	DeleteDocumentRequest
	DeleteDocumentResponse
	BulkRequest
	BulkResponse
	SearchRequest
	SearchResponse
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type Document struct {
	Id     string               `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Fields *google_protobuf.Any `protobuf:"bytes,2,opt,name=fields" json:"fields,omitempty"`
}

func (m *Document) Reset()                    { *m = Document{} }
func (m *Document) String() string            { return proto1.CompactTextString(m) }
func (*Document) ProtoMessage()               {}
func (*Document) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Document) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Document) GetFields() *google_protobuf.Any {
	if m != nil {
		return m.Fields
	}
	return nil
}

type UpdateRequest struct {
	Method   string    `protobuf:"bytes,1,opt,name=method" json:"method,omitempty"`
	Document *Document `protobuf:"bytes,2,opt,name=document" json:"document,omitempty"`
}

func (m *UpdateRequest) Reset()                    { *m = UpdateRequest{} }
func (m *UpdateRequest) String() string            { return proto1.CompactTextString(m) }
func (*UpdateRequest) ProtoMessage()               {}
func (*UpdateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *UpdateRequest) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *UpdateRequest) GetDocument() *Document {
	if m != nil {
		return m.Document
	}
	return nil
}

type GetIndexRequest struct {
	IncludeIndexMapping bool `protobuf:"varint,1,opt,name=include_index_mapping,json=includeIndexMapping" json:"include_index_mapping,omitempty"`
	IncludeIndexType    bool `protobuf:"varint,2,opt,name=include_index_type,json=includeIndexType" json:"include_index_type,omitempty"`
	IncludeKvstore      bool `protobuf:"varint,3,opt,name=include_kvstore,json=includeKvstore" json:"include_kvstore,omitempty"`
	IncludeKvconfig     bool `protobuf:"varint,4,opt,name=include_kvconfig,json=includeKvconfig" json:"include_kvconfig,omitempty"`
}

func (m *GetIndexRequest) Reset()                    { *m = GetIndexRequest{} }
func (m *GetIndexRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetIndexRequest) ProtoMessage()               {}
func (*GetIndexRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *GetIndexRequest) GetIncludeIndexMapping() bool {
	if m != nil {
		return m.IncludeIndexMapping
	}
	return false
}

func (m *GetIndexRequest) GetIncludeIndexType() bool {
	if m != nil {
		return m.IncludeIndexType
	}
	return false
}

func (m *GetIndexRequest) GetIncludeKvstore() bool {
	if m != nil {
		return m.IncludeKvstore
	}
	return false
}

func (m *GetIndexRequest) GetIncludeKvconfig() bool {
	if m != nil {
		return m.IncludeKvconfig
	}
	return false
}

type GetIndexResponse struct {
	IndexPath    string               `protobuf:"bytes,1,opt,name=index_path,json=indexPath" json:"index_path,omitempty"`
	IndexMapping *google_protobuf.Any `protobuf:"bytes,2,opt,name=index_mapping,json=indexMapping" json:"index_mapping,omitempty"`
	IndexType    string               `protobuf:"bytes,3,opt,name=index_type,json=indexType" json:"index_type,omitempty"`
	Kvstore      string               `protobuf:"bytes,4,opt,name=kvstore" json:"kvstore,omitempty"`
	Kvconfig     *google_protobuf.Any `protobuf:"bytes,5,opt,name=kvconfig" json:"kvconfig,omitempty"`
	Succeeded    bool                 `protobuf:"varint,6,opt,name=succeeded" json:"succeeded,omitempty"`
	Message      string               `protobuf:"bytes,7,opt,name=message" json:"message,omitempty"`
}

func (m *GetIndexResponse) Reset()                    { *m = GetIndexResponse{} }
func (m *GetIndexResponse) String() string            { return proto1.CompactTextString(m) }
func (*GetIndexResponse) ProtoMessage()               {}
func (*GetIndexResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *GetIndexResponse) GetIndexPath() string {
	if m != nil {
		return m.IndexPath
	}
	return ""
}

func (m *GetIndexResponse) GetIndexMapping() *google_protobuf.Any {
	if m != nil {
		return m.IndexMapping
	}
	return nil
}

func (m *GetIndexResponse) GetIndexType() string {
	if m != nil {
		return m.IndexType
	}
	return ""
}

func (m *GetIndexResponse) GetKvstore() string {
	if m != nil {
		return m.Kvstore
	}
	return ""
}

func (m *GetIndexResponse) GetKvconfig() *google_protobuf.Any {
	if m != nil {
		return m.Kvconfig
	}
	return nil
}

func (m *GetIndexResponse) GetSucceeded() bool {
	if m != nil {
		return m.Succeeded
	}
	return false
}

func (m *GetIndexResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type PutDocumentRequest struct {
	Document *Document `protobuf:"bytes,1,opt,name=document" json:"document,omitempty"`
}

func (m *PutDocumentRequest) Reset()                    { *m = PutDocumentRequest{} }
func (m *PutDocumentRequest) String() string            { return proto1.CompactTextString(m) }
func (*PutDocumentRequest) ProtoMessage()               {}
func (*PutDocumentRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PutDocumentRequest) GetDocument() *Document {
	if m != nil {
		return m.Document
	}
	return nil
}

type PutDocumentResponse struct {
	Succeeded bool   `protobuf:"varint,1,opt,name=succeeded" json:"succeeded,omitempty"`
	Message   string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *PutDocumentResponse) Reset()                    { *m = PutDocumentResponse{} }
func (m *PutDocumentResponse) String() string            { return proto1.CompactTextString(m) }
func (*PutDocumentResponse) ProtoMessage()               {}
func (*PutDocumentResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *PutDocumentResponse) GetSucceeded() bool {
	if m != nil {
		return m.Succeeded
	}
	return false
}

func (m *PutDocumentResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type GetDocumentRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *GetDocumentRequest) Reset()                    { *m = GetDocumentRequest{} }
func (m *GetDocumentRequest) String() string            { return proto1.CompactTextString(m) }
func (*GetDocumentRequest) ProtoMessage()               {}
func (*GetDocumentRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *GetDocumentRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type GetDocumentResponse struct {
	Document  *Document `protobuf:"bytes,1,opt,name=document" json:"document,omitempty"`
	Succeeded bool      `protobuf:"varint,2,opt,name=succeeded" json:"succeeded,omitempty"`
	Message   string    `protobuf:"bytes,3,opt,name=message" json:"message,omitempty"`
}

func (m *GetDocumentResponse) Reset()                    { *m = GetDocumentResponse{} }
func (m *GetDocumentResponse) String() string            { return proto1.CompactTextString(m) }
func (*GetDocumentResponse) ProtoMessage()               {}
func (*GetDocumentResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *GetDocumentResponse) GetDocument() *Document {
	if m != nil {
		return m.Document
	}
	return nil
}

func (m *GetDocumentResponse) GetSucceeded() bool {
	if m != nil {
		return m.Succeeded
	}
	return false
}

func (m *GetDocumentResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type DeleteDocumentRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *DeleteDocumentRequest) Reset()                    { *m = DeleteDocumentRequest{} }
func (m *DeleteDocumentRequest) String() string            { return proto1.CompactTextString(m) }
func (*DeleteDocumentRequest) ProtoMessage()               {}
func (*DeleteDocumentRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *DeleteDocumentRequest) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

type DeleteDocumentResponse struct {
	Succeeded bool   `protobuf:"varint,1,opt,name=succeeded" json:"succeeded,omitempty"`
	Message   string `protobuf:"bytes,2,opt,name=message" json:"message,omitempty"`
}

func (m *DeleteDocumentResponse) Reset()                    { *m = DeleteDocumentResponse{} }
func (m *DeleteDocumentResponse) String() string            { return proto1.CompactTextString(m) }
func (*DeleteDocumentResponse) ProtoMessage()               {}
func (*DeleteDocumentResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *DeleteDocumentResponse) GetSucceeded() bool {
	if m != nil {
		return m.Succeeded
	}
	return false
}

func (m *DeleteDocumentResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type BulkRequest struct {
	BatchSize      int32            `protobuf:"varint,1,opt,name=batch_size,json=batchSize" json:"batch_size,omitempty"`
	UpdateRequests []*UpdateRequest `protobuf:"bytes,2,rep,name=update_requests,json=updateRequests" json:"update_requests,omitempty"`
}

func (m *BulkRequest) Reset()                    { *m = BulkRequest{} }
func (m *BulkRequest) String() string            { return proto1.CompactTextString(m) }
func (*BulkRequest) ProtoMessage()               {}
func (*BulkRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *BulkRequest) GetBatchSize() int32 {
	if m != nil {
		return m.BatchSize
	}
	return 0
}

func (m *BulkRequest) GetUpdateRequests() []*UpdateRequest {
	if m != nil {
		return m.UpdateRequests
	}
	return nil
}

type BulkResponse struct {
	PutCount      int32  `protobuf:"varint,1,opt,name=put_count,json=putCount" json:"put_count,omitempty"`
	PutErrorCount int32  `protobuf:"varint,2,opt,name=put_error_count,json=putErrorCount" json:"put_error_count,omitempty"`
	DeleteCount   int32  `protobuf:"varint,3,opt,name=delete_count,json=deleteCount" json:"delete_count,omitempty"`
	Succeeded     bool   `protobuf:"varint,4,opt,name=succeeded" json:"succeeded,omitempty"`
	Message       string `protobuf:"bytes,5,opt,name=message" json:"message,omitempty"`
}

func (m *BulkResponse) Reset()                    { *m = BulkResponse{} }
func (m *BulkResponse) String() string            { return proto1.CompactTextString(m) }
func (*BulkResponse) ProtoMessage()               {}
func (*BulkResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *BulkResponse) GetPutCount() int32 {
	if m != nil {
		return m.PutCount
	}
	return 0
}

func (m *BulkResponse) GetPutErrorCount() int32 {
	if m != nil {
		return m.PutErrorCount
	}
	return 0
}

func (m *BulkResponse) GetDeleteCount() int32 {
	if m != nil {
		return m.DeleteCount
	}
	return 0
}

func (m *BulkResponse) GetSucceeded() bool {
	if m != nil {
		return m.Succeeded
	}
	return false
}

func (m *BulkResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type SearchRequest struct {
	SearchRequest *google_protobuf.Any `protobuf:"bytes,1,opt,name=search_request,json=searchRequest" json:"search_request,omitempty"`
}

func (m *SearchRequest) Reset()                    { *m = SearchRequest{} }
func (m *SearchRequest) String() string            { return proto1.CompactTextString(m) }
func (*SearchRequest) ProtoMessage()               {}
func (*SearchRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *SearchRequest) GetSearchRequest() *google_protobuf.Any {
	if m != nil {
		return m.SearchRequest
	}
	return nil
}

type SearchResponse struct {
	SearchResult *google_protobuf.Any `protobuf:"bytes,1,opt,name=search_result,json=searchResult" json:"search_result,omitempty"`
	Succeeded    bool                 `protobuf:"varint,2,opt,name=succeeded" json:"succeeded,omitempty"`
	Message      string               `protobuf:"bytes,3,opt,name=message" json:"message,omitempty"`
}

func (m *SearchResponse) Reset()                    { *m = SearchResponse{} }
func (m *SearchResponse) String() string            { return proto1.CompactTextString(m) }
func (*SearchResponse) ProtoMessage()               {}
func (*SearchResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *SearchResponse) GetSearchResult() *google_protobuf.Any {
	if m != nil {
		return m.SearchResult
	}
	return nil
}

func (m *SearchResponse) GetSucceeded() bool {
	if m != nil {
		return m.Succeeded
	}
	return false
}

func (m *SearchResponse) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto1.RegisterType((*Document)(nil), "proto.Document")
	proto1.RegisterType((*UpdateRequest)(nil), "proto.UpdateRequest")
	proto1.RegisterType((*GetIndexRequest)(nil), "proto.GetIndexRequest")
	proto1.RegisterType((*GetIndexResponse)(nil), "proto.GetIndexResponse")
	proto1.RegisterType((*PutDocumentRequest)(nil), "proto.PutDocumentRequest")
	proto1.RegisterType((*PutDocumentResponse)(nil), "proto.PutDocumentResponse")
	proto1.RegisterType((*GetDocumentRequest)(nil), "proto.GetDocumentRequest")
	proto1.RegisterType((*GetDocumentResponse)(nil), "proto.GetDocumentResponse")
	proto1.RegisterType((*DeleteDocumentRequest)(nil), "proto.DeleteDocumentRequest")
	proto1.RegisterType((*DeleteDocumentResponse)(nil), "proto.DeleteDocumentResponse")
	proto1.RegisterType((*BulkRequest)(nil), "proto.BulkRequest")
	proto1.RegisterType((*BulkResponse)(nil), "proto.BulkResponse")
	proto1.RegisterType((*SearchRequest)(nil), "proto.SearchRequest")
	proto1.RegisterType((*SearchResponse)(nil), "proto.SearchResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Index service

type IndexClient interface {
	GetIndex(ctx context.Context, in *GetIndexRequest, opts ...grpc.CallOption) (*GetIndexResponse, error)
	PutDocument(ctx context.Context, in *PutDocumentRequest, opts ...grpc.CallOption) (*PutDocumentResponse, error)
	GetDocument(ctx context.Context, in *GetDocumentRequest, opts ...grpc.CallOption) (*GetDocumentResponse, error)
	DeleteDocument(ctx context.Context, in *DeleteDocumentRequest, opts ...grpc.CallOption) (*DeleteDocumentResponse, error)
	Bulk(ctx context.Context, in *BulkRequest, opts ...grpc.CallOption) (*BulkResponse, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error)
}

type indexClient struct {
	cc *grpc.ClientConn
}

func NewIndexClient(cc *grpc.ClientConn) IndexClient {
	return &indexClient{cc}
}

func (c *indexClient) GetIndex(ctx context.Context, in *GetIndexRequest, opts ...grpc.CallOption) (*GetIndexResponse, error) {
	out := new(GetIndexResponse)
	err := grpc.Invoke(ctx, "/proto.Index/GetIndex", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) PutDocument(ctx context.Context, in *PutDocumentRequest, opts ...grpc.CallOption) (*PutDocumentResponse, error) {
	out := new(PutDocumentResponse)
	err := grpc.Invoke(ctx, "/proto.Index/PutDocument", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) GetDocument(ctx context.Context, in *GetDocumentRequest, opts ...grpc.CallOption) (*GetDocumentResponse, error) {
	out := new(GetDocumentResponse)
	err := grpc.Invoke(ctx, "/proto.Index/GetDocument", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) DeleteDocument(ctx context.Context, in *DeleteDocumentRequest, opts ...grpc.CallOption) (*DeleteDocumentResponse, error) {
	out := new(DeleteDocumentResponse)
	err := grpc.Invoke(ctx, "/proto.Index/DeleteDocument", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) Bulk(ctx context.Context, in *BulkRequest, opts ...grpc.CallOption) (*BulkResponse, error) {
	out := new(BulkResponse)
	err := grpc.Invoke(ctx, "/proto.Index/Bulk", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *indexClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	out := new(SearchResponse)
	err := grpc.Invoke(ctx, "/proto.Index/Search", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Index service

type IndexServer interface {
	GetIndex(context.Context, *GetIndexRequest) (*GetIndexResponse, error)
	PutDocument(context.Context, *PutDocumentRequest) (*PutDocumentResponse, error)
	GetDocument(context.Context, *GetDocumentRequest) (*GetDocumentResponse, error)
	DeleteDocument(context.Context, *DeleteDocumentRequest) (*DeleteDocumentResponse, error)
	Bulk(context.Context, *BulkRequest) (*BulkResponse, error)
	Search(context.Context, *SearchRequest) (*SearchResponse, error)
}

func RegisterIndexServer(s *grpc.Server, srv IndexServer) {
	s.RegisterService(&_Index_serviceDesc, srv)
}

func _Index_GetIndex_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIndexRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).GetIndex(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Index/GetIndex",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).GetIndex(ctx, req.(*GetIndexRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_PutDocument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutDocumentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).PutDocument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Index/PutDocument",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).PutDocument(ctx, req.(*PutDocumentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_GetDocument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDocumentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).GetDocument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Index/GetDocument",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).GetDocument(ctx, req.(*GetDocumentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_DeleteDocument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteDocumentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).DeleteDocument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Index/DeleteDocument",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).DeleteDocument(ctx, req.(*DeleteDocumentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_Bulk_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BulkRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).Bulk(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Index/Bulk",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).Bulk(ctx, req.(*BulkRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Index_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IndexServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Index/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IndexServer).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Index_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Index",
	HandlerType: (*IndexServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetIndex",
			Handler:    _Index_GetIndex_Handler,
		},
		{
			MethodName: "PutDocument",
			Handler:    _Index_PutDocument_Handler,
		},
		{
			MethodName: "GetDocument",
			Handler:    _Index_GetDocument_Handler,
		},
		{
			MethodName: "DeleteDocument",
			Handler:    _Index_DeleteDocument_Handler,
		},
		{
			MethodName: "Bulk",
			Handler:    _Index_Bulk_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _Index_Search_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/index_service.proto",
}

func init() { proto1.RegisterFile("proto/index_service.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 741 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xa4, 0x54, 0xdd, 0x4e, 0xdb, 0x4a,
	0x10, 0xc6, 0xf9, 0x23, 0x99, 0xfc, 0xa1, 0xcd, 0x81, 0x13, 0x7c, 0x40, 0xe2, 0x58, 0x47, 0x07,
	0xaa, 0xa2, 0xd0, 0xd2, 0x8b, 0x0a, 0x55, 0x5c, 0xd0, 0xd2, 0xd2, 0xaa, 0x45, 0x45, 0x86, 0x5e,
	0x47, 0xc6, 0x1e, 0x12, 0x8b, 0xc4, 0x76, 0xed, 0x5d, 0xd4, 0x70, 0xdb, 0x07, 0xe8, 0x45, 0xdf,
	0xa3, 0xaf, 0xd1, 0xd7, 0xaa, 0xbc, 0x3f, 0xc6, 0x9b, 0xa4, 0x41, 0x6d, 0xaf, 0xac, 0xfd, 0xf6,
	0x9b, 0x6f, 0xbe, 0x19, 0xcf, 0x2c, 0xac, 0x47, 0x71, 0x48, 0xc3, 0x3d, 0x3f, 0xf0, 0xf0, 0x53,
	0x3f, 0xc1, 0xf8, 0xc6, 0x77, 0xb1, 0xc7, 0x31, 0x52, 0xe6, 0x1f, 0x73, 0x7d, 0x10, 0x86, 0x83,
	0x11, 0xee, 0xf1, 0xd3, 0x25, 0xbb, 0xda, 0x73, 0x82, 0x89, 0x60, 0x58, 0xaf, 0xa1, 0x7a, 0x1c,
	0xba, 0x6c, 0x8c, 0x01, 0x25, 0x2d, 0x28, 0xf8, 0x5e, 0xd7, 0xd8, 0x32, 0x76, 0x6a, 0x76, 0xc1,
	0xf7, 0xc8, 0x2e, 0x54, 0xae, 0x7c, 0x1c, 0x79, 0x49, 0xb7, 0xb0, 0x65, 0xec, 0xd4, 0xf7, 0xff,
	0xea, 0x09, 0x9d, 0x9e, 0xd2, 0xe9, 0x1d, 0x05, 0x13, 0x5b, 0x72, 0xac, 0x0b, 0x68, 0x7e, 0x88,
	0x3c, 0x87, 0xa2, 0x8d, 0x1f, 0x19, 0x26, 0x94, 0xac, 0x41, 0x65, 0x8c, 0x74, 0x18, 0x2a, 0x49,
	0x79, 0x22, 0x0f, 0xa1, 0xea, 0xc9, 0x94, 0x52, 0xb8, 0x2d, 0x14, 0x7b, 0xca, 0x89, 0x9d, 0x11,
	0xac, 0xef, 0x06, 0xb4, 0x4f, 0x90, 0xbe, 0x49, 0x8b, 0x53, 0xc2, 0xfb, 0xb0, 0xea, 0x07, 0xee,
	0x88, 0x79, 0xd8, 0x17, 0x45, 0x8f, 0x9d, 0x28, 0xf2, 0x83, 0x01, 0xcf, 0x53, 0xb5, 0x3b, 0xf2,
	0x92, 0xc7, 0x9c, 0x8a, 0x2b, 0xb2, 0x0b, 0x44, 0x8f, 0xa1, 0x93, 0x08, 0x79, 0xfa, 0xaa, 0xbd,
	0x92, 0x0f, 0xb8, 0x98, 0x44, 0x48, 0xb6, 0xa1, 0xad, 0xd8, 0xd7, 0x37, 0x09, 0x0d, 0x63, 0xec,
	0x16, 0x39, 0xb5, 0x25, 0xe1, 0xb7, 0x02, 0x25, 0x0f, 0x60, 0xe5, 0x8e, 0xe8, 0x86, 0xc1, 0x95,
	0x3f, 0xe8, 0x96, 0x38, 0xb3, 0x9d, 0x31, 0x05, 0x6c, 0x7d, 0x29, 0xc0, 0xca, 0x5d, 0x25, 0x49,
	0x14, 0x06, 0x09, 0x92, 0x4d, 0x00, 0x61, 0x27, 0x72, 0xe8, 0x50, 0xf6, 0xa9, 0xc6, 0x91, 0x33,
	0x87, 0x0e, 0xc9, 0x01, 0x34, 0xf5, 0x0a, 0x17, 0xfd, 0x88, 0x86, 0x9f, 0x2f, 0x38, 0x53, 0xe6,
	0x85, 0x16, 0x73, 0xca, 0xbc, 0xc2, 0x2e, 0x2c, 0xab, 0xca, 0x4a, 0xfc, 0x4e, 0x1d, 0xc9, 0x23,
	0xa8, 0x66, 0xa5, 0x94, 0x17, 0xa4, 0xcb, 0x58, 0x64, 0x03, 0x6a, 0x09, 0x73, 0x5d, 0x44, 0x0f,
	0xbd, 0x6e, 0x85, 0x57, 0x7f, 0x07, 0xa4, 0x99, 0xc6, 0x98, 0x24, 0xce, 0x00, 0xbb, 0xcb, 0x22,
	0x93, 0x3c, 0x5a, 0x47, 0x40, 0xce, 0x18, 0xcd, 0x7e, 0xba, 0xfc, 0xbb, 0xf9, 0xf1, 0x30, 0xee,
	0x1b, 0x8f, 0x53, 0xe8, 0x68, 0x12, 0xb2, 0xad, 0x9a, 0x23, 0x63, 0x81, 0xa3, 0x82, 0xee, 0xe8,
	0x3f, 0x20, 0x27, 0x38, 0xe3, 0x68, 0x6a, 0x2f, 0xac, 0x5b, 0xe8, 0x68, 0x2c, 0x99, 0xf4, 0x57,
	0x8c, 0xeb, 0x0e, 0x0b, 0x0b, 0x1c, 0x16, 0x75, 0x87, 0xdb, 0xb0, 0x7a, 0x8c, 0x23, 0xa4, 0x78,
	0x9f, 0xc9, 0x33, 0x58, 0x9b, 0x26, 0xfe, 0x61, 0x73, 0xae, 0xa1, 0xfe, 0x9c, 0x8d, 0xae, 0x55,
	0xc2, 0x4d, 0x80, 0x4b, 0x87, 0xba, 0xc3, 0x7e, 0xe2, 0xdf, 0x22, 0xd7, 0x29, 0xdb, 0x35, 0x8e,
	0x9c, 0xfb, 0xb7, 0x48, 0x0e, 0xa1, 0xcd, 0xf8, 0x73, 0xd0, 0x8f, 0x45, 0x40, 0xfa, 0x8a, 0x14,
	0xf9, 0x34, 0x89, 0xa6, 0x68, 0x8f, 0x85, 0xdd, 0x62, 0xf9, 0x63, 0x62, 0x7d, 0x33, 0xa0, 0x21,
	0xb2, 0x49, 0xd7, 0xff, 0x40, 0x2d, 0x62, 0xb4, 0xef, 0x86, 0x4c, 0xb6, 0xb7, 0x6c, 0x57, 0x23,
	0x46, 0x5f, 0xa4, 0x67, 0xf2, 0x3f, 0xb4, 0xd3, 0x4b, 0x8c, 0xe3, 0x30, 0x96, 0x94, 0x02, 0xa7,
	0x34, 0x23, 0x46, 0x5f, 0xa6, 0xa8, 0xe0, 0xfd, 0x0b, 0x0d, 0x8f, 0x37, 0x45, 0x92, 0x8a, 0x9c,
	0x54, 0x17, 0x98, 0xa0, 0x68, 0xdd, 0x29, 0x2d, 0xe8, 0x4e, 0x59, 0xef, 0xce, 0x3b, 0x68, 0x9e,
	0xa3, 0x13, 0xbb, 0x43, 0xd5, 0x9f, 0x67, 0xd0, 0x4a, 0x38, 0xa0, 0x1a, 0x20, 0x87, 0x62, 0xfe,
	0x36, 0x35, 0x93, 0x7c, 0xb0, 0xf5, 0xd9, 0x80, 0x96, 0x92, 0x93, 0x0d, 0x38, 0x80, 0x66, 0xa6,
	0x97, 0xb0, 0xd1, 0x62, 0xb9, 0x86, 0x92, 0x4b, 0x99, 0xbf, 0x3b, 0x6c, 0xfb, 0x5f, 0x8b, 0x50,
	0xe6, 0xef, 0x15, 0x39, 0x84, 0xaa, 0x7a, 0xbb, 0xc8, 0x9a, 0xfc, 0x81, 0x53, 0xcf, 0xb2, 0xf9,
	0xf7, 0x0c, 0x2e, 0x9c, 0x5b, 0x4b, 0xe4, 0x15, 0xd4, 0x73, 0x6b, 0x4a, 0xd6, 0x25, 0x73, 0x76,
	0xfb, 0x4d, 0x73, 0xde, 0x55, 0x5e, 0x27, 0xb7, 0x79, 0x99, 0xce, 0xec, 0xce, 0x66, 0x3a, 0x73,
	0x16, 0xd5, 0x5a, 0x22, 0xef, 0xa1, 0xa5, 0x2f, 0x07, 0xd9, 0x50, 0xab, 0x3a, 0x6f, 0xb9, 0xcc,
	0xcd, 0x9f, 0xdc, 0x66, 0x82, 0x8f, 0xa1, 0x94, 0x4e, 0x2b, 0x21, 0x92, 0x98, 0x5b, 0x14, 0xb3,
	0xa3, 0x61, 0x59, 0xc8, 0x53, 0xa8, 0x88, 0x3f, 0x4c, 0xd4, 0x46, 0x68, 0xf3, 0x63, 0xae, 0x4e,
	0xa1, 0x2a, 0xf0, 0xb2, 0xc2, 0xf1, 0x27, 0x3f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x38, 0x75, 0xb7,
	0x4d, 0xf8, 0x07, 0x00, 0x00,
}
