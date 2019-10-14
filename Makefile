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

GOOS ?=
GOARCH ?=
GO111MODULE ?= on
CGO_ENABLED ?= 0
CGO_CFLAGS ?=
CGO_LDFLAGS ?=
BUILD_TAGS ?=
BIN_EXT ?=
VERSION ?=
DOCKER_REPOSITORY ?= mosuka

PACKAGES = $(shell $(GO) list ./... | grep -v '/vendor/')

PROTOBUFS = $(shell find . -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq | grep -v /vendor/)

TARGET_PACKAGES = $(shell find . -name 'main.go' -print0 | xargs -0 -n1 dirname | sort | uniq | grep -v /vendor/)

GRPC_GATEWAY_PATH = $(shell $(GO) list -m -f "{{.Dir}}" github.com/grpc-ecosystem/grpc-gateway)

ifeq ($(GOOS),)
  GOOS = $(shell go version | awk -F ' ' '{print $$NF}' | awk -F '/' '{print $$1}')
endif

ifeq ($(GOARCH),)
  GOARCH = $(shell go version | awk -F ' ' '{print $$NF}' | awk -F '/' '{print $$2}')
endif

ifeq ($(VERSION),)
  VERSION = latest
endif
LDFLAGS = -ldflags "-s -w -X \"github.com/mosuka/blast/version.Version=$(VERSION)\""

ifeq ($(GOOS),windows)
  BIN_EXT = .exe
endif

GO := GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) GO111MODULE=$(GO111MODULE) go

.DEFAULT_GOAL := build

.PHONY: clean
clean:
	@echo ">> cleaning binaries"
	rm -rf ./bin
	rm -rf ./data
	rm -rf ./dist

.PHONY: echo-env
echo-env:
	@echo ">> echo environment variables"
	@echo "   GOOS              = $(GOOS)"
	@echo "   GOARCH            = $(GOARCH)"
	@echo "   GO111MODULE       = $(GO111MODULE)"
	@echo "   CGO_ENABLED       = $(CGO_ENABLED)"
	@echo "   CGO_CFLAGS        = $(CGO_CFLAGS)"
	@echo "   CGO_LDFLAGS       = $(CGO_LDFLAGS)"
	@echo "   BUILD_TAGS        = $(BUILD_TAGS)"
	@echo "   BIN_EXT           = $(BIN_EXT)"
	@echo "   VERSION           = $(VERSION)"
	@echo "   DOCKER_REPOSITORY = $(DOCKER_REPOSITORY)"
	@echo "   PACKAGES          = $(PACKAGES)"
	@echo "   PROTOBUFS         = $(PROTOBUFS)"
	@echo "   TARGET_PACKAGES   = $(TARGET_PACKAGES)"
	@echo "   LDFLAGS           = $(LDFLAGS)"
	@echo "   GRPC_GATEWAY_PATH = $(GRPC_GATEWAY_PATH)"

.PHONY: format
format:
	@echo ">> formatting code"
	@$(GO) fmt $(PACKAGES)

.PHONY: protoc
protoc: echo-env
	@echo ">> generating proto3 code"
	@echo "   GRPC_GATEWAY_PATH = $(GRPC_GATEWAY_PATH)"
	@for proto_dir in $(PROTOBUFS); do echo $$proto_dir; protoc --proto_path=. --proto_path=${GRPC_GATEWAY_PATH} --proto_path=${GRPC_GATEWAY_PATH}/third_party/googleapis --proto_path=$$proto_dir --go_out=plugins=grpc:$(GOPATH)/src $$proto_dir/*.proto || exit 1; done
	@for proto_dir in $(PROTOBUFS); do echo $$proto_dir; protoc --proto_path=. --proto_path=${GRPC_GATEWAY_PATH} --proto_path=${GRPC_GATEWAY_PATH}/third_party/googleapis --proto_path=$$proto_dir --grpc-gateway_out=logtostderr=true,allow_delete_body=true:$(GOPATH)/src $$proto_dir/*.proto || exit 1; done
	@for proto_dir in $(PROTOBUFS); do echo $$proto_dir; protoc --proto_path=. --proto_path=${GRPC_GATEWAY_PATH} --proto_path=${GRPC_GATEWAY_PATH}/third_party/googleapis --proto_path=$$proto_dir --swagger_out=logtostderr=true,allow_delete_body=true:. $$proto_dir/*.proto || exit 1; done

.PHONY: test
test: echo-env
	@echo ">> testing all packages"
	@$(GO) test -v -tags="$(BUILD_TAGS)" $(PACKAGES)

.PHONY: coverage
coverage: echo-env
	@echo ">> checking coverage of all packages"
	$(GO) test -coverprofile=./cover.out -tags="$(BUILD_TAGS)" $(PACKAGES)
	$(GO) tool cover -html=cover.out -o cover.html

.PHONY: build
build: echo-env
	@echo ">> building binaries"
	for target_pkg in $(TARGET_PACKAGES); do echo $$target_pkg; $(GO) build -tags="$(BUILD_TAGS)" $(LDFLAGS) -o ./bin/`basename $$target_pkg`$(BIN_EXT) $$target_pkg || exit 1; done

.PHONY: install
install: echo-env
	@echo ">> installing binaries"
	for target_pkg in $(TARGET_PACKAGES); do echo $$target_pkg; $(GO) install -tags="$(BUILD_TAGS)" $(LDFLAGS) $$target_pkg || exit 1; done

.PHONY: dist
dist: echo-env
	@echo ">> packaging binaries"
	mkdir -p ./dist/$(GOOS)-$(GOARCH)/bin
	for target_pkg in $(TARGET_PACKAGES); do echo $$target_pkg; $(GO) build -tags="$(BUILD_TAGS)" $(LDFLAGS) -o ./dist/$(GOOS)-$(GOARCH)/bin/`basename $$target_pkg`$(BIN_EXT) $$target_pkg || exit 1; done
	(cd ./dist/$(GOOS)-$(GOARCH); tar zcfv ../blast-${VERSION}.$(GOOS)-$(GOARCH).tar.gz .)

.PHONY: git-tag
git-tag: echo-env
	@echo ">> tagging github"
ifeq ($(VERSION),$(filter $(VERSION),latest master ""))
	@echo "please specify VERSION"
else
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
endif

.PHONY: docker-build
docker-build: echo-env
	@echo ">> building docker container image"
	docker build -t $(DOCKER_REPOSITORY)/blast:latest --build-arg VERSION=$(VERSION) .
	docker tag $(DOCKER_REPOSITORY)/blast:latest $(DOCKER_REPOSITORY)/blast:$(VERSION)

.PHONY: docker-push
docker-push: echo-env
	@echo ">> pushing docker container image"
	docker push $(DOCKER_REPOSITORY)/blast:latest
	docker push $(DOCKER_REPOSITORY)/blast:$(VERSION)

.PHONY: docker-pull
docker-pull: echo-env
	@echo ">> pulling docker container image"
	docker pull $(DOCKER_REPOSITORY):$(VERSION)
