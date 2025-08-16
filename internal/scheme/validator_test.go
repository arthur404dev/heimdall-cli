package scheme

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateScheme(t *testing.T) {
	tests := []struct {
		name        string
		scheme      *Scheme
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid scheme with all required colors",
			scheme: &Scheme{
				Name: "Valid Scheme",
				Colours: map[string]string{
					"base00": "#1a1a1a",
					"base01": "#2a2a2a",
					"base02": "#3a3a3a",
					"base03": "#4a4a4a",
					"base04": "#5a5a5a",
					"base05": "#6a6a6a",
					"base06": "#7a7a7a",
					"base07": "#8a8a8a",
					"base08": "#9a9a9a",
					"base09": "#aaaaaa",
					"base0A": "#bababa",
					"base0B": "#cacaca",
					"base0C": "#dadada",
					"base0D": "#eaeaea",
					"base0E": "#fafafa",
					"base0F": "#ffffff",
				},
			},
			expectError: false,
		},
		{
			name: "scheme missing name",
			scheme: &Scheme{
				Colours: map[string]string{
					"base00": "#000000",
				},
			},
			expectError: true,
			errorMsg:    "scheme name is required",
		},
		{
			name: "scheme with nil colors map",
			scheme: &Scheme{
				Name: "No Colors",
			},
			expectError: true,
			errorMsg:    "colors map is required",
		},
		{
			name: "scheme missing required color keys",
			scheme: &Scheme{
				Name: "Incomplete",
				Colours: map[string]string{
					"base00": "#000000",
					"base01": "#111111",
				},
			},
			expectError: true,
			errorMsg:    "missing required color keys",
		},
		{
			name: "scheme with invalid hex color",
			scheme: &Scheme{
				Name: "Invalid Colors",
				Colours: map[string]string{
					"base00": "#1a1a1a",
					"base01": "not-a-hex",
					"base02": "#3a3a3a",
					"base03": "#4a4a4a",
					"base04": "#5a5a5a",
					"base05": "#6a6a6a",
					"base06": "#7a7a7a",
					"base07": "#8a8a8a",
					"base08": "#9a9a9a",
					"base09": "#aaaaaa",
					"base0A": "#bababa",
					"base0B": "#cacaca",
					"base0C": "#dadada",
					"base0D": "#eaeaea",
					"base0E": "#fafafa",
					"base0F": "#ffffff",
				},
			},
			expectError: true,
			errorMsg:    "invalid hex color format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScheme(tt.scheme)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#000000", true},
		{"#fff", true},
		{"#ABCDEF", true},
		{"#12345678", true}, // With alpha
		{"000000", true},    // Without #
		{"fff", true},
		{"#gggggg", false},  // Invalid hex chars
		{"#12345", false},   // Wrong length
		{"#1234567", false}, // Wrong length
		{"not-hex", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := isValidHexColor(tt.color)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid JSON with all fields",
			json: `{
				"name": "Test Scheme",
				"colours": {
					"base00": "#000000",
					"base01": "#111111",
					"base02": "#222222",
					"base03": "#333333",
					"base04": "#444444",
					"base05": "#555555",
					"base06": "#666666",
					"base07": "#777777",
					"base08": "#888888",
					"base09": "#999999",
					"base0A": "#aaaaaa",
					"base0B": "#bbbbbb",
					"base0C": "#cccccc",
					"base0D": "#dddddd",
					"base0E": "#eeeeee",
					"base0F": "#ffffff"
				}
			}`,
			expectError: false,
		},
		{
			name:        "invalid JSON syntax",
			json:        `{"name": "Bad JSON"`,
			expectError: true,
			errorMsg:    "invalid JSON syntax",
		},
		{
			name: "JSON with wrong type",
			json: `{
				"name": 123,
				"colours": {}
			}`,
			expectError: true,
			errorMsg:    "invalid type",
		},
		{
			name: "valid JSON but invalid scheme",
			json: `{
				"name": "Missing Colors",
				"colours": {
					"base00": "#000000"
				}
			}`,
			expectError: true,
			errorMsg:    "missing required color keys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, err := ValidateJSON([]byte(tt.json))

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, scheme)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, scheme)
			}
		})
	}
}

func TestMigrateTextFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkColors map[string]string
	}{
		{
			name: "format with equals sign",
			input: `
				base00=#1a1a1a
				base01=#2a2a2a
				base02=#3a3a3a
				base03=#4a4a4a
				base04=#5a5a5a
				base05=#6a6a6a
				base06=#7a7a7a
				base07=#8a8a8a
				base08=#9a9a9a
				base09=#aaaaaa
				base0A=#bababa
				base0B=#cacaca
				base0C=#dadada
				base0D=#eaeaea
				base0E=#fafafa
				base0F=#ffffff
			`,
			expectError: false,
			checkColors: map[string]string{
				"base00": "#1a1a1a",
				"base0F": "#ffffff",
			},
		},
		{
			name: "format with colon",
			input: `
				base00: #1a1a1a
				base01: #2a2a2a
				base02: #3a3a3a
				base03: #4a4a4a
				base04: #5a5a5a
				base05: #6a6a6a
				base06: #7a7a7a
				base07: #8a8a8a
				base08: #9a9a9a
				base09: #aaaaaa
				base0A: #bababa
				base0B: #cacaca
				base0C: #dadada
				base0D: #eaeaea
				base0E: #fafafa
				base0F: #ffffff
			`,
			expectError: false,
			checkColors: map[string]string{
				"base00": "#1a1a1a",
				"base0F": "#ffffff",
			},
		},
		{
			name: "format with spaces",
			input: `
				base00 1a1a1a
				base01 2a2a2a
				base02 3a3a3a
				base03 4a4a4a
				base04 5a5a5a
				base05 6a6a6a
				base06 7a7a7a
				base07 8a8a8a
				base08 9a9a9a
				base09 aaaaaa
				base0A bababa
				base0B cacaca
				base0C dadada
				base0D eaeaea
				base0E fafafa
				base0F ffffff
			`,
			expectError: false,
			checkColors: map[string]string{
				"base00": "#1a1a1a",
				"base0F": "#ffffff",
			},
		},
		{
			name: "format with comments",
			input: `
				# This is a comment
				base00=#1a1a1a  # Background
				base01=#2a2a2a
				// Another comment style
				base02=#3a3a3a
				base03=#4a4a4a
				base04=#5a5a5a
				base05=#6a6a6a
				base06=#7a7a7a
				base07=#8a8a8a
				base08=#9a9a9a
				base09=#aaaaaa
				base0A=#bababa
				base0B=#cacaca
				base0C=#dadada
				base0D=#eaeaea
				base0E=#fafafa
				base0F=#ffffff
			`,
			expectError: false,
			checkColors: map[string]string{
				"base00": "#1a1a1a",
			},
		},
		{
			name: "incomplete color set",
			input: `
				base00=#000000
				base01=#111111
			`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme, err := migrateTextFormat([]byte(tt.input))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, scheme)

				// Check specific colors
				for key, expectedValue := range tt.checkColors {
					assert.Equal(t, expectedValue, scheme.Colours[key])
				}

				// Check that all required colors are present
				for _, key := range RequiredColorKeys {
					_, exists := scheme.Colours[key]
					assert.True(t, exists, "Missing required key: %s", key)
				}
			}
		})
	}
}

func TestSanitizeScheme(t *testing.T) {
	tests := []struct {
		name     string
		input    *Scheme
		expected *Scheme
	}{
		{
			name: "add hash prefix to colors",
			input: &Scheme{
				Colours: map[string]string{
					"base00": "1a1a1a",
					"base01": "#2a2a2a",
				},
			},
			expected: &Scheme{
				Name:    "Unnamed Scheme",
				Flavour: "default",
				Mode:    "dark",
				Colours: map[string]string{
					"base00": "#1a1a1a",
					"base01": "#2a2a2a",
				},
			},
		},
		{
			name: "set default values",
			input: &Scheme{
				Colours: map[string]string{},
			},
			expected: &Scheme{
				Name:    "Unnamed Scheme",
				Flavour: "default",
				Mode:    "dark",
				Colours: map[string]string{},
			},
		},
		{
			name: "handle nil colors map",
			input: &Scheme{
				Name: "Test",
			},
			expected: &Scheme{
				Name:    "Test",
				Flavour: "default",
				Mode:    "dark",
				Colours: map[string]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SanitizeScheme(tt.input)
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestValidationErrors(t *testing.T) {
	// Test single error
	err := ValidationError{
		Field:   "colours.base00",
		Message: "invalid color",
	}
	assert.Contains(t, err.Error(), "colours.base00")
	assert.Contains(t, err.Error(), "invalid color")

	// Test multiple errors
	errs := ValidationErrors{
		{Field: "name", Message: "is required"},
		{Field: "colours", Message: "missing keys"},
	}
	errStr := errs.Error()
	assert.Contains(t, errStr, "multiple validation errors")
	assert.Contains(t, errStr, "name")
	assert.Contains(t, errStr, "colours")

	// Test empty errors
	emptyErrs := ValidationErrors{}
	assert.Equal(t, "", emptyErrs.Error())
}
