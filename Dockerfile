#  Copyright (c) 2018 Minoru Osuka
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

#FROM golang:1.10.3
#
#ARG VERSION=
#ARG BUILD_TAGS=
#
#COPY . /go/src/github.com/mosuka/blast
#
#RUN apt-get update && \
#    apt-get install -y iproute netcat lsof jq libxml2-utils xmlstarlet tar && \
#    apt-get clean && \
#    cd /go/src/github.com/mosuka/blast && \
#    make GOOS=linux GOARCH=amd64 VERSION=${VERSION} BUILD_TAG="${BUILD_TAGS}" build
#
#
#FROM alpine:3.8
#
#MAINTAINER Minoru Osuka "minoru.osuka@gmail.com"
#
#RUN apk --no-cache update
#
#COPY --from=0 /go/src/github.com/mosuka/blast/bin/blast /usr/bin/blast
#
#EXPOSE 10000 10001 10002
#
#ENTRYPOINT [ "/usr/bin/blast" ]
#CMD        [ "--help" ]

FROM ubuntu:18.10

ARG VERSION=
ARG BUILD_TAGS=
ARG CGO_ENABLED=0

ENV GOPATH /go

COPY . /go/src/github.com/mosuka/blast

RUN apt-get update && \
    apt-get install -y git \
                       golang \
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
                       build-essential && \
    apt-get clean && \

    update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 80 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-8 80 && \
    update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 90 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-4.8 90 && \

    go get -u -v github.com/blevesearch/cld2 && \
    cd $GOPATH/src/github.com/blevesearch/cld2 && \
    git clone https://github.com/CLD2Owners/cld2.git && \
    cd cld2/internal && \
    ./compile_libs.sh && \
    cp *.so /usr/local/lib && \

    cd /go/src/github.com/mosuka/blast && \
    make GOOS=linux GOARCH=amd64 VERSION=${VERSION} CGO_ENABLED=${CGO_ENABLED} BUILD_TAG="${BUILD_TAGS}" build

FROM ubuntu:18.10

MAINTAINER Minoru Osuka "minoru.osuka@gmail.com"

RUN apt-get update && \
    apt-get install -y libicu-dev \
                       libleveldb-dev \
                       libstemmer-dev \
                       libgflags-dev \
                       libsnappy-dev \
                       zlib1g-dev \
                       libbz2-dev \
                       liblz4-dev \
                       libzstd-dev \
                       librocksdb-dev && \
    apt-get clean

COPY --from=0 /go/src/github.com/blevesearch/cld2/cld2/internal/*.so /usr/local/lib
COPY --from=0 /go/src/github.com/mosuka/blast/bin/blast /usr/bin/blast

EXPOSE 10000 10001 10002

ENTRYPOINT [ "/usr/bin/blast" ]
CMD        [ "--help" ]
