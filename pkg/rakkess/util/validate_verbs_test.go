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

package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateVerbs(t *testing.T) {
	tests := []struct {
		name     string
		verbs    []string
		expected string
	}{
		{
			name:  "only valid verbs",
			verbs: []string{"list", "get", "proxy"},
		},
		{
			name:     "only invalid verbs",
			verbs:    []string{"lust", "git", "poxy"},
			expected: "unexpected verbs: [git lust poxy]",
		},
		{
			name:     "valid and invalid verbs",
			verbs:    []string{"list", "git", "proxy"},
			expected: "unexpected verbs: [git]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := ValidateVerbs(test.verbs)
			if test.expected != "" {
				assert.EqualError(t, actual, test.expected)
			} else {
				assert.NoError(t, actual)
			}
		})
	}
}
