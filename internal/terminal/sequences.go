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
// Format: OSC sequences for terminal color configuration
func (sb *SequenceBuilder) GenerateSequences(colours map[string]string) ([]string, error) {
	var sequences []string

	// Generate sequences for special colors first (using OSC 10, 11, 12)
	// These are more universally supported than the extended indices
	// Try both "foreground" and "text" for compatibility
	if fg, exists := colours["foreground"]; exists {
		sequences = append(sequences, sb.buildOSCSequence(10, fg)) // OSC 10 - foreground
	} else if fg, exists := colours["text"]; exists {
		sequences = append(sequences, sb.buildOSCSequence(10, fg)) // OSC 10 - foreground
	}

	if bg, exists := colours["background"]; exists {
		sequences = append(sequences, sb.buildOSCSequence(11, bg)) // OSC 11 - background
	}

	if cursor, exists := colours["cursor"]; exists {
		sequences = append(sequences, sb.buildOSCSequence(12, cursor)) // OSC 12 - cursor
	} else if cursor, exists := colours["text"]; exists {
		// Use text color for cursor if no explicit cursor color
		sequences = append(sequences, sb.buildOSCSequence(12, cursor)) // OSC 12 - cursor
	}

	// Generate sequences for 16 standard colors
	// Try both "colour" and "term" prefixes for compatibility
	for i := 0; i < 16; i++ {
		colourKey := fmt.Sprintf("colour%d", i)
		termKey := fmt.Sprintf("term%d", i)

		var hexValue string
		var exists bool

		if hexValue, exists = colours[colourKey]; !exists {
			hexValue, exists = colours[termKey]
		}

		if exists {
			sequence, err := sb.buildColorSequence(i, hexValue)
			if err != nil {
				return nil, fmt.Errorf("failed to build sequence for color %d: %w", i, err)
			}
			sequences = append(sequences, sequence)
		}
	}

	return sequences, nil
}

// buildOSCSequence creates an OSC sequence for special colors
func (sb *SequenceBuilder) buildOSCSequence(code int, hexColor string) string {
	// Remove # prefix if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Return raw escape sequence (without printf wrapper for direct terminal application)
	return fmt.Sprintf("\\033]%d;#%s\\007", code, hexColor)
}

// buildColorSequence creates a single ANSI escape sequence for a color
func (sb *SequenceBuilder) buildColorSequence(index int, hexColor string) (string, error) {
	// Remove # prefix if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Validate hex color format
	if len(hexColor) != 6 {
		return "", fmt.Errorf("invalid hex color format: %s (expected 6 characters)", hexColor)
	}

	// Return raw escape sequence (without printf wrapper for direct terminal application)
	sequence := fmt.Sprintf("\\033]4;%d;#%s\\007", index, hexColor)

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
func (sb *SequenceBuilder) FormatSequencesForShell(sequences []string, schemeName string) string {
	var builder strings.Builder

	builder.WriteString("#!/bin/bash\n")
	builder.WriteString("# Heimdall Terminal Color Sequences\n")
	builder.WriteString(fmt.Sprintf("# Scheme: %s\n", schemeName))
	builder.WriteString("# Generated automatically - source this file to apply colors\n\n")

	// Add color comments for clarity
	builder.WriteString("# Special colors\n")
	for i, sequence := range sequences {
		if i < 3 { // First 3 are special colors (foreground, background, cursor)
			// Wrap raw sequences in printf for shell execution
			builder.WriteString(fmt.Sprintf("printf '%s'\n", sequence))
		}
	}

	if len(sequences) > 3 {
		builder.WriteString("\n# Standard colors (0-15)\n")
		colorNames := []string{
			"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
			"bright black", "bright red", "bright green", "bright yellow",
			"bright blue", "bright magenta", "bright cyan", "bright white",
		}

		for i := 3; i < len(sequences) && i-3 < len(colorNames); i++ {
			// Wrap raw sequences in printf for shell execution
			builder.WriteString(fmt.Sprintf("printf '%s'  # %s\n", sequences[i], colorNames[i-3]))
		}
	}

	builder.WriteString("\n# End of sequences\n")

	return builder.String()
}
