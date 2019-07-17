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

	"github.com/blevesearch/bleve"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/config"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/maputils"
	"github.com/mosuka/blast/protobuf"
	"go.uber.org/zap"
)

type RaftFSM struct {
	path        string
	indexConfig *config.IndexConfig
	logger      *zap.Logger

	metadata      maputils.Map
	metadataMutex sync.RWMutex

	index *Index
}

func NewRaftFSM(path string, indexConfig *config.IndexConfig, logger *zap.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		path:        path,
		indexConfig: indexConfig,
		logger:      logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	var err error

	f.metadata = maputils.Map{}

	f.index, err = NewIndex(f.path, f.indexConfig, f.logger)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

func (f *RaftFSM) Stop() error {
	err := f.index.Close()
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

func (f *RaftFSM) GetNodeConfig(nodeId string) (map[string]interface{}, error) {
	f.metadataMutex.RLock()
	defer f.metadataMutex.RUnlock()

	nodeConfig, err := f.metadata.Get(nodeId)
	if err != nil {
		f.logger.Error(err.Error(), zap.String("node_id", nodeId))
		if err == maputils.ErrNotFound {
			return nil, blasterrors.ErrNotFound
		}
		return nil, err
	}

	return nodeConfig.(maputils.Map).ToMap(), nil
}

func (f *RaftFSM) SetNodeConfig(nodeId string, nodeConfig map[string]interface{}) error {
	f.metadataMutex.RLock()
	defer f.metadataMutex.RUnlock()

	err := f.metadata.Merge(nodeId, nodeConfig)
	if err != nil {
		f.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))
		return err
	}

	return nil
}

func (f *RaftFSM) DeleteNodeConfig(nodeId string) error {
	f.metadataMutex.RLock()
	defer f.metadataMutex.RUnlock()

	err := f.metadata.Delete(nodeId)
	if err != nil {
		f.logger.Error(err.Error(), zap.String("node_id", nodeId))
		return err
	}

	return nil
}

func (f *RaftFSM) GetDocument(id string) (map[string]interface{}, error) {
	fields, err := f.index.Get(id)
	if err != nil {
		f.logger.Error(err.Error())
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
		err = f.SetNodeConfig(data["node_id"].(string), data["node_config"].(map[string]interface{}))
		return &fsmResponse{error: err}
	case deleteNode:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.DeleteNodeConfig(data["node_id"].(string))
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
		doc := &protobuf.Document{}
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
