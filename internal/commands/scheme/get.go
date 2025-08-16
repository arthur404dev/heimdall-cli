package scheme

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
	"github.com/spf13/cobra"
)

// getCommand creates the scheme get subcommand
func getCommand() *cobra.Command {
	var (
		property   string
		jsonOut    bool
		getName    bool
		getFlavour bool
		getMode    bool
		getVariant bool
		noColor    bool
	)

	cmd := &cobra.Command{
		Use:   "get [property]",
		Short: "Get current scheme or specific property",
		Long: `Get the current color scheme or a specific property.
		
Properties:
  name     - Scheme name
  flavour  - Scheme flavour
  mode     - Scheme mode (dark/light)
  variant  - Scheme variant
  colors   - All colors (JSON format)
  <color>  - Specific color value (e.g., base, text, primary)
		
Examples:
  heimdall scheme get              # Show current scheme info with colors
  heimdall scheme get -n           # Get scheme name only
  heimdall scheme get -f           # Get flavour only
  heimdall scheme get -m           # Get mode only
  heimdall scheme get -v           # Get variant only
  heimdall scheme get name         # Get scheme name
  heimdall scheme get colors       # Get all colors as JSON
  heimdall scheme get base         # Get base color value
  heimdall scheme get --json       # Output full scheme as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := scheme.NewManager()

			// Get current scheme
			currentScheme, err := manager.GetCurrent()
			if err != nil {
				return err
			}

			// Handle compatibility flags
			if getName {
				fmt.Println(currentScheme.Name)
				return nil
			}

			if getFlavour {
				fmt.Println(currentScheme.Flavour)
				return nil
			}

			if getMode {
				fmt.Println(currentScheme.Mode)
				return nil
			}

			if getVariant {
				fmt.Println(currentScheme.Variant)
				return nil
			}

			// If property specified in args
			if len(args) > 0 {
				property = args[0]
			}

			// Handle JSON output for full scheme
			if jsonOut && property == "" {
				data, err := json.MarshalIndent(currentScheme, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal scheme: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			// Handle specific property
			switch property {
			case "":
				// Show basic info with colored output by default
				return displaySchemeInfo(currentScheme, !noColor)

			case "name":
				fmt.Println(currentScheme.Name)

			case "flavour":
				fmt.Println(currentScheme.Flavour)

			case "mode":
				fmt.Println(currentScheme.Mode)

			case "variant":
				fmt.Println(currentScheme.Variant)

			case "colors":
				colors := currentScheme.GetColors()
				if jsonOut {
					data, err := json.MarshalIndent(colors, "", "  ")
					if err != nil {
						return fmt.Errorf("failed to marshal colors: %w", err)
					}
					fmt.Println(string(data))
				} else {
					return displayColorsWithPreview(colors, !noColor)
				}

			default:
				// Try to get specific color
				if colorValue, ok := currentScheme.Colours[property]; ok {
					if !noColor {
						// Display color with preview
						displayColorWithPreview(property, colorValue)
					} else {
						fmt.Println(colorValue)
					}
				} else {
					return fmt.Errorf("unknown property or color: %s", property)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&getName, "name", "n", false, "Print current scheme name")
	cmd.Flags().BoolVarP(&getFlavour, "flavour", "f", false, "Print current flavour")
	cmd.Flags().BoolVarP(&getMode, "mode", "m", false, "Print current mode")
	cmd.Flags().BoolVarP(&getVariant, "variant", "v", false, "Print current variant")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	return cmd
}

// displaySchemeInfo displays scheme information with colored output
func displaySchemeInfo(s *scheme.Scheme, useColor bool) error {
	if useColor {
		// Use ANSI color codes directly
		fmt.Printf("\033[36;1mCurrent Color Scheme\033[0m\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━\n")

		fmt.Printf("\033[34mName:\033[0m    %s\n", s.Name)
		fmt.Printf("\033[34mFlavour:\033[0m %s\n", s.Flavour)
		fmt.Printf("\033[34mMode:\033[0m    %s\n", s.Mode)

		if s.Variant != "" {
			fmt.Printf("\033[34mVariant:\033[0m %s\n", s.Variant)
		}

		// Show source if available
		if s.Source != "" {
			sourceDisplay := string(s.Source)
			sourceColor := ""
			switch s.Source {
			case scheme.SourceUser:
				sourceColor = "\033[32m" // Green for user
			case scheme.SourceGenerated:
				sourceColor = "\033[33m" // Yellow for generated
			case scheme.SourceBundled:
				sourceColor = "\033[36m" // Cyan for bundled
			}
			fmt.Printf("\033[34mSource:\033[0m  %s%s\033[0m\n", sourceColor, sourceDisplay)
		}

		fmt.Printf("\n")
		fmt.Printf("\033[34mTerminal Colors:\033[0m\n")

		// Display standard 16 terminal colors
		// First check for numbered colors (color0-color15)
		hasNumberedColors := false
		for i := 0; i < 16; i++ {
			if _, ok := s.Colours[fmt.Sprintf("color%d", i)]; ok {
				hasNumberedColors = true
				break
			}
		}

		if hasNumberedColors {
			// Display numbered terminal colors in a grid
			for i := 0; i < 8; i++ {
				colorKey := fmt.Sprintf("color%d", i)
				if colorValue, ok := s.Colours[colorKey]; ok {
					fmt.Printf("  %-8s ", fmt.Sprintf("[%d]", i))
					// Ensure consistent handling of # prefix
					if !strings.HasPrefix(colorValue, "#") {
						colorValue = "#" + colorValue
					}
					displayColorBlock(colorValue)
					// Safely truncate to max 7 chars if needed
					displayValue := colorValue
					if len(colorValue) > 7 {
						displayValue = colorValue[:7]
					}
					fmt.Printf(" %-8s", displayValue)
				}

				// Display bright variant next to it
				brightKey := fmt.Sprintf("color%d", i+8)
				if colorValue, ok := s.Colours[brightKey]; ok {
					fmt.Printf("    %-8s ", fmt.Sprintf("[%d]", i+8))
					// Ensure consistent handling of # prefix
					if !strings.HasPrefix(colorValue, "#") {
						colorValue = "#" + colorValue
					}
					displayColorBlock(colorValue)
					// Safely truncate to max 7 chars if needed
					displayValue := colorValue
					if len(colorValue) > 7 {
						displayValue = colorValue[:7]
					}
					fmt.Printf(" %-8s", displayValue)
				}
				fmt.Printf("\n")
			}
		} else {
			// Fallback to named colors
			// Map common color names to terminal color indices
			namedColors := []struct {
				name  string
				keys  []string
				index int
			}{
				{"Black", []string{"black", "color0", "base", "background"}, 0},
				{"Red", []string{"red", "color1", "error"}, 1},
				{"Green", []string{"green", "color2", "success"}, 2},
				{"Yellow", []string{"yellow", "color3", "warning"}, 3},
				{"Blue", []string{"blue", "color4", "primary"}, 4},
				{"Magenta", []string{"magenta", "color5", "mauve", "pink", "tertiary"}, 5},
				{"Cyan", []string{"cyan", "color6", "teal", "sapphire", "secondary"}, 6},
				{"White", []string{"white", "color7", "text", "foreground"}, 7},
				{"Bright Black", []string{"brightBlack", "color8", "surface0", "overlay0"}, 8},
				{"Bright Red", []string{"brightRed", "color9", "maroon"}, 9},
				{"Bright Green", []string{"brightGreen", "color10"}, 10},
				{"Bright Yellow", []string{"brightYellow", "color11", "peach"}, 11},
				{"Bright Blue", []string{"brightBlue", "color12", "lavender", "sky"}, 12},
				{"Bright Magenta", []string{"brightMagenta", "color13", "flamingo", "rosewater"}, 13},
				{"Bright Cyan", []string{"brightCyan", "color14"}, 14},
				{"Bright White", []string{"brightWhite", "color15", "subtext0", "subtext1"}, 15},
			}

			for _, nc := range namedColors {
				var colorValue string
				found := false
				for _, key := range nc.keys {
					if val, ok := s.Colours[key]; ok {
						colorValue = val
						found = true
						break
					}
				}

				if found {
					fmt.Printf("  %-14s ", fmt.Sprintf("[%d] %s:", nc.index, nc.name))
					// Ensure consistent handling of # prefix
					if !strings.HasPrefix(colorValue, "#") {
						colorValue = "#" + colorValue
					}
					displayColorBlock(colorValue)
					// Safely truncate to max 7 chars if needed
					displayValue := colorValue
					if len(colorValue) > 7 {
						displayValue = colorValue[:7]
					}
					fmt.Printf(" %s\n", displayValue)
				}
			}
		}

		// Show base/text colors separately if not already shown
		fmt.Printf("\n\033[34mBase Colors:\033[0m\n")
		if baseColor, ok := s.Colours["base"]; ok {
			fmt.Printf("  Background: ")
			// Ensure consistent handling of # prefix
			if !strings.HasPrefix(baseColor, "#") {
				baseColor = "#" + baseColor
			}
			displayColorBlock(baseColor)
			fmt.Printf(" %s\n", baseColor)
		} else if bgColor, ok := s.Colours["background"]; ok {
			fmt.Printf("  Background: ")
			// Ensure consistent handling of # prefix
			if !strings.HasPrefix(bgColor, "#") {
				bgColor = "#" + bgColor
			}
			displayColorBlock(bgColor)
			fmt.Printf(" %s\n", bgColor)
		}

		if textColor, ok := s.Colours["text"]; ok {
			fmt.Printf("  Foreground: ")
			// Ensure consistent handling of # prefix
			if !strings.HasPrefix(textColor, "#") {
				textColor = "#" + textColor
			}
			displayColorBlock(textColor)
			fmt.Printf(" %s\n", textColor)
		} else if fgColor, ok := s.Colours["foreground"]; ok {
			fmt.Printf("  Foreground: ")
			// Ensure consistent handling of # prefix
			if !strings.HasPrefix(fgColor, "#") {
				fgColor = "#" + fgColor
			}
			displayColorBlock(fgColor)
			fmt.Printf(" %s\n", fgColor)
		}

	} else {
		fmt.Printf("Current scheme: %s/%s/%s\n", s.Name, s.Flavour, s.Mode)
		if s.Variant != "" {
			fmt.Printf("Variant: %s\n", s.Variant)
		}
	}

	return nil
}

// displayColorsWithPreview displays all colors with color previews
func displayColorsWithPreview(colors map[string]string, useColor bool) error {
	if useColor {
		fmt.Printf("\033[36;1mComplete Color Palette\033[0m\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━\n")

		// Categorize colors
		terminalColors := make(map[string]string)
		baseColors := make(map[string]string)
		semanticColors := make(map[string]string)
		materialColors := make(map[string]string)
		specialColors := make(map[string]string)
		otherColors := make(map[string]string)

		// Sort colors into categories
		for key, value := range colors {
			switch {
			// Terminal colors (color0-color15, term0-term15)
			case strings.HasPrefix(key, "color") || strings.HasPrefix(key, "term"):
				terminalColors[key] = value

			// Base colors
			case key == "background" || key == "foreground" || key == "base" ||
				key == "text" || key == "mantle" || key == "crust" ||
				strings.HasPrefix(key, "surface") || strings.HasPrefix(key, "overlay"):
				baseColors[key] = value

			// Semantic colors (Catppuccin style)
			case key == "red" || key == "green" || key == "yellow" || key == "blue" ||
				key == "pink" || key == "teal" || key == "cyan" || key == "magenta" ||
				key == "orange" || key == "peach" || key == "maroon" || key == "lavender" ||
				key == "mauve" || key == "sapphire" || key == "sky" || key == "flamingo" ||
				key == "rosewater" || key == "purple" || key == "white" || key == "black" ||
				strings.HasPrefix(key, "subtext"):
				semanticColors[key] = value

			// Material You colors
			case strings.Contains(key, "primary") || strings.Contains(key, "Primary") ||
				strings.Contains(key, "secondary") || strings.Contains(key, "Secondary") ||
				strings.Contains(key, "tertiary") || strings.Contains(key, "Tertiary") ||
				strings.Contains(key, "error") || strings.Contains(key, "Error") ||
				strings.Contains(key, "success") || strings.Contains(key, "Success") ||
				strings.Contains(key, "neutral") || strings.Contains(key, "Neutral") ||
				strings.Contains(key, "Container") || strings.Contains(key, "Fixed") ||
				strings.HasPrefix(key, "on"):
				materialColors[key] = value

			// Special colors
			case key == "cursor" || key == "cursor_text" || key == "selection" ||
				key == "selection_text" || key == "url":
				specialColors[key] = value

			default:
				otherColors[key] = value
			}
		}

		// Display terminal colors first if present
		if len(terminalColors) > 0 {
			fmt.Printf("\n\033[35mTerminal Colors\033[0m\n")
			fmt.Printf("───────────────\n")
			// Sort and display numbered colors
			for i := 0; i < 16; i++ {
				key := fmt.Sprintf("color%d", i)
				if value, ok := terminalColors[key]; ok {
					displayColorWithPreview(fmt.Sprintf("%-8s", key), value)
				}
				// Also check for term variant
				key = fmt.Sprintf("term%d", i)
				if value, ok := terminalColors[key]; ok {
					displayColorWithPreview(fmt.Sprintf("%-8s", key), value)
				}
			}
		}

		// Display base colors
		if len(baseColors) > 0 {
			fmt.Printf("\n\033[35mBase Colors\033[0m\n")
			fmt.Printf("───────────\n")
			// Display in specific order if available
			orderedBase := []string{"background", "foreground", "base", "text", "crust", "mantle",
				"surface", "surface0", "surface1", "surface2",
				"overlay0", "overlay1", "overlay2"}
			for _, key := range orderedBase {
				if value, ok := baseColors[key]; ok {
					displayColorWithPreview(key, value)
				}
			}
		}

		// Display semantic colors
		if len(semanticColors) > 0 {
			fmt.Printf("\n\033[35mSemantic Colors\033[0m\n")
			fmt.Printf("───────────────\n")
			// Display in rainbow order
			orderedSemantic := []string{"red", "maroon", "peach", "orange", "yellow",
				"green", "teal", "cyan", "sapphire", "sky", "blue", "lavender",
				"purple", "mauve", "pink", "flamingo", "rosewater",
				"white", "subtext0", "subtext1", "black"}
			for _, key := range orderedSemantic {
				if value, ok := semanticColors[key]; ok {
					displayColorWithPreview(key, value)
				}
			}
			// Display any remaining semantic colors
			for key, value := range semanticColors {
				found := false
				for _, ordered := range orderedSemantic {
					if key == ordered {
						found = true
						break
					}
				}
				if !found {
					displayColorWithPreview(key, value)
				}
			}
		}

		// Display Material You colors
		if len(materialColors) > 0 {
			fmt.Printf("\n\033[35mMaterial You Colors\033[0m\n")
			fmt.Printf("───────────────────\n")
			// Group by type
			primaryKeys := []string{}
			secondaryKeys := []string{}
			tertiaryKeys := []string{}
			errorKeys := []string{}
			successKeys := []string{}
			neutralKeys := []string{}

			for key := range materialColors {
				switch {
				case strings.Contains(strings.ToLower(key), "primary"):
					primaryKeys = append(primaryKeys, key)
				case strings.Contains(strings.ToLower(key), "secondary"):
					secondaryKeys = append(secondaryKeys, key)
				case strings.Contains(strings.ToLower(key), "tertiary"):
					tertiaryKeys = append(tertiaryKeys, key)
				case strings.Contains(strings.ToLower(key), "error"):
					errorKeys = append(errorKeys, key)
				case strings.Contains(strings.ToLower(key), "success"):
					successKeys = append(successKeys, key)
				case strings.Contains(strings.ToLower(key), "neutral"):
					neutralKeys = append(neutralKeys, key)
				}
			}

			// Display grouped
			for _, key := range primaryKeys {
				displayColorWithPreview(key, materialColors[key])
			}
			for _, key := range secondaryKeys {
				displayColorWithPreview(key, materialColors[key])
			}
			for _, key := range tertiaryKeys {
				displayColorWithPreview(key, materialColors[key])
			}
			for _, key := range errorKeys {
				displayColorWithPreview(key, materialColors[key])
			}
			for _, key := range successKeys {
				displayColorWithPreview(key, materialColors[key])
			}
			for _, key := range neutralKeys {
				displayColorWithPreview(key, materialColors[key])
			}
		}

		// Display special colors
		if len(specialColors) > 0 {
			fmt.Printf("\n\033[35mSpecial Colors\033[0m\n")
			fmt.Printf("──────────────\n")
			for key, value := range specialColors {
				displayColorWithPreview(key, value)
			}
		}

		// Display any other colors
		if len(otherColors) > 0 {
			fmt.Printf("\n\033[35mOther Colors\033[0m\n")
			fmt.Printf("────────────\n")
			for key, value := range otherColors {
				displayColorWithPreview(key, value)
			}
		}

	} else {
		// Non-colored output - still organize by category
		fmt.Printf("# Terminal Colors\n")
		for i := 0; i < 16; i++ {
			key := fmt.Sprintf("color%d", i)
			if value, ok := colors[key]; ok {
				fmt.Printf("%s: %s\n", key, value)
			}
		}

		fmt.Printf("\n# Base Colors\n")
		baseKeys := []string{"background", "foreground", "base", "text", "crust", "mantle"}
		for _, key := range baseKeys {
			if value, ok := colors[key]; ok {
				fmt.Printf("%s: %s\n", key, value)
			}
		}

		fmt.Printf("\n# All Colors\n")
		for key, value := range colors {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	return nil
}

// displayColorWithPreview displays a single color with preview
func displayColorWithPreview(name, hexValue string) {
	// Ensure we have a properly formatted hex color
	// Handle both with and without # prefix consistently
	colorValue := hexValue
	if !strings.HasPrefix(colorValue, "#") {
		colorValue = "#" + colorValue
	}
	fmt.Printf("  %-20s ", name+":")
	displayColorBlock(colorValue)
	fmt.Printf(" %s\n", colorValue)
}

// displayColorBlock displays a colored block for the given hex color
func displayColorBlock(hexColor string) {
	// Create ANSI color code from hex
	if len(hexColor) == 7 && hexColor[0] == '#' {
		// Convert hex to RGB
		r, g, b := hexToRGB(hexColor[1:])

		// Create 24-bit color escape sequence
		fmt.Printf("\033[48;2;%d;%d;%dm  \033[0m", r, g, b)
	} else {
		fmt.Printf("██")
	}
}

// hexToRGB converts hex color to RGB values
func hexToRGB(hex string) (int, int, int) {
	if len(hex) != 6 {
		return 0, 0, 0
	}

	var r, g, b int
	fmt.Sscanf(hex[0:2], "%x", &r)
	fmt.Sscanf(hex[2:4], "%x", &g)
	fmt.Sscanf(hex[4:6], "%x", &b)

	return r, g, b
}
