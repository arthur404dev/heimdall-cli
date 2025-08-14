package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Scheme struct {
	Name    string            `json:"name"`
	Flavour string            `json:"flavour"`
	Mode    string            `json:"mode"`
	Colours map[string]string `json:"colours"`
}

// Define the order for organized output
var colorOrder = []string{
	// Core colors
	"background",
	"foreground",
	"text",

	// Base colors
	"base",
	"mantle",
	"crust",

	// Primary colors
	"primary",
	"onPrimary",
	"primaryContainer",
	"onPrimaryContainer",
	"primaryFixed",
	"primaryFixedDim",
	"onPrimaryFixed",
	"onPrimaryFixedVariant",
	"primary_paletteKeyColor",

	// Secondary colors
	"secondary",
	"onSecondary",
	"secondaryContainer",
	"onSecondaryContainer",
	"secondaryFixed",
	"secondaryFixedDim",
	"onSecondaryFixed",
	"onSecondaryFixedVariant",
	"secondary_paletteKeyColor",

	// Tertiary colors
	"tertiary",
	"onTertiary",
	"tertiaryContainer",
	"onTertiaryContainer",
	"tertiaryFixed",
	"tertiaryFixedDim",
	"onTertiaryFixed",
	"onTertiaryFixedVariant",
	"tertiary_paletteKeyColor",

	// Error colors
	"error",
	"onError",
	"errorContainer",
	"onErrorContainer",

	// Success colors
	"success",
	"onSuccess",
	"successContainer",
	"onSuccessContainer",

	// Surface colors
	"surface",
	"onSurface",
	"surfaceDim",
	"surfaceBright",
	"surfaceVariant",
	"onSurfaceVariant",
	"surfaceContainerLowest",
	"surfaceContainerLow",
	"surfaceContainer",
	"surfaceContainerHigh",
	"surfaceContainerHighest",
	"surface0",
	"surface1",
	"surface2",
	"surfaceTint",

	// Inverse colors
	"inverseSurface",
	"inverseOnSurface",
	"inversePrimary",

	// Background
	"onBackground",

	// Outline
	"outline",
	"outlineVariant",

	// Other Material You
	"shadow",
	"scrim",

	// Palette key colors
	"neutral_paletteKeyColor",
	"neutral_variant_paletteKeyColor",

	// Overlay colors
	"overlay0",
	"overlay1",
	"overlay2",

	// Text variants
	"subtext0",
	"subtext1",

	// Catppuccin colors
	"rosewater",
	"flamingo",
	"pink",
	"mauve",
	"red",
	"maroon",
	"peach",
	"yellow",
	"green",
	"teal",
	"sky",
	"sapphire",
	"blue",
	"lavender",

	// Terminal colors term0-15
	"term0",
	"term1",
	"term2",
	"term3",
	"term4",
	"term5",
	"term6",
	"term7",
	"term8",
	"term9",
	"term10",
	"term11",
	"term12",
	"term13",
	"term14",
	"term15",

	// Terminal colors color0-15
	"color0",
	"color1",
	"color2",
	"color3",
	"color4",
	"color5",
	"color6",
	"color7",
	"color8",
	"color9",
	"color10",
	"color11",
	"color12",
	"color13",
	"color14",
	"color15",
}

// Get the complete Material You template from gruvbox
func getCompleteTemplate() map[string]bool {
	template := make(map[string]bool)
	for _, key := range colorOrder {
		template[key] = true
	}
	return template
}

