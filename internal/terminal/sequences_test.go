package terminal

import (
	"strings"
	"testing"
)

func TestSequenceBuilder_GenerateSequences(t *testing.T) {
	sb := NewSequenceBuilder()

	// Test with sample catppuccin mocha colors
	colours := map[string]string{
		"colour0":    "1e1e2e",
		"colour1":    "f38ba8",
		"colour2":    "a6e3a1",
		"colour3":    "f9e2af",
		"colour4":    "89b4fa",
		"colour5":    "f5c2e7",
		"colour6":    "94e2d5",
		"colour7":    "bac2de",
		"colour8":    "585b70",
		"colour9":    "f38ba8",
		"colour10":   "a6e3a1",
		"colour11":   "f9e2af",
		"colour12":   "89b4fa",
		"colour13":   "f5c2e7",
		"colour14":   "94e2d5",
		"colour15":   "a6adc8",
		"background": "1e1e2e",
		"foreground": "cdd6f4",
		"cursor":     "f5e0dc",
	}

	sequences, err := sb.GenerateSequences(colours)
	if err != nil {
		t.Fatalf("GenerateSequences failed: %v", err)
	}

	// Should have 16 standard colors + 3 special colors = 19 sequences
	expectedCount := 19
	if len(sequences) != expectedCount {
		t.Errorf("Expected %d sequences, got %d", expectedCount, len(sequences))
	}

	// Test first sequence (colour0)
	expectedFirst := "\\033]4;0;rgb:1e/1e/2e\\033\\\\"
	if sequences[0] != expectedFirst {
		t.Errorf("Expected first sequence %q, got %q", expectedFirst, sequences[0])
	}

	// Test background sequence (should be index 256)
	found := false
	expectedBg := "\\033]4;256;rgb:1e/1e/2e\\033\\\\"
	for _, seq := range sequences {
		if seq == expectedBg {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected background sequence %q not found", expectedBg)
	}
}

func TestSequenceBuilder_buildColorSequence(t *testing.T) {
	sb := NewSequenceBuilder()

	tests := []struct {
		name     string
		index    int
		hexColor string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid color without #",
			index:    0,
			hexColor: "1a1b26",
			expected: "\\033]4;0;rgb:1a/1b/26\\033\\\\",
			wantErr:  false,
		},
		{
			name:     "valid color with #",
			index:    1,
			hexColor: "#f7768e",
			expected: "\\033]4;1;rgb:f7/76/8e\\033\\\\",
			wantErr:  false,
		},
		{
			name:     "special color index",
			index:    256,
			hexColor: "1a1b26",
			expected: "\\033]4;256;rgb:1a/1b/26\\033\\\\",
			wantErr:  false,
		},
		{
			name:     "invalid hex length",
			index:    0,
			hexColor: "1a1b2",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid hex characters",
			index:    0,
			hexColor: "gghhii",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sb.buildColorSequence(tt.index, tt.hexColor)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSequenceBuilder_parseHexColor(t *testing.T) {
	sb := NewSequenceBuilder()

	tests := []struct {
		name    string
		hex     string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantErr bool
	}{
		{
			name:    "valid hex",
			hex:     "1a1b26",
			wantR:   0x1a,
			wantG:   0x1b,
			wantB:   0x26,
			wantErr: false,
		},
		{
			name:    "white color",
			hex:     "ffffff",
			wantR:   255,
			wantG:   255,
			wantB:   255,
			wantErr: false,
		},
		{
			name:    "black color",
			hex:     "000000",
			wantR:   0,
			wantG:   0,
			wantB:   0,
			wantErr: false,
		},
		{
			name:    "invalid length",
			hex:     "1a1b2",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			hex:     "gghhii",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, err := sb.parseHexColor(tt.hex)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("Expected RGB(%d, %d, %d), got RGB(%d, %d, %d)",
					tt.wantR, tt.wantG, tt.wantB, r, g, b)
			}
		})
	}
}

func TestSequenceBuilder_ValidateSequenceFormat(t *testing.T) {
	sb := NewSequenceBuilder()

	tests := []struct {
		name     string
		sequence string
		wantErr  bool
	}{
		{
			name:     "valid sequence",
			sequence: "\\033]4;0;rgb:1a/1b/26\\033\\\\",
			wantErr:  false,
		},
		{
			name:     "valid special color",
			sequence: "\\033]4;256;rgb:ff/ff/ff\\033\\\\",
			wantErr:  false,
		},
		{
			name:     "missing start",
			sequence: "4;0;rgb:1a/1b/26\\033\\\\",
			wantErr:  true,
		},
		{
			name:     "missing end",
			sequence: "\\033]4;0;rgb:1a/1b/26",
			wantErr:  true,
		},
		{
			name:     "invalid format",
			sequence: "\\033]4;0;1a1b26\\033\\\\",
			wantErr:  true,
		},
		{
			name:     "invalid RGB format",
			sequence: "\\033]4;0;rgb:1a-1b-26\\033\\\\",
			wantErr:  true,
		},
		{
			name:     "invalid hex in RGB",
			sequence: "\\033]4;0;rgb:gg/hh/ii\\033\\\\",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sb.ValidateSequenceFormat(tt.sequence)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSequenceBuilder_FormatSequencesForShell(t *testing.T) {
	sb := NewSequenceBuilder()

	sequences := []string{
		"\\033]4;0;rgb:1a/1b/26\\033\\\\",
		"\\033]4;1;rgb:f7/76/8e\\033\\\\",
	}

	result := sb.FormatSequencesForShell(sequences)

	// Check that it starts with shebang
	if !strings.HasPrefix(result, "#!/bin/bash") {
		t.Errorf("Expected shell script to start with shebang")
	}

	// Check that it contains printf statements
	if !strings.Contains(result, "printf") {
		t.Errorf("Expected shell script to contain printf statements")
	}

	// Check that escape sequences are converted
	if strings.Contains(result, "\\033") {
		t.Errorf("Expected escape sequences to be converted for shell")
	}

	// Check that it contains actual escape characters
	if !strings.Contains(result, "\033") {
		t.Errorf("Expected shell script to contain actual escape characters")
	}
}

func TestSequenceBuilder_MissingColors(t *testing.T) {
	sb := NewSequenceBuilder()

	// Test with missing colors
	colours := map[string]string{
		"colour0": "1a1b26",
		"colour1": "f7768e",
		// Missing colour2-colour15
		"background": "1a1b26",
		// Missing foreground and cursor
	}

	sequences, err := sb.GenerateSequences(colours)
	if err != nil {
		t.Fatalf("GenerateSequences failed: %v", err)
	}

	// Should only have sequences for colors that exist
	expectedCount := 3 // colour0, colour1, background
	if len(sequences) != expectedCount {
		t.Errorf("Expected %d sequences, got %d", expectedCount, len(sequences))
	}
}
