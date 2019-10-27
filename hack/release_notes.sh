#!/usr/bin/env bash

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

set -euo pipefail

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if ! [[ -x "${DIR}/release-notes" ]]; then
  echo >&2 'Installing release-notes'
  cd "${DIR}/tools"
  GOBIN="$DIR" GO111MODULE=on go install github.com/corneliusweig/release-notes
  cd -
fi

# you can pass your github token with --token here if you run out of requests
"${DIR}/release-notes" corneliusweig rakkess

echo
echo "Thanks to all the contributors for this release: "
git log "$(git describe --tags --abbrev=0)".. --format="%aN" --reverse | sort --unique | sed 's:^:- :'
echo
