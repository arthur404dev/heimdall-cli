package scheme

import (
	"encoding/json"
	"fmt"

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

			// Handle caelestia-compatible flags
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
		// Get base colors for styling
		baseColor := "#" + s.Colours["base"]
		textColor := "#" + s.Colours["text"]

		// Use ANSI color codes directly
		fmt.Printf("\033[36;1mCurrent Color Scheme\033[0m\n")
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━\n")

		fmt.Printf("\033[34mName:\033[0m    %s\n", s.Name)
		fmt.Printf("\033[34mFlavour:\033[0m %s\n", s.Flavour)
		fmt.Printf("\033[34mMode:\033[0m    %s\n", s.Mode)

		if s.Variant != "" {
			fmt.Printf("\033[34mVariant:\033[0m %s\n", s.Variant)
		}

		fmt.Printf("\n")
		fmt.Printf("\033[34mColors:\033[0m\n")

		// Display color preview
		fmt.Printf("  Background: ")
		displayColorBlock(baseColor)
		fmt.Printf(" %s\n", baseColor)

		fmt.Printf("  Foreground: ")
		displayColorBlock(textColor)
		fmt.Printf(" %s\n", textColor)

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
		fmt.Printf("\033[36;1mColor Palette\033[0m\n")
		fmt.Printf("━━━━━━━━━━━━━\n")

		for key, value := range colors {
			displayColorWithPreview(key, value)
		}
	} else {
		for key, value := range colors {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	return nil
}

// displayColorWithPreview displays a single color with preview
func displayColorWithPreview(name, hexValue string) {
	colorValue := "#" + hexValue
	fmt.Printf("%-12s ", name+":")
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