// Generate missing Material You colors based on existing colors
func generateMissingColors(colors map[string]string) map[string]string {
	complete := make(map[string]string)

	// Copy existing colors
	for k, v := range colors {
		complete[k] = v
	}

	// Ensure we have core colors
	if _, ok := complete["background"]; !ok {
		if v, ok := complete["base"]; ok {
			complete["background"] = v
		} else {
			complete["background"] = "#1e1e2e" // Default dark background
		}
	}

	if _, ok := complete["foreground"]; !ok {
		if v, ok := complete["text"]; ok {
			complete["foreground"] = v
		} else {
			complete["foreground"] = "#cdd6f4" // Default light text
		}
	}

	if _, ok := complete["text"]; !ok {
		complete["text"] = complete["foreground"]
	}

	// Generate base colors if missing
	if _, ok := complete["base"]; !ok {
		complete["base"] = complete["background"]
	}

	if _, ok := complete["mantle"]; !ok {
		// Slightly darker than base
		complete["mantle"] = darkenColor(complete["base"], 0.05)
	}

	if _, ok := complete["crust"]; !ok {
		// Even darker
		complete["crust"] = darkenColor(complete["base"], 0.1)
	}

	// Generate primary colors if missing
	if _, ok := complete["primary"]; !ok {
		if v, ok := complete["blue"]; ok {
			complete["primary"] = v
		} else if v, ok := complete["term4"]; ok {
			complete["primary"] = v
		} else {
			complete["primary"] = "#89b4fa"
		}
	}

	// Generate surface colors
	bg := complete["background"]
	if _, ok := complete["surface"]; !ok {
		complete["surface"] = bg
	}
	if _, ok := complete["surfaceDim"]; !ok {
		complete["surfaceDim"] = bg
	}
	if _, ok := complete["surfaceBright"]; !ok {
		complete["surfaceBright"] = lightenColor(bg, 0.15)
	}
	if _, ok := complete["surfaceContainerLowest"]; !ok {
		complete["surfaceContainerLowest"] = darkenColor(bg, 0.05)
	}
	if _, ok := complete["surfaceContainerLow"]; !ok {
		complete["surfaceContainerLow"] = lightenColor(bg, 0.03)
	}
	if _, ok := complete["surfaceContainer"]; !ok {
		complete["surfaceContainer"] = lightenColor(bg, 0.06)
	}
	if _, ok := complete["surfaceContainerHigh"]; !ok {
		complete["surfaceContainerHigh"] = lightenColor(bg, 0.09)
	}
	if _, ok := complete["surfaceContainerHighest"]; !ok {
		complete["surfaceContainerHighest"] = lightenColor(bg, 0.12)
	}

	// Surface 0,1,2
	if _, ok := complete["surface0"]; !ok {
		complete["surface0"] = lightenColor(bg, 0.08)
	}
	if _, ok := complete["surface1"]; !ok {
		complete["surface1"] = lightenColor(bg, 0.10)
	}
	if _, ok := complete["surface2"]; !ok {
		complete["surface2"] = lightenColor(bg, 0.12)
	}

	// Generate on-colors
	fg := complete["foreground"]
	if _, ok := complete["onBackground"]; !ok {
		complete["onBackground"] = fg
	}
	if _, ok := complete["onSurface"]; !ok {
		complete["onSurface"] = fg
	}
	if _, ok := complete["onPrimary"]; !ok {
		complete["onPrimary"] = darkenColor(complete["primary"], 0.7)
	}

	// Generate container colors
	if _, ok := complete["primaryContainer"]; !ok {
		complete["primaryContainer"] = darkenColor(complete["primary"], 0.3)
	}
	if _, ok := complete["onPrimaryContainer"]; !ok {
		complete["onPrimaryContainer"] = darkenColor(complete["primary"], 0.8)
	}

	// Secondary colors
	if _, ok := complete["secondary"]; !ok {
		if v, ok := complete["teal"]; ok {
			complete["secondary"] = v
		} else if v, ok := complete["term6"]; ok {
			complete["secondary"] = v
		} else {
			complete["secondary"] = "#94e2d5"
		}
	}

	// Tertiary colors
	if _, ok := complete["tertiary"]; !ok {
		if v, ok := complete["pink"]; ok {
			complete["tertiary"] = v
		} else if v, ok := complete["term5"]; ok {
			complete["tertiary"] = v
		} else {
			complete["tertiary"] = "#f5c2e7"
		}
	}

	// Error colors
	if _, ok := complete["error"]; !ok {
		if v, ok := complete["red"]; ok {
			complete["error"] = v
		} else if v, ok := complete["term1"]; ok {
			complete["error"] = v
		} else {
			complete["error"] = "#f38ba8"
		}
	}

	// Generate missing Material You properties
	generateMaterialYouColors(complete)

	// Ensure terminal colors exist
	ensureTerminalColors(complete)

	// Add catppuccin color names if missing
	ensureCatppuccinColors(complete)

	return complete
}

