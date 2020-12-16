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
OUT_DIR = bin
BINARY = skywalking-satellite

RELEASE_BIN = skywalking-satellite-$(VERSION)-bin
RELEASE_SRC = skywalking-satellite-$(VERSION)-src

OS = $(shell uname)

GO = go
GO_PATH = $$($(GO) env GOPATH)
GO_BUILD = $(GO) build
GO_GET = $(GO) get
GO_TEST = $(GO) test
GO_LINT = $(GO_PATH)/bin/golangci-lint
GO_LICENSER = $(GO_PATH)/bin/go-licenser
GO_PACKR = $(GO_PATH)/bin/packr2
GO_BUILD_FLAGS = -v
GO_BUILD_LDFLAGS = -X main.version=$(VERSION)
GQL_GEN = $(GO_PATH)/bin/gqlgen

PLATFORMS := windows linux darwin
os = $(word 1, $@)
ARCH = amd64

SHELL = /bin/bash

all: clean license deps lint test build

tools:
	$(GO_PACKR) -v || $(GO_GET) -u github.com/gobuffalo/packr/v2/...
	$(GO_LINT) version || curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GO_PATH)/bin v1.33.0
	$(GO_LICENSER) -version || GO111MODULE=off $(GO_GET) -u github.com/elastic/go-licenser

deps: tools
	$(GO_GET) -v -t -d ./...

.PHONY: lint
lint: tools
	$(GO_LINT) run -v ./...

.PHONE: test
test: clean lint
	$(GO_TEST) ./... -coverprofile=coverage.txt -covermode=atomic

.PHONY: license
license: clean tools
	$(GO_LICENSER) -d -licensor='Apache Software Foundation (ASF)' .

.PHONY: verify
verify: clean license lint test

.PHONY: clean
clean: tools
	-rm -rf coverage.txt

.PHONY: build
build: deps windows linux darwin


.PHONY: $(PLATFORMS)
$(PLATFORMS):
	mkdir -p $(OUT_DIR)
	GOOS=$(os) GOARCH=$(ARCH) $(GO_BUILD) $(GO_BUILD_FLAGS) -ldflags "$(GO_BUILD_LDFLAGS)" -o $(OUT_DIR)/$(BINARY)-$(VERSION)-$(os)-$(ARCH) ./cmd
