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

package indexer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/testutils"
)

func TestSingleNode(t *testing.T) {
	tmpDir, err := testutils.TmpDir()
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	curDir, _ := os.Getwd()

	logLevel := "DEBUG"
	logFilename := ""
	logMaxSize := 500
	logMaxBackups := 3
	logMaxAge := 30
	logCompress := false

	httpAccessLogFilename := ""
	httpAccessLogMaxSize := 500
	httpAccessLogMaxBackups := 3
	httpAccessLogMaxAge := 30
	httpAccessLogCompress := false

	nodeId := "indexer1"
	bindPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	bindAddr := fmt.Sprintf(":%d", bindPort)
	grpcPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	grpcAddr := fmt.Sprintf(":%d", grpcPort)
	httpPort, err := testutils.TmpPort()
	if err != nil {
		t.Errorf("%v", err)
	}
	httpAddr := fmt.Sprintf(":%d", httpPort)
	dataDir := filepath.Join(tmpDir, "store")
	peerAddr := ""

	indexMappingFile := filepath.Join(curDir, "../example/wiki_index_mapping.json")
	indexType := "upside_down"
	indexStorageType := "boltdb"

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

	// metadata
	metadata := map[string]interface{}{
		"bind_addr": bindAddr,
		"grpc_addr": grpcAddr,
		"http_addr": httpAddr,
		"data_dir":  dataDir,
	}

	// index mapping
	indexMapping := mapping.NewIndexMapping()
	if indexMappingFile != "" {
		_, err := os.Stat(indexMappingFile)
		if err == nil {
			// read index mapping file
			f, err := os.Open(indexMappingFile)
			if err != nil {
				t.Errorf("%v", err)
			}
			defer func() {
				_ = f.Close()
			}()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				t.Errorf("%v", err)
			}

			err = json.Unmarshal(b, indexMapping)
			if err != nil {
				t.Errorf("%v", err)
			}
		} else if os.IsNotExist(err) {
			t.Errorf("%v", err)
		}
	}
	err = indexMapping.Validate()
	if err != nil {
		t.Errorf("%v", err)
	}

	// IndexMappingImpl -> JSON
	indexMappingJSON, err := json.Marshal(indexMapping)
	if err != nil {
		t.Errorf("%v", err)
	}
	// JSON -> map[string]interface{}
	var indexMappingMap map[string]interface{}
	err = json.Unmarshal(indexMappingJSON, &indexMappingMap)
	if err != nil {
		t.Errorf("%v", err)
	}

	indexConfig := map[string]interface{}{
		"index_mapping":      indexMappingMap,
		"index_type":         indexType,
		"index_storage_type": indexStorageType,
	}

	server, err := NewServer("", "", nodeId, metadata, peerAddr, indexConfig, logger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(5 * time.Second)

	// create gRPC client
	client, err := grpc.NewClient(grpcAddr)
	defer func() {
		_ = client.Close()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// liveness
	liveness, err := client.LivenessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := protobuf.LivenessProbeResponse_ALIVE.String()
	act1 := liveness
	if exp1 != act1 {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	// readiness
	readiness, err := client.ReadinessProbe()
	if err != nil {
		t.Errorf("%v", err)
	}
	exp2 := protobuf.ReadinessProbeResponse_READY.String()
	act2 := readiness
	if exp1 != act1 {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	// get node
	node, err := client.GetNode(nodeId)
	if err != nil {
		t.Errorf("%v", err)
	}
	exp3 := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": bindAddr,
			"grpc_addr": grpcAddr,
			"http_addr": httpAddr,
			"data_dir":  dataDir,
		},
		"state": "Leader",
	}
	act3 := node
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", exp3, act3)
	}

	// get cluster
	cluster, err := client.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	exp4 := map[string]interface{}{
		nodeId: map[string]interface{}{
			"metadata": map[string]interface{}{
				"bind_addr": bindAddr,
				"grpc_addr": grpcAddr,
				"http_addr": httpAddr,
				"data_dir":  dataDir,
			},
			"state": "Leader",
		},
	}
	act4 := cluster
	if !reflect.DeepEqual(exp4, act4) {
		t.Errorf("expected content to see %v, saw %v", exp4, act4)
	}

	// get index config
	val5, err := client.GetIndexConfig()
	if err != nil {
		t.Errorf("%v", err)
	}
	exp5 := indexConfig
	act5 := val5
	if !reflect.DeepEqual(exp5, act5) {
		t.Errorf("expected content to see %v, saw %v", exp5, act5)
	}

	// get index stats
	val6, err := client.GetIndexStats()
	if err != nil {
		t.Errorf("%v", err)
	}
	exp6 := map[string]interface{}{
		"index": map[string]interface{}{
			"analysis_time":                float64(0),
			"batches":                      float64(0),
			"deletes":                      float64(0),
			"errors":                       float64(0),
			"index_time":                   float64(0),
			"num_plain_text_bytes_indexed": float64(0),
			"term_searchers_finished":      float64(0),
			"term_searchers_started":       float64(0),
			"updates":                      float64(0),
		},
		"search_time": float64(0),
		"searches":    float64(0),
	}
	act6 := val6
	if !reflect.DeepEqual(exp6, act6) {
		t.Errorf("expected content to see %v, saw %v", exp6, act6)
	}
}
