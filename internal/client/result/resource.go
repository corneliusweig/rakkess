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
	"cmp"
	"sort"
	"strings"

	"github.com/corneliusweig/rakkess/internal/printer"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceAccess holds the access result for all resources.
type ResourceAccess map[string]map[string]Access

// Print implements MatrixPrinter.Print. It prints a tab-separated table with a header.
func (ra ResourceAccess) Table(verbs []string) *printer.Table {
	var groupResources []schema.GroupResource
	for name := range ra {
		groupResources = append(groupResources, schema.ParseGroupResource(name))
	}
	sort.Slice(groupResources, func(i, j int) bool {
		x := groupResources[i]
		y := groupResources[j]
		// first sort by group, then resource
		if x.Group != y.Group {
			return cmp.Less(x.Group, y.Group)
		}
		return cmp.Less(x.Resource, y.Resource)
	})

	upperVerbs := make([]string, 0, len(verbs))
	for _, v := range verbs {
		upperVerbs = append(upperVerbs, strings.ToUpper(v))
	}

	p := printer.TableWithHeaders(nil)

	// table body
	lastGroup := ""
	for i, gr := range groupResources {
		// print the API group and verbs when the API group changes, or for the first API group (which often is "")
		if gr.Group != lastGroup || i == 0 {

			if i != 0 {
				p.AddRow([]string{" "}, printer.None) // at least one "none" outcome needed to get the tabprinter aligning all columns
			}

			displayGroup := gr.Group
			if displayGroup == "" {
				displayGroup = "core"
			}

			p.AddRow(append([]string{displayGroup + ":"}, upperVerbs...), printer.None)
			lastGroup = gr.Group
		}

		var outcomes []printer.Outcome

		res := ra[gr.String()]
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
		p.AddRow([]string{gr.Resource}, outcomes...)
	}
	return p
}
