package update

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Version represents a semantic version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
	Raw        string
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(versionStr string) (*Version, error) {
	// Clean up the version string (remove 'v' prefix if present)
	versionStr = strings.TrimPrefix(versionStr, "v")
	versionStr = strings.TrimSpace(versionStr)

	// Semantic version regex pattern
	// Matches: MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
	pattern := `^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(versionStr)
	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
		Raw:        versionStr,
	}, nil
}

// String returns the string representation of the version
func (v *Version) String() string {
	result := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Prerelease != "" {
		result += "-" + v.Prerelease
	}
	if v.Build != "" {
		result += "+" + v.Build
	}
	return result
}

// Compare compares two versions
// Returns:
//
//	-1 if v < other
//	 0 if v == other
//	 1 if v > other
func (v *Version) Compare(other *Version) int {
	// Compare major version
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor version
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch version
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare prerelease versions
	// No prerelease > prerelease (1.0.0 > 1.0.0-alpha)
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}

	// If both have prereleases, compare them
	if v.Prerelease != "" && other.Prerelease != "" {
		return comparePrereleases(v.Prerelease, other.Prerelease)
	}

	return 0
}

// IsNewer returns true if v is newer than other
func (v *Version) IsNewer(other *Version) bool {
	return v.Compare(other) > 0
}

// IsOlder returns true if v is older than other
func (v *Version) IsOlder(other *Version) bool {
	return v.Compare(other) < 0
}

// IsEqual returns true if v equals other
func (v *Version) IsEqual(other *Version) bool {
	return v.Compare(other) == 0
}

// comparePrereleases compares two prerelease version strings
func comparePrereleases(a, b string) int {
	// Split by dots
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	// Compare each part
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		aIsNum := isNumeric(aParts[i])
		bIsNum := isNumeric(bParts[i])

		if aIsNum && bIsNum {
			// Both numeric, compare as numbers
			aNum, _ := strconv.Atoi(aParts[i])
			bNum, _ := strconv.Atoi(bParts[i])
			if aNum != bNum {
				if aNum < bNum {
					return -1
				}
				return 1
			}
		} else if !aIsNum && !bIsNum {
			// Both strings, compare lexically
			cmp := strings.Compare(aParts[i], bParts[i])
			if cmp != 0 {
				return cmp
			}
		} else {
			// Mixed types, numeric < string
			if aIsNum {
				return -1
			}
			return 1
		}
	}

	// If all parts are equal, the one with fewer parts is less
	if len(aParts) < len(bParts) {
		return -1
	}
	if len(aParts) > len(bParts) {
		return 1
	}

	return 0
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// VersionMetadata contains additional version information
type VersionMetadata struct {
	Version  *Version
	Commit   string
	Date     string
	BuiltBy  string
	Channel  string
	Platform string
	Arch     string
}

// GetCurrentVersion returns the current binary version with metadata
func GetCurrentVersion() (*VersionMetadata, error) {
	// Import the version from the root command
	// This will be set via ldflags during build
	versionStr := getCurrentVersionString()

	version, err := ParseVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version: %w", err)
	}

	return &VersionMetadata{
		Version:  version,
		Commit:   getCurrentCommit(),
		Date:     getCurrentDate(),
		BuiltBy:  getCurrentBuiltBy(),
		Channel:  detectChannel(version),
		Platform: getCurrentPlatform(),
		Arch:     getCurrentArch(),
	}, nil
}

// Helper functions to get version information
// These will be linked to the actual values from the root command

func getCurrentVersionString() string {
	// This will be replaced with actual version from root command
	// For now, return a default
	return "0.2.0"
}

func getCurrentCommit() string {
	return "unknown"
}

func getCurrentDate() string {
	return "unknown"
}

func getCurrentBuiltBy() string {
	return "unknown"
}

func getCurrentPlatform() string {
	return runtime.GOOS
}

func getCurrentArch() string {
	return runtime.GOARCH
}

// detectChannel determines the release channel from the version
func detectChannel(v *Version) string {
	if v.Prerelease == "" {
		return "stable"
	}

	prerelease := strings.ToLower(v.Prerelease)
	if strings.Contains(prerelease, "beta") {
		return "beta"
	}
	if strings.Contains(prerelease, "alpha") {
		return "alpha"
	}
	if strings.Contains(prerelease, "nightly") || strings.Contains(prerelease, "dev") {
		return "nightly"
	}

	return "prerelease"
}
