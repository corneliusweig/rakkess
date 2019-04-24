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
	"k8s.io/cli-runtime/pkg/genericclioptions"
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

func TestRakkessOptions_ExpandServiceAccount(t *testing.T) {
	tests := []struct {
		name           string
		serviceAccount string
		namespace      string
		impersonate    string
		expected       string
		expectedErr    string
	}{
		{
			name:        "no serviceAccount given",
			impersonate: "original-impersonate",
			expected:    "original-impersonate",
		},
		{
			name:           "unqualified serviceAccount and namespace",
			serviceAccount: "some-sa",
			namespace:      "some-ns",
			expected:       "system:serviceaccount:some-ns:some-sa",
		},
		{
			name:           "qualified serviceAccount",
			serviceAccount: "some-ns:some-sa",
			expected:       "system:serviceaccount:some-ns:some-sa",
		},
		{
			name:           "unqualified serviceAccount without namespace",
			serviceAccount: "some-ns",
			expectedErr:    "fully qualify the serviceAccount",
		},
		{
			name:           "qualified serviceAccount and impersonate",
			serviceAccount: "some-ns",
			impersonate:    "other-impersonatino",
			expectedErr:    "--sa cannot be mixed with --as",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := &RakkessOptions{
				ConfigFlags: &genericclioptions.ConfigFlags{
					Impersonate: &test.impersonate,
					Namespace:   &test.namespace,
				},
				AsServiceAccount: test.serviceAccount,
			}

			err := opts.ExpandServiceAccount()
			if test.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErr)
			} else {
				assert.Equal(t, test.expected, *opts.ConfigFlags.Impersonate)
			}
		})
	}
}
