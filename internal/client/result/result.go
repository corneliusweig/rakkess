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

import "io"

type Access uint8

// This encodes the access of the given subject to the resource+verb combination.
const (
	AccessDenied Access = iota
	AccessAllowed
	AccessNotApplicable
	AccessRequestErr
)

// CodeConverter converts an access code to a human-readable string.
type CodeConverter func(Access) string

// MatrixPrinter needs to be implemented by result types.
type MatrixPrinter interface {
	// Print writes the result for the requestedVerbs to w using the code converter.
	Print(w io.Writer, converter CodeConverter, requestedVerbs []string)
}
