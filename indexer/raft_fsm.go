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
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	pbindex "github.com/mosuka/blast/protobuf/index"
	blastraft "github.com/mosuka/blast/protobuf/raft"
)

type RaftFSM struct {
	cluster *blastraft.Cluster

	path string

	index *Index

	indexConfig map[string]interface{}

	logger *log.Logger
}

func NewRaftFSM(path string, indexConfig map[string]interface{}, logger *log.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		cluster:     &blastraft.Cluster{Nodes: make(map[string]*blastraft.Metadata, 0)},
		path:        path,
		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	var err error

	f.index, err = NewIndex(f.path, f.indexConfig, f.logger)
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) Close() error {
	err := f.index.Close()
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) GetNode(node *blastraft.Node) (*blastraft.Node, error) {
	if _, exist := f.cluster.Nodes[node.Id]; exist {
		return &blastraft.Node{
			Id:       node.Id,
			Metadata: f.cluster.Nodes[node.Id],
		}, nil
	} else {
		return nil, blasterrors.ErrNotFound
	}
}

func (f *RaftFSM) applySetNode(node *blastraft.Node) interface{} {
	if _, exist := f.cluster.Nodes[node.Id]; !exist {
		f.cluster.Nodes[node.Id] = node.Metadata
		return nil
	} else {
		return errors.New("already exists")
	}
}

func (f *RaftFSM) applyDeleteNode(node *blastraft.Node) interface{} {
	if _, exist := f.cluster.Nodes[node.Id]; exist {
		delete(f.cluster.Nodes, node.Id)
		return nil
	} else {
		return blasterrors.ErrNotFound
	}
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

func (f *RaftFSM) Apply(l *raft.Log) interface{} {
	var c pbindex.IndexCommand
	err := proto.Unmarshal(l.Data, &c)
	if err != nil {
		return err
	}

	f.logger.Printf("[DEBUG] Apply %v", c)

	switch c.Type {
	case pbindex.IndexCommand_SET_NODE:
		// Any -> Node
		nodeInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if nodeInstance == nil {
			return errors.New("nil")
		}
		node := nodeInstance.(*blastraft.Node)

		return f.applySetNode(node)
	case pbindex.IndexCommand_DELETE_NODE:
		// Any -> Node
		nodeInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if nodeInstance == nil {
			return errors.New("nil")
		}
		node := nodeInstance.(*blastraft.Node)

		return f.applyDeleteNode(node)
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

func (f *RaftFSM) IndexConfig() (map[string]interface{}, error) {
	return f.index.Config()
}

func (f *RaftFSM) IndexStats() (map[string]interface{}, error) {
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
