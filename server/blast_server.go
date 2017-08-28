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
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/cluster"
	"github.com/mosuka/blast/proto"
	"github.com/mosuka/blast/service"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

//var (
//	STATE_ACTIVE = "active"
//	STATE_READY  = "ready"
//	STATE_DOWN   = "down"
//)

type BlastServer struct {
	cluster.BlastCluster

	hostname   string
	port       int
	server     *grpc.Server
	service    *service.BlastService
	collection string
	//shard              string
	//node               string
}

func NewBlastServer(port int,
	indexPath string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{},
	etcdEndpoints []string, etcdDialTimeout int, etcdRequestTimeout int, collection string) (*BlastServer, error) {

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

	var blastCluster cluster.BlastCluster
	if len(etcdEndpoints) > 0 {
		blastCluster, err = cluster.NewBlastCluster(etcdEndpoints, etcdDialTimeout)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to connect the cluster")
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(etcdRequestTimeout)*time.Millisecond)
		defer cancel()

		// fetch index mapping
		fetchedIndexMapping, err := blastCluster.GetIndexMapping(ctx, collection)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the index mapping")
			return nil, errors.New("failed to fetch the index mapping")
		}
		indexMapping = fetchedIndexMapping
		log.Info("succeeded in fetch the index mapping")

		// fetch index type
		fetchedIndexType, err := blastCluster.GetIndexType(ctx, collection)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the index type")
			return nil, errors.New("failed to fetch the index type")
		}
		indexType = fetchedIndexType
		log.Info("succeeded in fetch the index type")

		// fetch kvstore
		fetchedKvstore, err := blastCluster.GetKvstore(ctx, collection)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the kvstore")
			return nil, errors.New("failed to fetch the kvstore")
		}
		kvstore = fetchedKvstore
		log.Info("succeeded in fetch the kvstore")

		// fetch kvconfig
		fetchedKvconfig, err := blastCluster.GetKvconfig(ctx, collection)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to fetch the kvconfig")
			return nil, errors.New("failed to fetch the kvconfig")
		}
		kvconfig = fetchedKvconfig
		log.Info("succeeded in fetch the kvconfig")
	}

	svr := grpc.NewServer()
	svc := service.NewBlastService(indexPath, indexMapping, indexType, kvstore, kvconfig)
	proto.RegisterIndexServer(svr, svc)

	return &BlastServer{
		BlastCluster: blastCluster,
		hostname:     hostname,
		port:         port,
		server:       svr,
		service:      svc,
		collection:   collection,
	}, nil
}

func (s *BlastServer) Start() error {
	// create listener
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

	// start server
	go func() {
		s.service.OpenIndex()
		s.server.Serve(listener)
		return
	}()
	log.Info("server has been started")

	//s.BlastCluster.Watch();

	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15000)*time.Millisecond)
			defer cancel()

			s.BlastCluster.Watch(ctx, s.collection)
		}
		return
	}()
	log.Info("watching the cluster information has been started")

	//if s.etcdClient != nil {
	//	// start watching
	//	go func() {
	//		for {
	//			s.watchCluster()
	//		}
	//		return
	//	}()
	//	log.Info("watching the cluster information has been started")
	//
	//	// join cluster
	//	err = s.joinCluster()
	//	if err != nil {
	//		log.WithFields(log.Fields{
	//			"cluster": s.cluster,
	//			"shard":   s.shard,
	//			"node":    s.node,
	//			"error":   err.Error(),
	//		}).Error("the blast server failed to join the cluster")
	//		return err
	//	}
	//	log.WithFields(log.Fields{
	//		"cluster": s.cluster,
	//		"shard":   s.shard,
	//		"node":    s.node,
	//	}).Info("the blast server has been joined the cluster")
	//}

	return nil
}

func (s *BlastServer) Stop() error {
	if s.BlastCluster != nil {
		err := s.BlastCluster.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to disconnect the cluster")
		}
	}

	//if s.etcdClient != nil {
	//	err := s.leaveCluster()
	//	if err != nil {
	//		log.WithFields(log.Fields{
	//			"cluster": s.cluster,
	//			"shard":   s.shard,
	//			"node":    s.node,
	//			"error":   err.Error(),
	//		}).Error("the blast server failed to leave the cluster")
	//	}
	//	log.WithFields(log.Fields{
	//		"cluster": s.cluster,
	//		"shard":   s.shard,
	//		"node":    s.node,
	//	}).Info("the blast server has been left the cluster")
	//
	//	err = s.etcdClient.Close()
	//	if err != nil {
	//		log.WithFields(log.Fields{
	//			"error": err.Error(),
	//		}).Error("failed to disconnect from the etcd")
	//	}
	//	log.Info("disconnected from the etcd")
	//}

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

//func (s *BlastServer) watchCluster() {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.etcdRequestTimeout)*time.Millisecond)
//	defer cancel()
//
//	keyCluster := fmt.Sprintf("/blast/clusters/%s", s.cluster)
//
//	rch := s.etcdClient.Watch(ctx, keyCluster, clientv3.WithPrefix())
//	for wresp := range rch {
//		for _, ev := range wresp.Events {
//			log.WithFields(log.Fields{
//				"type":  ev.Type,
//				"key":   fmt.Sprintf("%s", ev.Kv.Key),
//				"value": fmt.Sprintf("%s", ev.Kv.Value),
//			}).Info("the cluster information has been changed")
//		}
//	}
//
//	return
//}

//func (s *BlastServer) joinCluster() error {
//	if s.cluster == "" {
//		return fmt.Errorf("cluster is required")
//	}
//	if s.shard == "" {
//		return fmt.Errorf("shard is required")
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.etcdRequestTimeout)*time.Millisecond)
//	defer cancel()
//
//	keyNode := fmt.Sprintf("/blast/clusters/%s/shards/%s/nodes/%s", s.cluster, s.shard, s.node)
//
//	_, err := s.etcdKv.Put(ctx, keyNode, STATE_READY)
//	if err != nil {
//		log.WithFields(log.Fields{
//			"key":   keyNode,
//			"value": STATE_READY,
//			"error": err,
//		}).Error("failed to put the data")
//		return err
//	}
//	log.WithFields(log.Fields{
//		"key":   keyNode,
//		"value": STATE_READY,
//	}).Info("put the data")
//
//	return nil
//}

//func (s *BlastServer) leaveCluster() error {
//	if s.cluster == "" {
//		return fmt.Errorf("cluster is required")
//	}
//	if s.shard == "" {
//		return fmt.Errorf("shard is required")
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.etcdRequestTimeout)*time.Millisecond)
//	defer cancel()
//
//	keyNode := fmt.Sprintf("/blast/clusters/%s/shards/%s/nodes/%s", s.cluster, s.shard, s.node)
//
//	_, err := s.etcdKv.Delete(ctx, keyNode)
//	if err != nil {
//		log.WithFields(log.Fields{
//			"key":   keyNode,
//			"error": err,
//		}).Error("failed to delete the data")
//		return err
//	}
//	log.WithFields(log.Fields{
//		"key": keyNode,
//	}).Info("deleted the data")
//
//	return nil
//}
