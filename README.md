<!--
 Copyright (c) 2019 Minoru Osuka

 Licensed to the Apache Software Foundation (ASF) under one or more
 contributor license agreements.  See the NOTICE file distributed with
 this work for additional information regarding copyright ownership.
 The ASF licenses this file to You under the Apache License, Version 2.0
 (the "License"); you may not use this file except in compliance with
 the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# Blast

Blast is a full-text search and indexing server written in [Go](https://golang.org) built on top of [Bleve](http://www.blevesearch.com).  
It provides functions through [gRPC](http://www.grpc.io) ([HTTP/2](https://en.wikipedia.org/wiki/HTTP/2) + [Protocol Buffers](https://developers.google.com/protocol-buffers/)) or traditional [RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer) API ([HTTP/1.1](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) + [JSON](http://www.json.org)).  
Blast implements a [Raft consensus algorithm](https://raft.github.io/) by [hashicorp/raft](https://github.com/hashicorp/raft). It achieves consensus across all the nodes, ensuring that every change made to the system is made to a quorum of nodes, or none at all.
Blast makes it easy for programmers to develop search applications with advanced features.


## Features

- Full-text search/indexing
- Faceted search
- Spatial/Geospatial search
- Search result highlighting
- Distributed search/indexing
- Index replication
- Bringing up cluster
- Cluster Federation
- An easy-to-use HTTP API
- CLI is available
- Docker container image is available


## Installing dependencies

Blast requires some C/C++ libraries if you need to enable cld2, icu, libstemmer or leveldb. The following sections are instructions for satisfying dependencies on particular platforms.

### Ubuntu 18.10

```bash
$ sudo apt-get update
$ sudo apt-get install -y \
    libicu-dev \
    libstemmer-dev \
    libleveldb-dev \
    gcc-4.8 \
    g++-4.8 \
    build-essential

$ sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 80
$ sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-8 80
$ sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 90
$ sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-4.8 90

$ export GOPATH=${HOME}/go
$ mkdir -p ${GOPATH}/src/github.com/blevesearch
$ cd ${GOPATH}/src/github.com/blevesearch
$ git clone https://github.com/blevesearch/cld2.git
$ cd ${GOPATH}/src/github.com/blevesearch/cld2
$ git clone https://github.com/CLD2Owners/cld2.git
$ cd cld2/internal
$ ./compile_libs.sh
$ sudo cp *.so /usr/local/lib
```

### macOS High Sierra Version 10.13.6

```bash
$ brew install \
    icu4c \
    leveldb

$ export GOPATH=${HOME}/go
$ go get -u -v github.com/blevesearch/cld2
$ cd ${GOPATH}/src/github.com/blevesearch/cld2
$ git clone https://github.com/CLD2Owners/cld2.git
$ cd cld2/internal
$ perl -p -i -e 's/soname=/install_name,/' compile_libs.sh
$ ./compile_libs.sh
$ sudo cp *.so /usr/local/lib
```


## Building Blast

When you satisfied dependencies, let's build Blast for Linux as following:

```bash
$ mkdir -p ${GOPATH}/src/github.com/mosuka
$ cd ${GOPATH}/src/github.com/mosuka
$ git clone https://github.com/mosuka/blast.git
$ cd blast
$ make build
```

If you want to build for other platform, set `GOOS`, `GOARCH` environment variables. For example, build for macOS like following:

```bash
$ make \
    GOOS=darwin \
    build
```

Blast supports some [Bleve Extensions (blevex)](https://github.com/blevesearch/blevex). If you want to build with them, please set `CGO_LDFLAGS`, `CGO_CFLAGS`, `CGO_ENABLED` and `BUILD_TAGS`. For example, build LevelDB to be available for index storage as follows:

```bash
$ make \
    GOOS=linux \
    BUILD_TAGS=leveldb \
    CGO_ENABLED=1 \
    build
```

You can enable all the Bleve extensions supported by Blast as follows:

###  Linux

```bash
$ make \
    GOOS=linux \
    BUILD_TAGS="kagome icu libstemmer cld2 cznicb leveldb badger" \
    CGO_ENABLED=1 \
    build
```

### macOS

```bash
$ make \
    GOOS=darwin \
    BUILD_TAGS="kagome icu libstemmer cld2 cznicb leveldb badger" \
    CGO_ENABLED=1 \
    CGO_LDFLAGS="-L/usr/local/opt/icu4c/lib" \
    CGO_CFLAGS="-I/usr/local/opt/icu4c/include" \
    build
```

### Build flags

Please refer to the following table for details of Bleve Extensions:

| BUILD_TAGS | CGO_ENABLED | Description |
| ---------- | ----------- | ----------- |
| cld2       | 1           | Enable Compact Language Detector |
| kagome     | 0           | Enable Japanese Language Analyser |
| icu        | 1           | Enable ICU Tokenizer, Thai Language Analyser |
| libstemmer | 1           | Enable Language Stemmer (Danish, German, English, Spanish, Finnish, French, Hungarian, Italian, Dutch, Norwegian, Portuguese, Romanian, Russian, Swedish, Turkish) |
| cznicb     | 0           | Enable cznicb KV store |
| leveldb    | 1           | Enable LevelDB |
| badger     | 0           | Enable Badger (This feature is considered experimental) |

If you want to enable the feature whose `CGO_ENABLE` is `1`, please install it referring to the Installing dependencies section above.

### Binaries

You can see the binary file when build successful like so:

```bash
$ ls ./bin
blast
```


## Testing Blast

If you want to test your changes, run command like following:

```bash
$ make \
    test
```

You can test with all the Bleve extensions supported by Blast as follows:

###  Linux

```bash
$ make \
    GOOS=linux \
    BUILD_TAGS="kagome icu libstemmer cld2 cznicb leveldb badger" \
    CGO_ENABLED=1 \
    test
```

### macOS

```bash
$ make \
    GOOS=darwin \
    BUILD_TAGS="kagome icu libstemmer cld2 cznicb leveldb badger" \
    CGO_ENABLED=1 \
    CGO_LDFLAGS="-L/usr/local/opt/icu4c/lib" \
    CGO_CFLAGS="-I/usr/local/opt/icu4c/include" \
    test
```


## Packaging Blast

###  Linux

```bash
$ make \
    GOOS=linux \
    BUILD_TAGS="kagome icu libstemmer cld2 cznicb leveldb badger" \
    CGO_ENABLED=1 \
    dist
```

### macOS

```bash
$ make \
    GOOS=darwin \
    BUILD_TAGS="kagome icu libstemmer cld2 cznicb leveldb badger" \
    CGO_ENABLED=1 \
    CGO_LDFLAGS="-L/usr/local/opt/icu4c/lib" \
    CGO_CFLAGS="-I/usr/local/opt/icu4c/include" \
    dist
```


## Starting Blast in standalone mode

![standalone](https://user-images.githubusercontent.com/970948/59768879-138f5180-92e0-11e9-8b33-c7b1a93e0893.png)

Running a Blast in standalone mode is easy. Start a indexer like so:

```bash
$ ./bin/blast indexer start \
    --grpc-address=:5000 \
    --grpc-gateway-address=:6000 \
    --http-address=:8000 \
    --node-id=indexer1 \
    --node-address=:2000 \
    --data-dir=/tmp/blast/indexer1 \
    --raft-storage-type=boltdb \
    --index-mapping-file=./example/wiki_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb
```

Please refer to following document for details of index mapping:
- http://blevesearch.com/docs/Terminology/
- http://blevesearch.com/docs/Text-Analysis/
- http://blevesearch.com/docs/Index-Mapping/
- https://github.com/blevesearch/bleve/blob/master/mapping/index.go#L43

You can check the node with the following command:

```bash
$ ./bin/blast indexer node info --grpc-address=:5000 | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "node": {
    "id": "indexer1",
    "bind_address": ":2000",
    "state": 3,
    "metadata": {
      "grpc_address": ":5000",
      "grpc_gateway_address": ":6000",
      "http_address": ":8000"
    }
  }
}
```

You can now put, get, search and delete the documents via CLI.  

### Indexing a document via CLI

For document indexing, execute the following command:

```bash
$ ./bin/blast indexer index --grpc-address=:5000 enwiki_1 '
{
  "fields": {
    "title_en": "Search engine (computing)",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "_type": "enwiki"
  }
}
' | jq .
```

or

```bash
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/wiki_doc_enwiki_1.json | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{}
```

### Getting a document via CLI

Getting a document is as following:

```bash
$ ./bin/blast indexer get --grpc-address=:5000 enwiki_1 | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "fields": {
    "_type": "enwiki",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  }
}
```

### Searching documents via CLI

Searching documents is as like following:

```bash
$ ./bin/blast indexer search --grpc-address=:5000 --file=./example/wiki_search_request.json | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "search_result": {
    "status": {
      "total": 1,
      "failed": 0,
      "successful": 1
    },
    "request": {
      "query": {
        "query": "+_all:search"
      },
      "size": 10,
      "from": 0,
      "highlight": {
        "style": "html",
        "fields": [
          "title",
          "text"
        ]
      },
      "fields": [
        "*"
      ],
      "facets": {
        "Timestamp range": {
          "size": 10,
          "field": "timestamp",
          "date_ranges": [
            {
              "end": "2010-12-31T23:59:59Z",
              "name": "2001 - 2010",
              "start": "2001-01-01T00:00:00Z"
            },
            {
              "end": "2020-12-31T23:59:59Z",
              "name": "2011 - 2020",
              "start": "2011-01-01T00:00:00Z"
            }
          ]
        },
        "Type count": {
          "size": 10,
          "field": "_type"
        }
      },
      "explain": false,
      "sort": [
        "-_score",
        "_id",
        "-timestamp"
      ],
      "includeLocations": false
    },
    "hits": [
      {
        "index": "/tmp/blast/indexer1/index",
        "id": "enwiki_1",
        "score": 0.09703538256409851,
        "locations": {
          "text_en": {
            "search": [
              {
                "pos": 2,
                "start": 2,
                "end": 8,
                "array_positions": null
              },
              {
                "pos": 20,
                "start": 118,
                "end": 124,
                "array_positions": null
              },
              {
                "pos": 33,
                "start": 195,
                "end": 201,
                "array_positions": null
              },
              {
                "pos": 68,
                "start": 415,
                "end": 421,
                "array_positions": null
              },
              {
                "pos": 73,
                "start": 438,
                "end": 444,
                "array_positions": null
              },
              {
                "pos": 76,
                "start": 458,
                "end": 466,
                "array_positions": null
              }
            ]
          },
          "title_en": {
            "search": [
              {
                "pos": 1,
                "start": 0,
                "end": 6,
                "array_positions": null
              }
            ]
          }
        },
        "sort": [
          "_score",
          "enwiki_1",
          " \u0001\u0015\u001f\u0004~80Pp\u0000"
        ],
        "fields": {
          "_type": "enwiki",
          "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_en": "Search engine (computing)"
        }
      }
    ],
    "total_hits": 1,
    "max_score": 0.09703538256409851,
    "took": 122105,
    "facets": {
      "Timestamp range": {
        "field": "timestamp",
        "total": 1,
        "missing": 0,
        "other": 0,
        "date_ranges": [
          {
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z",
            "end": "2020-12-31T23:59:59Z",
            "count": 1
          }
        ]
      },
      "Type count": {
        "field": "_type",
        "total": 1,
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "term": "enwiki",
            "count": 1
          }
        ]
      }
    }
  }
}
```

Please refer to following document for details of search request and result:
- http://blevesearch.com/docs/Query/
- http://blevesearch.com/docs/Query-String-Query/
- http://blevesearch.com/docs/Sorting/
- https://github.com/blevesearch/bleve/blob/master/search.go#L267
- https://github.com/blevesearch/bleve/blob/master/search.go#L443

### Deleting a document via CLI

Deleting a document is as following:

```bash
$ ./bin/blast indexer delete --grpc-address=:5000 enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```json
{}
```

### Indexing documents in bulk via CLI

Indexing documents in bulk, run the following command:

```bash
$ ./bin/blast indexer index --grpc-address=:5000 --file=./example/wiki_bulk_index.jsonl --bulk | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "count": 36
}
```

### Deleting documents in bulk via CLI

Deleting documents in bulk, run the following command:

```bash
$ ./bin/blast indexer delete --grpc-address=:5000 --file=./example/wiki_bulk_delete.txt | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "count": 36
}
```


## Using HTTP REST API

Also you can do above commands via HTTP REST API that listened port 5002.

### Indexing a document via HTTP REST API

Indexing a document via HTTP is as following:

```bash
$ curl -X PUT 'http://127.0.0.1:6000/v1/documents/enwiki_1' -H 'Content-Type: application/json' --data-binary '
{
  "fields": {
    "title_en": "Search engine (computing)",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "_type": "enwiki"
  }
}
' | jq .
```

or

```bash
$ curl -X PUT 'http://127.0.0.1:6000/v1/documents' -H 'Content-Type: application/json' --data-binary @./example/wiki_doc_enwiki_1.json | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{}
```

### Getting a document via HTTP REST API

Getting a document via HTTP is as following:

```bash
$ curl -X GET 'http://127.0.0.1:6000/v1/documents/enwiki_1' -H 'Content-Type: application/json' | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "fields": {
    "_type": "enwiki",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  }
}
```

### Searching documents via HTTP REST API

Searching documents via HTTP is as following:

```bash
$ curl -X POST 'http://127.0.0.1:6000/v1/search' -H 'Content-Type: application/json' --data-binary @./example/wiki_search_request.json | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "search_result": {
    "status": {
      "total": 1,
      "failed": 0,
      "successful": 1
    },
    "request": {
      "query": {
        "query": "+_all:search"
      },
      "size": 10,
      "from": 0,
      "highlight": {
        "style": "html",
        "fields": [
          "title",
          "text"
        ]
      },
      "fields": [
        "*"
      ],
      "facets": {
        "Timestamp range": {
          "size": 10,
          "field": "timestamp",
          "date_ranges": [
            {
              "end": "2010-12-31T23:59:59Z",
              "name": "2001 - 2010",
              "start": "2001-01-01T00:00:00Z"
            },
            {
              "end": "2020-12-31T23:59:59Z",
              "name": "2011 - 2020",
              "start": "2011-01-01T00:00:00Z"
            }
          ]
        },
        "Type count": {
          "size": 10,
          "field": "_type"
        }
      },
      "explain": false,
      "sort": [
        "-_score",
        "_id",
        "-timestamp"
      ],
      "includeLocations": false
    },
    "hits": [
      {
        "index": "/tmp/blast/indexer1/index",
        "id": "enwiki_1",
        "score": 0.09703538256409851,
        "locations": {
          "text_en": {
            "search": [
              {
                "pos": 2,
                "start": 2,
                "end": 8,
                "array_positions": null
              },
              {
                "pos": 20,
                "start": 118,
                "end": 124,
                "array_positions": null
              },
              {
                "pos": 33,
                "start": 195,
                "end": 201,
                "array_positions": null
              },
              {
                "pos": 68,
                "start": 415,
                "end": 421,
                "array_positions": null
              },
              {
                "pos": 73,
                "start": 438,
                "end": 444,
                "array_positions": null
              },
              {
                "pos": 76,
                "start": 458,
                "end": 466,
                "array_positions": null
              }
            ]
          },
          "title_en": {
            "search": [
              {
                "pos": 1,
                "start": 0,
                "end": 6,
                "array_positions": null
              }
            ]
          }
        },
        "sort": [
          "_score",
          "enwiki_1",
          " \u0001\u0015\u001f\u0004~80Pp\u0000"
        ],
        "fields": {
          "_type": "enwiki",
          "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_en": "Search engine (computing)"
        }
      }
    ],
    "total_hits": 1,
    "max_score": 0.09703538256409851,
    "took": 323568,
    "facets": {
      "Timestamp range": {
        "field": "timestamp",
        "total": 1,
        "missing": 0,
        "other": 0,
        "date_ranges": [
          {
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z",
            "end": "2020-12-31T23:59:59Z",
            "count": 1
          }
        ]
      },
      "Type count": {
        "field": "_type",
        "total": 1,
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "term": "enwiki",
            "count": 1
          }
        ]
      }
    }
  }
}
```

### Deleting a document via HTTP REST API

Deleting a document via HTTP is as following:

```bash
$ curl -X DELETE 'http://127.0.0.1:6000/v1/documents/enwiki_1' -H 'Content-Type: application/json' | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{}
```

### Indexing documents in bulk via HTTP REST API

Indexing documents in bulk via HTTP is as following:

```bash
$ curl -X PUT 'http://127.0.0.1:6000/v1/bulk' -H 'Content-Type: application/x-ndjson' --data-binary @./example/wiki_bulk_index.jsonl | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "count": 36
}
```

### Deleting documents in bulk via HTTP REST API

Deleting documents in bulk via HTTP is as following:

```bash
$ curl -X DELETE 'http://127.0.0.1:6000/v1/bulk' -H 'Content-Type: text/plain' --data-binary @./example/wiki_bulk_delete.txt | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "count": 36
}
```


## Starting Blast in cluster mode

![cluster](https://user-images.githubusercontent.com/970948/59768677-bf846d00-92df-11e9-8a70-92496ff55ce7.png)

Blast can easily bring up a cluster. Running a Blast in standalone is not fault tolerant. If you need to improve fault tolerance, start two more indexers as follows:

First of all, start a indexer in standalone.

```bash
$ ./bin/blast indexer start \
    --grpc-address=:5000 \
    --grpc-gateway-address=:6000 \
    --http-address=:8000 \
    --node-id=indexer1 \
    --node-address=:2000 \
    --data-dir=/tmp/blast/indexer1 \
    --raft-storage-type=boltdb \
    --index-mapping-file=./example/wiki_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb
