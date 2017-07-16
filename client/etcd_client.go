//  Copyright (c) 2017 Minoru Osuka
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

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/coreos/etcd/clientv3"
	"strconv"
	"time"
)

type EtcdClientWrapper struct {
	client         *clientv3.Client
	kv             clientv3.KV
	requestTimeout int
}

func NewEtcdClientWrapper(endpoints []string, requestTimeout int) (*EtcdClientWrapper, error) {
	if len(endpoints) <= 0 {
		return nil, nil
	}

	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		Context:     context.Background(),
	}
	c, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &EtcdClientWrapper{
		client:         c,
		kv:             clientv3.NewKV(c),
		requestTimeout: requestTimeout,
	}, nil
}

func (c *EtcdClientWrapper) PutCluster(name string, shards int, indexPath string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{}) error {
	var err error

	err = c.PutShards(name, shards)
	if err != nil {
		return err
	}

	err = c.PutIndexPath(name, indexPath)
	if err != nil {
		return err
	}

	err = c.PutIndexMapping(name, indexMapping)
	if err != nil {
		return err
	}

	err = c.PutIndexType(name, indexType)
	if err != nil {
		return err
	}

	err = c.PutKvstore(name, kvstore)
	if err != nil {
		return err
	}

	err = c.PutKvconfig(name, kvconfig)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) GetCluster(name string) (interface{}, error) {
	shards, err := c.GetShards(name)
	if err != nil {
		return nil, err
	}

	indexPath, err := c.GetIndexPath(name)
	if err != nil {
		return nil, err
	}

	indexMapping, err := c.GetIndexMapping(name)
	if err != nil {
		return nil, err
	}

	indexType, err := c.GetIndexType(name)
	if err != nil {
		return nil, err
	}

	kvstore, err := c.GetKvstore(name)
	if err != nil {
		return nil, err
	}

	kvconfig, err := c.GetKvconfig(name)
	if err != nil {
		return nil, err
	}

	r := struct {
		Name         string                    `json:"name,omitempty"`
		Shards       int                       `json:"shards,omitempty"`
		IndexPath    string                    `json:"index_path,omitempty"`
		IndexMapping *mapping.IndexMappingImpl `json:"index_mapping,omitempty"`
		IndexType    string                    `json:"index_type,omitempty"`
		Kvstore      string                    `json:"kvstore,omitempty"`
		Kvconfig     map[string]interface{}    `json:"kvconfig,omitempty"`
	}{
		Name:         name,
		Shards:       shards,
		IndexPath:    indexPath,
		IndexMapping: indexMapping,
		IndexType:    indexType,
		Kvstore:      kvstore,
		Kvconfig:     kvconfig,
	}

	return r, nil
}

func (c *EtcdClientWrapper) DeleteCluster(name string) error {
	err := c.DeleteShards(name)
	if err != nil {
		return err
	}

	err = c.DeleteIndexPath(name)
	if err != nil {
		return err
	}

	err = c.DeleteIndexMapping(name)
	if err != nil {
		return err
	}

	err = c.DeleteIndexType(name)
	if err != nil {
		return err
	}

	err = c.DeleteKvstore(name)
	if err != nil {
		return err
	}

	err = c.DeleteKvconfig(name)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) PutShards(name string, shards int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyShards := fmt.Sprintf("%s/%s", keyRoot, "shards")

	_, err := c.kv.Put(ctx, keyShards, strconv.Itoa(shards))
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) GetShards(name string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyShards := fmt.Sprintf("%s/%s", keyRoot, "shards")

	var shards int

	resp, err := c.kv.Get(ctx, keyShards)
	if err != nil {
		return 0, err
	}
	for _, ev := range resp.Kvs {
		shards, err = strconv.Atoi(string(ev.Value))
		if err != nil {
			return 0, err
		}
	}

	return shards, nil
}

