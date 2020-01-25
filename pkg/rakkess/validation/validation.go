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

package validation

import (
	"fmt"

	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Options validates RakkessOptions. Fields validated:
// - OutputFormat
// - Verbs
func Options(opts *options.RakkessOptions) error {
	if err := verbs(opts.Verbs); err != nil {
		return err
	}
	return OutputFormat(opts.OutputFormat)
}

func OutputFormat(format string) error {
	for _, o := range constants.ValidOutputFormats {
		if o == format {
			return nil
		}
	}
	return fmt.Errorf("unexpected output format: %s", format)
}

func verbs(verbs []string) error {
	valid := sets.NewString(constants.ValidVerbs...)
	given := sets.NewString(verbs...)
	difference := given.Difference(valid)

	if difference.Len() > 0 {
		return fmt.Errorf("unexpected verbs: %s", difference.List())
	}

	return nil
}
