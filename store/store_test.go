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

package store

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestStore_Reader(t *testing.T) {
	config := DefaultStoreConfig()

	tmpDir, err := ioutil.TempDir("/tmp", "blast")
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	config.Dir = tmpDir

	defer os.RemoveAll(tmpDir)

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	writer, err := store.Writer()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer writer.Close()

	writer.Put([]byte("1"), []byte("AAA"))

	reader, err := store.Reader()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer reader.Close()

	b, err := reader.Get([]byte("1"))

	expectedValue := "AAA"
	actualValue := string(b)

	if expectedValue != actualValue {
		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
	}
}

func TestStore_Writer(t *testing.T) {
	config := DefaultStoreConfig()

	tmpDir, err := ioutil.TempDir("/tmp", "blast")
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	config.Dir = tmpDir

	defer os.RemoveAll(tmpDir)

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	writer, err := store.Writer()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer writer.Close()

	writer.Put([]byte("1"), []byte("AAA"))

	reader, err := store.Reader()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer reader.Close()

	b, err := reader.Get([]byte("1"))

	expectedValue := "AAA"
	actualValue := string(b)

	if expectedValue != actualValue {
		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
	}
}

func TestStore_Iterator(t *testing.T) {
	config := DefaultStoreConfig()

	tmpDir, err := ioutil.TempDir("/tmp", "blast")
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	config.Dir = tmpDir

	defer os.RemoveAll(tmpDir)

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	writer, err := store.Writer()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer writer.Close()

	writer.Put([]byte("1"), []byte("AAA"))
	writer.Put([]byte("2"), []byte("BBB"))
	writer.Put([]byte("3"), []byte("CCC"))

	iterator, err := store.Iterator()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer iterator.Close()

	expectedCnt := 3
	actualCnt := 0
	for iterator.Next() {
		t.Logf("%s=%v", iterator.Key(), string(iterator.Value()))
		actualCnt++
	}

	if expectedCnt != actualCnt {
		t.Errorf("unexpected error.  expected %d, actual %d", expectedCnt, actualCnt)
	}
}

func TestStore_Bulker(t *testing.T) {
	config := DefaultStoreConfig()

	tmpDir, err := ioutil.TempDir("/tmp", "blast")
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	config.Dir = tmpDir

	defer os.RemoveAll(tmpDir)

	store, err := NewStore(config)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}

	bulker, err := store.Bulker(5)
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	//defer bulker.Close()

	bulker.Put([]byte("1"), []byte("AAA"))
	bulker.Put([]byte("2"), []byte("BBB"))
	bulker.Put([]byte("3"), []byte("CCC"))
	bulker.Delete([]byte("2"))
	bulker.Close()

	iterator, err := store.Iterator()
	if err != nil {
		t.Fatalf("unexpected error. %v", err)
	}
	defer iterator.Close()

	expectedCnt := 2
	actualCnt := 0
	for iterator.Next() {
		t.Logf("%s=%v", iterator.Key(), string(iterator.Value()))
		actualCnt++
	}

	if expectedCnt != actualCnt {
		t.Errorf("unexpected error.  expected %d, actual %d", expectedCnt, actualCnt)
	}
}
