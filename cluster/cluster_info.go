package cluster

import (
	"github.com/blevesearch/bleve/mapping"
)

type ClusterInfo struct {
	Name         string                    `json:"name,omitempty"`
	Shards       int                       `json:"shards,omitempty"`
	IndexPath    string                    `json:"index_path,omitempty"`
	IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
	IndexType    string                    `json:"index_type,omitempty"`
	Kvstore      string                    `json:"kvstore,omitempty"`
	Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
}
