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

	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/stretchr/testify/assert"
)

func TestMainHelp(t *testing.T) {
	origOpts := rakkessOptions
	newOpts, _, stdout, stderr := options.NewTestRakkessOptions()

	defer func(args []string) {
		os.Args = args
		rakkessOptions = origOpts
	}(os.Args)
	os.Args = []string{"rakkess", "help"}
	rakkessOptions = newOpts

	err := Execute()

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "Available Commands:")
	assert.Equal(t, "", stderr.String())
}

func TestMainUnknownCommand(t *testing.T) {
	origOpts := rakkessOptions
	newOpts, _, _, _ := options.NewTestRakkessOptions()

	defer func(args []string) {
		os.Args = args
		rakkessOptions = origOpts
	}(os.Args)
	os.Args = []string{"rakkess", "unknown"}
	rakkessOptions = newOpts

	err := Execute()

	assert.Error(t, err)
}

func TestMainVersionCommand(t *testing.T) {
	origOpts := rakkessOptions
	newOpts, _, stdout, stderr := options.NewTestRakkessOptions()

	defer func(args []string) {
		os.Args = args
		rakkessOptions = origOpts
	}(os.Args)
	os.Args = []string{"rakkess", "version", "--full"}
	rakkessOptions = newOpts

	err := Execute()

	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "rakkess: ")
	assert.Equal(t, "", stderr.String())
}
