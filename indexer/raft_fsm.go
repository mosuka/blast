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
	"sync"

	"github.com/blevesearch/bleve/mapping"

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/index"
	"go.uber.org/zap"
)

type RaftFSM struct {
	path             string
	indexMapping     *mapping.IndexMappingImpl
	indexType        string
	indexStorageType string
	logger           *zap.Logger

	cluster      *index.Cluster
	clusterMutex sync.RWMutex

	index *Index
}

func NewRaftFSM(path string, indexMapping *mapping.IndexMappingImpl, indexType string, indexStorageType string, logger *zap.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		path:             path,
		indexMapping:     indexMapping,
		indexType:        indexType,
		indexStorageType: indexStorageType,
		logger:           logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	f.logger.Info("initialize cluster")
	f.cluster = &index.Cluster{Nodes: make(map[string]*index.Node, 0)}

	f.logger.Info("initialize index")
	var err error
	f.index, err = NewIndex(f.path, f.indexMapping, f.indexType, f.indexStorageType, f.logger)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

func (f *RaftFSM) Stop() error {
	f.logger.Info("close index")
	err := f.index.Close()
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

func (f *RaftFSM) GetNode(nodeId string) (*index.Node, error) {
	f.clusterMutex.RLock()
	defer f.clusterMutex.RUnlock()

	node, ok := f.cluster.Nodes[nodeId]
	if !ok {
		return nil, blasterrors.ErrNotFound
	}

	return node, nil
}

func (f *RaftFSM) SetNode(node *index.Node) error {
	f.clusterMutex.RLock()
	defer f.clusterMutex.RUnlock()

	f.cluster.Nodes[node.Id] = node

	return nil
}

func (f *RaftFSM) DeleteNode(nodeId string) error {
	f.clusterMutex.RLock()
	defer f.clusterMutex.RUnlock()

	if _, ok := f.cluster.Nodes[nodeId]; !ok {
		return blasterrors.ErrNotFound
	}

	delete(f.cluster.Nodes, nodeId)

	return nil
}

func (f *RaftFSM) GetDocument(id string) (map[string]interface{}, error) {
	fields, err := f.index.Get(id)
	if err != nil {
		switch err {
		case blasterrors.ErrNotFound:
			f.logger.Debug(err.Error(), zap.String("id", id))
		default:
			f.logger.Error(err.Error(), zap.String("id", id))
		}
		return nil, err
	}

	return fields, nil
}

func (f *RaftFSM) IndexDocument(id string, fields map[string]interface{}) error {
	err := f.index.Index(id, fields)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

func (f *RaftFSM) IndexDocuments(docs []map[string]interface{}) (int, error) {
	count, err := f.index.BulkIndex(docs)
	if err != nil {
		f.logger.Error(err.Error())
		return -1, err
	}

	return count, nil
}

func (f *RaftFSM) DeleteDocument(id string) error {
	err := f.index.Delete(id)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

func (f *RaftFSM) DeleteDocuments(ids []string) (int, error) {
	count, err := f.index.BulkDelete(ids)
	if err != nil {
		f.logger.Error(err.Error())
		return -1, err
	}

	return count, nil
}

func (f *RaftFSM) Search(request *bleve.SearchRequest) (*bleve.SearchResult, error) {
	result, err := f.index.Search(request)
	if err != nil {
		f.logger.Error(err.Error())
		return nil, err
	}

	return result, nil
}

func (f *RaftFSM) GetIndexConfig() (map[string]interface{}, error) {
	return f.index.Config()
}

func (f *RaftFSM) GetIndexStats() (map[string]interface{}, error) {
	return f.index.Stats()
}

type fsmResponse struct {
	error error
}

type fsmIndexDocumentResponse struct {
	count int
	error error
}

type fsmDeleteDocumentResponse struct {
	count int
	error error
}

func (f *RaftFSM) Apply(l *raft.Log) interface{} {
	var msg message
	err := json.Unmarshal(l.Data, &msg)
	if err != nil {
		return err
	}

	switch msg.Command {
	case setNode:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		b, err := json.Marshal(data["node"])
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		var node *index.Node
		err = json.Unmarshal(b, &node)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.SetNode(node)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		return &fsmResponse{error: err}
	case deleteNode:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.DeleteNode(data["id"].(string))
		return &fsmResponse{error: err}
	case indexDocument:
		var data []map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmIndexDocumentResponse{count: -1, error: err}
		}
		count, err := f.IndexDocuments(data)
		return &fsmIndexDocumentResponse{count: count, error: err}
	case deleteDocument:
		var data []string
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmDeleteDocumentResponse{count: -1, error: err}
		}
		count, err := f.DeleteDocuments(data)
		return &fsmDeleteDocumentResponse{count: count, error: err}
	default:
		err = errors.New("unsupported command")
		f.logger.Error(err.Error())
		return &fsmResponse{error: err}
	}
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.logger.Info("snapshot")

	return &RaftFSMSnapshot{
		index:  f.index,
		logger: f.logger,
	}, nil
}

func (f *RaftFSM) Restore(rc io.ReadCloser) error {
	f.logger.Info("restore")

	defer func() {
		err := rc.Close()
		if err != nil {
			f.logger.Error(err.Error())
		}
	}()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	docCount := 0

	buff := proto.NewBuffer(data)
	for {
		doc := &index.Document{}
		err = buff.DecodeMessage(doc)
		if err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			f.logger.Error(err.Error())
			continue
		}

		fields, err := protobuf.MarshalAny(doc.Fields)
		if err != nil {
			f.logger.Error(err.Error())
			continue
		}
		if fields == nil {
			f.logger.Error("value is nil")
			continue
		}
		fieldsMap := *fields.(*map[string]interface{})

		err = f.index.Index(doc.Id, fieldsMap)
		if err != nil {
			f.logger.Error(err.Error())
			continue
		}

		docCount = docCount + 1
	}

	f.logger.Info("restore", zap.Int("count", docCount))

	return nil
}

// ---------------------

type RaftFSMSnapshot struct {
	index  *Index
	logger *zap.Logger
}

func (f *RaftFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	f.logger.Info("persist")

	defer func() {
		err := sink.Close()
		if err != nil {
			f.logger.Error(err.Error())
		}
	}()

	ch := f.index.SnapshotItems()

	docCount := 0

	for {
		doc := <-ch
		if doc == nil {
			break
		}

		docBytes, err := json.Marshal(doc)
		if err != nil {
			f.logger.Error(err.Error())
			continue
		}

		_, err = sink.Write(docBytes)
		if err != nil {
			f.logger.Error(err.Error())
			continue
		}

		docCount = docCount + 1
	}

	f.logger.Info("persist", zap.Int("count", docCount))

	return nil
}

func (f *RaftFSMSnapshot) Release() {
	f.logger.Info("release")
}
