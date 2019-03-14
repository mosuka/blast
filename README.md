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

Blast is a full text search and indexing server written in [Go](https://golang.org) built on top of [Bleve](http://www.blevesearch.com).  
It provides functions through [gRPC](http://www.grpc.io) ([HTTP/2](https://en.wikipedia.org/wiki/HTTP/2) + [Protocol Buffers](https://developers.google.com/protocol-buffers/)) or traditional [RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer) API ([HTTP/1.1](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) + [JSON](http://www.json.org)).  
Blast implements [Raft consensus algorithm](https://raft.github.io/) by [hashicorp/raft](https://github.com/hashicorp/raft). It achieve consensus across all the instances of the nodes, ensuring that every change made to the system is made to a quorum of nodes, or none at all.
Blast makes it easy for programmers to develop search applications with advanced features.


## Features

- Full-text search and indexing
- Faceting
- Result highlighting
- Easy deployment
- Bringing up cluster
- Index replication
- An easy-to-use HTTP API
- CLI is also available
- Docker container image is available


## Installing dependencies

Blast requires some C/C++ libraries. The following sections are instructions for satisfying dependencies on particular platforms.


### Ubuntu 18.10

```bash
$ sudo apt-get install -y \
    libicu-dev \
    libstemmer-dev \
    gcc-4.8 \
    g++-4.8 \
    build-essential

$ sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 80
$ sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-8 80
$ sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 90
$ sudo update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-4.8 90

$ export GOPATH=${HOME}/go
$ go get -u -v github.com/blevesearch/cld2
$ cd ${GOPATH}/src/github.com/blevesearch/cld2
$ git clone https://github.com/CLD2Owners/cld2.git
$ cd cld2/internal
$ ./compile_libs.sh
$ sudo cp *.so /usr/local/lib
```


### macOS High Sierra Version 10.13.6

```bash
$ brew install \
    icu4c

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
$ git clone git@github.com:mosuka/blast.git
$ cd blast
$ make build
```

If you want to build for other platform, set `GOOS`, `GOARCH` environment variables. For example, build for macOS like following:

```bash
$ make \
    GOOS=darwin \
    build
```

Blast supports some [Bleve Extensions (blevex)](https://github.com/blevesearch/blevex). If you want to build with them, please set `CGO_LDFLAGS`, `CGO_CFLAGS`, `CGO_ENABLED` and `BUILD_TAGS`. For example, enable Japanese Language Analyzer like following:

```bash
$ make \
    BUILD_TAGS=kagome \
    build
```

You can enable supported Bleve Extensions for like following:


###  Linux

```bash
$ make \
    GOOS=linux \
    BUILD_TAGS="kagome icu libstemmer cld2" \
    CGO_ENABLED=1 \
    build
```


#### macOS

```bash
$ make GOOS=darwin \
    GOOS=darwin \
    BUILD_TAGS="kagome icu libstemmer cld2" \
    CGO_ENABLED=1 \
    CGO_LDFLAGS="-L/usr/local/opt/icu4c/lib" \
    CGO_CFLAGS="-I/usr/local/opt/icu4c/include" \
    build
```


Please refer to the following table for details of Bleve Extensions:

| BUILD_TAGS                 | CGO_ENABLED | Description                                  |
| -------------------------- | ----------- | -------------------------------------------- |
| cld2                       | 1           | Enable Compact Language Detector             |
| kagome                     | 0           | Enable Japanese Language Analyser            |
| icu                        | 1           | Enable ICU Tokenizer, Thai Language Analyser |
| libstemmer                 | 1           | Enable Language Stemmer (Danish, German, English, Spanish, Finnish, French, Hungarian, Italian, Dutch, Norwegian, Portuguese, Romanian, Russian, Swedish, Turkish) |

You can see the binary file when build successful like so:

```bash
$ ls ./bin
blast-index
```


## Starting Blast index node

Running a Blast index node is easy. Start Blast data node like so:

```bash
$ ./bin/blast-index start --node-id=index1 --data-dir=/tmp/blast/index1 --bind-addr=:6060 --grpc-addr=:5050 --http-addr=:8080 --index-mapping-file ./example/index_mapping.json
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
$ cat ./example/doc_enwiki_1.json | xargs -0 ./bin/blast-index index --grpc-addr=:5050 enwiki_1
```


### Getting a document via CLI

Getting a document is as following:

```bash
$ ./bin/blast-index get --grpc-addr=:5050 enwiki_1
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


### Searching documents via CLI

Searching documents is as like following:

```bash
$ cat ./example/search_request.json | xargs -0 ./bin/blast-index search --grpc-addr=:5050
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
      "index": "/tmp/blast/index1/index",
      "id": "enwiki_1",
      "score": 0.09634961191421738,
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
        "contributor": "unknown",
        "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
        "timestamp": "2018-07-04T05:41:00Z",
        "title_en": "Search engine (computing)"
      }
    }
  ],
  "total_hits": 1,
  "max_score": 0.09634961191421738,
  "took": 362726,
  "facets": {
    "Contributor count": {
      "field": "contributor",
      "total": 1,
      "missing": 0,
      "other": 0,
      "terms": [
        {
          "term": "unknown",
          "count": 1
        }
      ]
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
$ ./bin/blast-index delete --grpc-addr=:5050 enwiki_1
```


## Using HTTP REST API

Also you can do above commands via HTTP REST API that listened port 8080.


### Indexing a document via HTTP REST API

Putting a document via HTTP is as following:

```bash
$ curl -X PUT 'http://127.0.0.1:8080/documents/enwiki_1' -d @./example/doc_enwiki_1.json
```


### Getting a document via HTTP REST API

Getting a document via HTTP is as following:

```bash
$ curl -X GET 'http://127.0.0.1:8080/documents/enwiki_1'
```


### Searching documents via HTTP REST API

Searching documents via HTTP is as following:

```bash
$ curl -X POST 'http://127.0.0.1:8080/search' -d @./example/search_request.json
```


### Deleting a document via HTTP REST API

Deleting a document via HTTP is as following:

```bash
$ curl -X DELETE 'http://127.0.0.1:8080/documents/enwiki_1'
```


## Bringing up a cluster

Blast is easy to bring up the cluster. Blast data node is already running, but that is not fault tolerant. If you need to increase the fault tolerance, bring up 2 more data nodes like so:

```bash
$ ./bin/blast-index start --node-id=index2 --data-dir=/tmp/blast/index2 --bind-addr=:6061 --grpc-addr=:5051 --http-addr=:8081 --index-mapping-file ./example/index_mapping.json --join-addr=:5050
$ ./bin/blast-index start --node-id=index3 --data-dir=/tmp/blast/index3 --bind-addr=:6062 --grpc-addr=:5052 --http-addr=:8082 --index-mapping-file ./example/index_mapping.json --join-addr=:5050
```

_Above example shows each Blast node running on the same host, so each node must listen on different ports. This would not be necessary if each node ran on a different host._

This instructs each new node to join an existing node, each node recognizes the joining clusters when started.
So you have a 3-node cluster. That way you can tolerate the failure of 1 node. You can check the peers with the following command:

```bash
$ ./bin/blast-index cluster --grpc-addr=:5050
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "nodes": [
    {
      "id": "index1",
      "bind_addr": ":6060",
      "grpc_addr": ":5050",
      "http_addr": ":8080",
      "leader": true,
      "data_dir": "/tmp/blast/index1"
    },
    {
      "id": "index2",
      "bind_addr": ":6061",
      "grpc_addr": ":5051",
      "http_addr": ":8081",
      "data_dir": "/tmp/blast/index2"
    },
    {
      "id": "index3",
      "bind_addr": ":6062",
      "grpc_addr": ":5052",
      "http_addr": ":8082",
      "data_dir": "/tmp/blast/index3"
    }
  ]
}
```

Recommend 3 or more odd number of nodes in the cluster. In failure scenarios, data loss is inevitable, so avoid deploying single nodes.

The following command indexes documents to any node in the cluster:

```bash
$ cat ./example/doc_enwiki_1.json | xargs -0 ./bin/blast-index index --grpc-addr=:5050 enwiki_1
```

So, you can get the document from the node specified by the above command as follows:

```bash
$ ./bin/blast-index get --grpc-addr=:5050 enwiki_1
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
$ ./bin/blast-index get --grpc-addr=:5051 enwiki_1
$ ./bin/blast-index get --grpc-addr=:5052 enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "enwiki_1",
  "fields": {
    "_type": "enwiki",
    "contributor": "unknown",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  }
}
```


## Blast on Docker

### Building Docker container image on localhost

You can build the Docker container image like so:

```bash
$ make docker
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


### Running Blast index node on Docker

Running a Blast data node on Docker. Start Blast data node like so:

```bash
$ docker run --rm --name blast-index1 \
    -p 5050:5050 \
    -p 6060:6060 \
    -p 8080:8080 \
    -v $(pwd)/example:/opt/blast/example \
    mosuka/blast:latest blast-index start \
      --node-id=blast-index1 \
      --bind-addr=:6060 \
      --grpc-addr=:5050 \
      --http-addr=:8080 \
      --data-dir=/tmp/blast/index1 \
      --index-mapping-file=/opt/blast/example/index_mapping.json
```

You can execute the command in docker container as follows:

```bash
$ docker exec -it blast-index1 blast-index node --grpc-addr=:5050
```
