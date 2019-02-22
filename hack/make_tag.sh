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

echo -n "Please enter the old tag (maybe $(git describe HEAD --always --tag | sed 's:-.*::')): "
read -r OLD_TAG

echo -n "Please enter the new tag: "
read -r NEW_TAG


find . -type f -not \( -path './hack/*' -o -path './.github/*' -o -path './out/*' -o -path './doc/releases/*' \) -print0 |
    xargs -0 sed -i "s:${OLD_TAG}:${NEW_TAG}:g"

echo "Please check what has changed:"
echo "-------------------->8--------------------"
git diff
echo "-------------------->8--------------------"

echo -n "looks good? (y/n)  "
read -r YESNO

if [[ "${YESNO}" != "y" ]]
then
    echo "Aborting"
    exit 1
fi

git commit -a -m "Create tag ${NEW_TAG}"
git tag "${NEW_TAG}"
