Bleve Badger Backend
=====================
> [Blevesearch](https://github.com/blevesearch/bleve) kvstore implementation based on [Badger](https://github.com/dgraph-io/badger) forked from [https://github.com/akhenakh/bleve/tree/badger](https://github.com/akhenakh/bleve/tree/badger) with alot of improvements and fixes.

Usage
==========
> `➜ go get github.com/alash3al/bbadger` .

```go
package main

import (
	"fmt"

	"github.com/alash3al/bbadger"
	"github.com/blevesearch/bleve"
)

func main() {
	// create/open bleveIndex
	index, err := bbadger.BleveIndex("/tmp/badger/indexName", bleve.NewIndexMapping())

    // index some data
    err = index.Index(identifier, your_data)

    // search for some text
    query := bleve.NewMatchQuery("text")
    search := bleve.NewSearchRequest(query)
    searchResults, err := index.Search(search)
}

```
