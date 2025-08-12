package scheme

import (
	"fmt"
	"strings"

	"github.com/heimdall-cli/heimdall/internal/scheme"
	"github.com/spf13/cobra"
)

// installCommand creates the scheme install subcommand
func installCommand() *cobra.Command {
	var (
		all bool
	)

	cmd := &cobra.Command{
		Use:   "install [scheme-name]",
		Short: "Install bundled color schemes",
		Long: `Install bundled color schemes to your local scheme directory.
		
Available bundled schemes:
  - Catppuccin (Mocha, Macchiato, Frappe, Latte)
  - Gruvbox (Dark, Light)
  - Rosé Pine (Main, Moon, Dawn)
  - OneDark
  - Dracula
  - Tokyo Night (Night, Storm, Day)

Examples:
  heimdall scheme install "Catppuccin Mocha"  # Install specific scheme
  heimdall scheme install --all                # Install all bundled schemes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := scheme.NewManager()

			// Install all bundled schemes
			if all {
				fmt.Println("Installing all bundled schemes...")
				if err := manager.InstallAllBundledSchemes(); err != nil {
					return fmt.Errorf("failed to install bundled schemes: %w", err)
				}
				fmt.Println("Successfully installed all bundled schemes")
				return nil
			}

			// Install specific scheme
			if len(args) == 0 {
				// List available bundled schemes
				names, err := scheme.ListBundledSchemeNames()
				if err != nil {
					return fmt.Errorf("failed to list bundled schemes: %w", err)
				}

				if len(names) == 0 {
					fmt.Println("No bundled schemes available")
					return nil
				}

				fmt.Println("Available bundled schemes:")
				for _, name := range names {
					fmt.Printf("  - %s\n", name)
				}
				fmt.Println("\nTo install a scheme, run:")
				fmt.Println("  heimdall scheme install \"<scheme-name>\"")
				fmt.Println("\nTo install all schemes, run:")
				fmt.Println("  heimdall scheme install --all")
				return nil
			}

			schemeName := strings.Join(args, " ")
			fmt.Printf("Installing scheme: %s\n", schemeName)
			if err := manager.InstallBundledScheme(schemeName); err != nil {
				return fmt.Errorf("failed to install scheme: %w", err)
			}
			fmt.Printf("Successfully installed %s\n", schemeName)

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Install all bundled schemes")

	return cmd
}
