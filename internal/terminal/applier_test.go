package terminal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewApplier(t *testing.T) {
	applier := NewApplier()
	if applier == nil {
		t.Fatal("NewApplier() returned nil")
	}
	if applier.sequenceBuilder == nil {
		t.Fatal("NewApplier() did not initialize sequenceBuilder")
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"0", true},
		{"1", true},
		{"123", true},
		{"abc", false},
		{"12a", false},
		{"a12", false},
		{"ptmx", false},
	}

	for _, test := range tests {
		result := isNumeric(test.input)
		if result != test.expected {
			t.Errorf("isNumeric(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestCheckWritePermission(t *testing.T) {
	applier := NewApplier()

	// Test with a non-existent file
	nonExistentPath := "/dev/pts/999999"
	writable := applier.checkWritePermission(nonExistentPath)
	if writable {
		t.Errorf("checkWritePermission(%q) = true, expected false for non-existent file", nonExistentPath)
	}

	// Test with /dev/null (should be writable)
	devNullPath := "/dev/null"
	if _, err := os.Stat(devNullPath); err == nil {
		writable := applier.checkWritePermission(devNullPath)
		if !writable {
			t.Errorf("checkWritePermission(%q) = false, expected true for /dev/null", devNullPath)
		}
	}
}

func TestDetectActiveTerminals(t *testing.T) {
	applier := NewApplier()

	// This test will vary based on the system state
	// We'll just ensure it doesn't crash and returns a slice
	terminals, err := applier.detectActiveTerminals()
	if err != nil {
		// On some systems /dev/pts might not exist or be accessible
		// This is acceptable for testing
		t.Logf("detectActiveTerminals() returned error (expected on some systems): %v", err)
		return
	}

	if terminals == nil {
		t.Fatal("detectActiveTerminals() returned nil slice")
	}

	// Log the number of terminals found for debugging
	t.Logf("Found %d active terminals", len(terminals))

	// Validate each terminal device
	for i, terminal := range terminals {
		if terminal == nil {
			t.Errorf("Terminal %d is nil", i)
			continue
		}
		if terminal.Path == "" {
			t.Errorf("Terminal %d has empty path", i)
		}
		if !filepath.IsAbs(terminal.Path) {
			t.Errorf("Terminal %d path %q is not absolute", i, terminal.Path)
		}
		if terminal.PID < 0 {
			t.Errorf("Terminal %d has invalid PID %d", i, terminal.PID)
		}
	}
}

func TestGetActiveTerminalCount(t *testing.T) {
	applier := NewApplier()

	count, err := applier.GetActiveTerminalCount()
	if err != nil {
		// On some systems /dev/pts might not exist or be accessible
		t.Logf("GetActiveTerminalCount() returned error (expected on some systems): %v", err)
		return
	}

	if count < 0 {
		t.Errorf("GetActiveTerminalCount() = %d, expected non-negative", count)
	}

	t.Logf("Active terminal count: %d", count)
}

func TestListActiveTerminals(t *testing.T) {
	applier := NewApplier()

	terminals, err := applier.ListActiveTerminals()
	if err != nil {
		// On some systems /dev/pts might not exist or be accessible
		t.Logf("ListActiveTerminals() returned error (expected on some systems): %v", err)
		return
	}

	if terminals == nil {
		t.Fatal("ListActiveTerminals() returned nil")
	}

	t.Logf("Listed %d active terminals", len(terminals))
}

func TestApplyToTerminals(t *testing.T) {
	applier := NewApplier()

	// Test with sample colors
	colours := map[string]string{
		"colour0":    "1a1b26",
		"colour1":    "f7768e",
		"colour2":    "9ece6a",
		"colour3":    "e0af68",
		"colour4":    "7aa2f7",
		"colour5":    "bb9af7",
		"colour6":    "7dcfff",
		"colour7":    "c0caf5",
		"colour8":    "414868",
		"colour9":    "f7768e",
		"colour10":   "9ece6a",
		"colour11":   "e0af68",
		"colour12":   "7aa2f7",
		"colour13":   "bb9af7",
		"colour14":   "7dcfff",
		"colour15":   "c0caf5",
		"background": "1a1b26",
		"foreground": "c0caf5",
		"cursor":     "c0caf5",
	}

	// This should not fail even if no terminals are found
	err := applier.ApplyToTerminals(colours, "test-scheme")
	if err != nil {
		t.Errorf("ApplyToTerminals() returned error: %v", err)
	}
}

func TestApplySequencesWithFallback(t *testing.T) {
	applier := NewApplier()

	// Test with sample colors
	colours := map[string]string{
		"colour0":    "1a1b26",
		"colour1":    "f7768e",
		"background": "1a1b26",
		"foreground": "c0caf5",
	}

	// This should never fail due to fallback behavior
	err := applier.ApplySequencesWithFallback(colours, "test-scheme")
	if err != nil {
		t.Errorf("ApplySequencesWithFallback() returned error: %v", err)
	}
}

func TestApplyToTerminalsWithInvalidColors(t *testing.T) {
	applier := NewApplier()

	// Test with invalid colors
	colours := map[string]string{
		"colour0": "invalid",
		"colour1": "gggggg",
	}

	err := applier.ApplyToTerminals(colours, "test-scheme")
	if err == nil {
		t.Error("ApplyToTerminals() should have failed with invalid colors")
	}
}

func TestProcessUsesTerminal(t *testing.T) {
	applier := NewApplier()

	// Test with invalid PID
	result := applier.processUsesTerminal(-1, "0")
	if result {
		t.Error("processUsesTerminal() should return false for invalid PID")
	}

	// Test with non-existent PID
	result = applier.processUsesTerminal(999999, "0")
	if result {
		t.Error("processUsesTerminal() should return false for non-existent PID")
	}
}

func TestFindTerminalPID(t *testing.T) {
	applier := NewApplier()

	// Test with non-existent terminal
	_, err := applier.findTerminalPID("/dev/pts/999999")
	if err == nil {
		t.Error("findTerminalPID() should return error for non-existent terminal")
	}
}

func TestCheckTerminalDevice(t *testing.T) {
	applier := NewApplier()

	// Test with non-existent device
	_, err := applier.checkTerminalDevice("/dev/pts/999999")
	if err == nil {
		t.Error("checkTerminalDevice() should return error for non-existent device")
	}

	// Test with non-device file (if /tmp exists)
	if _, err := os.Stat("/tmp"); err == nil {
		_, err := applier.checkTerminalDevice("/tmp")
		if err == nil {
			t.Error("checkTerminalDevice() should return error for non-device file")
		}
	}
}

// Benchmark tests
func BenchmarkDetectActiveTerminals(b *testing.B) {
	applier := NewApplier()

	for i := 0; i < b.N; i++ {
		_, _ = applier.detectActiveTerminals()
	}
}

func BenchmarkIsNumeric(b *testing.B) {
	testStrings := []string{"0", "123", "abc", "12a", "ptmx"}

	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			isNumeric(s)
		}
	}
}

func BenchmarkApplyToTerminals(b *testing.B) {
	applier := NewApplier()
	colours := map[string]string{
		"colour0":    "1a1b26",
		"colour1":    "f7768e",
		"background": "1a1b26",
		"foreground": "c0caf5",
	}

	for i := 0; i < b.N; i++ {
		_ = applier.ApplyToTerminals(colours, "test-scheme")
	}
}
