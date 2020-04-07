package marshaler

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/protobuf"
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

	// test kvs.Node
	node := &protobuf.Node{
		RaftAddress: ":7000",
		State:       "Leader",
		Metadata: &protobuf.Metadata{
			GrpcAddress: ":9000",
			HttpAddress: ":8000",
		},
	}

	nodeAny := &any.Any{}
	err = UnmarshalAny(node, nodeAny)
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedType = "protobuf.Node"
	actualType = nodeAny.TypeUrl
	if expectedType != actualType {
		t.Errorf("expected content to see %s, saw %s", expectedType, actualType)
	}

	expectedValue = []byte(`{"raft_address":":7000","metadata":{"grpc_address":":9000","http_address":":8000"},"state":"Leader"}`)
	actualValue = nodeAny.Value
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

	// raft.Node
	dataAny = &any.Any{
		TypeUrl: "protobuf.Node",
		Value:   []byte(`{"raft_address":":7000","metadata":{"grpc_address":":9000","http_address":":8000"},"state":"Leader"}`),
	}

	data, err = MarshalAny(dataAny)
	if err != nil {
		t.Errorf("%v", err)
	}
	node := data.(*protobuf.Node)

	if node.RaftAddress != ":7000" {
		t.Errorf("expected content to see %v, saw %v", ":7000", node.RaftAddress)
	}
	if node.Metadata.GrpcAddress != ":9000" {
		t.Errorf("expected content to see %v, saw %v", ":9000", node.Metadata.GrpcAddress)
	}
	if node.Metadata.HttpAddress != ":8000" {
		t.Errorf("expected content to see %v, saw %v", ":8000", node.Metadata.HttpAddress)
	}
	if node.State != "Leader" {
		t.Errorf("expected content to see %v, saw %v", "Leader", node.State)
	}
}
