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
