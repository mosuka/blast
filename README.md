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
- Index replication
- Bringing up cluster
- An easy-to-use HTTP API
- CLI is available
- Docker container image is available


## Install build dependencies

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


## Build

Building Blast as following:

```bash
$ mkdir -p ${GOPATH}/src/github.com/mosuka
$ cd ${GOPATH}/src/github.com/mosuka
$ git clone https://github.com/mosuka/blast.git
$ cd blast
$ make
```

If you omit `GOOS` or `GOARCH`, it will build the binary of the platform you are using.  
If you want to specify the target platform, please set `GOOS` and `GOARCH` environment variables.

### Linux

```bash
$ make GOOS=linux build
```

### macOS

```bash
$ make GOOS=darwin build
```

### Windows

```bash
$ make GOOS=windows build
```

## Build with extensions

Blast supports some Bleve Extensions (blevex). If you want to build with them, please set CGO_LDFLAGS, CGO_CFLAGS, CGO_ENABLED and BUILD_TAGS. For example, build LevelDB to be available for index storage as follows:

```bash
$ make GOOS=linux \
       BUILD_TAGS=icu \
       CGO_ENABLED=1 \
       build
```

### Linux

```bash
$ make GOOS=linux \
       BUILD_TAGS="kagome icu libstemmer cld2" \
       CGO_ENABLED=1 \
       build
```

### macOS

```bash
$ make GOOS=darwin \
       BUILD_TAGS="kagome icu libstemmer cld2" \
       CGO_ENABLED=1 \
       CGO_LDFLAGS="-L/usr/local/opt/icu4c/lib" \
       CGO_CFLAGS="-I/usr/local/opt/icu4c/include" \
       build
```

### Buil flags

Refer to the following table for the build flags of the supported Bleve extensions:

| BUILD_TAGS | CGO_ENABLED | Description |
| ---------- | ----------- | ----------- |
| cld2       | 1           | Enable Compact Language Detector |
| kagome     | 0           | Enable Japanese Language Analyser |
| icu        | 1           | Enable ICU Tokenizer, Thai Language Analyser |
| libstemmer | 1           | Enable Language Stemmer (Danish, German, English, Spanish, Finnish, French, Hungarian, Italian, Dutch, Norwegian, Portuguese, Romanian, Russian, Swedish, Turkish) |

If you want to enable the feature whose `CGO_ENABLE` is `1`, please install it referring to the Install build dependencies section above.


## Binary

You can see the binary file when build successful like so:

```bash
$ ls ./bin
blast
```


## Test

If you want to test your changes, run command like following:

```bash
$ make test
```

If you want to specify the target platform, set `GOOS` and `GOARCH` environment variables in the same way as the build.


## Package

To create a distribution package, run the following command:

```bash
$ make dist
```


## Configure

Blast can change its startup options with configuration files, environment variables, and command line arguments.  
Refer to the following table for the options that can be configured.

| CLI Flag | Environment variable | Configuration File | Description |
| --- | --- | --- | --- |
| --config-file | - | - | config file. if omitted, blast.yaml in /etc and home directory will be searched |
| --id | BLAST_ID | id | node ID |
| --raft-address | BLAST_RAFT_ADDRESS | raft_address | Raft server listen address |
| --grpc-address | BLAST_GRPC_ADDRESS | grpc_address | gRPC server listen address |
| --http-address | BLAST_HTTP_ADDRESS | http_address | HTTP server listen address |
| --data-directory | BLAST_DATA_DIRECTORY | data_directory | data directory which store the index and Raft logs |
| --mapping-file | BLAST_MAPPING_FILE | mapping_file | path to the index mapping file |
| --peer-grpc-address | BLAST_PEER_GRPC_ADDRESS | peer_grpc_address | listen address of the existing gRPC server in the joining cluster |
| --certificate-file | BLAST_CERTIFICATE_FILE | certificate_file | path to the client server TLS certificate file |
| --key-file | BLAST_KEY_FILE | key_file | path to the client server TLS key file |
| --common-name | BLAST_COMMON_NAME | common_name | certificate common name |
| --log-level | BLAST_LOG_LEVEL | log_level | log level |
| --log-file | BLAST_LOG_FILE | log_file | log file |
| --log-max-size | BLAST_LOG_MAX_SIZE | log_max_size | max size of a log file in megabytes |
| --log-max-backups | BLAST_LOG_MAX_BACKUPS | log_max_backups | max backup count of log files |
| --log-max-age | BLAST_LOG_MAX_AGE | log_max_age | max age of a log file in days |
| --log-compress | BLAST_LOG_COMPRESS | log_compress | compress a log file |


