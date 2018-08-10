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

## Building Blast

Building Blast requires Go 1.9 or later. To build Blast on Linux like so:

```bash
$ git clone git@github.com:mosuka/blast.git
$ cd blast
$ make build
```

If you want to build Blast other platform, please set `GOOS`、`GOARCH` like following:

```bash
$ make GOOS=darwin build
```

You can see the binary file when build successful like so:

```bash
$ ls ./bin
blast
```

## Running Blast node

Running a Blast node is easy. Start Blast node like so:

```bash
$ ./bin/blast start --bind-addr=localhost:10000 --grpc-addr=localhost:10001 --http-addr=localhost:10002 --node-id=node0 --raft-dir=/tmp/blast/noade0/raft --store-dir=/tmp/blast/node0/store --index-dir=/tmp/blast/node0/index
```

## Using Blast CLI

You can now put, get, search and delete the document(s) via CLI.

### Putting a document

Putting a document is as following:

```bash
$ cat ./example/doc1.json | xargs -0 ./bin/blast put --grpc-addr=localhost:10001 --pretty-print 1
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
$ ./bin/blast get --grpc-addr=localhost:10001 --pretty-print 1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "1",
  "fields": {
    "category": "Library",
    "description": "Bleve is a full-text search and indexing library for Go.",
    "name": "Bleve",
    "popularity": 3,
    "release": "2014-04-18T00:00:00Z"
  },
  "success": true
}
```

### Deleting a document

Deleting a document is as following:

```
$ ./bin/blast delete --grpc-addr=localhost:10001 --pretty-print 1
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
  "put_count": 7,
  "success": true
}
```

### Searching documents

Searching documents is as like following:

