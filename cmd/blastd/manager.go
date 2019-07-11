// Copyright (c) 2019 Minoru Osuka
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

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/indexutils"

	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/manager"
	"github.com/urfave/cli"
)

func startManager(c *cli.Context) error {
	logLevel := c.GlobalString("log-level")
	logFilename := c.GlobalString("log-file")
	logMaxSize := c.GlobalInt("log-max-size")
	logMaxBackups := c.GlobalInt("log-max-backups")
	logMaxAge := c.GlobalInt("log-max-age")
	logCompress := c.GlobalBool("log-compress")

	grpcLogLevel := c.GlobalString("grpc-log-level")
	grpcLogFilename := c.GlobalString("grpc-log-file")
	grpcLogMaxSize := c.GlobalInt("grpc-log-max-size")
	grpcLogMaxBackups := c.GlobalInt("grpc-log-max-backups")
	grpcLogMaxAge := c.GlobalInt("grpc-log-max-age")
	grpcLogCompress := c.GlobalBool("grpc-log-compress")

	httpAccessLogFilename := c.GlobalString("http-access-log-file")
	httpAccessLogMaxSize := c.GlobalInt("http-access-log-max-size")
	httpAccessLogMaxBackups := c.GlobalInt("http-access-log-max-backups")
	httpAccessLogMaxAge := c.GlobalInt("http-access-log-max-age")
	httpAccessLogCompress := c.GlobalBool("http-access-log-compress")

	nodeId := c.String("node-id")
	bindAddr := c.String("bind-addr")
	grpcAddr := c.String("grpc-addr")
	httpAddr := c.String("http-addr")
	dataDir := c.String("data-dir")
	raftStorageType := c.String("raft-storage-type")
	peerAddr := c.String("peer-addr")

	indexMappingFile := c.String("index-mapping-file")
	indexType := c.String("index-type")
	indexStorageType := c.String("index-storage-type")

	// create logger
	logger := logutils.NewLogger(
		logLevel,
		logFilename,
		logMaxSize,
		logMaxBackups,
		logMaxAge,
		logCompress,
	)

	// create logger
	grpcLogger := logutils.NewGRPCLogger(
		grpcLogLevel,
		grpcLogFilename,
		grpcLogMaxSize,
		grpcLogMaxBackups,
		grpcLogMaxAge,
		grpcLogCompress,
	)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger(
		httpAccessLogFilename,
		httpAccessLogMaxSize,
		httpAccessLogMaxBackups,
		httpAccessLogMaxAge,
		httpAccessLogCompress,
	)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()
	if peerAddr != "" {
		err := clusterConfig.SetPeerAddr(peerAddr)
		if err != nil {
			return err
		}
	}

	// create node config
	nodeConfig := config.NewNodeConfigFromMap(
		map[string]interface{}{
			"node_id":           nodeId,
			"bind_addr":         bindAddr,
			"grpc_addr":         grpcAddr,
			"http_addr":         httpAddr,
			"data_dir":          dataDir,
			"raft_storage_type": raftStorageType,
		},
	)

	var err error

	// create index mapping
	var indexMapping *mapping.IndexMappingImpl
	if indexMappingFile != "" {
		indexMapping, err = indexutils.NewIndexMappingFromFile(indexMappingFile)
		if err != nil {
			return err
		}
	} else {
		indexMapping = mapping.NewIndexMapping()
	}

	// create index config
	indexConfig := config.DefaultIndexConfig()
	err = indexConfig.SetIndexMapping(indexMapping)
	if err != nil {
		return err
	}
	err = indexConfig.SetIndexType(indexType)
	if err != nil {
		return err
	}
	err = indexConfig.SetIndexStorageType(indexStorageType)
	if err != nil {
		return err
	}

	svr, err := manager.NewServer(clusterConfig, nodeConfig, indexConfig, logger.Named(nodeId), grpcLogger.Named(nodeId), httpAccessLogger)
	if err != nil {
		return err
	}

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Kill, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go svr.Start()

	<-quitCh

	svr.Stop()

	return nil
}
