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
		property string
		jsonOut  bool
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
  heimdall scheme get              # Show current scheme info
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
				// Show basic info
				fmt.Printf("Current scheme: %s/%s/%s\n",
					currentScheme.Name, currentScheme.Flavour, currentScheme.Mode)
				if currentScheme.Variant != "" {
					fmt.Printf("Variant: %s\n", currentScheme.Variant)
				}
				if currentScheme.Metadata.Author != "" {
					fmt.Printf("Author: %s\n", currentScheme.Metadata.Author)
				}
				if currentScheme.Metadata.Description != "" {
					fmt.Printf("Description: %s\n", currentScheme.Metadata.Description)
				}

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
					for key, value := range colors {
						fmt.Printf("%s: %s\n", key, value)
					}
				}

			default:
				// Try to get specific color
				if color, ok := currentScheme.Colors[property]; ok {
					fmt.Println(color.Hex)
				} else {
					return fmt.Errorf("unknown property or color: %s", property)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output as JSON")

	return cmd
}
