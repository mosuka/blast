# Copyright (c) 2019 Minoru Osuka
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# 		http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.12.1-stretch

ARG VERSION

ENV GOPATH /go

COPY . ${GOPATH}/src/github.com/mosuka/blast

RUN echo "deb http://ftp.us.debian.org/debian/ jessie main contrib non-free" >> /etc/apt/sources.list && \
    echo "deb-src http://ftp.us.debian.org/debian/ jessie main contrib non-free" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get install -y \
      git \
      golang \
      libicu-dev \
      libstemmer-dev \
      libleveldb-dev \
      gcc-4.8 \
      g++-4.8 \
      build-essential && \
    apt-get clean && \
    update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-6 80 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-6 80 && \
    update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 90 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-4.8 90 && \
    go get -u -v github.com/blevesearch/cld2 && \
    cd ${GOPATH}/src/github.com/blevesearch/cld2 && \
    git clone https://github.com/CLD2Owners/cld2.git && \
    cd cld2/internal && \
    ./compile_libs.sh && \
    cp *.so /usr/local/lib && \
    cd ${GOPATH}/src/github.com/mosuka/blast && \
    make \
      GOOS=linux \
      GOARCH=amd64 \
      CGO_ENABLED=1 \
      BUILD_TAGS="kagome icu libstemmer cld2 cznicb leveldb badger" \
      VERSION="${VERSION}" \
      build

FROM debian:stretch-slim

MAINTAINER Minoru Osuka "minoru.osuka@gmail.com"

RUN apt-get update && \
    apt-get install -y \
      libicu-dev \
      libstemmer-dev \
      libleveldb-dev && \
    apt-get clean

COPY --from=0 /go/src/github.com/blevesearch/cld2/cld2/internal/*.so /usr/local/lib/
COPY --from=0 /go/src/github.com/mosuka/blast/bin/* /usr/bin/
COPY --from=0 /go/src/github.com/mosuka/blast/docker-entrypoint.sh /usr/bin/

EXPOSE 5050 6060 8080

ENTRYPOINT [ "/usr/bin/docker-entrypoint.sh" ]
CMD        [ "blast-indexer", "--help" ]
