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
