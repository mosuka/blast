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

import "github.com/boltdb/bolt"

type Iterator struct {
	store     *Store
	tx        *bolt.Tx
	cursor    *bolt.Cursor
	key       []byte
	value     []byte
	firstTime bool
}

func NewIterator(s *Store) (*Iterator, error) {
	var err error

	var tx *bolt.Tx
	if tx, err = s.db.Begin(false); err != nil {
		return nil, err
	}

	return &Iterator{
		store:     s,
		tx:        tx,
		cursor:    tx.Bucket([]byte(Bucket)).Cursor(),
		firstTime: true,
	}, nil
}

func (i *Iterator) Next() bool {
	if i.firstTime {
		i.key, i.value = i.cursor.First()
		i.firstTime = false
	} else {
		i.key, i.value = i.cursor.Next()
	}

	if i.key != nil {
		return true
	}

	return false
}

func (i *Iterator) Key() []byte {
	return i.key
}

func (i *Iterator) Value() []byte {
	return i.value
}

func (i *Iterator) Close() error {
	defer i.tx.Rollback()

	return nil
}
