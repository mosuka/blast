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
	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/indexutils"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/manager"
	"github.com/urfave/cli"
)

func clusterNodeStart(c *cli.Context) error {
	peerGrpcAddr := c.String("peer-grpc-address")

	grpcAddr := c.String("grpc-address")
	httpAddr := c.String("http-address")

	nodeId := c.String("node-id")
	nodeAddr := c.String("node-address")
	dataDir := c.String("data-dir")
	raftStorageType := c.String("raft-storage-type")

	indexMappingFile := c.String("index-mapping-file")
	indexType := c.String("index-type")
	indexStorageType := c.String("index-storage-type")

	logLevel := c.String("log-level")
	logFilename := c.String("log-file")
	logMaxSize := c.Int("log-max-size")
	logMaxBackups := c.Int("log-max-backups")
	logMaxAge := c.Int("log-max-age")
	logCompress := c.Bool("log-compress")

	grpcLogLevel := c.String("grpc-log-level")
	grpcLogFilename := c.String("grpc-log-file")
	grpcLogMaxSize := c.Int("grpc-log-max-size")
	grpcLogMaxBackups := c.Int("grpc-log-max-backups")
	grpcLogMaxAge := c.Int("grpc-log-max-age")
	grpcLogCompress := c.Bool("grpc-log-compress")

	httpLogFilename := c.String("http-log-file")
	httpLogMaxSize := c.Int("http-log-max-size")
	httpLogMaxBackups := c.Int("http-log-max-backups")
	httpLogMaxAge := c.Int("http-log-max-age")
	httpLogCompress := c.Bool("http-log-compress")

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
	httpLogger := logutils.NewApacheCombinedLogger(
		httpLogFilename,
		httpLogMaxSize,
		httpLogMaxBackups,
		httpLogMaxAge,
		httpLogCompress,
	)

	// create cluster config
	clusterConfig := config.DefaultClusterConfig()
	if peerGrpcAddr != "" {
		clusterConfig.PeerAddr = peerGrpcAddr
	}

	// create node config
	nodeConfig := &config.NodeConfig{
		NodeId:          nodeId,
		BindAddr:        nodeAddr,
		GRPCAddr:        grpcAddr,
		HTTPAddr:        httpAddr,
		DataDir:         dataDir,
		RaftStorageType: raftStorageType,
	}

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
	indexConfig := &config.IndexConfig{
		IndexMapping:     indexMapping,
		IndexType:        indexType,
		IndexStorageType: indexStorageType,
	}

	svr, err := manager.NewServer(clusterConfig, nodeConfig, indexConfig, logger.Named(nodeId), grpcLogger.Named(nodeId), httpLogger)
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
