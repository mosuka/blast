// Copyright 2015 ikawaha
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// 	You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lattice

import "sync"

// BosEosID represents Reserved identifier of node id.
const BosEosID int = -1

// NodeClass codes.
const (
	DUMMY NodeClass = iota
	KNOWN
	UNKNOWN
	USER
)

// NodeClass represents a node type.
type NodeClass int

// String returns a string representation of a node class.
func (nc NodeClass) String() string {
	switch nc {
	case DUMMY:
		return "DUMMY"
	case KNOWN:
		return "KNOWN"
	case UNKNOWN:
		return "UNKNOWN"
	case USER:
		return "USER"
	}
	return "UNDEF"
}

type node struct {
	ID      int
	Start   int
	Class   NodeClass
	Cost    int32
	Left    int32
	Right   int32
	Weight  int32
	Surface string
	Prev    *node
}

var nodePool = sync.Pool{
	New: func() interface{} {
		return new(node)
	},
}

func newNode() *node {
	return nodePool.Get().(*node)
}
