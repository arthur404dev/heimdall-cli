package manager

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		hasError bool
	}{
		{
			name:     "seconds as number",
			input:    "30",
			expected: 30 * time.Second,
			hasError: false,
		},
		{
			name:     "standard go duration",
			input:    "30m",
			expected: 30 * time.Minute,
			hasError: false,
		},
		{
			name:     "standard go duration with seconds",
			input:    "1h30m45s",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
			hasError: false,
		},
		{
			name:     "hours and minutes",
			input:    "2h30m",
			expected: 2*time.Hour + 30*time.Minute,
			hasError: false,
		},
		{
			name:     "hours and minutes with space",
			input:    "2h 30m",
			expected: 2*time.Hour + 30*time.Minute,
			hasError: false,
		},
		{
			name:     "decimal hours",
			input:    "1.5h",
			expected: 90 * time.Minute,
			hasError: false,
		},
		{
			name:     "decimal hours with fraction",
			input:    "2.25h",
			expected: 2*time.Hour + 15*time.Minute,
			hasError: false,
		},
		{
			name:     "only minutes",
			input:    "45m",
			expected: 45 * time.Minute,
			hasError: false,
		},
		{
			name:     "only seconds",
			input:    "120s",
			expected: 120 * time.Second,
			hasError: false,
		},
		{
			name:     "complex duration",
			input:    "3h15m30s",
			expected: 3*time.Hour + 15*time.Minute + 30*time.Second,
			hasError: false,
		},
		{
			name:     "invalid format",
			input:    "invalid",
			expected: 0,
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "zero duration",
			input:    "0",
			expected: 0,
			hasError: true,
		},
		{
			name:     "negative duration",
			input:    "-30m",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid unit",
			input:    "30x",
			expected: 0,
			hasError: true,
		},
		{
			name:     "mixed valid and invalid",
			input:    "1h30x",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %s: %s", tt.input, err.Error())
				}
				if result != tt.expected {
					t.Errorf("Expected %v for input %s, got %v", tt.expected, tt.input, result)
				}
			}
		})
	}
}

