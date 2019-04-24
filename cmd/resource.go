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
	"github.com/corneliusweig/rakkess/pkg/rakkess"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	rakkessSubjectLong = `
Show all subjects with access to a given resource

This command slices the authorization space (subject, resource, verb)
along a plane of fixed resource.

Rakkess retrieves all (Cluster)Roles plus their bindings and evaluates the
authorization for the given resource and verbs. The result is shown as a
matrix with verbs in the horizontal and subjects in the vertical direction.

Note that the effective access right may differ from the shown results due to
group membership such as 'system:unauthenticated'.

More on https://github.com/corneliusweig/rakkess/blob/v0.3.0/doc/USAGE.md#usage
`

	rakkessSubjectExamples = `
  Review access to deployments in any namespace
  $ rakkess resource deployments

  Review access to deployments in the default namespace (with shorthands)
  $ rakkess r deploy --namespace default

  Review access to deployments with custom verbs
  $ rakkess r deploy --verbs get,watch,deletecollection

  Review access to a config-map with a specific name
  $ rakkess r cm config-map-name --verbs=all
`
)

// resourceCmd represents the resource command
var resourceCmd = &cobra.Command{
	Use:     "resource",
	Aliases: []string{"r", "for"},
	Short:   "Show all subjects with access to a given resource",
	Args:    cobra.RangeArgs(1, 2),
	Long:    rakkessSubjectLong,
	Example: rakkessSubjectExamples,
	Run: func(cmd *cobra.Command, args []string) {
		resource := args[0]
		var resourceName string
		if len(args) == 2 {
			resourceName = args[1]
		}
		if err := rakkess.Subject(rakkessOptions, resource, resourceName); err != nil {
			logrus.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	AddRakkessFlags(resourceCmd)
}
