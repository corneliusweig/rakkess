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

package cmd

import (
	"context"
	"strings"

	rakkess "github.com/corneliusweig/rakkess/internal"
	"github.com/corneliusweig/rakkess/internal/constants"
	"github.com/corneliusweig/rakkess/internal/diff"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const (
	diffLongHelp = `
Show diff of access rights between two different settings.

The diff command accepts the same options as the root command. Those settings
define the original settings (◀). In addition, you can pass positional args to
patch the original settings. The patched settings define the access settings
to compare against (▶).

The positional arguments must have the pattern 'flagname=value' where flagname
is the flag to override. For example:

  $ rakkess diff -n kube-system --sa=coredns sa=attachdetach-controller

This uses the 'coredns' service-account for the left hand side (◀), and the
'attachdetach-controller' service-account for the right hand side (▶). Note
that the overrides have the same pattern as the original flag without the '--'.

The output only shows differences between the two settings. Resources with the
same access rules are skipped.

More on https://github.com/corneliusweig/rakkess/blob/v0.4.7/doc/USAGE.md#usage
`

	diffExamples = `
  Review diff of access rights between context one and two
   $ rakkess diff --context=one context=two
   or if one is the active context simply
   $ rakkess diff context=two

  Review diff of access rights between service accounts
   $ rakkess diff -n kube-system --sa=coredns sa=attachdetach-controller
`
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:     "diff",
	Short:   "Show access rule differences between two settings",
	Args:    cobra.ArbitraryArgs,
	Long:    constants.HelpTextMapName(diffLongHelp),
	Example: constants.HelpTextMapName(diffExamples),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		catchCtrlC(cancel)

		overrides := make(map[string]string)
		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				klog.Exitf("Diff arg needs to set a value (example flag=value), got %s", arg)
			}
			overrides[parts[0]] = parts[1]
		}
		if len(overrides) == 0 {
			klog.Exitf("Nothing to diff against. See --help for examples.")
		}

		opts.ExpandServiceAccount()
		left, err := rakkess.Resource(ctx, opts)
		if err != nil {
			klog.Exitf("Original options failed with %v", err)
		}

		for name, value := range overrides {
			klog.V(2).Infof("Overriding flag %s=%s", name, value)
			if err := cmd.Flags().Set(name, value); err != nil {
				klog.Exitf("Overriding flags failed with %v", err)
			}
		}

		opts.ExpandServiceAccount()
		right, err := rakkess.Resource(ctx, opts)
		if err != nil {
			klog.Exitf("Modified options failed with %v", err)
		}

		t := diff.Diff(left, right, opts.Verbs)
		t.Render(opts.Streams.Out, "left-right")
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	AddRakkessFlags(diffCmd)
	diffCmd.Flags().StringVar(&opts.AsServiceAccount, constants.FlagServiceAccount, "", "similar to --as, but impersonate as service-account. The argument must be qualified <namespace>:<sa-name> or be combined with the --namespace option. Takes precedence over --as.")
}
