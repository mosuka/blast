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

type BlastServer struct {
	cluster.BlastCluster

	hostname   string
	port       int
	server     *grpc.Server
	service    *service.BlastService
	collection string
	node       string
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
		node:         node,
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

	if s.BlastCluster != nil {
		s.watchCollection()
		log.WithFields(log.Fields{
			"collection": s.collection,
		}).Info("watching the cluster information has been started")
	}

	// start server
	go func() {
		s.service.OpenIndex()
		s.server.Serve(listener)
		return
	}()
	log.Info("server has been started")

	if s.BlastCluster != nil {
		err := s.joinCluster()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to join the collection")
		} else {
			log.WithFields(log.Fields{
				"collection": s.collection,
			}).Info("server has been joined the collection")
		}
	}

	return nil
}

func (s *BlastServer) Stop() error {
	if s.BlastCluster != nil {
		err := s.leaveCluster()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to leave the collection")
		} else {
			log.WithFields(log.Fields{
				"collection": s.collection,
			}).Info("server has been left the collection")
		}

		err = s.BlastCluster.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to disconnect the cluster")
		}
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

func (s *BlastServer) joinCluster() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15000)*time.Millisecond)
	defer cancel()

	err := s.BlastCluster.PutNode(ctx, s.collection, s.node, cluster.STATE_ACTIVE)

	return err
}

func (s *BlastServer) leaveCluster() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15000)*time.Millisecond)
	defer cancel()

	err := s.BlastCluster.DeleteNode(ctx, s.collection, s.node)

	return err
}

func (s *BlastServer) watchCollection() {
	go func() {
		var ctx context.Context
		var cancel context.CancelFunc

		for {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(15000)*time.Millisecond)
			s.BlastCluster.Watch(ctx, s.collection)
		}
		defer cancel()

		return
	}()

	return
}
