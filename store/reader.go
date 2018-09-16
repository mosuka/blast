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

type Reader struct {
	store *Store
}

func NewReader(s *Store) (*Reader, error) {
	return &Reader{
		store: s,
	}, nil
}

func (r *Reader) Get(key []byte) ([]byte, error) {
	start := time.Now()
	defer Metrics(start, "Reader", "Get")

	var err error

	var value []byte
	if err = r.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(Bucket))
		value = bucket.Get(key)
		return nil
	}); err != nil {
		r.store.logger.Printf("[ERR] boltdb: Failed to get data: %s: %v", key, err)
		return nil, err
	}

	if value == nil {
		r.store.logger.Printf("[DEBUG] boltdb: No such data: %s", key)
	} else {
		r.store.logger.Printf("[DEBUG] boltdb: Data has been got: %s: %v", key, value)
	}

	return value, nil
}

func (r *Reader) Close() error {
	return nil
}
