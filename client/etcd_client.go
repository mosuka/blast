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
	"github.com/coreos/etcd/clientv3/clientv3util"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var (
	STATE_ACTIVE = "active"
	STATE_DOWN   = "down"
)

type EtcdClient struct {
	endpoints      []string
	dialTimeout    int
	requestTimeout int
	client         *clientv3.Client
	kv             clientv3.KV
}

func NewEtcdClient(endpoints []string, dialTimeout int, requestTimeout int) (*EtcdClient, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(dialTimeout) * time.Millisecond,
		Context:     context.Background(),
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"endpoints":      endpoints,
			"dialTimeout":    dialTimeout,
			"requestTimeout": requestTimeout,
		}).Error("failed to connect etcd endpoints")

		return nil, fmt.Errorf("failed to connect etcd endpoints")
	}

	log.WithFields(log.Fields{
		"endpoints":      endpoints,
		"dialTimeout":    dialTimeout,
		"requestTimeout": requestTimeout,
	}).Info("succeeded in connect to etcd endpoints")

	return &EtcdClient{
		endpoints:      endpoints,
		dialTimeout:    dialTimeout,
		requestTimeout: requestTimeout,
		client:         client,
		kv:             clientv3.NewKV(client),
	}, nil
}

func (c *EtcdClient) AddCluster(clusterName string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyCluster := fmt.Sprintf("/blast/clusters/%s", clusterName)

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyCluster)).
			Then(clientv3.OpPut(keyCluster, STATE_DOWN)).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err := c.kv.Put(ctx, keyCluster, STATE_DOWN)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) RemoveCluster(clusterName string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyCluster := fmt.Sprintf("/blast/clusters/%s", clusterName)

	_, err := c.kv.Txn(ctx).
		If(clientv3util.KeyExists(keyCluster)).
		Then(clientv3.OpDelete(keyCluster)).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) PutNumberOfShards(clusterName string, numberOfShards int, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyNumberOfShards := fmt.Sprintf("/blast/clusters/%s/numberOfShards", clusterName)

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyNumberOfShards)).
			Then(clientv3.OpPut(keyNumberOfShards, strconv.Itoa(numberOfShards))).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err := c.kv.Put(ctx, keyNumberOfShards, strconv.Itoa(numberOfShards))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) GetNumberOfShards(clusterName string) (int, error) {
	if clusterName == "" {
		return 0, fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyNumberOfShards := fmt.Sprintf("/blast/clusters/%s/numberOfShards", clusterName)

	var numberOfShards int

	resp, err := c.kv.Get(ctx, keyNumberOfShards)
	if err != nil {
		return 0, err
	}
	for _, ev := range resp.Kvs {
		numberOfShards, err = strconv.Atoi(string(ev.Value))
		if err != nil {
			return 0, err
		}
	}

	if numberOfShards <= 0 {
		return 0, fmt.Errorf("numberOfShards is 0")
	}

	return numberOfShards, nil
}

func (c *EtcdClient) DeleteNumberOfShards(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyNumberOfShards := fmt.Sprintf("/blast/clusters/%s/numberOfShards", clusterName)

	_, err := c.kv.Delete(ctx, keyNumberOfShards)
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) PutIndexMapping(clusterName string, indexMapping *mapping.IndexMappingImpl, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/indexMapping", clusterName)

	var bytesIndexMapping []byte
	var err error
	if indexMapping != nil {
		bytesIndexMapping, err = json.Marshal(indexMapping)
		if err != nil {
			return err
		}
	}

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyIndexMapping)).
			Then(clientv3.OpPut(keyIndexMapping, string(bytesIndexMapping))).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err := c.kv.Put(ctx, keyIndexMapping, string(bytesIndexMapping))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) GetIndexMapping(clusterName string) (*mapping.IndexMappingImpl, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/indexMapping", clusterName)

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

	if indexMapping == nil {
		return nil, fmt.Errorf("indexMapping is nil")
	}

	return indexMapping, nil
}

