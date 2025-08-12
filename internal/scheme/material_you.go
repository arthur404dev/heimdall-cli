package scheme

import (
	"fmt"
	"strings"
)

// MaterialYouColors represents a full Material You color scheme
type MaterialYouColors map[string]string

// ConvertToMaterialYou converts a traditional color scheme to Material You format
func ConvertToMaterialYou(colors map[string]string, variant string) MaterialYouColors {
	m := make(MaterialYouColors)

	// Get base colors
	bg := colors["background"]
	fg := colors["foreground"]

	// Map terminal colors directly
	for i := 0; i < 16; i++ {
		if c, ok := colors[fmt.Sprintf("color%d", i)]; ok {
			m[fmt.Sprintf("term%d", i)] = strings.TrimPrefix(c, "#")
		}
	}

	// Set Material Design 3 surface colors based on background
	m["background"] = strings.TrimPrefix(bg, "#")
	m["surface"] = strings.TrimPrefix(bg, "#")
	m["base"] = strings.TrimPrefix(bg, "#")
	m["mantle"] = strings.TrimPrefix(bg, "#")
	m["crust"] = darken(bg, 0.05)

	// Surface variants
	m["surfaceDim"] = strings.TrimPrefix(bg, "#")
	m["surfaceBright"] = lighten(bg, 0.15)
	m["surfaceContainerLowest"] = darken(bg, 0.08)
	m["surfaceContainerLow"] = lighten(bg, 0.05)
	m["surfaceContainer"] = lighten(bg, 0.08)
	m["surfaceContainerHigh"] = lighten(bg, 0.11)
	m["surfaceContainerHighest"] = lighten(bg, 0.15)

	// Surface overlays
	m["surface0"] = lighten(bg, 0.05)
	m["surface1"] = lighten(bg, 0.10)
	m["surface2"] = lighten(bg, 0.15)
	m["overlay0"] = lighten(bg, 0.20)
	m["overlay1"] = lighten(bg, 0.25)
	m["overlay2"] = lighten(bg, 0.30)

	// Text colors
	m["text"] = strings.TrimPrefix(fg, "#")
	m["onBackground"] = strings.TrimPrefix(fg, "#")
	m["onSurface"] = strings.TrimPrefix(fg, "#")
	m["subtext1"] = darken(fg, 0.10)
	m["subtext0"] = darken(fg, 0.20)

	// Get primary color (usually blue - color4)
	primary := getColorOrDefault(colors, "color4", "#89B4FA")
	m["primary"] = strings.TrimPrefix(primary, "#")
	m["primary_paletteKeyColor"] = darken(primary, 0.20)
	m["primaryContainer"] = darken(primary, 0.30)
	m["onPrimary"] = contrastColor(primary)
	m["onPrimaryContainer"] = "#ffffff"
	m["inversePrimary"] = darken(primary, 0.40)
	m["primaryFixed"] = lighten(primary, 0.10)
	m["primaryFixedDim"] = strings.TrimPrefix(primary, "#")
	m["onPrimaryFixed"] = darken(primary, 0.60)
	m["onPrimaryFixedVariant"] = darken(primary, 0.40)

	// Secondary color (usually cyan - color6)
	secondary := getColorOrDefault(colors, "color6", "#94E2D5")
	m["secondary"] = strings.TrimPrefix(secondary, "#")
	m["secondary_paletteKeyColor"] = darken(secondary, 0.20)
	m["secondaryContainer"] = darken(secondary, 0.30)
	m["onSecondary"] = contrastColor(secondary)
	m["onSecondaryContainer"] = lighten(secondary, 0.40)
	m["secondaryFixed"] = lighten(secondary, 0.10)
	m["secondaryFixedDim"] = strings.TrimPrefix(secondary, "#")
	m["onSecondaryFixed"] = darken(secondary, 0.60)
	m["onSecondaryFixedVariant"] = darken(secondary, 0.40)

	// Tertiary color (usually magenta - color5)
	tertiary := getColorOrDefault(colors, "color5", "#F5C2E7")
	m["tertiary"] = strings.TrimPrefix(tertiary, "#")
	m["tertiary_paletteKeyColor"] = darken(tertiary, 0.20)
	m["tertiaryContainer"] = darken(tertiary, 0.30)
	m["onTertiary"] = contrastColor(tertiary)
	m["onTertiaryContainer"] = "#000000"
	m["tertiaryFixed"] = lighten(tertiary, 0.10)
	m["tertiaryFixedDim"] = strings.TrimPrefix(tertiary, "#")
	m["onTertiaryFixed"] = darken(tertiary, 0.60)
	m["onTertiaryFixedVariant"] = darken(tertiary, 0.40)

	// Error color (usually red - color1)
	errorColor := getColorOrDefault(colors, "color1", "#F38BA8")
	m["error"] = strings.TrimPrefix(errorColor, "#")
	m["onError"] = contrastColor(errorColor)
	m["errorContainer"] = darken(errorColor, 0.40)
	m["onErrorContainer"] = lighten(errorColor, 0.40)

	// Success colors (usually green - color2)
	success := getColorOrDefault(colors, "color2", "#A6E3A1")
	m["success"] = strings.TrimPrefix(success, "#")
	m["onSuccess"] = contrastColor(success)
	m["successContainer"] = darken(success, 0.30)
	m["onSuccessContainer"] = lighten(success, 0.40)

	// Neutral colors
	m["neutral_paletteKeyColor"] = darken(fg, 0.30)
	m["neutral_variant_paletteKeyColor"] = darken(fg, 0.35)

	// Surface variants
	m["surfaceVariant"] = lighten(bg, 0.15)
	m["onSurfaceVariant"] = darken(fg, 0.10)
	m["inverseSurface"] = strings.TrimPrefix(fg, "#")
	m["inverseOnSurface"] = strings.TrimPrefix(bg, "#")

	// Outline
	m["outline"] = darken(fg, 0.30)
	m["outlineVariant"] = lighten(bg, 0.15)

	// Shadows and scrim
	m["shadow"] = "000000"
	m["scrim"] = "000000"
	m["surfaceTint"] = strings.TrimPrefix(primary, "#")

	// Theme-specific colors (for Catppuccin compatibility)
	// These map to the standard terminal colors with Catppuccin names
	if colors["color15"] != "" {
		m["rosewater"] = lighten(colors["color15"], 0.05)
		m["flamingo"] = lighten(colors["color15"], 0.00)
		m["pink"] = strings.TrimPrefix(getColorOrDefault(colors, "color5", "#F5C2E7"), "#")
		m["mauve"] = strings.TrimPrefix(getColorOrDefault(colors, "color13", "#F5C2E7"), "#")
		m["red"] = strings.TrimPrefix(getColorOrDefault(colors, "color1", "#F38BA8"), "#")
		m["maroon"] = strings.TrimPrefix(getColorOrDefault(colors, "color9", "#F38BA8"), "#")
		m["peach"] = strings.TrimPrefix(getColorOrDefault(colors, "color3", "#F9E2AF"), "#")
		m["yellow"] = strings.TrimPrefix(getColorOrDefault(colors, "color11", "#F9E2AF"), "#")
		m["green"] = strings.TrimPrefix(getColorOrDefault(colors, "color2", "#A6E3A1"), "#")
		m["teal"] = strings.TrimPrefix(getColorOrDefault(colors, "color10", "#A6E3A1"), "#")
		m["sky"] = strings.TrimPrefix(getColorOrDefault(colors, "color14", "#94E2D5"), "#")
		m["sapphire"] = strings.TrimPrefix(getColorOrDefault(colors, "color6", "#94E2D5"), "#")
		m["blue"] = strings.TrimPrefix(getColorOrDefault(colors, "color4", "#89B4FA"), "#")
		m["lavender"] = strings.TrimPrefix(getColorOrDefault(colors, "color12", "#89B4FA"), "#")
	}

	return m
}

