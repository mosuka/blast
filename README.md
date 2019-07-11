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


#### macOS

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
blast blastd
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


#### macOS

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


#### macOS

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
$ ./bin/blastd \
    indexer \
    --node-id=indexer1 \
    --bind-addr=:5000 \
    --grpc-addr=:5001 \
    --http-addr=:5002 \
    --data-dir=/tmp/blast/indexer1 \
    --index-mapping-file=./example/wiki_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb
```

Please refer to following document for details of index mapping:
- http://blevesearch.com/docs/Terminology/
- http://blevesearch.com/docs/Text-Analysis/
- http://blevesearch.com/docs/Index-Mapping/
- https://github.com/blevesearch/bleve/blob/master/mapping/index.go#L43

You can now put, get, search and delete the documents via CLI.  


### Indexing a document via CLI

For document indexing, execute the following command:

```bash
$ cat ./example/wiki_doc_enwiki_1.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```bash
1
```


### Getting a document via CLI

Getting a document is as following:

```bash
$ ./bin/blast get document --grpc-addr=:5001 enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "_type": "enwiki",
  "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
  "timestamp": "2018-07-04T05:41:00Z",
  "title_en": "Search engine (computing)"
}
```


### Searching documents via CLI

Searching documents is as like following:

```bash
$ cat ./example/wiki_search_request.json | xargs -0 ./bin/blast search --grpc-addr=:5001
```

You can see the result in JSON format. The result of the above command is:

```json
{
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
      "Contributor count": {
        "size": 10,
        "field": "contributor"
      },
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
      }
    },
    "explain": false,
    "sort": [
      "-_score"
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
        "_score"
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
  "took": 201951,
  "facets": {
    "Contributor count": {
      "field": "contributor",
      "total": 0,
      "missing": 1,
      "other": 0
    },
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
$ ./bin/blast delete document --grpc-addr=:5001 enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```bash
1
```


### Indexing documents in bulk via CLI

Indexing documents in bulk, run the following command:

```bash
$ cat ./example/wiki_bulk_index.json | xargs -0 ./bin/blast set document --grpc-addr=:5001
```

You can see the result in JSON format. The result of the above command is:

```bash
4
```


### Deleting documents in bulk via CLI

Deleting documents in bulk, run the following command:

```bash
$ cat ./example/wiki_bulk_delete.json | xargs -0 ./bin/blast delete document --grpc-addr=:5001
```

You can see the result in JSON format. The result of the above command is:

```bash
4
```


## Using HTTP REST API

Also you can do above commands via HTTP REST API that listened port 5002.


### Indexing a document via HTTP REST API

Indexing a document via HTTP is as following:

```bash
$ curl -X PUT 'http://127.0.0.1:5002/documents/enwiki_1' -d @./example/wiki_doc_enwiki_1.json
```


### Getting a document via HTTP REST API

Getting a document via HTTP is as following:

```bash
$ curl -X GET 'http://127.0.0.1:5002/documents/enwiki_1'
```


### Searching documents via HTTP REST API

Searching documents via HTTP is as following:

```bash
$ curl -X POST 'http://127.0.0.1:5002/search' -d @./example/wiki_search_request.json
```


### Deleting a document via HTTP REST API

Deleting a document via HTTP is as following:

```bash
$ curl -X DELETE 'http://127.0.0.1:5002/documents/enwiki_1'
```


### Indexing documents in bulk via HTTP REST API

Indexing documents in bulk via HTTP is as following:

```bash
$ curl -X PUT 'http://127.0.0.1:5002/documents' -d @./example/wiki_bulk_index.json
```


### Deleting documents in bulk via HTTP REST API

Deleting documents in bulk via HTTP is as following:

```bash
$ curl -X DELETE 'http://127.0.0.1:5002/documents' -d @./example/wiki_bulk_delete.json
```


## Starting Blast in cluster mode

![cluster](https://user-images.githubusercontent.com/970948/59768677-bf846d00-92df-11e9-8a70-92496ff55ce7.png)

Blast can easily bring up a cluster. Running a Blast in standalone is not fault tolerant. If you need to improve fault tolerance, start two more indexers as follows:

First of all, start a indexer in standalone.

```bash
$ ./bin/blastd \
    indexer \
    --node-id=indexer1 \
    --bind-addr=:5000 \
    --grpc-addr=:5001 \
    --http-addr=:5002 \
    --data-dir=/tmp/blast/indexer1 \
    --index-mapping-file=./example/wiki_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb
