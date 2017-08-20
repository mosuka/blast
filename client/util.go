package client

import (
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
)

type GetDocumentResponse struct {
	Id        string                 `json:"id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Succeeded bool                   `json:"succeeded"`
	Message   string                 `json:"message,omitempty"`
}

type PutDocumentResponse struct {
	Id        string                 `json:"id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Succeeded bool                   `json:"succeeded"`
	Message   string                 `json:"message,omitempty"`
}

type DeleteDocumentResponse struct {
	Id        string `json:"id,omitempty"`
	Succeeded bool   `json:"succeeded"`
	Message   string `json:"message,omitempty"`
}

type BulkResponse struct {
	PutCount      int32  `json:"put_count"`
	PutErrorCount int32  `json:"put_error_count"`
	DeleteCount   int32  `json:"delete_count"`
	Succeeded     bool   `json:"succeeded"`
	Message       string `json:"message,omitempty"`
}

type SearchResponse struct {
	SearchResult *bleve.SearchResult `json:"search_result"`
	Succeeded    bool                `json:"succeeded"`
	Message      string              `json:"message,omitempty"`
}

type GetIndexResponse struct {
	IndexPath    string                    `json:"index_path,omitempty"`
	IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
	IndexType    string                    `json:"index_type,omitempty"`
	Kvstore      string                    `json:"kvstore,omitempty"`
	Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
	Succeeded    bool                      `json:"succeeded"`
	Message      string                    `json:"message,omitempty"`
}
