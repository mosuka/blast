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

package manager

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/raft"
	"github.com/mosuka/blast/logutils"
	"github.com/mosuka/blast/protobuf/management"
)

func TestRaftFSM_GetNode(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	_ = fsm.SetNodeConfig(
		"node1",
		&management.Node{
			BindAddress: "2100",
			State:       raft.Leader.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5100",
				HttpAddress: "8100",
			},
		},
	)
	_ = fsm.SetNodeConfig(
		"node2",
		&management.Node{
			BindAddress: "2110",
			State:       raft.Follower.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5110",
				HttpAddress: "8110",
			},
		},
	)
	_ = fsm.SetNodeConfig(
		"node3",
		&management.Node{
			BindAddress: "2120",
			State:       raft.Follower.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5120",
				HttpAddress: "8120",
			},
		},
	)

	val1, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Fatalf("%v", err)
	}

	exp1 := &management.Node{
		BindAddress: "2110",
		State:       raft.Follower.String(),
		Metadata: &management.Metadata{
			GrpcAddress: "5110",
			HttpAddress: "8110",
		},
	}

	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Fatalf("expected content to see %v, saw %v", exp1, act1)
	}

}

func TestRaftFSM_SetNode(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	_ = fsm.SetNodeConfig(
		"node1",
		&management.Node{
			BindAddress: "2100",
			State:       raft.Leader.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5100",
				HttpAddress: "8100",
			},
		},
	)
	_ = fsm.SetNodeConfig(
		"node2",
		&management.Node{
			BindAddress: "2110",
			State:       raft.Follower.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5110",
				HttpAddress: "8110",
			},
		},
	)
	_ = fsm.SetNodeConfig(
		"node3",
		&management.Node{
			BindAddress: "2120",
			State:       raft.Follower.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5120",
				HttpAddress: "8120",
			},
		},
	)

	val1, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp1 := &management.Node{
		BindAddress: "2110",
		State:       raft.Follower.String(),
		Metadata: &management.Metadata{
			GrpcAddress: "5110",
			HttpAddress: "8110",
		},
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Fatalf("expected content to see %v, saw %v", exp1, act1)
	}

	_ = fsm.SetNodeConfig(
		"node2",
		&management.Node{
			BindAddress: "2110",
			State:       raft.Shutdown.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5110",
				HttpAddress: "8110",
			},
		},
	)

	val2, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp2 := &management.Node{
		BindAddress: "2110",
		State:       raft.Shutdown.String(),
		Metadata: &management.Metadata{
			GrpcAddress: "5110",
			HttpAddress: "8110",
		},
	}

	act2 := val2
	if !reflect.DeepEqual(exp2, act2) {
		t.Fatalf("expected content to see %v, saw %v", exp2, act2)
	}
}

func TestRaftFSM_DeleteNode(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	_ = fsm.SetNodeConfig(
		"node1",
		&management.Node{
			BindAddress: "2100",
			State:       raft.Leader.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5100",
				HttpAddress: "8100",
			},
		},
	)
	_ = fsm.SetNodeConfig(
		"node2",
		&management.Node{
			BindAddress: "2110",
			State:       raft.Follower.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5110",
				HttpAddress: "8110",
			},
		},
	)
	_ = fsm.SetNodeConfig(
		"node3",
		&management.Node{
			BindAddress: "2120",
			State:       raft.Follower.String(),
			Metadata: &management.Metadata{
				GrpcAddress: "5120",
				HttpAddress: "8120",
			},
		},
	)

	val1, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp1 := &management.Node{
		BindAddress: "2110",
		State:       raft.Follower.String(),
		Metadata: &management.Metadata{
			GrpcAddress: "5110",
			HttpAddress: "8110",
		},
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Fatalf("expected content to see %v, saw %v", exp1, act1)
	}

	err = fsm.DeleteNodeConfig("node2")
	if err != nil {
		t.Fatalf("%v", err)
	}

	val2, err := fsm.GetNodeConfig("node2")
	if err == nil {
		t.Fatalf("expected error: %v", err)
	}

	act1 = val2
	if reflect.DeepEqual(nil, act1) {
		t.Fatalf("expected content to see nil, saw %v", act1)
	}
}

