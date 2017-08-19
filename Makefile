#  Copyright (c) 2017 Minoru Osuka
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

VERSION = 0.1.0

LDFLAGS = -ldflags "-X \"github.com/mosuka/blast/version.Version=${VERSION}\""

GO := GO15VENDOREXPERIMENT=1 go
PACKAGES = $(shell $(GO) list ./... | grep -v '/vendor/')
PROTOBUFS = $(shell find . -name '*.proto' | sort --unique | grep -v /vendor/)
TARGET_PACKAGES = $(shell find . -name 'main.go' -print0 | xargs -0 -n1 dirname | sort --unique | grep -v /vendor/)
#BUILD_TAGS = "-tags=lang"
BUILD_TAGS = "-tags=''"

vendoring:
	@echo ">> vendoring dependencies"
	gvt restore

protoc:
	@echo ">> generating proto3 code"
	@for proto_file in $(PROTOBUFS); do echo $$proto_file; protoc --go_out=plugins=grpc:. $$proto_file; done

format:
	@echo ">> formatting code"
	@$(GO) fmt $(PACKAGES)

test:
	@echo ">> running all tests"
	@$(GO) test $(PACKAGES)

build:
	@echo ">> building binaries"
	@for target_pkg in $(TARGET_PACKAGES); do echo $$target_pkg; $(GO) build ${BUILD_TAGS} ${LDFLAGS} -o ./bin/`basename $$target_pkg` $$target_pkg; done

install:
	@echo ">> installing binaries"
	@for target_pkg in $(TARGET_PACKAGES); do echo $$target_pkg; $(GO) install ${BUILD_TAGS} ${LDFLAGS} $$target_pkg; done
