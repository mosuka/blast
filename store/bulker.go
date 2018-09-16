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

type Bulker struct {
	store     *Store
	tx        *bolt.Tx
	batchSize int
	bucket    *bolt.Bucket
	count     int
}

func NewBulker(s *Store, b int) (*Bulker, error) {
	var err error

	var tx *bolt.Tx
	if tx, err = s.db.Begin(true); err != nil {
		return nil, err
	}

	return &Bulker{
		store:     s,
		tx:        tx,
		batchSize: b,
		bucket:    tx.Bucket([]byte(Bucket)),
		count:     0,
	}, nil
}

func (b *Bulker) Put(key []byte, value []byte) error {
	start := time.Now()
	defer Metrics(start, "Bulker", "Put")

	var err error

	if err = b.bucket.Put(key, value); err != nil {
		b.store.logger.Printf("[ERR] boltdb: Failed to put data: %s: %v: %v", key, value, err)
		return err
	}

	b.count++

	if b.count >= b.batchSize {
		if err = b.tx.Commit(); err != nil {
			b.store.logger.Printf("[ERR] boltdb: Failed to commit transaction: %v", err)
			return err
		}

		b.count = 0

		if b.tx, err = b.store.db.Begin(true); err != nil {
			b.store.logger.Printf("[ERR] boltdb: Failed to create transaction: %v", err)
			return err
		}
		b.bucket = b.tx.Bucket([]byte(Bucket))
	}

	return nil
}

func (b *Bulker) Delete(key []byte) error {
	start := time.Now()
	defer Metrics(start, "Bulker", "Delete")

	var err error

	if err = b.bucket.Delete(key); err != nil {
		b.store.logger.Printf("[ERR] boltdb: Failed to delete data: %s: %v", key, err)
		return err
	}

	b.count++

	if b.count >= b.batchSize {
		if err = b.tx.Commit(); err != nil {
			b.store.logger.Printf("[ERR] boltdb: Failed to commit transaction: %v", err)
			return err
		}

		b.count = 0

		if b.tx, err = b.store.db.Begin(true); err != nil {
			b.store.logger.Printf("[ERR] boltdb: Failed to create transaction: %v", err)
			return err
		}
		b.bucket = b.tx.Bucket([]byte(Bucket))
	}

	return nil
}

func (b *Bulker) Close() error {
	var err error

	if b.count > 0 {
		if err = b.tx.Commit(); err != nil {
			b.store.logger.Printf("[ERR] boltdb: Failed to commit transaction: %v", err)
			return err
		}
	}

	defer b.tx.Rollback()

	return nil
}
