# Copyright 2020 Cornelius Weig
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
export CGO_ENABLED ?= 0

PROJECT   ?= rakkess
REPOPATH  ?= github.com/corneliusweig/$(PROJECT)
COMMIT    := $(shell git rev-parse HEAD)
VERSION   ?= $(shell git describe --always --tags --dirty)
GOOS      ?= $(shell go env GOOS)
GOPATH    ?= $(shell go env GOPATH)

BUILDDIR   := out
PLATFORMS  ?= darwin/amd64 darwin/arm64 windows/amd64 linux/amd64
DISTFILE   := $(BUILDDIR)/$(VERSION).tar.gz
ASSETS     := $(BUILDDIR)/rakkess-amd64-darwin.tar.gz \
              $(BUILDDIR)/rakkess-arm64-darwin.tar.gz \
              $(BUILDDIR)/rakkess-amd64-linux.tar.gz \
              $(BUILDDIR)/rakkess-amd64-windows.zip
ASSETSKREW := $(BUILDDIR)/access-matrix-amd64-darwin.tar.gz \
              $(BUILDDIR)/access-matrix-arm64-darwin.tar.gz \
              $(BUILDDIR)/access-matrix-amd64-linux.tar.gz \
              $(BUILDDIR)/access-matrix-amd64-windows.zip
CHECKSUMS  := $(patsubst %,%.sha256,$(ASSETS) $(ASSETSKREW))

VERSION_PACKAGE := $(REPOPATH)/internal/version

DATE_FMT = %Y-%m-%dT%H:%M:%SZ
ifdef SOURCE_DATE_EPOCH
    # GNU and BSD date require different options for a fixed date
    BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "+$(DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "+$(DATE_FMT)" 2>/dev/null)
else
    BUILD_DATE ?= $(shell date "+$(DATE_FMT)")
endif
GO_LDFLAGS :="-s -w
GO_LDFLAGS += -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -X $(VERSION_PACKAGE).buildDate=$(BUILD_DATE)
GO_LDFLAGS += -X $(VERSION_PACKAGE).gitCommit=$(COMMIT)
GO_LDFLAGS +="

ifdef ZOPFLI
  COMPRESS:=zopfli -c
else
  COMPRESS:=gzip --best -k -c
endif

define doUPX
	upx -9q $@
endef

GO_FILES  := $(shell find . -type f -name '*.go')

.PHONY: test
test:
	go test -tags accessmatrix ./...

.PHONY: help
help:
	@echo 'Valid make targets:'
	@echo '  - all:      build binaries for all supported platforms'
	@echo '  - clean:    clean up build directory'
	@echo '  - coverage: run unit tests with coverage'
	@echo '  - deploy:   build artifacts for a new deployment'
	@echo '  - dev:      build the binary for the current platform'
	@echo '  - dist:     create a tar archive of the source code'
	@echo '  - help:     print this help'
	@echo '  - lint:     run golangci-lint'
	@echo '  - test:     run unit tests'
	@echo '  - build-rakkess:        build binaries for all supported platforms'
	@echo '  - build-access-matrix:  build binaries for all supported platforms'

.PHONY: coverage
coverage: $(BUILDDIR)
	go test -coverprofile=$(BUILDDIR)/coverage.txt -covermode=atomic ./...

.PHONY: all
all: lint test dev

.PHONY: dev
dev: CGO_ENABLED := 1
dev: GO_LDFLAGS := $(subst -s -w,,$(GO_LDFLAGS))
dev:
	go build -race -ldflags $(GO_LDFLAGS) -o rakkess main.go

# TODO(corneliusweig): gox does not support the -trimpath flag, see https://github.com/mitchellh/gox/pull/138
build-rakkess: $(GO_FILES) $(BUILDDIR)
	GOFLAGS="-trimpath" gox -osarch="$(PLATFORMS)" -tags netgo -ldflags $(GO_LDFLAGS) -output="out/rakkess-{{.Arch}}-{{.OS}}"

build-access-matrix: $(GO_FILES) $(BUILDDIR)
	GOFLAGS="-trimpath" gox -osarch="$(PLATFORMS)" -tags accessmatrix,netgo -ldflags $(GO_LDFLAGS) -output="out/access-matrix-{{.Arch}}-{{.OS}}"

.PHONY: lint
lint:
	hack/run_lint.sh

.PRECIOUS: %.zip
%.zip: %.exe
	cp LICENSE $(BUILDDIR) && \
	cd $(BUILDDIR) && \
	zip $(patsubst $(BUILDDIR)/%, %, $@) LICENSE $(patsubst $(BUILDDIR)/%, %, $<)

.PRECIOUS: %.gz
%.gz: %
	$(COMPRESS) "$<" > "$@"

%.tar: %
	cp LICENSE $(BUILDDIR)
	tar cf "$@" -C $(BUILDDIR) LICENSE $(patsubst $(BUILDDIR)/%,%,$^)

$(BUILDDIR):
	mkdir -p "$@"

%.sha256: %
	shasum -a 256 $< > $@

.INTERMEDIATE: $(DISTFILE:.gz=)
$(DISTFILE:.gz=): $(BUILDDIR)
	git archive --prefix="rakkess-$(VERSION)/" --format=tar HEAD > "$@"

.PHONY: deploy
deploy: $(CHECKSUMS)
	$(RM) $(BUILDDIR)/LICENSE

.PHONY: dist
dist: $(DISTFILE)

.PHONY: clean
clean:
	$(RM) -r $(BUILDDIR) rakkess

$(BUILDDIR)/rakkess-arm64-darwin: build-rakkess
$(BUILDDIR)/rakkess-amd64-darwin: build-rakkess
$(BUILDDIR)/rakkess-amd64-linux: build-rakkess
	$(doUPX)
$(BUILDDIR)/rakkess-amd64-windows.exe: build-rakkess
	$(doUPX)

$(BUILDDIR)/access-matrix-arm64-darwin: build-access-matrix
$(BUILDDIR)/access-matrix-amd64-darwin: build-access-matrix
$(BUILDDIR)/access-matrix-amd64-linux: build-access-matrix
	$(doUPX)
$(BUILDDIR)/access-matrix-amd64-windows.exe: build-access-matrix
	$(doUPX)
