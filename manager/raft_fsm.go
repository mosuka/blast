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

	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/maputils"
	"go.uber.org/zap"
)

type RaftFSM struct {
	path   string
	logger *zap.Logger

	metadata maputils.Map
	data     maputils.Map
}

func NewRaftFSM(path string, logger *zap.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		path:   path,
		logger: logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	f.logger.Info("initialize metadata")
	f.metadata = maputils.Map{}

	f.logger.Info("initialize store data")
	f.data = maputils.Map{}

	return nil
}

func (f *RaftFSM) Stop() error {
	return nil
}

func (f *RaftFSM) GetNodeConfig(nodeId string) (map[string]interface{}, error) {
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

func (f *RaftFSM) applySetNodeConfig(nodeId string, nodeConfig map[string]interface{}) error {
	err := f.metadata.Merge(nodeId, nodeConfig)
	if err != nil {
		f.logger.Error(err.Error(), zap.String("node_id", nodeId), zap.Any("node_config", nodeConfig))
		return err
	}

	return nil
}

func (f *RaftFSM) applyDeleteNodeConfig(nodeId string) error {
	err := f.metadata.Delete(nodeId)
	if err != nil {
		f.logger.Error(err.Error(), zap.String("node_id", nodeId))
		return err
	}

	return nil
}

func (f *RaftFSM) Get(key string) (interface{}, error) {
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

func (f *RaftFSM) applySet(key string, value interface{}, merge bool) error {
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

func (f *RaftFSM) delete(keys []string, data interface{}) (interface{}, error) {
	var err error

	if len(keys) >= 1 {
		key := keys[0]

		switch data.(type) {
		case map[string]interface{}:
			if _, exist := data.(map[string]interface{})[key]; exist {
				if len(keys) > 1 {
					data.(map[string]interface{})[key], err = f.delete(keys[1:], data.(map[string]interface{})[key])
					if err != nil {
						return nil, err
					}
				} else {
					mm := data.(map[string]interface{})
					delete(mm, key)
					data = mm
				}
			} else {
				return nil, blasterrors.ErrNotFound
			}
		}
	}

	return data, nil
}

//func (f *RaftFSM) applyDelete(key string) interface{} {
func (f *RaftFSM) applyDelete(key string) error {
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
		err = f.applySetNodeConfig(data["node_id"].(string), data["node_config"].(map[string]interface{}))
		return &fsmResponse{error: err}
	case deleteNode:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.applyDeleteNodeConfig(data["node_id"].(string))
		return &fsmResponse{error: err}
	case setKeyValue:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.applySet(data["key"].(string), data["value"], true)
		return &fsmResponse{error: err}
	case deleteKeyValue:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			f.logger.Error(err.Error())
			return &fsmResponse{error: err}
		}
		err = f.applyDelete(data["key"].(string))
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
