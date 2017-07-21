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

package service

import (
	"github.com/mosuka/blast/util"
	"io/ioutil"
	"os"
	"testing"
)

func TestBlastGRPCService(t *testing.T) {
	dir, _ := os.Getwd()

	path, _ := ioutil.TempDir("/tmp", "indigo")
	indexMappingPath := dir + "/../etc/index_mapping.json"
	indexType := "upside_down"
	kvstore := "boltdb"
	kvconfigPath := dir + "/../etc/kvconfig.json"

	indexMappingFile, err := os.Open(indexMappingPath)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer indexMappingFile.Close()

	indexMapping, err := util.NewIndexMapping(indexMappingFile)
	if err != nil {
		t.Errorf("could not load IndexMapping : %v", err)
	}

	kvconfigFile, err := os.Open(kvconfigPath)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer kvconfigFile.Close()

	kvconfig, err := util.NewKvconfig(kvconfigFile)
	if err != nil {
		t.Errorf("could not load kvconfig : %v", err)
	}
	kvconfig["path"] = path + "/store"

	s := NewBlastGRPCService(path, indexMapping, indexType, kvstore, kvconfig)
	if s == nil {
		t.Fatalf("unexpected error.  expected not nil, actual %v", s)
	}

	if s.Path != path {
		t.Errorf("unexpected error.  expected %v, actual %v", path, s.Path)
	}
}
