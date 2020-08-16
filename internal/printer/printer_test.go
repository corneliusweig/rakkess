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

package printer

import (
	"bytes"
	"io"
	"testing"

	"github.com/corneliusweig/rakkess/internal/client/result"
	"github.com/stretchr/testify/assert"
)

type accessResult map[string]result.Access

func buildAccess() accessResult {
	return make(map[string]result.Access)
}
func (a accessResult) withResult(result result.Access, verbs ...string) accessResult {
	for _, v := range verbs {
		a[v] = result
	}
	return a
}
func (a accessResult) allowed(verbs ...string) accessResult {
	return a.withResult(result.AccessAllowed, verbs...)
}
func (a accessResult) denied(verbs ...string) accessResult {
	return a.withResult(result.AccessDenied, verbs...)
}
func (a accessResult) get() map[string]result.Access {
	return a
}

const HEADER = "NAME       GET  LIST\n"

func TestPrintResults(t *testing.T) {
	tests := []struct {
		name          string
		verbs         []string
		given         result.ResourceAccess
		expected      string
		expectedColor string
		expectedASCII string
	}{
		{
			"single result, all allowed",
			[]string{"get", "list"},
			[]result.ResourceAccessItem{{Name: "resource1", Access: buildAccess().allowed("get", "list").get()}},
			HEADER + "resource1  ✔    ✔\n",
			HEADER + "resource1  \033[32m✔\033[0m    \033[32m✔\033[0m\n",
			HEADER + "resource1  yes  yes\n",
		},
		{
			"single result, all forbidden",
			[]string{"get", "list"},
			[]result.ResourceAccessItem{
				{Name: "resource1", Access: buildAccess().denied("get", "list").get()},
			},
			HEADER + "resource1  ✖    ✖\n",
			HEADER + "resource1  \033[31m✖\033[0m    \033[31m✖\033[0m\n",
			HEADER + "resource1  no   no\n",
		},
		{
			"single result, all not applicable",
			[]string{"get", "list"},
			[]result.ResourceAccessItem{
				{Name: "resource1", Access: buildAccess().withResult(result.AccessNotApplicable, "get", "list").get()},
			},
			HEADER + "resource1       \n",
			HEADER + "resource1  \033[0m\033[0m     \033[0m\033[0m\n",
			HEADER + "resource1  n/a  n/a\n",
		},
		{
			"single result, all ERR",
			[]string{"get", "list"},
			[]result.ResourceAccessItem{
				{Name: "resource1", Access: buildAccess().withResult(result.AccessRequestErr, "get", "list").get()},
			},
			HEADER + "resource1  ERR  ERR\n",
			HEADER + "resource1  \033[35mERR\033[0m  \033[35mERR\033[0m\n",
			HEADER + "resource1  ERR  ERR\n",
		},
		{
			"single result, mixed",
			[]string{"get", "list"},
			[]result.ResourceAccessItem{
				{Name: "resource1", Access: buildAccess().allowed("list").denied("get").get()},
			},
			HEADER + "resource1  ✖    ✔\n",
			"",
			HEADER + "resource1  no   yes\n",
		},
		{
			"many results",
			[]string{"get"},
			[]result.ResourceAccessItem{
				{Name: "resource1", Access: buildAccess().denied("get").get()},
				{Name: "resource2", Access: buildAccess().allowed("get").get()},
				{Name: "resource3", Access: buildAccess().denied("get").get()},
			},
			"NAME       GET\nresource1  ✖\nresource2  ✔\nresource3  ✖\n",
			"",
			"NAME       GET\nresource1  no\nresource2  yes\nresource3  no\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			PrintResults(buf, test.verbs, "icon-table", test.given)

			assert.Equal(t, test.expected, buf.String())

			buf = &bytes.Buffer{}

			PrintResults(buf, test.verbs, "ascii-table", test.given)

			assert.Equal(t, test.expectedASCII, buf.String())
		})
	}

	for _, test := range tests[0:4] {
		isTerminal = func(w io.Writer) bool {
			return true
		}
		defer func() {
			isTerminal = isTerminalImpl
		}()

		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			PrintResults(buf, test.verbs, "icon-table", test.given)

			assert.Equal(t, test.expectedColor, buf.String())

			buf = &bytes.Buffer{}

			PrintResults(buf, test.verbs, "ascii-table", test.given)

			assert.Equal(t, test.expectedASCII, buf.String())
		})
	}
}
