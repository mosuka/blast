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

package cluster

import (
	"encoding/json"
	"testing"
)

func TestNode(t *testing.T) {
	n := NewNode()

	actualNodeState := n.State
	expectNodeState := STATE_DOWN
	if actualNodeState != expectNodeState {
		t.Fatalf("actual=%v, expect=%v", actualNodeState, expectNodeState)
	}
}

func TestShard(t *testing.T) {
	s := NewShard()

	actualShardState := s.State
	expectShardState := STATE_DOWN
	if actualShardState != expectShardState {
		t.Fatalf("actual=%v, expect=%v", actualShardState, expectShardState)
	}

	actualNumNodes := len(s.Nodes)
	expectNumNodes := 0
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}

	s.AddNode("node1")

	actualNumNodes = len(s.Nodes)
	expectNumNodes = 1
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}

	s.AddNode("node2")

	actualNumNodes = len(s.Nodes)
	expectNumNodes = 2
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}

	s.AddNode("node3")

	actualNumNodes = len(s.Nodes)
	expectNumNodes = 3
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}

	s.RemoveNode("node2")

	actualNumNodes = len(s.Nodes)
	expectNumNodes = 2
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}

	s.RemoveNode("node1")

	actualNumNodes = len(s.Nodes)
	expectNumNodes = 1
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}

	s.RemoveNode("node3")

	actualNumNodes = len(s.Nodes)
	expectNumNodes = 0
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}
}

func TestCluster(t *testing.T) {
	c := NewCluster()

	c.AddShard("shard1")
	c.AddNode("shard1", "node1")
	c.AddNode("shard1", "node2")

	c.AddShard("shard2")
	c.AddNode("shard2", "node3")
	c.AddNode("shard2", "node4")

	actualNumShards := len(c.Shards)
	expectNumShards := 2
	if actualNumShards != expectNumShards {
		t.Fatalf("actual=%v, expect=%v", actualNumShards, expectNumShards)
	}

	actualActiveNodes := c.Shards["shard1"].GetActiveNodes()
	expectActiveNodes := make(map[string]*Node)
	actualActiveNodesBytes, _ := json.Marshal(actualActiveNodes)
	expectActiveNodesBytes, _ := json.Marshal(expectActiveNodes)
	if string(actualActiveNodesBytes) != string(expectActiveNodesBytes) {
		t.Fatalf("actual=%v, expect=%v", actualActiveNodes, expectActiveNodes)
	}

	c.Shards["shard1"].Nodes["node1"].State = STATE_ACTIVE

	actualActiveNodes = c.Shards["shard1"].GetActiveNodes()
	actualActiveNodesBytes, _ = json.Marshal(actualActiveNodes)
	actualActiveNodesStr := string(actualActiveNodesBytes)
	expectActiveNodesStr := `{"node1":{"state":"active"}}`
	if actualActiveNodesStr != expectActiveNodesStr {
		t.Fatalf("actual=%v, expect=%v", actualActiveNodesStr, expectActiveNodesStr)
	}

	c.RemoveNode("shard2", "node3")

	actualNumNodes := len(c.Shards["shard2"].Nodes)
	expectNumNodes := 1
	if actualNumNodes != expectNumNodes {
		t.Fatalf("actual=%v, expect=%v", actualNumNodes, expectNumNodes)
	}

	actualMinShards := c.GetMinimumShards()
	actualMinShardsBytes, _ := json.Marshal(actualMinShards)
	actualMinShardsStr := string(actualMinShardsBytes)
	expectMinShardsStr := `{"shard2":{"nodes":{"node4":{"state":"down"}},"state":"down"}}`
	if actualMinShardsStr != expectMinShardsStr {
		t.Fatalf("actual=%v, expect=%v", actualMinShardsStr, expectMinShardsStr)
	}

}
