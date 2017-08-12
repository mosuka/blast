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

type EtcdClientWrapper struct {
	client         *clientv3.Client
	kv             clientv3.KV
	requestTimeout int
}

func NewEtcdClientWrapper(endpoints []string, dialTimeout int, requestTimeout int) (*EtcdClientWrapper, error) {
	if len(endpoints) <= 0 {
		return nil, fmt.Errorf("endpoints are required")
	}

	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(dialTimeout) * time.Millisecond,
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

//func (c *EtcdClientWrapper) AddWorker(clusterName string, name string, disableOverwrite bool) error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
//	defer cancel()
//
//	keyRoot := fmt.Sprintf("/blast/%s", clusterName)
//	keyWorkers := fmt.Sprintf("%s/%s", keyRoot, "workers")
//
//	if disableOverwrite {
//		_, err := c.kv.Txn(ctx).
//			If(clientv3util.KeyMissing(keyWorkers)).
//			Then(clientv3.OpPut(keyWorkers, strconv.Itoa(shards))).
//			Commit()
//		if err != nil {
//			return err
//		}
//	} else {
//		_, err := c.kv.Put(ctx, keyWorkers, strconv.Itoa(shards))
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

func (c *EtcdClientWrapper) RemoveWorker(clusterName string, name string, disableOverwrite bool) error {

	return nil
}

func (c *EtcdClientWrapper) GetWorkers(clusterName string) error {

	return nil
}

func (c *EtcdClientWrapper) PutNumberOfShards(clusterName string, numberOfShards int, disableOverwrite bool) error {
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

func (c *EtcdClientWrapper) GetNumberOfShards(clusterName string) (int, error) {
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

func (c *EtcdClientWrapper) DeleteNumberOfShards(clusterName string) error {
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

func (c *EtcdClientWrapper) PutIndexMapping(clusterName string, indexMapping *mapping.IndexMappingImpl, disableOverwrite bool) error {
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

func (c *EtcdClientWrapper) GetIndexMapping(clusterName string) (*mapping.IndexMappingImpl, error) {
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

func (c *EtcdClientWrapper) DeleteIndexMapping(clusterName string) error {
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

func (c *EtcdClientWrapper) PutIndexType(clusterName string, indexType string, disableOverwrite bool) error {
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

func (c *EtcdClientWrapper) GetIndexType(clusterName string) (string, error) {
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

func (c *EtcdClientWrapper) DeleteIndexType(clusterName string) error {
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

func (c *EtcdClientWrapper) PutKvstore(clusterName string, kvstore string, disableOverwrite bool) error {
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

func (c *EtcdClientWrapper) GetKvstore(clusterName string) (string, error) {
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

func (c *EtcdClientWrapper) DeleteKvstore(clusterName string) error {
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

func (c *EtcdClientWrapper) PutKvconfig(clusterName string, kvconfig map[string]interface{}, disableOverwrite bool) error {
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

func (c *EtcdClientWrapper) GetKvconfig(clusterName string) (map[string]interface{}, error) {
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

func (c *EtcdClientWrapper) DeleteKvconfig(clusterName string) error {
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

func (c *EtcdClientWrapper) Watch(clusterName string) error {
	if clusterName == "" {
		return fmt.Errorf("clusterName is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.requestTimeout)*time.Millisecond)
	defer cancel()

	keyCluster := fmt.Sprintf("/blast/clusters/%s", clusterName)

	rch := c.client.Watch(ctx, keyCluster, clientv3.WithPrefix())

	for wresp := range rch {
		for _, ev := range wresp.Events {
			log.WithFields(log.Fields{
				"type":  ev.Type,
				"key":   string(ev.Kv.Key),
				"value": string(ev.Kv.Value),
			}).Info("information changes in cluster has detected")
		}
	}

	return nil
}

func (c *EtcdClientWrapper) Close() error {
	return c.client.Close()
}