```

Then, start two more indexers.

```bash
$ ./bin/blastd \
    indexer \
    --peer-addr=:5001 \
    --node-id=indexer2 \
    --bind-addr=:5010 \
    --grpc-addr=:5011 \
    --http-addr=:5012 \
    --data-dir=/tmp/blast/indexer2

$ ./bin/blastd \
    indexer \
    --peer-addr=:5001 \
    --node-id=indexer3 \
    --bind-addr=:5020 \
    --grpc-addr=:5021 \
    --http-addr=:5022 \
    --data-dir=/tmp/blast/indexer3
```

_Above example shows each Blast node running on the same host, so each node must listen on different ports. This would not be necessary if each node ran on a different host._

This instructs each new node to join an existing node, specifying `--peer-addr=:5001`. Each node recognizes the joining clusters when started.
So you have a 3-node cluster. That way you can tolerate the failure of 1 node. You can check the peers in the cluster with the following command:


```bash
$ ./bin/blast get cluster --grpc-addr=:5001
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "indexer1": {
    "node_config": {
      "bind_addr": ":5000",
      "data_dir": "/tmp/blast/indexer1",
      "grpc_addr": ":5001",
      "http_addr": ":5002",
      "node_id": "indexer1",
      "raft_storage_type": "boltdb"
    },
    "state": "Leader"
  },
  "indexer2": {
    "node_config": {
      "bind_addr": ":5010",
      "data_dir": "/tmp/blast/indexer2",
      "grpc_addr": ":5011",
      "http_addr": ":5012",
      "node_id": "indexer2",
      "raft_storage_type": "boltdb"
    },
    "state": "Follower"
  },
  "indexer3": {
    "node_config": {
      "bind_addr": ":5020",
      "data_dir": "/tmp/blast/indexer3",
      "grpc_addr": ":5021",
      "http_addr": ":5022",
      "node_id": "indexer3",
      "raft_storage_type": "boltdb"
    },
    "state": "Follower"
  }
}
```

Recommend 3 or more odd number of nodes in the cluster. In failure scenarios, data loss is inevitable, so avoid deploying single nodes.

The following command indexes documents to any node in the cluster:

```bash
$ cat ./example/wiki_doc_enwiki_1.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 enwiki_1
```

So, you can get the document from the node specified by the above command as follows:

```bash
$ ./bin/blast get document --grpc-addr=:5001 enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "_type": "enwiki",
  "contributor": "unknown",
  "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
  "timestamp": "2018-07-04T05:41:00Z",
  "title_en": "Search engine (computing)"
}
```

You can also get the same document from other nodes in the cluster as follows:

```bash
$ ./bin/blast get document --grpc-addr=:5011 enwiki_1
$ ./bin/blast get document --grpc-addr=:5021 enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "_type": "enwiki",
  "contributor": "unknown",
  "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
  "timestamp": "2018-07-04T05:41:00Z",
  "title_en": "Search engine (computing)"
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

### Bring up the manager cluster.

Manager can also bring up a cluster like an indexer. Specify a common index mapping for federation at startup.

```bash
$ ./bin/blastd \
    manager \
    --node-id=manager1 \
    --bind-addr=:15000 \
    --grpc-addr=:15001 \
    --http-addr=:15002 \
    --data-dir=/tmp/blast/manager1 \
    --raft-storage-type=badger \
    --index-mapping-file=./example/wiki_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb

$ ./bin/blastd \
    manager \
    --peer-addr=:15001 \
    --node-id=manager2 \
    --bind-addr=:15010 \
    --grpc-addr=:15011 \
    --http-addr=:15012 \
    --data-dir=/tmp/blast/manager2 \
    --raft-storage-type=badger

$ ./bin/blastd \
    manager \
    --peer-addr=:15001 \
    --node-id=manager3 \
    --bind-addr=:15020 \
    --grpc-addr=:15021 \
    --http-addr=:15022 \
    --data-dir=/tmp/blast/manager3
```

### Bring up the indexer cluster.

Federated mode differs from cluster mode that it specifies the manager in start up to bring up indexer cluster.  
The following example starts two 3-node clusters.

