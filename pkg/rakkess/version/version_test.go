/*
Copyright 2019 Cornelius Weig

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/blang/semver"
)

func TestParseVersion(t *testing.T) {
	var tests = []struct {
		name      string
		given     string
		expected  semver.Version
		shouldErr bool
	}{
		{
			name:     "parse version correct",
			given:    "v3.14.15",
			expected: semver.MustParse("3.14.15"),
		},
		{
			name:     "parse with trailing text",
			given:    "v2.71.82-dirty",
			expected: semver.MustParse("2.71.82-dirty"),
		},
		{
			name:     "fail parse without leading v",
			given:    "2.71.82",
			expected: semver.MustParse("2.71.82"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseVersion(test.given)
			if test.shouldErr {
				assert.Error(t, err, "parse should fail")
			} else {
				assert.Equal(t, test.expected, actual, "parse should succeed")

			}
		})
	}
}
