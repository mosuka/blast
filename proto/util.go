package proto

import (
	"encoding/json"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
	"reflect"
)

var (
	typeRegistry = make(map[string]reflect.Type)
)

func init() {
	typeRegistry["mapping.IndexMappingImpl"] = reflect.TypeOf(mapping.IndexMappingImpl{})
	typeRegistry["bleve.IndexStat"] = reflect.TypeOf(bleve.IndexStat{})
	typeRegistry["interface {}"] = reflect.TypeOf((map[string]interface{})(nil))
	typeRegistry["util.BulkRequest"] = reflect.TypeOf(BulkRequest{})
	typeRegistry["bleve.SearchRequest"] = reflect.TypeOf(bleve.SearchRequest{})
	typeRegistry["bleve.SearchResult"] = reflect.TypeOf(bleve.SearchResult{})
}

func MarshalAny(instance interface{}) (google_protobuf.Any, error) {
	var message google_protobuf.Any

	if instance == nil {
		return message, nil
	}

	value, err := json.Marshal(instance)
	if err != nil {
		return message, err
	}

	message.TypeUrl = reflect.TypeOf(instance).Elem().String()
	message.Value = value

	return message, nil
}

func UnmarshalAny(message *google_protobuf.Any) (interface{}, error) {
	if message == nil {
		return nil, nil
	}

	typeUrl := message.TypeUrl
	value := message.Value

	instance := reflect.New(typeRegistry[typeUrl]).Interface()

	err := json.Unmarshal(value, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}
