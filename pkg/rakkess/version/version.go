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

package version

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/blang/semver"
)

var version, gitCommit, buildDate string
var platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)

type BuildInfo struct {
	BuildDate string
	Compiler  string
	GitCommit string
	GoVersion string
	Platform  string
	Version   string
}

// GetBuildInfo returns build information about the binary
func GetBuildInfo() *BuildInfo {
	// These vars are set via -ldflags settings during 'go build'
	return &BuildInfo{
		Version:   version,
		GitCommit: gitCommit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  platform,
	}
}

func ParseVersion(version string) (semver.Version, error) {
	version = strings.TrimLeft(strings.TrimSpace(version), "v")
	return semver.Parse(version)
}