func (c *EtcdClient) DeleteIndexMapping(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/indexMapping", clusterName)

	_, err := c.kv.Txn(ctx).
		If(clientv3util.KeyExists(keyIndexMapping)).
		Then(clientv3.OpDelete(keyIndexMapping)).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) PutIndexType(clusterName string, indexType string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyIndexType := fmt.Sprintf("/blast/clusters/%s/indexType", clusterName)

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyIndexType)).
			Then(clientv3.OpPut(keyIndexType, indexType)).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err := c.kv.Put(ctx, keyIndexType, indexType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) GetIndexType(clusterName string) (string, error) {
	if clusterName == "" {
		return "", fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyIndexType := fmt.Sprintf("/blast/clusters/%s/indexType", clusterName)

	var indexType string

	resp, err := c.kv.Get(ctx, keyIndexType)
	if err != nil {
		return "", err
	}
	for _, ev := range resp.Kvs {
		indexType = string(ev.Value)
	}

	if indexType == "" {
		return "", fmt.Errorf("indexType is \"\"")
	}

	return indexType, nil
}

func (c *EtcdClient) DeleteIndexType(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyIndexType := fmt.Sprintf("/blast/clusters/%s/indexType", clusterName)

	_, err := c.kv.Txn(ctx).
		If(clientv3util.KeyExists(keyIndexType)).
		Then(clientv3.OpDelete(keyIndexType)).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) PutKvstore(clusterName string, kvstore string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", clusterName)

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyKvstore)).
			Then(clientv3.OpPut(keyKvstore, kvstore)).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err := c.kv.Put(ctx, keyKvstore, kvstore)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) GetKvstore(clusterName string) (string, error) {
	if clusterName == "" {
		return "", fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", clusterName)

	var kvstore string

	resp, err := c.kv.Get(ctx, keyKvstore)
	if err != nil {
		return "", err
	}
	for _, ev := range resp.Kvs {
		kvstore = string(ev.Value)
	}

	if kvstore == "" {
		return "", fmt.Errorf("kvstore is \"\"")
	}

	return kvstore, nil
}

func (c *EtcdClient) DeleteKvstore(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", clusterName)

	_, err := c.kv.Txn(ctx).
		If(clientv3util.KeyExists(keyKvstore)).
		Then(clientv3.OpDelete(keyKvstore)).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) PutKvconfig(clusterName string, kvconfig map[string]interface{}, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", clusterName)

	var bytesKvconfig []byte
	var err error
	if kvconfig != nil {
		bytesKvconfig, err = json.Marshal(kvconfig)
		if err != nil {
			return err
		}
	}

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyKvconfig)).
			Then(clientv3.OpPut(keyKvconfig, string(bytesKvconfig))).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err = c.kv.Put(ctx, keyKvconfig, string(bytesKvconfig))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) GetKvconfig(clusterName string) (map[string]interface{}, error) {
	if clusterName == "" {
		return nil, fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", clusterName)

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

	if kvconfig == nil {
		return nil, fmt.Errorf("kvconfig is nil")
	}

	return kvconfig, nil
}

func (c *EtcdClient) DeleteKvconfig(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", clusterName)

	_, err := c.kv.Txn(ctx).
		If(clientv3util.KeyExists(keyKvconfig)).
		Then(clientv3.OpDelete(keyKvconfig)).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) AddShard(clusterName string, shardName string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}
	if shardName == "" {
		return fmt.Errorf("shardName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyShard := fmt.Sprintf("/blast/clusters/%s/shards/%s", clusterName, shardName)

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyShard)).
			Then(clientv3.OpPut(keyShard, STATE_DOWN)).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err := c.kv.Put(ctx, keyShard, STATE_DOWN)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) RemoveShard(clusterName string, shardName string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}
	if shardName == "" {
		return fmt.Errorf("shardName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyShard := fmt.Sprintf("/blast/clusters/%s/shards/%s", clusterName, shardName)

	_, err := c.kv.Txn(ctx).
		If(clientv3util.KeyExists(keyShard)).
		Then(clientv3.OpDelete(keyShard)).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) AddNode(clusterName string, shardName string, nodeName string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}
	if shardName == "" {
		return fmt.Errorf("shardName is required")
	}
	if nodeName == "" {
		return fmt.Errorf("nodeName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyNode := fmt.Sprintf("/blast/clusters/%s/shards/%s/nodes/%s", clusterName, shardName, nodeName)

	if disableOverwrite {
		_, err := c.kv.Txn(ctx).
			If(clientv3util.KeyMissing(keyNode)).
			Then(clientv3.OpPut(keyNode, STATE_DOWN)).
			Commit()
		if err != nil {
			return err
		}
	} else {
		_, err := c.kv.Put(ctx, keyNode, STATE_DOWN)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *EtcdClient) RemoveNode(clusterName string, shardName string, nodeName string, disableOverwrite bool) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}
	if shardName == "" {
		return fmt.Errorf("shardName is required")
	}
	if nodeName == "" {
		return fmt.Errorf("nodeName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyNode := fmt.Sprintf("/blast/clusters/%s/shards/%s/nodes/%s", clusterName, shardName, nodeName)

	_, err := c.kv.Txn(ctx).
		If(clientv3util.KeyExists(keyNode)).
		Then(clientv3.OpDelete(keyNode)).
		Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *EtcdClient) Close() error {
	log.WithFields(log.Fields{
		"endpoints":      c.endpoints,
		"dialTimeout":    c.dialTimeout,
		"requestTimeout": c.requestTimeout,
	}).Info("disconnect etcd endpoints")

	return c.client.Close()
}
