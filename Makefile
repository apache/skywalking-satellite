#
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

VERSION ?= latest
HUB ?= apache
OUT_DIR = bin
BINARY = skywalking-satellite

RELEASE_BIN = skywalking-satellite-$(VERSION)-bin
RELEASE_SRC = skywalking-satellite-$(VERSION)-src

PLUGIN_DOC_BASE_DIR = docs
PLUGIN_DOC_PLUGIN_DIR = /en/setup/plugins
PLUGIN_DOC_MENU = /menu.yml

OSNAME := $(if $(findstring Darwin,$(shell uname)),darwin,linux)

SH = sh
GO = go
GIT = git
PROTOC = protoc
GO_PATH = $$($(GO) env GOPATH)
GO_BUILD = $(GO) build
GO_GET = $(GO) get
GO_TEST = $(GO) test
GO_LINT = $(GO_PATH)/bin/golangci-lint
GO_BUILD_FLAGS = -v
GO_BUILD_LDFLAGS = -X main.version=$(VERSION) -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore -w -s
GO_TEST_LDFLAGS = -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn
GQL_GEN = $(GO_PATH)/bin/gqlgen

PLATFORMS := linux darwin windows
os = $(word 1, $@)
ARCH = amd64

SHELL = /bin/bash

all: deps verify build gen-docs check

.PHONY: tools
tools:
	$(GO_LINT) version || curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_PATH)/bin v1.48.0

deps: tools
	$(GO_GET) -v -t -d ./...

.PHONY: gen-docs
gen-docs: build
	$(OUT_DIR)/$(BINARY)-$(VERSION)-$(OSNAME)-$(ARCH) docs -output=$(PLUGIN_DOC_BASE_DIR) -menu=$(PLUGIN_DOC_MENU) -plugins=$(PLUGIN_DOC_PLUGIN_DIR)

.PHONY: lint
lint: tools
	$(GO_LINT) run -v --timeout 15m ./...

.PHONY: test
test: clean
	$(GO_TEST) -ldflags "$(GO_TEST_LDFLAGS)" ./... -coverprofile=coverage.txt -covermode=atomic

.PHONY: verify
verify: clean lint test

.PHONY: clean
clean: tools
	-rm -rf coverage.txt

.PHONY: build
build: clean deps linux darwin windows

.PHONY: check
check: clean
	$(GO) mod tidy > /dev/null
	@if [ ! -z "`git status -s`" ]; then \
		echo "Following files are not consistent with CI:"; \
		git status -s; \
		git diff; \
		exit 1; \
	fi

.PHONY: docker
docker:
	docker build --build-arg VERSION=$(VERSION) -t $(HUB)/skywalking-satellite:v$(VERSION) --no-cache . -f docker/Dockerfile

.PHONY: docker.push
docker.push:
	docker push $(HUB)/skywalking-satellite:v$(VERSION)

.PHONY: release
release:
	/bin/sh tools/release/create_bin_release.sh
	/bin/sh tools/release/create_source_release.sh

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p $(OUT_DIR)
	GOOS=$(os) GOARCH=$(ARCH) $(GO_BUILD) $(GO_BUILD_FLAGS) -ldflags "$(GO_BUILD_LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH) ./cmd
