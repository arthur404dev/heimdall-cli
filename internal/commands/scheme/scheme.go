package scheme

import (
	"github.com/spf13/cobra"
)

// Command creates the scheme command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheme",
		Short: "Manage color schemes",
		Long: `Manage color schemes for theming.
		
Available subcommands:
  list    - List available schemes, flavours, or modes
  get     - Get current scheme or specific property
  set     - Set the active scheme
  install - Install bundled color schemes
  bundled - Show bundled schemes with details`,
	}

	// Add subcommands
	cmd.AddCommand(listCommand())
	cmd.AddCommand(getCommand())
	cmd.AddCommand(setCommand())
	cmd.AddCommand(installCommand())
	cmd.AddCommand(bundledCommand())

	return cmd
}
