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

package config

import (
	"fmt"
	"testing"

	"github.com/mosuka/blast/strutils"
)

func TestDefaultNodeConfig(t *testing.T) {
	expConfig := &NodeConfig{
		NodeId:          fmt.Sprintf("node-%s", strutils.RandStr(5)),
		BindAddr:        ":2000",
		GRPCAddr:        ":5000",
		HTTPAddr:        ":8000",
		DataDir:         "/tmp/blast",
		RaftStorageType: "boltdb",
	}
	actConfig := DefaultNodeConfig()

	if expConfig.BindAddr != actConfig.BindAddr {
		t.Fatalf("expected content to see %v, saw %v", expConfig.BindAddr, actConfig.BindAddr)
	}

	if expConfig.GRPCAddr != actConfig.GRPCAddr {
		t.Fatalf("expected content to see %v, saw %v", expConfig.GRPCAddr, actConfig.GRPCAddr)
	}

	if expConfig.HTTPAddr != actConfig.HTTPAddr {
		t.Fatalf("expected content to see %v, saw %v", expConfig.HTTPAddr, actConfig.HTTPAddr)
	}

	if expConfig.DataDir != actConfig.DataDir {
		t.Fatalf("expected content to see %v, saw %v", expConfig.DataDir, actConfig.DataDir)
	}

	if expConfig.RaftStorageType != actConfig.RaftStorageType {
		t.Fatalf("expected content to see %v, saw %v", expConfig.RaftStorageType, actConfig.RaftStorageType)
	}
}
