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

package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/corneliusweig/rakkess/pkg/rakkess"
	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rakkessOptions = options.NewRakkessOptions()
	v              string
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

More on https://github.com/corneliusweig/rakkess/blob/v0.2.1/doc/USAGE.md#usage
`

	rakkessExamples = `
  Review access to cluster-scoped resources
  $ rakkess

  Review access to namespaced resources in 'default'
  $ rakkess --namespace default

  Review access as a different user
  $ rakkess --as other-user

  Review access for different verbs
  $ rakkess --verbs get,watch,patch
`
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "rakkess",
	Short:   "Review access - show an access matrix for all resources",
	Long:    rakkessLongDescription,
	Example: rakkessExamples,
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		catchCtrlC(cancel)

		if err := rakkess.Resource(ctx, rakkessOptions); err != nil {
			logrus.Error(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", constants.DefaultLogLevel.String(), "Log level (debug, info, warn, error, fatal, panic)")

	AddRakkessFlags(rootCmd)

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := SetUpLogs(rakkessOptions.Streams.ErrOut, v); err != nil {
			return err
		}
		return nil
	}
}

func AddRakkessFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVar(&rakkessOptions.Verbs, "verbs", []string{"list", "create", "update", "delete"}, fmt.Sprintf("show access for verbs out of %s", constants.ValidVerbs))
	cmd.Flags().StringVarP(&rakkessOptions.OutputFormat, "output", "o", "icon-table", fmt.Sprintf("output format out of %s", constants.ValidOutputFormats))

	rakkessOptions.ConfigFlags.AddFlags(cmd.Flags())
}

func SetUpLogs(out io.Writer, level string) error {
	logrus.SetOutput(out)
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return errors.Wrap(err, "parsing log level")
	}
	logrus.SetLevel(lvl)
	logrus.Debugf("Set log-level to %s", level)
	return nil
}
