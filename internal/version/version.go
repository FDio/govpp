//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package version keeps track of versioning info for GoVPP.
package version

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

// Following variables should normally be updated via `-ldflags "-X ..."`.
// However, the version string is hard-coded to ensure it is always included
// even with bare go build/install.
var (
	name       = "govpp"
	version    = "v0.14.0-dev"
	commit     = "unknown"
	branch     = "HEAD"
	buildStamp = ""
	buildUser  = "unknown"
	buildHost  = "unknown"

	buildDate time.Time
)

func init() {
	buildstampInt64, _ := strconv.ParseInt(buildStamp, 10, 64)
	if buildstampInt64 == 0 {
		modTime, _ := binaryModTime()
		buildstampInt64 = modTime.Unix()
	}
	buildDate = time.Unix(buildstampInt64, 0)
}

func binaryModTime() (time.Time, error) {
	// Get the path of the currently running binary
	binaryPath, err := os.Executable()
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to get current binary path: %w", err)
	}

	// Get the file info for the binary
	fileInfo, err := os.Stat(binaryPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to get file info for binary: %w", err)
	}

	// Return the modification time
	return fileInfo.ModTime(), nil
}

// Version return semver string.
func Version() string {
	return version
}

func Info() string {
	return fmt.Sprintf(`%s %s`, name, version)
}

func Verbose() string {
	return fmt.Sprintf(`%s
  Version:      %s
  Branch:   	%s
  Revision: 	%s
  Built by:  	%s@%s 
  Build date:	%s
  Go runtime:	%s (%s/%s)`,
		name,
		version, branch, commit,
		buildUser, buildHost, buildDate.Format(time.UnixDate),
		runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}

func Short() string {
	return fmt.Sprintf(`%s %s`, name, version)
}

func BuildTime() string {
	stamp := buildDate.Format(time.UnixDate)
	if !buildDate.IsZero() {
		stamp += fmt.Sprintf(" (%s)", timeAgo(buildDate))
	}
	return stamp
}

func BuiltBy() string {
	return fmt.Sprintf("%s@%s (%s %s/%s)",
		buildUser, buildHost, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}

func timeAgo(t time.Time) string {
	const timeDay = time.Hour * 24
	if ago := time.Since(t); ago > timeDay {
		return fmt.Sprintf("%v days ago", float64(ago.Round(timeDay)/timeDay))
	} else if ago > time.Hour {
		return fmt.Sprintf("%v hours ago", ago.Round(time.Hour).Hours())
	} else if ago > time.Minute {
		return fmt.Sprintf("%v minutes ago", ago.Round(time.Minute).Minutes())
	}
	return "just now"
}
