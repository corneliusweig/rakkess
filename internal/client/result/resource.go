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
	"sort"
	"strings"

	"github.com/corneliusweig/rakkess/internal/printer"
)

// ResourceAccess holds the access result for all resources.
type ResourceAccess map[string]map[string]Access

// Print implements MatrixPrinter.Print. It prints a tab-separated table with a header.
func (ra ResourceAccess) ToPrinter(verbs []string) *printer.Printer {
	var names []string
	for name := range ra {
		names = append(names, name)
	}
	sort.Strings(names)

	// table header
	headers := []string{"NAME"}
	for _, v := range verbs {
		headers = append(headers, strings.ToUpper(v))
	}

	p := printer.New(headers)

	// table body
	for _, name := range names {
		var outcomes []printer.Outcome

		res := ra[name]
		for _, v := range verbs {
			var o printer.Outcome
			switch res[v] {
			case Denied:
				o = printer.Down
			case Allowed:
				o = printer.Up
			case NotApplicable:
				o = printer.None
			case RequestErr:
				o = printer.Err
			}
			outcomes = append(outcomes, o)
		}
		p.AddRow([]string{name}, outcomes...)
	}
	return p
}
