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
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/client"
	"github.com/mosuka/blast/proto"
	"github.com/mosuka/blast/service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

type BlastServer struct {
	host        string
	port        int
	etcdClient  *client.EtcdClientWrapper
	clusterName string
	server      *grpc.Server
	service     *service.BlastService
}

func NewStandaloneMode(port int, indexPath string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{}) *BlastServer {
	host, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get hostname")
		return nil
	}

	svr := grpc.NewServer()
	svc := service.NewBlastService(indexPath, indexMapping, indexType, kvstore, kvconfig)
	proto.RegisterIndexServer(svr, svc)

	return &BlastServer{
		host:    host,
		port:    port,
		server:  svr,
		service: svc,
	}
}

func NewClusterMode(port int, indexPath string, etcdServers []string, etcdDialTimeout int, etcdRequestTimeout int, clusterName string) *BlastServer {
	host, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get hostname")
		return nil
	}

	etcdClient, err := client.NewEtcdClientWrapper(etcdServers, etcdDialTimeout, etcdRequestTimeout)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to connect etcd server")
		return nil
	}

	indexMapping, err := etcdClient.GetIndexMapping(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"clusterName": clusterName,
			"error":       err.Error(),
		}).Error("failed to get index mapping")
		return nil
	}
	log.WithFields(log.Fields{
		"clusterName": clusterName,
	}).Info("succeeded in get index mapping")

	indexType, err := etcdClient.GetIndexType(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"clusterName": clusterName,
			"error":       err.Error(),
		}).Error("failed to get index type")
		return nil
	}
	log.WithFields(log.Fields{
		"clusterName": clusterName,
	}).Info("succeeded in get index type")

	kvstore, err := etcdClient.GetKvstore(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"clusterName": clusterName,
			"error":       err.Error(),
		}).Error("failed to get kvstore")
		return nil
	}
	log.WithFields(log.Fields{
		"clusterName": clusterName,
	}).Info("succeeded in get kvstore")

	kvconfig, err := etcdClient.GetKvconfig(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"clusterName": clusterName,
			"error":       err.Error(),
		}).Error("failed to get kvconfig")
		return nil
	}
	log.WithFields(log.Fields{
		"clusterName": clusterName,
	}).Info("succeeded in get kvconfig")

	svr := grpc.NewServer()
	svc := service.NewBlastService(indexPath, indexMapping, indexType, kvstore, kvconfig)
	proto.RegisterIndexServer(svr, svc)

	return &BlastServer{
		host:        host,
		port:        port,
		etcdClient:  etcdClient,
		clusterName: clusterName,
		server:      svr,
		service:     svc,
	}
}

func (s *BlastServer) joinCluster() error {
	if s.etcdClient != nil {
		log.WithFields(log.Fields{
			"cluster": s.clusterName,
			"host":    s.host,
			"port":    s.port,
		}).Info("join a cluster")
	}

	return nil
}

func (s *BlastServer) leaveCluster() error {
	if s.etcdClient != nil {
		log.WithFields(log.Fields{
			"cluster": s.clusterName,
			"host":    s.host,
			"port":    s.port,
		}).Info("leave a cluster")
	}

	return nil
}

func (s *BlastServer) Watch() {
	log.Infof("start watching %s", s.clusterName)

	go func() {
		for {
			err := s.etcdClient.Watch(s.clusterName)
			if err != nil {
				log.WithFields(log.Fields{
					"clusterName": s.clusterName,
					"error":       err.Error(),
				}).Error("failed to watch cluster")

				time.Sleep(time.Duration(3000) * time.Millisecond)

				continue
			}
		}
		return
	}()

	return
}

func (s *BlastServer) Start() error {
	if s.etcdClient != nil && s.clusterName != "" {
		s.Watch()
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err == nil {
		log.WithFields(log.Fields{
			"port": s.port,
		}).Info("create listener")
	} else {
		log.WithFields(log.Fields{
			"port":  s.port,
			"error": err.Error(),
		}).Error("failed to create listener")
		return err
	}

	go func() {
		s.service.OpenIndex()
		s.server.Serve(listener)
		return
	}()
	log.WithFields(log.Fields{
		"host": s.host,
		"port": s.port,
	}).Info("The Blast server started")

	err = s.joinCluster()

	return nil
}

func (s *BlastServer) Stop() error {
	err := s.leaveCluster()

	if s.etcdClient != nil {
		err := s.etcdClient.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to close etcd client")
			return err
		}
	}

	err = s.service.CloseIndex()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to close index")
		return err
	}

	s.server.GracefulStop()

	log.WithFields(log.Fields{
		"host": s.host,
		"port": s.port,
	}).Info("The Blast server stopped")

	return nil
}
