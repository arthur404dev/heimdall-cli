package scheme

import (
	"fmt"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/scheme"
	"github.com/arthur404dev/heimdall-cli/internal/theme"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/spf13/cobra"
)

// setCommand creates the scheme set subcommand
func setCommand() *cobra.Command {
	var (
		noApply bool
		variant string
	)

	cmd := &cobra.Command{
		Use:   "set <scheme> [flavour] [mode]",
		Short: "Set the active color scheme",
		Long: `Set the active color scheme and apply theme.
		
Arguments:
  scheme  - Scheme name (required)
  flavour - Flavour name (optional, defaults to first available)
  mode    - Mode: dark or light (optional, defaults to dark)
		
Examples:
  heimdall scheme set rosepine            # Use rosepine with defaults
  heimdall scheme set rosepine main       # Use rosepine/main with default mode
  heimdall scheme set rosepine main dark  # Use rosepine/main/dark
  heimdall scheme set catppuccin mocha dark --variant blue`,
		Args: cobra.RangeArgs(1, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := scheme.NewManager()

			schemeName := args[0]
			flavour := ""
			mode := "dark"

			// Parse arguments
			if len(args) > 1 {
				flavour = args[1]
			}
			if len(args) > 2 {
				mode = args[2]
			}

			// If flavour not specified, get first available
			if flavour == "" {
				flavours, err := manager.ListFlavours(schemeName)
				if err != nil {
					return fmt.Errorf("failed to list flavours: %w", err)
				}
				if len(flavours) == 0 {
					return fmt.Errorf("no flavours available for scheme %s", schemeName)
				}
				flavour = flavours[0]
				logger.Info("Using default flavour", "flavour", flavour)
			}

			// Validate mode
			if mode != "dark" && mode != "light" {
				return fmt.Errorf("invalid mode: %s (must be 'dark' or 'light')", mode)
			}

			// Load the scheme (with fallback to bundled)
			newScheme, err := manager.LoadSchemeWithFallback(schemeName, flavour, mode)
			if err != nil {
				return fmt.Errorf("failed to load scheme: %w", err)
			}

			// Set variant if specified
			if variant != "" {
				newScheme.Variant = variant
			}

			// Save as current scheme
			if err := manager.SetScheme(newScheme); err != nil {
				return fmt.Errorf("failed to set scheme: %w", err)
			}

			logger.Info("Scheme set",
				"scheme", schemeName,
				"flavour", flavour,
				"mode", mode,
				"variant", variant)

			// Apply theme unless disabled
			if !noApply {
				if err := applyTheme(newScheme); err != nil {
					logger.Error("Failed to apply theme", "error", err)
					return fmt.Errorf("failed to apply theme: %w", err)
				}

				// Send notification
				notifier := notify.NewNotifier()
				notifier.Send(&notify.Notification{
					Summary: "Scheme Changed",
					Body:    fmt.Sprintf("Applied %s/%s/%s", schemeName, flavour, mode),
					Urgency: notify.UrgencyNormal,
				})
			}

			fmt.Printf("Scheme set to %s/%s/%s\n", schemeName, flavour, mode)
			if variant != "" {
				fmt.Printf("Variant: %s\n", variant)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noApply, "no-apply", false, "Don't apply theme after setting scheme")
	cmd.Flags().StringVar(&variant, "variant", "", "Scheme variant (e.g., blue, green)")

	return cmd
}

// applyTheme applies the theme for the current scheme
func applyTheme(s *scheme.Scheme) error {
	// Load configuration
	if err := config.Load(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := config.Get()

	// Create theme applier
	applier := theme.NewApplier(paths.ConfigDir, paths.DataDir)

	// Get colors as string map
	colors := s.GetColors()

	// Determine which apps to theme based on config
	apps := []string{}

	if cfg.Theme.EnableBtop {
		apps = append(apps, "btop")
	}
	if cfg.Theme.EnableDiscord {
		apps = append(apps, "discord")
	}
	if cfg.Theme.EnableFuzzel {
		apps = append(apps, "fuzzel")
	}
	if cfg.Theme.EnableGtk {
		apps = append(apps, "gtk")
	}
	if cfg.Theme.EnableQt {
		apps = append(apps, "qt")
	}
	if cfg.Theme.EnableSpicetify {
		apps = append(apps, "spicetify")
	}

	// Apply theme to each app
	var errors []string
	for _, app := range apps {
		if err := applier.ApplyTheme(app, colors, s.Mode); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", app, err))
			logger.Error("Failed to apply theme", "app", app, "error", err)
		} else {
			logger.Info("Applied theme", "app", app)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to apply theme to some apps:\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
