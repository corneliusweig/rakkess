#!/usr/bin/env bash

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

set -euo pipefail

HACK=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
GOPATH="$(go env GOPATH)"

if ! [[ -x "$GOPATH/bin/golangci-lint" ]]
then
   echo 'Installing golangci-lint'
   "${HACK}"/install_golangci-lint.sh -b "$GOPATH/bin" v1.30.0
fi

"$GOPATH/bin/golangci-lint" run \
		--no-config \
		-D errcheck \
		-E goconst \
		-E gocritic \
		-E goimports \
		-E golint \
		-E gosec \
		-E gosimple \
		-E interfacer \
		-E maligned \
		-E misspell \
		-E unconvert \
		-E unparam \
		-E stylecheck \
		-E staticcheck \
		-E structcheck \
		-E prealloc \
		--skip-dirs hack