```

Then, start two more indexers.

```bash
$ ./bin/blast indexer start \
    --peer-grpc-address=:5000 \
    --grpc-address=:5010 \
    --grpc-gateway-address=:6010 \
    --http-address=:8010 \
    --node-id=indexer2 \
    --node-address=:2010 \
    --data-dir=/tmp/blast/indexer2 \
    --raft-storage-type=boltdb

$ ./bin/blast indexer start \
    --peer-grpc-address=:5000 \
    --grpc-address=:5020 \
    --grpc-gateway-address=:6020 \
    --http-address=:8020 \
    --node-id=indexer3 \
    --node-address=:2020 \
    --data-dir=/tmp/blast/indexer3 \
    --raft-storage-type=boltdb
```

_Above example shows each Blast node running on the same host, so each node must listen on different ports. This would not be necessary if each node ran on a different host._

This instructs each new node to join an existing node, specifying `--peer-addr=:5001`. Each node recognizes the joining clusters when started.
So you have a 3-node cluster. That way you can tolerate the failure of 1 node. You can check the peers in the cluster with the following command:

```bash
$ ./bin/blast indexer cluster info --grpc-address=:5000 | jq .
```

or

```bash
$ curl -X GET 'http://127.0.0.1:6000/v1/cluster/status' -H 'Content-Type: application/json' | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "cluster": {
    "nodes": {
      "indexer1": {
        "id": "indexer1",
        "bind_address": ":2000",
        "state": 1,
        "metadata": {
          "grpc_address": ":5000",
          "grpc_gateway_address": ":6000",
          "http_address": ":8000"
        }
      },
      "indexer2": {
        "id": "indexer2",
        "bind_address": ":2010",
        "state": 1,
        "metadata": {
          "grpc_address": ":5010",
          "grpc_gateway_address": ":6010",
          "http_address": ":8010"
        }
      },
      "indexer3": {
        "id": "indexer3",
        "bind_address": ":2020",
        "state": 3,
        "metadata": {
          "grpc_address": ":5020",
          "grpc_gateway_address": ":6020",
          "http_address": ":8020"
        }
      }
    }
  }
}
```

Recommend 3 or more odd number of nodes in the cluster. In failure scenarios, data loss is inevitable, so avoid deploying single nodes.

The following command indexes documents to any node in the cluster:

```bash
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/wiki_doc_enwiki_1.json | jq .
```

So, you can get the document from the node specified by the above command as follows:

```bash
$ ./bin/blast indexer get --grpc-address=:5000 enwiki_1 | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "fields": {
    "_type": "enwiki",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  }
}
```

You can also get the same document from other nodes in the cluster as follows:

```bash
$ ./bin/blast indexer get --grpc-address=:5010 enwiki_1 | jq .
$ ./bin/blast indexer get --grpc-address=:5020 enwiki_1 | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "fields": {
    "_type": "enwiki",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  }
}
```


## Starting Blast in federated mode (experimental)

![federation](https://user-images.githubusercontent.com/970948/59768498-6f0d0f80-92df-11e9-8538-2a1c6e44c30a.png)

Running a Blast in cluster mode allows you to replicate the index among indexers in a cluster to improve fault tolerance.  
However, as the index grows, performance degradation can become an issue. Therefore, instead of providing a large single physical index, it is better to distribute indices across multiple indexers.  
Blast provides a federated mode to enable distributed search and indexing.

Blast provides the following type of node for federation:
- manager: Manager manage common index mappings to index across multiple indexers. It also manages information and status of clusters that participate in the federation.
- dispatcher: Dispatcher is responsible for distributed search or indexing of each indexer. In the case of a index request, send document to each cluster based on the document ID. And in the case of a search request, the same query is sent to each cluster, then the search results are merged and returned to the client.

### Bring up the manager cluster

Manager can also bring up a cluster like an indexer. Specify a common index mapping for federation at startup.

```bash
$ ./bin/blast manager start \
    --grpc-address=:5100 \
    --grpc-gateway-address=:6100 \
    --http-address=:8100 \
    --node-id=manager1 \
    --node-address=:2100 \
    --data-dir=/tmp/blast/manager1 \
    --raft-storage-type=boltdb \
    --index-mapping-file=./example/wiki_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb

