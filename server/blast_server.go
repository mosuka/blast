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

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/coreos/etcd/clientv3"
	"github.com/mosuka/blast/proto"
	"github.com/mosuka/blast/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

var (
	STATE_ACTIVE = "active"
	STATE_READY  = "ready"
	STATE_DOWN   = "down"
)

type BlastServer struct {
	hostname           string
	port               int
	server             *grpc.Server
	service            *service.BlastService
	etcdClient         *clientv3.Client
	etcdKv             clientv3.KV
	etcdRequestTimeout int
	cluster            string
	shard              string
	node               string
}

func NewBlastServer(port int, indexPath string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{}, etcdEndpoints []string, etcdDialTimeout int, etcdRequestTimeout int, cluster string, shard string) (*BlastServer, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get the hostname")
		return nil, err
	}
	log.WithFields(log.Fields{
		"hostname": hostname,
	}).Info("got the hostname")

	node := fmt.Sprintf("%s:%d", hostname, port)
	log.WithFields(log.Fields{
		"node": node,
	}).Info("determine the node")

	var etcdClient *clientv3.Client
	if len(etcdEndpoints) > 0 {
		etcdConfig := clientv3.Config{
			Endpoints:   etcdEndpoints,
			DialTimeout: time.Duration(etcdDialTimeout) * time.Millisecond,
			Context:     context.Background(),
		}
		etcdClient, err = clientv3.New(etcdConfig)
		if err != nil {
			log.WithFields(log.Fields{
				"etcdConfig": etcdConfig,
				"error":      err.Error(),
			}).Error("failed to create the etcd client")
			return nil, err
		}
		log.WithFields(log.Fields{
			"etcdConfig": etcdConfig,
		}).Info("created the etcd client")
	}

	var etcdKv clientv3.KV
	if etcdClient != nil {
		etcdKv = clientv3.NewKV(etcdClient)
		log.Info("created the etcd kv")
	}

	if etcdKv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(etcdRequestTimeout)*time.Millisecond)
		defer cancel()

		// fetch index mapping
		keyIndexMapping := fmt.Sprintf("/blast/clusters/%s/indexMapping", cluster)
		kvresp, err := etcdKv.Get(ctx, keyIndexMapping)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the index mapping")

			return nil, err
		}
		log.Info("succeeded in fetch the index mapping")

		for _, ev := range kvresp.Kvs {
			err = json.Unmarshal(ev.Value, &indexMapping)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("failed to unmarshal the index mapping")

				return nil, err
			}
		}

		// fetch index type
		keyIndexType := fmt.Sprintf("/blast/clusters/%s/indexType", cluster)
		kvresp, err = etcdKv.Get(ctx, keyIndexType)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the index type")

			return nil, err
		}
		log.Info("succeeded in fetch the index type")

		for _, ev := range kvresp.Kvs {
			indexType = string(ev.Value)
		}

		// fetch kvstore
		keyKvstore := fmt.Sprintf("/blast/clusters/%s/kvstore", cluster)
		kvresp, err = etcdKv.Get(ctx, keyKvstore)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the kvstore")

			return nil, err
		}
		log.Info("succeeded in fetch the kvstore")

		for _, ev := range kvresp.Kvs {
			kvstore = string(ev.Value)
		}

		// fetch kvconfig
		keyKvconfig := fmt.Sprintf("/blast/clusters/%s/kvconfig", cluster)
		kvresp, err = etcdKv.Get(ctx, keyKvconfig)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the kvconfig")

			return nil, err
		}
		log.Info("succeeded in fetch the kvconfig")

		for _, ev := range kvresp.Kvs {
			err = json.Unmarshal(ev.Value, &kvconfig)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("failed to unmarshal the kvconfig")

				return nil, err
			}
		}
	}

	svr := grpc.NewServer()
	svc := service.NewBlastService(indexPath, indexMapping, indexType, kvstore, kvconfig)
	proto.RegisterIndexServer(svr, svc)

	return &BlastServer{
		hostname:           hostname,
		port:               port,
		server:             svr,
		service:            svc,
		etcdClient:         etcdClient,
		etcdKv:             etcdKv,
		etcdRequestTimeout: etcdRequestTimeout,
		cluster:            cluster,
		shard:              shard,
		node:               node,
	}, nil
}

