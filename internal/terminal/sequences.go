package terminal

import (
	"fmt"
	"strings"
)

// SequenceBuilder generates ANSI escape sequences for terminal color application
type SequenceBuilder struct{}

// NewSequenceBuilder creates a new sequence builder
func NewSequenceBuilder() *SequenceBuilder {
	return &SequenceBuilder{}
}

// GenerateSequences creates ANSI escape sequences for all colors
// Format: \033]4;0;rgb:1a/1b/26\033\\
func (sb *SequenceBuilder) GenerateSequences(colours map[string]string) ([]string, error) {
	var sequences []string

	// Generate sequences for 16 standard colors (colour0-colour15)
	for i := 0; i < 16; i++ {
		colourKey := fmt.Sprintf("colour%d", i)
		if hexValue, exists := colours[colourKey]; exists {
			sequence, err := sb.buildColorSequence(i, hexValue)
			if err != nil {
				return nil, fmt.Errorf("failed to build sequence for %s: %w", colourKey, err)
			}
			sequences = append(sequences, sequence)
		}
	}

	// Generate sequences for special colors
	specialColors := map[string]int{
		"background": 256, // Background color
		"foreground": 257, // Foreground color
		"cursor":     258, // Cursor color
	}

	for colorName, colorIndex := range specialColors {
		if hexValue, exists := colours[colorName]; exists {
			sequence, err := sb.buildColorSequence(colorIndex, hexValue)
			if err != nil {
				return nil, fmt.Errorf("failed to build sequence for %s: %w", colorName, err)
			}
			sequences = append(sequences, sequence)
		}
	}

	return sequences, nil
}

// buildColorSequence creates a single ANSI escape sequence for a color
func (sb *SequenceBuilder) buildColorSequence(index int, hexColor string) (string, error) {
	// Remove # prefix if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Validate hex color format
	if len(hexColor) != 6 {
		return "", fmt.Errorf("invalid hex color format: %s (expected 6 characters)", hexColor)
	}

	// Parse RGB components
	r, g, b, err := sb.parseHexColor(hexColor)
	if err != nil {
		return "", fmt.Errorf("failed to parse hex color %s: %w", hexColor, err)
	}

	// Build ANSI sequence: \033]4;index;rgb:rr/gg/bb\033\\
	sequence := fmt.Sprintf("\\033]4;%d;rgb:%02x/%02x/%02x\\033\\\\", index, r, g, b)

	return sequence, nil
}

// parseHexColor converts hex string to RGB values
func (sb *SequenceBuilder) parseHexColor(hex string) (r, g, b uint8, err error) {
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("hex color must be 6 characters long")
	}

	// Parse red component
	var rInt, gInt, bInt int
	if _, err := fmt.Sscanf(hex[0:2], "%x", &rInt); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid red component: %s", hex[0:2])
	}

	// Parse green component
	if _, err := fmt.Sscanf(hex[2:4], "%x", &gInt); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid green component: %s", hex[2:4])
	}

	// Parse blue component
	if _, err := fmt.Sscanf(hex[4:6], "%x", &bInt); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid blue component: %s", hex[4:6])
	}

	return uint8(rInt), uint8(gInt), uint8(bInt), nil
}

// ValidateSequenceFormat validates that a sequence matches the expected ANSI format
func (sb *SequenceBuilder) ValidateSequenceFormat(sequence string) error {
	// Check if sequence starts with \033]4;
	if !strings.HasPrefix(sequence, "\\033]4;") {
		return fmt.Errorf("sequence must start with \\033]4;")
	}

	// Check if sequence ends with \033\\
	if !strings.HasSuffix(sequence, "\\033\\\\") {
		return fmt.Errorf("sequence must end with \\033\\\\")
	}

	// Extract the middle part (index;rgb:rr/gg/bb)
	middle := strings.TrimPrefix(sequence, "\\033]4;")
	middle = strings.TrimSuffix(middle, "\\033\\\\")

	// Split by semicolon to get index and rgb parts
	parts := strings.Split(middle, ";")
	if len(parts) != 2 {
		return fmt.Errorf("sequence must have format \\033]4;index;rgb:rr/gg/bb\\033\\\\")
	}

	// Validate index is numeric
	var index int
	if _, err := fmt.Sscanf(parts[0], "%d", &index); err != nil {
		return fmt.Errorf("invalid color index: %s", parts[0])
	}

	// Validate RGB format
	rgbPart := parts[1]
	if !strings.HasPrefix(rgbPart, "rgb:") {
		return fmt.Errorf("RGB part must start with 'rgb:'")
	}

	// Extract RGB values
	rgbValues := strings.TrimPrefix(rgbPart, "rgb:")
	rgbComponents := strings.Split(rgbValues, "/")
	if len(rgbComponents) != 3 {
		return fmt.Errorf("RGB values must be in format rr/gg/bb")
	}

	// Validate each RGB component is valid hex
	for i, component := range rgbComponents {
		if len(component) != 2 {
			return fmt.Errorf("RGB component %d must be 2 hex digits", i)
		}
		var value int
		if _, err := fmt.Sscanf(component, "%x", &value); err != nil {
			return fmt.Errorf("invalid hex value in RGB component %d: %s", i, component)
		}
	}

	return nil
}

// FormatSequencesForShell formats sequences for shell sourcing
func (sb *SequenceBuilder) FormatSequencesForShell(sequences []string) string {
	var builder strings.Builder

	builder.WriteString("#!/bin/bash\n")
	builder.WriteString("# Heimdall Terminal Color Sequences\n")
	builder.WriteString("# Generated automatically - source this file to apply colors\n\n")

	for _, sequence := range sequences {
		// Convert escape sequences to actual escape characters for shell
		shellSequence := strings.ReplaceAll(sequence, "\\033", "\033")
		shellSequence = strings.ReplaceAll(shellSequence, "\\\\", "\\")

		builder.WriteString(fmt.Sprintf("printf '%s'\n", shellSequence))
	}

	builder.WriteString("\n# End of sequences\n")

	return builder.String()
}
