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

type BlastServer struct {
	host       string
	port       int
	server     *grpc.Server
	service    *service.BlastService
	etcdClient *client.EtcdClient
	cluster    string
	shard      string
}

func NewBlastServer(port int, indexPath string, indexMapping *mapping.IndexMappingImpl, indexType string, kvstore string, kvconfig map[string]interface{}, etcdEndpoints []string, etcdDialTimeout int, etcdRequestTimeout int, cluster string, shard string) (*BlastServer, error) {
	host, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to get hostname")
		return nil, err
	}

	var etcdClient *client.EtcdClient
	if len(etcdEndpoints) > 0 {
		etcdClient, err = client.NewEtcdClient(etcdEndpoints, etcdDialTimeout, etcdRequestTimeout)
		if err != nil {
			return nil, err
		}
	}

	svr := grpc.NewServer()
	svc := service.NewBlastService(indexPath, indexMapping, indexType, kvstore, kvconfig)
	proto.RegisterIndexServer(svr, svc)

	return &BlastServer{
		host:       host,
		port:       port,
		server:     svr,
		service:    svc,
		etcdClient: etcdClient,
		cluster:    cluster,
		shard:      shard,
	}, nil
}

func (s *BlastServer) Start() error {
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

	return nil
}

func (s *BlastServer) Stop() error {

	if s.etcdClient != nil {
		s.etcdClient.Close()
	}

	err := s.service.CloseIndex()
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
