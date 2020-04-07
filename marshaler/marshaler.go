package marshaler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
)

var (
	DefaultContentType = "application/json"
)

type BlastMarshaler struct{}

func (*BlastMarshaler) ContentType() string {
	return DefaultContentType
}

func (m *BlastMarshaler) Marshal(v interface{}) ([]byte, error) {
	switch v.(type) {
	case *protobuf.GetResponse:
		var fields map[string]interface{}
		if err := json.Unmarshal(v.(*protobuf.GetResponse).Fields, &fields); err != nil {
			return nil, err
		}
		resp := map[string]interface{}{
			"fields": fields,
		}
		if value, err := json.Marshal(resp); err == nil {
			return value, nil
		} else {
			return nil, err
		}
	case *protobuf.SearchResponse:
		var searchResult map[string]interface{}
		if err := json.Unmarshal(v.(*protobuf.SearchResponse).SearchResult, &searchResult); err != nil {
			return nil, err
		}
		resp := map[string]interface{}{
			"search_result": searchResult,
		}
		if value, err := json.Marshal(resp); err == nil {
			return value, nil
		} else {
			return nil, err
		}
	case *protobuf.MappingResponse:
		var m map[string]interface{}
		if err := json.Unmarshal(v.(*protobuf.MappingResponse).Mapping, &m); err != nil {
			return nil, err
		}
		resp := map[string]interface{}{
			"mapping": m,
		}
		if value, err := json.Marshal(resp); err == nil {
			return value, nil
		} else {
			return nil, err
		}
	case *protobuf.MetricsResponse:
		value := v.(*protobuf.MetricsResponse).Metrics
		return value, nil
	default:
		return json.Marshal(v)
	}
}

func (m *BlastMarshaler) Unmarshal(data []byte, v interface{}) error {
	switch v.(type) {
	case *protobuf.SetRequest:
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		if i, ok := m["id"].(string); ok {
			v.(*protobuf.SetRequest).Id = i
		}

		if f, ok := m["fields"].(map[string]interface{}); ok {
			fieldsBytes, err := json.Marshal(f)
			if err != nil {
				return err
			}
			v.(*protobuf.SetRequest).Fields = fieldsBytes
		}
		return nil
	case *protobuf.BulkIndexRequest:
		v.(*protobuf.BulkIndexRequest).Requests = make([]*protobuf.SetRequest, 0)

		reader := bufio.NewReader(bytes.NewReader(data))
		for {
			docBytes, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					if len(docBytes) > 0 {
						r := &protobuf.SetRequest{}
						if err := m.Unmarshal(docBytes, r); err != nil {
							continue
						}
						v.(*protobuf.BulkIndexRequest).Requests = append(v.(*protobuf.BulkIndexRequest).Requests, r)
					}
					break
				}
			}
			if len(docBytes) > 0 {
				r := &protobuf.SetRequest{}
				if err := m.Unmarshal(docBytes, r); err != nil {
					continue
				}
				v.(*protobuf.BulkIndexRequest).Requests = append(v.(*protobuf.BulkIndexRequest).Requests, r)
			}
		}
		return nil
	case *protobuf.BulkDeleteRequest:
		v.(*protobuf.BulkDeleteRequest).Requests = make([]*protobuf.DeleteRequest, 0)

		reader := bufio.NewReader(bytes.NewReader(data))
		for {
			docBytes, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF || err == io.ErrClosedPipe {
					if len(docBytes) > 0 {
						r := &protobuf.DeleteRequest{
							Id: strings.TrimSpace(string(docBytes)),
						}
						v.(*protobuf.BulkDeleteRequest).Requests = append(v.(*protobuf.BulkDeleteRequest).Requests, r)
					}
					break
				}
			}
			if len(docBytes) > 0 {
				r := &protobuf.DeleteRequest{
					Id: strings.TrimSpace(string(docBytes)),
				}
				v.(*protobuf.BulkDeleteRequest).Requests = append(v.(*protobuf.BulkDeleteRequest).Requests, r)
			}
		}
		return nil
	case *protobuf.SearchRequest:
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		f, ok := m["search_request"]
		if !ok {
			return errors.ErrNil
		}
		searchRequestBytes, err := json.Marshal(f)
		if err != nil {
			return err
		}
		v.(*protobuf.SearchRequest).SearchRequest = searchRequestBytes
		return nil
	default:
		return json.Unmarshal(data, v)
	}
}

func (m *BlastMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	return runtime.DecoderFunc(
		func(v interface{}) error {
			buffer, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}

			return m.Unmarshal(buffer, v)
		},
	)
}

func (m *BlastMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	return json.NewEncoder(w)
}

func (m *BlastMarshaler) Delimiter() []byte {
	return []byte("\n")
}
