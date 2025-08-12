package scheme

import (
	"fmt"
	"sort"
	"strings"

	"github.com/heimdall-cli/heimdall/internal/scheme"
	"github.com/spf13/cobra"
)

// bundledCommand creates the scheme bundled subcommand
func bundledCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundled",
		Short: "List bundled color schemes",
		Long: `List all bundled color schemes with their details.
		
These schemes are embedded in the binary and can be used directly
without installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			schemes, err := scheme.GetBundledSchemes()
			if err != nil {
				return fmt.Errorf("failed to get bundled schemes: %w", err)
			}

			if len(schemes) == 0 {
				fmt.Println("No bundled schemes available")
				return nil
			}

			// Group schemes by family
			families := make(map[string][]scheme.BundledScheme)
			for _, s := range schemes {
				families[s.Family] = append(families[s.Family], s)
			}

			// Sort families
			familyNames := make([]string, 0, len(families))
			for name := range families {
				familyNames = append(familyNames, name)
			}
			sort.Strings(familyNames)

			fmt.Println("Bundled Color Schemes:")
			fmt.Println()

			for _, family := range familyNames {
				// Capitalize family name
				displayName := strings.ToUpper(family[:1]) + family[1:]
				fmt.Printf("## %s\n", displayName)

				// Sort schemes within family by flavour
				familySchemes := families[family]
				sort.Slice(familySchemes, func(i, j int) bool {
					return familySchemes[i].Flavour < familySchemes[j].Flavour
				})

				for _, s := range familySchemes {
					fmt.Printf("  - %s (%s)\n", s.Name, s.Variant)
					if s.Author != "" {
						fmt.Printf("    Author: %s\n", s.Author)
					}
				}
				fmt.Println()
			}

			fmt.Println("Usage:")
			fmt.Println("  To use a bundled scheme directly:")
			fmt.Println("    heimdall scheme set <family> <flavour> <mode>")
			fmt.Println()
			fmt.Println("  To install a bundled scheme locally:")
			fmt.Println("    heimdall scheme install \"<scheme-name>\"")
			fmt.Println()
			fmt.Println("  To install all bundled schemes:")
			fmt.Println("    heimdall scheme install --all")

			return nil
		},
	}

	return cmd
}
