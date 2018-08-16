//  Copyright (c) 2018 Minoru Osuka
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

package bleve

import (
	"log"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/boltdb"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/logging"
)

type IndexConfig struct {
	Path         string                    `json:"path,omitempty"`
	IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
	IndexType    string                    `json:"index_type,omitempty"`
	Kvstore      string                    `json:"kvstore,omitempty"`
	Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
}

func DefaultConfig() *IndexConfig {
	return &IndexConfig{
		Path:         "./data/index",
		IndexMapping: mapping.NewIndexMapping(),
		IndexType:    upsidedown.Name,
		Kvstore:      boltdb.Name,
		Kvconfig: map[string]interface{}{
			"create_if_missing": true,
			"error_if_exists":   false,
		},
	}
}

type Index struct {
	index  bleve.Index
	logger *log.Logger
}

func NewIndex(config *IndexConfig) (*Index, error) {
	var err error

	var idx bleve.Index
	if _, err = os.Stat(config.Path); os.IsNotExist(err) {
		if idx, err = bleve.NewUsing(config.Path, config.IndexMapping, config.IndexType, config.Kvstore, config.Kvconfig); err != nil {
			return nil, err
		}
	} else {
		if idx, err = bleve.OpenUsing(config.Path, config.Kvconfig); err != nil {
			return nil, err
		}
	}

	return &Index{
		index:  idx,
		logger: logging.DefaultLogger(),
	}, nil
}

func (i *Index) SetLogger(logger *log.Logger) {
	i.logger = logger
	return
}

func (i *Index) Searcher() (*Searcher, error) {
	var err error

	var searcher *Searcher
	if searcher, err = NewSearcher(i); err != nil {
		i.logger.Printf("[ERR] bleve: Failed to create searcher: %s", err.Error())
		return nil, err
	}

	return searcher, nil
}

func (i *Index) Indexer() (*Indexer, error) {
	var err error

	var indexer *Indexer
	if indexer, err = NewIndexer(i); err != nil {
		i.logger.Printf("[ERR] bleve: Failed to create indexer: %s", err.Error())
		return nil, err
	}

	return indexer, nil
}

func (i *Index) Bulker(batchSize int) (*Bulker, error) {
	var err error

	var bulker *Bulker
	if bulker, err = NewBulker(i, batchSize); err != nil {
		i.logger.Printf("[ERR] bleve: Failed to create bulker: %v", err)
		return nil, err
	}

	return bulker, nil
}

func (i *Index) Close() error {
	if err := i.index.Close(); err != nil {
		i.logger.Printf("[ERR] bleve: Failed to close index: %s", err.Error())
		return err
	}

	return nil
}
