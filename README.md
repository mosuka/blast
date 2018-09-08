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
$ cat ./example/doc_enwiki_1.json | xargs -0 ./bin/blast put --grpc-addr=localhost:10001 --pretty-print enwiki_1
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
$ ./bin/blast get --grpc-addr=localhost:10001 --pretty-print enwiki_1
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
  },
  "success": true
}
```

### Deleting a document

Deleting a document is as following:

```
$ ./bin/blast delete --grpc-addr=localhost:10001 --pretty-print enwiki_1
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
  "put_count": 36,
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
      "Contributor count": {
        "field": "contributor",
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "count": 12,
            "term": "unknown"
          }
        ],
        "total": 12
      },
      "Timestamp range": {
        "date_ranges": [
          {
            "count": 12,
            "end": "2020-12-31T23:59:59Z",
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z"
          }
        ],
        "field": "timestamp",
        "missing": 0,
        "other": 0,
        "total": 12
      }
    },
    "hits": [
      {
        "fields": {
          "_type": "enwiki",
          "contributor": "unknown",
          "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_en": "Search engine (computing)"
        },
        "id": "enwiki_1",
        "index": "/tmp/blast/node1/index",
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
              },
              {
                "array_positions": null,
                "end": 421,
                "pos": 68,
                "start": 415
              },
              {
                "array_positions": null,
                "end": 444,
                "pos": 73,
                "start": 438
              },
              {
                "array_positions": null,
                "end": 466,
                "pos": 76,
                "start": 458
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
        "score": 0.633816551223398,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "arwiki",
          "contributor": "unknown",
          "text_ar": "محرك البحث (بالإنجليزية: Search engine) هو نظام لإسترجاع المعلومات صمم للمساعدة على البحث عن المعلومات المخزنة على أي نظام حاسوبي. تعرض نتائج البحث عادة على شكل قائمة لأماكن تواجد المعلومات ومرتبة وفق معايير معينة. تسمح محركات البحث باختصار مدة البحث والتغلب على مشكلة أحجام البيانات المتصاعدة (إغراق معلوماتي).",
          "timestamp": "2018-03-25T18:04:00Z",
          "title_ar": "محرك بحث"
        },
        "id": "arwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_ar": {
            "search": [
              {
                "array_positions": null,
                "end": 51,
                "pos": 4,
                "start": 45
              }
            ]
          }
        },
        "score": 0.2584513642405507,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "hiwiki",
          "contributor": "unknown",
          "text_hi": "ऐसे कम्प्यूटर प्रोग्राम खोजी इंजन (search engine) कहलाते हैं जो किसी कम्प्यूटर सिस्टम पर भण्डारित सूचना में से वांछित सूचना को ढूढ निकालते हैं। ये इंजन प्राप्त परिणामों को प्रायः एक सूची के रूप में प्रस्तुत करते हैं जिससे वांछित सूचना की प्रकृति और उसकी स्थिति का पता चलता है। खोजी इंजन किसी सूचना तक अपेक्षाकृत बहुत कम समय में पहुँचने में हमारी सहायता करते हैं। वे 'सूचना ओवरलोड' से भी हमे बचाते हैं। खोजी इंजन का सबसे प्रचलित रूप 'वेब खोजी इंजन' है जो वर्ल्ड वाइड वेब पर सूचना खोजने के लिये प्रयुक्त होता है।",
          "timestamp": "2017-10-19T20:09:00Z",
          "title_hi": "खोज इंजन"
        },
        "id": "hiwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_hi": {
            "search": [
              {
                "array_positions": null,
                "end": 99,
                "pos": 6,
                "start": 93
              }
            ]
          }
        },
        "score": 0.2315458066284217,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "itwiki",
          "contributor": "unknown",
          "text_it": "Nell'ambito delle tecnologie di Internet, un motore di ricerca (in inglese search engine) è un sistema automatico che, su richiesta, analizza un insieme di dati (spesso da esso stesso raccolti) e restituisce un indice dei contenuti disponibili[1] classificandoli in modo automatico in base a formule statistico-matematiche che ne indichino il grado di rilevanza data una determinata chiave di ricerca. Uno dei campi in cui i motori di ricerca trovano maggiore utilizzo è quello dell'information retrieval e nel web. I motori di ricerca più utilizzati nel 2017 sono stati: Google, Bing, Baidu, Qwant, Yandex, Ecosia, DuckDuckGo.",
          "timestamp": "2018-07-16T12:20:00Z",
          "title_it": "Motore di ricerca"
        },
        "id": "itwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_it": {
            "search": [
              {
                "array_positions": null,
                "end": 81,
                "pos": 12,
                "start": 75
              }
            ]
          }
        },
        "score": 0.22855799635019447,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "thwiki",
          "contributor": "unknown",
          "text_th": "เสิร์ชเอนจิน (search engine) หรือ โปรแกรมค้นหา คือ โปรแกรมที่ช่วยในการสืบค้นหาข้อมูล โดยเฉพาะข้อมูลบนอินเทอร์เน็ต โดยครอบคลุมทั้งข้อความ รูปภาพ ภาพเคลื่อนไหว เพลง ซอฟต์แวร์ แผนที่ ข้อมูลบุคคล กลุ่มข่าว และอื่น ๆ ซึ่งแตกต่างกันไปแล้วแต่โปรแกรมหรือผู้ให้บริการแต่ละราย. เสิร์ชเอนจินส่วนใหญ่จะค้นหาข้อมูลจากคำสำคัญ (คีย์เวิร์ด) ที่ผู้ใช้ป้อนเข้าไป จากนั้นก็จะแสดงรายการผลลัพธ์ที่มันคิดว่าผู้ใช้น่าจะต้องการขึ้นมา ในปัจจุบัน เสิร์ชเอนจินบางตัว เช่น กูเกิล จะบันทึกประวัติการค้นหาและการเลือกผลลัพธ์ของผู้ใช้ไว้ด้วย และจะนำประวัติที่บันทึกไว้นั้น มาช่วยกรองผลลัพธ์ในการค้นหาครั้งต่อ ๆ ไป",
          "timestamp": "2016-06-18T11:06:00Z",
          "title_th": "เสิร์ชเอนจิน"
        },
        "id": "thwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_th": {
            "search": [
              {
                "array_positions": null,
                "end": 44,
                "pos": 4,
                "start": 38
              }
            ]
          }
        },
        "score": 0.2018569611073604,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "zhwiki",
          "contributor": "unknown",
          "text_zh": "搜索引擎（英语：search engine）是一种信息检索系统，旨在协助搜索存储在计算机系统中的信息。搜索结果一般被称为“hits”，通常会以表单的形式列出。网络搜索引擎是最常见、公开的一种搜索引擎，其功能为搜索万维网上储存的信息.",
          "timestamp": "2018-08-27T05:47:00Z",
          "title_zh": "搜索引擎"
        },
        "id": "zhwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_zh": {
            "search": [
              {
                "array_positions": null,
                "end": 30,
                "pos": 5,
                "start": 24
              }
            ]
          }
        },
        "score": 0.19986816567601795,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "bgwiki",
          "contributor": "unknown",
          "text_bg": "Търсачка или търсеща машина (на английски: Web search engine) е специализиран софтуер за извличане на информация, съхранена в компютърна система или мрежа. Това може да е персонален компютър, Интернет, корпоративна мрежа и т.н. Без допълнителни уточнения, най-често под търсачка се разбира уеб(-)търсачка, която търси в Интернет. Други видове търсачки са корпоративните търсачки, които търсят в интранет мрежите, личните търсачки – за индивидуалните компютри и мобилните търсачки. В търсачката потребителят (търсещият) прави запитване за съдържание, отговарящо на определен критерий (обикновено такъв, който съдържа определени думи и фрази). В резултат се получават списък от точки, които отговарят, пълно или частично, на този критерий. Търсачките обикновено използват редовно подновявани индекси, за да оперират бързо и ефикасно. Някои търсачки също търсят в информацията, която е на разположение в нюзгрупите и други големи бази данни. За разлика от Уеб директориите, които се поддържат от хора редактори, търсачките оперират алгоритмично. Повечето Интернет търсачки са притежавани от различни корпорации.",
          "timestamp": "2018-07-11T11:03:00Z",
          "title_bg": "Търсачка"
        },
        "id": "bgwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_bg": {
            "search": [
              {
                "array_positions": null,
                "end": 88,
                "pos": 8,
                "start": 82
              }
            ]
          }
        },
        "score": 0.18275270089950202,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "idwiki",
          "contributor": "unknown",
          "text_id": "Mesin pencari web atau mesin telusur web (bahasa Inggris: web search engine) adalah program komputer yang dirancang untuk melakukan pencarian atas berkas-berkas yang tersimpan dalam layanan www, ftp, publikasi milis, ataupun news group dalam sebuah ataupun sejumlah komputer peladen dalam suatu jaringan. Mesin pencari merupakan perangkat penelusur informasi dari dokumen-dokumen yang tersedia. Hasil pencarian umumnya ditampilkan dalam bentuk daftar yang seringkali diurutkan menurut tingkat akurasi ataupun rasio pengunjung atas suatu berkas yang disebut sebagai hits. Informasi yang menjadi target pencarian bisa terdapat dalam berbagai macam jenis berkas seperti halaman situs web, gambar, ataupun jenis-jenis berkas lainnya. Beberapa mesin pencari juga diketahui melakukan pengumpulan informasi atas data yang tersimpan dalam suatu basis data ataupun direktori web. Sebagian besar mesin pencari dijalankan oleh perusahaan swasta yang menggunakan algoritme kepemilikan dan basis data tertutup, di antaranya yang paling populer adalah safari Google (MSN Search dan Yahoo!). Telah ada beberapa upaya menciptakan mesin pencari dengan sumber terbuka (open source), contohnya adalah Htdig, Nutch, Egothor dan OpenFTS.",
          "timestamp": "2017-11-20T17:47:00Z",
          "title_id": "Mesin pencari web"
        },
        "id": "idwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_id": {
            "search": [
              {
                "array_positions": null,
                "end": 68,
                "pos": 11,
                "start": 62
              }
            ]
          }
        },
        "score": 0.17060026115019133,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "ptwiki",
          "contributor": "unknown",
          "text_pt": "Motor de pesquisa (português europeu) ou ferramenta de busca (português brasileiro) ou buscador (em inglês: search engine) é um programa desenhado para procurar palavras-chave fornecidas pelo utilizador em documentos e bases de dados. No contexto da internet, um motor de pesquisa permite procurar palavras-chave em documentos alojados na world wide web, como aqueles que se encontram armazenados em websites. Os motores de busca surgiram logo após o aparecimento da Internet, com a intenção de prestar um serviço extremamente importante: a busca de qualquer informação na rede, apresentando os resultados de uma forma organizada, e também com a proposta de fazer isto de uma maneira rápida e eficiente. A partir deste preceito básico, diversas empresas se desenvolveram, chegando algumas a valer milhões de dólares. Entre as maiores empresas encontram-se o Google, o Yahoo, o Bing, o Lycos, o Cadê e, mais recentemente, a Amazon.com com o seu mecanismo de busca A9 porém inativo. Os buscadores se mostraram imprescindíveis para o fluxo de acesso e a conquista novos visitantes. Antes do advento da Web, havia sistemas para outros protocolos ou usos, como o Archie para sites FTP anônimos e o Veronica para o Gopher (protocolo de redes de computadores que foi desenhado para indexar repositórios de documentos na Internet, baseado-se em menus).",
          "timestamp": "2017-11-09T14:38:00Z",
          "title_pt": "Motor de busca"
        },
        "id": "ptwiki_1",
        "index": "/tmp/blast/node1/index",
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
        "score": 0.1688012644026126,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "frwiki",
          "contributor": "unknown",
          "text_fr": "Un moteur de recherche est une application web permettant de trouver des ressources à partir d'une requête sous forme de mots. Les ressources peuvent être des pages web, des articles de forums Usenet, des images, des vidéos, des fichiers, etc. Certains sites web offrent un moteur de recherche comme principale fonctionnalité ; on appelle alors « moteur de recherche » le site lui-même. Ce sont des instruments de recherche sur le web sans intervention humaine, ce qui les distingue des annuaires. Ils sont basés sur des « robots », encore appelés « bots », « spiders «, « crawlers » ou « agents », qui parcourent les sites à intervalles réguliers et de façon automatique pour découvrir de nouvelles adresses (URL). Ils suivent les liens hypertextes qui relient les pages les unes aux autres, les uns après les autres. Chaque page identifiée est alors indexée dans une base de données, accessible ensuite par les internautes à partir de mots-clés. C'est par abus de langage qu'on appelle également « moteurs de recherche » des sites web proposant des annuaires de sites web : dans ce cas, ce sont des instruments de recherche élaborés par des personnes qui répertorient et classifient des sites web jugés dignes d'intérêt, et non des robots d'indexation. Les moteurs de recherche ne s'appliquent pas qu'à Internet : certains moteurs sont des logiciels installés sur un ordinateur personnel. Ce sont des moteurs dits « de bureau » qui combinent la recherche parmi les fichiers stockés sur le PC et la recherche parmi les sites Web — on peut citer par exemple Exalead Desktop, Google Desktop et Copernic Desktop Search, Windex Server, etc. On trouve également des métamoteurs, c'est-à-dire des sites web où une même recherche est lancée simultanément sur plusieurs moteurs de recherche, les résultats étant ensuite fusionnés pour être présentés à l'internaute. On peut citer dans cette catégorie Ixquick, Mamma, Kartoo, Framabee ou Lilo.",
          "timestamp": "2018-05-30T15:15:00Z",
          "title_fr": "Moteur de recherche"
        },
        "id": "frwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_fr": {
            "search": [
              {
                "array_positions": null,
                "end": 1662,
                "pos": 253,
                "start": 1656
              }
            ]
          }
        },
        "score": 0.14418354794511895,
        "sort": [
          "_score"
        ]
      }
    ],
    "max_score": 0.633816551223398,
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
    "took": 571914,
    "total_hits": 12
  },
  "success": true
}
```


## Using HTTP REST API

Also you can do above commands via HTTP REST API that listened port 10001 (address port + 1).

### Putting a document

Putting a document via HTTP is as following:

```bash
$ curl -X PUT 'http://localhost:10002/rest/enwiki_1?pretty-print=true' -d @./example/doc_enwiki_1.json
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
$ curl -X GET 'http://localhost:10002/rest/enwiki_1?pretty-print=true'
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
  },
  "success": true
}
```

### Deleting a document

Deleting a document via HTTP is as following:

```bash
$ curl -X DELETE 'http://localhost:10002/rest/enwiki_1?pretty-print=true'
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
  "put_count": 36,
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
      "Contributor count": {
        "field": "contributor",
        "missing": 0,
        "other": 0,
        "terms": [
          {
            "count": 12,
            "term": "unknown"
          }
        ],
        "total": 12
      },
      "Timestamp range": {
        "date_ranges": [
          {
            "count": 12,
            "end": "2020-12-31T23:59:59Z",
            "name": "2011 - 2020",
            "start": "2011-01-01T00:00:00Z"
          }
        ],
        "field": "timestamp",
        "missing": 0,
        "other": 0,
        "total": 12
      }
    },
    "hits": [
      {
        "fields": {
          "_type": "enwiki",
          "contributor": "unknown",
          "text_en": "A search engine is an information retrieval system designed to help find information stored on a computer system. The search results are usually presented in a list and are commonly called hits. Search engines help to minimize the time required to find information and the amount of information which must be consulted, akin to other techniques for managing information overload. The most public, visible form of a search engine is a Web search engine which searches for information on the World Wide Web.",
          "timestamp": "2018-07-04T05:41:00Z",
          "title_en": "Search engine (computing)"
        },
        "id": "enwiki_1",
        "index": "/tmp/blast/node1/index",
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
              },
              {
                "array_positions": null,
                "end": 421,
                "pos": 68,
                "start": 415
              },
              {
                "array_positions": null,
                "end": 444,
                "pos": 73,
                "start": 438
              },
              {
                "array_positions": null,
                "end": 466,
                "pos": 76,
                "start": 458
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
        "score": 0.633816551223398,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "arwiki",
          "contributor": "unknown",
          "text_ar": "محرك البحث (بالإنجليزية: Search engine) هو نظام لإسترجاع المعلومات صمم للمساعدة على البحث عن المعلومات المخزنة على أي نظام حاسوبي. تعرض نتائج البحث عادة على شكل قائمة لأماكن تواجد المعلومات ومرتبة وفق معايير معينة. تسمح محركات البحث باختصار مدة البحث والتغلب على مشكلة أحجام البيانات المتصاعدة (إغراق معلوماتي).",
          "timestamp": "2018-03-25T18:04:00Z",
          "title_ar": "محرك بحث"
        },
        "id": "arwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_ar": {
            "search": [
              {
                "array_positions": null,
                "end": 51,
                "pos": 4,
                "start": 45
              }
            ]
          }
        },
        "score": 0.2584513642405507,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "hiwiki",
          "contributor": "unknown",
          "text_hi": "ऐसे कम्प्यूटर प्रोग्राम खोजी इंजन (search engine) कहलाते हैं जो किसी कम्प्यूटर सिस्टम पर भण्डारित सूचना में से वांछित सूचना को ढूढ निकालते हैं। ये इंजन प्राप्त परिणामों को प्रायः एक सूची के रूप में प्रस्तुत करते हैं जिससे वांछित सूचना की प्रकृति और उसकी स्थिति का पता चलता है। खोजी इंजन किसी सूचना तक अपेक्षाकृत बहुत कम समय में पहुँचने में हमारी सहायता करते हैं। वे 'सूचना ओवरलोड' से भी हमे बचाते हैं। खोजी इंजन का सबसे प्रचलित रूप 'वेब खोजी इंजन' है जो वर्ल्ड वाइड वेब पर सूचना खोजने के लिये प्रयुक्त होता है।",
          "timestamp": "2017-10-19T20:09:00Z",
          "title_hi": "खोज इंजन"
        },
        "id": "hiwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_hi": {
            "search": [
              {
                "array_positions": null,
                "end": 99,
                "pos": 6,
                "start": 93
              }
            ]
          }
        },
        "score": 0.2315458066284217,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "itwiki",
          "contributor": "unknown",
          "text_it": "Nell'ambito delle tecnologie di Internet, un motore di ricerca (in inglese search engine) è un sistema automatico che, su richiesta, analizza un insieme di dati (spesso da esso stesso raccolti) e restituisce un indice dei contenuti disponibili[1] classificandoli in modo automatico in base a formule statistico-matematiche che ne indichino il grado di rilevanza data una determinata chiave di ricerca. Uno dei campi in cui i motori di ricerca trovano maggiore utilizzo è quello dell'information retrieval e nel web. I motori di ricerca più utilizzati nel 2017 sono stati: Google, Bing, Baidu, Qwant, Yandex, Ecosia, DuckDuckGo.",
          "timestamp": "2018-07-16T12:20:00Z",
          "title_it": "Motore di ricerca"
        },
        "id": "itwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_it": {
            "search": [
              {
                "array_positions": null,
                "end": 81,
                "pos": 12,
                "start": 75
              }
            ]
          }
        },
        "score": 0.22855799635019447,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "thwiki",
          "contributor": "unknown",
          "text_th": "เสิร์ชเอนจิน (search engine) หรือ โปรแกรมค้นหา คือ โปรแกรมที่ช่วยในการสืบค้นหาข้อมูล โดยเฉพาะข้อมูลบนอินเทอร์เน็ต โดยครอบคลุมทั้งข้อความ รูปภาพ ภาพเคลื่อนไหว เพลง ซอฟต์แวร์ แผนที่ ข้อมูลบุคคล กลุ่มข่าว และอื่น ๆ ซึ่งแตกต่างกันไปแล้วแต่โปรแกรมหรือผู้ให้บริการแต่ละราย. เสิร์ชเอนจินส่วนใหญ่จะค้นหาข้อมูลจากคำสำคัญ (คีย์เวิร์ด) ที่ผู้ใช้ป้อนเข้าไป จากนั้นก็จะแสดงรายการผลลัพธ์ที่มันคิดว่าผู้ใช้น่าจะต้องการขึ้นมา ในปัจจุบัน เสิร์ชเอนจินบางตัว เช่น กูเกิล จะบันทึกประวัติการค้นหาและการเลือกผลลัพธ์ของผู้ใช้ไว้ด้วย และจะนำประวัติที่บันทึกไว้นั้น มาช่วยกรองผลลัพธ์ในการค้นหาครั้งต่อ ๆ ไป",
          "timestamp": "2016-06-18T11:06:00Z",
          "title_th": "เสิร์ชเอนจิน"
        },
        "id": "thwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_th": {
            "search": [
              {
                "array_positions": null,
                "end": 44,
                "pos": 4,
                "start": 38
              }
            ]
          }
        },
        "score": 0.2018569611073604,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "zhwiki",
          "contributor": "unknown",
          "text_zh": "搜索引擎（英语：search engine）是一种信息检索系统，旨在协助搜索存储在计算机系统中的信息。搜索结果一般被称为“hits”，通常会以表单的形式列出。网络搜索引擎是最常见、公开的一种搜索引擎，其功能为搜索万维网上储存的信息.",
          "timestamp": "2018-08-27T05:47:00Z",
          "title_zh": "搜索引擎"
        },
        "id": "zhwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_zh": {
            "search": [
              {
                "array_positions": null,
                "end": 30,
                "pos": 5,
                "start": 24
              }
            ]
          }
        },
        "score": 0.19986816567601795,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "bgwiki",
          "contributor": "unknown",
          "text_bg": "Търсачка или търсеща машина (на английски: Web search engine) е специализиран софтуер за извличане на информация, съхранена в компютърна система или мрежа. Това може да е персонален компютър, Интернет, корпоративна мрежа и т.н. Без допълнителни уточнения, най-често под търсачка се разбира уеб(-)търсачка, която търси в Интернет. Други видове търсачки са корпоративните търсачки, които търсят в интранет мрежите, личните търсачки – за индивидуалните компютри и мобилните търсачки. В търсачката потребителят (търсещият) прави запитване за съдържание, отговарящо на определен критерий (обикновено такъв, който съдържа определени думи и фрази). В резултат се получават списък от точки, които отговарят, пълно или частично, на този критерий. Търсачките обикновено използват редовно подновявани индекси, за да оперират бързо и ефикасно. Някои търсачки също търсят в информацията, която е на разположение в нюзгрупите и други големи бази данни. За разлика от Уеб директориите, които се поддържат от хора редактори, търсачките оперират алгоритмично. Повечето Интернет търсачки са притежавани от различни корпорации.",
          "timestamp": "2018-07-11T11:03:00Z",
          "title_bg": "Търсачка"
        },
        "id": "bgwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_bg": {
            "search": [
              {
                "array_positions": null,
                "end": 88,
                "pos": 8,
                "start": 82
              }
            ]
          }
        },
        "score": 0.18275270089950202,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "idwiki",
          "contributor": "unknown",
          "text_id": "Mesin pencari web atau mesin telusur web (bahasa Inggris: web search engine) adalah program komputer yang dirancang untuk melakukan pencarian atas berkas-berkas yang tersimpan dalam layanan www, ftp, publikasi milis, ataupun news group dalam sebuah ataupun sejumlah komputer peladen dalam suatu jaringan. Mesin pencari merupakan perangkat penelusur informasi dari dokumen-dokumen yang tersedia. Hasil pencarian umumnya ditampilkan dalam bentuk daftar yang seringkali diurutkan menurut tingkat akurasi ataupun rasio pengunjung atas suatu berkas yang disebut sebagai hits. Informasi yang menjadi target pencarian bisa terdapat dalam berbagai macam jenis berkas seperti halaman situs web, gambar, ataupun jenis-jenis berkas lainnya. Beberapa mesin pencari juga diketahui melakukan pengumpulan informasi atas data yang tersimpan dalam suatu basis data ataupun direktori web. Sebagian besar mesin pencari dijalankan oleh perusahaan swasta yang menggunakan algoritme kepemilikan dan basis data tertutup, di antaranya yang paling populer adalah safari Google (MSN Search dan Yahoo!). Telah ada beberapa upaya menciptakan mesin pencari dengan sumber terbuka (open source), contohnya adalah Htdig, Nutch, Egothor dan OpenFTS.",
          "timestamp": "2017-11-20T17:47:00Z",
          "title_id": "Mesin pencari web"
        },
        "id": "idwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_id": {
            "search": [
              {
                "array_positions": null,
                "end": 68,
                "pos": 11,
                "start": 62
              }
            ]
          }
        },
        "score": 0.17060026115019133,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "ptwiki",
          "contributor": "unknown",
          "text_pt": "Motor de pesquisa (português europeu) ou ferramenta de busca (português brasileiro) ou buscador (em inglês: search engine) é um programa desenhado para procurar palavras-chave fornecidas pelo utilizador em documentos e bases de dados. No contexto da internet, um motor de pesquisa permite procurar palavras-chave em documentos alojados na world wide web, como aqueles que se encontram armazenados em websites. Os motores de busca surgiram logo após o aparecimento da Internet, com a intenção de prestar um serviço extremamente importante: a busca de qualquer informação na rede, apresentando os resultados de uma forma organizada, e também com a proposta de fazer isto de uma maneira rápida e eficiente. A partir deste preceito básico, diversas empresas se desenvolveram, chegando algumas a valer milhões de dólares. Entre as maiores empresas encontram-se o Google, o Yahoo, o Bing, o Lycos, o Cadê e, mais recentemente, a Amazon.com com o seu mecanismo de busca A9 porém inativo. Os buscadores se mostraram imprescindíveis para o fluxo de acesso e a conquista novos visitantes. Antes do advento da Web, havia sistemas para outros protocolos ou usos, como o Archie para sites FTP anônimos e o Veronica para o Gopher (protocolo de redes de computadores que foi desenhado para indexar repositórios de documentos na Internet, baseado-se em menus).",
          "timestamp": "2017-11-09T14:38:00Z",
          "title_pt": "Motor de busca"
        },
        "id": "ptwiki_1",
        "index": "/tmp/blast/node1/index",
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
        "score": 0.1688012644026126,
        "sort": [
          "_score"
        ]
      },
      {
        "fields": {
          "_type": "frwiki",
          "contributor": "unknown",
          "text_fr": "Un moteur de recherche est une application web permettant de trouver des ressources à partir d'une requête sous forme de mots. Les ressources peuvent être des pages web, des articles de forums Usenet, des images, des vidéos, des fichiers, etc. Certains sites web offrent un moteur de recherche comme principale fonctionnalité ; on appelle alors « moteur de recherche » le site lui-même. Ce sont des instruments de recherche sur le web sans intervention humaine, ce qui les distingue des annuaires. Ils sont basés sur des « robots », encore appelés « bots », « spiders «, « crawlers » ou « agents », qui parcourent les sites à intervalles réguliers et de façon automatique pour découvrir de nouvelles adresses (URL). Ils suivent les liens hypertextes qui relient les pages les unes aux autres, les uns après les autres. Chaque page identifiée est alors indexée dans une base de données, accessible ensuite par les internautes à partir de mots-clés. C'est par abus de langage qu'on appelle également « moteurs de recherche » des sites web proposant des annuaires de sites web : dans ce cas, ce sont des instruments de recherche élaborés par des personnes qui répertorient et classifient des sites web jugés dignes d'intérêt, et non des robots d'indexation. Les moteurs de recherche ne s'appliquent pas qu'à Internet : certains moteurs sont des logiciels installés sur un ordinateur personnel. Ce sont des moteurs dits « de bureau » qui combinent la recherche parmi les fichiers stockés sur le PC et la recherche parmi les sites Web — on peut citer par exemple Exalead Desktop, Google Desktop et Copernic Desktop Search, Windex Server, etc. On trouve également des métamoteurs, c'est-à-dire des sites web où une même recherche est lancée simultanément sur plusieurs moteurs de recherche, les résultats étant ensuite fusionnés pour être présentés à l'internaute. On peut citer dans cette catégorie Ixquick, Mamma, Kartoo, Framabee ou Lilo.",
          "timestamp": "2018-05-30T15:15:00Z",
          "title_fr": "Moteur de recherche"
        },
        "id": "frwiki_1",
        "index": "/tmp/blast/node1/index",
        "locations": {
          "text_fr": {
            "search": [
              {
                "array_positions": null,
                "end": 1662,
                "pos": 253,
                "start": 1656
              }
            ]
          }
        },
        "score": 0.14418354794511895,
        "sort": [
          "_score"
        ]
      }
    ],
    "max_score": 0.633816551223398,
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
    "took": 1754440,
    "total_hits": 12
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
                    --raft-node-id=node2 \
                    --raft-dir=/tmp/blast/node2/raft \
                    --store-dir=/tmp/blast/node2/store \
                    --index-dir=/tmp/blast/node2/index \
                    --peer-grpc-addr=localhost:10001 \
                    --index-mapping-file=./etc/index_mapping.json
