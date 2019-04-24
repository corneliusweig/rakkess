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
	"os"
	"text/template"

	"github.com/corneliusweig/rakkess/pkg/rakkess/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	versionTemplate = `{{.Version}}
`
	fullInfoTemplate = `rakkess:    {{.Version}}
platform:   {{.Platform}}
git commit: {{.GitCommit}}
build date: {{.BuildDate}}
go version: {{.GoVersion}}
compiler:   {{.Compiler}}
`
	flagFull = "full"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Args:  cobra.NoArgs,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolP(flagFull, "f", false, "print extended version information")
}

func runVersion(cmd *cobra.Command, args []string) {
	var tpl string

	if cmd.Flag(flagFull).Changed {
		tpl = fullInfoTemplate
	} else {
		tpl = versionTemplate
	}

	var t = template.Must(template.New("info").Parse(tpl))

	if err := t.Execute(os.Stdout, version.GetBuildInfo()); err != nil {
		logrus.Warn("Could not print version info")
	}
}
