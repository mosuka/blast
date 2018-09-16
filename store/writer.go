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
	"time"

	"github.com/boltdb/bolt"
)

type Writer struct {
	store *Store
}

func NewWriter(s *Store) (*Writer, error) {
	return &Writer{
		store: s,
	}, nil
}

func (w *Writer) Put(key []byte, value []byte) error {
	start := time.Now()
	defer Metrics(start, "Writer", "Put")

	var err error

	if err = w.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(Bucket))
		err := bucket.Put(key, value)
		if err != nil {
			w.store.logger.Printf("[ERR] boltdb: Failed to put data: %s: %v: %v", key, value, err)
			return err
		}
		return nil
	}); err != nil {
		w.store.logger.Printf("[ERR] boltdb: Failed to update data: %s: %v: %v", key, value, err)
		return err
	}

	w.store.logger.Printf("[DEBUG] boltdb: Data has been put: %s: %v", key, value)
	return nil
}

func (w *Writer) Delete(key []byte) error {
	start := time.Now()
	defer Metrics(start, "Writer", "Delete")

	var err error

	if err = w.store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(Bucket))
		err := bucket.Delete(key)
		if err != nil {
			w.store.logger.Printf("[ERR] boltdb: Failed to delete data: %s: %v", key, err)
			return err
		}
		return nil
	}); err != nil {
		w.store.logger.Printf("[ERR] boltdb: Failed to update data: %s, %v", key, err)
		return err
	}

	w.store.logger.Printf("[DEBUG] boltdb: Data has been deleted: %s", key)
	return nil
}

func (w *Writer) Close() error {
	return nil
}
