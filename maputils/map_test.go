package maputils

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_splitKey(t *testing.T) {
	key1 := "/a/b/c/d"
	keys1 := splitKey(key1)
	exp1 := []string{"a", "b", "c", "d"}
	act1 := keys1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	key2 := "/"
	keys2 := splitKey(key2)
	exp2 := make([]string, 0)
	act2 := keys2
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	key3 := ""
	keys3 := splitKey(key3)
	exp3 := make([]string, 0)
	act3 := keys3
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", exp3, act3)
	}
}

func Test_makeSelector(t *testing.T) {
	key1 := "/a/b/c/d"
	selector1 := makeSelector(key1)
	exp1 := "a.b.c.d"
	act1 := selector1
	if exp1 != act1 {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	key2 := "/"
	selector2 := makeSelector(key2)
	exp2 := ""
	act2 := selector2
	if exp2 != act2 {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	key3 := ""
	selector3 := makeSelector(key3)
	exp3 := ""
	act3 := selector3
	if exp3 != act3 {
		t.Errorf("expected content to see %v, saw %v", exp3, act3)
	}
}

func Test_normalize(t *testing.T) {
	data1 := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	val1 := normalize(data1)
	exp1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}
}

func Test_makeMap(t *testing.T) {
	val1 := makeMap("/a/b/c", "C").(Map)
	exp1 := Map{
		"a": Map{
			"b": Map{
				"c": "C",
			},
		},
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	val2 := makeMap("a/b", map[string]interface{}{"c": "C"}).(Map)
	exp2 := Map{
		"a": Map{
			"b": Map{
				"c": "C",
			},
		},
	}
	act2 := val2
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}
}

func Test_Has(t *testing.T) {
	map1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}

	val1, err := map1.Has("a/b/c")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := true
	act1 := val1
	if exp1 != act1 {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	val2, err := map1.Get("a/b/f")
	if err != ErrNotFound {
		t.Errorf("%v", err)
	}
	exp2 := false
	act2 := val2
	if exp2 == act2 {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}
}

func Test_Set(t *testing.T) {
	map1 := Map{}

	err := map1.Set("/", Map{"a": "A"})
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := Map{
		"a": "A",
	}
	act1 := map1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	err = map1.Set("/", Map{"A": "a"})
	if err != nil {
		t.Errorf("%v", err)
	}
	exp2 := Map{
		"A": "a",
	}
	act2 := map1
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	err = map1.Set("/", Map{"A": 1})
	if err != nil {
		t.Errorf("%v", err)
	}
	exp3 := Map{
		"A": 1,
	}
	act3 := map1
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	err = map1.Set("/A", "AAA")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp4 := Map{
		"A": "AAA",
	}
	act4 := map1
	if !reflect.DeepEqual(exp4, act4) {
		t.Errorf("expected content to see %v, saw %v", exp4, act4)
	}

	err = map1.Set("/B", "BBB")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp5 := Map{
		"A": "AAA",
		"B": "BBB",
	}
	act5 := map1
	if !reflect.DeepEqual(exp5, act5) {
		t.Errorf("expected content to see %v, saw %v", exp5, act5)
	}

	err = map1.Set("/C", map[string]interface{}{"D": "CCC-DDD"})
	if err != nil {
		t.Errorf("%v", err)
	}
	exp6 := Map{
		"A": "AAA",
		"B": "BBB",
		"C": Map{
			"D": "CCC-DDD",
		},
	}
	act6 := map1
	if !reflect.DeepEqual(exp6, act6) {
		t.Errorf("expected content to see %v, saw %v", exp6, act6)
	}
}

func Test_Merge(t *testing.T) {
	map1 := Map{}

	err := map1.Merge("/", Map{"a": "A"})
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := Map{
		"a": "A",
	}
	act1 := map1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	err = map1.Merge("/a", "a")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp2 := Map{
		"a": "a",
	}
	act2 := map1
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}

	err = map1.Merge("/", Map{"a": 1})
	if err != nil {
		t.Errorf("%v", err)
	}
	exp3 := Map{
		"a": 1,
	}
	act3 := map1
	if !reflect.DeepEqual(exp3, act3) {
		t.Errorf("expected content to see %v, saw %v", exp3, act3)
	}

	err = map1.Merge("/", Map{"b": 2})
	if err != nil {
		t.Errorf("%v", err)
	}
	exp4 := Map{
		"a": 1,
		"b": 2,
	}
	act4 := map1
	if !reflect.DeepEqual(exp4, act4) {
		t.Errorf("expected content to see %v, saw %v", exp4, act4)
	}

	err = map1.Merge("/c", 3)
	if err != nil {
		t.Errorf("%v", err)
	}
	exp5 := Map{
		"a": 1,
		"b": 2,
		"c": 3,
	}
	act5 := map1
	if !reflect.DeepEqual(exp5, act5) {
		t.Errorf("expected content to see %v, saw %v", exp5, act5)
	}

}