$ ./bin/blast manager start \
    --peer-grpc-address=:5100 \
    --grpc-address=:5110 \
    --grpc-gateway-address=:6110 \
    --http-address=:8110 \
    --node-id=manager2 \
    --node-address=:2110 \
    --data-dir=/tmp/blast/manager2 \
    --raft-storage-type=boltdb

$ ./bin/blast manager start \
    --peer-grpc-address=:5100 \
    --grpc-address=:5120 \
    --grpc-gateway-address=:6120 \
    --http-address=:8120 \
    --node-id=manager3 \
    --node-address=:2120 \
    --data-dir=/tmp/blast/manager3 \
    --raft-storage-type=boltdb
```

### Bring up the indexer cluster

Federated mode differs from cluster mode that it specifies the manager in start up to bring up indexer cluster.  
The following example starts two 3-node clusters.

```bash
$ ./bin/blast indexer start \
    --manager-grpc-address=:5100 \
    --shard-id=shard1 \
    --grpc-address=:5000 \
    --grpc-gateway-address=:6000 \
    --http-address=:8000 \
    --node-id=indexer1 \
    --node-address=:2000 \
    --data-dir=/tmp/blast/indexer1 \
    --raft-storage-type=boltdb

$ ./bin/blast indexer start \
    --manager-grpc-address=:5100 \
    --shard-id=shard1 \
    --grpc-address=:5010 \
    --grpc-gateway-address=:6010 \
    --http-address=:8010 \
    --node-id=indexer2 \
    --node-address=:2010 \
    --data-dir=/tmp/blast/indexer2 \
    --raft-storage-type=boltdb

