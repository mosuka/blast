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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type JsonMarshaler struct{}

// ContentType always Returns "application/json".
func (*JsonMarshaler) ContentType() string {
	return "application/json"
}

// Marshal marshals "v" into JSON
func (j *JsonMarshaler) Marshal(v interface{}) ([]byte, error) {
	switch v.(type) {
	case *index.GetResponse:
		value, err := protobuf.MarshalAny(v.(*index.GetResponse).Document.Fields)
		if err != nil {
			return nil, err
		}
		return json.Marshal(
			map[string]interface{}{
				"document": map[string]interface{}{
					"id":     v.(*index.GetResponse).Document.Id,
					"fields": value,
				},
			},
		)
	default:
		return json.Marshal(v)
	}
}

// Unmarshal unmarshals JSON data into "v".
func (j *JsonMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads JSON stream from "r".
func (j *JsonMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			buffer, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}

			switch v.(type) {
			case *index.IndexRequest:
				var tmpValue map[string]interface{}
				err = json.Unmarshal(buffer, &tmpValue)
				if err != nil {
					return err
				}
				//id, ok := tmpValue["id"].(string)
				//if ok {
				//	v.(*index.IndexRequest).Id = id
				//}
				fields, ok := tmpValue["fields"]
				if !ok {
					return errors.New("value does not exist")
				}
				v.(*index.IndexRequest).Fields = &any.Any{}
				return protobuf.UnmarshalAny(fields, v.(*index.IndexRequest).Fields)
			default:
				return json.Unmarshal(buffer, v)
			}
		},
	)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *JsonMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// Delimiter for newline encoded JSON streams.
func (j *JsonMarshaler) Delimiter() []byte {
	return []byte("\n")
}

type JsonlMarshaler struct{}

// ContentType always Returns "application/json".
func (*JsonlMarshaler) ContentType() string {
	return "application/json"
}

// Marshal marshals "v" into JSON
func (j *JsonlMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal unmarshals JSON data into "v".
func (j *JsonlMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads JSON-LINE stream from "r".
func (j *JsonlMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			buffer, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}

			switch v.(type) {
			case *index.BulkIndexRequest:
				docs := make([]*index.Document, 0)
				reader := bufio.NewReader(bytes.NewReader(buffer))
				for {
					docBytes, err := reader.ReadBytes('\n')
					if err != nil {
						if err == io.EOF || err == io.ErrClosedPipe {
							if len(docBytes) > 0 {
								doc := &index.Document{}
								err = index.UnmarshalDocument(docBytes, doc)
								if err != nil {
									return err
								}
								docs = append(docs, doc)
							}
							break
						}
					}

					if len(docBytes) > 0 {
						doc := &index.Document{}
						err = index.UnmarshalDocument(docBytes, doc)
						if err != nil {
							return err
						}
						docs = append(docs, doc)
					}
				}
				v.(*index.BulkIndexRequest).Documents = docs
				return nil
			default:
				return json.Unmarshal(buffer, v)
			}
		},
	)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *JsonlMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// Delimiter for newline encoded JSON streams.
func (j *JsonlMarshaler) Delimiter() []byte {
	return []byte("\n")
}

type TextMarshaler struct{}

// ContentType always Returns "application/json".
func (*TextMarshaler) ContentType() string {
	return "application/json"
}

// Marshal marshals "v" into JSON
func (j *TextMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal unmarshals JSON data into "v".
func (j *TextMarshaler) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewDecoder returns a Decoder which reads text stream from "r".
func (j *TextMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			buffer, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}

			switch v.(type) {
			case *index.BulkDeleteRequest:
				ids := make([]string, 0)
				reader := bufio.NewReader(bytes.NewReader(buffer))
				for {
					//idBytes, err := reader.ReadBytes('\n')
					idBytes, _, err := reader.ReadLine()
					if err != nil {
						if err == io.EOF || err == io.ErrClosedPipe {
							if len(idBytes) > 0 {
								ids = append(ids, string(idBytes))
							}
							break
						}
					}

					if len(idBytes) > 0 {
						ids = append(ids, string(idBytes))
					}
				}
				v.(*index.BulkDeleteRequest).Ids = ids
				return nil
			default:
				return json.Unmarshal(buffer, v)
			}
		},
	)
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *TextMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

// Delimiter for newline encoded JSON streams.
func (j *TextMarshaler) Delimiter() []byte {
	return []byte("\n")
}

type GRPCGateway struct {
	grpcGatewayAddr string
	grpcAddr        string
	logger          *zap.Logger

	ctx      context.Context
	cancel   context.CancelFunc
	listener net.Listener
}

func NewGRPCGateway(grpcGatewayAddr string, grpcAddr string, logger *zap.Logger) (*GRPCGateway, error) {
	return &GRPCGateway{
		grpcGatewayAddr: grpcGatewayAddr,
		grpcAddr:        grpcAddr,
		logger:          logger,
	}, nil
}

func (s *GRPCGateway) Start() error {
	s.ctx, s.cancel = NewGRPCContext()

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption("application/json", new(JsonMarshaler)),
		runtime.WithMarshalerOption("application/x-ndjson", new(JsonlMarshaler)),
		runtime.WithMarshalerOption("text/plain", new(TextMarshaler)),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := index.RegisterIndexHandlerFromEndpoint(s.ctx, mux, s.grpcAddr, opts)
	if err != nil {
		return err
	}

	s.listener, err = net.Listen("tcp", s.grpcGatewayAddr)
	if err != nil {
		return err
	}

	err = http.Serve(s.listener, mux)
	if err != nil {
		return err
	}

	return nil
}

func (s *GRPCGateway) Stop() error {
	defer s.cancel()

	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *GRPCGateway) GetAddress() (string, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.listener.Addr().String())
	if err != nil {
		return "", err
	}

	v4Addr := ""
	if tcpAddr.IP.To4() != nil {
		v4Addr = tcpAddr.IP.To4().String()
	}
	port := tcpAddr.Port

	return fmt.Sprintf("%s:%d", v4Addr, port), nil
}