```bash
$ cat ./example/search_request.json | xargs -0 ./bin/blast search --grpc-addr=localhost:10001 --pretty-print
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "search_result": {
    "facets": {
      "Category count": {
        "field": "category",
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "count": 4,
            "term": "library"
          },
          {
            "count": 3,
            "term": "server"
          }
        ],
        "total": 7
      },
      "Popularity range": {
        "field": "popularity",
        "missing": 0,
        "numeric_ranges": [
          {
            "count": 2,
            "max": 4,
            "min": 3,
            "name": "more than or equal to 3 and less than 4"
          },
          {
            "count": 2,
            "min": 5,
            "name": "more than or equal to 5"
          },
          {
            "count": 1,
            "max": 2,
            "min": 1,
            "name": "more than or equal to 1 and less than 2"
          },
          {
            "count": 1,
            "max": 3,
            "min": 2,
            "name": "more than or equal to 2 and less than 3"
          },
          {
            "count": 1,
            "max": 5,
            "min": 4,
            "name": "more than or equal to 4 and less than 5"
          }
        ],
        "other": 0,
        "total": 7
      },
      "Release date range": {
        "date_ranges": [
          {
            "count": 4,
            "end": "2010-12-31T23:59:59Z",
            "name": "2001 - 2010",
            "start": "2001-01-01T00:00:00Z"
          },
          {
            "count": 2,
            "end": "2020-12-31T23:59:59Z",
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z"
          }
        ],
        "field": "release",
        "missing": 0,
        "other": 0,
        "total": 6
      }
    },
    "hits": [
      {
        "fields": {
          "category": "Library",
          "description": "Bleve is a full-text search and indexing library for Go.",
          "name": "Bleve",
          "popularity": 3,
          "release": "2014-04-18T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Bleve is a full-text search and indexing library for Go."
          ],
          "name": [
            "\u003cmark\u003eBleve\u003c/mark\u003e"
          ]
        },
        "id": "1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "bleve": [
              {
                "array_positions": null,
                "end": 5,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Library",
          "description": "Apache Lucene is a high-performance, full-featured text search engine library written entirely in Java.",
          "name": "Lucene",
          "popularity": 4,
          "release": "2000-03-30T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Apache Lucene is a high-performance, full-featured text search engine library written entirely in Java."
          ],
          "name": [
            "\u003cmark\u003eLucene\u003c/mark\u003e"
          ]
        },
        "id": "2",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "lucene": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Library",
          "description": "Whoosh is a fast, featureful full-text indexing and searching library implemented in pure Python.",
          "name": "Whoosh",
          "popularity": 3,
          "release": "2008-02-20T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Whoosh is a fast, featureful full-text indexing and searching library implemented in pure Python."
          ],
          "name": [
            "\u003cmark\u003eWhoosh\u003c/mark\u003e"
          ]
        },
        "id": "3",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "whoosh": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Library",
          "description": "Ferret is a super fast, highly configurable search library written in Ruby.",
          "name": "Ferret",
          "popularity": 2,
          "release": "2005-10-01T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Ferret is a super fast, highly configurable search library written in Ruby."
          ],
          "name": [
            "\u003cmark\u003eFerret\u003c/mark\u003e"
          ]
        },
        "id": "4",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "ferret": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Server",
          "description": "Solr is an open source enterprise search platform, written in Java, from the Apache Lucene project.",
          "name": "Solr",
          "popularity": 5,
          "release": "2006-12-22T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Solr is an open source enterprise search platform, written in Java, from the Apache Lucene project."
          ],
          "name": [
            "\u003cmark\u003eSolr\u003c/mark\u003e"
          ]
        },
        "id": "5",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "solr": [
              {
                "array_positions": null,
                "end": 4,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Server",
          "description": "Elasticsearch is a search engine based on Lucene, written in Java.",
          "name": "Elasticsearch",
          "popularity": 5,
          "release": "2010-02-08T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Elasticsearch is a search engine based on Lucene, written in Java."
          ],
          "name": [
            "\u003cmark\u003eElasticsearch\u003c/mark\u003e"
          ]
        },
        "id": "6",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "elasticsearch": [
              {
                "array_positions": null,
                "end": 13,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Server",
          "description": "Blast is a full-text search and indexing server written in Go, built on top of Bleve.",
          "name": "Blast",
          "popularity": 1,
          "release": "2017-01-13T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Blast is a full-text search and indexing server written in Go, built on top of Bleve."
          ],
          "name": [
            "\u003cmark\u003eBlast\u003c/mark\u003e"
          ]
        },
        "id": "7",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "blast": [
              {
                "array_positions": null,
                "end": 5,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      }
    ],
    "max_score": 0.12163776688600772,
    "request": {
      "explain": false,
      "facets": {
        "Category count": {
          "field": "category",
          "size": 10
        },
        "Popularity range": {
          "field": "popularity",
          "numeric_ranges": [
            {
              "max": 1,
              "name": "less than 1"
            },
            {
              "max": 2,
              "min": 1,
              "name": "more than or equal to 1 and less than 2"
            },
            {
              "max": 3,
              "min": 2,
              "name": "more than or equal to 2 and less than 3"
            },
            {
              "max": 4,
              "min": 3,
              "name": "more than or equal to 3 and less than 4"
            },
            {
              "max": 5,
              "min": 4,
              "name": "more than or equal to 4 and less than 5"
            },
            {
              "min": 5,
              "name": "more than or equal to 5"
            }
          ],
          "size": 10
        },
        "Release date range": {
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
          "field": "release",
          "size": 10
        }
      },
      "fields": [
        "*"
      ],
      "from": 0,
      "highlight": {
        "fields": [
          "name",
          "description"
        ],
        "style": "html"
      },
      "includeLocations": false,
      "query": {
        "query": "name:*"
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
    "took": 446410,
    "total_hits": 7
  },
  "success": true
}
```


## Using HTTP REST API

Also you can do above commands via HTTP REST API that listened port 10001 (address port + 1).

### Putting a document

Putting a document via HTTP is as following:

```bash
$ curl -X PUT 'http://localhost:10002/rest/1?pretty-print=true' -d @./example/doc1.json
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
$ curl -X GET 'http://localhost:10002/rest/1?pretty-print=true'
```

You can see the result in JSON format. The result of the above request is:

```json
{
  "id": "1",
  "fields": {
    "category": "Library",
    "description": "Bleve is a full-text search and indexing library for Go.",
    "name": "Bleve",
    "popularity": 3,
    "release": "2014-04-18T00:00:00Z"
  },
  "success": true
}
```

### Deleting a document

Deleting a document via HTTP is as following:

```bash
$ curl -X DELETE 'http://localhost:10002/rest/1?pretty-print=true'
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
  "put_count": 7,
  "success": true
}
```

### Searching documents

Searching documents via HTTP is as following:

```bash
$ curl -X POST 'http://localhost:10002/rest/_search?pretty-print=true' -d @./example/search_request.json
```

You can see the result in JSON format. The result of the above request is:

