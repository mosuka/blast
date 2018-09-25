<!--
 Copyright (c) 2018 Minoru Osuka

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

Blast
======

Blast is a full text search and indexing server written in [Go](https://golang.org) built on top of the [Bleve](http://www.blevesearch.com). It provides functions through [gRPC](http://www.grpc.io) ([HTTP/2](https://en.wikipedia.org/wiki/HTTP/2) + [Protocol Buffers](https://developers.google.com/protocol-buffers/)) or traditional [RESTful](https://en.wikipedia.org/wiki/Representational_state_transfer) API ([HTTP/1.1](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) + [JSON](http://www.json.org)).  
Blast uses [Raft](https://en.wikipedia.org/wiki/Raft_(computer_science)) consensus algorithm to achieve consensus across all the instances of the nodes, ensuring that every change made to the system is made to a quorum of nodes, or none at all.  
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
$ sudo apt-get install -y \
    libicu-dev \
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
$ brew install \
    icu4c \
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
blastd
```

## Running Blast data node

Running a Blast data node is easy. Start Blast data node like so:

```bash
$ ./bin/blastd data \
    --raft-addr=127.0.0.1:10000 \
    --grpc-addr=127.0.0.1:10001 \
    --http-addr=127.0.0.1:10002 \
    --raft-node-id=node1 \
    --raft-dir=/tmp/blast/node1/raft \
    --store-dir=/tmp/blast/node1/store \
    --index-dir=/tmp/blast/node1/index \
    --index-mapping-file=./etc/index_mapping.json
```

## Using Blast CLI

You can now put, get, search and delete the document(s) via CLI.

### Putting a document

Putting a document is as following:

```bash
$ cat ./example/doc_enwiki_1.json | xargs -0 ./bin/blast put --grpc-addr=127.0.0.1:10001 enwiki_1
```

### Getting a document

Getting a document is as following:

```bash
$ ./bin/blast get --grpc-addr=127.0.0.1:10001 enwiki_1
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

### Deleting a document

Deleting a document is as following:

```bash
$ ./bin/blast delete --grpc-addr=127.0.0.1:10001 enwiki_1
```

### Bulk update

Bulk update is as following:

```bash
$ cat ./example/bulk_put_request.json | xargs -0 ./bin/blast bulk --grpc-addr=127.0.0.1:10001
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "put_count": 36
}
```

### Searching documents

Searching documents is as like following:

```bash
$ cat ./example/search_request.json | xargs -0 ./bin/blast search --grpc-addr=127.0.0.1:10001
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
      "index": "/tmp/blast/node1/index",
      "id": "enwiki_1",
      "score": 0.633816551223398,
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
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "arwiki_1",
      "score": 0.2584513642405507,
      "locations": {
        "text_ar": {
          "search": [
            {
              "pos": 4,
              "start": 45,
              "end": 51,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "arwiki",
        "contributor": "unknown",
        "text_ar": "محرك البحث (بالإنجليزية: Search engine) هو نظام لإسترجاع المعلومات صمم للمساعدة على البحث عن المعلومات المخزنة على أي نظام حاسوبي. تعرض نتائج البحث عادة على شكل قائمة لأماكن تواجد المعلومات ومرتبة وفق معايير معينة. تسمح محركات البحث باختصار مدة البحث والتغلب على مشكلة أحجام البيانات المتصاعدة (إغراق معلوماتي).",
        "timestamp": "2018-03-25T18:04:00Z",
        "title_ar": "محرك بحث"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "hiwiki_1",
      "score": 0.2315458066284217,
      "locations": {
        "text_hi": {
          "search": [
            {
              "pos": 6,
              "start": 93,
              "end": 99,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "hiwiki",
        "contributor": "unknown",
        "text_hi": "ऐसे कम्प्यूटर प्रोग्राम खोजी इंजन (search engine) कहलाते हैं जो किसी कम्प्यूटर सिस्टम पर भण्डारित सूचना में से वांछित सूचना को ढूढ निकालते हैं। ये इंजन प्राप्त परिणामों को प्रायः एक सूची के रूप में प्रस्तुत करते हैं जिससे वांछित सूचना की प्रकृति और उसकी स्थिति का पता चलता है। खोजी इंजन किसी सूचना तक अपेक्षाकृत बहुत कम समय में पहुँचने में हमारी सहायता करते हैं। वे 'सूचना ओवरलोड' से भी हमे बचाते हैं। खोजी इंजन का सबसे प्रचलित रूप 'वेब खोजी इंजन' है जो वर्ल्ड वाइड वेब पर सूचना खोजने के लिये प्रयुक्त होता है।",
        "timestamp": "2017-10-19T20:09:00Z",
        "title_hi": "खोज इंजन"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "itwiki_1",
      "score": 0.22855799635019447,
      "locations": {
        "text_it": {
          "search": [
            {
              "pos": 12,
              "start": 75,
              "end": 81,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "itwiki",
        "contributor": "unknown",
        "text_it": "Nell'ambito delle tecnologie di Internet, un motore di ricerca (in inglese search engine) è un sistema automatico che, su richiesta, analizza un insieme di dati (spesso da esso stesso raccolti) e restituisce un indice dei contenuti disponibili[1] classificandoli in modo automatico in base a formule statistico-matematiche che ne indichino il grado di rilevanza data una determinata chiave di ricerca. Uno dei campi in cui i motori di ricerca trovano maggiore utilizzo è quello dell'information retrieval e nel web. I motori di ricerca più utilizzati nel 2017 sono stati: Google, Bing, Baidu, Qwant, Yandex, Ecosia, DuckDuckGo.",
        "timestamp": "2018-07-16T12:20:00Z",
        "title_it": "Motore di ricerca"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "thwiki_1",
      "score": 0.2018569611073604,
      "locations": {
        "text_th": {
          "search": [
            {
              "pos": 4,
              "start": 38,
              "end": 44,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "thwiki",
        "contributor": "unknown",
        "text_th": "เสิร์ชเอนจิน (search engine) หรือ โปรแกรมค้นหา คือ โปรแกรมที่ช่วยในการสืบค้นหาข้อมูล โดยเฉพาะข้อมูลบนอินเทอร์เน็ต โดยครอบคลุมทั้งข้อความ รูปภาพ ภาพเคลื่อนไหว เพลง ซอฟต์แวร์ แผนที่ ข้อมูลบุคคล กลุ่มข่าว และอื่น ๆ ซึ่งแตกต่างกันไปแล้วแต่โปรแกรมหรือผู้ให้บริการแต่ละราย. เสิร์ชเอนจินส่วนใหญ่จะค้นหาข้อมูลจากคำสำคัญ (คีย์เวิร์ด) ที่ผู้ใช้ป้อนเข้าไป จากนั้นก็จะแสดงรายการผลลัพธ์ที่มันคิดว่าผู้ใช้น่าจะต้องการขึ้นมา ในปัจจุบัน เสิร์ชเอนจินบางตัว เช่น กูเกิล จะบันทึกประวัติการค้นหาและการเลือกผลลัพธ์ของผู้ใช้ไว้ด้วย และจะนำประวัติที่บันทึกไว้นั้น มาช่วยกรองผลลัพธ์ในการค้นหาครั้งต่อ ๆ ไป",
        "timestamp": "2016-06-18T11:06:00Z",
        "title_th": "เสิร์ชเอนจิน"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "zhwiki_1",
      "score": 0.19986816567601795,
      "locations": {
        "text_zh": {
          "search": [
            {
              "pos": 5,
              "start": 24,
              "end": 30,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "zhwiki",
        "contributor": "unknown",
        "text_zh": "搜索引擎（英语：search engine）是一种信息检索系统，旨在协助搜索存储在计算机系统中的信息。搜索结果一般被称为“hits”，通常会以表单的形式列出。网络搜索引擎是最常见、公开的一种搜索引擎，其功能为搜索万维网上储存的信息.",
        "timestamp": "2018-08-27T05:47:00Z",
        "title_zh": "搜索引擎"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "bgwiki_1",
      "score": 0.18275270089950202,
      "locations": {
        "text_bg": {
          "search": [
            {
              "pos": 8,
              "start": 82,
              "end": 88,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "bgwiki",
        "contributor": "unknown",
        "text_bg": "Търсачка или търсеща машина (на английски: Web search engine) е специализиран софтуер за извличане на информация, съхранена в компютърна система или мрежа. Това може да е персонален компютър, Интернет, корпоративна мрежа и т.н. Без допълнителни уточнения, най-често под търсачка се разбира уеб(-)търсачка, която търси в Интернет. Други видове търсачки са корпоративните търсачки, които търсят в интранет мрежите, личните търсачки – за индивидуалните компютри и мобилните търсачки. В търсачката потребителят (търсещият) прави запитване за съдържание, отговарящо на определен критерий (обикновено такъв, който съдържа определени думи и фрази). В резултат се получават списък от точки, които отговарят, пълно или частично, на този критерий. Търсачките обикновено използват редовно подновявани индекси, за да оперират бързо и ефикасно. Някои търсачки също търсят в информацията, която е на разположение в нюзгрупите и други големи бази данни. За разлика от Уеб директориите, които се поддържат от хора редактори, търсачките оперират алгоритмично. Повечето Интернет търсачки са притежавани от различни корпорации.",
        "timestamp": "2018-07-11T11:03:00Z",
        "title_bg": "Търсачка"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "idwiki_1",
      "score": 0.17060026115019133,
      "locations": {
        "text_id": {
          "search": [
            {
              "pos": 11,
              "start": 62,
              "end": 68,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "idwiki",
        "contributor": "unknown",
        "text_id": "Mesin pencari web atau mesin telusur web (bahasa Inggris: web search engine) adalah program komputer yang dirancang untuk melakukan pencarian atas berkas-berkas yang tersimpan dalam layanan www, ftp, publikasi milis, ataupun news group dalam sebuah ataupun sejumlah komputer peladen dalam suatu jaringan. Mesin pencari merupakan perangkat penelusur informasi dari dokumen-dokumen yang tersedia. Hasil pencarian umumnya ditampilkan dalam bentuk daftar yang seringkali diurutkan menurut tingkat akurasi ataupun rasio pengunjung atas suatu berkas yang disebut sebagai hits. Informasi yang menjadi target pencarian bisa terdapat dalam berbagai macam jenis berkas seperti halaman situs web, gambar, ataupun jenis-jenis berkas lainnya. Beberapa mesin pencari juga diketahui melakukan pengumpulan informasi atas data yang tersimpan dalam suatu basis data ataupun direktori web. Sebagian besar mesin pencari dijalankan oleh perusahaan swasta yang menggunakan algoritme kepemilikan dan basis data tertutup, di antaranya yang paling populer adalah safari Google (MSN Search dan Yahoo!). Telah ada beberapa upaya menciptakan mesin pencari dengan sumber terbuka (open source), contohnya adalah Htdig, Nutch, Egothor dan OpenFTS.",
        "timestamp": "2017-11-20T17:47:00Z",
        "title_id": "Mesin pencari web"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "ptwiki_1",
      "score": 0.1688012644026126,
      "locations": {
        "text_pt": {
          "search": [
            {
              "pos": 16,
              "start": 111,
              "end": 117,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "ptwiki",
        "contributor": "unknown",
        "text_pt": "Motor de pesquisa (português europeu) ou ferramenta de busca (português brasileiro) ou buscador (em inglês: search engine) é um programa desenhado para procurar palavras-chave fornecidas pelo utilizador em documentos e bases de dados. No contexto da internet, um motor de pesquisa permite procurar palavras-chave em documentos alojados na world wide web, como aqueles que se encontram armazenados em websites. Os motores de busca surgiram logo após o aparecimento da Internet, com a intenção de prestar um serviço extremamente importante: a busca de qualquer informação na rede, apresentando os resultados de uma forma organizada, e também com a proposta de fazer isto de uma maneira rápida e eficiente. A partir deste preceito básico, diversas empresas se desenvolveram, chegando algumas a valer milhões de dólares. Entre as maiores empresas encontram-se o Google, o Yahoo, o Bing, o Lycos, o Cadê e, mais recentemente, a Amazon.com com o seu mecanismo de busca A9 porém inativo. Os buscadores se mostraram imprescindíveis para o fluxo de acesso e a conquista novos visitantes. Antes do advento da Web, havia sistemas para outros protocolos ou usos, como o Archie para sites FTP anônimos e o Veronica para o Gopher (protocolo de redes de computadores que foi desenhado para indexar repositórios de documentos na Internet, baseado-se em menus).",
        "timestamp": "2017-11-09T14:38:00Z",
        "title_pt": "Motor de busca"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "frwiki_1",
      "score": 0.14418354794511895,
      "locations": {
        "text_fr": {
          "search": [
            {
              "pos": 253,
              "start": 1656,
              "end": 1662,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "frwiki",
        "contributor": "unknown",
        "text_fr": "Un moteur de recherche est une application web permettant de trouver des ressources à partir d'une requête sous forme de mots. Les ressources peuvent être des pages web, des articles de forums Usenet, des images, des vidéos, des fichiers, etc. Certains sites web offrent un moteur de recherche comme principale fonctionnalité ; on appelle alors « moteur de recherche » le site lui-même. Ce sont des instruments de recherche sur le web sans intervention humaine, ce qui les distingue des annuaires. Ils sont basés sur des « robots », encore appelés « bots », « spiders «, « crawlers » ou « agents », qui parcourent les sites à intervalles réguliers et de façon automatique pour découvrir de nouvelles adresses (URL). Ils suivent les liens hypertextes qui relient les pages les unes aux autres, les uns après les autres. Chaque page identifiée est alors indexée dans une base de données, accessible ensuite par les internautes à partir de mots-clés. C'est par abus de langage qu'on appelle également « moteurs de recherche » des sites web proposant des annuaires de sites web : dans ce cas, ce sont des instruments de recherche élaborés par des personnes qui répertorient et classifient des sites web jugés dignes d'intérêt, et non des robots d'indexation. Les moteurs de recherche ne s'appliquent pas qu'à Internet : certains moteurs sont des logiciels installés sur un ordinateur personnel. Ce sont des moteurs dits « de bureau » qui combinent la recherche parmi les fichiers stockés sur le PC et la recherche parmi les sites Web — on peut citer par exemple Exalead Desktop, Google Desktop et Copernic Desktop Search, Windex Server, etc. On trouve également des métamoteurs, c'est-à-dire des sites web où une même recherche est lancée simultanément sur plusieurs moteurs de recherche, les résultats étant ensuite fusionnés pour être présentés à l'internaute. On peut citer dans cette catégorie Ixquick, Mamma, Kartoo, Framabee ou Lilo.",
        "timestamp": "2018-05-30T15:15:00Z",
        "title_fr": "Moteur de recherche"
      }
    }
  ],
  "total_hits": 12,
  "max_score": 0.633816551223398,
  "took": 480187,
  "facets": {
    "Contributor count": {
      "field": "contributor",
      "total": 12,
      "missing": 0,
      "other": 0,
      "terms": [
        {
          "term": "unknown",
          "count": 12
        }
      ]
    },
    "Timestamp range": {
      "field": "timestamp",
      "total": 12,
      "missing": 0,
      "other": 0,
      "date_ranges": [
        {
          "name": "2011 - 2020",
          "start": "2011-01-01T00:00:00Z",
          "end": "2020-12-31T23:59:59Z",
          "count": 12
        }
      ]
    }
  }
}
```


## Using HTTP REST API

Also you can do above commands via HTTP REST API that listened port 10001 (address port + 1).

### Putting a document

Putting a document via HTTP is as following:

```bash
$ curl -X PUT 'http://127.0.0.1:10002/rest/enwiki_1' -d @./example/doc_enwiki_1.json
```

You can see the result in JSON format. The result of the above request is:

```json
{}
```

### Getting a document

Getting a document via HTTP is as following:

```bash
$ curl -X GET 'http://127.0.0.1:10002/rest/enwiki_1'
```

You can see the result in JSON format. The result of the above request is:

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

### Deleting a document

Deleting a document via HTTP is as following:

```bash
$ curl -X DELETE 'http://127.0.0.1:10002/rest/enwiki_1'
```

You can see the result in JSON format. The result of the above request is:

```json
{}
```

### Bulk update

Bulk update is as following:

```bash
$ curl -X POST 'http://127.0.0.1:10002/rest/_bulk' -d @./example/bulk_put_request.json
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "put_count": 36
}
```

### Searching documents

Searching documents via HTTP is as following:

```bash
$ curl -X POST 'http://127.0.0.1:10002/rest/_search' -d @./example/search_request.json
```

You can see the result in JSON format. The result of the above request is:

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
      "index": "/tmp/blast/node1/index",
      "id": "enwiki_1",
      "score": 0.633816551223398,
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
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "arwiki_1",
      "score": 0.2584513642405507,
      "locations": {
        "text_ar": {
          "search": [
            {
              "pos": 4,
              "start": 45,
              "end": 51,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "arwiki",
        "contributor": "unknown",
        "text_ar": "محرك البحث (بالإنجليزية: Search engine) هو نظام لإسترجاع المعلومات صمم للمساعدة على البحث عن المعلومات المخزنة على أي نظام حاسوبي. تعرض نتائج البحث عادة على شكل قائمة لأماكن تواجد المعلومات ومرتبة وفق معايير معينة. تسمح محركات البحث باختصار مدة البحث والتغلب على مشكلة أحجام البيانات المتصاعدة (إغراق معلوماتي).",
        "timestamp": "2018-03-25T18:04:00Z",
        "title_ar": "محرك بحث"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "hiwiki_1",
      "score": 0.2315458066284217,
      "locations": {
        "text_hi": {
          "search": [
            {
              "pos": 6,
              "start": 93,
              "end": 99,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "hiwiki",
        "contributor": "unknown",
        "text_hi": "ऐसे कम्प्यूटर प्रोग्राम खोजी इंजन (search engine) कहलाते हैं जो किसी कम्प्यूटर सिस्टम पर भण्डारित सूचना में से वांछित सूचना को ढूढ निकालते हैं। ये इंजन प्राप्त परिणामों को प्रायः एक सूची के रूप में प्रस्तुत करते हैं जिससे वांछित सूचना की प्रकृति और उसकी स्थिति का पता चलता है। खोजी इंजन किसी सूचना तक अपेक्षाकृत बहुत कम समय में पहुँचने में हमारी सहायता करते हैं। वे 'सूचना ओवरलोड' से भी हमे बचाते हैं। खोजी इंजन का सबसे प्रचलित रूप 'वेब खोजी इंजन' है जो वर्ल्ड वाइड वेब पर सूचना खोजने के लिये प्रयुक्त होता है।",
        "timestamp": "2017-10-19T20:09:00Z",
        "title_hi": "खोज इंजन"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "itwiki_1",
      "score": 0.22855799635019447,
      "locations": {
        "text_it": {
          "search": [
            {
              "pos": 12,
              "start": 75,
              "end": 81,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "itwiki",
        "contributor": "unknown",
        "text_it": "Nell'ambito delle tecnologie di Internet, un motore di ricerca (in inglese search engine) è un sistema automatico che, su richiesta, analizza un insieme di dati (spesso da esso stesso raccolti) e restituisce un indice dei contenuti disponibili[1] classificandoli in modo automatico in base a formule statistico-matematiche che ne indichino il grado di rilevanza data una determinata chiave di ricerca. Uno dei campi in cui i motori di ricerca trovano maggiore utilizzo è quello dell'information retrieval e nel web. I motori di ricerca più utilizzati nel 2017 sono stati: Google, Bing, Baidu, Qwant, Yandex, Ecosia, DuckDuckGo.",
        "timestamp": "2018-07-16T12:20:00Z",
        "title_it": "Motore di ricerca"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "thwiki_1",
      "score": 0.2018569611073604,
      "locations": {
        "text_th": {
          "search": [
            {
              "pos": 4,
              "start": 38,
              "end": 44,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "thwiki",
        "contributor": "unknown",
        "text_th": "เสิร์ชเอนจิน (search engine) หรือ โปรแกรมค้นหา คือ โปรแกรมที่ช่วยในการสืบค้นหาข้อมูล โดยเฉพาะข้อมูลบนอินเทอร์เน็ต โดยครอบคลุมทั้งข้อความ รูปภาพ ภาพเคลื่อนไหว เพลง ซอฟต์แวร์ แผนที่ ข้อมูลบุคคล กลุ่มข่าว และอื่น ๆ ซึ่งแตกต่างกันไปแล้วแต่โปรแกรมหรือผู้ให้บริการแต่ละราย. เสิร์ชเอนจินส่วนใหญ่จะค้นหาข้อมูลจากคำสำคัญ (คีย์เวิร์ด) ที่ผู้ใช้ป้อนเข้าไป จากนั้นก็จะแสดงรายการผลลัพธ์ที่มันคิดว่าผู้ใช้น่าจะต้องการขึ้นมา ในปัจจุบัน เสิร์ชเอนจินบางตัว เช่น กูเกิล จะบันทึกประวัติการค้นหาและการเลือกผลลัพธ์ของผู้ใช้ไว้ด้วย และจะนำประวัติที่บันทึกไว้นั้น มาช่วยกรองผลลัพธ์ในการค้นหาครั้งต่อ ๆ ไป",
        "timestamp": "2016-06-18T11:06:00Z",
        "title_th": "เสิร์ชเอนจิน"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "zhwiki_1",
      "score": 0.19986816567601795,
      "locations": {
        "text_zh": {
          "search": [
            {
              "pos": 5,
              "start": 24,
              "end": 30,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "zhwiki",
        "contributor": "unknown",
        "text_zh": "搜索引擎（英语：search engine）是一种信息检索系统，旨在协助搜索存储在计算机系统中的信息。搜索结果一般被称为“hits”，通常会以表单的形式列出。网络搜索引擎是最常见、公开的一种搜索引擎，其功能为搜索万维网上储存的信息.",
        "timestamp": "2018-08-27T05:47:00Z",
        "title_zh": "搜索引擎"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "bgwiki_1",
      "score": 0.18275270089950202,
      "locations": {
        "text_bg": {
          "search": [
            {
              "pos": 8,
              "start": 82,
              "end": 88,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "bgwiki",
        "contributor": "unknown",
        "text_bg": "Търсачка или търсеща машина (на английски: Web search engine) е специализиран софтуер за извличане на информация, съхранена в компютърна система или мрежа. Това може да е персонален компютър, Интернет, корпоративна мрежа и т.н. Без допълнителни уточнения, най-често под търсачка се разбира уеб(-)търсачка, която търси в Интернет. Други видове търсачки са корпоративните търсачки, които търсят в интранет мрежите, личните търсачки – за индивидуалните компютри и мобилните търсачки. В търсачката потребителят (търсещият) прави запитване за съдържание, отговарящо на определен критерий (обикновено такъв, който съдържа определени думи и фрази). В резултат се получават списък от точки, които отговарят, пълно или частично, на този критерий. Търсачките обикновено използват редовно подновявани индекси, за да оперират бързо и ефикасно. Някои търсачки също търсят в информацията, която е на разположение в нюзгрупите и други големи бази данни. За разлика от Уеб директориите, които се поддържат от хора редактори, търсачките оперират алгоритмично. Повечето Интернет търсачки са притежавани от различни корпорации.",
        "timestamp": "2018-07-11T11:03:00Z",
        "title_bg": "Търсачка"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "idwiki_1",
      "score": 0.17060026115019133,
      "locations": {
        "text_id": {
          "search": [
            {
              "pos": 11,
              "start": 62,
              "end": 68,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "idwiki",
        "contributor": "unknown",
        "text_id": "Mesin pencari web atau mesin telusur web (bahasa Inggris: web search engine) adalah program komputer yang dirancang untuk melakukan pencarian atas berkas-berkas yang tersimpan dalam layanan www, ftp, publikasi milis, ataupun news group dalam sebuah ataupun sejumlah komputer peladen dalam suatu jaringan. Mesin pencari merupakan perangkat penelusur informasi dari dokumen-dokumen yang tersedia. Hasil pencarian umumnya ditampilkan dalam bentuk daftar yang seringkali diurutkan menurut tingkat akurasi ataupun rasio pengunjung atas suatu berkas yang disebut sebagai hits. Informasi yang menjadi target pencarian bisa terdapat dalam berbagai macam jenis berkas seperti halaman situs web, gambar, ataupun jenis-jenis berkas lainnya. Beberapa mesin pencari juga diketahui melakukan pengumpulan informasi atas data yang tersimpan dalam suatu basis data ataupun direktori web. Sebagian besar mesin pencari dijalankan oleh perusahaan swasta yang menggunakan algoritme kepemilikan dan basis data tertutup, di antaranya yang paling populer adalah safari Google (MSN Search dan Yahoo!). Telah ada beberapa upaya menciptakan mesin pencari dengan sumber terbuka (open source), contohnya adalah Htdig, Nutch, Egothor dan OpenFTS.",
        "timestamp": "2017-11-20T17:47:00Z",
        "title_id": "Mesin pencari web"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "ptwiki_1",
      "score": 0.1688012644026126,
      "locations": {
        "text_pt": {
          "search": [
            {
              "pos": 16,
              "start": 111,
              "end": 117,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "ptwiki",
        "contributor": "unknown",
        "text_pt": "Motor de pesquisa (português europeu) ou ferramenta de busca (português brasileiro) ou buscador (em inglês: search engine) é um programa desenhado para procurar palavras-chave fornecidas pelo utilizador em documentos e bases de dados. No contexto da internet, um motor de pesquisa permite procurar palavras-chave em documentos alojados na world wide web, como aqueles que se encontram armazenados em websites. Os motores de busca surgiram logo após o aparecimento da Internet, com a intenção de prestar um serviço extremamente importante: a busca de qualquer informação na rede, apresentando os resultados de uma forma organizada, e também com a proposta de fazer isto de uma maneira rápida e eficiente. A partir deste preceito básico, diversas empresas se desenvolveram, chegando algumas a valer milhões de dólares. Entre as maiores empresas encontram-se o Google, o Yahoo, o Bing, o Lycos, o Cadê e, mais recentemente, a Amazon.com com o seu mecanismo de busca A9 porém inativo. Os buscadores se mostraram imprescindíveis para o fluxo de acesso e a conquista novos visitantes. Antes do advento da Web, havia sistemas para outros protocolos ou usos, como o Archie para sites FTP anônimos e o Veronica para o Gopher (protocolo de redes de computadores que foi desenhado para indexar repositórios de documentos na Internet, baseado-se em menus).",
        "timestamp": "2017-11-09T14:38:00Z",
        "title_pt": "Motor de busca"
      }
    },
    {
      "index": "/tmp/blast/node1/index",
      "id": "frwiki_1",
      "score": 0.14418354794511895,
      "locations": {
        "text_fr": {
          "search": [
            {
              "pos": 253,
              "start": 1656,
              "end": 1662,
              "array_positions": null
            }
          ]
        }
      },
      "sort": [
        "_score"
      ],
      "fields": {
        "_type": "frwiki",
        "contributor": "unknown",
        "text_fr": "Un moteur de recherche est une application web permettant de trouver des ressources à partir d'une requête sous forme de mots. Les ressources peuvent être des pages web, des articles de forums Usenet, des images, des vidéos, des fichiers, etc. Certains sites web offrent un moteur de recherche comme principale fonctionnalité ; on appelle alors « moteur de recherche » le site lui-même. Ce sont des instruments de recherche sur le web sans intervention humaine, ce qui les distingue des annuaires. Ils sont basés sur des « robots », encore appelés « bots », « spiders «, « crawlers » ou « agents », qui parcourent les sites à intervalles réguliers et de façon automatique pour découvrir de nouvelles adresses (URL). Ils suivent les liens hypertextes qui relient les pages les unes aux autres, les uns après les autres. Chaque page identifiée est alors indexée dans une base de données, accessible ensuite par les internautes à partir de mots-clés. C'est par abus de langage qu'on appelle également « moteurs de recherche » des sites web proposant des annuaires de sites web : dans ce cas, ce sont des instruments de recherche élaborés par des personnes qui répertorient et classifient des sites web jugés dignes d'intérêt, et non des robots d'indexation. Les moteurs de recherche ne s'appliquent pas qu'à Internet : certains moteurs sont des logiciels installés sur un ordinateur personnel. Ce sont des moteurs dits « de bureau » qui combinent la recherche parmi les fichiers stockés sur le PC et la recherche parmi les sites Web — on peut citer par exemple Exalead Desktop, Google Desktop et Copernic Desktop Search, Windex Server, etc. On trouve également des métamoteurs, c'est-à-dire des sites web où une même recherche est lancée simultanément sur plusieurs moteurs de recherche, les résultats étant ensuite fusionnés pour être présentés à l'internaute. On peut citer dans cette catégorie Ixquick, Mamma, Kartoo, Framabee ou Lilo.",
        "timestamp": "2018-05-30T15:15:00Z",
        "title_fr": "Moteur de recherche"
      }
    }
  ],
  "total_hits": 12,
  "max_score": 0.633816551223398,
  "took": 480187,
  "facets": {
    "Contributor count": {
      "field": "contributor",
      "total": 12,
      "missing": 0,
      "other": 0,
      "terms": [
        {
          "term": "unknown",
          "count": 12
        }
      ]
    },
    "Timestamp range": {
      "field": "timestamp",
      "total": 12,
      "missing": 0,
      "other": 0,
      "date_ranges": [
        {
          "name": "2011 - 2020",
          "start": "2011-01-01T00:00:00Z",
          "end": "2020-12-31T23:59:59Z",
          "count": 12
        }
      ]
    }
  }
}
```

## Bringing up a cluster

Blast is easy to bring up the cluster. Blast data node is already running, but that is not fault tolerant. If you need to increase the fault tolerance, bring up 2 more data nodes like so:

```bash
$ ./bin/blastd data \
    --raft-addr=127.0.0.1:11000 \
    --grpc-addr=127.0.0.1:11001 \
    --http-addr=127.0.0.1:11002 \
    --raft-node-id=node2 \
    --raft-dir=/tmp/blast/node2/raft \
    --store-dir=/tmp/blast/node2/store \
    --index-dir=/tmp/blast/node2/index \
    --index-mapping-file=./etc/index_mapping.json \
    --peer-grpc-addr=127.0.0.1:10001
$ ./bin/blastd data \
    --raft-addr=127.0.0.1:12000 \
    --grpc-addr=127.0.0.1:12001 \
    --http-addr=127.0.0.1:12002 \
    --raft-node-id=node3 \
    --raft-dir=/tmp/blast/node3/raft \
    --store-dir=/tmp/blast/node3/store \
    --index-dir=/tmp/blast/node3/index \
    --index-mapping-file=./etc/index_mapping.json \
    --peer-grpc-addr=127.0.0.1:10001
```

_Above example shows each Blast node running on the same host, so each node must listen on different ports. This would not be necessary if each node ran on a different host._

So you have a 3-node cluster. That way you can tolerate the failure of 1 node. You can check the peers with the following command:

```bash
$ ./bin/blast cluster --grpc-addr=127.0.0.1:10001
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "nodes": [
    {
      "id": "node1",
      "raft_addr": "127.0.0.1:10000",
      "grpc_addr": "127.0.0.1:10001",
      "http_addr": "127.0.0.1:10002",
      "leader": true
    },
    {
      "id": "node2",
      "raft_addr": "127.0.0.1:11000",
      "grpc_addr": "127.0.0.1:11001",
      "http_addr": "127.0.0.1:11002"
    },
    {
      "id": "node3",
      "raft_addr": "127.0.0.1:12000",
      "grpc_addr": "127.0.0.1:12001",
      "http_addr": "127.0.0.1:12002"
    }
  ]
}
```

Recommend 3 or more odd number of nodes in the cluster. In failure scenarios, data loss is inevitable, so avoid deploying single nodes.

This tells each new node to join the existing node. Once joined, each node now knows about the key:

Following command puts a document to node0:

```bash
$ cat ./example/doc_enwiki_1.json | xargs -0 ./bin/blast put --grpc-addr=127.0.0.1:10001 enwiki_1
```

So, you can get a document from node1 like following:

```bash
$ ./bin/blast get --grpc-addr=127.0.0.1:10001 enwiki_1
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

Also, you can get same document from node2 (127.0.0.1:11001) like following:

```bash
$ ./bin/blast get --grpc-addr=127.0.0.1:11001 enwiki_1
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

Lastly, you can get same document from node3 (127.0.0.1:12001) like following:

```bash
$ ./bin/blast get --grpc-addr=127.0.0.1:12001 enwiki_1
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

### Running Blast index node on Docker

Running a Blast data node on Docker. Start Blast data node like so:

```bash
$ docker run --rm --name blast1 \
    -p 10000:10000 \
    -p 10001:10001 \
    -p 10002:10002 \
    mosuka/blast:latest blastd data \
      --bind-addr=:10000 \
      --grpc-addr=:10001 \
      --http-addr=:10002 \
      --raft-node-id=node1
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
$ docker-compose rm
```
