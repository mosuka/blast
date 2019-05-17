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

package indexer

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/maputils"
	"github.com/mosuka/blast/protobuf"
	pbindex "github.com/mosuka/blast/protobuf/index"
)

type RaftFSM struct {
	cluster maputils.Map

	path string

	index *Index

	indexConfig map[string]interface{}

	logger *log.Logger
}

func NewRaftFSM(path string, indexConfig map[string]interface{}, logger *log.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		path:        path,
		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	var err error

	f.cluster = maputils.Map{}

	f.index, err = NewIndex(f.path, f.indexConfig, f.logger)
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) Stop() error {
	err := f.index.Close()
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) GetNode(id string) (map[string]interface{}, error) {
	value, err := f.cluster.Get(id)
	if err != nil {
		return nil, err
	}

	return value.(maputils.Map).ToMap(), nil
}

func (f *RaftFSM) applySetNode(id string, value map[string]interface{}) interface{} {
	err := f.cluster.Merge(id, value)
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) applyDeleteNode(id string) interface{} {
	err := f.cluster.Delete(id)
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) GetDocument(id string) (map[string]interface{}, error) {
	fields, err := f.index.Get(id)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func (f *RaftFSM) applyIndexDocument(id string, fields map[string]interface{}) interface{} {
	f.logger.Printf("[DEBUG] index %s, %v", id, fields)

	err := f.index.Index(id, fields)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	return nil
}

func (f *RaftFSM) applyDeleteDocument(id string) interface{} {
	err := f.index.Delete(id)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	return nil
}

func (f *RaftFSM) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := f.index.Search(request)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (f *RaftFSM) Apply(l *raft.Log) interface{} {
	var msg message
	err := json.Unmarshal(l.Data, &msg)
	if err != nil {
		return err
	}

	f.logger.Printf("[DEBUG] Apply %v", msg)

	switch msg.Command {
	case setNode:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			return err
		}
		return f.applySetNode(data["id"].(string), data["metadata"].(map[string]interface{}))
	case deleteNode:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			return err
		}
		return f.applyDeleteNode(data["id"].(string))
	case indexDocument:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			return err
		}
		return f.applyIndexDocument(data["id"].(string), data["fields"].(map[string]interface{}))
	case deleteDocument:
		var data string
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			return err
		}
		return f.applyDeleteDocument(data)
	default:
		return errors.New("command type not support")
	}
}

func (f *RaftFSM) GetIndexConfig() (map[string]interface{}, error) {
	return f.index.Config()
}

func (f *RaftFSM) GetIndexStats() (map[string]interface{}, error) {
	return f.index.Stats()
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &IndexFSMSnapshot{
		index:  f.index,
		logger: f.logger,
	}, nil
}

func (f *RaftFSM) Restore(rc io.ReadCloser) error {
	defer func() {
		err := rc.Close()
		if err != nil {
			f.logger.Printf("[ERR] %v", err)
		}
	}()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	docCount := 0

	buff := proto.NewBuffer(data)
	for {
		doc := &pbindex.Document{}
		err = buff.DecodeMessage(doc)
		if err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			f.logger.Printf("[ERR] %v", err)
			return err
		}

		fields, err := protobuf.MarshalAny(doc.Fields)
		if err != nil {
			return err
		}
		if fields == nil {
			return nil
		}
		fieldsMap := *fields.(*map[string]interface{})

		err = f.index.Index(doc.Id, fieldsMap)
		if err != nil {
			f.logger.Printf("[ERR] %v", err)
			return err
		}

		f.logger.Printf("[DEBUG] restore %v %v", doc.Id, doc.Fields)
		docCount = docCount + 1
	}

	f.logger.Printf("[INFO] %d documents were restored", docCount)

	return nil
}

// ---------------------

type IndexFSMSnapshot struct {
	index  *Index
	logger *log.Logger
}

func (f *IndexFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	defer func() {
		err := sink.Close()
		if err != nil {
			f.logger.Printf("[ERR] %v", err)
		}
	}()

	ch := f.index.SnapshotItems()

	docCount := 0

	for {
		doc := <-ch
		if doc == nil {
			break
		}

		docCount = docCount + 1
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return err
		}

		_, err = sink.Write(docBytes)
		if err != nil {
			return err
		}
	}
	f.logger.Printf("[INFO] %d documents were persisted", docCount)

	return nil
}

func (f *IndexFSMSnapshot) Release() {
	f.logger.Printf("[INFO] release")
}