```json
{
  "search_result": {
    "facets": {
      "Category count": {
        "field": "category",
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "count": 4,
            "term": "library"
          },
          {
            "count": 3,
            "term": "server"
          }
        ],
        "total": 7
      },
      "Popularity range": {
        "field": "popularity",
        "missing": 0,
        "numeric_ranges": [
          {
            "count": 2,
            "max": 4,
            "min": 3,
            "name": "more than or equal to 3 and less than 4"
          },
          {
            "count": 2,
            "min": 5,
            "name": "more than or equal to 5"
          },
          {
            "count": 1,
            "max": 2,
            "min": 1,
            "name": "more than or equal to 1 and less than 2"
          },
          {
            "count": 1,
            "max": 3,
            "min": 2,
            "name": "more than or equal to 2 and less than 3"
          },
          {
            "count": 1,
            "max": 5,
            "min": 4,
            "name": "more than or equal to 4 and less than 5"
          }
        ],
        "other": 0,
        "total": 7
      },
      "Release date range": {
        "date_ranges": [
          {
            "count": 4,
            "end": "2010-12-31T23:59:59Z",
            "name": "2001 - 2010",
            "start": "2001-01-01T00:00:00Z"
          },
          {
            "count": 2,
            "end": "2020-12-31T23:59:59Z",
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z"
          }
        ],
        "field": "release",
        "missing": 0,
        "other": 0,
        "total": 6
      }
    },
    "hits": [
      {
        "fields": {
          "category": "Library",
          "description": "Bleve is a full-text search and indexing library for Go.",
          "name": "Bleve",
          "popularity": 3,
          "release": "2014-04-18T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Bleve is a full-text search and indexing library for Go."
          ],
          "name": [
            "\u003cmark\u003eBleve\u003c/mark\u003e"
          ]
        },
        "id": "1",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "bleve": [
              {
                "array_positions": null,
                "end": 5,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Library",
          "description": "Apache Lucene is a high-performance, full-featured text search engine library written entirely in Java.",
          "name": "Lucene",
          "popularity": 4,
          "release": "2000-03-30T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Apache Lucene is a high-performance, full-featured text search engine library written entirely in Java."
          ],
          "name": [
            "\u003cmark\u003eLucene\u003c/mark\u003e"
          ]
        },
        "id": "2",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "lucene": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Library",
          "description": "Whoosh is a fast, featureful full-text indexing and searching library implemented in pure Python.",
          "name": "Whoosh",
          "popularity": 3,
          "release": "2008-02-20T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Whoosh is a fast, featureful full-text indexing and searching library implemented in pure Python."
          ],
          "name": [
            "\u003cmark\u003eWhoosh\u003c/mark\u003e"
          ]
        },
        "id": "3",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "whoosh": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Library",
          "description": "Ferret is a super fast, highly configurable search library written in Ruby.",
          "name": "Ferret",
          "popularity": 2,
          "release": "2005-10-01T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Ferret is a super fast, highly configurable search library written in Ruby."
          ],
          "name": [
            "\u003cmark\u003eFerret\u003c/mark\u003e"
          ]
        },
        "id": "4",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "ferret": [
              {
                "array_positions": null,
                "end": 6,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Server",
          "description": "Solr is an open source enterprise search platform, written in Java, from the Apache Lucene project.",
          "name": "Solr",
          "popularity": 5,
          "release": "2006-12-22T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Solr is an open source enterprise search platform, written in Java, from the Apache Lucene project."
          ],
          "name": [
            "\u003cmark\u003eSolr\u003c/mark\u003e"
          ]
        },
        "id": "5",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "solr": [
              {
                "array_positions": null,
                "end": 4,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Server",
          "description": "Elasticsearch is a search engine based on Lucene, written in Java.",
          "name": "Elasticsearch",
          "popularity": 5,
          "release": "2010-02-08T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Elasticsearch is a search engine based on Lucene, written in Java."
          ],
          "name": [
            "\u003cmark\u003eElasticsearch\u003c/mark\u003e"
          ]
        },
        "id": "6",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "elasticsearch": [
              {
                "array_positions": null,
                "end": 13,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "category": "Server",
          "description": "Blast is a full-text search and indexing server written in Go, built on top of Bleve.",
          "name": "Blast",
          "popularity": 1,
          "release": "2017-01-13T00:00:00Z"
        },
        "fragments": {
          "description": [
            "Blast is a full-text search and indexing server written in Go, built on top of Bleve."
          ],
          "name": [
            "\u003cmark\u003eBlast\u003c/mark\u003e"
          ]
        },
        "id": "7",
        "index": "/tmp/blast/node0/index",
        "locations": {
          "name": {
            "blast": [
              {
                "array_positions": null,
                "end": 5,
                "pos": 1,
                "start": 0
              }
            ]
          }
        },
        "score": 0.12163776688600772,
        "sort": [
          "_score"
        ]
      }
    ],
    "max_score": 0.12163776688600772,
    "request": {
      "explain": false,
      "facets": {
        "Category count": {
          "field": "category",
          "size": 10
        },
        "Popularity range": {
          "field": "popularity",
          "numeric_ranges": [
            {
              "max": 1,
              "name": "less than 1"
            },
            {
              "max": 2,
              "min": 1,
              "name": "more than or equal to 1 and less than 2"
            },
            {
              "max": 3,
              "min": 2,
              "name": "more than or equal to 2 and less than 3"
            },
            {
              "max": 4,
              "min": 3,
              "name": "more than or equal to 3 and less than 4"
            },
            {
              "max": 5,
              "min": 4,
              "name": "more than or equal to 4 and less than 5"
            },
            {
              "min": 5,
              "name": "more than or equal to 5"
            }
          ],
          "size": 10
        },
        "Release date range": {
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
          "field": "release",
          "size": 10
        }
      },
      "fields": [
        "*"
      ],
      "from": 0,
      "highlight": {
        "fields": [
          "name",
          "description"
        ],
        "style": "html"
      },
      "includeLocations": false,
      "query": {
        "query": "name:*"
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
    "took": 280479,
    "total_hits": 7
  },
  "success": true
}
```