$ ./bin/blast indexer start \
    --manager-grpc-address=:5100 \
    --shard-id=shard1 \
    --grpc-address=:5020 \
    --grpc-gateway-address=:6020 \
    --http-address=:8020 \
    --node-id=indexer3 \
    --node-address=:2020 \
    --data-dir=/tmp/blast/indexer3 \
    --raft-storage-type=boltdb

$ ./bin/blast indexer start \
    --manager-grpc-address=:5100 \
    --shard-id=shard2 \
    --grpc-address=:5030 \
    --grpc-gateway-address=:6030 \
    --http-address=:8030 \
    --node-id=indexer4 \
    --node-address=:2030 \
    --data-dir=/tmp/blast/indexer4 \
    --raft-storage-type=boltdb

$ ./bin/blast indexer start \
    --manager-grpc-address=:5100 \
    --shard-id=shard2 \
    --grpc-address=:5040 \
    --grpc-gateway-address=:6040 \
    --http-address=:8040 \
    --node-id=indexer5 \
    --node-address=:2040 \
    --data-dir=/tmp/blast/indexer5 \
    --raft-storage-type=boltdb

$ ./bin/blast indexer start \
    --manager-grpc-address=:5100 \
    --shard-id=shard2 \
    --grpc-address=:5050 \
    --grpc-gateway-address=:6050 \
    --http-address=:8050 \
    --node-id=indexer6 \
    --node-address=:2050 \
    --data-dir=/tmp/blast/indexer6 \
    --raft-storage-type=boltdb
