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
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	blasterrors "github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/protobuf/management"
	blastraft "github.com/mosuka/blast/protobuf/raft"
	"github.com/stretchr/objx"
)

type RaftFSM struct {
	cluster *blastraft.Cluster

	data objx.Map

	logger *log.Logger
}

func NewRaftFSM(path string, logger *log.Logger) (*RaftFSM, error) {
	return &RaftFSM{
		cluster: &blastraft.Cluster{Nodes: make([]*blastraft.Node, 0)},
		data:    objx.New(map[string]interface{}{}),
		logger:  logger,
	}, nil
}

func (f *RaftFSM) Close() error {
	return nil
}

func (f *RaftFSM) GetMetadata(node *blastraft.Node) (*blastraft.Node, error) {
	for _, n := range f.cluster.Nodes {
		if n.Id == node.Id {
			return n, nil
		}
	}

	return nil, blasterrors.ErrNotFound
}

func (f *RaftFSM) applySetMetadata(node *blastraft.Node) interface{} {
	for _, n := range f.cluster.Nodes {
		if n.Id == node.Id {
			return errors.New("already exists")
		}
	}

	f.cluster.Nodes = append(f.cluster.Nodes, node)

	return nil
}

func (f *RaftFSM) applyDeleteMetadata(node *blastraft.Node) interface{} {
	for i, n := range f.cluster.Nodes {
		if n.Id == node.Id {
			return append(f.cluster.Nodes[:i], f.cluster.Nodes[i+1:]...)
		}
	}

	return blasterrors.ErrNotFound
}

func (f *RaftFSM) pathKeys(path string) []string {
	keys := make([]string, 0)
	for _, k := range strings.Split(path, "/") {
		if k != "" {
			keys = append(keys, k)
		}
	}

	return keys
}

func (f *RaftFSM) makeSafePath(path string) string {
	keys := f.pathKeys(path)

	safePath := strings.Join(keys, "/")
	//safePath = strings.Join(strings.Split(safePath, "."), "_")
	safePath = strings.Join(strings.Split(safePath, "/"), objx.PathSeparator)

	return safePath
}

func (f *RaftFSM) Get(key string) (interface{}, error) {
	var value *objx.Value

	if key == "/" {
		value = f.data.Value()
	} else {
		path := f.makeSafePath(key)
		if f.data.Has(path) {
			value = f.data.Get(path)
		} else {
			return nil, blasterrors.ErrNotFound
		}
	}

	var retValue interface{}
	if value.IsBool() {
		retValue = value.Bool()
	} else if value.IsBoolSlice() {
		retValue = value.BoolSlice()
	} else if value.IsComplex64() {
		retValue = value.Complex64()
	} else if value.IsComplex64Slice() {
		retValue = value.Complex64Slice()
	} else if value.IsComplex128() {
		retValue = value.Complex128()
	} else if value.IsComplex128Slice() {
		retValue = value.Complex128Slice()
	} else if value.IsFloat32() {
		retValue = value.Float32()
	} else if value.IsFloat32Slice() {
		retValue = value.Float32Slice()
	} else if value.IsFloat64() {
		retValue = value.Float64()
	} else if value.IsFloat64Slice() {
		retValue = value.Float64Slice()
	} else if value.IsInt() {
		retValue = value.Int()
	} else if value.IsIntSlice() {
		retValue = value.IntSlice()
	} else if value.IsInt8() {
		retValue = value.IsInt8()
	} else if value.IsInt8Slice() {
		retValue = value.Int8Slice()
	} else if value.IsInt16() {
		retValue = value.Int16()
	} else if value.IsInt16Slice() {
		retValue = value.Int16Slice()
	} else if value.IsInt32() {
		retValue = value.Int32()
	} else if value.IsInt32Slice() {
		retValue = value.Int32Slice()
	} else if value.IsInt64() {
		retValue = value.IsInt64()
	} else if value.IsInt64Slice() {
		retValue = value.Int64Slice()
	} else if value.IsStr() {
		retValue = value.Str()
	} else if value.IsStrSlice() {
		retValue = value.StrSlice()
	} else if value.IsUint() {
		retValue = value.Uint()
	} else if value.IsUintSlice() {
		retValue = value.UintSlice()
	} else if value.IsUint8() {
		retValue = value.Uint8()
	} else if value.IsUint8Slice() {
		retValue = value.Uint8Slice()
	} else if value.IsUint16() {
		retValue = value.Uint16()
	} else if value.IsUint16Slice() {
		retValue = value.Uint16Slice()
	} else if value.IsUint32() {
		retValue = value.Uint32()
	} else if value.IsUint32Slice() {
		retValue = value.Uint32Slice()
	} else if value.IsUint64() {
		retValue = value.Uint64()
	} else if value.IsUint64Slice() {
		retValue = value.Uint64Slice()
	} else if value.IsUintptr() {
		retValue = value.Uintptr()
	} else if value.IsUintptrSlice() {
		retValue = value.UintptrSlice()
	} else if value.IsMSI() || value.IsObjxMap() {
		b, err := json.Marshal(value.MSI())
		if err != nil {
			return nil, err
		}
		var mm map[string]interface{}
		err = json.Unmarshal(b, &mm)
		if err != nil {
			return nil, err
		}
		retValue = mm
	} else if value.IsMSISlice() || value.IsObjxMapSlice() {
		b, err := json.Marshal(value.MSISlice())
		if err != nil {
			return nil, err
		}
		var mm []map[string]interface{}
		err = json.Unmarshal(b, &mm)
		if err != nil {
			return nil, err
		}
		retValue = mm
	} else if value.IsInterSlice() {
		retValue = value.InterSlice()
	} else if value.IsInter() {
		retValue = value.Inter()
	} else if value.IsNil() {
		retValue = nil
	}

	return retValue, nil
}

