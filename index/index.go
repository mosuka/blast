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

package index

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/boltdb"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/mapping"
	blastlog "github.com/mosuka/blast/log"
)

const (
	DefaultDir              = "/tmp/blast/data/index"
	DefaultIndexMappingFile = ""
	DefaultIndexType        = upsidedown.Name
	DefaultKvstore          = boltdb.Name
)

type IndexConfig struct {
	Dir          string                    `json:"dir,omitempty"`
	IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
	IndexType    string                    `json:"index_type,omitempty"`
	Kvstore      string                    `json:"kvstore,omitempty"`
	Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
}

func DefaultIndexConfig() *IndexConfig {
	return &IndexConfig{
		Dir:          DefaultDir,
		IndexMapping: mapping.NewIndexMapping(),
		IndexType:    DefaultIndexType,
		Kvstore:      DefaultKvstore,
		Kvconfig: map[string]interface{}{
			"create_if_missing": true,
			"error_if_exists":   true,
		},
	}
}

func (c *IndexConfig) SetIndexMapping(indexMappingFile string) error {
	var err error

	f, err := os.Open(indexMappingFile)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, c.IndexMapping)
	if err != nil {
		return err
	}

	return nil
}

func NewIndexMapping(file string) (*mapping.IndexMappingImpl, error) {
	var err error

	m := mapping.NewIndexMapping()

	if file == "" {
		return m, nil
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type Index struct {
	index  bleve.Index
	logger *log.Logger
}

func NewIndex(config *IndexConfig) (*Index, error) {
	var err error

	var idx bleve.Index
	if _, err = os.Stat(config.Dir); os.IsNotExist(err) {
		if idx, err = bleve.NewUsing(config.Dir, config.IndexMapping, config.IndexType, config.Kvstore, config.Kvconfig); err != nil {
			return nil, err
		}
	} else {
		if idx, err = bleve.OpenUsing(config.Dir, config.Kvconfig); err != nil {
			return nil, err
		}
	}

	return &Index{
		index:  idx,
		logger: blastlog.DefaultLogger(),
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