## Start

Starting server is easy as follows:

```bash
$ ./bin/blast start \
              --id=node1 \
              --raft-address=:7000 \
              --http-address=:8000 \
              --grpc-address=:9000 \
              --data-directory=/tmp/blast/node1 \
              --mapping-file=./examples/example_mapping.json
```

You can get the node information with the following command:

```bash
$ ./bin/blast node | jq .
```

or the following URL:

```bash
$ curl -X GET http://localhost:8000/v1/node | jq .
```

The result of the above command is:

```json
{
  "node": {
    "raft_address": ":7000",
    "metadata": {
      "grpc_address": ":9000",
      "http_address": ":8000"
    },
    "state": "Leader"
  }
}
```

## Health check

You can check the health status of the node.

```bash
$ ./bin/blast healthcheck | jq .
```

Also provides the following REST APIs

### Liveness prove

This endpoint always returns 200 and should be used to check server health.

```bash
$ curl -X GET http://localhost:8000/v1/liveness_check | jq .
```

### Readiness probe

This endpoint returns 200 when server is ready to serve traffic (i.e. respond to queries).

```bash
$ curl -X GET http://localhost:8000/v1/readiness_check | jq .
```

## Put a document

To put a document, execute the following command:

```bash
$ ./bin/blast set 1 '
{
  "fields": {
    "title": "Search engine (computing)",
    "text": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "_type": "example"
  }
}
' | jq .
```

or, you can use the RESTful API as follows:

```bash
$ curl -X PUT 'http://127.0.0.1:8000/v1/documents/1' --data-binary '
{
  "fields": {
    "title": "Search engine (computing)",
    "text": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "_type": "example"
  }
}
' | jq .
```

or

```bash
$ curl -X PUT 'http://127.0.0.1:8000/v1/documents/1' -H "Content-Type: application/json" --data-binary @./examples/example_doc_1.json
```

## Get a document

To get a document, execute the following command:

```bash
$ ./bin/blast get 1 | jq .
```

or, you can use the RESTful API as follows:

```bash
$ curl -X GET 'http://127.0.0.1:8000/v1/documents/1' | jq .
```

You can see the result. The result of the above command is:

```json
{
  "fields": {
    "_type": "example",
    "text": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title": "Search engine (computing)"
  }
}
```

## Search documents

To search documents, execute the following command:

```bash
$ ./bin/blast search '
{
  "search_request": {
    "query": {
      "query": "+_all:search"
    },
    "size": 10,
    "from": 0,
    "fields": [
      "*"
    ],
    "sort": [
      "-_score"
    ]
  }
}
' | jq .
```

or, you can use the RESTful API as follows:

```bash
$ curl -X POST 'http://127.0.0.1:8000/v1/search' --data-binary '
{
  "search_request": {
    "query": {
      "query": "+_all:search"
    },
    "size": 10,
    "from": 0,
    "fields": [
      "*"
    ],
    "sort": [
      "-_score"
    ]
  }
}
' | jq .
```

You can see the result. The result of the above command is:

```json
{
  "search_result": {
    "facets": null,
    "hits": [
      {
        "fields": {
          "_type": "example",
          "text": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title": "Search engine (computing)"
        },
        "id": "1",
        "index": "/tmp/blast/node1/index",
        "score": 0.09703538256409851,
        "sort": [
          "_score"
        ]
      }
    ],
    "max_score": 0.09703538256409851,
    "request": {
      "explain": false,
      "facets": null,
      "fields": [
        "*"
      ],
      "from": 0,
      "highlight": null,
      "includeLocations": false,
      "query": {
        "query": "+_all:search"
      },
      "search_after": null,
      "search_before": null,
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
    "took": 171880,
    "total_hits": 1
  }
}
```

