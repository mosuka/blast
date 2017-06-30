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

package grpc

import (
	"github.com/mosuka/blast/util"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestBlastGRPCServer(t *testing.T) {
	dir, _ := os.Getwd()

	port := 0
	path, _ := ioutil.TempDir("/tmp", "indigo")
	indexMappingPath := dir + "/../../example/index_mapping.json"
	indexType := "upside_down"
	kvstore := "boltdb"
	kvconfigPath := dir + "/../../example/kvconfig.json"

	indexMappingFile, err := os.Open(indexMappingPath)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer indexMappingFile.Close()

	indexMapping, err := util.NewIndexMapping(indexMappingFile)
	if err != nil {
		t.Errorf("could not load IndexMapping: %v", err)
	}

	kvconfigFile, err := os.Open(kvconfigPath)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer kvconfigFile.Close()

	kvconfig, err := util.NewKvconfig(kvconfigFile)
	if err != nil {
		t.Errorf("could not load kvconfig %v", err)
	}
	kvconfig["path"] = path + "/store"

	gRPCServer := NewBlastGRPCServer(port, path, indexMapping, indexType, kvstore, kvconfig)

	if gRPCServer == nil {
		t.Fatalf("unexpected error.  expected not nil, actual %v", gRPCServer)
	}

	err = gRPCServer.Start(true)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	time.Sleep(10 * time.Second)

	err = gRPCServer.Stop(false)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	os.RemoveAll(path)
}