func generateMaterialYouColors(colors map[string]string) {
	// Fixed colors
	if _, ok := colors["primaryFixed"]; !ok {
		colors["primaryFixed"] = lightenColor(colors["primary"], 0.2)
	}
	if _, ok := colors["primaryFixedDim"]; !ok {
		colors["primaryFixedDim"] = colors["primary"]
	}
	if _, ok := colors["onPrimaryFixed"]; !ok {
		colors["onPrimaryFixed"] = darkenColor(colors["primary"], 0.6)
	}
	if _, ok := colors["onPrimaryFixedVariant"]; !ok {
		colors["onPrimaryFixedVariant"] = darkenColor(colors["primary"], 0.4)
	}

	// Do the same for secondary and tertiary
	if _, ok := colors["secondaryContainer"]; !ok {
		colors["secondaryContainer"] = darkenColor(colors["secondary"], 0.3)
	}
	if _, ok := colors["onSecondary"]; !ok {
		colors["onSecondary"] = darkenColor(colors["secondary"], 0.7)
	}
	if _, ok := colors["onSecondaryContainer"]; !ok {
		colors["onSecondaryContainer"] = lightenColor(colors["secondary"], 0.3)
	}

	// Error containers
	if _, ok := colors["errorContainer"]; !ok {
		colors["errorContainer"] = darkenColor(colors["error"], 0.5)
	}
	if _, ok := colors["onError"]; !ok {
		colors["onError"] = darkenColor(colors["error"], 0.8)
	}
	if _, ok := colors["onErrorContainer"]; !ok {
		colors["onErrorContainer"] = lightenColor(colors["error"], 0.3)
	}

	// Success colors (green-based)
	if _, ok := colors["success"]; !ok {
		if v, ok := colors["green"]; ok {
			colors["success"] = v
		} else if v, ok := colors["term2"]; ok {
			colors["success"] = v
		} else {
			colors["success"] = "#a6e3a1"
		}
	}
	if _, ok := colors["successContainer"]; !ok {
		colors["successContainer"] = darkenColor(colors["success"], 0.4)
	}
	if _, ok := colors["onSuccess"]; !ok {
		colors["onSuccess"] = darkenColor(colors["success"], 0.7)
	}
	if _, ok := colors["onSuccessContainer"]; !ok {
		colors["onSuccessContainer"] = lightenColor(colors["success"], 0.3)
	}

	// Surface variants
	if _, ok := colors["surfaceVariant"]; !ok {
		colors["surfaceVariant"] = lightenColor(colors["surface"], 0.1)
	}
	if _, ok := colors["onSurfaceVariant"]; !ok {
		colors["onSurfaceVariant"] = darkenColor(colors["onSurface"], 0.1)
	}
	if _, ok := colors["surfaceTint"]; !ok {
		colors["surfaceTint"] = colors["primary"]
	}

	// Inverse colors
	if _, ok := colors["inverseSurface"]; !ok {
		colors["inverseSurface"] = colors["onSurface"]
	}
	if _, ok := colors["inverseOnSurface"]; !ok {
		colors["inverseOnSurface"] = colors["surface"]
	}
	if _, ok := colors["inversePrimary"]; !ok {
		colors["inversePrimary"] = darkenColor(colors["primary"], 0.5)
	}

	// Outline colors
	if _, ok := colors["outline"]; !ok {
		colors["outline"] = lightenColor(colors["surface"], 0.3)
	}
	if _, ok := colors["outlineVariant"]; !ok {
		colors["outlineVariant"] = lightenColor(colors["surface"], 0.15)
	}

	// Other
	if _, ok := colors["shadow"]; !ok {
		colors["shadow"] = "#000000"
	}
	if _, ok := colors["scrim"]; !ok {
		colors["scrim"] = "#000000"
	}

	// Palette key colors
	if _, ok := colors["primary_paletteKeyColor"]; !ok {
		colors["primary_paletteKeyColor"] = colors["primaryContainer"]
	}
	if _, ok := colors["secondary_paletteKeyColor"]; !ok {
		colors["secondary_paletteKeyColor"] = colors["secondaryContainer"]
	}
	if _, ok := colors["tertiary_paletteKeyColor"]; !ok {
		colors["tertiary_paletteKeyColor"] = colors["tertiaryContainer"]
	}
	if _, ok := colors["neutral_paletteKeyColor"]; !ok {
		colors["neutral_paletteKeyColor"] = colors["surface2"]
	}
	if _, ok := colors["neutral_variant_paletteKeyColor"]; !ok {
		colors["neutral_variant_paletteKeyColor"] = colors["surfaceVariant"]
	}

	// Complete remaining fixed colors
	completeFixedColors(colors, "secondary")
	completeFixedColors(colors, "tertiary")
}