```

### Start up the dispatcher

Finally, start the dispatcher with a manager that manages the target federation so that it can perform distributed search and indexing.

```bash
$ ./bin/blast dispatcher start \
    --manager-grpc-address=:5100 \
    --grpc-address=:5200 \
    --grpc-gateway-address=:6200 \
    --http-address=:8200
```

### Check the cluster info

```bash
$ ./bin/blast manager cluster info --grpc-address=:5100 | jq .
$ ./bin/blast indexer cluster info --grpc-address=:5000 | jq .
$ ./bin/blast indexer cluster info --grpc-address=:5030 | jq .
$ ./bin/blast manager get cluster --grpc-address=:5100 --format=json | jq .
```

```bash
$ ./bin/blast dispatcher index --grpc-address=:5200 --file=./example/wiki_bulk_index.jsonl --bulk | jq .
```

```bash
$ ./bin/blast dispatcher search --grpc-address=:5200 --file=./example/wiki_search_request_simple.json | jq .
```

```bash
$ ./bin/blast dispatcher delete --grpc-address=:5200 --file=./example/wiki_bulk_delete.txt | jq .
```


## Blast on Docker

### Building Docker container image on localhost

You can build the Docker container image like so:

```bash
$ make docker-build
```

### Pulling Docker container image from docker.io

You can also use the Docker container image already registered in docker.io like so:

```bash
$ docker pull mosuka/blast:latest
```

See https://hub.docker.com/r/mosuka/blast/tags/

### Pulling Docker container image from docker.io

You can also use the Docker container image already registered in docker.io like so:

```bash
$ docker pull mosuka/blast:latest
```

### Running Indexer on Docker

Running a Blast data node on Docker. Start Blast data node like so:

```bash
$ docker run --rm --name blast-indexer1 \
    -p 2000:2000 \
    -p 5000:5000 \
    -p 6000:6000 \
    -p 8000:8000 \
    -v $(pwd)/example:/opt/blast/example \
    mosuka/blast:latest blast indexer start \
      --grpc-address=:5000 \
      --grpc-gateway-address=:6000 \
      --http-address=:8000 \
      --node-id=blast-indexer1 \
      --node-address=:2000 \
      --data-dir=/tmp/blast/indexer1 \
      --raft-storage-type=boltdb \
      --index-mapping-file=/opt/blast/example/wiki_index_mapping.json \
      --index-type=upside_down \
      --index-storage-type=boltdb