func (f *RaftFSM) applySet(key string, value interface{}) interface{} {
	// convert to JSON string
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	valueMap, err := objx.FromJSON(string(valueJSON))
	if err != nil {
		return err
	}

	if key == "/" {
		f.data = valueMap
	} else {
		f.data.Set(f.makeSafePath(key), valueMap)
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
				// key exists
				if len(keys) > 1 {
					data.(map[string]interface{})[key], err = f.delete(keys[1:], data.(map[string]interface{})[key])
					if err != nil {
						return nil, err
					}
				} else {
					//
					mm := data.(map[string]interface{})
					delete(mm, key)
					data = mm
				}
			} else {
				// key does not exist
				return nil, blasterrors.ErrNotFound
			}
		}
	}

	return data, nil
}

func (f *RaftFSM) applyDelete(key string) interface{} {
	if key == "/" {
		f.data = objx.Map{}
		return nil
	}

	// get data as map[string]interface{}
	dataJSON, err := json.Marshal(&f.data)
	if err != nil {
		return err
	}
	var dataMap map[string]interface{}
	err = json.Unmarshal(dataJSON, &dataMap)
	if err != nil {
		return err
	}

	// delete by key
	data, err := f.delete(f.pathKeys(key), dataMap)
	if err != nil {
		return err
	}

	// interface{} -> JSON
	dataJSON, err = json.Marshal(&data)
	if err != nil {
		return err
	}

	f.data, err = objx.FromJSON(string(dataJSON))
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
		// Any -> Node
		nodeInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if nodeInstance == nil {
			return errors.New("nil")
		}
		node := nodeInstance.(*blastraft.Node)

		return f.applySetMetadata(node)
	case management.ManagementCommand_DELETE_METADATA:
		// Any -> Node
		nodeInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if nodeInstance == nil {
			return errors.New("nil")
		}
		node := nodeInstance.(*blastraft.Node)

		return f.applyDeleteMetadata(node)
	case management.ManagementCommand_PUT_KEY_VALUE_PAIR:
		// Any -> KeyValuePair
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
		// Any -> KeyValuePair
		kvpInstance, err := protobuf.MarshalAny(c.Data)
		if err != nil {
			return err
		}
		if kvpInstance == nil {
			return errors.New("nil")
		}
		kvp := kvpInstance.(*management.KeyValuePair)

		return f.applyDelete(kvp.Key)
	default:
		return errors.New("command type not support")
	}
}

func (f *RaftFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &KVSFSMSnapshot{
		federation: f.data,
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

	err = json.Unmarshal(data, &f.data)
	if err != nil {
		f.logger.Printf("[ERR] %v", err)
		return err
	}

	f.logger.Printf("[INFO] federation was restored: %v", f.data)

	return nil
}

type KVSFSMSnapshot struct {
	federation objx.Map
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
