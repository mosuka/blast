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
	"errors"
	"io"
	"io/ioutil"
	"log"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	pbindex "github.com/mosuka/blast/protobuf/index"
	blastraft "github.com/mosuka/blast/protobuf/raft"
)

type RaftFSM struct {
	index *Index

	metadata map[string]*blastraft.Node

	logger *log.Logger
}

func NewRaftFSM(path string, indexMapping *mapping.IndexMappingImpl, indexStorageType string, logger *log.Logger) (*RaftFSM, error) {
	index, err := NewIndex(path, indexMapping, indexStorageType, logger)
	if err != nil {
		return nil, err
	}

	return &RaftFSM{
		metadata: make(map[string]*blastraft.Node, 0),
		index:    index,
		logger:   logger,
	}, nil
}

func (f *RaftFSM) Close() error {
	err := f.index.Close()
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) Get(id string) (map[string]interface{}, error) {
	fields, err := f.index.Get(id)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func (f *RaftFSM) applyIndex(id string, fields map[string]interface{}) interface{} {
	f.logger.Printf("[DEBUG] index %s, %v", id, fields)

	err := f.index.Index(id, fields)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	return nil
}

func (f *RaftFSM) applyDelete(id string) interface{} {
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

func (f *RaftFSM) GetMetadata(nodeId string) (*blastraft.Node, error) {
	node, exists := f.metadata[nodeId]
	if !exists {
		return nil, blasterrors.ErrNotFound
	}
	if node == nil {
		return nil, errors.New("nil")
	}
	value := node

	return value, nil
}

func (f *RaftFSM) applySetMetadata(nodeId string, node *blastraft.Node) interface{} {
	f.metadata[nodeId] = node

	return nil
}

func (f *RaftFSM) applyDeleteMetadata(nodeId string) interface{} {
	_, exists := f.metadata[nodeId]
	if exists {
		delete(f.metadata, nodeId)
	}

	return nil
}

func (f *RaftFSM) Apply(l *raft.Log) interface{} {
	var c pbindex.IndexCommand
	err := proto.Unmarshal(l.Data, &c)
	if err != nil {
		return err
	}

	f.logger.Printf("[DEBUG] Apply %v", c)

	switch c.Type {
	case pbindex.IndexCommand_SET_METADATA:
		// Any -> Node
		nodeInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if nodeInstance == nil {
			return errors.New("nil")
		}
		metadata := nodeInstance.(*blastraft.Node)

		return f.applySetMetadata(metadata.Id, metadata)
	case pbindex.IndexCommand_DELETE_METADATA:
		// Any -> Node
		metadataInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if metadataInstance == nil {
			return errors.New("nil")
		}
		metadata := *metadataInstance.(*blastraft.Node)

		return f.applyDeleteMetadata(metadata.Id)
	case pbindex.IndexCommand_INDEX_DOCUMENT:
		// Any -> Document
		docInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if docInstance == nil {
			return errors.New("nil")
		}
		doc := *docInstance.(*pbindex.Document)

		// Any -> map[string]interface{}
		fieldsInstance, err := protobuf.MarshalAny(doc.Fields)
		if err != nil {
			return err
		}
		if fieldsInstance == nil {
			return errors.New("nil")
		}
		fields := *fieldsInstance.(*map[string]interface{})

		return f.applyIndex(doc.Id, fields)
	case pbindex.IndexCommand_DELETE_DOCUMENT:
		// Any -> Document
		docInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if docInstance == nil {
			return errors.New("nil")
		}
		doc := *docInstance.(*pbindex.Document)

		return f.applyDelete(doc.Id)
	default:
		return errors.New("command type not support")
	}
}

func (f *RaftFSM) Stats() (map[string]interface{}, error) {
	stats, err := f.index.Stats()
	if err != nil {
		return nil, err
	}

	return stats, nil
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

		// Any -> map[string]interface{}
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

		buff := proto.NewBuffer([]byte{})
		err := buff.EncodeMessage(doc)
		if err != nil {
			return err
		}

		_, err = sink.Write(buff.Bytes())
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
