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
	"strings"
	"testing"

	"github.com/corneliusweig/rakkess/internal/client/result"
	"github.com/stretchr/testify/assert"
)

const HEADER = "NAME       GET  LIST\n"

func parseInput(t *testing.T, in []string) result.ResourceAccess {
	ret := make(result.ResourceAccess)
	for _, s := range in {
		parts := strings.Split(s, ":")
		if len(parts) != 3 {
			t.Errorf("parseInput got %q, want format 'name:verb1:result'", s)
			return nil
		}
		name := parts[0]
		verb := parts[1]
		var access result.Access
		switch parts[2] {
		case "n/a":
			access = result.AccessNotApplicable
		case "no":
			access = result.AccessDenied
		case "ok":
			access = result.AccessAllowed
		case "err":
			access = result.AccessRequestErr
		}

		r := ret[name]
		if r == nil {
			r = make(map[string]result.Access)
			ret[name] = r
		}
		r[verb] = access
	}
	return ret
}

func TestPrintResults(t *testing.T) {
	tests := []struct {
		name      string
		verbs     []string
		input     []string // for example: "name:verb1,verb2"
		want      string
		wantColor string
		wantASCII string
	}{
		{
			"single result, all allowed",
			[]string{"get", "list"},
			[]string{"resource1:get:ok", "resource1:list:ok"},
			HEADER + "resource1  ✔    ✔\n",
			HEADER + "resource1  \033[32m✔\033[0m    \033[32m✔\033[0m\n",
			HEADER + "resource1  yes  yes\n",
		},
		{
			"single result, all forbidden",
			[]string{"get", "list"},
			[]string{"resource1:get:no", "resource1:list:no"},
			HEADER + "resource1  ✖    ✖\n",
			HEADER + "resource1  \033[31m✖\033[0m    \033[31m✖\033[0m\n",
			HEADER + "resource1  no   no\n",
		},
		{
			"single result, all not applicable",
			[]string{"get", "list"},
			[]string{"resource1:get:n/a", "resource1:list:n/a"},
			HEADER + "resource1       \n",
			HEADER + "resource1  \033[0m\033[0m     \033[0m\033[0m\n",
			HEADER + "resource1  n/a  n/a\n",
		},
		{
			"single result, all ERR",
			[]string{"get", "list"},
			[]string{"resource1:get:err", "resource1:list:err"},
			HEADER + "resource1  ERR  ERR\n",
			HEADER + "resource1  \033[35mERR\033[0m  \033[35mERR\033[0m\n",
			HEADER + "resource1  ERR  ERR\n",
		},
		{
			"single result, mixed",
			[]string{"get", "list"},
			[]string{"resource1:get:no", "resource1:list:ok"},
			HEADER + "resource1  ✖    ✔\n",
			"",
			HEADER + "resource1  no   yes\n",
		},
		{
			"many results",
			[]string{"get"},
			[]string{"resource1:get:no", "resource2:get:ok", "resource3:get:no"},
			"NAME       GET\nresource1  ✖\nresource2  ✔\nresource3  ✖\n",
			"",
			"NAME       GET\nresource1  no\nresource2  yes\nresource3  no\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := parseInput(t, test.input)

			buf := &bytes.Buffer{}
			PrintResults(buf, test.verbs, "icon-table", in)
			assert.Equal(t, test.want, buf.String())

			buf = &bytes.Buffer{}
			PrintResults(buf, test.verbs, "ascii-table", in)
			assert.Equal(t, test.wantASCII, buf.String())
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
			in := parseInput(t, test.input)

			buf := &bytes.Buffer{}
			PrintResults(buf, test.verbs, "icon-table", in)
			assert.Equal(t, test.wantColor, buf.String())

			buf = &bytes.Buffer{}
			PrintResults(buf, test.verbs, "ascii-table", in)
			assert.Equal(t, test.wantASCII, buf.String())
		})
	}
}
