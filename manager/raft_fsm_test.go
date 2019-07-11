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

	"github.com/mosuka/blast/logutils"
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
		t.Errorf("%v", err)
	}

	_ = fsm.applySetNodeConfig("node1", map[string]interface{}{
		"bind_addr": ":16060",
		"grpc_addr": ":17070",
		"http_addr": ":18080",
	})
	_ = fsm.applySetNodeConfig("node2", map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
	})
	_ = fsm.applySetNodeConfig("node3", map[string]interface{}{
		"bind_addr": ":16062",
		"grpc_addr": ":17072",
		"http_addr": ":18082",
	})

	val1, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Errorf("%v", err)
	}

	exp1 := map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

}

func TestRaftFSM_SetNode(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	_ = fsm.applySetNodeConfig("node1", map[string]interface{}{
		"bind_addr": ":16060",
		"grpc_addr": ":17070",
		"http_addr": ":18080",
	})
	_ = fsm.applySetNodeConfig("node2", map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
	})
	_ = fsm.applySetNodeConfig("node3", map[string]interface{}{
		"bind_addr": ":16062",
		"grpc_addr": ":17072",
		"http_addr": ":18082",
	})

	val1, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	_ = fsm.applySetNodeConfig("node2", map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
		"leader":    true,
	})

	val2, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp2 := map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
		"leader":    true,
	}
	act2 := val2
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}
}

func TestRaftFSM_DeleteNode(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	_ = fsm.applySetNodeConfig("node1", map[string]interface{}{
		"bind_addr": ":16060",
		"grpc_addr": ":17070",
		"http_addr": ":18080",
	})
	_ = fsm.applySetNodeConfig("node2", map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
	})
	_ = fsm.applySetNodeConfig("node3", map[string]interface{}{
		"bind_addr": ":16062",
		"grpc_addr": ":17072",
		"http_addr": ":18082",
	})

	val1, err := fsm.GetNodeConfig("node2")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := map[string]interface{}{
		"bind_addr": ":16061",
		"grpc_addr": ":17071",
		"http_addr": ":18081",
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	err = fsm.applyDeleteNodeConfig("node2")
	if err != nil {
		t.Errorf("%v", err)
	}

	val2, err := fsm.GetNodeConfig("node2")
	if err == nil {
		t.Errorf("expected error: %v", err)
	}

	act1 = val2
	if reflect.DeepEqual(nil, act1) {
		t.Errorf("expected content to see nil, saw %v", act1)
	}
}

func TestRaftFSM_Get(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	err = fsm.applySet("/", map[string]interface{}{"a": 1}, false)
	if err != nil {
		t.Errorf("%v", err)
	}

	value, err := fsm.Get("/a")
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedValue := 1
	actualValue := value
	if expectedValue != actualValue {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}

func TestRaftFSM_Set(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	// set {"a": 1}
	err = fsm.applySet("/", map[string]interface{}{
		"a": 1,
	}, false)
	if err != nil {
		t.Errorf("%v", err)
	}
	val1, err := fsm.Get("/")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := map[string]interface{}{
		"a": 1,
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	// merge {"a": "A"}
	_ = fsm.applySet("/", map[string]interface{}{
		"a": "A",
	}, true)
	if err != nil {
		t.Errorf("%v", err)
	}
	val2, err := fsm.Get("/")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp2 := map[string]interface{}{
		"a": "A",
	}
	act2 := val2
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	// set {"a": {"b": "AB"}}
	err = fsm.applySet("/", map[string]interface{}{
		"a": map[string]interface{}{
			"b": "AB",
		},
	}, false)
	if err != nil {
		t.Errorf("%v", err)
	}

	val3, err := fsm.Get("/")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp3 := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "AB",
		},
	}
	act3 := val3
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", exp3, act3)
	}

	// merge {"a": {"c": "AC"}}
	err = fsm.applySet("/", map[string]interface{}{
		"a": map[string]interface{}{
			"c": "AC",
		},
	}, true)
	if err != nil {
		t.Errorf("%v", err)
	}
	val4, err := fsm.Get("/")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp4 := map[string]interface{}{
		"a": map[string]interface{}{
			"b": "AB",
			"c": "AC",
		},
	}
	act4 := val4
	if !reflect.DeepEqual(exp4, act4) {
		t.Errorf("expected content to see %v, saw %v", exp4, act4)
	}

	// set {"a": 1}
	err = fsm.applySet("/", map[string]interface{}{
		"a": 1,
	}, false)
	if err != nil {
		t.Errorf("%v", err)
	}
	val5, err := fsm.Get("/")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp5 := map[string]interface{}{
		"a": 1,
	}
	act5 := val5
	if !reflect.DeepEqual(exp5, act5) {
		t.Errorf("expected content to see %v, saw %v", exp5, act5)
	}

	// TODO: merge {"a": {"c": "AC"}}
	//fsm.applySet("/", map[string]interface{}{
	//	"a": map[string]interface{}{
	//		"c": "AC",
	//	},
	//}, true)
	//val6, err := fsm.Get("/")
	//if err != nil {
	//	t.Errorf("%v", err)
	//}
	//exp6 := map[string]interface{}{
	//	"a": map[string]interface{}{
	//		"c": "AC",
	//	},
	//}
	//act6 := val6
	//if !reflect.DeepEqual(exp6, act6) {
	//	t.Errorf("expected content to see %v, saw %v", exp6, act6)
	//}
}

func TestRaftFSM_Delete(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("%v", err)
	}
	defer func() {
		err := os.RemoveAll(tmp)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	logger := logutils.NewLogger("DEBUG", "", 100, 5, 3, false)

	fsm, err := NewRaftFSM(tmp, logger)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = fsm.Start()
	defer func() {
		err := fsm.Stop()
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	if err != nil {
		t.Errorf("%v", err)
	}

	err = fsm.applySet("/", map[string]interface{}{"a": 1}, false)
	if err != nil {
		t.Errorf("%v", err)
	}

	value, err := fsm.Get("/a")
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedValue := 1
	actualValue := value
	if expectedValue != actualValue {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}

	err = fsm.applyDelete("/a")
	if err != nil {
		t.Errorf("%v", err)
	}

	value, err = fsm.Get("/a")
	if err == nil {
		t.Errorf("expected nil: %v", err)
	}

	actualValue = value
	if nil != actualValue {
		t.Errorf("expected content to see %v, saw %v", expectedValue, actualValue)
	}
}
