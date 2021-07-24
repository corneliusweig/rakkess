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
	"context"
	"flag"
	"fmt"
	"strings"

	rakkess "github.com/corneliusweig/rakkess/internal"
	"github.com/corneliusweig/rakkess/internal/constants"
	"github.com/corneliusweig/rakkess/internal/diff"
	"github.com/corneliusweig/rakkess/internal/options"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	opts     = options.NewRakkessOptions()
	diffWith []string
)

const (
	rakkessLongDescription = `
Show an access matrix for all server resources

This command slices the authorization space (subject, resource, verb)
along a plane of fixed subject.

Rakkess retrieves the full list of server resources, checks access for
the current user with the given verbs, and prints the result as a matrix.
This complements the usual "kubectl auth can-i" command, which works for
a single resource and a single verb.

When passing the --diff-with flag, the matrix shows only the diff of the access
rights. The diff-with flag takes overrides in the form "flag=value". It accepts
the same flags as rakkess itself (without the leading --). The flag can be
repeated.
For example: --diff-with context=b --diff-with sa=kube-system:job-controller

More on https://github.com/corneliusweig/rakkess/blob/v0.4.7/doc/USAGE.md#usage
`

	rakkessExamples = `
  Review access to cluster-scoped resources
   $ rakkess

  Review access to namespaced resources in 'default'
   $ rakkess --namespace default

  Review access as a different user
   $ rakkess --as other-user

  Review access as a service-account
   $ rakkess --sa kube-system:namespace-controller

  Review access for different verbs
   $ rakkess --verbs get,watch,patch

  Review access rights diff with another service account
   $ rakkess --diff-with sa=kube-system:namespace-controller
`
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     constants.CommandName,
	Short:   "Review access - show an access matrix for all resources",
	Long:    constants.HelpTextMapName(rakkessLongDescription),
	Example: constants.HelpTextMapName(rakkessExamples),
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		catchCtrlC(cancel)

		res, err := rakkess.Resource(ctx, opts)
		if err != nil {
			return err
		}
		if diffWith == nil {
			t := res.Table(opts.Verbs)
			t.Render(opts.Streams.Out, opts.OutputFormat)
			return nil
		}

		orig := res
		flags := cmd.Flags()

		for _, arg := range diffWith {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("diffWith expects format flag=value, got %s", arg)
			}
			name, value := parts[0], parts[1]
			fl := flags.Lookup(name)
			if fl == nil && len(name) == 1 {
				fl = flags.ShorthandLookup(name)
			}
			if fl == nil {
				return fmt.Errorf("flag %q does not exist", name)
			}
			klog.V(2).Infof("Override flag %s=%s", name, value)
			if err := fl.Value.Set(value); err != nil {
				return fmt.Errorf("failed to set %s=%s", name, value)
			}
		}
		_ = opts.ExpandServiceAccount() // expand again in case `--sa` was overridden
		mod, err := rakkess.Resource(ctx, opts)
		if err != nil {
			return fmt.Errorf("with modified flags: %v", err)
		}

		t := diff.Diff(orig, mod, opts.Verbs)
		t.Render(opts.Streams.Out, opts.OutputFormat)
		return nil
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if n := opts.ConfigFlags.Namespace; n == nil || *n == "" {
			fmt.Fprintf(opts.Streams.Out, "No namespace given, this implies cluster scope (try -n if this is not intended)\n")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	rootCmd.SetOutput(opts.Streams.Out)
	return rootCmd.Execute()
}

func init() {
	klog.InitFlags(flag.CommandLine)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	AddRakkessFlags(rootCmd)
	rootCmd.Flags().StringVar(&opts.AsServiceAccount, constants.FlagServiceAccount, "", "similar to --as, but impersonate as service-account. The argument must be qualified <namespace>:<sa-name> or be combined with the --namespace option. Takes precedence over --as.")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		opts.ExpandVerbs()
	}
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return opts.ExpandServiceAccount()
	}
}

// AddRakkessFlags sets up common flags for subcommands.
func AddRakkessFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&opts.Verbs, constants.FlagVerbs, []string{"list", "create", "update", "delete"}, fmt.Sprintf("show access for verbs out of (%s)", strings.Join(constants.ValidVerbs, ", ")))
	cmd.Flags().StringVarP(&opts.OutputFormat, constants.FlagOutput, "o", "icon-table", fmt.Sprintf("output format out of (%s)", strings.Join(constants.ValidOutputFormats, ", ")))
	cmd.Flags().StringSliceVar(&diffWith, constants.FlagDiffWith, nil, "Show diff for modified call. For example --diff-with=namespace=kube-system.")

	opts.ConfigFlags.AddFlags(cmd.Flags())
}
