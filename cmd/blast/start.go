//  Copyright (c) 2018 Minoru Osuka
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blevesearch/bleve/mapping"
	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/grpc/client"
	grpcserver "github.com/mosuka/blast/grpc/server"
	httpserver "github.com/mosuka/blast/http/server"
	"github.com/mosuka/blast/index/bleve"
	"github.com/mosuka/blast/logging"
	"github.com/mosuka/blast/protobuf"
	braft "github.com/mosuka/blast/raft"
	"github.com/mosuka/blast/service"
	"github.com/mosuka/blast/store/boltdb"
	"github.com/mosuka/blast/version"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var logo = `
  ____  _           _   
 | __ )| | __ _ ___| |_ 
 |  _ \| |/ _' / __| __|  The lightweight distributed
 | |_) | | (_| \__ \ |_   indexing and search server.
 |____/|_|\__,_|___/\__|  version ` + version.Version + `
`

func start(c *cli.Context) {
	// Display logo.
	fmt.Println(logo)

	bindAddr := c.String("bind-addr")
	grpcAddr := c.String("grpc-addr")
	httpAddr := c.String("http-addr")

	nodeID := c.String("node-id")
	raftDir := c.String("raft-dir")
	retainSnapshotCount := c.Int("retain-snapshot-count")
	raftTimeout := c.String("raft-timeout")

	storeDir := c.String("store-dir")

	indexDir := c.String("index-dir")
	indexMapping := c.String("index-mapping")
	indexType := c.String("index-type")
	indexKvstore := c.String("index-kvstore")

	peerGRPCAddr := c.String("peer-grpc-addr")
	maxSendMsgSize := c.Int("max-send-msg-size")
	maxRecvMsgSize := c.Int("max-recv-msg-size")

	logLevel := c.String("log-level")
	logFilename := c.String("log-filename")
	logMaxSize := c.Int("log-max-size")
	logMaxBackups := c.Int("log-max-backups")
	logMaxAge := c.Int("log-max-age")
	logCompress := c.Bool("log-compress")

	httpAccessLogFilename := c.String("http-access-log-filename")
	httpAccessLogMaxSize := c.Int("http-access-log-max-size")
	httpAccessLogMaxBackups := c.Int("http-access-log-max-backups")
	httpAccessLogMaxAge := c.Int("http-access-log-max-age")
	httpAccessLogCompress := c.Bool("http-access-log-compress")

	var err error

	// Raft config
	raftConfig := braft.DefaultConfig()
	if nodeID != "" {
		raftConfig.Config.LocalID = raft.ServerID(nodeID)
	}
	if raftDir != "" {
		raftConfig.Path = raftDir
	}
	if retainSnapshotCount > 0 {
		raftConfig.RetainSnapshotCount = retainSnapshotCount
	}
	if raftTimeout != "" {
		if raftConfig.Timeout, err = time.ParseDuration(raftTimeout); err != nil {
			fmt.Fprint(os.Stderr, errors.Wrap(err, "Failed to parse raft timeout"))
			return
		}
	}

	// Store config
	storeConfig := boltdb.DefaultConfig()
	if storeDir != "" {
		storeConfig.Path = storeDir
	}

	// Index config
	indexConfig := bleve.DefaultConfig()
	if indexDir != "" {
		indexConfig.Path = indexDir
	}
	if indexMapping != "" {
		var imf *os.File
		if imf, err = os.Open(indexMapping); err != nil {
			fmt.Fprint(os.Stderr, errors.Wrap(err, "Failed to open index mapping"))
			return
		}
		defer imf.Close()

		var imb []byte
		if imb, err = ioutil.ReadAll(imf); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to read index mapping"))
			return
		}

		im := mapping.NewIndexMapping()
		if err = json.Unmarshal(imb, &im); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to unmarshal index mapping"))
			return
		}
		indexConfig.IndexMapping = im
	}
	if indexType != "" {
		indexConfig.IndexType = indexType
	}
	if indexKvstore != "" {
		indexConfig.Kvstore = indexKvstore
	}

	// Create logger
	logger := logging.Logger(
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
	var bootstrap bool
	bootstrap = peerGRPCAddr == "" || peerGRPCAddr == grpcAddr

	// Create Service
	var svc *service.KVSService
	if svc, err = service.NewKVSService(bindAddr, raftConfig, bootstrap, storeConfig, indexConfig); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to create service"))
		return
	}
	svc.SetLogger(logger)

	// Start service
	if err = svc.Start(); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to start service"))
		return
	}
	defer svc.Stop()

	// Create gRPC server
	var grpcServer *grpcserver.GRPCServer
	if grpcServer, err = grpcserver.NewGRPCServer(grpcAddr, svc); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to create gRPC Server"))
		return
	}
	defer grpcServer.Stop()
	grpcServer.SetLogger(logger)
	grpcServer.SetMaxSendMessageSize(maxSendMsgSize)
	grpcServer.SetMaxReceiveMessageSize(maxRecvMsgSize)

	// Start gRPC server
	if err = grpcServer.Start(); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to start gRPC Server"))
		return
	}

	// Create HTTP access logger
	httpAccessLogger := logging.HTTPAccessLogger(
		httpAccessLogFilename,
		httpAccessLogMaxSize,
		httpAccessLogMaxBackups,
		httpAccessLogMaxAge,
		httpAccessLogCompress,
	)

	// Create HTTP server
	var httpServer *httpserver.HTTPServer
	if httpServer, err = httpserver.NewHTTPServer(httpAddr, grpcAddr); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to initialize HTTP Server"))
		return
	}
	defer httpServer.Stop()

	// Setup HTTP server
	httpServer.SetLogger(logger)
	httpServer.SetHTTPAccessLogger(httpAccessLogger)
	httpServer.SetMaxSendMessageSize(maxSendMsgSize)
	httpServer.SetMaxReceiveMessageSize(maxRecvMsgSize)

	// Start HTTP server
	if err = httpServer.Start(); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to start HTTP Server"))
		return
	}

	var joinReq *protobuf.JoinRequest
	joinReq = &protobuf.JoinRequest{
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
		if _, err = svc.WaitForLeader(60 * time.Second); err != nil {
			fmt.Fprintln(os.Stderr, errors.Wrap(err, "Failed to detect leader node"))
			return
		}

		// Put a metadata of bootstrap node
		svc.PutMetadata(joinReq)
	} else {
		// If node is not bootstrap, make the join request.
		var grpcClient *client.GRPCClient
		if grpcClient, err = client.NewGRPCClient(peerGRPCAddr, maxSendMsgSize, maxRecvMsgSize); err != nil {
			fmt.Fprintln(os.Stderr, errors.New(err.Error()))
			return
		}
		defer grpcClient.Close()

		grpcClient.Join(joinReq)
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
