package scheme

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

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
		noApply      bool
		variant      string
		setName      string
		setFlavour   string
		setMode      string
		setVariant   string
		randomScheme bool
		enableNotify bool
	)

	cmd := &cobra.Command{
		Use:   "set [scheme] [flavour] [mode]",
		Short: "Set the active color scheme",
		Long: `Set the active color scheme and apply theme.
		
Arguments:
  scheme  - Scheme name (optional when using flags)
  flavour - Flavour name (optional, defaults to first available)
  mode    - Mode: dark or light (optional, defaults to dark)
		
Examples:
  heimdall scheme set rosepine            # Use rosepine with defaults
  heimdall scheme set rosepine main       # Use rosepine/main with default mode
  heimdall scheme set rosepine main dark  # Use rosepine/main/dark
  heimdall scheme set -n catppuccin -f mocha -m dark -v blue
  heimdall scheme set -r                  # Random scheme selection
  heimdall scheme set --notify rosepine   # With desktop notifications`,
		Args: cobra.RangeArgs(0, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := scheme.NewManager()

			var schemeName, flavour, mode string

			// Handle random scheme selection
			if randomScheme {
				return setRandomScheme(manager, !noApply, enableNotify)
			}

			// Handle caelestia-compatible flags
			if setName != "" || setFlavour != "" || setMode != "" || setVariant != "" {
				return setSchemeByFlags(manager, setName, setFlavour, setMode, setVariant, !noApply, enableNotify)
			}

			// Handle positional arguments
			if len(args) == 0 {
				return fmt.Errorf("scheme name is required when not using flags")
			}

			schemeName = args[0]
			mode = "dark" // default

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
			if setVariant != "" {
				newScheme.Variant = setVariant
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

				// Send notification if enabled
				if enableNotify {
					notifier := notify.NewNotifier()
					notifier.Send(&notify.Notification{
						Summary: "Scheme Changed",
						Body:    fmt.Sprintf("Applied %s/%s/%s", schemeName, flavour, mode),
						Urgency: notify.UrgencyNormal,
					})
				}
			}

			fmt.Printf("Scheme set to %s/%s/%s\n", schemeName, flavour, mode)
			if setVariant != "" {
				fmt.Printf("Variant: %s\n", setVariant)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noApply, "no-apply", false, "Don't apply theme after setting scheme")
	cmd.Flags().StringVarP(&setName, "name", "n", "", "Set scheme name")
	cmd.Flags().StringVarP(&setFlavour, "flavour", "f", "", "Set flavour")
	cmd.Flags().StringVarP(&setMode, "mode", "m", "", "Set mode")
	cmd.Flags().StringVarP(&setVariant, "variant", "v", "", "Set variant")
	cmd.Flags().BoolVarP(&randomScheme, "random", "r", false, "Random scheme selection")
	cmd.Flags().BoolVar(&enableNotify, "notify", false, "Enable desktop notifications")

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

// setRandomScheme selects and applies a random scheme
func setRandomScheme(manager *scheme.Manager, shouldApplyTheme, shouldNotify bool) error {
	// Get all available schemes
	schemes, err := manager.ListSchemes()
	if err != nil {
		return fmt.Errorf("failed to list schemes: %w", err)
	}

	if len(schemes) == 0 {
		return fmt.Errorf("no schemes available")
	}

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Pick random scheme
	randomScheme := schemes[rand.Intn(len(schemes))]

	// Get flavours for the random scheme
	flavours, err := manager.ListFlavours(randomScheme)
	if err != nil {
		return fmt.Errorf("failed to list flavours for %s: %w", randomScheme, err)
	}

	if len(flavours) == 0 {
		return fmt.Errorf("no flavours available for scheme %s", randomScheme)
	}

	// Pick random flavour
	randomFlavour := flavours[rand.Intn(len(flavours))]

	// Get modes for the random scheme/flavour
	modes, err := manager.ListModes(randomScheme, randomFlavour)
	if err != nil {
		return fmt.Errorf("failed to list modes for %s/%s: %w", randomScheme, randomFlavour, err)
	}

	if len(modes) == 0 {
		return fmt.Errorf("no modes available for scheme %s/%s", randomScheme, randomFlavour)
	}

	// Pick random mode
	randomMode := modes[rand.Intn(len(modes))]

	// Load and apply the random scheme
	newScheme, err := manager.LoadSchemeWithFallback(randomScheme, randomFlavour, randomMode)
	if err != nil {
		return fmt.Errorf("failed to load random scheme: %w", err)
	}

	// Save as current scheme
	if err := manager.SetScheme(newScheme); err != nil {
		return fmt.Errorf("failed to set random scheme: %w", err)
	}

	logger.Info("Random scheme selected",
		"scheme", randomScheme,
		"flavour", randomFlavour,
		"mode", randomMode)

	// Apply theme if enabled
	if shouldApplyTheme {
		if err := applyTheme(newScheme); err != nil {
			logger.Error("Failed to apply theme", "error", err)
			return fmt.Errorf("failed to apply theme: %w", err)
		}

		// Send notification if enabled
		if shouldNotify {
			notifier := notify.NewNotifier()
			notifier.Send(&notify.Notification{
				Summary: "Random Scheme Applied",
				Body:    fmt.Sprintf("Applied %s/%s/%s", randomScheme, randomFlavour, randomMode),
				Urgency: notify.UrgencyNormal,
			})
		}
	}

	fmt.Printf("Random scheme set to %s/%s/%s\n", randomScheme, randomFlavour, randomMode)
	return nil
}

// setSchemeByFlags sets scheme using individual flags
func setSchemeByFlags(manager *scheme.Manager, name, flavour, mode, variant string, shouldApplyTheme, shouldNotify bool) error {
	// Get current scheme to fill in missing values
	current, err := manager.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current scheme: %w", err)
	}

	// Use current values if flags not provided
	if name == "" {
		name = current.Name
	}
	if flavour == "" {
		flavour = current.Flavour
	}
	if mode == "" {
		mode = current.Mode
	}
	if variant == "" {
		variant = current.Variant
	}

	// Validate mode
	if mode != "dark" && mode != "light" {
		return fmt.Errorf("invalid mode: %s (must be 'dark' or 'light')", mode)
	}

	// Load the scheme
	newScheme, err := manager.LoadSchemeWithFallback(name, flavour, mode)
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

	logger.Info("Scheme set by flags",
		"scheme", name,
		"flavour", flavour,
		"mode", mode,
		"variant", variant)

	// Apply theme if enabled
	if shouldApplyTheme {
		if err := applyTheme(newScheme); err != nil {
			logger.Error("Failed to apply theme", "error", err)
			return fmt.Errorf("failed to apply theme: %w", err)
		}

		// Send notification if enabled
		if shouldNotify {
			notifier := notify.NewNotifier()
			notifier.Send(&notify.Notification{
				Summary: "Scheme Changed",
				Body:    fmt.Sprintf("Applied %s/%s/%s", name, flavour, mode),
				Urgency: notify.UrgencyNormal,
			})
		}
	}

	fmt.Printf("Scheme set to %s/%s/%s\n", name, flavour, mode)
	if variant != "" {
		fmt.Printf("Variant: %s\n", variant)
	}

	return nil
}
