FROM golang:1.15.6-buster

ARG VERSION

ENV GOPATH /go

COPY . ${GOPATH}/src/github.com/mosuka/blast

RUN echo "deb http://ftp.us.debian.org/debian/ jessie main contrib non-free" >> /etc/apt/sources.list && \
    echo "deb-src http://ftp.us.debian.org/debian/ jessie main contrib non-free" >> /etc/apt/sources.list && \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y \
            git \
           # golang \
            libicu-dev \
            libstemmer-dev \
            gcc-4.8 \
            g++-4.8 \
            build-essential && \
    apt-get clean && \
    #update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-6 80 && \
    #update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-6 80 && \
    update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-4.8 90 && \
    update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-4.8 90 && \
    go get -u -v github.com/blevesearch/cld2 && \
    cd ${GOPATH}/src/github.com/blevesearch/cld2 && \
    git clone https://github.com/CLD2Owners/cld2.git && \
    cd cld2/internal && \
    ./compile_libs.sh && \
    cp *.so /usr/local/lib && \
    cd ${GOPATH}/src/github.com/mosuka/blast && \
    make GOOS=linux \
         GOARCH=amd64 \
         CGO_ENABLED=1 \
         BUILD_TAGS="kagome icu libstemmer cld2" \
         VERSION="${VERSION}" \
         build

FROM debian:buster-slim

MAINTAINER Minoru Osuka "minoru.osuka@gmail.com"

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y \
            libicu-dev \
            libstemmer-dev && \
    apt-get clean && \
    rm -rf /var/cache/apk/*

COPY --from=0 /go/src/github.com/blevesearch/cld2/cld2/internal/*.so /usr/local/lib/
COPY --from=0 /go/src/github.com/mosuka/blast/bin/* /usr/bin/

EXPOSE 7000 8000 9000

ENTRYPOINT [ "/usr/bin/blast" ]
CMD        [ "start" ]
