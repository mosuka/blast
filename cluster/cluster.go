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
	"fmt"
	"math"
	"math/rand"
	"time"
)

var (
	STATE_ACTIVE = "active"
	STATE_DOWN   = "down"
)

type Node struct {
	State string `json:"state,omitempty"`
}

func NewNode() *Node {
	return &Node{
		State: STATE_DOWN,
	}
}

type Shard struct {
	Nodes map[string]*Node `json:"nodes,omitempty"`
	State string           `json:"state,omitempty"`
}

func NewShard() *Shard {
	return &Shard{
		Nodes: make(map[string]*Node),
		State: STATE_DOWN,
	}
}

func (m *Shard) AddNode(name string) error {
	_, exist := m.Nodes[name]
	if exist {
		return fmt.Errorf("%s already exists in this shard", name)
	}

	node := NewNode()
	m.Nodes[name] = node

	return nil
}

func (m *Shard) RemoveNode(name string) error {
	_, exist := m.Nodes[name]
	if !exist {
		return fmt.Errorf("%s does not exist in this shard", name)
	}

	delete(m.Nodes, name)

	return nil
}

func (m *Shard) GetActiveNodes() map[string]*Node {
	ret := make(map[string]*Node)

	for name, node := range m.Nodes {
		if node.State == STATE_ACTIVE {
			ret[name] = node
		}
	}

	return ret
}

func (m *Shard) GetActiveNode() map[string]*Node {
	activeNodes := m.GetActiveNodes()

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(activeNodes))

	ret := make(map[string]*Node)
	i := 0
	for name, node := range activeNodes {
		if index == i {
			ret[name] = node
			break
		} else {
			i++
		}
	}

	return ret
}

type Cluster struct {
	Shards map[string]*Shard `json:"shards,omitempty"`
	State  string            `json:"state,omitempty"`
}

func NewCluster() *Cluster {
	return &Cluster{
		Shards: make(map[string]*Shard),
		State:  STATE_DOWN,
	}
}

func (m *Cluster) AddShard(name string) error {
	_, exist := m.Shards[name]
	if exist {
		return fmt.Errorf("%s already exists in this cluster", name)
	}

	shard := NewShard()
	m.Shards[name] = shard

	return nil
}

func (m *Cluster) RemoveShard(name string) error {
	_, exist := m.Shards[name]
	if !exist {
		return fmt.Errorf("%s does not exist in this cluster", name)
	}

	delete(m.Shards, name)

	return nil
}

func (m *Cluster) GetMinimumShards() map[string]*Shard {
	min := math.MaxInt32
	for _, shard := range m.Shards {
		numNodes := len(shard.GetActiveNodes())
		if numNodes < min {
			min = numNodes
		}
	}

	ret := make(map[string]*Shard)
	for name, shard := range m.Shards {
		if len(shard.GetActiveNodes()) <= min {
			ret[name] = shard
		}
	}

	return ret
}

func (m *Cluster) GetActiveShards() map[string]*Shard {
	ret := make(map[string]*Shard)

	for name, shard := range m.Shards {
		if shard.State == STATE_ACTIVE {
			ret[name] = shard
		}
	}

	return ret
}

func (m *Cluster) AddNode(shardName string, nodeName string) error {
	_, exist := m.Shards[shardName]
	if !exist {
		return fmt.Errorf("%s does not exist in this cluster", shardName)
	}

	err := m.Shards[shardName].AddNode(nodeName)
	if err != nil {
		return err
	}

	return nil
}

func (m *Cluster) RemoveNode(shardName string, nodeName string) error {
	_, exist := m.Shards[shardName]
	if !exist {
		return fmt.Errorf("%s does not exist in this cluster", shardName)
	}

	err := m.Shards[shardName].RemoveNode(nodeName)
	if err != nil {
		return err
	}

	return nil
}

func (m *Cluster) GetSearchNodes() map[string]*Node {
	ret := make(map[string]*Node)

	for _, shard := range m.GetActiveShards() {
		for name, node := range shard.GetActiveNode() {
			ret[name] = node
		}
	}

	return ret
}
