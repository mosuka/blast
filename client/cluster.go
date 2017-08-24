package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/coreos/etcd/clientv3"
)

type Cluster interface {
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
	Close() error
}

type cluster struct {
	client *clientv3.Client
}

func NewCluster(cfg clientv3.Config) (Cluster, error) {
	c, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	return &cluster{
		client: c,
	}, nil
}

func (c *cluster) PutIndexMapping(ctx context.Context, cluster string, indexMapping *mapping.IndexMappingImpl) error {
	keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/index_mapping", cluster)

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

func (c *cluster) GetIndexMapping(ctx context.Context, cluster string) (*mapping.IndexMappingImpl, error) {
	keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/index_mapping", cluster)

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

	return indexMapping, nil
}

func (c *cluster) DeleteIndexMapping(ctx context.Context, cluster string) error {
	keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/index_mapping", cluster)

	_, err := c.client.Delete(ctx, keyIndexMapping)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) PutIndexType(ctx context.Context, cluster string, indexType string) error {
	keyIndexType := fmt.Sprintf("/blast/clusters/%s/index_type", cluster)

	_, err := c.client.Put(ctx, keyIndexType, indexType)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) GetIndexType(ctx context.Context, cluster string) (string, error) {
	keyIndexType := fmt.Sprintf("/blast/clusters/%s/index_type", cluster)

	var indexType string

	kvresp, err := c.client.Get(ctx, keyIndexType)
	if err != nil {
		return "", err
	}
	for _, ev := range kvresp.Kvs {
		indexType = string(ev.Value)
	}

	return indexType, nil
}

func (c *cluster) DeleteIndexType(ctx context.Context, cluster string) error {
	keyIndexType := fmt.Sprintf("/blast/clusters/%s/index_type", cluster)

	_, err := c.client.Delete(ctx, keyIndexType)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) PutKvstore(ctx context.Context, cluster string, kvstore string) error {
	keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", cluster)

	_, err := c.client.Put(ctx, keyKvstore, kvstore)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) GetKvstore(ctx context.Context, cluster string) (string, error) {
	keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", cluster)

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

func (c *cluster) DeleteKvstore(ctx context.Context, cluster string) error {
	keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", cluster)

	_, err := c.client.Delete(ctx, keyKvstore)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) PutKvconfig(ctx context.Context, cluster string, kvconfig map[string]interface{}) error {
	keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", cluster)

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

func (c *cluster) GetKvconfig(ctx context.Context, cluster string) (map[string]interface{}, error) {
	keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", cluster)

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

func (c *cluster) DeleteKvconfig(ctx context.Context, cluster string) error {
	keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", cluster)

	_, err := c.client.Delete(ctx, keyKvconfig)
	if err != nil {
		return err
	}

	return nil
}

func (c *cluster) Close() error {
	return c.client.Close()
}
