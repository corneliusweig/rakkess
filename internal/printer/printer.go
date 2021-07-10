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
	"sort"
	"strings"
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
	once       sync.Once
)

type Outcome uint8

const (
	None Outcome = iota
	Up
	Down
	Err
)

type Row struct {
	Label   string
	Entries []Outcome
}
type Printer struct {
	Headers []string
	Rows    []Row
}

func New(headers []string) *Printer {
	return &Printer{
		Headers: headers,
	}
}

func (p *Printer) AddRow(label string, outcomes ...Outcome) {
	row := Row{
		Label:   label,
		Entries: outcomes,
	}
	p.Rows = append(p.Rows, row)
}

func (p *Printer) Print(out io.Writer, conv func(Outcome) string) {
	once.Do(func() { initTerminal(out) })
	w := tabwriter.NewWriter(out, 4, 8, 2, ' ', tabwriter.SmashEscape|tabwriter.StripEscape)
	defer w.Flush()

	// table header
	for i, header := range p.Headers {
		if i == 0 {
			fmt.Fprint(w, header)
		} else {
			fmt.Fprintf(w, "\t%s", strings.ToUpper(header))
		}
	}
	fmt.Fprint(w, "\n")

	// table body
	sort.Slice(p.Rows, func(i, j int) bool { return p.Rows[i].Label < p.Rows[j].Label })
	for _, row := range p.Rows {
		fmt.Fprintf(w, "%s", row.Label)
		for _, e := range row.Entries {
			fmt.Fprintf(w, "\t%s", conv(e))
		}
		fmt.Fprint(w, "\n")
	}
}

// MatrixPrinter needs to be implemented by result types.
type MatrixPrinter interface {
	// Print writes the result for the requestedVerbs to w using the code converter.
	Print(w io.Writer, converter result.CodeConverter, verbs []string)
}

// PrintResults configures the table style and delegates printing to result.MatrixPrinter.
func PrintResults(out io.Writer, requestedVerbs []string, outputFormat string, results MatrixPrinter) {
	w := tabwriter.NewWriter(out, 4, 8, 2, ' ', tabwriter.SmashEscape|tabwriter.StripEscape)
	defer w.Flush()

	once.Do(func() { initTerminal(out) })

	codeConverter := humanreadableAccessCode
	if isTerminal(out) {
		codeConverter = colorHumanreadableAccessCode
	}
	if outputFormat == "ascii-table" {
		codeConverter = asciiAccessCode
	}

	results.Print(w, codeConverter, requestedVerbs)
}

func humanreadableAccessCode(code result.Access) string {
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

func colorHumanreadableAccessCode(code result.Access) string {
	return fmt.Sprintf("\xff\033[%dm\xff%s\xff\033[0m\xff", codeToColor(code), humanreadableAccessCode(code))
}

func codeToColor(code result.Access) color {
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

func asciiAccessCode(code result.Access) string {
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
