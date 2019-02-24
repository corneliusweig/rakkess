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

package client

import (
	"sort"
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
		input  []Result
		sorted []Result
	}{
		{
			name:   "two inputs",
			input:  []Result{{Name: "b second"}, {Name: "a first"}},
			sorted: []Result{{Name: "a first"}, {Name: "b second"}},
		},
		{
			name:   "three inputs",
			input:  []Result{{Name: "b second"}, {Name: "c third"}, {Name: "a first"}},
			sorted: []Result{{Name: "a first"}, {Name: "b second"}, {Name: "c third"}},
		},
		{
			name: "three inputs, stable",
			input: []Result{
				{Name: "same", Access: makeResult("b", 1)},
				{Name: "same", Access: makeResult("a", 2)},
				{Name: "same", Access: makeResult("c", 3)},
			},
			sorted: []Result{
				{Name: "same", Access: makeResult("b", 1)},
				{Name: "same", Access: makeResult("a", 2)},
				{Name: "same", Access: makeResult("c", 3)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sort.Stable(sortableResult(test.input))
			assert.Equal(t, test.sorted, test.input)
		})
	}
}
