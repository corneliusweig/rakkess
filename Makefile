# Copyright 2019 Cornelius Weig
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export GO111MODULE ?= on
export GOARCH      ?= amd64
export CGO_ENABLED ?= 0

PROJECT   ?= rakkess
REPOPATH  ?= github.com/corneliusweig/$(PROJECT)
COMMIT    := $(shell git rev-parse HEAD)
VERSION   ?= $(shell git describe --always --tags --dirty)
GOOS      ?= $(shell go env GOOS)
GOPATH    ?= $(shell go env GOPATH)

BUILDDIR  := out
PLATFORMS ?= linux windows darwin
DISTFILE  := $(BUILDDIR)/$(VERSION).tar.gz
TARGETS   := $(patsubst %,$(BUILDDIR)/$(PROJECT)-%-$(GOARCH),$(PLATFORMS))
ASSETS    := $(BUILDDIR)/rakkess-linux-$(GOARCH).gz $(BUILDDIR)/rakkess-darwin-$(GOARCH).gz $(BUILDDIR)/rakkess-windows-$(GOARCH).zip
BUNDLE    := $(BUILDDIR)/bundle.tar.gz
CHECKSUMS := $(patsubst %,%.sha256,$(ASSETS))
CHECKSUMS += $(BUNDLE).sha256

VERSION_PACKAGE := $(REPOPATH)/pkg/rakkess/version

GO_LDFLAGS :="
GO_LDFLAGS += -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -X $(VERSION_PACKAGE).buildDate=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS += -X $(VERSION_PACKAGE).gitCommit=$(COMMIT)
GO_LDFLAGS +="

GO_FILES  := $(shell find . -type f -name '*.go')

.PHONY: test
test:
	go test ./...

.PHONY: help
help:
	@echo 'Valid make targets:'
	@echo '  - all:      build binaries for all supported platforms'
	@echo '  - clean:    clean up build directory'
	@echo '  - coverage: run unit tests with coverage'
	@echo '  - deploy:   build artifacts for a new deployment'
	@echo '  - dev:      build the binary for the current platform'
	@echo '  - help:     print this help'
	@echo '  - install:  install the `rakkess` binary in your gopath'
	@echo '  - lint:     run golangci-lint
	@echo '  - test:     run unit tests'

.PHONY: coverage
coverage: $(BUILDDIR)
	go test -coverprofile=$(BUILDDIR)/coverage.txt -covermode=atomic ./...

.PHONY: all
all: $(TARGETS)

.PHONY: dev
dev: GO_FLAGS := -race
dev: CGO_ENABLED := 1
dev: $(BUILDDIR)/rakkess-linux-$(GOARCH)
	@mv $< $(PROJECT)

$(BUILDDIR)/$(PROJECT)-%-$(GOARCH): $(GO_FILES) $(BUILDDIR)
	GOOS=$* go build $(GO_FLAGS) -ldflags $(GO_LDFLAGS) -o $@ main.go

install: $(BUILDDIR)/$(PROJECT)-$(GOOS)-$(GOARCH)
	@mv -i $< $(GOPATH)/bin/$(PROJECT)

.PHONY: lint
lint:
	hack/run_lint.sh

%.zip: %
	zip $@ $<

%.gz: %
	gzip --best -k $<

$(BUNDLE): $(TARGETS)
	tar czf $(BUNDLE) -C $(BUILDDIR) $(patsubst $(BUILDDIR)/%,%,$(TARGETS))

$(BUILDDIR):
	mkdir -p "$@"

%.sha256: %
	shasum -a 256 $< > $@

.PHONY: deploy
deploy: $(CHECKSUMS) $(ASSETS)
	git archive --prefix="rakkess-$(VERSION)/" --format=tar.gz HEAD > $(DISTFILE)

.PHONY: clean
clean:
	$(RM) $(TARGETS) $(CHECKSUMS) $(DISTFILE) $(BUNDLE)
