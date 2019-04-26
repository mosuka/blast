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

package protobuf

import (
	"bytes"
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/protobuf/index"
	"github.com/mosuka/blast/protobuf/raft"
)

func TestMarshalAny_Slice(t *testing.T) {
	data := []interface{}{"a", 1}

	dataAny := &any.Any{}
	err := UnmarshalAny(data, dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType := "[]interface {}"
	actualType := dataAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue := []byte(`["a",1]`)
	actualValue := dataAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestMarshalAny_Map(t *testing.T) {
	data := map[string]interface{}{"a": 1, "b": 2, "c": 3}

	dataAny := &any.Any{}
	err := UnmarshalAny(data, dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedMapType := "map[string]interface {}"
	actualMapType := dataAny.TypeUrl
	if expectedMapType != actualMapType {
		t.Errorf("expected content to see %s, saw %s", expectedMapType, actualMapType)
	}

	expectedValue := []byte(`{"a":1,"b":2,"c":3}`)
	actualValue := dataAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestMarshalAny_Document(t *testing.T) {
	fieldsMap := map[string]interface{}{"f1": "aaa", "f2": 222, "f3": "ccc"}
	fieldsAny := &any.Any{}
	err := UnmarshalAny(fieldsMap, fieldsAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	data := &index.Document{
		Id:     "1",
		Fields: fieldsAny,
	}

	dataAny := &any.Any{}
	err = UnmarshalAny(data, dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType := "index.Document"
	actualType := dataAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue := []byte(`{"id":"1","fields":{"type_url":"map[string]interface {}","value":"eyJmMSI6ImFhYSIsImYyIjoyMjIsImYzIjoiY2NjIn0="}}`)
	actualValue := dataAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestMarshalAny_Node(t *testing.T) {
	data := &raft.Node{
		Id: "node1",
		Metadata: &raft.Metadata{
			GrpcAddr: ":5050",
			DataDir:  "/tmp/blast/index1",
			BindAddr: ":6060",
			HttpAddr: ":8080",
		},
		Leader: true,
	}

	dataAny := &any.Any{}
	err := UnmarshalAny(data, dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType := "raft.Node"
	actualType := dataAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue := []byte(`{"id":"node1","metadata":{"bind_addr":":6060","grpc_addr":":5050","http_addr":":8080","data_dir":"/tmp/blast/index1"},"leader":true}`)
	actualValue := dataAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestMarshalAny_SearchRequest(t *testing.T) {
	data := bleve.NewSearchRequest(bleve.NewQueryStringQuery("blast"))

	dataAny := &any.Any{}
	err := UnmarshalAny(data, dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType := "bleve.SearchRequest"
	actualType := dataAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue := []byte(`{"query":{"query":"blast"},"size":10,"from":0,"highlight":null,"fields":null,"facets":null,"explain":false,"sort":["-_score"],"includeLocations":false}`)
	actualValue := dataAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestMarshalAny_SearchResult(t *testing.T) {
	data := &bleve.SearchResult{
		Total: 10,
	}

	dataAny := &any.Any{}
	err := UnmarshalAny(data, dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType := "bleve.SearchResult"
	actualType := dataAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue := []byte(`{"status":null,"request":null,"hits":null,"total_hits":10,"max_score":0,"took":0,"facets":null}`)
	actualValue := dataAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestUnmarshalAny_Slice(t *testing.T) {
	dataAny := &any.Any{
		TypeUrl: "[]interface {}",
		Value:   []byte(`["a",1]`),
	}

	ins, err := MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	data := *ins.(*[]interface{})

	expected1 := "a"
	actual1 := data[0]
	if expected1 != actual1 {
		t.Errorf("expected content to see %v, saw %v", expected1, actual1)
	}

	expected2 := float64(1)
	actual2 := data[1]
	if expected2 != actual2 {
		t.Errorf("expected content to see %v, saw %v", expected2, actual2)
	}
}

func TestUnmarshalAny_Map(t *testing.T) {
	dataAny := &any.Any{
		TypeUrl: "map[string]interface {}",
		Value:   []byte(`{"a":1,"b":2,"c":3}`),
	}

	ins, err := MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	data := *ins.(*map[string]interface{})

	expected1 := float64(1)
	actual1 := data["a"]
	if expected1 != actual1 {
		t.Errorf("expected content to see %v, saw %v", expected1, actual1)
	}

	expected2 := float64(2)
	actual2 := data["b"]
	if expected2 != actual2 {
		t.Errorf("expected content to see %v, saw %v", expected2, actual2)
	}

	expected3 := float64(3)
	actual3 := data["c"]
	if expected3 != actual3 {
		t.Errorf("expected content to see %v, saw %v", expected3, actual3)
	}
}

func TestUnmarshalAny_Document(t *testing.T) {
	dataAny := &any.Any{
		TypeUrl: "index.Document",
		Value:   []byte(`{"id":"1","fields":{"type_url":"map[string]interface {}","value":"eyJmMSI6ImFhYSIsImYyIjoyMjIsImYzIjoiY2NjIn0="}}`),
	}

	ins, err := MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	data := *ins.(*index.Document)

	expected1 := "1"
	actual1 := data.Id
	if expected1 != actual1 {
		t.Errorf("expected content to see %v, saw %v", expected1, actual1)
	}

	expected2 := "map[string]interface {}"
	actual2 := data.Fields.TypeUrl
	if expected2 != actual2 {
		t.Errorf("expected content to see %v, saw %v", expected2, actual2)
	}

	expected3 := []byte(`{"f1":"aaa","f2":222,"f3":"ccc"}`)
	actual3 := data.Fields.Value
	if !bytes.Equal(expected3, actual3) {
		t.Errorf("expected content to see %v, saw %v", expected3, actual3)
	}
}

func TestUnmarshalAny_SearchRequest(t *testing.T) {
	dataAny := &any.Any{
		TypeUrl: "bleve.SearchRequest",
		Value:   []byte(`{"query":{"query":"blast"},"size":10,"from":0,"highlight":null,"fields":null,"facets":null,"explain":false,"sort":["-_score"],"includeLocations":false}`),
	}

	ins, err := MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	data := *ins.(*bleve.SearchRequest)

	expected1 := bleve.NewQueryStringQuery("blast").Query
	actual1 := data.Query.(*query.QueryStringQuery).Query
	if expected1 != actual1 {
		t.Errorf("expected content to see %v, saw %v", expected1, actual1)
	}
}

func TestUnmarshalAny_SearchResult(t *testing.T) {
	dataAny := &any.Any{
		TypeUrl: "bleve.SearchResult",
		Value:   []byte(`{"status":null,"request":null,"hits":null,"total_hits":10,"max_score":0,"took":0,"facets":null}`),
	}

	ins, err := MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	data := *ins.(*bleve.SearchResult)

	expected1 := uint64(10)
	actual1 := data.Total
	if expected1 != actual1 {
		t.Errorf("expected content to see %v, saw %v", expected1, actual1)
	}
}
