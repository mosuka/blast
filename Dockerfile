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

FROM golang:1.10.3

ARG VERSION=
ARG BUILD_TAGS=

COPY . /go/src/github.com/mosuka/blast

RUN apt-get update && \
    apt-get install -y iproute netcat lsof jq libxml2-utils xmlstarlet tar && \
    apt-get clean && \
    cd /go/src/github.com/mosuka/blast && \
    make GOOS=linux GOARCH=amd64 VERSION=${VERSION} BUILD_TAG="${BUILD_TAGS}" build


FROM alpine:3.8

MAINTAINER Minoru Osuka "minoru.osuka@gmail.com"

RUN apk --no-cache update

COPY --from=0 /go/src/github.com/mosuka/blast/bin/blast /usr/bin/blast

EXPOSE 10000 10001 10002

ENTRYPOINT [ "/usr/bin/blast" ]
CMD        [ "--help" ]