// Helper functions for color manipulation
func getColorOrDefault(colors map[string]string, key, defaultColor string) string {
	if c, ok := colors[key]; ok {
		return c
	}
	return defaultColor
}

// darken darkens a hex color by the given amount (0-1)
func darken(hexColor string, amount float64) string {
	hex := strings.TrimPrefix(hexColor, "#")
	if len(hex) != 6 {
		return hex
	}

	r := hexToInt(hex[0:2])
	g := hexToInt(hex[2:4])
	b := hexToInt(hex[4:6])

	r = int(float64(r) * (1 - amount))
	g = int(float64(g) * (1 - amount))
	b = int(float64(b) * (1 - amount))

	return fmt.Sprintf("%02x%02x%02x", r, g, b)
}

// lighten lightens a hex color by the given amount (0-1)
func lighten(hexColor string, amount float64) string {
	hex := strings.TrimPrefix(hexColor, "#")
	if len(hex) != 6 {
		return hex
	}

	r := hexToInt(hex[0:2])
	g := hexToInt(hex[2:4])
	b := hexToInt(hex[4:6])

	r = r + int(float64(255-r)*amount)
	g = g + int(float64(255-g)*amount)
	b = b + int(float64(255-b)*amount)

	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return fmt.Sprintf("%02x%02x%02x", r, g, b)
}

// contrastColor returns black or white depending on the background
func contrastColor(hexColor string) string {
	hex := strings.TrimPrefix(hexColor, "#")
	if len(hex) != 6 {
		return "ffffff"
	}

	r := hexToInt(hex[0:2])
	g := hexToInt(hex[2:4])
	b := hexToInt(hex[4:6])

	// Calculate luminance
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255

	if luminance > 0.5 {
		return "000000"
	}
	return "ffffff"
}

// hexToInt converts a hex string to an integer
func hexToInt(hex string) int {
	var val int
	fmt.Sscanf(hex, "%x", &val)
	return val
}
