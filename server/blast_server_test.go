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
	"github.com/mosuka/blast/util"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestBlastServer(t *testing.T) {
	dir, _ := os.Getwd()

	port := 0
	indexPath, _ := ioutil.TempDir("/tmp", "blast")
	indexPath = indexPath + "/index/data"
	indexMappingPath := dir + "/../etc/index_mapping.json"
	indexType := "upside_down"
	kvstore := "boltdb"
	kvconfigPath := dir + "/../etc/kvconfig.json"
	etcdEndpoints := []string{}
	etcdDialTimeout := 5000
	etcdRequestTimeout := 5000
	cluster := ""
	shard := ""

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

	gRPCServer, err := NewBlastServer(port, indexPath, indexMapping, indexType, kvstore, kvconfig, etcdEndpoints, etcdDialTimeout, etcdRequestTimeout, cluster, shard)

	if gRPCServer == nil {
		t.Fatalf("unexpected error. expected not nil, actual %v", gRPCServer)
	}

	err = gRPCServer.Start()

	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	time.Sleep(10 * time.Second)

	err = gRPCServer.Stop()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	os.RemoveAll(indexPath)
}
