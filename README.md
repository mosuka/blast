Blast
======

The Blast is a full text search and indexing server written in [Go](https://golang.org) built on top of the [Bleve](http://www.blevesearch.com), [Bolt](https://github.com/boltdb/bolt) and [Raft](https://github.com/hashicorp/raft).  
Blast server provides functions through [gRPC](http://www.grpc.io) ([HTTP/2](https://en.wikipedia.org/wiki/HTTP/2) + [Protocol Buffers](https://developers.google.com/protocol-buffers/)) or traditional [RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer) API ([HTTP/1.1](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) + [JSON](http://www.json.org)), and uses [Raft](https://en.wikipedia.org/wiki/Raft_(computer_science)) to achieve consensus across all the instances of the nodes, ensuring that every change made to the system is made to a quorum of nodes, or none at all.  
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
- Docker container image is also available

## Building Blast

Blast requires Bleve and [Bleve Extensions](https://github.com/blevesearch/blevex). Some Bleve Extensions requires C/C++ libraries. The following sections are instructions for satisfying dependencies on particular platforms.

### Requirement

- Go
- Git

### Ubuntu 18.10

```
$ sudo apt-get install -y libicu-dev \
                          libleveldb-dev \
                          libstemmer-dev \
                          libgflags-dev \
                          libsnappy-dev \
                          zlib1g-dev \
                          libbz2-dev \
                          liblz4-dev \
                          libzstd-dev \
                          librocksdb-dev \
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
$ brew install icu4c \
               leveldb \
               rocksdb \
               zstd

$ CGO_LDFLAGS="-L/usr/local/opt/icu4c/lib" \
  CGO_CFLAGS="-I/usr/local/opt/icu4c/include" \
  go get -u -v github.com/blevesearch/cld2
$ cd ${GOPATH}/src/github.com/blevesearch/cld2
$ git clone https://github.com/CLD2Owners/cld2.git
$ cd cld2/internal
$ perl -p -i -e 's/soname=/install_name,/' compile_libs.sh
$ ./compile_libs.sh
$ sudo cp *.so /usr/local/lib
```

### Build Blast

Build Blast for Linux as following:

```bash
$ git clone git@github.com:mosuka/blast.git
$ cd blast
$ make build
```

If you want to build for other platform, set `GOOS`, `GOARCH`. For example, build for macOS like following:

```bash
$ GOOS=darwin \
  make build
```

If you want to build Blast with Bleve and Bleve Extentions (blevex), please set `CGO_LDFLAGS`, `CGO_CFLAGS`, `CGO_ENABLED` and `BUILD_TAGS`. For example, enable Japanese Language Analyzer like following:

```bash
$ BUILD_TAGS=kagome \
  make build
```

You can enable all Bleve Extensions for macOS like following:

```bash
$ GOOS=darwin \
  CGO_LDFLAGS="-L/usr/local/opt/icu4c/lib -L/usr/local/opt/rocksdb/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd" \
  CGO_CFLAGS="-I/usr/local/opt/icu4c/include -I/usr/local/opt/rocksdb/include" \
  CGO_ENABLED=1 \
  BUILD_TAGS="full" \
  make build
```

Also, you can enable all Bleve Extensions for Linux like following:

```bash
$ GOOS=linux \
  CGO_LDFLAGS="-L/usr/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd" \
  CGO_CFLAGS="-I/usr/include/rocksdb" \
  CGO_ENABLED=1 \
  BUILD_TAGS=full \
  make build
```

Please refer to the following table for details of Bleve Extensions:

| | CGO_ENABLED | BUILD_TAGS |
| --- | --- | --- |
| Compact Language Detector | 1 | cld2 |
| cznicb KV store |  | cznicb |
| Thai Language Analyser | 1 | icu |
| Japanese Language Analyser |  | kagome |
| LevelDB | 1 | leveldb |
| Danish, German, English, Spanish, Finnish, French, Hungarian, Italian, Dutch, Norwegian, Portuguese, Romanian, Russian, Swedish, Turkish Language Stemmer | 1 | libstemmer |
| RocksDB | 1 | rocksdb |

You can see the binary file when build successful like so:

```bash
$ ls ./bin
blast
```

## Running Blast node

Running a Blast node is easy. Start Blast node like so:

```bash
$ ./bin/blast start --bind-addr=localhost:10000 \
                    --grpc-addr=localhost:10001 \
                    --http-addr=localhost:10002 \
                    --node-id=node1 \
                    --raft-dir=/tmp/blast/noade1/raft \
                    --store-dir=/tmp/blast/node1/store \
                    --index-dir=/tmp/blast/node1/index \
                    --index-mapping=./example/wikipedia_index_mapping.json
```

## Using Blast CLI

You can now put, get, search and delete the document(s) via CLI.

### Putting a document

Putting a document is as following:

```bash
$ cat ./example/enwiki_doc1.json | xargs -0 ./bin/blast put --grpc-addr=localhost:10001 --pretty-print enwiki_doc1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "success": true
}
```

### Getting a document

Getting a document is as following:

```bash
$ ./bin/blast get --grpc-addr=localhost:10001 --pretty-print enwiki_doc1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "enwiki_doc1",
  "fields": {
    "_type": "enwiki",
    "contributor": "unknown",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  },
  "success": true
}
```

### Deleting a document

Deleting a document is as following:

```
$ ./bin/blast delete --grpc-addr=localhost:10001 --pretty-print enwiki_doc1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "success": true
}
```

### Bulk update

Bulk update is as following:

```
$ cat ./example/bulk_put_request.json | xargs -0 ./bin/blast bulk --pretty-print
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "put_count": 3,
  "success": true
}
```

### Searching documents

Searching documents is as like following:

```bash
$ cat ./example/wikipedia_search_request.json | xargs -0 ./bin/blast search --grpc-addr=localhost:10001 --pretty-print
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "search_result": {
    "facets": {
      "Contributor count": {
        "field": "contributor",
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "count": 3,
            "term": "unknown"
          }
        ],
        "total": 3
      },
      "Timestamp range": {
        "date_ranges": [
          {
            "count": 3,
            "end": "2020-12-31T23:59:59Z",
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z"
          }
        ],
        "field": "timestamp",
        "missing": 0,
        "other": 0,
        "total": 3
      }
    },
    "hits": [
      {
        "fields": {
          "_type": "enwiki",
          "contributor": "unknown",
          "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_en": "Search engine (computing)"
        },
        "id": "enwiki_doc1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "text_en": {
            "search": [
              {
                "array_positions": null,
                "end": 8,
                "pos": 2,
                "start": 2
              },
              {
                "array_positions": null,
                "end": 124,
                "pos": 20,
                "start": 118
              },
              {
                "array_positions": null,
                "end": 201,
                "pos": 33,
                "start": 195
              }
            ]
          },
          "title_en": {
            "search": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.18706386191354732,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "ptwiki",
          "contributor": "unknown",
          "text_pt": "Motor de pesquisa (português europeu) ou ferramenta de busca (português brasileiro) ou buscador (em inglês: search engine) é um programa desenhado para procurar palavras-chave fornecidas pelo utilizador em documentos e bases de dados. No contexto da internet, um motor de pesquisa permite procurar palavras-chave em documentos alojados na world wide web, como aqueles que se encontram armazenados em websites.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_pt": "Motor de busca"
        },
        "id": "ptwiki_doc1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "text_pt": {
            "search": [
              {
                "array_positions": null,
                "end": 117,
                "pos": 16,
                "start": 111
              }
            ]
          }
        },
        "score": 0.09273589609475594,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "jawiki",
          "contributor": "unknown",
          "text_ja": "検索エンジン（けんさくエンジン、英語: search engine）は、狭義にはインターネットに存在する情報（ウェブページ、ウェブサイト、画像フみを提供していたウェブサイトそのものを検索エンジンと呼んだが、現在では様々なサービスが加わったポータルサイト化が進んだため、検索をサービスの一つとしてトを単に検索サイトと呼ぶことはなくなっている。広義には、インターネットに限定せず情報を検索するシステム全般を含む。",
          "timestamp": "2018-05-30T00:52:00Z",
          "title_ja": "検索エンジン"
        },
        "id": "jawiki_doc1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "text_ja": {
            "search": [
              {
                "array_positions": null,
                "end": 62,
                "pos": 11,
                "start": 56
              }
            ]
          }
        },
        "score": 0.05875099443799347,
        "sort": [
          "_score"
        ]
      }
    ],
    "max_score": 0.18706386191354732,
    "request": {
      "explain": false,
      "facets": {
        "Contributor count": {
          "field": "contributor",
          "size": 10
        },
        "Timestamp range": {
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
          ],
          "field": "timestamp",
          "size": 10
        }
      },
      "fields": [
        "*"
      ],
      "from": 0,
      "highlight": {
        "fields": [
          "title",
          "text"
        ],
        "style": "html"
      },
      "includeLocations": false,
      "query": {
        "query": "+_all:search"
      },
      "size": 10,
      "sort": [
        "-_score"
      ]
    },
    "status": {
      "failed": 0,
      "successful": 1,
      "total": 1
    },
    "took": 318296,
    "total_hits": 3
  },
  "success": true
}
```


## Using HTTP REST API

Also you can do above commands via HTTP REST API that listened port 10001 (address port + 1).

### Putting a document

Putting a document via HTTP is as following:

```bash
$ curl -X PUT 'http://localhost:10002/rest/enwiki_doc1?pretty-print=true' -d @./example/enwiki_doc1.json
```

You can see the result in JSON format. The result of the above request is:

```json
{
  "success": true
}
```

### Getting a document

Getting a document via HTTP is as following:

```bash
$ curl -X GET 'http://localhost:10002/rest/enwiki_doc1?pretty-print=true'
```

You can see the result in JSON format. The result of the above request is:

```json
{
  "id": "enwiki_doc1",
  "fields": {
    "_type": "enwiki",
    "contributor": "unknown",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  },
  "success": true
}
```

### Deleting a document

Deleting a document via HTTP is as following:

```bash
$ curl -X DELETE 'http://localhost:10002/rest/enwiki_doc1?pretty-print=true'
```

You can see the result in JSON format. The result of the above request is:

```json
{
  "success": true
}
```

### Bulk update

Bulk update is as following:

```
$ curl -X POST 'http://localhost:10002/rest/_bulk?pretty-print=true' -d @./example/bulk_put_request.json
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "put_count": 3,
  "success": true
}
```

### Searching documents

Searching documents via HTTP is as following:

```bash
$ curl -X POST 'http://localhost:10002/rest/_search?pretty-print=true' -d @./example/wikipedia_search_request.json
```

You can see the result in JSON format. The result of the above request is:

```json
{
  "search_result": {
    "facets": {
      "Contributor count": {
        "field": "contributor",
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "count": 3,
            "term": "unknown"
          }
        ],
        "total": 3
      },
      "Timestamp range": {
        "date_ranges": [
          {
            "count": 3,
            "end": "2020-12-31T23:59:59Z",
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z"
          }
        ],
        "field": "timestamp",
        "missing": 0,
        "other": 0,
        "total": 3
      }
    },
    "hits": [
      {
        "fields": {
          "_type": "enwiki",
          "contributor": "unknown",
          "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_en": "Search engine (computing)"
        },
        "id": "enwiki_doc1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "text_en": {
            "search": [
              {
                "array_positions": null,
                "end": 8,
                "pos": 2,
                "start": 2
              },
              {
                "array_positions": null,
                "end": 124,
                "pos": 20,
                "start": 118
              },
              {
                "array_positions": null,
                "end": 201,
                "pos": 33,
                "start": 195
              }
            ]
          },
          "title_en": {
            "search": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.18706386191354732,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "ptwiki",
          "contributor": "unknown",
          "text_pt": "Motor de pesquisa (português europeu) ou ferramenta de busca (português brasileiro) ou buscador (em inglês: search engine) é um programa desenhado para procurar palavras-chave fornecidas pelo utilizador em documentos e bases de dados. No contexto da internet, um motor de pesquisa permite procurar palavras-chave em documentos alojados na world wide web, como aqueles que se encontram armazenados em websites.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_pt": "Motor de busca"
        },
        "id": "ptwiki_doc1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "text_pt": {
            "search": [
              {
                "array_positions": null,
                "end": 117,
                "pos": 16,
                "start": 111
              }
            ]
          }
        },
        "score": 0.09273589609475594,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "jawiki",
          "contributor": "unknown",
          "text_ja": "検索エンジン（けんさくエンジン、英語: search engine）は、狭義にはインターネットに存在する情報（ウェブページ、ウェブサイト、画像フム。インターネットの普及初期には、検索としての機能のみを提供していたウェブサイトそのものを検索エンジンと呼んだが、現在では様々なサービスが加わったポー情報を検索するシステム全般を含む。",
          "timestamp": "2018-05-30T00:52:00Z",
          "title_ja": "検索エンジン"
        },
        "id": "jawiki_doc1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "text_ja": {
            "search": [
              {
                "array_positions": null,
                "end": 62,
                "pos": 11,
                "start": 56
              }
            ]
          }
        },
        "score": 0.05875099443799347,
        "sort": [
          "_score"
        ]
      }
    ],
    "max_score": 0.18706386191354732,
    "request": {
      "explain": false,
      "facets": {
        "Contributor count": {
          "field": "contributor",
          "size": 10
        },
        "Timestamp range": {
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
          ],
          "field": "timestamp",
          "size": 10
        }
      },
      "fields": [
        "*"
      ],
      "from": 0,
      "highlight": {
        "fields": [
          "title",
          "text"
        ],
        "style": "html"
      },
      "includeLocations": false,
      "query": {
        "query": "+_all:search"
      },
      "size": 10,
      "sort": [
        "-_score"
      ]
    },
    "status": {
      "failed": 0,
      "successful": 1,
      "total": 1
    },
    "took": 332682,
    "total_hits": 3
  },
  "success": true
}
```

## Bringing up a cluster

Blast is easy to bring up the cluster. Blast node is already running, but that is not fault tolerant. If you need to increase the fault tolerance, bring up 2 more nodes like so:

```bash
$ ./bin/blast start --bind-addr=localhost:11000 \
                    --grpc-addr=localhost:11001 \
                    --http-addr=localhost:11002 \
                    --node-id=node2 \
                    --raft-dir=/tmp/blast/node2/raft \
                    --store-dir=/tmp/blast/node2/store \
                    --index-dir=/tmp/blast/node2/index \
                    --peer-grpc-addr=localhost:10001 \
                    --index-mapping=./example/wikipedia_index_mapping.json
$ ./bin/blast start --bind-addr=localhost:12000 \
                    --grpc-addr=localhost:12001 \
                    --http-addr=localhost:12002 \
                    --node-id=node3 \
                    --raft-dir=/tmp/blast/node3/raft \
                    --store-dir=/tmp/blast/node3/store \
                    --index-dir=/tmp/blast/node3/index \
                    --peer-grpc-addr=localhost:10001 \
                    --index-mapping=./example/wikipedia_index_mapping.json
```

_Above example shows each Blast node running on the same host, so each node must listen on different ports. This would not be necessary if each node ran on a different host._

So you have a 3-node cluster. That way you can tolerate the failure of 1 node. You can check the peers with the following command:

```bash
$ ./bin/blast peers --grpc-addr=localhost:10001 --pretty-print
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "peers": [
    {
      "address": "127.0.0.1:10000",
      "leader": true,
      "metadata": {
        "grpc_address": "localhost:10001",
        "http_address": "localhost:10002"
      },
      "node_id": "node1"
    },
    {
      "address": "127.0.0.1:11000",
      "metadata": {
        "grpc_address": "localhost:11001",
        "http_address": "localhost:11002"
      },
      "node_id": "node2"
    },
    {
      "address": "127.0.0.1:12000",
      "metadata": {
        "grpc_address": "localhost:12001",
        "http_address": "localhost:12002"
      },
      "node_id": "node3"
    }
  ],
  "success": true
}
```

Recommend 3 or more odd number of nodes in the cluster. In failure scenarios, data loss is inevitable, so avoid deploying single nodes.

This tells each new node to join the existing node. Once joined, each node now knows about the key:

Following command puts a document to node0:

```bash
$ cat ./example/enwiki_doc1.json | xargs -0 ./bin/blast put --grpc-addr=localhost:10001 --pretty-print enwiki_doc1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "success": true
}
```

So, you can get a document from node1 like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:10001 --pretty-print enwiki_doc1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "enwiki_doc1",
  "fields": {
    "_type": "enwiki",
    "contributor": "unknown",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  },
  "success": true
}
```

Also, you can get same document from node2 (localhost:11001) like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:11001 --pretty-print enwiki_doc1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "enwiki_doc1",
  "fields": {
    "_type": "enwiki",
    "contributor": "unknown",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  },
  "success": true
}
```

Lastly, you can get same document from node3 (localhost:12001) like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:12001 --pretty-print enwiki_doc1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "enwiki_doc1",
  "fields": {
    "_type": "enwiki",
    "contributor": "unknown",
    "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title_en": "Search engine (computing)"
  },
  "success": true
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
$ docker pull mosuka/blast:v0.2.0
```

See https://hub.docker.com/r/mosuka/blast/tags/

### Running Blast node on Docker

Running a Blast node on Docker. Start Blast node like so:

```bash
$ docker run --rm --name blast1 \
    -p 10000:10000 \
    -p 10001:10001 \
    -p 10002:10002 \
    mosuka/blast:v0.3.0 start \
    --bind-addr=:10000 \
    --grpc-addr=:10001 \
    --http-addr=:10002 \
    --node-id=node1
```

### Running Blast cluster on Docker Compose

You can also bringing up cluster with Docker Compose.
First, you need to start 1st node of cluster like so:

```bash
$ docker-compose up -d blast1
```

Then, you can start other nodes like so:

```bash
$ docker-compose up -d blast2 blast3
```

All nodes are stopped as follows:

```bash
$ docker-compose stop
```
