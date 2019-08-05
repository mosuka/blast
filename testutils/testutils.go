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

package testutils

import (
	"io/ioutil"
	"net"

	"github.com/mosuka/blast/config"
	"github.com/mosuka/blast/indexutils"
)

func TmpDir() string {
	tmp, _ := ioutil.TempDir("", "")
	return tmp
}

func TmpPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return -1
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1
	}

	defer func() {
		_ = l.Close()
	}()

	return l.Addr().(*net.TCPAddr).Port
}

//func TmpNodeConfig() *config.NodeConfig {
//	c := config.DefaultNodeConfig()
//
//	c.BindAddr = fmt.Sprintf(":%d", TmpPort())
//	c.GRPCAddr = fmt.Sprintf(":%d", TmpPort())
//	c.HTTPAddr = fmt.Sprintf(":%d", TmpPort())
//	c.DataDir = TmpDir()
//
//	return c
//}

func TmpIndexConfig(indexMappingFile string, indexType string, indexStorageType string) (*config.IndexConfig, error) {
	indexMapping, err := indexutils.NewIndexMappingFromFile(indexMappingFile)
	if err != nil {
		return config.DefaultIndexConfig(), err
	}

	indexConfig := &config.IndexConfig{
		IndexMapping:     indexMapping,
		IndexType:        indexType,
		IndexStorageType: indexStorageType,
	}

	return indexConfig, nil
}
