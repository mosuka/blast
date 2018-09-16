// Copyright (c) 2018 Minoru Osuka
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
	"fmt"
	"github.com/mosuka/blast/version"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/index"
	blastlog "github.com/mosuka/blast/log"
	"github.com/mosuka/blast/node/data/client"
	"github.com/mosuka/blast/node/data/protobuf"
	"github.com/mosuka/blast/node/data/server"
	"github.com/mosuka/blast/node/data/service"
	blastraft "github.com/mosuka/blast/raft"
	"github.com/mosuka/blast/store"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var logo = `
    ____   __              __ 
   / __ ) / /____ _ _____ / /_
  / __ \ / // __ '// ___// __/  The lightweight distributed
 / /_/ // // /_/ /(__  )/ /_    indexing and search server.
/_.___//_/ \__,_//____/ \__/    version ` + version.Version + `
`

func data(c *cli.Context) {
	// Display logo.
	fmt.Println(logo)

	bindAddr := c.String("bind-addr")
	grpcAddr := c.String("grpc-addr")
	httpAddr := c.String("http-addr")

	nodeID := c.String("raft-node-id")
	raftDir := c.String("raft-dir")
	snapshotCount := c.Int("raft-snapshot-count")
	raftTimeout := c.String("raft-timeout")

	storeDir := c.String("store-dir")

	indexDir := c.String("index-dir")
	indexMappingFile := c.String("index-mapping-file")
	indexType := c.String("index-type")
	indexKvstore := c.String("index-kvstore")

	peerGRPCAddr := c.String("peer-grpc-addr")

	logLevel := c.String("log-level")
	logFilename := c.String("log-file")
	logMaxSize := c.Int("log-max-size")
	logMaxBackups := c.Int("log-max-backups")
	logMaxAge := c.Int("log-max-age")
	logCompress := c.Bool("log-compress")

	httpAccessLogFilename := c.String("http-access-log-file")
	httpAccessLogMaxSize := c.Int("http-access-log-max-size")
	httpAccessLogMaxBackups := c.Int("http-access-log-max-backups")
	httpAccessLogMaxAge := c.Int("http-access-log-max-age")
	httpAccessLogCompress := c.Bool("http-access-log-compress")

	var err error

	// Raft config
	raftConfig := blastraft.DefaultRaftConfig()
	raftConfig.Config.LocalID = raft.ServerID(nodeID)
	raftConfig.Dir = raftDir
	raftConfig.SnapshotCount = snapshotCount
	raftConfig.Timeout, err = time.ParseDuration(raftTimeout)
	if err != nil {
		fmt.Fprint(os.Stderr, errors.Wrap(err, "Failed to parse raft timeout"))
		return
	}

	// Store config
	storeConfig := store.DefaultStoreConfig()
	storeConfig.Dir = storeDir

	// Index config
	indexConfig := index.DefaultIndexConfig()
	indexConfig.Dir = indexDir
	indexConfig.IndexType = indexType
	indexConfig.Kvstore = indexKvstore
	if indexMappingFile != "" {
		err = indexConfig.SetIndexMapping(indexMappingFile)
		if err != nil {
			fmt.Fprint(os.Stderr, errors.Wrap(err, "Failed to read index mapping file"))
			return
		}
	}

	// Create logger
	logger := blastlog.Logger(
		logLevel,
		"",
		log.LstdFlags|log.Lmicroseconds|log.LUTC,
		logFilename,
		logMaxSize,
		logMaxBackups,
		logMaxAge,
		logCompress,
	)

	// Check bootstrap node
	bootstrap := peerGRPCAddr == "" || peerGRPCAddr == grpcAddr

	// Create Service
	svc, err := service.NewService(bindAddr, raftConfig, bootstrap, storeConfig, indexConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to create service"))
		return
	}
	svc.SetLogger(logger)

	// Start service
	err = svc.Start()
	defer svc.Stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to start service"))
		return
	}

	// Create data server
	dataGRPCServer, err := server.NewGRPCServer(grpcAddr, svc)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to create gRPC Server"))
		return
	}
	dataGRPCServer.SetLogger(logger)

	// Start gRPC server
	err = dataGRPCServer.Start()
	defer dataGRPCServer.Stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to start gRPC Server"))
		return
	}

	// Create HTTP access logger
	httpAccessLogger := blastlog.HTTPAccessLogger(
		httpAccessLogFilename,
		httpAccessLogMaxSize,
		httpAccessLogMaxBackups,
		httpAccessLogMaxAge,
		httpAccessLogCompress,
	)

	// Create HTTP server
	httpServer, err := server.NewHTTPServer(httpAddr, grpcAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to initialize HTTP Server"))
		return
	}

	// Setup HTTP server
	httpServer.SetLogger(logger)
	httpServer.SetHTTPAccessLogger(httpAccessLogger)

	// Start HTTP server
	err = httpServer.Start()
	defer httpServer.Stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to start HTTP Server"))
		return
	}

	joinReq := &protobuf.JoinRequest{
		NodeId:  nodeID,
		Address: bindAddr,
		Metadata: &protobuf.Metadata{
			GrpcAddress: grpcAddr,
			HttpAddress: httpAddr,
		},
	}

	if bootstrap {
		// If node is bootstrap, put metadata into service.
		// Wait for leader detected
		_, err = svc.WaitForLeader(60 * time.Second)
		if err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to detect leader node"))
			return
		}

		// Put a metadata of bootstrap node
		svc.PutMetadata(joinReq)
	} else {
		// If node is not bootstrap, make the join request.
		dataClient, err := client.NewGRPCClient(peerGRPCAddr)
		defer dataClient.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, errors.New(err.Error()))
			return
		}

		dataClient.Join(joinReq)
	}

	// Wait signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	_ = <-signalChan

	return
}
