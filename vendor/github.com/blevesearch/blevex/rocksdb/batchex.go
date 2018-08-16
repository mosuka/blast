//  Copyright (c) 2016 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package rocksdb

/*
#include <stdio.h>
#include <stdlib.h>
#include "rocksdb/c.h"

char *blevex_rocksdb_execute_direct_batch(
    rocksdb_t* db,
    const unsigned char writeoptions_sync,
    const unsigned char writeoptions_disable_WAL,
    const int num_sets,
    const char* const* set_keys,
    const size_t* set_keys_sizes,
    const char* const* set_vals,
    const size_t* set_vals_sizes,
    int num_deletes,
    const char* const* delete_keys,
    const size_t* delete_keys_sizes,
    int num_merges,
    const char* const* merge_keys,
    const size_t* merge_keys_sizes,
    const char* const* merge_vals,
    const size_t* merge_vals_sizes) {
    rocksdb_writebatch_t* b = rocksdb_writebatch_create();

    if (num_sets > 0) {
        rocksdb_writebatch_putv(b,
            num_sets, set_keys, set_keys_sizes,
            num_sets, set_vals, set_vals_sizes);
    }
    if (num_deletes > 0) {
        rocksdb_writebatch_deletev(b,
            num_deletes, delete_keys, delete_keys_sizes);
    }
    if (num_merges > 0) {
        rocksdb_writebatch_mergev(b,
            num_merges, merge_keys, merge_keys_sizes,
            num_merges, merge_vals, merge_vals_sizes);
    }

    char *errMsg = NULL;

    rocksdb_writeoptions_t *options = rocksdb_writeoptions_create();

    rocksdb_writeoptions_set_sync(options, writeoptions_sync);
    rocksdb_writeoptions_disable_WAL(options, writeoptions_disable_WAL);

    rocksdb_write(db, options, b, &errMsg);

    rocksdb_writeoptions_destroy(options);

    rocksdb_writebatch_destroy(b);

    return errMsg;
}

void blevex_rocksdb_alloc_direct_batch(size_t totalBytes, size_t n, void **out) {
    out[0] = malloc(totalBytes);
    out[1] = malloc(n * sizeof(char *));
    out[2] = malloc(n * sizeof(size_t));
}

void blevex_rocksdb_free_direct_batch(void **bufs) {
    free(bufs[0]);
    free(bufs[1]);
    free(bufs[2]);
}
*/
import "C"

import (
	"errors"
	"reflect"
	"unsafe"

	"github.com/blevesearch/bleve/index/store"
)

type BatchEx struct {
	cbufs []unsafe.Pointer
	buf   []byte

	num_sets       int
	set_keys       []*C.char
	set_keys_sizes []C.size_t
	set_vals       []*C.char
	set_vals_sizes []C.size_t

	num_deletes       int
	delete_keys       []*C.char
	delete_keys_sizes []C.size_t

	num_merges       int
	merge_keys       []*C.char
	merge_keys_sizes []C.size_t
	merge_vals       []*C.char
	merge_vals_sizes []C.size_t
}

func newBatchEx(o store.KVBatchOptions) *BatchEx {
	s := o.NumSets
	ss := s + o.NumSets
	ssd := ss + o.NumDeletes
	ssdm := ssd + o.NumMerges
	ssdmm := ssdm + o.NumMerges

	cbufs := make([]unsafe.Pointer, 3)

	C.blevex_rocksdb_alloc_direct_batch(C.size_t(o.TotalBytes),
		C.size_t(ssdmm), (*unsafe.Pointer)(&cbufs[0]))

	buf := unsafeToByteSlice(cbufs[0], o.TotalBytes)
	arr_ptr_char := unsafeToCPtrCharSlice(cbufs[1], ssdmm)
	arr_size_t := unsafeToCSizeTSlice(cbufs[2], ssdmm)

	return &BatchEx{
		cbufs:             cbufs,
		buf:               buf,
		set_keys:          arr_ptr_char[0:s],
		set_keys_sizes:    arr_size_t[0:s],
		set_vals:          arr_ptr_char[s:ss],
		set_vals_sizes:    arr_size_t[s:ss],
		delete_keys:       arr_ptr_char[ss:ssd],
		delete_keys_sizes: arr_size_t[ss:ssd],
		merge_keys:        arr_ptr_char[ssd:ssdm],
		merge_keys_sizes:  arr_size_t[ssd:ssdm],
		merge_vals:        arr_ptr_char[ssdm:ssdmm],
		merge_vals_sizes:  arr_size_t[ssdm:ssdmm],
	}
}

func (b *BatchEx) Set(key, val []byte) {
	b.set_keys[b.num_sets] = (*C.char)(unsafe.Pointer(&key[0]))
	b.set_keys_sizes[b.num_sets] = (C.size_t)(len(key))
	b.set_vals[b.num_sets] = (*C.char)(unsafe.Pointer(&val[0]))
	b.set_vals_sizes[b.num_sets] = (C.size_t)(len(val))
	b.num_sets += 1
}

func (b *BatchEx) Delete(key []byte) {
	b.delete_keys[b.num_deletes] = (*C.char)(unsafe.Pointer(&key[0]))
	b.delete_keys_sizes[b.num_deletes] = (C.size_t)(len(key))
	b.num_deletes += 1
}

