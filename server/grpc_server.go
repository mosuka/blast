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
)

type blastGRPCServer struct {
	hostname   string
	port       int
	server     *grpc.Server
	listener   net.Listener
	service    *service.BlastGRPCService
	etcdClient *client.EtcdClientWrapper
}

func NewBlastGRPCServer(port int) *blastGRPCServer {
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get hostname")
		return nil
	}

	return &blastGRPCServer{
		hostname: hostname,
		port:     port,
	}
}

func (s *blastGRPCServer) StartOnCluster(etcdServers []string, requestTimeout int, clusterName string) error {
	etcdClient, err := client.NewEtcdClientWrapper(etcdServers, requestTimeout)
	if err == nil {
		log.WithFields(log.Fields{
			"etcdServers":    etcdServers,
			"requestTimeout": requestTimeout,
		}).Info("succeeded in connect to etcd servers")
	} else {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to connect etcd server")
		return nil
	}
	s.etcdClient = etcdClient

	indexPath, err := s.etcdClient.GetIndexPath(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get indexPath")
		return nil
	}

	indexMapping, err := s.etcdClient.GetIndexMapping(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get indexMapping")
		return nil
	}

	indexType, err := s.etcdClient.GetIndexType(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get indexType")
		return nil
	}

	kvstore, err := s.etcdClient.GetKvstore(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get kvstore")
		return nil
	}

	kvconfig, err := s.etcdClient.GetKvconfig(clusterName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get kvconfig")
		return nil
	}

	return s.Start(indexPath, indexMapping, indexType, kvstore, kvconfig)
}

func (s *blastGRPCServer) Start(indexPath string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{}) error {
	s.server = grpc.NewServer()
	s.service = service.NewBlastGRPCService(indexPath, indexMapping, indexType, kvstore, kvconfig)

	proto.RegisterBlastServer(s.server, s.service)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err == nil {
		log.WithFields(log.Fields{
			"port": s.port,
		}).Info("create listener")
	} else {
		log.WithFields(log.Fields{
			"port":  s.port,
			"error": err.Error(),
		}).Error("failed to create listener")
		return nil
	}
	s.listener = l

	go func() {
		s.service.OpenIndex()
		s.server.Serve(s.listener)
		return
	}()

	log.WithFields(log.Fields{
		"host": s.hostname,
		"port": s.port,
	}).Info("The Blast server started")

	return nil
}

func (s *blastGRPCServer) Stop() error {
	if s.etcdClient != nil {
		err := s.etcdClient.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to close etcd client")
			return nil
		}
	}

	err := s.service.CloseIndex()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to close index")
		return nil
	}

	s.server.GracefulStop()

	log.WithFields(log.Fields{
		"host": s.hostname,
		"port": s.port,
	}).Info("The Blast server stopped")

	return nil
}
