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

package options

import (
	"testing"

	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
	"github.com/stretchr/testify/assert"
)

func TestRakkessOptions_ExpandVerbs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "* wildcard",
			input:    []string{"*"},
			expected: constants.ValidVerbs,
		},
		{
			name:     "all wildcard",
			input:    []string{"*"},
			expected: constants.ValidVerbs,
		},
		{
			name:     "wildcard mixed with other verbs",
			input:    []string{"list", "*", "get"},
			expected: constants.ValidVerbs,
		},
		{
			name:     "no wildcard",
			input:    []string{"list", "get"},
			expected: []string{"list", "get"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := &RakkessOptions{Verbs: test.input}
			opts.ExpandVerbs()

			assert.Equal(t, test.expected, opts.Verbs)
		})
	}
}
