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

// resourceCmd represents the resource command
var resourceCmd = &cobra.Command{
	Use:     "resource",
	Aliases: []string{"for", "for-resource"},
	Short:   "Show access matrix for a given resource",
	Args:    cobra.ExactArgs(1),
	Long:    `todo`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := rakkess.RakkessSubject(rakkessOptions, args[0]); err != nil {
			logrus.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	AddRakkessFlags(resourceCmd)
}
