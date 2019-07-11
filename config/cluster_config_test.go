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

import "testing"

func TestClusterConfig_SetManagerAddr(t *testing.T) {
	managerAddr := ":15000"
	cfg := DefaultClusterConfig()
	err := cfg.SetManagerAddr(managerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actManagerAddr, err := cfg.GetManagerAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if managerAddr != actManagerAddr {
		t.Fatalf("expected content to see %v, saw %v", managerAddr, actManagerAddr)
	}
}

func TestClusterConfig_GetManagerAddr(t *testing.T) {
	managerAddr := ":15000"
	cfg := DefaultClusterConfig()
	err := cfg.SetManagerAddr(managerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actManagerAddr, err := cfg.GetManagerAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if managerAddr != actManagerAddr {
		t.Fatalf("expected content to see %v, saw %v", managerAddr, actManagerAddr)
	}
}

func TestClusterConfig_SetClusterId(t *testing.T) {
	clusterId := "cluster1"
	cfg := DefaultClusterConfig()
	err := cfg.SetClusterId(clusterId)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actClusterId, err := cfg.GetClusterId()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if clusterId != actClusterId {
		t.Fatalf("expected content to see %v, saw %v", clusterId, actClusterId)
	}
}

func TestClusterConfig_GetClusterId(t *testing.T) {
	clusterId := "cluster1"
	cfg := DefaultClusterConfig()
	err := cfg.SetClusterId(clusterId)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actClusterId, err := cfg.GetClusterId()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if clusterId != actClusterId {
		t.Fatalf("expected content to see %v, saw %v", clusterId, actClusterId)
	}
}

func TestClusterConfig_SetPeerAddr(t *testing.T) {
	peerAddr := ":5000"
	cfg := DefaultClusterConfig()
	err := cfg.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actPeerAddr, err := cfg.GetPeerAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if peerAddr != actPeerAddr {
		t.Fatalf("expected content to see %v, saw %v", peerAddr, actPeerAddr)
	}
}

func TestClusterConfig_GetPeerAddr(t *testing.T) {
	peerAddr := ":5000"
	cfg := DefaultClusterConfig()
	err := cfg.SetPeerAddr(peerAddr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	actPeerAddr, err := cfg.GetPeerAddr()
	if err != nil {
		t.Fatalf("%v", err)
	}

	if peerAddr != actPeerAddr {
		t.Fatalf("expected content to see %v, saw %v", peerAddr, actPeerAddr)
	}
}