func (c *EtcdClientWrapper) DeleteShards(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyShards := fmt.Sprintf("%s/%s", keyRoot, "shards")

	_, err := c.kv.Delete(ctx, keyShards)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) PutIndexPath(name string, indexPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexPath := fmt.Sprintf("%s/%s", keyRoot, "indexPath")

	_, err := c.kv.Put(ctx, keyIndexPath, indexPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) GetIndexPath(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexPath := fmt.Sprintf("%s/%s", keyRoot, "indexPath")

	var indexPath string

	resp, err := c.kv.Get(ctx, keyIndexPath)
	if err != nil {
		return "", err
	}
	for _, ev := range resp.Kvs {
		indexPath = string(ev.Value)
	}

	return indexPath, nil
}

func (c *EtcdClientWrapper) DeleteIndexPath(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexPath := fmt.Sprintf("%s/%s", keyRoot, "indexPath")

	_, err := c.kv.Delete(ctx, keyIndexPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) PutIndexMapping(name string, indexMapping *mapping.IndexMappingImpl) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexMapping := fmt.Sprintf("%s/%s", keyRoot, "indexMapping")

	var bytesIndexMapping []byte
	var err error
	if indexMapping != nil {
		bytesIndexMapping, err = json.Marshal(indexMapping)
		if err != nil {
			return err
		}
	}

	_, err = c.kv.Put(ctx, keyIndexMapping, string(bytesIndexMapping))
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) GetIndexMapping(name string) (*mapping.IndexMappingImpl, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexMapping := fmt.Sprintf("%s/%s", keyRoot, "indexMapping")

	var indexMapping *mapping.IndexMappingImpl

	resp, err := c.kv.Get(ctx, keyIndexMapping)
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		err = json.Unmarshal(ev.Value, &indexMapping)
		if err != nil {
			return nil, err
		}
	}

	return indexMapping, nil
}

func (c *EtcdClientWrapper) DeleteIndexMapping(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexMapping := fmt.Sprintf("%s/%s", keyRoot, "indexMapping")

	_, err := c.kv.Delete(ctx, keyIndexMapping)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) PutIndexType(name string, indexType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexType := fmt.Sprintf("%s/%s", keyRoot, "indexType")

	_, err := c.kv.Put(ctx, keyIndexType, indexType)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) GetIndexType(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexType := fmt.Sprintf("%s/%s", keyRoot, "indexType")

	var indexType string

	resp, err := c.kv.Get(ctx, keyIndexType)
	if err != nil {
		return "", err
	}
	for _, ev := range resp.Kvs {
		indexType = string(ev.Value)
	}

	return indexType, nil
}

func (c *EtcdClientWrapper) DeleteIndexType(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyIndexType := fmt.Sprintf("%s/%s", keyRoot, "indexType")

	_, err := c.kv.Delete(ctx, keyIndexType)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) PutKvstore(name string, kvstore string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyKvstore := fmt.Sprintf("%s/%s", keyRoot, "kvstore")

	_, err := c.kv.Put(ctx, keyKvstore, kvstore)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) GetKvstore(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyKvstore := fmt.Sprintf("%s/%s", keyRoot, "kvstore")

	var kvstore string

	resp, err := c.kv.Get(ctx, keyKvstore)
	if err != nil {
		return "", err
	}
	for _, ev := range resp.Kvs {
		kvstore = string(ev.Value)
	}

	return kvstore, nil
}

func (c *EtcdClientWrapper) DeleteKvstore(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyKvstore := fmt.Sprintf("%s/%s", keyRoot, "kvstore")

	_, err := c.kv.Delete(ctx, keyKvstore)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) PutKvconfig(name string, kvconfig map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyKvconfig := fmt.Sprintf("%s/%s", keyRoot, "kvconfig")

	var bytesKvconfig []byte
	var err error
	if kvconfig != nil {
		bytesKvconfig, err = json.Marshal(kvconfig)
		if err != nil {
			return err
		}
	}

	_, err = c.kv.Put(ctx, keyKvconfig, string(bytesKvconfig))
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) GetKvconfig(name string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyKvconfig := fmt.Sprintf("%s/%s", keyRoot, "kvconfig")

	var kvconfig map[string]interface{}

	resp, err := c.kv.Get(ctx, keyKvconfig)
	if err != nil {
		return nil, err
	}
	for _, ev := range resp.Kvs {
		err = json.Unmarshal(ev.Value, &kvconfig)
		if err != nil {
			return nil, err
		}
	}

	return kvconfig, nil
}

func (c *EtcdClientWrapper) DeleteKvconfig(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyRoot := fmt.Sprintf("/blast/%s", name)
	keyKvconfig := fmt.Sprintf("%s/%s", keyRoot, "kvconfig")

	_, err := c.kv.Delete(ctx, keyKvconfig)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClientWrapper) Close() error {
	return c.client.Close()
}