## Delete a document

Deleting a document, execute the following command:

```bash
$ ./bin/blast delete 1
```

or, you can use the RESTful API as follows:

```bash
$ curl -X DELETE 'http://127.0.0.1:8000/v1/documents/1'
```

## Index documents in bulk

To index documents in bulk, execute the following command:

```bash
$ ./bin/blast bulk-index --file ./examples/example_bulk_index.json
```

or, you can use the RESTful API as follows:

```bash
$ curl -X PUT 'http://127.0.0.1:8000/v1/documents' -H "Content-Type: application/x-ndjson" --data-binary @./examples/example_bulk_index.json
```

## Delete documents in bulk

To delete documents in bulk, execute the following command:

```bash
$ ./bin/blast bulk-delete --file ./examples/example_bulk_delete.txt
```

or, you can use the RESTful API as follows:

```bash
$ curl -X DELETE 'http://127.0.0.1:8000/v1/documents' -H "Content-Type: text/plain" --data-binary @./examples/example_bulk_delete.txt
```

## Bringing up a cluster

Blast is easy to bring up the cluster. the node is already running, but that is not fault tolerant. If you need to increase the fault tolerance, bring up 2 more data nodes like so:

```bash
$ ./bin/blast start \
              --id=node2 \
              --raft-address=:7001 \
              --http-address=:8001 \
              --grpc-address=:9001 \
              --peer-grpc-address=:9000 \
              --data-directory=/tmp/blast/node2 \
              --mapping-file=./examples/example_mapping.json
```

```bash
$ ./bin/blast start \
              --id=node3 \
              --raft-address=:7002 \
              --http-address=:8002 \
              --grpc-address=:9002 \
              --peer-grpc-address=:9000 \
              --data-directory=/tmp/blast/node3 \
              --mapping-file=./examples/example_mapping.json
```


_Above example shows each Blast node running on the same host, so each node must listen on different ports. This would not be necessary if each node ran on a different host._

This instructs each new node to join an existing node, each node recognizes the joining clusters when started.
So you have a 3-node cluster. That way you can tolerate the failure of 1 node. You can check the cluster with the following command:

```bash
$ ./bin/blast cluster | jq .
```

or, you can use the RESTful API as follows:

```bash
$ curl -X GET 'http://127.0.0.1:8000/v1/cluster' | jq .
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "cluster": {
    "nodes": {
      "node1": {
        "raft_address": ":7000",
        "metadata": {
          "grpc_address": ":9000",
          "http_address": ":8000"
        },
        "state": "Leader"
      },
      "node2": {
        "raft_address": ":7001",
        "metadata": {
          "grpc_address": ":9001",
          "http_address": ":8001"
        },
        "state": "Follower"
      },
      "node3": {
        "raft_address": ":7002",
        "metadata": {
          "grpc_address": ":9002",
          "http_address": ":8002"
        },
        "state": "Follower"
      }
    },
    "leader": "node1"
  }
}
```

Recommend 3 or more odd number of nodes in the cluster. In failure scenarios, data loss is inevitable, so avoid deploying single nodes.

The above example, the node joins to the cluster at startup, but you can also join the node that already started on standalone mode to the cluster later, as follows:

```bash
$ ./bin/blast join --grpc-address=:9000 node2 127.0.0.1:9001
```

or, you can use the RESTful API as follows:

```bash
$ curl -X PUT 'http://127.0.0.1:8000/v1/cluster/node2' --data-binary '
{
  "raft_address": ":7001",
  "metadata": {
    "grpc_address": ":9001",
    "http_address": ":8001"
  }
}
'
```

To remove a node from the cluster, execute the following command:

```bash
$ ./bin/blast leave --grpc-address=:9000 node2
```

or, you can use the RESTful API as follows:

```bash
$ curl -X DELETE 'http://127.0.0.1:8000/v1/cluster/node2'
```

The following command indexes documents to any node in the cluster:

```bash
$ ./bin/blast set 1 '
{
  "fields": {
    "title": "Search engine (computing)",
    "text": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "_type": "example"
  }
}
' --grpc-address=:9000 | jq .
```