## Bringing up a cluster

Blast is easy to bring up the cluster. Blast node is already running, but that is not fault tolerant. If you need to increase the fault tolerance, bring up 2 more nodes like so:

```bash
$ ./bin/blast start --bind-addr=localhost:11000 --grpc-addr=localhost:11001 --http-addr=localhost:11002 --node-id=node1 --raft-dir=/tmp/blast/node1/raft --store-dir=/tmp/blast/node1/store --index-dir=/tmp/blast/node1/index --peer-grpc-addr=localhost:10001
$ ./bin/blast start --bind-addr=localhost:12000 --grpc-addr=localhost:12001 --http-addr=localhost:12002 --node-id=node2 --raft-dir=/tmp/blast/node2/raft --store-dir=/tmp/blast/node2/store --index-dir=/tmp/blast/node2/index --peer-grpc-addr=localhost:10001
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
      "node_id": "node0"
    },
    {
      "address": "127.0.0.1:11000",
      "metadata": {
        "grpc_address": "localhost:11001",
        "http_address": "localhost:11002"
      },
      "node_id": "node1"
    },
    {
      "address": "127.0.0.1:12000",
      "metadata": {
        "grpc_address": "localhost:12001",
        "http_address": "localhost:12002"
      },
      "node_id": "node2"
    }
  ],
  "success": true
}
```

Recommend 3 or more odd number of nodes in the cluster. In failure scenarios, data loss is inevitable, so avoid deploying single nodes.

This tells each new node to join the existing node. Once joined, each node now knows about the key:

Following command puts a document to node0:

```bash
$ cat ./example/doc1.json | xargs -0 ./bin/blast put --grpc-addr=localhost:10001 --pretty-print 1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "success": true
}
```

So, you can get a document from node0 like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:10001 --pretty-print 1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "1",
  "fields": {
    "category": "Library",
    "description": "Bleve is a full-text search and indexing library for Go.",
    "name": "Bleve",
    "popularity": 3,
    "release": "2014-04-18T00:00:00Z"
  },
  "success": true
}
```

Also, you can get same document from node1 (localhost:11001) like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:11001 --pretty-print 1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "1",
  "fields": {
    "category": "Library",
    "description": "Bleve is a full-text search and indexing library for Go.",
    "name": "Bleve",
    "popularity": 3,
    "release": "2014-04-18T00:00:00Z"
  },
  "success": true
}
```

Lastly, you can get same document from node2 (localhost:12001) like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:12001 --pretty-print 1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "id": "1",
  "fields": {
    "category": "Library",
    "description": "Bleve is a full-text search and indexing library for Go.",
    "name": "Bleve",
    "popularity": 3,
    "release": "2014-04-18T00:00:00Z"
  },
  "success": true
}
```

