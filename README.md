# Blast

The Blast is a full text search and indexing server written in [Go](https://golang.org) that built on top of the [Bleve](http://www.blevesearch.com). Blast server provides functions through [gRPC](http://www.grpc.io) ([HTTP/2](https://en.wikipedia.org/wiki/HTTP/2) + [Protocol Buffers](https://developers.google.com/protocol-buffers/)).  
This repository includes Blast CLI that is a command line interface for controlling Blast server, and Blast REST server that provides a traditional [RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer) API ([HTTP/1.1](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) + [JSON](http://www.json.org)).  
Blast makes it easy for programmers to develop search applications with advanced features.


## Features

- Full-text search and indexing
- Faceting
- Result highlighting
- Text analysis

For more detailed information, refer to the [Bleve document](http://www.blevesearch.com/docs/Home/).

## Blast server

### Blast server config file (blast.yaml)

The Blast server parameters are described in blast.yaml.

```yaml
log_format: text
log_output: ""
log_level: "info"

port: 10000
path: "./data/index"

index_mapping: ""
index_type: "upside_down"
kvstore: "boltdb"
kvconfig: ""

delete_index_at_startup: false
delete_index_at_shutdown: false
```

### Blast server parameters and environment variables

| Parameter name | Environment variable | Command line option | Description |
| --- | --- | --- | --- |
| config                   | BLAST_CONFIG                   | --config                   | config file path. |
| log_format               | BLAST_LOG_FORMAT               | --log-format               | log format. `text`, `color` and `json` are available. Default is `text` |
| log_output               | BLAST_LOG_OUTPUT               | --log-output               | log output path. Default is `stdout` |
| log_level                | BLAST_LOG_LEVEL                | --log-level                | log level. `debug`, `info`, `warn`, `error`, `fatal` and `panic` are available. Default is `info` |
| port                     | BLAST_PORT                     | --port                     | port number. default is `10000` |
| path                     | BLAST_PATH                     | --path                     | index directory path. Default is `/var/blast/data/index` |
| index_mapping            | BLAST_INDEX_MAPPING            | --index-mapping            | index mapping path. Default is `""` |
| index_type               | BLAST_INDEX_TYPE               | --index-type               | index type. `upside_down` is available. Default is `upside_down` |
| kvstore                  | BLAST_KVSTORE                  | --kvstore                  | kvstore. `boltdb`, `goleveldb`, `gtreap` and `moss` are available. Default is `boltdb` |
| kvconfig                 | BLAST_KVCONFIG                 | --kvconfig                 | kvconfig path. Default is `""` |
| delete_index_at_startup  | BLAST_DELETE_INDEX_AT_STARTUP  | --delete-index-at-startup  | delete index at startup. Default is `false` |
| delete_index_at_shutdown | BLAST_DELETE_INDEX_AT_SHUTDOWN | --delete-index-at-shutdown | delete index at shutdown. Default is `false` |


### Index Mapping

You can specify the index mapping describes how to your data model should be indexed. it contains all of the details about which fields your documents can contain, and how those fields should be dealt with when adding documents to the index, or when querying those fields. The example is following:

```json
{
  "types": {
    "document": {
      "enabled": true,
      "dynamic": true,
      "properties": {
        "category": {
          "enabled": true,
          "dynamic": true,
          "fields": [
            {
              "type": "text",
              "analyzer": "keyword",
              "store": true,
              "index": true,
              "include_term_vectors": true,
              "include_in_all": true
            }
          ],
          "default_analyzer": ""
        },
        "description": {
          "enabled": true,
          "dynamic": true,
          "fields": [
            {
              "type": "text",
              "analyzer": "en",
              "store": true,
              "index": true,
              "include_term_vectors": true,
              "include_in_all": true
            }
          ],
          "default_analyzer": ""
        },
        "name": {
          "enabled": true,
          "dynamic": true,
          "fields": [
            {
              "type": "text",
              "analyzer": "en",
              "store": true,
              "index": true,
              "include_term_vectors": true,
              "include_in_all": true
            }
          ],
          "default_analyzer": ""
        },
        "popularity": {
          "enabled": true,
          "dynamic": true,
          "fields": [
            {
              "type": "number",
              "store": true,
              "index": true,
              "include_in_all": true
            }
          ],
          "default_analyzer": ""
        },
        "release": {
          "enabled": true,
          "dynamic": true,
          "fields": [
            {
              "type": "datetime",
              "store": true,
              "index": true,
              "include_in_all": true
            }
          ],
          "default_analyzer": ""
        },
        "type": {
          "enabled": true,
          "dynamic": true,
          "fields": [
            {
              "type": "text",
              "analyzer": "keyword",
              "store": true,
              "index": true,
              "include_term_vectors": true,
              "include_in_all": true
            }
          ],
          "default_analyzer": ""
        }
      },
      "default_analyzer": ""
    }
  },
  "default_mapping": {
    "enabled": true,
    "dynamic": true,
    "default_analyzer": ""
  },
  "type_field": "type",
  "default_type": "document",
  "default_analyzer": "standard",
  "default_datetime_parser": "dateTimeOptional",
  "default_field": "_all",
  "store_dynamic": true,
  "index_dynamic": true,
  "analysis": {}
}
```

See [Introduction to Index Mappings](http://www.blevesearch.com/docs/Index-Mapping/) and [type IndexMappingImpl](https://godoc.org/github.com/blevesearch/bleve/mapping#IndexMappingImpl) for more details.  


### Start Blast server

The `blast` command starts Blast server. You can display a help message by specifying `-h` or `--help` option.

```sh
$ ./bin/blast --config ./example/blast.yaml --index-mapping ./example/index_mapping.json
```


## Blast CLI

### Get the index information from Blast server

The `get index` command retrieves an index information about existing opened index. You can display a help message by specifying the `- h` or` --help` option.

```sh
$ ./bin/blastcli get index --include-index-mapping --include-index-type --include-kvstore --include-kvconfig
```

The result of the above `get index` command is:

```json
{
  "path": "/var/blast/data/index",
  "index_mapping": {
    "types": {
      "document": {
        "enabled": true,
        "dynamic": true,
        "properties": {
          "category": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "keyword",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "description": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "en",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "name": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "en",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "popularity": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "number",
                "store": true,
                "index": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "release": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "datetime",
                "store": true,
                "index": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "type": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "keyword",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          }
        },
        "default_analyzer": ""
      }
    },
    "default_mapping": {
      "enabled": true,
      "dynamic": true,
      "default_analyzer": ""
    },
    "type_field": "type",
    "default_type": "document",
    "default_analyzer": "standard",
    "default_datetime_parser": "dateTimeOptional",
    "default_field": "_all",
    "store_dynamic": true,
    "index_dynamic": true,
    "analysis": {}
  },
  "index_type": "upside_down",
  "kvstore": "boltdb",
  "kvconfig": {}
}
```

### Put document format

```json
{
  "document": {
    "id": "1",
    "fields": {
      "name": "Bleve",
      "description": "Bleve is a full-text search and indexing library for Go.",
      "category": "Library",
      "popularity": 3.0,
      "release": "2014-04-18T00:00:00Z",
      "type": "document"
    }
  }
}
```


### Put the document to Blast server

The `put document` command adds or updates a JSON formatted document in a specified index. You can display a help message by specifying the `- h` or` --help` option.  
The document example is following:

```sh
$ ./bin/blastcli put document --resource ./example/document_1.json
```

The result of the above `put document` command is:

```json
{
  "put_count": 1
}
```


### Get the document from Blast server

The `get document` command retrieves a JSON formatted document on its id from a specified index. You can display a help message by specifying the `- h` or` --help` option.

```sh
$ ./bin/blastcli get document --id 1
```

The result of the above `get document` command is:

```json
{
  "document": {
    "id": "1",
    "fields": {
      "name": "Bleve",
      "description": "Bleve is a full-text search and indexing library for Go.",
      "category": "Library",
      "popularity": 3.0,
      "release": "2014-04-18T00:00:00Z",
      "type": "document"
    }
  }
}
```


### Delete the document from Blast server

The `delete document` command deletes a document on its id from a specified index. You can display a help message by specifying the `- h` or` --help` option.

```sh
$ ./bin/blastcli delete document --id 1
```

The result of the above `delete document` command is:

```json
{
  "delete_count": 1
}
```


### Bulk document format

```json
{
  "batch_size": 1000,
  "requests": [
    {
      "method": "put",
      "document": {
        "id": "1",
        "fields": {
          "name": "Bleve",
          "description": "Bleve is a full-text search and indexing library for Go.",
          "category": "Library",
          "popularity": 3,
          "release": "2014-04-18T00:00:00Z",
          "type": "document"
        }
      }
    },
    {
      "method": "delete",
      "document": {
        "id": "2"
      }
    },
    {
      "method": "delete",
      "document": {
        "id": "3"
      }
    },
    {
      "method": "delete",
      "document": {
        "id": "4"
      }
    },
    {
      "method": "delete",
      "document": {
        "id": "5"
      }
    },
    {
      "method": "delete",
      "document": {
        "id": "6"
      }
    },
    {
      "method": "put",
      "document": {
        "id": "7",
        "fields": {
          "name": "Blast",
          "description": "Blast is a full-text search and indexing server written in Go, built on top of Bleve.",
          "category": "Server",
          "popularity": 1,
          "release": "2017-01-13T00:00:00Z",
          "type": "document"
        }
      }
    }
  ]
}
```

### Index the documents in bulk to Blast server

The `bulk` command makes it possible to perform many put/delete operations in a single command execution. This can greatly increase the indexing speed. You can display a help message by specifying the `- h` or` --help` option.
The bulk example is following:

```sh
$ ./bin/blastcli bulk --resource ./example/bulk_put.json
```

The result of the above `bulk` command is:

```json
{
  "put_count": 7
}
```

### Search request format

```json
{
  "search_request": {
    "query": {
      "query": "name:*"
    },
    "size": 10,
    "from": 0,
    "fields": [
      "*"
    ],
    "sort": [
      "-_score"
    ],
    "facets": {
      "Category count": {
        "size": 10,
        "field": "category"
      },
      "Popularity range": {
        "size": 10,
        "field": "popularity",
        "numeric_ranges": [
          {
            "name": "less than 1",
            "max": 1
          },
          {
            "name": "more than or equal to 1 and less than 2",
            "min": 1,
            "max": 2
          },
          {
            "name": "more than or equal to 2 and less than 3",
            "min": 2,
            "max": 3
          },
          {
            "name": "more than or equal to 3 and less than 4",
            "min": 3,
            "max": 4
          },
          {
            "name": "more than or equal to 4 and less than 5",
            "min": 4,
            "max": 5
          },
          {
            "name": "more than or equal to 5",
            "min": 5
          }
        ]
      },
      "Release date range": {
        "size": 10,
        "field": "release",
        "date_ranges": [
          {
            "name": "2001 - 2010",
            "start": "2001-01-01T00:00:00Z",
            "end": "2010-12-31T23:59:59Z"
          },
          {
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z",
            "end": "2020-12-31T23:59:59Z"
          }
        ]
      }
    },
    "highlight": {
      "style": "html",
      "fields": [
        "name",
        "description"
      ]
    }
  }
}
```

See [Queries](http://www.blevesearch.com/docs/Query/), [Query String Query](http://www.blevesearch.com/docs/Query-String-Query/) and [type SearchRequest](https://godoc.org/github.com/blevesearch/bleve#SearchRequest) for more details.


### Search the documents from Blast server

The `search` command can be executed with a search request, which includes the Query, within its file. Here is an example:
You can display a help message by specifying the `- h` or` --help` option.

```sh
$ ./bin/blastcli search --resource ./example/search_request.json
```

The result of the above `search` command is:

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
            "term": "Library"
          },
          {
            "count": 3,
            "term": "Server"
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
          "release": "2014-04-18T00:00:00Z",
          "type": "document"
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
        "index": "/var/blast/data/index",
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
          "release": "2000-03-30T00:00:00Z",
          "type": "document"
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
        "index": "/var/blast/data/index",
        "locations": {
          "name": {
            "lucen": [
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
          "release": "2008-02-20T00:00:00Z",
          "type": "document"
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
        "index": "/var/blast/data/index",
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
          "release": "2005-10-01T00:00:00Z",
          "type": "document"
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
        "index": "/var/blast/data/index",
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
          "release": "2006-12-22T00:00:00Z",
          "type": "document"
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
        "index": "/var/blast/data/index",
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
          "release": "2010-02-08T00:00:00Z",
          "type": "document"
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
        "index": "/var/blast/data/index",
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
          "release": "2017-01-13T00:00:00Z",
          "type": "document"
        },
        "fragments": {
          "description": [
            "Blast is a full-text search and indexing server written in Go, built on top of Bleve."
          ],
          "name": [
            "\u003cmark\u003eBleve\u003c/mark\u003e"
          ]
        },
        "id": "7",
        "index": "/var/blast/data/index",
        "locations": {
          "name": {
            "blast": [
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
    "took": 4103742,
    "total_hits": 7
  }
}
```

See [type SearchResult](https://godoc.org/github.com/blevesearch/bleve#SearchResult) for more details.



## Blast Rest server

### Blast REST server config file (blastrest.yaml)

The Blast REST server parameters are described in blastrest.yaml.

```yaml
log_format: text
log_output: ""
log_level: "info"

port: 20000
base_uri: "api"
server: "localhost:10000"
```

### Blast REST server parameters and environment variables

| Parameter name | Environment variable | Command line option | Description |
| --- | --- | --- | --- |
| log_format | BLASTREST_LOG_FORMAT | --log-format | log format. `text`, `color` and `json` are available. Default is `text` |
| log_output | BLASTREST_LOG_OUTPUT | --log-output | log output path. Default is `stdout` |
| log_level  | BLASTREST_LOG_LEVEL  | --log-level  | log level. `debug`, `info`, `warn`, `error`, `fatal` and `panic` are available. Default is `info` |
| port       | BLASTREST_PORT       | --port       | port number. default is `1289` |
| base_uri   | BLASTREST_BASE_URI   | --base-uri   | index directory path. Default is `/var/blast/data/index` |
| server     | BLASTREST_SERVER     | --server     | index mapping path. Default is `""` |


### Start Blast REST server

the `blastrest` command starts Blast REST server. You can display a help message by specifying `-h` or `--help` option.

```sh
$ blastret
```

### Get the index information from Blast server via Blast REST server

The get index API retrieves an index information about index.

#### Endpoint URL:

```
http://<hostname>:<port>/<base_uri>/?<query_parameter>...
```

#### Query Parameters:

| Parameter | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| includeIndexMapping | boolean | No | If true, index mapping will returned with result. |
| includeIndexType    | boolean | No | If true, index type will returned with result. |
| includeKvstore      | boolean | No | If true, kvstore will returned with result. |
| includeKvconfig     | boolean | No | If true, kvconfig will returned with result. |

```sh
$ curl -s -X GET "http://localhost:2289/api/?includeIndexMapping=true&includeIndexType=true&includeKvstore=true&includeKvconfig=true"
```

The result of the above command is:

```json
{
  "path": "/var/blast/data/index",
  "index_mapping": {
    "types": {
      "document": {
        "enabled": true,
        "dynamic": true,
        "properties": {
          "category": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "keyword",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "description": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "en",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "name": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "en",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "popularity": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "number",
                "store": true,
                "index": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "release": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "datetime",
                "store": true,
                "index": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          },
          "type": {
            "enabled": true,
            "dynamic": true,
            "fields": [
              {
                "type": "text",
                "analyzer": "keyword",
                "store": true,
                "index": true,
                "include_term_vectors": true,
                "include_in_all": true
              }
            ],
            "default_analyzer": ""
          }
        },
        "default_analyzer": ""
      }
    },
    "default_mapping": {
      "enabled": true,
      "dynamic": true,
      "default_analyzer": ""
    },
    "type_field": "type",
    "default_type": "document",
    "default_analyzer": "standard",
    "default_datetime_parser": "dateTimeOptional",
    "default_field": "_all",
    "store_dynamic": true,
    "index_dynamic": true,
    "analysis": {}
  },
  "index_type": "upside_down",
  "kvstore": "boltdb",
  "kvconfig": {
    "create_if_missing": true,
    "error_if_exists": true,
    "path": "/var/blast/data/index/store"
  }
}
```

### Put the document to the Blast Server via the Blast REST Server

The put document API adds or updates a JSON formatted document in an index.

#### Endpoint URL:

```
http://<hostname>:<port>/<base_uri>/<id>?<query_parameters>...
```

#### Path Parameters:

| Parameter | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| id | string | Yes | Document ID. |

```sh
$ curl -s -X PUT -H "Content-Type: application/json" --data-binary @example/document_1.json "http://localhost:2289/api/1"
```

The result of the above command is:

```json
{
  "put_count": 1
}
```

### Get the document from the Blast Server via the Blast REST Server

The get document API retrieves a JSON formatted document on its id from an index.

#### Endpoint URL:

```
http://<hostname>:<port>/<base_uri>/<id>?<query_parameters>...
```

#### Path Parameters:

| Parameter | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| id | string | Yes | Document ID. |

```sh
$ curl -s -X GET "http://localhost:2289/api/1"
```

The result of the above command is:

```json
{
  "document": {
    "id": "1",
    "fields": {
      "name": "Bleve",
      "description": "Bleve is a full-text search and indexing library for Go.",
      "category": "Library",
      "popularity": 3.0,
      "release": "2014-04-18T00:00:00Z",
      "type": "document"
    }
  }
}
```

### Delete the document from the Blast Server via the Blast REST Server

The delete document API deletes a document on its id from an index.

#### Endpoint URL:

```
http://<hostname>:<port>/<base_uri>/<id>?<query_parameters>...
```

#### Path Parameters:

| Parameter | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| id | string | Yes | Document ID. |

```sh
$ curl -s -X DELETE "http://localhost:2289/api/1"
```

The result of the above command is:

```json
{
  "delete_count": 1
}
```

### Index the documents in bulk to the Blast Server via the Blast REST Server

The bulk API makes it possible to perform many put/delete operations in a single command execution. This can greatly increase the indexing speed.

#### Endpoint URL:

```
http://<hostname>:<port>/<base_uri>/_bulk?<query_parameters>...
```

#### Query Parameters:

| Parameter | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| batchSize | integer | No | Batch size of bulk request. Default is 1000. |

```sh
$ curl -s -X POST -H "Content-Type: application/json" --data-binary @example/bulk_put.json "http://localhost:2289/api/_bulk"
```

The result of the above command is:

```text
{
  "put_count": 7
}
```

### Search the documents from the Blast Server via the blast REST Server

The search API can be executed with a search request, which includes the Query, within its file.

#### Endpoint URL:

```
http://<hostname>:<port>/<base_uri>/_search?<query_parameters>...
```

#### Query Parameters:

| Parameter | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| query            | string   | No | Query string. |
| size             | integer  | No | Number of hits to return. Default is `10`. |
| from             | integer  | No | Starting from index of the hits to return. Default is `0`. |
| explain          | boolean  | No | Contain an explanation of how scoring of the hits was computed. |
| fields           | string   | No | Specify a set of fields to return. Example `name,description` |
| sort             | string   | No | Sorting to perform. Example `-_score,name` |
| facets           | string   | No | Faceting to perform. Example `{"cat_count":{"size":10,"field":"category"}` |
| highlight        | string   | No | Highlighting to perform. Example `{"style":"html","fields":["name","description"]}` |
| highlightStyle   | string   | No | Highlighting style. Default is `html` |
| highlightField   | string   | No | Specify a set of fields to highlight. |
| includeLocations | boolean  | No | Include terms locations. |

```sh
$ curl -s -X POST -H "Content-Type: application/json" --data-binary @example/search_request.json "http://localhost:20000/api/_search"
```

The result of the above command is:

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
        "query": "name:*"
      },
      "size": 10,
      "from": 0,
      "highlight": {
        "style": "html",
        "fields": [
          "name",
          "description"
        ]
      },
      "fields": [
        "*"
      ],
      "facets": {
        "Category count": {
          "size": 10,
          "field": "category"
        },
        "Popularity range": {
          "size": 10,
          "field": "popularity",
          "numeric_ranges": [
            {
              "name": "less than 1",
              "max": 1
            },
            {
              "name": "more than or equal to 1 and less than 2",
              "min": 1,
              "max": 2
            },
            {
              "name": "more than or equal to 2 and less than 3",
              "min": 2,
              "max": 3
            },
            {
              "name": "more than or equal to 3 and less than 4",
              "min": 3,
              "max": 4
            },
            {
              "name": "more than or equal to 4 and less than 5",
              "min": 4,
              "max": 5
            },
            {
              "name": "more than or equal to 5",
              "min": 5
            }
          ]
        },
        "Release date range": {
          "size": 10,
          "field": "release",
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
        "index": "/var/blast/data/index",
        "id": "1",
        "score": 0.12163776688600772,
        "locations": {
          "name": {
            "bleve": [
              {
                "pos": 1,
                "start": 0,
                "end": 5,
                "array_positions": null
              }
            ]
          }
        },
        "fragments": {
          "description": [
            "Bleve is a full-text search and indexing library for Go."
          ],
          "name": [
            "\u003cmark\u003eBleve\u003c/mark\u003e"
          ]
        },
        "sort": [
          "_score"
        ],
        "fields": {
          "category": "Library",
          "description": "Bleve is a full-text search and indexing library for Go.",
          "name": "Bleve",
          "popularity": 3,
          "release": "2014-04-18T00:00:00Z",
          "type": "document"
        }
      },
      {
        "index": "/var/blast/data/index",
        "id": "2",
        "score": 0.12163776688600772,
        "locations": {
          "name": {
            "lucen": [
              {
                "pos": 1,
                "start": 0,
                "end": 6,
                "array_positions": null
              }
            ]
          }
        },
        "fragments": {
          "description": [
            "Apache Lucene is a high-performance, full-featured text search engine library written entirely in Java."
          ],
          "name": [
            "\u003cmark\u003eLucene\u003c/mark\u003e"
          ]
        },
        "sort": [
          "_score"
        ],
        "fields": {
          "category": "Library",
          "description": "Apache Lucene is a high-performance, full-featured text search engine library written entirely in Java.",
          "name": "Lucene",
          "popularity": 4,
          "release": "2000-03-30T00:00:00Z",
          "type": "document"
        }
      },
      {
        "index": "/var/blast/data/index",
        "id": "3",
        "score": 0.12163776688600772,
        "locations": {
          "name": {
            "whoosh": [
              {
                "pos": 1,
                "start": 0,
                "end": 6,
                "array_positions": null
              }
            ]
          }
        },
        "fragments": {
          "description": [
            "Whoosh is a fast, featureful full-text indexing and searching library implemented in pure Python."
          ],
          "name": [
            "\u003cmark\u003eWhoosh\u003c/mark\u003e"
          ]
        },
        "sort": [
          "_score"
        ],
        "fields": {
          "category": "Library",
          "description": "Whoosh is a fast, featureful full-text indexing and searching library implemented in pure Python.",
          "name": "Whoosh",
          "popularity": 3,
          "release": "2008-02-20T00:00:00Z",
          "type": "document"
        }
      },
      {
        "index": "/var/blast/data/index",
        "id": "4",
        "score": 0.12163776688600772,
        "locations": {
          "name": {
            "ferret": [
              {
                "pos": 1,
                "start": 0,
                "end": 6,
                "array_positions": null
              }
            ]
          }
        },
        "fragments": {
          "description": [
            "Ferret is a super fast, highly configurable search library written in Ruby."
          ],
          "name": [
            "\u003cmark\u003eFerret\u003c/mark\u003e"
          ]
        },
        "sort": [
          "_score"
        ],
        "fields": {
          "category": "Library",
          "description": "Ferret is a super fast, highly configurable search library written in Ruby.",
          "name": "Ferret",
          "popularity": 2,
          "release": "2005-10-01T00:00:00Z",
          "type": "document"
        }
      },
      {
        "index": "/var/blast/data/index",
        "id": "5",
        "score": 0.12163776688600772,
        "locations": {
          "name": {
            "solr": [
              {
                "pos": 1,
                "start": 0,
                "end": 4,
                "array_positions": null
              }
            ]
          }
        },
        "fragments": {
          "description": [
            "Solr is an open source enterprise search platform, written in Java, from the Apache Lucene project."
          ],
          "name": [
            "\u003cmark\u003eSolr\u003c/mark\u003e"
          ]
        },
        "sort": [
          "_score"
        ],
        "fields": {
          "category": "Server",
          "description": "Solr is an open source enterprise search platform, written in Java, from the Apache Lucene project.",
          "name": "Solr",
          "popularity": 5,
          "release": "2006-12-22T00:00:00Z",
          "type": "document"
        }
      },
      {
        "index": "/var/blast/data/index",
        "id": "6",
        "score": 0.12163776688600772,
        "locations": {
          "name": {
            "elasticsearch": [
              {
                "pos": 1,
                "start": 0,
                "end": 13,
                "array_positions": null
              }
            ]
          }
        },
        "fragments": {
          "description": [
            "Elasticsearch is a search engine based on Lucene, written in Java."
          ],
          "name": [
            "\u003cmark\u003eElasticsearch\u003c/mark\u003e"
          ]
        },
        "sort": [
          "_score"
        ],
        "fields": {
          "category": "Server",
          "description": "Elasticsearch is a search engine based on Lucene, written in Java.",
          "name": "Elasticsearch",
          "popularity": 5,
          "release": "2010-02-08T00:00:00Z",
          "type": "document"
        }
      },
      {
        "index": "/var/blast/data/index",
        "id": "7",
        "score": 0.12163776688600772,
        "locations": {
          "name": {
            "blast": [
              {
                "pos": 1,
                "start": 0,
                "end": 6,
                "array_positions": null
              }
            ]
          }
        },
        "fragments": {
          "description": [
            "Blast is a full-text search and indexing server written in Go, built on top of Bleve."
          ],
          "name": [
            "\u003cmark\u003eBlast\u003c/mark\u003e"
          ]
        },
        "sort": [
          "_score"
        ],
        "fields": {
          "category": "Server",
          "description": "Blast is a full-text search and indexing server written in Go, built on top of Bleve.",
          "name": "Blast",
          "popularity": 1,
          "release": "2017-01-13T00:00:00Z",
          "type": "document"
        }
      }
    ],
    "total_hits": 7,
    "max_score": 0.12163776688600772,
    "took": 786061,
    "facets": {
      "Category count": {
        "field": "category",
        "total": 7,
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "term": "Library",
            "count": 4
          },
          {
            "term": "Server",
            "count": 3
          }
        ]
      },
      "Popularity range": {
        "field": "popularity",
        "total": 7,
        "missing": 0,
        "other": 0,
        "numeric_ranges": [
          {
            "name": "more than or equal to 3 and less than 4",
            "min": 3,
            "max": 4,
            "count": 2
          },
          {
            "name": "more than or equal to 5",
            "min": 5,
            "count": 2
          },
          {
            "name": "more than or equal to 1 and less than 2",
            "min": 1,
            "max": 2,
            "count": 1
          },
          {
            "name": "more than or equal to 2 and less than 3",
            "min": 2,
            "max": 3,
            "count": 1
          },
          {
            "name": "more than or equal to 4 and less than 5",
            "min": 4,
            "max": 5,
            "count": 1
          }
        ]
      },
      "Release date range": {
        "field": "release",
        "total": 6,
        "missing": 0,
        "other": 0,
        "date_ranges": [
          {
            "name": "2001 - 2010",
            "start": "2001-01-01T00:00:00Z",
            "end": "2010-12-31T23:59:59Z",
            "count": 4
          },
          {
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z",
            "end": "2020-12-31T23:59:59Z",
            "count": 2
          }
        ]
      }
    }
  }
}
```


## License

Apache License Version 2.0
