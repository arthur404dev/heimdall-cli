package scheme

import (
	"fmt"
	"sort"

	"github.com/heimdall-cli/heimdall/internal/scheme"
	"github.com/spf13/cobra"
)

// listCommand creates the scheme list subcommand
func listCommand() *cobra.Command {
	var (
		schemeName string
		flavour    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available schemes, flavours, or modes",
		Long: `List available color schemes, flavours, or modes.
		
Examples:
  heimdall scheme list                    # List all schemes
  heimdall scheme list -s rosepine        # List flavours for rosepine
  heimdall scheme list -s rosepine -f main # List modes for rosepine/main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := scheme.NewManager()

			// List modes for specific scheme/flavour
			if schemeName != "" && flavour != "" {
				modes, err := manager.ListModes(schemeName, flavour)
				if err != nil {
					return err
				}

				sort.Strings(modes)
				fmt.Printf("Available modes for %s/%s:\n", schemeName, flavour)
				for _, mode := range modes {
					fmt.Printf("  - %s\n", mode)
				}
				return nil
			}

			// List flavours for specific scheme
			if schemeName != "" {
				flavours, err := manager.ListFlavours(schemeName)
				if err != nil {
					return err
				}

				sort.Strings(flavours)
				fmt.Printf("Available flavours for %s:\n", schemeName)
				for _, f := range flavours {
					fmt.Printf("  - %s\n", f)
				}
				return nil
			}

			// List all schemes
			schemes, err := manager.ListSchemes()
			if err != nil {
				return err
			}

			if len(schemes) == 0 {
				fmt.Println("No schemes available")
				return nil
			}

			sort.Strings(schemes)
			fmt.Println("Available schemes:")
			for _, s := range schemes {
				fmt.Printf("  - %s\n", s)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&schemeName, "scheme", "s", "", "Scheme name")
	cmd.Flags().StringVarP(&flavour, "flavour", "f", "", "Flavour name")

	return cmd
}
