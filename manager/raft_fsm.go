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
	"log"

	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/maputils"
)

type RaftFSM struct {
	cluster maputils.Map

	path string

	data maputils.Map

	logger *log.Logger
}

func NewRaftFSM(path string, logger *log.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		path:   path,
		logger: logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	f.logger.Print("[INFO] initialize data")
	f.cluster = maputils.Map{}
	f.data = maputils.Map{}
	return nil
}

func (f *RaftFSM) Stop() error {
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

func (f *RaftFSM) Get(key string) (interface{}, error) {
	value, err := f.data.Get(key)
	if err != nil {
		switch err {
		case maputils.ErrNotFound:
			return nil, blasterrors.ErrNotFound
		default:
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

func (f *RaftFSM) applySet(key string, value interface{}, merge bool) interface{} {
	if merge {
		err := f.data.Merge(key, value)
		if err != nil {
			return err
		}
	} else {
		err := f.data.Set(key, value)
		if err != nil {
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

func (f *RaftFSM) applyDelete(key string) interface{} {
	err := f.data.Delete(key)
	if err != nil {
		if err == maputils.ErrNotFound {
			return blasterrors.ErrNotFound
		}
		return err
	}

	return nil
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
	case setKeyValue:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			return err
		}

		return f.applySet(data["key"].(string), data["value"], true)
	case deleteKeyValue:
		var data map[string]interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			return err
		}

		return f.applyDelete(data["key"].(string))
	default:
		return errors.New("command type not support")
	}
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &RaftFSMSnapshot{
		data:   f.data,
		logger: f.logger,
	}, nil
}

func (f *RaftFSM) Restore(rc io.ReadCloser) error {
	f.logger.Print("[INFO] restore data")

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

	err = json.Unmarshal(data, &f.data)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	return nil
}

type RaftFSMSnapshot struct {
	data   maputils.Map
	logger *log.Logger
}

func (f *RaftFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	f.logger.Printf("[INFO] persist data")

	defer func() {
		err := sink.Close()
		if err != nil {
			f.logger.Printf("[ERR] %v", err)
		}
	}()

	buff, err := json.Marshal(f.data)
	if err != nil {
		return err
	}

	_, err = sink.Write(buff)
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSMSnapshot) Release() {
	f.logger.Printf("[INFO] release")
}
