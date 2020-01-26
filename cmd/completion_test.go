// +build !accessmatrix

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

package cmd

import (
	"os"
	"testing"

	"github.com/corneliusweig/rakkess/internal/options"
	"github.com/stretchr/testify/assert"
)

func TestMainCompletionCommand(t *testing.T) {
	tests := [][]string{
		{"rakkess", "completion", "zsh"},
		{"rakkess", "completion", "bash"},
	}

	for _, testargs := range tests {
		t.Run(testargs[2], func(t *testing.T) {
			origOpts := rakkessOptions
			newOpts, _, stdout, stderr := options.NewTestRakkessOptions()

			defer func(args []string) {
				os.Args = args
				rakkessOptions = origOpts
			}(os.Args)
			os.Args = testargs
			rakkessOptions = newOpts

			err := Execute()

			assert.NoError(t, err)
			assert.NotEmpty(t, stdout.String())
			assert.Equal(t, "", stderr.String())
		})

	}
}
