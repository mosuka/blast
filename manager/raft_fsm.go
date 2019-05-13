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

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/maputils"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	blastraft "github.com/mosuka/blast/protobuf/raft"
)

type RaftFSM struct {
	cluster *blastraft.Cluster

	path string

	data maputils.Map

	logger *log.Logger
}

func NewRaftFSM(path string, logger *log.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		cluster: &blastraft.Cluster{Nodes: make(map[string]*blastraft.Metadata, 0)},
		path:    path,
		logger:  logger,
	}, nil
}

func (f *RaftFSM) Start() error {
	f.logger.Print("[INFO] initialize data")
	f.data = maputils.Map{}
	return nil
}

func (f *RaftFSM) Stop() error {
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

func (f *RaftFSM) Get(key string) (interface{}, error) {
	value, err := f.data.Get(key)
	if err != nil {
		return nil, err
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
		return err
	}

	return nil
}

func (f *RaftFSM) Apply(l *raft.Log) interface{} {
	var c management.ManagementCommand
	err := proto.Unmarshal(l.Data, &c)
	if err != nil {
		return err
	}

	f.logger.Printf("[DEBUG] Apply %v", c)

	switch c.Type {
	case management.ManagementCommand_SET_NODE:
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
	case management.ManagementCommand_DELETE_NODE:
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
	case management.ManagementCommand_PUT_KEY_VALUE_PAIR:
		// Any -> KeyValuePair
		kvpInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if kvpInstance == nil {
			return errors.New("nil")
		}
		kvp := *kvpInstance.(*management.KeyValuePair)

		// Any -> interface{}
		value, err := protobuf.MarshalAny(kvp.Value)

		v := value.(*map[string]interface{})

		return f.applySet(kvp.Key, *v, kvp.Merge)
	case management.ManagementCommand_DELETE_KEY_VALUE_PAIR:
		// Any -> KeyValuePair
		kvpInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if kvpInstance == nil {
			return errors.New("nil")
		}
		kvp := *kvpInstance.(*management.KeyValuePair)

		return f.applyDelete(kvp.Key)
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