```

You can execute the command in docker container as follows:

```bash
$ docker exec -it blast-indexer1 blast indexer node info --grpc-address=:5000
```

### Running cluster on Docker compose

Also, running a Blast cluster on Docker compose. 

```bash
$ docker-compose up -d manager1
$ docker-compose up -d indexer1
$ docker-compose up -d indexer2
$ docker-compose up -d indexer3
$ docker-compose up -d indexer4
$ docker-compose up -d indexer5
$ docker-compose up -d indexer6
$ docker-compose up -d dispatcher1
$ docker-compose ps
$ ./bin/blast manager get --grpc-address=127.0.0.1:5110 /cluster | jq .
```

```bash
$ docker-compose down
```


## Wikipedia example

This section explain how to index Wikipedia dump to Blast.

### Install wikiextractor

```bash
$ cd ${HOME}
$ git clone git@github.com:attardi/wikiextractor.git
```

### Download wikipedia dump

```bash
$ curl -o ~/tmp/enwiki-20190101-pages-articles.xml.bz2 https://dumps.wikimedia.org/enwiki/20190101/enwiki-20190101-pages-articles.xml.bz2
```

### Parsing wikipedia dump

```bash
$ cd wikiextractor
$ ./WikiExtractor.py -o ~/tmp/enwiki --json ~/tmp/enwiki-20190101-pages-articles.xml.bz2
```

### Starting Indexer

```bash
$ ./bin/blast indexer start \
    --grpc-address=:5000 \
    --grpc-gateway-address=:6000 \
    --http-address=:8000 \
    --node-id=indexer1 \
    --node-address=:2000 \
    --data-dir=/tmp/blast/indexer1 \
    --raft-storage-type=boltdb \
    --index-mapping-file=./example/enwiki_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb
