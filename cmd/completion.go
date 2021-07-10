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
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	completionLongDescription = `
	Outputs shell completion for the given shell (bash or zsh)

	OS X:
		$ source $(brew --prefix)/etc/bash_completion
		$ rakkess completion bash > ~/.rakkess-completion  # for bash users
		$ rakkess completion zsh > ~/.rakkess-completion   # for zsh users
		$ source ~/.rakkess-completion
	Ubuntu:
		$ source /etc/bash-completion
		$ source <(rakkess completion bash) # for bash users
		$ source <(rakkess completion zsh)  # for zsh users

	Additionally, you may want to output the completion to a file and source in your .bashrc
`

	zshCompdef = "\ncompdef _rakkess rakkess\n"
)

var completionCmd = &cobra.Command{
	Use:       "completion SHELL",
	Short:     "Output shell completion for the given shell (bash or zsh)",
	Long:      completionLongDescription,
	ValidArgs: []string{"bash", "zsh"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires 1 arg, found %d", len(args))
		}
		return cobra.OnlyValidArgs(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		out := rakkessOptions.Streams.Out
		var err error
		switch args[0] {
		case "bash":
			err = rootCmd.GenBashCompletion(out)
		case "zsh":
			err = getZshCompletion(out)
		}
		if err != nil {
			klog.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func getZshCompletion(out io.Writer) error {
	if err := rootCmd.GenZshCompletion(out); err != nil {
		return err
	}
	_, err := io.WriteString(out, zshCompdef)
	return err
}