func TestParseDurationEdgeCases(t *testing.T) {
	t.Run("very large number", func(t *testing.T) {
		result, err := ParseDuration("999999999")
		if err != nil {
			t.Errorf("Should handle large numbers: %s", err.Error())
		}
		expected := 999999999 * time.Second
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("fractional seconds in decimal hours", func(t *testing.T) {
		result, err := ParseDuration("0.5h")
		if err != nil {
			t.Errorf("Should handle fractional hours: %s", err.Error())
		}
		expected := 30 * time.Minute
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("multiple spaces", func(t *testing.T) {
		result, err := ParseDuration("1h  30m   15s")
		if err != nil {
			t.Errorf("Should handle multiple spaces: %s", err.Error())
		}
		expected := 1*time.Hour + 30*time.Minute + 15*time.Second
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("case sensitivity", func(t *testing.T) {
		// Should work with lowercase
		result, err := ParseDuration("1h30m")
		if err != nil {
			t.Errorf("Should handle lowercase: %s", err.Error())
		}
		expected := 1*time.Hour + 30*time.Minute
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{
			name:     "zero duration",
			input:    0,
			expected: "unlimited",
		},
		{
			name:     "only seconds",
			input:    45 * time.Second,
			expected: "45s",
		},
		{
			name:     "only minutes",
			input:    30 * time.Minute,
			expected: "30m",
		},
		{
			name:     "only hours",
			input:    2 * time.Hour,
			expected: "2h",
		},
		{
			name:     "minutes and seconds",
			input:    5*time.Minute + 30*time.Second,
			expected: "5m 30s",
		},
		{
			name:     "hours and minutes",
			input:    2*time.Hour + 30*time.Minute,
			expected: "2h 30m",
		},
		{
			name:     "hours and seconds",
			input:    1*time.Hour + 45*time.Second,
			expected: "1h 45s",
		},
		{
			name:     "hours, minutes, and seconds",
			input:    2*time.Hour + 15*time.Minute + 30*time.Second,
			expected: "2h 15m 30s",
		},
		{
			name:     "exactly one hour",
			input:    1 * time.Hour,
			expected: "1h",
		},
		{
			name:     "exactly one minute",
			input:    1 * time.Minute,
			expected: "1m",
		},
		{
			name:     "exactly one second",
			input:    1 * time.Second,
			expected: "1s",
		},
		{
			name:     "large duration",
			input:    25*time.Hour + 61*time.Minute + 75*time.Second,
			expected: "26h 2m 15s", // 61m = 1h1m, 75s = 1m15s
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s for duration %v, got %s", tt.expected, tt.input, result)
			}
		})
	}
}

func TestFormatDurationEdgeCases(t *testing.T) {
	t.Run("very large duration", func(t *testing.T) {
		duration := 1000*time.Hour + 500*time.Minute + 200*time.Second
		result := FormatDuration(duration)

		// Should handle overflow correctly
		// 500m = 8h20m, 200s = 3m20s
		// Total: 1000h + 8h + 20m + 3m + 20s = 1008h 23m 20s
		expected := "1008h 23m 20s"
		if result != expected {
			t.Errorf("Expected %s for large duration, got %s", expected, result)
		}
	})

	t.Run("fractional seconds", func(t *testing.T) {
		// Go duration truncates fractional seconds
		duration := 1*time.Hour + 30*time.Minute + 500*time.Millisecond
		result := FormatDuration(duration)
		expected := "1h 30m" // Milliseconds are truncated
		if result != expected {
			t.Errorf("Expected %s for fractional duration, got %s", expected, result)
		}
	})

	t.Run("negative duration", func(t *testing.T) {
		// Negative durations should be handled gracefully
		duration := -30 * time.Minute
		result := FormatDuration(duration)
		// The function should handle this case (implementation dependent)
		if result == "" {
			t.Error("Should handle negative duration gracefully")
		}
	})
}

// Integration tests
func TestParseDurationRoundTrip(t *testing.T) {
	testCases := []string{
		"30s",
		"5m",
		"2h",
		"1h30m",
		"2h15m30s",
		"45m30s",
	}

	for _, input := range testCases {
		t.Run("round trip "+input, func(t *testing.T) {
			// Parse the duration
			parsed, err := ParseDuration(input)
			if err != nil {
				t.Fatalf("Failed to parse %s: %s", input, err.Error())
			}

			// Format it back
			formatted := FormatDuration(parsed)

			// Parse the formatted version
			reparsed, err := ParseDuration(formatted)
			if err != nil {
				t.Fatalf("Failed to reparse %s: %s", formatted, err.Error())
			}

			// Should be the same
			if parsed != reparsed {
				t.Errorf("Round trip failed: %s -> %v -> %s -> %v", input, parsed, formatted, reparsed)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseDuration(b *testing.B) {
	testCases := []string{
		"30",
		"30m",
		"1h30m",
		"2h15m30s",
		"1.5h",
	}

	for _, tc := range testCases {
		b.Run("parse_"+tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ParseDuration(tc)
				if err != nil {
					b.Fatalf("Parse error: %s", err.Error())
				}
			}
		})
	}
}

func BenchmarkFormatDuration(b *testing.B) {
	testCases := []time.Duration{
		30 * time.Second,
		30 * time.Minute,
		1*time.Hour + 30*time.Minute,
		2*time.Hour + 15*time.Minute + 30*time.Second,
	}

	for i, tc := range testCases {
		b.Run("format_"+tc.String(), func(b *testing.B) {
			for j := 0; j < b.N; j++ {
				result := FormatDuration(tc)
				_ = result
			}
		})
		_ = i // Avoid unused variable
	}
}

// Property-based testing
func TestParseDurationProperties(t *testing.T) {
	t.Run("parsing zero always fails", func(t *testing.T) {
		zeroInputs := []string{"0", "0s", "0m", "0h"}
		for _, input := range zeroInputs {
			_, err := ParseDuration(input)
			if err == nil {
				t.Errorf("Expected error for zero duration input: %s", input)
			}
		}
	})

	t.Run("parsing positive durations succeeds", func(t *testing.T) {
		positiveInputs := []string{"1s", "1m", "1h", "30", "1h30m45s"}
		for _, input := range positiveInputs {
			result, err := ParseDuration(input)
			if err != nil {
				t.Errorf("Unexpected error for positive duration %s: %s", input, err.Error())
			}
			if result <= 0 {
				t.Errorf("Expected positive duration for %s, got %v", input, result)
			}
		}
	})

	t.Run("formatting never returns empty string for positive durations", func(t *testing.T) {
		durations := []time.Duration{
			1 * time.Second,
			1 * time.Minute,
			1 * time.Hour,
			1*time.Hour + 30*time.Minute + 45*time.Second,
		}
		for _, d := range durations {
			result := FormatDuration(d)
			if result == "" {
				t.Errorf("FormatDuration returned empty string for %v", d)
			}
		}
	})
}

// Error message tests
func TestParseDurationErrorMessages(t *testing.T) {
	errorCases := []struct {
		input       string
		expectedMsg string
	}{
		{
			input:       "invalid",
			expectedMsg: "invalid duration format",
		},
		{
			input:       "0",
			expectedMsg: "duration cannot be zero",
		},
		{
			input:       "",
			expectedMsg: "invalid duration format",
		},
		{
			input:       "30x",
			expectedMsg: "unknown time unit",
		},
	}

	for _, tc := range errorCases {
		t.Run("error_"+tc.input, func(t *testing.T) {
			_, err := ParseDuration(tc.input)
			if err == nil {
				t.Errorf("Expected error for input %s", tc.input)
			}
			if err != nil && !containsSubstring(err.Error(), tc.expectedMsg) {
				t.Errorf("Expected error message to contain '%s', got '%s'", tc.expectedMsg, err.Error())
			}
		})
	}
}

// Test utilities - using containsSubstring from session_test.go