func Test_Get(t *testing.T) {
	map1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}

	val1, err := map1.Get("a/b/c")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := "abc"
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

	val2, err := map1.Get("a")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp2 := Map{
		"b": Map{
			"c": "abc",
			"d": "abd",
		},
		"e": []interface{}{
			"ae1",
			"ae2",
		},
	}
	act2 := val2
	if !reflect.DeepEqual(exp2, act2) {
		t.Errorf("expected content to see %v, saw %v", exp2, act2)
	}
}

func Test_Delete(t *testing.T) {
	map1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}

	err := map1.Delete("a/b/c")
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := Map{
		"a": Map{
			"b": Map{
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	act1 := map1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}

}

func Test_ToMap(t *testing.T) {
	map1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	val1 := map1.ToMap()
	exp1 := map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	act1 := val1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}
}

func Test_FromMap(t *testing.T) {
	map1 := FromMap(map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	})
	exp1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	act1 := map1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}
}

func Test_ToJSON(t *testing.T) {
	map1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	val1, err := map1.ToJSON()
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := []byte(`{"a":{"b":{"c":"abc","d":"abd"},"e":["ae1","ae2"]}}`)
	act1 := val1
	if !bytes.Equal(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}
}

func Test_FromJSON(t *testing.T) {
	map1, err := FromJSON([]byte(`{"a":{"b":{"c":"abc","d":"abd"},"e":["ae1","ae2"]}}`))
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	act1 := map1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}
}

func Test_ToYAML(t *testing.T) {
	map1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}

	val1, err := map1.ToYAML()
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := []byte(`a:
  b:
    c: abc
    d: abd
  e:
  - ae1
  - ae2
`)
	act1 := val1
	if !bytes.Equal(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}
}

func Test_FromYAML(t *testing.T) {
	map1, err := FromYAML([]byte(`a:
  b:
    c: abc
    d: abd
  e:
  - ae1
  - ae2
`))
	if err != nil {
		t.Errorf("%v", err)
	}
	exp1 := Map{
		"a": Map{
			"b": Map{
				"c": "abc",
				"d": "abd",
			},
			"e": []interface{}{
				"ae1",
				"ae2",
			},
		},
	}
	act1 := map1
	if !reflect.DeepEqual(exp1, act1) {
		t.Errorf("expected content to see %v, saw %v", exp1, act1)
	}
}

//func Test_Get(t *testing.T) {
//	data1 := objx.Map{
//		"a": objx.Map{
//			"b": objx.Map{
//				"c": "abc",
//				"d": "abd",
//			},
//			"e": []interface{}{
//				"ae1",
//				"ae2",
//			},
//		},
//	}
//	key1 := "/"
//	val1, err := Get(data1, key1)
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//	exp1 := map[string]interface{}{
//		"a": map[string]interface{}{
//			"b": map[string]interface{}{
//				"c": "abc",
//				"d": "abd",
//			},
//			"e": []interface{}{
//				"ae1",
//				"ae2",
//			},
//		},
//	}
//	act1 := val1
//	if !reflect.DeepEqual(exp1, act1) {
//		t.Errorf("expected content to see %v, saw %v", exp1, act1)
//	}
//
//	key2 := "/a"
//	val2, err := Get(data1, key2)
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//	exp2 := map[string]interface{}{
//		"b": map[string]interface{}{
//			"c": "abc",
//			"d": "abd",
//		},
//		"e": []interface{}{
//			"ae1",
//			"ae2",
//		},
//	}
//	act2 := val2
//	if !reflect.DeepEqual(exp2, act2) {
//		t.Errorf("expected content to see %v, saw %v", exp2, act2)
//	}
//}

//func Test_Set(t *testing.T) {
//	data := map[string]interface{}{}
//
//	data, err := Set(data, "/", map[string]interface{}{"a": 1}, true)
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//
//	exp1 := 1
//	act1 := val1
//	if exp1 != act1 {
//		t.Errorf("expected content to see %v, saw %v", exp1, act1)
//	}
//
//	fsm.applySet("/b/bb", map[string]interface{}{"b": 1}, false)
//
//	val2, err := fsm.Get("/b")
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//
//	exp2 := map[string]interface{}{"bb": map[string]interface{}{"b": 1}}
//	act2 := val2.(map[string]interface{})
//	if !reflect.DeepEqual(exp2, act2) {
//		t.Errorf("expected content to see %v, saw %v", exp2, act2)
//	}
//
//	fsm.applySet("/", map[string]interface{}{"a": 1}, false)
//
//	val3, err := fsm.Get("/")
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//
//	exp3 := map[string]interface{}{"a": 1}
//	act3 := val3
//	if !reflect.DeepEqual(exp3, act3) {
//		t.Errorf("expected content to see %v, saw %v", exp3, act3)
//	}
//
//	fsm.applySet("/", map[string]interface{}{"b": 2}, true)
//
//	val4, err := fsm.Get("/")
//	if err != nil {
//		t.Errorf("%v", err)
//	}
//
//	exp4 := map[string]interface{}{"a": 1, "b": 2}
//	act4 := val4
//	if !reflect.DeepEqual(exp4, act4) {
//		t.Errorf("expected content to see %v, saw %v", exp4, act4)
//	}
//}
