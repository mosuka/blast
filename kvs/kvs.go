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

package kvs

import (
	"log"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/mosuka/blast/errors"
	"github.com/mosuka/blast/protobuf/kvs"
)

type KVS struct {
	dir      string
	valueDir string
	db       *badger.DB
	logger   *log.Logger
}

func NewKVS(dir string, valueDir string, logger *log.Logger) (*KVS, error) {
	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = valueDir
	opts.SyncWrites = false

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &KVS{
		dir:      dir,
		valueDir: valueDir,
		db:       db,
		logger:   logger,
	}, nil
}

func (b *KVS) Close() error {
	err := b.db.Close()
	if err != nil {
		return err
	}

	return nil
}

func (b *KVS) Get(key []byte) ([]byte, error) {
	start := time.Now()
	defer func() {
		b.logger.Printf("[DEBUG] get %v %f", key, float64(time.Since(start))/float64(time.Second))
	}()

	var value []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		v, err := item.Value()
		if err != nil {
			return err
		}
		value = append([]byte{}, v...)
		return nil
	})
	if err == badger.ErrKeyNotFound {
		b.logger.Printf("[DEBUG] %v", err)
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (b *KVS) Set(key []byte, value []byte) error {
	start := time.Now()
	defer func() {
		b.logger.Printf("[DEBUG] set %v %v %f", key, value, float64(time.Since(start))/float64(time.Second))
	}()

	err := b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *KVS) Delete(key []byte) error {
	start := time.Now()
	defer func() {
		b.logger.Printf("[DEBUG] delete %v %f", key, float64(time.Since(start))/float64(time.Second))
	}()

	err := b.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (b *KVS) SnapshotItems() <-chan *kvs.KeyValuePair {
	ch := make(chan *kvs.KeyValuePair, 1024)

	go func() {
		err := b.db.View(func(txn *badger.Txn) error {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchSize = 10
			it := txn.NewIterator(opts)
			defer it.Close()

			keyCount := 0
			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				k := item.Key()
				v, err := item.Value()
				if err != nil {
					b.logger.Printf("[WARN] %v", err)
					continue
				}

				kvp := &kvs.KeyValuePair{
					Key:   append([]byte{}, k...),
					Value: append([]byte{}, v...),
				}

				ch <- kvp
				keyCount = keyCount + 1

				if err != nil {
					return err
				}
			}

			b.logger.Print("[DEBUG] finished to write all key-value pairs to channel")
			ch <- nil

			b.logger.Printf("[INFO] snapshot total %d key-values pair", keyCount)

			return nil
		})
		if err != nil {
			b.logger.Printf("[WARN] %v", err)
		}
	}()

	return ch
}
