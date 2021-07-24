/*
Copyright 2021 Cornelius Weig

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

package diff

import (
	"sort"
	"strings"

	"github.com/corneliusweig/rakkess/internal/client/result"
	"github.com/corneliusweig/rakkess/internal/printer"
	"k8s.io/klog/v2"
)

// Diff takes two result sets and produces a printer that contains only the
// diff.
func Diff(left, right result.ResourceAccess, verbs []string) *printer.Table {
	// table header
	headers := []string{"NAME"}
	for _, v := range verbs {
		headers = append(headers, strings.ToUpper(v))
	}

	var names []string
	for name := range left {
		names = append(names, name)
	}
	sort.Strings(names)

	p := printer.TableWithHeaders(headers)

	for _, name := range names {
		l, r := left[name], right[name]
		klog.V(3).Infof("left=%v right=%v name=%s", l, r, name)

		skip := true
		var outcomes []printer.Outcome
		for _, verb := range verbs {
			ll, rr := l[verb], r[verb]
			var o printer.Outcome
			if ll != rr {
				skip = false
				if ll == result.Allowed {
					o = printer.Down
				}
				if rr == result.Allowed {
					o = printer.Up
				}
			}
			outcomes = append(outcomes, o)
		}
		if !skip {
			p.AddRow([]string{name}, outcomes...)
		}
	}

	for name := range right {
		if _, ok := left[name]; !ok {
			klog.Warning("Some differences may be hidden, please swap the roles to get the full picture.")
			break
		}
	}

	return p
}