```

### Indexing wikipedia dump

```bash
$ for FILE in $(find ~/tmp/enwiki -type f -name '*' | sort)
  do
    echo "Indexing ${FILE}"
    TIMESTAMP=$(date -u "+%Y-%m-%dT%H:%M:%SZ")
    DOCS=$(cat ${FILE} | jq -r '. + {fields: {url: .url, title_en: .title, text_en: .text, timestamp: "'${TIMESTAMP}'", _type: "enwiki"}} | del(.url) | del(.title) | del(.text) | del(.fields.id)' | jq -c)
    curl -s -X PUT -H 'Content-Type: application/x-ndjson' "http://127.0.0.1:6000/v1/bulk" --data-binary "${DOCS}"
    echo ""
  done
```


## Spatial/Geospatial search example

This section explain how to index Spatial/Geospatial data to Blast.

### Starting Indexer with Spatial/Geospatial index mapping

```bash
$ ./bin/blast indexer start \
    --grpc-address=:5000 \
    --http-address=:8000 \
    --node-id=indexer1 \
    --node-address=:2000 \
    --data-dir=/tmp/blast/indexer1 \
    --raft-storage-type=boltdb \
    --index-mapping-file=./example/geo_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb
```

### Indexing example Spatial/Geospatial data

```bash
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/geo_doc_1.json
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/geo_doc_2.json
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/geo_doc_3.json
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/geo_doc_4.json
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/geo_doc_5.json
$ ./bin/blast indexer index --grpc-address=:5000 --file ./example/geo_doc_6.json
```

### Searching example Spatial/Geospatial data

```bash
$ ./bin/blast indexer search --grpc-address=:5000 --file=./example/geo_search_request.json
```
