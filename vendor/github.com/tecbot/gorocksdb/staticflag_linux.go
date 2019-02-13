// +build static

package gorocksdb

// #cgo LDFLAGS: -l:librocksdb.a -l:libstdc++.a -l:libz.a -l:libbz2.a -l:libsnappy.a -lm
import "C"
