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

func (m *Document) GetFieldsActual() map[string]interface{} {
	if m == nil {
		return nil
	}

	fields, err := UnmarshalAny(m.Fields)
	if err != nil {
		return nil
	}

	if fields == nil {
		return nil
	}

	return *fields.(*map[string]interface{})
}

func (m *Document) SetFieldsActual(fields map[string]interface{}) {
	if m != nil {
		fieldsAny, err := MarshalAny(fields)
		if err == nil {
			m.Fields = &fieldsAny
		}
	}
}

func (m *GetIndexResponse) GetIndexMappingActual() *mapping.IndexMappingImpl {
	if m == nil {
		return nil
	}

	indexMapping, err := UnmarshalAny(m.IndexMapping)
	if err != nil {
		return nil
	}

	if indexMapping == nil {
		return nil
	}

	return indexMapping.(*mapping.IndexMappingImpl)
}

func (m *GetIndexResponse) SetIndexMappingActual(indexMapping *mapping.IndexMappingImpl) {
	if m != nil {
		indexMapingAny, err := MarshalAny(indexMapping)
		if err == nil {
			m.IndexMapping = &indexMapingAny
		}
	}
}

func (m *GetIndexResponse) GetKvconfigActual() *map[string]interface{} {
	if m == nil {
		return nil
	}

	kvconfig, err := UnmarshalAny(m.Kvconfig)
	if err != nil {
		return nil
	}

	if kvconfig == nil {
		return nil
	}

	return kvconfig.(*map[string]interface{})
}

func (m *GetIndexResponse) SetKvconfigActual(kvconfig map[string]interface{}) {
	if m != nil {
		kvconfigAny, err := MarshalAny(kvconfig)
		if err == nil {
			m.Kvconfig = &kvconfigAny
		}
	}
}

func (m *SearchRequest) GetSearchRequestActual() *bleve.SearchRequest {
	if m == nil {
		return nil
	}

	searchRequest, err := UnmarshalAny(m.SearchRequest)
	if err != nil {
		return nil
	}

	if searchRequest == nil {
		return nil
	}

	return searchRequest.(*bleve.SearchRequest)
}

func (m *SearchRequest) SetSearchRequestActual(searchRequest *bleve.SearchRequest) {
	if m != nil {
		searchRequestAny, err := MarshalAny(searchRequest)
		if err == nil {
			m.SearchRequest = &searchRequestAny
		}
	}
}

func (m *SearchResponse) GetSearchResultActual() *bleve.SearchResult {
	if m == nil {
		return nil
	}

	searchResult, err := UnmarshalAny(m.SearchResult)
	if err != nil {
		return nil
	}

	if searchResult == nil {
		return nil
	}

	return searchResult.(*bleve.SearchResult)
}

func (m *SearchResponse) SetSearchResultActual(searchResult *bleve.SearchResult) {
	if m != nil {
		searchResultAny, err := MarshalAny(searchResult)
		if err == nil {
			m.SearchResult = &searchResultAny
		}
	}
}
