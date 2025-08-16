package scheme

import (
	"fmt"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
	"github.com/spf13/cobra"
)

// installCommand creates the scheme install subcommand
func installCommand() *cobra.Command {
	var (
		all     bool
		userDir bool
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
				location := "data directory"
				if userDir {
					location = "user directory"
				}

				fmt.Printf("Installing all bundled schemes to %s...\n", location)

				var err error
				if userDir {
					err = manager.InstallAllBundledSchemesToUser()
				} else {
					err = manager.InstallAllBundledSchemes()
				}

				if err != nil {
					return fmt.Errorf("failed to install bundled schemes: %w", err)
				}
				fmt.Printf("Successfully installed all bundled schemes to %s\n", location)
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

			// Determine installation location
			location := "data directory"
			if userDir {
				location = "user directory"
			}

			fmt.Printf("Installing scheme: %s to %s\n", schemeName, location)

			// Install to appropriate location
			var err error
			if userDir {
				err = manager.InstallBundledSchemeToUser(schemeName)
			} else {
				err = manager.InstallBundledScheme(schemeName)
			}

			if err != nil {
				return fmt.Errorf("failed to install scheme: %w", err)
			}
			fmt.Printf("Successfully installed %s to %s\n", schemeName, location)

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Install all bundled schemes")
	cmd.Flags().BoolVar(&userDir, "user", false, "Install to user scheme directory")

	return cmd
}