func completeFixedColors(colors map[string]string, prefix string) {
	base := colors[prefix]
	if _, ok := colors[prefix+"Fixed"]; !ok {
		colors[prefix+"Fixed"] = lightenColor(base, 0.2)
	}
	if _, ok := colors[prefix+"FixedDim"]; !ok {
		colors[prefix+"FixedDim"] = base
	}
	if _, ok := colors["on"+capitalize(prefix)+"Fixed"]; !ok {
		colors["on"+capitalize(prefix)+"Fixed"] = darkenColor(base, 0.6)
	}
	if _, ok := colors["on"+capitalize(prefix)+"FixedVariant"]; !ok {
		colors["on"+capitalize(prefix)+"FixedVariant"] = darkenColor(base, 0.4)
	}
	if _, ok := colors[prefix+"Container"]; !ok {
		colors[prefix+"Container"] = darkenColor(base, 0.3)
	}
	if _, ok := colors["on"+capitalize(prefix)+"Container"]; !ok {
		colors["on"+capitalize(prefix)+"Container"] = lightenColor(base, 0.3)
	}
}

func ensureTerminalColors(colors map[string]string) {
	// Ensure both term and color formats exist
	for i := 0; i < 16; i++ {
		termKey := fmt.Sprintf("term%d", i)
		colorKey := fmt.Sprintf("color%d", i)

		// If we have one but not the other, copy it
		if val, ok := colors[termKey]; ok {
			if _, ok2 := colors[colorKey]; !ok2 {
				colors[colorKey] = val
			}
		} else if val, ok := colors[colorKey]; ok {
			colors[termKey] = val
		} else {
			// Generate default terminal colors if missing
			switch i {
			case 0: // black
				colors[termKey] = colors["surface1"]
				colors[colorKey] = colors["surface1"]
			case 1: // red
				colors[termKey] = colors["red"]
				colors[colorKey] = colors["red"]
			case 2: // green
				colors[termKey] = colors["green"]
				colors[colorKey] = colors["green"]
			case 3: // yellow
				colors[termKey] = colors["yellow"]
				colors[colorKey] = colors["yellow"]
			case 4: // blue
				colors[termKey] = colors["blue"]
				colors[colorKey] = colors["blue"]
			case 5: // magenta
				colors[termKey] = colors["pink"]
				colors[colorKey] = colors["pink"]
			case 6: // cyan
				colors[termKey] = colors["teal"]
				colors[colorKey] = colors["teal"]
			case 7: // white
				colors[termKey] = colors["subtext1"]
				colors[colorKey] = colors["subtext1"]
			case 8: // bright black
				colors[termKey] = colors["surface2"]
				colors[colorKey] = colors["surface2"]
			default:
				// Bright colors (9-15) same as normal
				colors[termKey] = colors[fmt.Sprintf("term%d", i-8)]
				colors[colorKey] = colors[fmt.Sprintf("color%d", i-8)]
			}
		}
	}
}

