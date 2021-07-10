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

	"github.com/stretchr/testify/assert"
)

const HEADER = "NAME       GET  LIST\n"

func TestPrintResults(t *testing.T) {
	tests := []struct {
		name      string
		p         *Printer
		want      string
		wantColor string
		wantASCII string
	}{
		{
			"single result, all allowed",
			&Printer{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Up, Up}},
				},
			},
			HEADER + "resource1  ✔    ✔\n",
			HEADER + "resource1  \033[32m✔\033[0m    \033[32m✔\033[0m\n",
			HEADER + "resource1  yes  yes\n",
		},
		{
			"single result, all forbidden",
			&Printer{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Down, Down}},
				},
			},
			HEADER + "resource1  ✖    ✖\n",
			HEADER + "resource1  \033[31m✖\033[0m    \033[31m✖\033[0m\n",
			HEADER + "resource1  no   no\n",
		},
		{
			"single result, all not applicable",
			&Printer{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{None, None}},
				},
			},
			HEADER + "resource1       \n",
			HEADER + "resource1  \033[0m\033[0m     \033[0m\033[0m\n",
			HEADER + "resource1  n/a  n/a\n",
		},
		{
			"single result, all ERR",
			&Printer{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Err, Err}},
				},
			},
			HEADER + "resource1  ERR  ERR\n",
			HEADER + "resource1  \033[35mERR\033[0m  \033[35mERR\033[0m\n",
			HEADER + "resource1  ERR  ERR\n",
		},
		{
			"single result, mixed",
			&Printer{
				Headers: []string{"NAME", "GET", "LIST"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Down, Up}},
				},
			},
			HEADER + "resource1  ✖    ✔\n",
			"",
			HEADER + "resource1  no   yes\n",
		},
		{
			"many results",
			&Printer{
				Headers: []string{"NAME", "GET"},
				Rows: []Row{
					{Intro: []string{"resource1"}, Entries: []Outcome{Down}},
					{Intro: []string{"resource2"}, Entries: []Outcome{Up}},
					{Intro: []string{"resource3"}, Entries: []Outcome{Err}},
				},
			},
			"NAME       GET\nresource1  ✖\nresource2  ✔\nresource3  ERR\n",
			"",
			"NAME       GET\nresource1  no\nresource2  yes\nresource3  ERR\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			test.p.Print(buf, "icon-table")
			assert.Equal(t, test.want, buf.String())

			buf = &bytes.Buffer{}
			test.p.Print(buf, "ascii-table")
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
			buf := &bytes.Buffer{}
			test.p.Print(buf, "icon-table")
			assert.Equal(t, test.wantColor, buf.String())

			buf = &bytes.Buffer{}
			test.p.Print(buf, "ascii-table")
			assert.Equal(t, test.wantASCII, buf.String())
		})
	}
}
