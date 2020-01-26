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
	"fmt"
	"io"
	"sync"

	"github.com/corneliusweig/rakkess/internal/client/result"
	"github.com/corneliusweig/tabwriter"
)

type color int

const (
	red    = color(31)
	green  = color(32)
	purple = color(35)
	none   = color(0)
)

var (
	isTerminal = isTerminalImpl
	terminit   sync.Once
)

// PrintResults configures the table style and delegates printing to result.MatrixPrinter.
func PrintResults(out io.Writer, requestedVerbs []string, outputFormat string, results result.MatrixPrinter) {
	w := tabwriter.NewWriter(out, 4, 8, 2, ' ', tabwriter.SmashEscape|tabwriter.StripEscape)
	defer w.Flush()

	terminit.Do(func() { initTerminal(out) })

	codeConverter := humanreadableAccessCode
	if isTerminal(out) {
		codeConverter = colorHumanreadableAccessCode
	}
	if outputFormat == "ascii-table" {
		codeConverter = asciiAccessCode
	}

	results.Print(w, codeConverter, requestedVerbs)
}

func humanreadableAccessCode(code int) string {
	switch code {
	case result.AccessAllowed:
		return "✔" // ✓
	case result.AccessDenied:
		return "✖" // ✕
	case result.AccessNotApplicable:
		return ""
	case result.AccessRequestErr:
		return "ERR"
	default:
		panic("unknown access code")
	}
}

func colorHumanreadableAccessCode(code int) string {
	return fmt.Sprintf("\xff\033[%dm\xff%s\xff\033[0m\xff", codeToColor(code), humanreadableAccessCode(code))
}

func codeToColor(code int) color {
	switch code {
	case result.AccessAllowed:
		return green
	case result.AccessDenied:
		return red
	case result.AccessNotApplicable:
		return none
	case result.AccessRequestErr:
		return purple
	}
	return none
}

func asciiAccessCode(code int) string {
	switch code {
	case result.AccessAllowed:
		return "yes"
	case result.AccessDenied:
		return "no"
	case result.AccessNotApplicable:
		return "n/a"
	case result.AccessRequestErr:
		return "ERR"
	default:
		panic("unknown access code")
	}
}
