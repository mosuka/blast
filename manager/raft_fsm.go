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

	path string

	data objx.Map

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
	f.data = objx.New(map[string]interface{}{})
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
	safePath = strings.Join(strings.Split(safePath, "/"), objx.PathSeparator)

	return safePath
}

func (f *RaftFSM) normalize(value interface{}) interface{} {
	var ret interface{}

	switch value.(type) {
	case objx.Map:
		ret = map[string]interface{}{}
		for k, v := range value.(objx.Map) {
			ret.(map[string]interface{})[k] = f.normalize(v)
		}
	case map[string]interface{}:
		ret = map[string]interface{}{}
		for k, v := range value.(map[string]interface{}) {
			ret.(map[string]interface{})[k] = f.normalize(v)
		}
	case *map[string]interface{}:
		ret = map[string]interface{}{}
		for k, v := range *value.(*map[string]interface{}) {
			ret.(map[string]interface{})[k] = f.normalize(v)
		}
	case []interface{}:
		ret = []interface{}{}
		for _, v := range value.([]interface{}) {
			ret = append(ret.([]interface{}), f.normalize(v))
		}
	case bool, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, complex64, complex128:
		ret = value
	}

	return ret
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
	} else if value.IsMSI() {
		retValue = f.normalize(value.MSI())
	} else if value.IsObjxMap() {
		retValue = f.normalize(value.ObjxMap())
	} else if value.IsMSISlice() {
		retValue = f.normalize(value.MSISlice())
	} else if value.IsObjxMapSlice() {
		retValue = f.normalize(value.ObjxMapSlice())
	} else if value.IsInterSlice() {
		retValue = f.normalize(value.InterSlice())
	} else if value.IsInter() {
		retValue = value.Inter()
	} else if value.IsNil() {
		retValue = nil
	}

	return retValue, nil
}

func (f *RaftFSM) makeMap(path string, value interface{}) interface{} {
	var ret interface{}

	keys := f.pathKeys(path)

	if len(keys) >= 1 {
		ret = map[string]interface{}{keys[0]: f.makeMap(strings.Join(keys[1:], "/"), value)}
	} else if len(keys) == 0 {
		ret = value
	}

	return ret
}

func (f *RaftFSM) applySet(key string, value interface{}, merge bool) interface{} {
	if merge {
		f.data = f.data.Merge(objx.New(f.makeMap(key, f.normalize(value))))
	} else {
		if key == "/" {
			f.data = objx.New(f.normalize(value))
		} else {
			path := f.makeSafePath(key)
			if f.data.Has(path) {
				f.data.Set(path, objx.New(f.normalize(value)))
			} else {
				f.data = f.data.Merge(objx.New(f.makeMap(key, f.normalize(value))))
			}
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
	if key == "/" {
		f.data = objx.Map{}
		return nil
	}

	dataMap := f.normalize(f.data)

	// delete by key
	data, err := f.delete(f.pathKeys(key), dataMap)
	if err != nil {
		return err
	}

	f.data = objx.New(data)

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
		kvp := kvpInstance.(*management.KeyValuePair)

		// Any -> interface{}
		value, err := protobuf.MarshalAny(kvp.Value)

		return f.applySet(kvp.Key, value, kvp.Merge)
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
	data   objx.Map
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