func ensureCatppuccinColors(colors map[string]string) {
	// Ensure catppuccin color names exist (they might map to other colors)
	catppuccinDefaults := map[string]string{
		"rosewater": colors["peach"],
		"flamingo":  colors["pink"],
		"pink":      colors["tertiary"],
		"mauve":     lightenColor(colors["tertiary"], 0.1),
		"red":       colors["error"],
		"maroon":    darkenColor(colors["error"], 0.1),
		"peach":     lightenColor(colors["yellow"], 0.2),
		"yellow":    colors["term3"],
		"green":     colors["term2"],
		"teal":      colors["secondary"],
		"sky":       lightenColor(colors["secondary"], 0.1),
		"sapphire":  darkenColor(colors["primary"], 0.1),
		"blue":      colors["primary"],
		"lavender":  lightenColor(colors["primary"], 0.1),
	}

	for name, defaultVal := range catppuccinDefaults {
		if _, ok := colors[name]; !ok && defaultVal != "" {
			colors[name] = defaultVal
		}
	}

	// Overlay colors
	if _, ok := colors["overlay0"]; !ok {
		colors["overlay0"] = lightenColor(colors["surface"], 0.2)
	}
	if _, ok := colors["overlay1"]; !ok {
		colors["overlay1"] = lightenColor(colors["surface"], 0.25)
	}
	if _, ok := colors["overlay2"]; !ok {
		colors["overlay2"] = lightenColor(colors["surface"], 0.3)
	}

	// Subtext colors
	if _, ok := colors["subtext0"]; !ok {
		colors["subtext0"] = darkenColor(colors["text"], 0.15)
	}
	if _, ok := colors["subtext1"]; !ok {
		colors["subtext1"] = darkenColor(colors["text"], 0.08)
	}
}

// Simple color manipulation functions
func lightenColor(hex string, factor float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "#" + hex
	}

	r, _ := parseHex(hex[0:2])
	g, _ := parseHex(hex[2:4])
	b, _ := parseHex(hex[4:6])

	r = r + int(float64(255-r)*factor)
	g = g + int(float64(255-g)*factor)
	b = b + int(float64(255-b)*factor)

	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func darkenColor(hex string, factor float64) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return "#" + hex
	}

	r, _ := parseHex(hex[0:2])
	g, _ := parseHex(hex[2:4])
	b, _ := parseHex(hex[4:6])

	r = int(float64(r) * (1 - factor))
	g = int(float64(g) * (1 - factor))
	b = int(float64(b) * (1 - factor))

	if r < 0 {
		r = 0
	}
	if g < 0 {
		g = 0
	}
	if b < 0 {
		b = 0
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func parseHex(s string) (int, error) {
	var val int
	_, err := fmt.Sscanf(s, "%x", &val)
	return val, err
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func standardizeScheme(schemePath string) error {
	// Read existing JSON
	data, err := os.ReadFile(schemePath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", schemePath, err)
	}

	var scheme Scheme
	if err := json.Unmarshal(data, &scheme); err != nil {
		return fmt.Errorf("failed to parse %s: %w", schemePath, err)
	}

	// Generate all missing colors
	completeColors := generateMissingColors(scheme.Colours)

	// Create ordered map for output
	orderedColors := make(map[string]string)

	// Add colors in the defined order
	for _, key := range colorOrder {
		if val, ok := completeColors[key]; ok {
			orderedColors[key] = val
		}
	}

	// Add any remaining colors not in our order (shouldn't happen but just in case)
	for key, val := range completeColors {
		if _, exists := orderedColors[key]; !exists {
			orderedColors[key] = val
		}
	}

	// Update scheme
	scheme.Colours = orderedColors

	// Custom marshal to maintain order
	output := fmt.Sprintf(`{
  "name": "%s",
  "flavour": "%s",
  "mode": "%s",
  "colours": {`, scheme.Name, scheme.Flavour, scheme.Mode)

	first := true
	for _, key := range colorOrder {
		if val, ok := orderedColors[key]; ok {
			if !first {
				output += ","
			}
			output += fmt.Sprintf("\n    \"%s\": \"%s\"", key, val)
			first = false
		}
	}

	// Add any extra colors
	var extraKeys []string
	for key := range orderedColors {
		found := false
		for _, orderedKey := range colorOrder {
			if key == orderedKey {
				found = true
				break
			}
		}
		if !found {
			extraKeys = append(extraKeys, key)
		}
	}

	sort.Strings(extraKeys)
	for _, key := range extraKeys {
		output += fmt.Sprintf(",\n    \"%s\": \"%s\"", key, orderedColors[key])
	}

	output += "\n  }\n}"

	// Write the file
	if err := os.WriteFile(schemePath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", schemePath, err)
	}

	fmt.Printf("Completed: %s\n", schemePath)
	return nil
}

func main() {
	schemesDir := "assets/schemes"

	err := filepath.WalkDir(schemesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			if err := standardizeScheme(path); err != nil {
				fmt.Printf("Error standardizing %s: %v\n", path, err)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAll themes now have complete Material You format!")
}
