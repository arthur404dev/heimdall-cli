//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Convert traditional color schemes to Material You format
func main() {
	// Read all JSON scheme files
	err := filepath.Walk("assets/schemes", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Read the JSON file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		var scheme map[string]interface{}
		if err := json.Unmarshal(data, &scheme); err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		// Get colors
		colors, ok := scheme["colors"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("no colors in %s", path)
		}

		// Convert to Material You format
		materialColors := convertToMaterialYou(colors)

		// Change extension from .json to .txt
		txtPath := strings.TrimSuffix(path, ".json") + ".txt"

		// Write as text file in Caelestia format
		var lines []string
		for key, value := range materialColors {
			lines = append(lines, fmt.Sprintf("%s %s", key, value))
		}

		content := strings.Join(lines, "\n")
		if err := os.WriteFile(txtPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", txtPath, err)
		}

		fmt.Printf("Converted %s -> %s\n", path, txtPath)
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func convertToMaterialYou(colors map[string]interface{}) map[string]string {
	m := make(map[string]string)

	// Helper to get color string
	getColor := func(key string, defaultVal string) string {
		if v, ok := colors[key].(string); ok {
			return strings.TrimPrefix(v, "#")
		}
		return defaultVal
	}

	// Get base colors
	bg := getColor("background", "1e1e2e")
	fg := getColor("foreground", "cdd6f4")

	// Map terminal colors
	for i := 0; i < 16; i++ {
		key := fmt.Sprintf("color%d", i)
		if c := getColor(key, ""); c != "" {
			m[fmt.Sprintf("term%d", i)] = c
		}
	}

	// Material Design 3 surface colors
	m["background"] = bg
	m["surface"] = bg
	m["base"] = bg
	m["mantle"] = bg
	m["crust"] = darken(bg, 0.05)

	// Surface variants
	m["surfaceDim"] = bg
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
	m["text"] = fg
	m["onBackground"] = fg
	m["onSurface"] = fg
	m["subtext1"] = darken(fg, 0.10)
	m["subtext0"] = darken(fg, 0.20)

	// Primary color (blue)
	primary := getColor("color4", "89b4fa")
	m["primary"] = primary
	m["primary_paletteKeyColor"] = darken(primary, 0.20)
	m["primaryContainer"] = darken(primary, 0.30)
	m["onPrimary"] = contrastColor(primary)
	m["onPrimaryContainer"] = "ffffff"
	m["inversePrimary"] = darken(primary, 0.40)
	m["primaryFixed"] = lighten(primary, 0.10)
	m["primaryFixedDim"] = primary
	m["onPrimaryFixed"] = darken(primary, 0.60)
	m["onPrimaryFixedVariant"] = darken(primary, 0.40)

	// Secondary color (cyan)
	secondary := getColor("color6", "94e2d5")
	m["secondary"] = secondary
	m["secondary_paletteKeyColor"] = darken(secondary, 0.20)
	m["secondaryContainer"] = darken(secondary, 0.30)
	m["onSecondary"] = contrastColor(secondary)
	m["onSecondaryContainer"] = lighten(secondary, 0.40)
	m["secondaryFixed"] = lighten(secondary, 0.10)
	m["secondaryFixedDim"] = secondary
	m["onSecondaryFixed"] = darken(secondary, 0.60)
	m["onSecondaryFixedVariant"] = darken(secondary, 0.40)

	// Tertiary color (magenta)
	tertiary := getColor("color5", "f5c2e7")
	m["tertiary"] = tertiary
	m["tertiary_paletteKeyColor"] = darken(tertiary, 0.20)
	m["tertiaryContainer"] = darken(tertiary, 0.30)
	m["onTertiary"] = contrastColor(tertiary)
	m["onTertiaryContainer"] = "000000"
	m["tertiaryFixed"] = lighten(tertiary, 0.10)
	m["tertiaryFixedDim"] = tertiary
	m["onTertiaryFixed"] = darken(tertiary, 0.60)
	m["onTertiaryFixedVariant"] = darken(tertiary, 0.40)

	// Error color (red)
	errorColor := getColor("color1", "f38ba8")
	m["error"] = errorColor
	m["onError"] = contrastColor(errorColor)
	m["errorContainer"] = darken(errorColor, 0.40)
	m["onErrorContainer"] = lighten(errorColor, 0.40)

	// Success colors (green)
	success := getColor("color2", "a6e3a1")
	m["success"] = success
	m["onSuccess"] = contrastColor(success)
	m["successContainer"] = darken(success, 0.30)
	m["onSuccessContainer"] = lighten(success, 0.40)

	// Neutral colors
	m["neutral_paletteKeyColor"] = darken(fg, 0.30)
	m["neutral_variant_paletteKeyColor"] = darken(fg, 0.35)

	// Surface variants
	m["surfaceVariant"] = lighten(bg, 0.15)
	m["onSurfaceVariant"] = darken(fg, 0.10)
	m["inverseSurface"] = fg
	m["inverseOnSurface"] = bg

	// Outline
	m["outline"] = darken(fg, 0.30)
	m["outlineVariant"] = lighten(bg, 0.15)

	// Shadows and scrim
	m["shadow"] = "000000"
	m["scrim"] = "000000"
	m["surfaceTint"] = primary

	// Theme-specific colors (Catppuccin style)
	m["rosewater"] = lighten(fg, 0.05)
	m["flamingo"] = lighten(fg, 0.00)
	m["pink"] = getColor("color5", "f5c2e7")
	m["mauve"] = getColor("color13", "f5c2e7")
	m["red"] = getColor("color1", "f38ba8")
	m["maroon"] = getColor("color9", "f38ba8")
	m["peach"] = getColor("color3", "f9e2af")
	m["yellow"] = getColor("color11", "f9e2af")
	m["green"] = getColor("color2", "a6e3a1")
	m["teal"] = getColor("color10", "a6e3a1")
	m["sky"] = getColor("color14", "94e2d5")
	m["sapphire"] = getColor("color6", "94e2d5")
	m["blue"] = getColor("color4", "89b4fa")
	m["lavender"] = getColor("color12", "89b4fa")

	return m
}

// Color manipulation functions
func hexToRGB(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}

	var r, g, b int
	fmt.Sscanf(hex[0:2], "%x", &r)
	fmt.Sscanf(hex[2:4], "%x", &g)
	fmt.Sscanf(hex[4:6], "%x", &b)
	return r, g, b
}

func rgbToHex(r, g, b int) string {
	return fmt.Sprintf("%02x%02x%02x", r, g, b)
}

func darken(hex string, amount float64) string {
	r, g, b := hexToRGB(hex)
	r = int(float64(r) * (1 - amount))
	g = int(float64(g) * (1 - amount))
	b = int(float64(b) * (1 - amount))
	return rgbToHex(r, g, b)
}

func lighten(hex string, amount float64) string {
	r, g, b := hexToRGB(hex)
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

	return rgbToHex(r, g, b)
}

func contrastColor(hex string) string {
	r, g, b := hexToRGB(hex)

	// Calculate luminance
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255

	if luminance > 0.5 {
		return "000000"
	}
	return "ffffff"
}
