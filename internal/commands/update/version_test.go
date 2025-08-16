package update

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Version
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "1.2.3",
			want: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Raw:   "1.2.3",
			},
		},
		{
			name:  "version with v prefix",
			input: "v2.0.0",
			want: &Version{
				Major: 2,
				Minor: 0,
				Patch: 0,
				Raw:   "2.0.0",
			},
		},
		{
			name:  "version with prerelease",
			input: "1.0.0-alpha.1",
			want: &Version{
				Major:      1,
				Minor:      0,
				Patch:      0,
				Prerelease: "alpha.1",
				Raw:        "1.0.0-alpha.1",
			},
		},
		{
			name:  "version with build metadata",
			input: "1.0.0+20130313144700",
			want: &Version{
				Major: 1,
				Minor: 0,
				Patch: 0,
				Build: "20130313144700",
				Raw:   "1.0.0+20130313144700",
			},
		},
		{
			name:  "version with prerelease and build",
			input: "1.0.0-beta+exp.sha.5114f85",
			want: &Version{
				Major:      1,
				Minor:      0,
				Patch:      0,
				Prerelease: "beta",
				Build:      "exp.sha.5114f85",
				Raw:        "1.0.0-beta+exp.sha.5114f85",
			},
		},
		{
			name:    "invalid version",
			input:   "not-a-version",
			wantErr: true,
		},
		{
			name:    "incomplete version",
			input:   "1.2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch {
					t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
				}
				if got.Prerelease != tt.want.Prerelease {
					t.Errorf("ParseVersion() Prerelease = %v, want %v", got.Prerelease, tt.want.Prerelease)
				}
				if got.Build != tt.want.Build {
					t.Errorf("ParseVersion() Build = %v, want %v", got.Build, tt.want.Build)
				}
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int // -1, 0, or 1
	}{
		{
			name:     "equal versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
		},
		{
			name:     "major version difference",
			v1:       "1.0.0",
			v2:       "2.0.0",
			expected: -1,
		},
		{
			name:     "minor version difference",
			v1:       "1.1.0",
			v2:       "1.2.0",
			expected: -1,
		},
		{
			name:     "patch version difference",
			v1:       "1.0.1",
			v2:       "1.0.2",
			expected: -1,
		},
		{
			name:     "prerelease vs stable",
			v1:       "1.0.0-alpha",
			v2:       "1.0.0",
			expected: -1,
		},
		{
			name:     "prerelease comparison",
			v1:       "1.0.0-alpha",
			v2:       "1.0.0-beta",
			expected: -1,
		},
		{
			name:     "prerelease with numbers",
			v1:       "1.0.0-alpha.1",
			v2:       "1.0.0-alpha.2",
			expected: -1,
		},
		{
			name:     "newer version",
			v1:       "2.1.0",
			v2:       "1.9.9",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := ParseVersion(tt.v1)
			if err != nil {
				t.Fatalf("Failed to parse v1: %v", err)
			}
			v2, err := ParseVersion(tt.v2)
			if err != nil {
				t.Fatalf("Failed to parse v2: %v", err)
			}

			result := v1.Compare(v2)
			if result != tt.expected {
				t.Errorf("Compare(%s, %s) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}

			// Test convenience methods
			if tt.expected < 0 && !v1.IsOlder(v2) {
				t.Errorf("IsOlder(%s, %s) = false, want true", tt.v1, tt.v2)
			}
			if tt.expected > 0 && !v1.IsNewer(v2) {
				t.Errorf("IsNewer(%s, %s) = false, want true", tt.v1, tt.v2)
			}
			if tt.expected == 0 && !v1.IsEqual(v2) {
				t.Errorf("IsEqual(%s, %s) = false, want true", tt.v1, tt.v2)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name  string
		input *Version
		want  string
	}{
		{
			name: "simple version",
			input: &Version{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: "1.2.3",
		},
		{
			name: "version with prerelease",
			input: &Version{
				Major:      1,
				Minor:      0,
				Patch:      0,
				Prerelease: "beta.1",
			},
			want: "1.0.0-beta.1",
		},
		{
			name: "version with build",
			input: &Version{
				Major: 1,
				Minor: 0,
				Patch: 0,
				Build: "20130313",
			},
			want: "1.0.0+20130313",
		},
		{
			name: "version with prerelease and build",
			input: &Version{
				Major:      2,
				Minor:      1,
				Patch:      0,
				Prerelease: "rc.1",
				Build:      "build.123",
			},
			want: "2.1.0-rc.1+build.123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectChannel(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "stable version",
			version: "1.0.0",
			want:    "stable",
		},
		{
			name:    "beta version",
			version: "1.0.0-beta.1",
			want:    "beta",
		},
		{
			name:    "alpha version",
			version: "1.0.0-alpha",
			want:    "alpha",
		},
		{
			name:    "nightly version",
			version: "1.0.0-nightly.20240101",
			want:    "nightly",
		},
		{
			name:    "dev version",
			version: "1.0.0-dev",
			want:    "nightly",
		},
		{
			name:    "other prerelease",
			version: "1.0.0-rc.1",
			want:    "prerelease",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("Failed to parse version: %v", err)
			}
			got := detectChannel(v)
			if got != tt.want {
				t.Errorf("detectChannel(%s) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
