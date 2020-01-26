/*
Copyright 2020 Cornelius Weig

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

package result

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortResult(t *testing.T) {
	makeResult := func(key string, value int) map[string]int {
		result := make(map[string]int)
		result[key] = value
		return result
	}
	tests := []struct {
		name   string
		input  []ResourceAccessItem
		sorted []ResourceAccessItem
	}{
		{
			name:   "two inputs",
			input:  []ResourceAccessItem{{Name: "b second"}, {Name: "a first"}},
			sorted: []ResourceAccessItem{{Name: "a first"}, {Name: "b second"}},
		},
		{
			name:   "three inputs",
			input:  []ResourceAccessItem{{Name: "b second"}, {Name: "c third"}, {Name: "a first"}},
			sorted: []ResourceAccessItem{{Name: "a first"}, {Name: "b second"}, {Name: "c third"}},
		},
		{
			name: "three inputs, stable",
			input: []ResourceAccessItem{
				{Name: "same", Access: makeResult("b", 1)},
				{Name: "same", Access: makeResult("a", 2)},
				{Name: "same", Access: makeResult("c", 3)},
			},
			sorted: []ResourceAccessItem{
				{Name: "same", Access: makeResult("b", 1)},
				{Name: "same", Access: makeResult("a", 2)},
				{Name: "same", Access: makeResult("c", 3)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := NewResourceAccess(test.input)
			assert.Equal(t, test.sorted, []ResourceAccessItem(actual))
		})
	}
}

func TestResourceAccess_Print(t *testing.T) {
	evenYesOddNo := func(i int) string {
		if i%2 == 0 {
			return "yes"
		}
		return "no"
	}
	tests := []struct {
		name     string
		items    ResourceAccess
		verbs    []string
		expected string
	}{
		{
			name: "single row with multiple verbs",
			items: []ResourceAccessItem{
				{
					Name:   "resource1",
					Access: map[string]int{"list": 1, "get": 2, "delete": 3},
				},
			},
			verbs:    []string{"list", "get", "delete"},
			expected: "NAME\tLIST\tGET\tDELETE\nresource1\tno\tyes\tno\n",
		},
		{
			name: "single row with ignored verbs",
			items: []ResourceAccessItem{
				{
					Name:   "resource1",
					Access: map[string]int{"list": 1, "get": 2, "delete": 3},
				},
			},
			verbs:    []string{"list"},
			expected: "NAME\tLIST\nresource1\tno\n",
		},
		{
			name: "multiple rows",
			items: []ResourceAccessItem{
				{
					Name:   "resource1",
					Access: map[string]int{"list": 1},
				},
				{
					Name:   "resource2",
					Access: map[string]int{"list": 1},
				},
				{
					Name:   "resource3",
					Access: map[string]int{"list": 2},
				},
			},
			verbs:    []string{"list"},
			expected: "NAME\tLIST\nresource1\tno\nresource2\tno\nresource3\tyes\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			test.items.Print(buf, evenYesOddNo, test.verbs)

			assert.Equal(t, test.expected, buf.String())
		})
	}
}
