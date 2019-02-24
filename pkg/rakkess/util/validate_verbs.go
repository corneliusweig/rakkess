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

	"k8s.io/apimachinery/pkg/util/sets"
)

func ValidateVerbs(verbs []string) error {
	valid := sets.NewString("get", "list", "watch", "create", "update", "delete", "proxy")
	given := sets.NewString(verbs...)
	difference := given.Difference(valid)

	if difference.Len() > 0 {
		return fmt.Errorf("unexpected verbs: %s", difference.List())
	}

	return nil
}
