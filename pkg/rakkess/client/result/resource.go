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

package result

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type CodeConverter func(int) string
type MatrixPrinter interface {
	Print(w io.Writer, converter CodeConverter, requestedVerbs []string)
}

type ResourceAccessItem struct {
	Name   string
	Access map[string]int
	Err    []error
}

type ResourceAccess []ResourceAccessItem

func NewResourceAccess(items []ResourceAccessItem) ResourceAccess {
	ra := ResourceAccess(items)
	sort.Stable(ra)
	return ra
}

func (ra ResourceAccess) Len() int      { return len(ra) }
func (ra ResourceAccess) Swap(i, j int) { ra[i], ra[j] = ra[j], ra[i] }
func (ra ResourceAccess) Less(i, j int) bool {
	ret := strings.Compare(ra[i].Name, ra[j].Name)
	if ret > 0 {
		return false
	} else if ret == 0 {
		return i < j
	}
	return true
}

func (ra ResourceAccess) Print(w io.Writer, converter CodeConverter, requestedVerbs []string) {
	// table header
	fmt.Fprint(w, "NAME")
	for _, v := range requestedVerbs {
		fmt.Fprintf(w, "\t%s", strings.ToUpper(v))
	}
	fmt.Fprint(w, "\n")

	// table body
	for _, r := range ra {
		fmt.Fprintf(w, "%s", r.Name)
		for _, v := range requestedVerbs {
			fmt.Fprintf(w, "\t%s", converter(r.Access[v]))
		}
		fmt.Fprint(w, "\n")
	}
}