func TestRaftFSM_Get(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = fsm.SetValue("/", map[string]interface{}{"a": 1}, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	value, err := fsm.GetValue("/a")
	if err != nil {
		t.Fatalf("%v", err)
	}

	expectedValue := 1
	actualValue := value
	if expectedValue != actualValue {
		t.Fatalf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestRaftFSM_Set(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	// set {"a": 1}
	err = fsm.SetValue("/", map[string]interface{}{
		"a": 1,
	}, false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val1, err := fsm.GetValue("/")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp1 := map[string]interface{}{
		"a": 1,
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Fatalf("expected content to see %v, saw %v", exp1, act1)
	}

	// merge {"a": "A"}
	_ = fsm.SetValue("/", map[string]interface{}{
		"a": "A",
	}, true)
	val2, err := fsm.GetValue("/")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp2 := map[string]interface{}{
		"a": "A",
	}
	act2 := val2
	if !reflect.DeepEqual(exp2, act2) {
		t.Fatalf("expected content to see %v, saw %v", exp2, act2)
	}

	// set {"a": {"b": "AB"}}
	err = fsm.SetValue("/", map[string]interface{}{
		"a": map[string]interface{}{
			"b": "AB",
		},
	}, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	val3, err := fsm.GetValue("/")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp3 := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "AB",
		},
	}
	act3 := val3
	if !reflect.DeepEqual(exp3, act3) {
		t.Fatalf("expected content to see %v, saw %v", exp3, act3)
	}

	// merge {"a": {"c": "AC"}}
	err = fsm.SetValue("/", map[string]interface{}{
		"a": map[string]interface{}{
			"c": "AC",
		},
	}, true)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val4, err := fsm.GetValue("/")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp4 := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "AB",
			"c": "AC",
		},
	}
	act4 := val4
	if !reflect.DeepEqual(exp4, act4) {
		t.Fatalf("expected content to see %v, saw %v", exp4, act4)
	}

	// set {"a": 1}
	err = fsm.SetValue("/", map[string]interface{}{
		"a": 1,
	}, false)
	if err != nil {
		t.Fatalf("%v", err)
	}
	val5, err := fsm.GetValue("/")
	if err != nil {
		t.Fatalf("%v", err)
	}
	exp5 := map[string]interface{}{
		"a": 1,
	}
	act5 := val5
	if !reflect.DeepEqual(exp5, act5) {
		t.Fatalf("expected content to see %v, saw %v", exp5, act5)
	}

	// TODO: merge {"a": {"c": "AC"}}
	//fsm.applySet("/", map[string]interface{}{
	//	"a": map[string]interface{}{
	//		"c": "AC",
	//	},
	//}, true)
	//val6, err := fsm.Get("/")
	//if err != nil {
	//	t.Fatalf("%v", err)
	//}
	//exp6 := map[string]interface{}{
	//	"a": map[string]interface{}{
	//		"c": "AC",
	//	},
	//}
	//act6 := val6
	//if !reflect.DeepEqual(exp6, act6) {
	//	t.Fatalf("expected content to see %v, saw %v", exp6, act6)
	//}
}

func TestRaftFSM_Delete(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Fatalf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Fatalf("%v", err)
		}
	}()
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = fsm.SetValue("/", map[string]interface{}{"a": 1}, false)
	if err != nil {
		t.Fatalf("%v", err)
	}

	value, err := fsm.GetValue("/a")
	if err != nil {
		t.Fatalf("%v", err)
	}

	expectedValue := 1
	actualValue := value
	if expectedValue != actualValue {
		t.Fatalf("expected content to see %v, saw %v", expectedValue, actualValue)
	}

	err = fsm.DeleteValue("/a")
	if err != nil {
		t.Fatalf("%v", err)
	}

	value, err = fsm.GetValue("/a")
	if err == nil {
		t.Fatalf("expected nil: %v", err)
	}

	actualValue = value
	if nil != actualValue {
		t.Fatalf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}
