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
	"fmt"
	"io"
	"sort"
	"strings"
)

// ResourceAccess holds the access result for all resources.
type ResourceAccess map[string]map[string]Access

// Print implements MatrixPrinter.Print. It prints a tab-separated table with a header.
func (ra ResourceAccess) Print(w io.Writer, converter CodeConverter, requestedVerbs []string) {
	var names []string
	for name := range ra {
		names = append(names, name)
	}
	sort.Strings(names)

	// table header
	fmt.Fprint(w, "NAME")
	for _, v := range requestedVerbs {
		fmt.Fprintf(w, "\t%s", strings.ToUpper(v))
	}
	fmt.Fprint(w, "\n")

	// table body
	for _, name := range names {
		fmt.Fprintf(w, "%s", name)
		res := ra[name]
		for _, v := range requestedVerbs {
			fmt.Fprintf(w, "\t%s", converter(res[v]))
		}
		fmt.Fprint(w, "\n")
	}
}
