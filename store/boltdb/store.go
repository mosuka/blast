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

package boltdb

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/mosuka/blast/logging"
)

const (
	Filename = "boltdb.db"
	Bucket   = "DB"

	DefaultDir  = "./data/store"
	DefaultMode = os.FileMode(0600)
)

type StoreConfig struct {
	Dir     string        `json:"dir,omitempty"`
	Mode    os.FileMode   `json:"mode,omitempty"`
	Options *bolt.Options `json:"options,omitempty"`
}

func DefaultStoreConfig() *StoreConfig {
	return &StoreConfig{
		Dir:     DefaultDir,
		Mode:    DefaultMode,
		Options: bolt.DefaultOptions,
	}
}

type Store struct {
	db     *bolt.DB
	logger *log.Logger
}

func NewStore(config *StoreConfig) (*Store, error) {
	var err error

	// Create directory
	if err := os.MkdirAll(config.Dir, 0755); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("store path not accessible: %v", err)
	}

	// Open boltdb
	var db *bolt.DB
	if db, err = bolt.Open(filepath.Join(config.Dir, Filename), config.Mode, config.Options); err != nil {
		return nil, err
	}

	// Create bucket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(Bucket))
		if err != nil {
			return err
		}
		return nil
	})

	return &Store{
		db:     db,
		logger: logging.DefaultLogger(),
	}, nil
}

func (s *Store) SetLogger(logger *log.Logger) {
	s.logger = logger
	return
}

func (s *Store) Reader() (*Reader, error) {
	var err error

	var reader *Reader
	if reader, err = NewReader(s); err != nil {
		s.logger.Printf("[ERR] boltdb: Failed to create reader: %v", err)
		return nil, err
	}

	return reader, nil
}

func (s *Store) Writer() (*Writer, error) {
	var err error

	var writer *Writer
	if writer, err = NewWriter(s); err != nil {
		s.logger.Printf("[ERR] boltdb: Failed to create writer: %v", err)
		return nil, err
	}

	return writer, nil
}

func (s *Store) Iterator() (*Iterator, error) {
	var err error

	var iterator *Iterator
	if iterator, err = NewIterator(s); err != nil {
		s.logger.Printf("[ERR] boltdb: Failed to create iterator: %v", err)
		return nil, err
	}

	return iterator, nil
}

func (s *Store) Bulker(batchSize int) (*Bulker, error) {
	var err error

	var bulker *Bulker
	if bulker, err = NewBulker(s, batchSize); err != nil {
		s.logger.Printf("[ERR] boltdb: Failed to create bulker: %v", err)
		return nil, err
	}

	return bulker, nil
}

func (s *Store) Close() error {
	var err error

	if err = s.db.Close(); err != nil {
		s.logger.Printf("[ERR] boltdb: Failed to close store: %v", err)
		return err
	}

	return nil
}
