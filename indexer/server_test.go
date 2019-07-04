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

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/grpc"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/testutils"
)

func TestIndexserStandalone(t *testing.T) {
	curDir, _ := os.Getwd()

	indexMappingFile := filepath.Join(curDir, "../example/wiki_index_mapping.json")
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
	err := indexMapping.Validate()
	if err != nil {
		t.Errorf("%v", err)
	}
	indexMappingJSON, err := json.Marshal(indexMapping)
	if err != nil {
		t.Errorf("%v", err)
	}
	var indexMappingMap map[string]interface{}
	err = json.Unmarshal(indexMappingJSON, &indexMappingMap)
	if err != nil {
		t.Errorf("%v", err)
	}

	indexType := "upside_down"
	indexStorageType := "boltdb"

	indexConfig := map[string]interface{}{
		"index_mapping":      indexMappingMap,
		"index_type":         indexType,
		"index_storage_type": indexStorageType,
	}

	// create logger
	logger := logutils.NewLogger("WARN", "", 500, 3, 30, false)

	// create HTTP access logger
	httpAccessLogger := logutils.NewApacheCombinedLogger("", 500, 3, 30, false)

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

	dataDir, err := testutils.TmpDir()
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(dataDir)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	peerAddr := ""

	// metadata
	metadata := map[string]interface{}{
		"bind_addr": bindAddr,
		"grpc_addr": grpcAddr,
		"http_addr": httpAddr,
		"data_dir":  dataDir,
	}

	server, err := NewServer("", "", nodeId, metadata, "boltdb", peerAddr, indexConfig, logger, httpAccessLogger)
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
	expNode := map[string]interface{}{
		"metadata": map[string]interface{}{
			"bind_addr": bindAddr,
			"grpc_addr": grpcAddr,
			"http_addr": httpAddr,
			"data_dir":  dataDir,
		},
		"state": "Leader",
	}
	actNode := node
	if !reflect.DeepEqual(expNode, actNode) {
		t.Errorf("expected content to see %v, saw %v", expNode, actNode)
	}

	// get cluster
	cluster, err := client.GetCluster()
	if err != nil {
		t.Errorf("%v", err)
	}
	expCluster := map[string]interface{}{
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
	actCluster := cluster
	if !reflect.DeepEqual(expCluster, actCluster) {
		t.Errorf("expected content to see %v, saw %v", expCluster, actCluster)
	}

	// get index config
	indexConfig1, err := client.GetIndexConfig()
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexConfig := indexConfig
	actIndexConfig := indexConfig1
	if !reflect.DeepEqual(expIndexConfig, actIndexConfig) {
		t.Errorf("expected content to see %v, saw %v", expIndexConfig, actIndexConfig)
	}

	// get index stats
	indexStats, err := client.GetIndexStats()
	if err != nil {
		t.Errorf("%v", err)
	}
	expIndexStats := map[string]interface{}{
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
	actIndexStats := indexStats
	if !reflect.DeepEqual(expIndexStats, actIndexStats) {
		t.Errorf("expected content to see %v, saw %v", expIndexStats, actIndexStats)
	}

	// put document
	endikiDocs := make([]map[string]interface{}, 0)
	enwiki1fieldsPath := filepath.Join(curDir, "../example/wiki_doc_enwiki_1.json")
	// read index mapping file
	enwiki1FieldsFile, err := os.Open(enwiki1fieldsPath)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		_ = enwiki1FieldsFile.Close()
	}()

	enwiki1FieldsByte, err := ioutil.ReadAll(enwiki1FieldsFile)
	if err != nil {
		t.Errorf("%v", err)
	}

	var enwiki1Fields map[string]interface{}
	err = json.Unmarshal(enwiki1FieldsByte, &enwiki1Fields)
	if err != nil {
		t.Errorf("%v", err)
	}
	enwiki1Doc := map[string]interface{}{
		"id":     "enwiki_1",
		"fields": enwiki1Fields,
	}
	endikiDocs = append(endikiDocs, enwiki1Doc)
	count, err := client.IndexDocument(endikiDocs)
	if err != nil {
		t.Errorf("%v", err)
	}
	expCount := 1
	actCount := count
	if expCount != actCount {
		t.Errorf("expected content to see %v, saw %v", expCount, actCount)
	}

	// get document
	fields1, err := client.GetDocument("enwiki_1")
	if err != nil {
		t.Errorf("%v", err)
	}
	expFields := enwiki1Fields
	actFields := fields1
	if !reflect.DeepEqual(expFields, actFields) {
		t.Errorf("expected content to see %v, saw %v", expFields, actFields)
	}

	// search
	searchRequestPath := filepath.Join(curDir, "../example/wiki_search_request.json")

	searchRequestFile, err := os.Open(searchRequestPath)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		_ = searchRequestFile.Close()
	}()

	searchRequestByte, err := ioutil.ReadAll(searchRequestFile)
	if err != nil {
		t.Errorf("%v", err)
	}

	searchRequest := bleve.NewSearchRequest(nil)
	err = json.Unmarshal(searchRequestByte, searchRequest)
	if err != nil {
		t.Errorf("%v", err)
	}

	searchResult1, err := client.Search(searchRequest)
	if err != nil {
		t.Errorf("%v", err)
	}
	expTotal := uint64(1)
	actTotal := searchResult1.Total
	if expTotal != actTotal {
		t.Errorf("expected content to see %v, saw %v", expTotal, actTotal)
	}

	// delete document
	count, err = client.DeleteDocument([]string{"enwiki_1"})
	if err != nil {
		t.Errorf("%v", err)
	}
	expCount = 1
	actCount = count
	if expCount != actCount {
		t.Errorf("expected content to see %v, saw %v", expIndexConfig, actIndexConfig)
	}

	// get document
	fields1, err = client.GetDocument("enwiki_1")
	if err != errors.ErrNotFound {
		t.Errorf("%v", err)
	}

	// search
	searchResult1, err = client.Search(searchRequest)
	if err != nil {
		t.Errorf("%v", err)
	}
	expTotal = uint64(0)
	actTotal = searchResult1.Total
	if expTotal != actTotal {
		t.Errorf("expected content to see %v, saw %v", expTotal, actTotal)
	}
}
