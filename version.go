// Copyright 2025 Edgeo SCADA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bacnet

// Version information for the bacnet package.
const (
	// Version is the current version of the bacnet package.
	Version = "1.0.0"

	// VersionMajor is the major version number.
	VersionMajor = 1

	// VersionMinor is the minor version number.
	VersionMinor = 0

	// VersionPatch is the patch version number.
	VersionPatch = 0
)

// VersionInfo contains detailed version information.
type VersionInfo struct {
	Version string
	Major   int
	Minor   int
	Patch   int
}

// GetVersion returns the current version information.
func GetVersion() VersionInfo {
	return VersionInfo{
		Version: Version,
		Major:   VersionMajor,
		Minor:   VersionMinor,
		Patch:   VersionPatch,
	}
}
