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
	"strings"
	"sync"

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
	Intro   []string
	Entries []Outcome
}
type Table struct {
	Headers []string
	Rows    []Row
}

func TableWithHeaders(headers []string) *Table {
	return &Table{
		Headers: headers,
	}
}

func (p *Table) AddRow(intro []string, outcomes ...Outcome) {
	row := Row{
		Intro:   intro,
		Entries: outcomes,
	}
	p.Rows = append(p.Rows, row)
}

func (p *Table) Render(out io.Writer, outputFormat string) {
	once.Do(func() { initTerminal(out) })

	conv := humanreadableAccessCode
	if isTerminal(out) {
		conv = colorHumanreadableAccessCode
	}
	if outputFormat == "ascii-table" {
		conv = asciiAccessCode
	}

	w := tabwriter.NewWriter(out, 4, 8, 2, ' ', tabwriter.SmashEscape|tabwriter.StripEscape)
	defer w.Flush()

	// table header
	for i, h := range p.Headers {
		if i == 0 {
			fmt.Fprint(w, h)
		} else {
			fmt.Fprintf(w, "\t%s", h)
		}
	}
	fmt.Fprint(w, "\n")

	// table body
	for _, row := range p.Rows {
		fmt.Fprintf(w, "%s", strings.Join(row.Intro, "\t"))
		for _, e := range row.Entries {
			fmt.Fprintf(w, "\t%s", conv(e)) // FIXME
		}
		fmt.Fprint(w, "\n")
	}
}

func humanreadableAccessCode(o Outcome) string {
	switch o {
	case None:
		return ""
	case Up:
		return "✔" // ✓
	case Down:
		return "✖" // ✕
	case Err:
		return "ERR"
	default:
		panic("unknown access code")
	}
}

func colorHumanreadableAccessCode(o Outcome) string {
	return fmt.Sprintf("\xff\033[%dm\xff%s\xff\033[0m\xff", codeToColor(o), humanreadableAccessCode(o))
}

func codeToColor(o Outcome) color {
	switch o {
	case None:
		return none
	case Up:
		return green
	case Down:
		return red
	case Err:
		return purple
	}
	return none
}

func asciiAccessCode(o Outcome) string {
	switch o {
	case None:
		return "n/a"
	case Up:
		return "yes"
	case Down:
		return "no"
	case Err:
		return "ERR"
	default:
		panic("unknown access code")
	}
}
