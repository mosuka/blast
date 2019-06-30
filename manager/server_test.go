package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/mosuka/blast/errors"

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

	nodeId := "manager1"
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

	server, err := NewServer(nodeId, metadata, peerAddr, indexConfig, logger, httpAccessLogger)
	defer func() {
		server.Stop()
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// start server
	server.Start()

	// sleep
	time.Sleep(10 * time.Second)

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

	// get index mapping
	val5, err := client.GetState("index_config/index_mapping")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp5 := indexMappingMap
	act5 := *val5.(*map[string]interface{})
	if !reflect.DeepEqual(exp5, act5) {
		t.Errorf("expected content to see %v, saw %v", exp5, act5)
	}

	// get index type
	val6, err := client.GetState("index_config/index_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp6 := indexType
	act6 := *val6.(*string)
	if exp6 != act6 {
		t.Errorf("expected content to see %v, saw %v", exp6, act6)
	}

	// get index storage type
	val7, err := client.GetState("index_config/index_storage_type")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp7 := indexStorageType
	act7 := *val7.(*string)
	if exp7 != act7 {
		t.Errorf("expected content to see %v, saw %v", exp7, act7)
	}

	// set value
	err = client.SetState("test/key8", "val8")
	if err != nil {
		t.Errorf("%v", err)
	}
	val8, err := client.GetState("test/key8")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp8 := "val8"
	act8 := *val8.(*string)
	if exp8 != act8 {
		t.Errorf("expected content to see %v, saw %v", exp8, act8)
	}

	// delete value
	err = client.DeleteState("test/key8")
	if err != nil {
		t.Errorf("%v", err)
	}
	val9, err := client.GetState("test/key8")
	if err != errors.ErrNotFound {
		t.Errorf("%v", err)
	}
	if val9 != nil {
		t.Errorf("%v", err)
	}

	// delete non-existing data
	err = client.DeleteState("test/non-existing")
	if err == nil {
		t.Errorf("%v", err)
	}
}
