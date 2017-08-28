package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"
)

type BlastCluster interface {
	PutIndexMapping(ctx context.Context, cluster string, indexMapping *mapping.IndexMappingImpl) error
	GetIndexMapping(ctx context.Context, cluster string) (*mapping.IndexMappingImpl, error)
	DeleteIndexMapping(ctx context.Context, cluster string) error
	PutIndexType(ctx context.Context, cluster string, indexType string) error
	GetIndexType(ctx context.Context, cluster string) (string, error)
	DeleteIndexType(ctx context.Context, cluster string) error
	PutKvstore(ctx context.Context, cluster string, kvstore string) error
	GetKvstore(ctx context.Context, cluster string) (string, error)
	DeleteKvstore(ctx context.Context, cluster string) error
	PutKvconfig(ctx context.Context, cluster string, kvconfig map[string]interface{}) error
	GetKvconfig(ctx context.Context, cluster string) (map[string]interface{}, error)
	DeleteKvconfig(ctx context.Context, cluster string) error
	Watch(ctx context.Context, cluster string) error
	Close() error
}

type blastCluster struct {
	client *clientv3.Client
}

func NewBlastCluster(endpoints []string, dialTimeout int) (BlastCluster, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(dialTimeout) * time.Millisecond,
		Context:     context.Background(),
	}

	c, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &blastCluster{
		client: c,
	}, nil
}

func (c *blastCluster) PutIndexMapping(ctx context.Context, collection string, indexMapping *mapping.IndexMappingImpl) error {
	keyIndexMapping := fmt.Sprintf("/blast/cluster/collections/%s/index_mapping", collection)

	bytesIndexMapping, err := json.Marshal(indexMapping)
	if err != nil {
		return err
	}

	_, err = c.client.Put(ctx, keyIndexMapping, string(bytesIndexMapping))
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) GetIndexMapping(ctx context.Context, collection string) (*mapping.IndexMappingImpl, error) {
	keyIndexMapping := fmt.Sprintf("/blast/cluster/collections/%s/index_mapping", collection)

	var indexMapping *mapping.IndexMappingImpl

	kvresp, err := c.client.Get(ctx, keyIndexMapping)
	if err != nil {
		return nil, err
	}
	for _, ev := range kvresp.Kvs {
		err = json.Unmarshal(ev.Value, &indexMapping)
		if err != nil {
			return nil, err
		}
	}

	if indexMapping == nil {
		return nil, errors.New("index mapping does not exist")
	}

	return indexMapping, nil
}

func (c *blastCluster) DeleteIndexMapping(ctx context.Context, collection string) error {
	keyIndexMapping := fmt.Sprintf("/blast/cluster/collections/%s/index_mapping", collection)

	_, err := c.client.Delete(ctx, keyIndexMapping)
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) PutIndexType(ctx context.Context, collection string, indexType string) error {
	keyIndexType := fmt.Sprintf("/blast/cluster/collections/%s/index_type", collection)

	_, err := c.client.Put(ctx, keyIndexType, indexType)
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) GetIndexType(ctx context.Context, collection string) (string, error) {
	keyIndexType := fmt.Sprintf("/blast/cluster/collections/%s/index_type", collection)

	var indexType string

	kvresp, err := c.client.Get(ctx, keyIndexType)
	if err != nil {
		return "", err
	}
	for _, ev := range kvresp.Kvs {
		indexType = string(ev.Value)
	}

	if indexType == "" {
		return "", errors.New("index type does not exist")
	}

	return indexType, nil
}

func (c *blastCluster) DeleteIndexType(ctx context.Context, collection string) error {
	keyIndexType := fmt.Sprintf("/blast/cluster/collections/%s/index_type", collection)

	_, err := c.client.Delete(ctx, keyIndexType)
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) PutKvstore(ctx context.Context, collection string, kvstore string) error {
	keyKvstore := fmt.Sprintf("/blast/cluster/collections/%s/kvstore", collection)

	_, err := c.client.Put(ctx, keyKvstore, kvstore)
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) GetKvstore(ctx context.Context, collection string) (string, error) {
	keyKvstore := fmt.Sprintf("/blast/cluster/collections/%s/kvstore", collection)

	var kvstore string

	kvresp, err := c.client.Get(ctx, keyKvstore)
	if err != nil {
		return "", err
	}
	for _, ev := range kvresp.Kvs {
		kvstore = string(ev.Value)
	}

	return kvstore, nil
}

func (c *blastCluster) DeleteKvstore(ctx context.Context, collection string) error {
	keyKvstore := fmt.Sprintf("/blast/cluster/collections/%s/kvstore", collection)

	_, err := c.client.Delete(ctx, keyKvstore)
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) PutKvconfig(ctx context.Context, collection string, kvconfig map[string]interface{}) error {
	keyKvconfig := fmt.Sprintf("/blast/cluster/collections/%s/kvconfig", collection)

	bytesKvconfig, err := json.Marshal(kvconfig)
	if err != nil {
		return err
	}

	_, err = c.client.Put(ctx, keyKvconfig, string(bytesKvconfig))
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) GetKvconfig(ctx context.Context, collection string) (map[string]interface{}, error) {
	keyKvconfig := fmt.Sprintf("/blast/cluster/collections/%s/kvconfig", collection)

	var kvconfig map[string]interface{}

	kvresp, err := c.client.Get(ctx, keyKvconfig)
	if err != nil {
		return nil, err
	}
	for _, ev := range kvresp.Kvs {
		err = json.Unmarshal(ev.Value, &kvconfig)
		if err != nil {
			return nil, err
		}
	}

	return kvconfig, nil
}

func (c *blastCluster) DeleteKvconfig(ctx context.Context, collection string) error {
	keyKvconfig := fmt.Sprintf("/blast/cluster/collections/%s/kvconfig", collection)

	_, err := c.client.Delete(ctx, keyKvconfig)
	if err != nil {
		return err
	}

	return nil
}

func (c *blastCluster) Watch(ctx context.Context, collection string) error {
	keyCluster := fmt.Sprintf("/blast/cluster/collections/%s", collection)

	rch := c.client.Watch(ctx, keyCluster, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			log.WithFields(log.Fields{
				"type":  ev.Type,
				"key":   fmt.Sprintf("%s", ev.Kv.Key),
				"value": fmt.Sprintf("%s", ev.Kv.Value),
			}).Info("the cluster information has been changed")
		}
	}

	return nil
}

func (c *blastCluster) Close() error {
	return c.client.Close()
}
