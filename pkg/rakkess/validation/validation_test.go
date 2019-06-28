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

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:   "valid format",
			format: "icon-table",
		},
		{
			name:     "invalid format",
			format:   "cassowary",
			expected: "unexpected output format: cassowary",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := OutputFormat(test.format)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}

func TestVerbs(t *testing.T) {
	tests := []struct {
		name     string
		verbs    []string
		expected string
	}{
		{
			name:  "only valid verbs",
			verbs: []string{"list", "get", "deletecollection"},
		},
		{
			name:     "only invalid verbs",
			verbs:    []string{"lust", "git", "poxy"},
			expected: "unexpected verbs: [git lust poxy]",
		},
		{
			name:     "valid and invalid verbs",
			verbs:    []string{"list", "git", "deletecollection"},
			expected: "unexpected verbs: [git]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := verbs(test.verbs)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}