$ ./bin/blast start --bind-addr=localhost:12000 \
                    --grpc-addr=localhost:12001 \
                    --http-addr=localhost:12002 \
                    --raft-node-id=node3 \
                    --raft-dir=/tmp/blast/node3/raft \
                    --store-dir=/tmp/blast/node3/store \
                    --index-dir=/tmp/blast/node3/index \
                    --peer-grpc-addr=localhost:10001 \
                    --index-mapping-file=./etc/index_mapping.json
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
$ cat ./example/doc_enwiki_1.json | xargs -0 ./bin/blast put --grpc-addr=localhost:10001 --pretty-print enwiki_1
```

You can see the result in JSON format. The result of the above command is:

```json
{
  "success": true
}
```

So, you can get a document from node1 like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:10001 --pretty-print enwiki_1
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
  },
  "success": true
}
```

Also, you can get same document from node2 (localhost:11001) like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:11001 --pretty-print enwiki_1
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
  },
  "success": true
}
```

Lastly, you can get same document from node3 (localhost:12001) like following:

```bash
$ ./bin/blast get --grpc-addr=localhost:12001 --pretty-print enwiki_1
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
$ docker pull mosuka/blast:latest
```

See https://hub.docker.com/r/mosuka/blast/tags/

### Running Blast node on Docker

Running a Blast node on Docker. Start Blast node like so:

```bash
$ docker run --rm --name blast1 \
    -p 10000:10000 \
    -p 10001:10001 \
    -p 10002:10002 \
    mosuka/blast:latest start \
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
