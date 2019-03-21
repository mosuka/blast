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

func TestMarshalAny(t *testing.T) {
	// test map[string]interface{}
	data := map[string]interface{}{"a": 1, "b": 2, "c": 3}

	mapAny := &any.Any{}
	err := UnmarshalAny(data, mapAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType := "map[string]interface {}"
	actualType := mapAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue := []byte(`{"a":1,"b":2,"c":3}`)
	actualValue := mapAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}

	// test index.Document
	fieldsMap := map[string]interface{}{"f1": "aaa", "f2": 222, "f3": "ccc"}
	fieldsAny := &any.Any{}
	err = UnmarshalAny(fieldsMap, fieldsAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	doc := &index.Document{
		Id:     "1",
		Fields: fieldsAny,
	}

	docAny := &any.Any{}
	err = UnmarshalAny(doc, docAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType = "index.Document"
	actualType = docAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue = []byte(`{"id":"1","fields":{"type_url":"map[string]interface {}","value":"eyJmMSI6ImFhYSIsImYyIjoyMjIsImYzIjoiY2NjIn0="}}`)
	actualValue = docAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}

	// test raft.Node
	node := &raft.Node{
		Id:       "node1",
		GrpcAddr: ":5050",
		DataDir:  "/tmp/blast/index1",
		BindAddr: ":6060",
		HttpAddr: ":8080",
		Leader:   true,
	}

	nodeAny := &any.Any{}
	err = UnmarshalAny(node, nodeAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType = "raft.Node"
	actualType = nodeAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue = []byte(`{"id":"node1","bind_addr":":6060","grpc_addr":":5050","http_addr":":8080","leader":true,"data_dir":"/tmp/blast/index1"}`)
	actualValue = nodeAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}

	// test bleve.SearchRequest
	searchReq := bleve.NewSearchRequest(bleve.NewQueryStringQuery("blast"))

	searchReqAny := &any.Any{}
	err = UnmarshalAny(searchReq, searchReqAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType = "bleve.SearchRequest"
	actualType = searchReqAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue = []byte(`{"query":{"query":"blast"},"size":10,"from":0,"highlight":null,"fields":null,"facets":null,"explain":false,"sort":["-_score"],"includeLocations":false}`)
	actualValue = searchReqAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}

	// test bleve.SearchResult
	searchReslt := &bleve.SearchResult{
		Total: 10,
	}

	searchResltAny := &any.Any{}
	err = UnmarshalAny(searchReslt, searchResltAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType = "bleve.SearchResult"
	actualType = searchResltAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue = []byte(`{"status":null,"request":null,"hits":null,"total_hits":10,"max_score":0,"took":0,"facets":null}`)
	actualValue = searchResltAny.Value
	if !bytes.Equal(expectedValue, actualValue) {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestUnmarshalAny(t *testing.T) {
	// test map[string]interface{}
	dataAny := &any.Any{
		TypeUrl: "map[string]interface {}",
		Value:   []byte(`{"a":1,"b":2,"c":3}`),
	}

	data, err := MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}
	dataMap := *data.(*map[string]interface{})

	if dataMap["a"] != float64(1) {
		t.Errorf("expected content to see %v, saw %v", 1, dataMap["a"])
	}
	if dataMap["b"] != float64(2) {
		t.Errorf("expected content to see %v, saw %v", 2, dataMap["b"])
	}
	if dataMap["c"] != float64(3) {
		t.Errorf("expected content to see %v, saw %v", 3, dataMap["c"])
	}

	// index.Document
	dataAny = &any.Any{
		TypeUrl: "index.Document",
		Value:   []byte(`{"id":"1","fields":{"type_url":"map[string]interface {}","value":"eyJmMSI6ImFhYSIsImYyIjoyMjIsImYzIjoiY2NjIn0="}}`),
	}

	data, err = MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}
	dataDoc := data.(*index.Document)

	if dataDoc.Id != "1" {
		t.Errorf("expected content to see %v, saw %v", "1", dataDoc.Id)
	}
	if dataDoc.Fields.TypeUrl != "map[string]interface {}" {
		t.Errorf("expected content to see %v, saw %v", "map[string]interface {}", dataDoc.Fields.TypeUrl)
	}
	if !bytes.Equal(dataDoc.Fields.Value, []byte(`{"f1":"aaa","f2":222,"f3":"ccc"}`)) {
		t.Errorf("expected content to see %v, saw %v", []byte("eyJmMSI6ImFhYSIsImYyIjoyMjIsImYzIjoiY2NjIn0="), dataDoc.Fields.Value)
	}

	// raft.Node
	dataAny = &any.Any{
		TypeUrl: "raft.Node",
		Value:   []byte(`{"id":"node1","bind_addr":":6060","grpc_addr":":5050","http_addr":":8080","leader":true,"data_dir":"/tmp/blast/index1"}`),
	}

	data, err = MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}
	dataNode := data.(*raft.Node)

	if dataNode.Id != "node1" {
		t.Errorf("expected content to see %v, saw %v", "node1", dataNode.Id)
	}
	if dataNode.HttpAddr != ":8080" {
		t.Errorf("expected content to see %v, saw %v", ":8080", dataNode.HttpAddr)
	}
	if dataNode.BindAddr != ":6060" {
		t.Errorf("expected content to see %v, saw %v", ":6060", dataNode.BindAddr)
	}
	if dataNode.GrpcAddr != ":5050" {
		t.Errorf("expected content to see %v, saw %v", ":5050", dataNode.BindAddr)
	}
	if dataNode.DataDir != "/tmp/blast/index1" {
		t.Errorf("expected content to see %v, saw %v", "/tmp/blast/index1", dataNode.DataDir)
	}
	if dataNode.Leader != true {
		t.Errorf("expected content to see %v, saw %v", true, dataNode.Leader)
	}

	// test bleve.SearchRequest
	dataAny = &any.Any{
		TypeUrl: "bleve.SearchRequest",
		Value:   []byte(`{"query":{"query":"blast"},"size":10,"from":0,"highlight":null,"fields":null,"facets":null,"explain":false,"sort":["-_score"],"includeLocations":false}`),
	}

	data, err = MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}
	searchRequest := data.(*bleve.SearchRequest)

	if searchRequest.Query.(*query.QueryStringQuery).Query != bleve.NewQueryStringQuery("blast").Query {
		t.Errorf("expected content to see %v, saw %v", bleve.NewQueryStringQuery("blast").Query, searchRequest.Query.(*query.QueryStringQuery).Query)
	}

	// test blast.SearchResult
	dataAny = &any.Any{
		TypeUrl: "bleve.SearchResult",
		Value:   []byte(`{"status":null,"request":null,"hits":null,"total_hits":10,"max_score":0,"took":0,"facets":null}`),
	}

	data, err = MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}
	searchResult := data.(*bleve.SearchResult)

	if searchResult.Total != 10 {
		t.Errorf("expected content to see %v, saw %v", 10, searchResult.Total)
	}

}
