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

VERSION ?=
GOOS = linux
GOARCH = amd64
BUILD_TAGS = "kagome"

GO := CGO_ENABLED=0 GO15VENDOREXPERIMENT=1 go

PACKAGES = $(shell $(GO) list ./... | grep -v '/vendor/')

PROTOBUFS = $(shell find . -name '*.proto' -print0 | xargs -0 -n1 dirname | sort --unique | grep -v /vendor/)

TARGET_PACKAGES = $(shell find . -name 'main.go' -print0 | xargs -0 -n1 dirname | sort --unique | grep -v /vendor/)

ifeq ($(VERSION),)
  VERSION = $(shell cat ./Versionfile)
endif
LDFLAGS = -ldflags "-X \"github.com/mosuka/blast/version.Version=${VERSION}\""


.DEFAULT_GOAL := build

.PHONY: init-deps
init-deps:
	@echo ">> initialize dependencies"
	dep init

.PHONY: install-deps
install-deps:
	@echo ">> initialize dependencies"
	dep ensure

.PHONY: update-deps
update-deps:
	@echo ">> initialize dependencies"
	dep ensure -update

.PHONY: protoc
protoc:
	@echo ">> generating proto3 code"
	@for proto_dir in $(PROTOBUFS); do echo $$proto_dir; protoc --proto_path=$$proto_dir --proto_path=./vendor/ --go_out=plugins=grpc:$$proto_dir $$proto_dir/*.proto || exit 1; done

.PHONY: format
format:
	@echo ">> formatting code"
	@$(GO) fmt $(PACKAGES)

.PHONY: test
test:
	@echo ">> testing all packages"
	@echo "   VERSION    = $(VERSION)"
	@echo "   BUILD_TAGS = $(BUILD_TAGS)"
	@$(GO) test -v -tags=${BUILD_TAGS} ${LDFLAGS} $(PACKAGES)

.PHONY: build
build:
	@echo ">> building binaries"
	@echo "   VERSION    = $(VERSION)"
	@echo "   GOOS       = $(GOOS)"
	@echo "   GOARCH     = $(GOARCH)"
	@echo "   BUILD_TAGS = $(BUILD_TAGS)"
	@for target_pkg in $(TARGET_PACKAGES); do echo $$target_pkg; GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO) build -tags=${BUILD_TAGS} ${LDFLAGS} -o ./bin/`basename $$target_pkg` $$target_pkg || exit 1; done

.PHONY: install
install:
	@echo ">> installing binaries"
	@echo "   VERSION    = $(VERSION)"
	@echo "   GOOS       = $(GOOS)"
	@echo "   GOARCH     = $(GOARCH)"
	@echo "   BUILD_TAGS = $(BUILD_TAGS)"
	@for target_pkg in $(TARGET_PACKAGES); do echo $$target_pkg; GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO) install -tags=${BUILD_TAGS} ${LDFLAGS} $$target_pkg || exit 1; done

.PHONY: dist
dist:
	@echo ">> packaging binaries"
	@echo "   VERSION = $(VERSION)"
	@echo "   GOOS    = $(GOOS)"
	@echo "   GOARCH  = $(GOARCH)"
	mkdir -p ./dist
	tar zcfv ./dist/blast-${VERSION}.$(GOOS)-$(GOARCH).tar.gz ./bin/* -C ./bin

.PHONY: docker
docker:
	@echo ">> building docker container image"
	@echo "   VERSION      = $(VERSION)"
	@echo "   GOOS         = $(GOOS)"
	@echo "   GOARCH       = $(GOARCH)"
	docker build -t blast:${VERSION} --build-arg OS=$(GOOS) --build-arg ARCH=$(GOARCH) --build-arg VERSION=${VERSION} .

.PHONY: clean
clean:
	@echo ">> cleaning binaries"
	rm -rf ./bin
	rm -rf ./dist
