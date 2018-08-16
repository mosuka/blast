//  Copyright (c) 2018 Minoru Osuka
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

package raft

import (
	"time"

	"github.com/hashicorp/raft"
)

type RaftConfig struct {
	Path                string        `json:"path,omitempty"`
	RetainSnapshotCount int           `json:"retain_snapshot_count,omitempty"`
	Timeout             time.Duration `json:"timeout,omitempty"`
	Config              *raft.Config  `json:"config,omitempty"`
}

func DefaultConfig() *RaftConfig {
	config := raft.DefaultConfig()
	config.LocalID = "node0"

	return &RaftConfig{
		Path:                "./data/raft",
		RetainSnapshotCount: 2,
		Timeout:             10 * time.Second,
		Config:              config,
	}
}
