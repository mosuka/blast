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
	"reflect"
	"testing"
)

func TestDefaultClusterConfig(t *testing.T) {
	exp := &ClusterConfig{}
	act := DefaultClusterConfig()
	if !reflect.DeepEqual(exp, act) {
		t.Fatalf("expected content to see %v, saw %v", exp, act)
	}

	expManagerAddr := ""
	actManagerAddr := act.ManagerAddr
	if expManagerAddr != actManagerAddr {
		t.Fatalf("expected content to see %v, saw %v", expManagerAddr, actManagerAddr)
	}

	expClusterId := ""
	actClusterId := act.ClusterId
	if expClusterId != actClusterId {
		t.Fatalf("expected content to see %v, saw %v", expClusterId, actClusterId)
	}

	expPeerAddr := ""
	actPeerAddr := act.PeerAddr
	if expPeerAddr != actPeerAddr {
		t.Fatalf("expected content to see %v, saw %v", expPeerAddr, actPeerAddr)
	}
}

func TestClusterConfig_1(t *testing.T) {
	expConfig := &ClusterConfig{
		ManagerAddr: ":12000",
		ClusterId:   "cluster1",
		PeerAddr:    ":5000",
	}
	actConfig := DefaultClusterConfig()
	actConfig.ManagerAddr = ":12000"
	actConfig.ClusterId = "cluster1"
	actConfig.PeerAddr = ":5000"

	expManagerAddr := expConfig.ManagerAddr
	actManagerAddr := actConfig.ManagerAddr
	if expManagerAddr != actManagerAddr {
		t.Fatalf("expected content to see %v, saw %v", expManagerAddr, actManagerAddr)
	}

	expClusterId := expConfig.ClusterId
	actClusterId := actConfig.ClusterId
	if expClusterId != actClusterId {
		t.Fatalf("expected content to see %v, saw %v", expClusterId, actClusterId)
	}

	expPeerAddr := expConfig.PeerAddr
	actPeerAddr := actConfig.PeerAddr
	if expPeerAddr != actPeerAddr {
		t.Fatalf("expected content to see %v, saw %v", expPeerAddr, actPeerAddr)
	}
}