So, you can get the document from the node specified by the above command as follows:

```bash
$ ./bin/blast get 1 --grpc-address=:9000 | jq .
```

You can see the result. The result of the above command is:

```text
value1
```

You can also get the same document from other nodes in the cluster as follows:

```bash
$ ./bin/blast get 1 --grpc-address=:9001 | jq .
$ ./bin/blast get 1 --grpc-address=:9002 | jq .
```

You can see the result. The result of the above command is:

```json
{
  "fields": {
    "_type": "example",
    "text": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
    "timestamp": "2018-07-04T05:41:00Z",
    "title": "Search engine (computing)"
  }
}
```


## Docker

### Build Docker container image

You can build the Docker container image like so:

```bash
$ make docker-build
```

### Pull Docker container image from docker.io

You can also use the Docker container image already registered in docker.io like so:

```bash
$ docker pull mosuka/blast:latest
```

See https://hub.docker.com/r/mosuka/blast/tags/

### Start on Docker

Running a Blast data node on Docker. Start Blast node like so:

```bash
$ docker run --rm --name blast-node1 \
    -p 7000:7000 \
    -p 8000:8000 \
    -p 9000:9000 \
    -v $(pwd)/etc/blast_mapping.json:/etc/blast_mapping.json \
    mosuka/blast:latest start \
      --id=node1 \
      --raft-address=:7000 \
      --http-address=:8000 \
      --grpc-address=:9000 \
      --data-directory=/tmp/blast/node1 \
      --mapping-file=/etc/blast_mapping.json
```

You can execute the command in docker container as follows:

```bash
$ docker exec -it blast-node1 blast node --grpc-address=:9000
```

## Securing Blast

Blast supports HTTPS access, ensuring that all communication between clients and a cluster is encrypted.

### Generating a certificate and private key

One way to generate the necessary resources is via [openssl](https://www.openssl.org/). For example:

```bash
$ openssl req -x509 -nodes -newkey rsa:4096 -keyout ./etc/blast_key.pem -out ./etc/blast_cert.pem -days 365 -subj '/CN=localhost'
Generating a 4096 bit RSA private key
............................++
........++
writing new private key to 'key.pem'
```

### Secure cluster example

Starting a node with HTTPS enabled, node-to-node encryption, and with the above configuration file. It is assumed the HTTPS X.509 certificate and key are at the paths server.crt and key.pem respectively.

```bash
$ ./bin/blast start \
             --id=node1 \
             --raft-address=:7000 \
             --http-address=:8000 \
             --grpc-address=:9000 \
             --peer-grpc-address=:9000 \
             --data-directory=/tmp/blast/node1 \
             --mapping-file=./etc/blast_mapping.json \
             --certificate-file=./etc/blast_cert.pem \
             --key-file=./etc/blast_key.pem \
             --common-name=localhost
```

```bash
$ ./bin/blast start \
             --id=node2 \
             --raft-address=:7001 \
             --http-address=:8001 \
             --grpc-address=:9001 \
             --peer-grpc-address=:9000 \
             --data-directory=/tmp/blast/node2 \
             --mapping-file=./etc/blast_mapping.json \
             --certificate-file=./etc/blast_cert.pem \
             --key-file=./etc/blast_key.pem \
             --common-name=localhost
```

```bash
$ ./bin/blast start \
             --id=node3 \
             --raft-address=:7002 \
             --http-address=:8002 \
             --grpc-address=:9002 \
             --peer-grpc-address=:9000 \
             --data-directory=/tmp/blast/node3 \
             --mapping-file=./etc/blast_mapping.json \
             --certificate-file=./etc/blast_cert.pem \
             --key-file=./etc/blast_key.pem \
             --common-name=localhost
```

You can access the cluster by adding a flag, such as the following command:

```bash
$ ./bin/blast cluster --grpc-address=:9000 --certificate-file=./etc/blast_cert.pem --common-name=localhost | jq .
```

or

```bash
$ curl -X GET https://localhost:8000/v1/cluster --cacert ./etc/cert.pem | jq .
```
