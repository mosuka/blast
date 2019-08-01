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

package manager

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"sync"

	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/maputils"
	"github.com/mosuka/blast/protobuf/management"
	"go.uber.org/zap"
)

type RaftFSM struct {
	path   string
	logger *zap.Logger

	cluster      *management.Cluster
	clusterMutex sync.RWMutex

	data      maputils.Map
	dataMutex sync.RWMutex
}

func NewRaftFSM(path string, logger *zap.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		path:   path,
		logger: logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	f.logger.Info("initialize metadata")
	f.cluster = &management.Cluster{Nodes: make(map[string]*management.Node, 0)}

	f.logger.Info("initialize store data")
	f.data = maputils.Map{}

	return nil
}

func (f *RaftFSM) Stop() error {
	return nil
}

func (f *RaftFSM) GetNode(nodeId string) (*management.Node, error) {
	f.clusterMutex.RLock()
	defer f.clusterMutex.RUnlock()

	node, ok := f.cluster.Nodes[nodeId]
	if !ok {
		return nil, blasterrors.ErrNotFound
	}

	return node, nil
}

func (f *RaftFSM) SetNode(node *management.Node) error {
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

func (f *RaftFSM) GetValue(key string) (interface{}, error) {
	value, err := f.data.Get(key)
	if err != nil {
		switch err {
		case maputils.ErrNotFound:
			f.logger.Debug("key does not found in the store data", zap.String("key", key))
			return nil, blasterrors.ErrNotFound
		default:
			f.logger.Error(err.Error(), zap.String("key", key))
			return nil, err
		}
	}

	var ret interface{}
	switch value.(type) {
	case maputils.Map:
		ret = value.(maputils.Map).ToMap()
	default:
		ret = value
	}

	return ret, nil
}

func (f *RaftFSM) SetValue(key string, value interface{}, merge bool) error {
	if merge {
		err := f.data.Merge(key, value)
		if err != nil {
			f.logger.Error(err.Error(), zap.String("key", key), zap.Any("value", value), zap.Bool("merge", merge))
			return err
		}
	} else {
		err := f.data.Set(key, value)
		if err != nil {
			f.logger.Error(err.Error(), zap.String("key", key), zap.Any("value", value), zap.Bool("merge", merge))
			return err
		}
	}

	return nil
}

func (f *RaftFSM) DeleteValue(key string) error {
	err := f.data.Delete(key)
	if err != nil {
		switch err {
		case maputils.ErrNotFound:
			f.logger.Debug("key does not found in the store data", zap.String("key", key))
			return blasterrors.ErrNotFound
		default:
			f.logger.Error(err.Error(), zap.String("key", key))
			return err
		}
	}

	return nil
}

type fsmResponse struct {
	error error
}

func (f *RaftFSM) Apply(l *raft.Log) interface{} {
	var msg message
	err := json.Unmarshal(l.Data, &msg)
	if err != nil {
		f.logger.Error(err.Error())
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
		var node *management.Node
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
	case setKeyValue:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.SetValue(data["key"].(string), data["value"], true)
		return &fsmResponse{error: err}
	case deleteKeyValue:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.DeleteValue(data["key"].(string))
		return &fsmResponse{error: err}
	default:
		err = errors.New("unsupported command")
		f.logger.Error(err.Error())
		return &fsmResponse{error: err}
	}
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.logger.Info("snapshot")

	return &RaftFSMSnapshot{
		data:   f.data,
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

	err = json.Unmarshal(data, &f.data)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

type RaftFSMSnapshot struct {
	data   maputils.Map
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

	buff, err := json.Marshal(f.data)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	_, err = sink.Write(buff)
	if err != nil {
		f.logger.Error(err.Error())
		return err
	}

	return nil
}

func (f *RaftFSMSnapshot) Release() {
	f.logger.Info("release")
}