func (s *BlastServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.WithFields(log.Fields{
			"port":  s.port,
			"error": err.Error(),
		}).Error("failed to create the listener")
		return err
	}
	log.WithFields(log.Fields{
		"port": s.port,
	}).Info("created the listener")

	go func() {
		s.service.OpenIndex()
		s.server.Serve(listener)
		return
	}()
	log.Info("the blast server has been started")

	if s.etcdClient != nil {
		go func() {
			for {
				s.watchCluster()
			}
			return
		}()
		log.Info("watching the cluster information has been started")

		err = s.joinCluster()
		if err != nil {
			log.WithFields(log.Fields{
				"cluster": s.cluster,
				"shard":   s.shard,
				"node":    s.node,
				"error":   err.Error(),
			}).Error("the blast server failed to join the cluster")
			return err
		}
		log.WithFields(log.Fields{
			"cluster": s.cluster,
			"shard":   s.shard,
			"node":    s.node,
		}).Info("the blast server has been joined the cluster")
	}

	return nil
}

func (s *BlastServer) Stop() error {
	if s.etcdClient != nil {
		err := s.leaveCluster()
		if err != nil {
			log.WithFields(log.Fields{
				"cluster": s.cluster,
				"shard":   s.shard,
				"node":    s.node,
				"error":   err.Error(),
			}).Error("the blast server failed to leave the cluster")
		}
		log.WithFields(log.Fields{
			"cluster": s.cluster,
			"shard":   s.shard,
			"node":    s.node,
		}).Info("the blast server has been left the cluster")

		err = s.etcdClient.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to disconnect from the etcd")
		}
		log.Info("disconnected from the etcd")
	}

	err := s.service.CloseIndex()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to close the index")
	}
	log.Info("closed the index")

	s.server.GracefulStop()

	log.Info("the blast server has been stopped")

	return err
}

func (s *BlastServer) watchCluster() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	keyCluster := fmt.Sprintf("/blast/clusters/%s", s.cluster)

	rch := s.etcdClient.Watch(ctx, keyCluster, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			log.WithFields(log.Fields{
				"type":  ev.Type,
				"key":   fmt.Sprintf("%s", ev.Kv.Key),
				"value": fmt.Sprintf("%s", ev.Kv.Value),
			}).Info("the cluster information has been changed")
		}
	}

	return
}

func (s *BlastServer) joinCluster() error {
	if s.cluster == "" {
		return fmt.Errorf("cluster is required")
	}
	if s.shard == "" {
		return fmt.Errorf("shard is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	keyNode := fmt.Sprintf("/blast/clusters/%s/shards/%s/nodes/%s", s.cluster, s.shard, s.node)

	_, err := s.etcdKv.Put(ctx, keyNode, STATE_READY)
	if err != nil {
		log.WithFields(log.Fields{
			"key":   keyNode,
			"value": STATE_READY,
			"error": err,
		}).Error("failed to put the data")
		return err
	}
	log.WithFields(log.Fields{
		"key":   keyNode,
		"value": STATE_READY,
	}).Info("put the data")

	return nil
}

func (s *BlastServer) leaveCluster() error {
	if s.cluster == "" {
		return fmt.Errorf("cluster is required")
	}
	if s.shard == "" {
		return fmt.Errorf("shard is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.etcdRequestTimeout)*time.Millisecond)
	defer cancel()

	keyNode := fmt.Sprintf("/blast/clusters/%s/shards/%s/nodes/%s", s.cluster, s.shard, s.node)

	_, err := s.etcdKv.Delete(ctx, keyNode)
	if err != nil {
		log.WithFields(log.Fields{
			"key":   keyNode,
			"error": err,
		}).Error("failed to delete the data")
		return err
	}
	log.WithFields(log.Fields{
		"key": keyNode,
	}).Info("deleted the data")

	return nil
}
