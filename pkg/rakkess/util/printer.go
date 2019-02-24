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

package util

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/corneliusweig/rakkess/pkg/rakkess/client"
)

func PrintResults(out io.Writer, requestedVerbs []string, results []client.Result) {
	w := tabwriter.NewWriter(out, 4, 8, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprint(w, "NAME")
	for _, v := range requestedVerbs {
		fmt.Fprintf(w, "\t%s", strings.ToUpper(v))
	}
	fmt.Fprint(w, "\n")

	for _, r := range results {
		fmt.Fprintf(w, "%s", r.Name)
		for _, v := range requestedVerbs {
			fmt.Fprintf(w, "\t%s", humanreadableAccessCode(r.Access[v]))
		}
		fmt.Fprint(w, "\n")
	}
}

func humanreadableAccessCode(code int) string {
	switch code {
	case client.AccessAllowed:
		return "yes"
	case client.AccessDenied:
		return "no"
	case client.AccessNotApplicable:
		return ""
	case client.AccessRequestErr:
		return "ERR"
	default:
		panic("unknown access code")
	}
}