func (b *BatchEx) Merge(key, val []byte) {
	b.merge_keys[b.num_merges] = (*C.char)(unsafe.Pointer(&key[0]))
	b.merge_keys_sizes[b.num_merges] = (C.size_t)(len(key))
	b.merge_vals[b.num_merges] = (*C.char)(unsafe.Pointer(&val[0]))
	b.merge_vals_sizes[b.num_merges] = (C.size_t)(len(val))
	b.num_merges += 1
}

func (b *BatchEx) Reset() {
	b.num_sets = 0
	b.num_deletes = 0
	b.num_merges = 0
}

func (b *BatchEx) Close() error {
	b.Reset()

	C.blevex_rocksdb_free_direct_batch((*unsafe.Pointer)(&b.cbufs[0]))

	b.cbufs = nil
	b.buf = nil
	b.set_keys = nil
	b.set_keys_sizes = nil
	b.set_vals = nil
	b.set_vals_sizes = nil
	b.delete_keys = nil
	b.delete_keys_sizes = nil
	b.merge_keys = nil
	b.merge_keys_sizes = nil
	b.merge_vals = nil
	b.merge_vals_sizes = nil

	return nil
}

func (b *BatchEx) execute(w *Writer) error {
	var num_sets C.int
	var set_keys **C.char
	var set_keys_sizes *C.size_t
	var set_vals **C.char
	var set_vals_sizes *C.size_t

	var num_deletes C.int
	var delete_keys **C.char
	var delete_keys_sizes *C.size_t

	var num_merges C.int
	var merge_keys **C.char
	var merge_keys_sizes *C.size_t
	var merge_vals **C.char
	var merge_vals_sizes *C.size_t

	if b.num_sets > 0 {
		num_sets = (C.int)(b.num_sets)
		set_keys = (**C.char)(unsafe.Pointer(&b.set_keys[0]))
		set_keys_sizes = (*C.size_t)(unsafe.Pointer(&b.set_keys_sizes[0]))
		set_vals = (**C.char)(unsafe.Pointer(&b.set_vals[0]))
		set_vals_sizes = (*C.size_t)(unsafe.Pointer(&b.set_vals_sizes[0]))
	}

	if b.num_deletes > 0 {
		num_deletes = (C.int)(b.num_deletes)
		delete_keys = (**C.char)(unsafe.Pointer(&b.delete_keys[0]))
		delete_keys_sizes = (*C.size_t)(unsafe.Pointer(&b.delete_keys_sizes[0]))
	}

	if b.num_merges > 0 {
		num_merges = (C.int)(b.num_merges)
		merge_keys = (**C.char)(unsafe.Pointer(&b.merge_keys[0]))
		merge_keys_sizes = (*C.size_t)(unsafe.Pointer(&b.merge_keys_sizes[0]))
		merge_vals = (**C.char)(unsafe.Pointer(&b.merge_vals[0]))
		merge_vals_sizes = (*C.size_t)(unsafe.Pointer(&b.merge_vals_sizes[0]))
	}

	// request fsync on write for safety by default (bleve's convention),
	// although rocksdb writeoptions normal default is false for sync.
	woptSync := C.uchar(1)
	if w.store.woptSyncUse {
		woptSync = boolToChar(w.store.woptSync)
	}

	woptDisableWAL := C.uchar(0)
	if w.store.woptDisableWALUse {
		woptDisableWAL = boolToChar(w.store.woptDisableWAL)
	}

	cErr := C.blevex_rocksdb_execute_direct_batch(
		(*C.rocksdb_t)(w.store.db.UnsafeGetDB()),
		woptSync,
		woptDisableWAL,
		num_sets,
		set_keys,
		set_keys_sizes,
		set_vals,
		set_vals_sizes,
		num_deletes,
		delete_keys,
		delete_keys_sizes,
		num_merges,
		merge_keys,
		merge_keys_sizes,
		merge_vals,
		merge_vals_sizes)
	if cErr != nil {
		err := errors.New(C.GoString(cErr))
		C.free(unsafe.Pointer(cErr))
		return err
	}

	return nil
}

// Originally from github.com/tecbot/gorocksdb/util.go.
func unsafeToByteSlice(data unsafe.Pointer, len int) []byte {
	var value []byte

	sH := (*reflect.SliceHeader)(unsafe.Pointer(&value))
	sH.Cap, sH.Len, sH.Data = len, len, uintptr(data)

	return value
}

func unsafeToCPtrCharSlice(data unsafe.Pointer, len int) []*C.char {
	var value []*C.char

	sH := (*reflect.SliceHeader)(unsafe.Pointer(&value))
	sH.Cap, sH.Len, sH.Data = len, len, uintptr(data)

	return value
}

func unsafeToCSizeTSlice(data unsafe.Pointer, len int) []C.size_t {
	var value []C.size_t

	sH := (*reflect.SliceHeader)(unsafe.Pointer(&value))
	sH.Cap, sH.Len, sH.Data = len, len, uintptr(data)

	return value
}

func boolToChar(b bool) C.uchar {
	if b {
		return 1
	}
	return 0
}