```bash
$ ./bin/blastd \
    indexer \
    --manager-addr=:15001 \
    --cluster-id=cluster1 \
    --node-id=indexer1 \
    --bind-addr=:5000 \
    --grpc-addr=:5001 \
    --http-addr=:5002 \
    --data-dir=/tmp/blast/indexer1

$ ./bin/blastd \
    indexer \
    --manager-addr=:15001 \
    --cluster-id=cluster1 \
    --node-id=indexer2 \
    --bind-addr=:5010 \
    --grpc-addr=:5011 \
    --http-addr=:5012 \
    --data-dir=/tmp/blast/indexer2

$ ./bin/blastd \
    indexer \
    --manager-addr=:15001 \
    --cluster-id=cluster1 \
    --node-id=indexer3 \
    --bind-addr=:5020 \
    --grpc-addr=:5021 \
    --http-addr=:5022 \
    --data-dir=/tmp/blast/indexer3

$ ./bin/blastd \
    indexer \
    --manager-addr=:15001 \
    --cluster-id=cluster2 \
    --node-id=indexer4 \
    --bind-addr=:5030 \
    --grpc-addr=:5031 \
    --http-addr=:5032 \
    --data-dir=/tmp/blast/indexer4

$ ./bin/blastd \
    indexer \
    --manager-addr=:15001 \
    --cluster-id=cluster2 \
    --node-id=indexer5 \
    --bind-addr=:5040 \
    --grpc-addr=:5041 \
    --http-addr=:5042 \
    --data-dir=/tmp/blast/indexer5

$ ./bin/blastd \
    indexer \
    --manager-addr=:15001 \
    --cluster-id=cluster2 \
    --node-id=indexer6 \
    --bind-addr=:5050 \
    --grpc-addr=:5051 \
    --http-addr=:5052 \
    --data-dir=/tmp/blast/indexer6
```

### Start up the dispatcher.

Finally, start the dispatcher with a manager that manages the target federation so that it can perform distributed search and indexing.

```bash
$ ./bin/blastd \
    dispatcher \
    --manager-addr=:15001 \
    --grpc-addr=:25001 \
    --http-addr=:25002
```

```bash
$ cat ./example/wiki_bulk_index.json | xargs -0 ./bin/blast set document --grpc-addr=:25001
```

```bash
$ cat ./example/wiki_search_request.json | xargs -0 ./bin/blast search --grpc-addr=:25001
```

```bash
$ cat ./example/wiki_bulk_delete.json | xargs -0 ./bin/blast delete document --grpc-addr=:25001
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
    -p 5000:5000 \
    -p 5001:5001 \
    -p 5002:5002 \
    -v $(pwd)/example:/opt/blast/example \
    mosuka/blast:latest blastd indexer \
      --node-id=blast-indexer1 \
      --bind-addr=:5000 \
      --grpc-addr=:5001 \
      --http-addr=:5002 \
      --data-dir=/tmp/blast/indexer1 \
      --index-mapping-file=/opt/blast/example/wiki_index_mapping.json \
      --index-storage-type=leveldb
```

You can execute the command in docker container as follows:

```bash
$ docker exec -it blast-indexer1 blast-indexer node --grpc-addr=:7070
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


### Indexing wikipedia dump

```shell
$ for FILE in $(find ~/tmp/enwiki -type f -name '*' | sort)
  do
    echo "Indexing ${FILE}"
    TIMESTAMP=$(date -u "+%Y-%m-%dT%H:%M:%SZ")
    DOCS=$(cat ${FILE} | jq -r '. + {fields: {url: .url, title_en: .title, text_en: .text, timestamp: "'${TIMESTAMP}'", _type: "enwiki"}} | del(.url) | del(.title) | del(.text) | del(.fields.id)' | jq -s)
    curl -s -X PUT -H 'Content-Type: application/json' "http://127.0.0.1:5002/documents" -d "${DOCS}"
  done
```


## Spatial/Geospatial search example

This section explain how to index Spatial/Geospatial data to Blast.

### Starting Indexer with Spatial/Geospatial index mapping

```bash
$ ./bin/blastd \
    indexer \
    --node-id=indexer1 \
    --bind-addr=:5000 \
    --grpc-addr=:5001 \
    --http-addr=:5002 \
    --data-dir=/tmp/blast/indexer1 \
    --index-mapping-file=./example/geo_index_mapping.json \
    --index-type=upside_down \
    --index-storage-type=boltdb
```

### Indexing example Spatial/Geospatial data

```bash
$ cat ./example/geo_doc1.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 geo_doc1
$ cat ./example/geo_doc2.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 geo_doc2
$ cat ./example/geo_doc3.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 geo_doc3
$ cat ./example/geo_doc4.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 geo_doc4
$ cat ./example/geo_doc5.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 geo_doc5
$ cat ./example/geo_doc6.json | xargs -0 ./bin/blast set document --grpc-addr=:5001 geo_doc6
```

### Searching example Spatial/Geospatial data

```bash
$ cat ./example/geo_search_request.json | xargs -0 ./bin/blast search --grpc-addr=:5001
```
