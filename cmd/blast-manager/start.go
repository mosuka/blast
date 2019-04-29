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
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/manager"
	"github.com/mosuka/blast/protobuf/raft"
	"github.com/mosuka/logutils"
	"github.com/urfave/cli"
)

func execStart(c *cli.Context) error {
	nodeId := c.String("node-id")
	bindAddr := c.String("bind-addr")
	grpcAddr := c.String("grpc-addr")
	httpAddr := c.String("http-addr")
	dataDir := c.String("data-dir")
	peerAddr := c.String("peer-addr")

	indexMappingFile := c.String("index-mapping-file")
	indexType := c.String("index-type")
	indexStorageType := c.String("index-storage-type")

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

	// create logger
	logger := logutils.NewLogger(
		logLevel,
		logFilename,
		logMaxSize,
		logMaxBackups,
		logMaxAge,
		logCompress,
	)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger(
		httpAccessLogFilename,
		httpAccessLogMaxSize,
		httpAccessLogMaxBackups,
		httpAccessLogMaxAge,
		httpAccessLogCompress,
	)

	// node
	node := &raft.Node{
		Id: nodeId,
		Metadata: &raft.Metadata{
			BindAddr: bindAddr,
			GrpcAddr: grpcAddr,
			HttpAddr: httpAddr,
			DataDir:  dataDir,
		},
		Leader: false,
	}

	// index mapping
	indexMapping := mapping.NewIndexMapping()
	if indexMappingFile != "" {
		_, err := os.Stat(indexMappingFile)
		if err == nil {
			// read index mapping file
			f, err := os.Open(indexMappingFile)
			if err != nil {
				return err
			}
			defer func() {
				_ = f.Close()
			}()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			err = json.Unmarshal(b, indexMapping)
			if err != nil {
				return err
			}
		} else if os.IsNotExist(err) {
			return err
		}
	}
	err := indexMapping.Validate()
	if err != nil {
		return err
	}

	// IndexMappingImpl -> JSON
	indexMappingJSON, err := json.Marshal(indexMapping)
	if err != nil {
		return err
	}
	// JSON -> map[string]interface{}
	var indexMappingMap map[string]interface{}
	err = json.Unmarshal(indexMappingJSON, &indexMappingMap)
	if err != nil {
		return err
	}

	indexConfig := map[string]interface{}{
		"index_mapping":      indexMappingMap,
		"index_type":         indexType,
		"index_storage_type": indexStorageType,
	}

	svr, err := manager.NewServer(node, peerAddr, indexConfig, logger, httpAccessLogger)
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
