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
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	pbraft "github.com/mosuka/blast/protobuf/raft"
	"github.com/mosuka/maputils"
)

type RaftFSM struct {
	metadata map[string]*pbraft.Node

	federation map[string]interface{}

	logger *log.Logger
}

func NewRaftFSM(path string, logger *log.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		metadata:   make(map[string]*pbraft.Node, 0),
		federation: make(map[string]interface{}, 0),
		logger:     logger,
	}, nil
}

func (f *RaftFSM) Close() error {
	return nil
}

func (f *RaftFSM) GetMetadata(nodeId string) (*pbraft.Node, error) {
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

func (f *RaftFSM) applySetMetadata(nodeId string, node *pbraft.Node) interface{} {
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

func (f *RaftFSM) Get(key string) (interface{}, error) {
	nm, err := maputils.NewNestedMap(f.federation)
	if err != nil {
		return nil, err
	}

	value, err := nm.Get(key)
	if err == maputils.ErrNotFound {
		return nil, blasterrors.ErrNotFound
	}

	return value, nil
}

func (f *RaftFSM) applySet(key string, value interface{}) interface{} {
	nm, err := maputils.NewNestedMap(f.federation)
	if err != nil {
		return err
	}

	err = nm.Set(key, value)
	if err != nil {
		return err
	}

	return nil
}

func (f *RaftFSM) applyDelete(key string) interface{} {
	nm, err := maputils.NewNestedMap(f.federation)
	if err != nil {
		return err
	}

	err = nm.Delete(key)
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
	case management.ManagementCommand_SET_METADATA:
		// Any -> raft.Node
		nodeInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if nodeInstance == nil {
			return errors.New("nil")
		}
		node := nodeInstance.(*pbraft.Node)

		return f.applySetMetadata(node.Id, node)
	case management.ManagementCommand_DELETE_METADATA:
		// Any -> raft.Node
		nodeInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if nodeInstance == nil {
			return errors.New("nil")
		}
		node := *nodeInstance.(*pbraft.Node)

		return f.applyDeleteMetadata(node.Id)
	case management.ManagementCommand_PUT_KEY_VALUE_PAIR:
		// Any -> federation.Node
		kvpInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if kvpInstance == nil {
			return errors.New("nil")
		}
		kvp := kvpInstance.(*management.KeyValuePair)

		// Any -> interface{}
		value, err := protobuf.MarshalAny(kvp.Value)

		return f.applySet(kvp.Key, value)
	case management.ManagementCommand_DELETE_KEY_VALUE_PAIR:
		// Any -> federation.Node
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
	return &KVSFSMSnapshot{
		federation: f.federation,
		logger:     f.logger,
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

	err = json.Unmarshal(data, &f.federation)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	f.logger.Printf("[INFO] federation was restored: %v", f.federation)

	return nil
}

// ---------------------

type KVSFSMSnapshot struct {
	federation map[string]interface{}
	logger     *log.Logger
}

func (f *KVSFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	f.logger.Printf("[INFO] start data persistence")

	defer func() {
		err := sink.Close()
		if err != nil {
			f.logger.Printf("[ERR] %v", err)
		}
	}()

	buff, err := json.Marshal(f.federation)
	if err != nil {
		return err
	}

	_, err = sink.Write(buff)
	if err != nil {
		return err
	}

	f.logger.Printf("[INFO] federation was persisted: %v", f.federation)

	return nil
}

func (f *KVSFSMSnapshot) Release() {
	f.logger.Printf("[INFO] release")
}
