package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/config/manager"
	"github.com/arthur404dev/heimdall-cli/internal/config/providers"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	mgr *manager.Manager
)

// Command returns the config command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [subcommand]",
		Short: "Manage heimdall configuration",
		Long: `Manage heimdall configuration with powerful discovery and exploration features.

The config command provides comprehensive configuration management including:
  • Discovery of all available options with descriptions
  • Search and filtering capabilities
  • Visual browsing with color-coded values
  • Automatic default handling (no config file required!)
  • Validation and migration support

Common Usage:
  heimdall config list                    # Browse all configuration options
  heimdall config search theme            # Find theme-related options
  heimdall config describe theme.enableGtk # Get detailed info about an option
  heimdall config effective --diff        # Show current config with customizations highlighted
  
Configuration Management:
  heimdall config get cli theme.enableGtk # Get a specific value
  heimdall config set cli theme.enableGtk false # Set a value
  heimdall config defaults --show         # Show all default values
  heimdall config refresh                 # Update config with new defaults

Advanced Features:
  heimdall config list --modified         # Show only customized values
  heimdall config list --category theme   # Filter by category
  heimdall config list --copy theme.enableGtk # Copy path to clipboard
  heimdall config validate cli            # Validate configuration

Note: You can run heimdall without any config file! Defaults are automatically applied.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the configuration manager
			mgr = manager.NewManager()

			// Try to load paths from main config
			configPath := os.Getenv("HEIMDALL_CONFIG")
			if configPath == "" {
				configPath = os.ExpandEnv("$HOME/.config/heimdall/config.json")
			}

			if _, err := os.Stat(configPath); err == nil {
				if err := mgr.LoadPathsFromConfig(configPath); err != nil {
					logger.Debug("Failed to load paths from config: %v", err)
				}
			}

			// Initialize the manager
			if err := mgr.Initialize(); err != nil {
				return fmt.Errorf("failed to initialize config manager: %w", err)
			}

			return nil
		},
	}

	// Add subcommands
	cmd.AddCommand(listCommand())
	cmd.AddCommand(getCommand())
	cmd.AddCommand(setCommand())
	cmd.AddCommand(validateCommand())
	cmd.AddCommand(saveCommand())
	cmd.AddCommand(loadCommand())
	cmd.AddCommand(schemaCommand())
	cmd.AddCommand(defaultsCommand())
	cmd.AddCommand(refreshCommand())

	// Add new discovery subcommands
	cmd.AddCommand(searchCommand())
	cmd.AddCommand(describeCommand())
	cmd.AddCommand(effectiveCommand())

	// Add 'all' subcommand for operations on all domains
	cmd.AddCommand(allCommand())

	// Register shell completions for config paths
	RegisterCompletions(cmd)

	return cmd
}

// listCommand lists all configuration options with descriptions
func listCommand() *cobra.Command {
	var category string
	var showTypes bool
	var filterType string
	var showModified bool
	var copyPath string
	var interactive bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration options with descriptions",
		Long: `List all available configuration options in a beautiful tree format.
		
Shows:
  - Configuration paths with visual indicators
  - Types and current values with color coding
  - Descriptions for each option
  - Default values with comparison
  - Modified values highlighted in different colors
  
Examples:
  heimdall config list                      # List all options
  heimdall config list --category theme     # List only theme options
  heimdall config list --type bool          # List only boolean options
  heimdall config list --modified           # Show only modified values
  heimdall config list --copy theme.enableGtk # Copy a config path to clipboard
  heimdall config list --interactive        # Interactive browsing mode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize metadata registry
			if err := config.InitializeRegistry(); err != nil {
				return fmt.Errorf("failed to initialize metadata: %w", err)
			}

			// Load current configuration to show values
			if err := config.Load(); err != nil {
				// Not critical, just means we won't show current values
				logger.Debug("Could not load config for current values: %v", err)
			}

			// Handle copy to clipboard
			if copyPath != "" {
				return copyConfigPath(copyPath)
			}

			var fields map[string]*config.FieldMetadata

			// Apply filters
			if category != "" {
				// Filter by category
				fields = config.MetadataRegistry.GetFieldsByPrefix(category)
				if len(fields) == 0 {
					return fmt.Errorf("no configuration options found for category '%s'", category)
				}
			} else {
				// Get all fields
				fields = config.MetadataRegistry.GetAllFields()
			}

			// Filter by type if specified
			if filterType != "" {
				fields = filterByType(fields, filterType)
			}

			// Filter to show only modified values if requested
			if showModified {
				fields = filterModifiedOnly(fields)
			}

			if len(fields) == 0 {
				if showModified {
					fmt.Println("No modified configuration values found.")
					fmt.Println("\nAll configuration options are using their default values.")
					fmt.Println("Use 'heimdall config list' to see all available options.")
				} else {
					fmt.Println("No configuration options found matching the filters.")
				}
				return nil
			}

			// Display in beautiful format with enhanced features
			if interactive {
				return displayInteractiveConfig(fields)
			}
			displayEnhancedConfigOptions(fields, showTypes)

			return nil
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category (theme, scheme, wallpaper, etc.)")
	cmd.Flags().BoolVarP(&showTypes, "types", "t", false, "Show type information for each option")
	cmd.Flags().StringVar(&filterType, "type", "", "Filter by type (bool, string, int, etc.)")
	cmd.Flags().BoolVarP(&showModified, "modified", "m", false, "Show only modified values")
	cmd.Flags().StringVar(&copyPath, "copy", "", "Copy a specific config path to clipboard")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Enable interactive browsing mode")

	return cmd
}

// getCommand gets a configuration value
func getCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [domain] [path]",
		Short: "Get a configuration value",
		Long: `Get a specific configuration value from a domain.

The domain can be:
  • cli - Main heimdall configuration
  • shell - Shell-specific configuration
  • all - Get value from all domains

Examples:
  heimdall config get cli theme.enableGtk      # Get GTK theme setting
  heimdall config get cli scheme.default       # Get default color scheme
  heimdall config get all appearance.colorScheme # Get from all domains`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]
			path := args[1]

			value, err := mgr.Get(domain, path)
			if err != nil {
				return err
			}

			// Format output based on type
			switch v := value.(type) {
			case string:
				fmt.Println(v)
			case bool, int, float64:
				fmt.Println(v)
			default:
				// For complex types, use JSON
				data, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format value: %w", err)
				}
				fmt.Println(string(data))
			}

			return nil
		},
	}
}

// setCommand sets a configuration value
func setCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set [domain] [path] [value]",
		Short: "Set a configuration value",
		Long: `Set a configuration value in a specific domain.

Values are automatically saved to disk after setting.
Complex values can be provided as JSON strings.

Examples:
  heimdall config set cli theme.enableGtk false
  heimdall config set cli scheme.default "catppuccin-mocha"
  heimdall config set cli wallpaper.directories '["~/Pictures", "~/Wallpapers"]'
  heimdall config set all appearance.colorScheme "gruvbox-dark"`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]
			path := args[1]
			valueStr := args[2]

			// Try to parse the value as JSON first
			var value interface{}
			if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
				// If not valid JSON, treat as string
				value = valueStr
			}

			if err := mgr.Set(domain, path, value); err != nil {
				return err
			}

			// Save the configuration
			if err := mgr.Save(domain); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Printf("✓ Set %s.%s to %v\n", domain, path, value)
			return nil
		},
	}
}

// validateCommand validates a configuration
func validateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [domain]",
		Short: "Validate a configuration against its schema",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]

			if err := mgr.Validate(domain); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			fmt.Printf("✓ Configuration '%s' is valid\n", domain)
			return nil
		},
	}
}

// saveCommand saves a configuration
func saveCommand() *cobra.Command {
	return &cobra.Command{
		Use:        "save [domain]",
		Short:      "Save a configuration to disk",
		Deprecated: "Configuration is automatically saved when using 'config set'. This command will be removed in v0.3.0",
		Args:       cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⚠️  Warning: This command is deprecated and will be removed in v0.3.0")
			fmt.Println("   Configuration is automatically saved when using 'config set'")
			fmt.Println()

			domain := args[0]

			if err := mgr.Save(domain); err != nil {
				return err
			}

			fmt.Printf("✓ Saved configuration '%s'\n", domain)
			return nil
		},
	}
}

// loadCommand loads a configuration
func loadCommand() *cobra.Command {
	return &cobra.Command{
		Use:        "load [domain]",
		Short:      "Load a configuration from disk",
		Deprecated: "Configuration is automatically loaded when needed. This command will be removed in v0.3.0",
		Args:       cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("⚠️  Warning: This command is deprecated and will be removed in v0.3.0")
			fmt.Println("   Configuration is automatically loaded when needed")
			fmt.Println()

			domain := args[0]

			if err := mgr.Load(domain); err != nil {
				return err
			}

			fmt.Printf("✓ Loaded configuration '%s'\n", domain)
			return nil
		},
	}
}

// schemaCommand displays the schema for a domain
func schemaCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "schema [domain]",
		Short: "Display the JSON schema for a configuration domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			domain := args[0]

			schema, err := mgr.GetSchema(domain)
			if err != nil {
				return err
			}

			data, err := schema.ToJSON()
			if err != nil {
				return fmt.Errorf("failed to format schema: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}
}

// defaultsCommand shows or resets configuration to defaults
func defaultsCommand() *cobra.Command {
	var force bool
	var showOnly bool
	var format string

	cmd := &cobra.Command{
		Use:   "defaults",
		Short: "Show or reset configuration to defaults",
		Long: `Show all default configuration values or reset configuration to defaults.
		
When used with --show flag:
  - Displays all default configuration values
  - Shows in tree format by default
  - Can output as JSON with --format json
  
When used without --show flag:
  - Backs up your current configuration
  - Resets all values to defaults
  - Preserves the backup in ~/.config/heimdall/config.json.backup

Examples:
  heimdall config defaults --show           # Show all defaults
  heimdall config defaults --show --format json  # Show defaults as JSON
  heimdall config defaults                  # Reset to defaults (with confirmation)
  heimdall config defaults --force          # Reset to defaults (no confirmation)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If show flag is set, just display defaults
			if showOnly {
				// Get default configuration
				defaultCfg := config.GetDefaults()

				switch format {
				case "json":
					// Output as JSON
					data, err := json.MarshalIndent(defaultCfg, "", "  ")
					if err != nil {
						return fmt.Errorf("failed to format defaults: %w", err)
					}
					fmt.Println(string(data))
				default:
					// Display as tree
					fmt.Printf("\033[36;1mDEFAULT CONFIGURATION VALUES\033[0m\n")
					fmt.Println(strings.Repeat("━", 80))
					fmt.Println()
					displayConfigStruct(reflect.ValueOf(defaultCfg).Elem(), reflect.TypeOf(defaultCfg).Elem(), "", false)
				}
				return nil
			}

			// Otherwise, proceed with reset logic
			// Import the main config package
			mainConfig := "github.com/arthur404dev/heimdall-cli/internal/config"
			_ = mainConfig // We'll use the actual config package

			configPath := os.ExpandEnv("$HOME/.config/heimdall/config.json")
			backupPath := configPath + ".backup"

			// Check if config exists
			configExists := false
			if _, err := os.Stat(configPath); err == nil {
				configExists = true
			}

			// If config exists and not forcing, ask for confirmation
			if configExists && !force {
				fmt.Println("⚠️  WARNING: This will reset your configuration to defaults!")
				fmt.Printf("Your current configuration will be backed up to: %s\n\n", backupPath)
				fmt.Print("Are you sure you want to continue? (y/N): ")

				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))

				if response != "y" && response != "yes" {
					fmt.Println("❌ Operation cancelled")
					return nil
				}
			}

			// Backup existing config if it exists
			if configExists {
				// Read current config
				currentData, err := os.ReadFile(configPath)
				if err != nil {
					return fmt.Errorf("failed to read current config: %w", err)
				}

				// Write backup with timestamp
				timestamp := time.Now().Format("20060102-150405")
				backupPathWithTime := fmt.Sprintf("%s.%s", backupPath, timestamp)
				if err := os.WriteFile(backupPathWithTime, currentData, 0644); err != nil {
					return fmt.Errorf("failed to create backup: %w", err)
				}

				// Also create a simple .backup file for easy access
				if err := os.WriteFile(backupPath, currentData, 0644); err != nil {
					// Not critical, just log
					logger.Warn("Failed to create simple backup file", "error", err)
				}

				fmt.Printf("✓ Current configuration backed up to:\n")
				fmt.Printf("  - %s (latest)\n", backupPath)
				fmt.Printf("  - %s (timestamped)\n", backupPathWithTime)
			}

			// Remove current config
			if configExists {
				if err := os.Remove(configPath); err != nil {
					return fmt.Errorf("failed to remove current config: %w", err)
				}
			}

			// Now load the config which will create a new one with defaults
			// We need to use the actual config package's functions
			if err := config.Load(); err != nil {
				return fmt.Errorf("failed to reset to defaults: %w", err)
			}

			fmt.Println("\n✓ Configuration has been reset to defaults!")
			fmt.Println("\nYou can restore your previous configuration with:")
			fmt.Printf("  cp %s %s\n", backupPath, configPath)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	cmd.Flags().BoolVarP(&showOnly, "show", "s", false, "Show default values without resetting")
	cmd.Flags().StringVar(&format, "format", "tree", "Output format for --show (tree, json)")

	return cmd
}

// refreshCommand refreshes configuration with new fields while preserving customizations
func refreshCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh configuration with new default fields",
		Long: `Refresh the heimdall configuration to include any new fields from updates.
		
This command will:
  - Load your current configuration
  - Merge in any new default fields
  - Preserve all your customizations
  - Save the updated configuration

This is useful after updating heimdall to ensure you have all new configuration options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := os.ExpandEnv("$HOME/.config/heimdall/config.json")

			// Check if config exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Println("No configuration file found. Creating new configuration with defaults...")
				// Just load the config which will create it with defaults
				if err := config.Load(); err != nil {
					return fmt.Errorf("failed to create configuration: %w", err)
				}
				fmt.Println("✓ Configuration created with defaults")
				return nil
			}

			// Create a backup first
			backupPath := configPath + ".refresh-backup"
			currentData, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read current config: %w", err)
			}

			if err := os.WriteFile(backupPath, currentData, 0644); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}

			fmt.Printf("✓ Current configuration backed up to: %s\n", backupPath)

			// Load the config - this will automatically merge defaults with existing values
			if err := config.Load(); err != nil {
				return fmt.Errorf("failed to refresh configuration: %w", err)
			}

			// The Load() function already saves the merged config, but let's make sure
			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save refreshed configuration: %w", err)
			}

			// Check what was added
			var oldConfig, newConfig map[string]interface{}
			if err := json.Unmarshal(currentData, &oldConfig); err == nil {
				if newData, err := os.ReadFile(configPath); err == nil {
					if err := json.Unmarshal(newData, &newConfig); err == nil {
						// Compare and show what was added
						added := findNewFields(oldConfig, newConfig, "")
						if len(added) > 0 {
							fmt.Println("\n✅ New configuration fields added:")
							for _, field := range added {
								fmt.Printf("  - %s\n", field)
							}
						} else {
							fmt.Println("\n✅ Configuration is already up to date")
						}
					}
				}
			}

			fmt.Println("\n✓ Configuration has been refreshed with latest defaults!")
			fmt.Println("Your customizations have been preserved.")

			return nil
		},
	}

	return cmd
}

// findNewFields compares two config maps and returns fields that are in new but not in old
func findNewFields(old, new map[string]interface{}, prefix string) []string {
	var added []string

	for key, newValue := range new {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		oldValue, exists := old[key]
		if !exists {
			// This field is new
			added = append(added, fullKey)
		} else {
			// Check nested objects
			if newMap, ok := newValue.(map[string]interface{}); ok {
				if oldMap, ok := oldValue.(map[string]interface{}); ok {
					// Recursively check nested fields
					nestedAdded := findNewFields(oldMap, newMap, fullKey)
					added = append(added, nestedAdded...)
				}
			}
		}
	}

	return added
}

// allCommand performs operations on all domains
func allCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all [operation]",
		Short: "Perform operations on all configuration domains",
		Long: `Perform operations on all configuration domains at once.

Examples:
  heimdall config all validate  # Validate all configurations
  heimdall config all save      # Save all configurations
  heimdall config all load      # Load all configurations`,
	}

	// Add subcommands for all operations
	cmd.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: "Validate all configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			var errors []string

			err := mgr.ApplyAll(func(domain string, provider providers.Provider) error {
				if err := mgr.Validate(domain); err != nil {
					errors = append(errors, fmt.Sprintf("%s: %v", domain, err))
					return nil // Continue with other domains
				}
				fmt.Printf("✓ %s: valid\n", domain)
				return nil
			})

			if len(errors) > 0 {
				fmt.Println("\nValidation errors:")
				for _, e := range errors {
					fmt.Printf("  ✗ %s\n", e)
				}
				return fmt.Errorf("validation failed for %d domain(s)", len(errors))
			}

			if err != nil {
				return err
			}

			fmt.Println("\n✓ All configurations are valid")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "save",
		Short: "Save all configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := mgr.SaveAll(); err != nil {
				return err
			}
			fmt.Println("✓ Saved all configurations")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "load",
		Short: "Load all configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := mgr.LoadAll(); err != nil {
				return err
			}
			fmt.Println("✓ Loaded all configurations")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get [path]",
		Short: "Get a value from all configurations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			domains := mgr.ListDomains()
			sort.Strings(domains)

			found := false
			for _, domain := range domains {
				value, err := mgr.Get(domain, path)
				if err != nil {
					// Path doesn't exist in this domain, skip
					continue
				}

				found = true
				// Format output
				switch v := value.(type) {
				case string:
					fmt.Printf("%s: %s\n", domain, v)
				case bool, int, float64:
					fmt.Printf("%s: %v\n", domain, v)
				default:
					data, _ := json.Marshal(v)
					fmt.Printf("%s: %s\n", domain, string(data))
				}
			}

			if !found {
				return fmt.Errorf("path '%s' not found in any configuration", path)
			}

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set [path] [value]",
		Short: "Set a value in all configurations that have the path",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			valueStr := args[1]

			// Try to parse the value as JSON first
			var value interface{}
			if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
				// If not valid JSON, treat as string
				value = valueStr
			}

			domains := mgr.ListDomains()
			updated := []string{}

			for _, domain := range domains {
				// Check if path exists in this domain
				if _, err := mgr.Get(domain, path); err != nil {
					// Path doesn't exist, skip
					continue
				}

				if err := mgr.Set(domain, path, value); err != nil {
					logger.Warn("Failed to set %s.%s: %v", domain, path, err)
					continue
				}

				if err := mgr.Save(domain); err != nil {
					logger.Warn("Failed to save %s: %v", domain, err)
					continue
				}

				updated = append(updated, domain)
				fmt.Printf("✓ Set %s.%s to %v\n", domain, path, value)
			}

			if len(updated) == 0 {
				return fmt.Errorf("path '%s' not found in any configuration", path)
			}

			return nil
		},
	})

	return cmd
}

// Helper function to format path for display
func formatPath(domain, path string) string {
	if path == "" {
		return domain
	}
	return fmt.Sprintf("%s.%s", domain, path)
}

// Helper function to parse domain and path from combined string
func parseDomainPath(combined string) (domain, path string) {
	parts := strings.SplitN(combined, ".", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// searchCommand searches for configuration options
func searchCommand() *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for configuration options by name or description",
		Long: `Search for configuration options using fuzzy matching.
		
Searches through:
  - Option names
  - Descriptions
  - JSON paths
  
Examples:
  heimdall config search theme       # Find all theme-related options
  heimdall config search "gtk"       # Find GTK-related options
  heimdall config search enable      # Find all enable/disable options
  heimdall config search --all       # Show all configuration options`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize metadata registry if not already done
			if err := config.InitializeRegistry(); err != nil {
				return fmt.Errorf("failed to initialize metadata: %w", err)
			}

			var results map[string]*config.FieldMetadata

			if showAll || (len(args) == 0) {
				// Show all fields
				results = config.MetadataRegistry.GetAllFields()
			} else {
				// Search for specific query
				query := strings.Join(args, " ")
				results = config.MetadataRegistry.SearchFields(query)
			}

			if len(results) == 0 {
				if showAll || len(args) == 0 {
					fmt.Println("No configuration options found.")
				} else {
					fmt.Printf("No configuration options found matching '%s'.\n", strings.Join(args, " "))
				}
				return nil
			}

			// Display results in a beautiful format
			displayConfigOptions(results)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all configuration options")

	return cmd
}

// describeCommand shows detailed information about a configuration option
func describeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "describe [path]",
		Short: "Show detailed information about a configuration option",
		Long: `Display detailed information about a specific configuration option.
		
Shows:
  - Full description
  - Type information
  - Default value
  - Example usage
  - Current value (if set)
  
Examples:
  heimdall config describe theme.enableGtk
  heimdall config describe scheme.default
  heimdall config describe wallpaper.directories`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			// Initialize metadata registry
			if err := config.InitializeRegistry(); err != nil {
				return fmt.Errorf("failed to initialize metadata: %w", err)
			}

			// Get field metadata
			field, exists := config.MetadataRegistry.GetFieldMetadata(path)
			if !exists {
				return fmt.Errorf("configuration option '%s' not found", path)
			}

			// Display detailed information
			displayFieldDetails(path, field)

			return nil
		},
	}
}

// effectiveCommand shows the effective (merged) configuration
func effectiveCommand() *cobra.Command {
	var format string
	var showDiff bool

	cmd := &cobra.Command{
		Use:   "effective",
		Short: "Show the effective configuration (user + defaults merged)",
		Long: `Display the effective configuration that is currently in use.
		
This shows the result of merging user configuration with defaults.
Options that differ from defaults are highlighted.
		
Examples:
  heimdall config effective                # Show effective config
  heimdall config effective --format json  # Output as JSON
  heimdall config effective --diff         # Highlight user customizations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load the configuration
			if err := config.Load(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Get effective configuration
			effectiveCfg := config.EffectiveConfig()

			switch format {
			case "json":
				// Output as JSON
				data, err := json.MarshalIndent(effectiveCfg, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format configuration: %w", err)
				}
				fmt.Println(string(data))

			case "yaml":
				// For YAML, we'd need to import a YAML library
				// For now, just use the tree format
				fallthrough

			default:
				// Display as tree with highlighting
				if showDiff {
					displayEffectiveConfigWithDiff(effectiveCfg)
				} else {
					displayEffectiveConfig(effectiveCfg)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "tree", "Output format (tree, json, yaml)")
	cmd.Flags().BoolVarP(&showDiff, "diff", "d", false, "Highlight values that differ from defaults")

	return cmd
}

// displayConfigOptions displays configuration options in a beautiful tree format
func displayConfigOptions(fields map[string]*config.FieldMetadata) {
	displayEnhancedConfigOptions(fields, false)
}

// displayEnhancedConfigOptions displays configuration options with enhanced features
func displayEnhancedConfigOptions(fields map[string]*config.FieldMetadata, showTypes bool) {
	// ANSI color codes
	const (
		colorReset   = "\033[0m"
		colorBold    = "\033[1m"
		colorCyan    = "\033[36m"
		colorGreen   = "\033[32m"
		colorYellow  = "\033[33m"
		colorBlue    = "\033[34m"
		colorGray    = "\033[90m"
		colorRed     = "\033[31m"
		colorMagenta = "\033[35m"
		colorOrange  = "\033[38;5;208m"
	)

	// Get defaults for comparison
	defaults := config.GetDefaults()
	defaultsValue := reflect.ValueOf(defaults).Elem()
	defaultsType := reflect.TypeOf(defaults).Elem()

	fmt.Printf("%s%sCONFIGURATION OPTIONS%s\n", colorCyan, colorBold, colorReset)
	fmt.Println(strings.Repeat("━", 80))

	// Show legend for color coding
	fmt.Printf("\n%sLegend:%s ", colorBold, colorReset)
	fmt.Printf("%s●%s Default  ", colorGray, colorReset)
	fmt.Printf("%s●%s Modified  ", colorMagenta, colorReset)
	fmt.Printf("%s●%s User Set  ", colorOrange, colorReset)
	fmt.Printf("%s✓%s Enabled  ", colorGreen, colorReset)
	fmt.Printf("%s✗%s Disabled\n", colorRed, colorReset)
	fmt.Println()

	// Group fields by category
	categories := make(map[string][]*config.FieldMetadata)
	var categoryOrder []string
	seenCategories := make(map[string]bool)

	// Create a sorted list of paths for consistent ordering
	var paths []string
	for path := range fields {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		field := fields[path]
		parts := strings.Split(path, ".")
		if len(parts) > 0 {
			category := parts[0]
			if !seenCategories[category] {
				seenCategories[category] = true
				categoryOrder = append(categoryOrder, category)
			}
			categories[category] = append(categories[category], field)
		}
	}

	// Display each category
	for i, category := range categoryOrder {
		// Count non-container fields in this category
		visibleFields := 0
		for _, field := range categories[category] {
			if !(field.Type == "object" && len(field.Children) > 0) {
				visibleFields++
			}
		}

		// Skip empty categories (when all fields are containers)
		if visibleFields == 0 {
			continue
		}

		// Format category name
		categoryTitle := strings.ToUpper(category[:1]) + category[1:]
		fmt.Printf("%s%s Settings (%s.*)%s\n", colorBlue, categoryTitle, category, colorReset)

		// Display fields in this category
		categoryFields := categories[category]
		for j, field := range categoryFields {
			// Skip container objects
			if field.Type == "object" && len(field.Children) > 0 {
				continue
			}

			// Determine tree character with better visual indicators
			isLast := j == len(categoryFields)-1
			treeChar := "├─"
			continueLine := "│ "
			if isLast {
				treeChar = "└─"
				continueLine = "  "
			}

			// Check if value is modified from default
			isModified := false
			isUserSet := config.IsUserSet(field.Path)
			currentValue := ""
			defaultValue := getDefaultValueByPath(defaultsValue, defaultsType, field.Path)

			if isUserSet {
				val := viper.Get(field.Path)
				isModified = !reflect.DeepEqual(val, defaultValue)

				// Format current value with color coding
				switch v := val.(type) {
				case bool:
					if v {
						currentValue = fmt.Sprintf("%s✓ true%s", colorGreen, colorReset)
					} else {
						currentValue = fmt.Sprintf("%s✗ false%s", colorRed, colorReset)
					}
				case string:
					if v != "" {
						color := colorYellow
						if isModified {
							color = colorMagenta
						}
						currentValue = fmt.Sprintf("%s\"%s\"%s", color, v, colorReset)
					}
				case []string:
					if len(v) > 0 {
						color := colorYellow
						if isModified {
							color = colorMagenta
						}
						currentValue = fmt.Sprintf("%s[%d items]%s", color, len(v), colorReset)
					}
				default:
					if v != nil {
						color := colorYellow
						if isModified {
							color = colorMagenta
						}
						currentValue = fmt.Sprintf("%s%v%s", color, v, colorReset)
					}
				}
			} else {
				// Show default value in gray
				switch v := defaultValue.(type) {
				case bool:
					if v {
						currentValue = fmt.Sprintf("%s✓ true (default)%s", colorGray, colorReset)
					} else {
						currentValue = fmt.Sprintf("%s✗ false (default)%s", colorGray, colorReset)
					}
				case string:
					if v != "" {
						currentValue = fmt.Sprintf("%s\"%s\" (default)%s", colorGray, v, colorReset)
					}
				default:
					if v != nil {
						currentValue = fmt.Sprintf("%s%v (default)%s", colorGray, v, colorReset)
					}
				}
			}

			// Format field name with modification indicator
			fieldName := field.Path
			if strings.HasPrefix(fieldName, category+".") {
				fieldName = strings.TrimPrefix(fieldName, category+".")
			}

			// Add modification indicator
			modIndicator := " "
			if isModified {
				modIndicator = fmt.Sprintf("%s●%s", colorMagenta, colorReset)
			} else if isUserSet {
				modIndicator = fmt.Sprintf("%s●%s", colorOrange, colorReset)
			} else {
				// Default value - show gray circle
				modIndicator = fmt.Sprintf("%s●%s", colorGray, colorReset)
			}

			// Format type if requested
			typeStr := ""
			if showTypes {
				typeStr = fmt.Sprintf(" %s[%s]%s", colorGray, field.Type, colorReset)
			}

			// Format default value for comparison
			defaultStr := ""
			if field.Default != "" && isModified {
				defaultStr = fmt.Sprintf(" %s(was: %s)%s", colorGray, field.Default, colorReset)
			}

			// Print field line with enhanced formatting
			fmt.Printf("%s %s %-25s%s %s%s\n",
				treeChar,
				modIndicator,
				fieldName,
				typeStr,
				currentValue,
				defaultStr)

			// Print description (indented)
			if field.Description != "" {
				fmt.Printf("%s └─ %s%s%s\n", continueLine, colorGray, field.Description, colorReset)
			}
		}

		// Add spacing between categories
		if i < len(categoryOrder)-1 {
			fmt.Println()
		}
	}

	// Show summary statistics
	fmt.Println()
	fmt.Println(strings.Repeat("─", 80))

	// Count only non-container fields
	totalFields := 0
	modifiedCount := 0
	userSetCount := 0

	for path, field := range fields {
		// Skip container objects in counting
		if field.Type == "object" && len(field.Children) > 0 {
			continue
		}

		totalFields++

		if config.IsUserSet(path) {
			userSetCount++
			val := viper.Get(path)
			defaultVal := getDefaultValueByPath(defaultsValue, defaultsType, path)
			if !reflect.DeepEqual(val, defaultVal) {
				modifiedCount++
			}
		}
	}

	// Only show summary if there are actual fields to display
	if totalFields > 0 {
		fmt.Printf("%sSummary:%s Total: %d | Modified: %s%d%s | User Set: %s%d%s | Using Defaults: %d\n",
			colorBold, colorReset,
			totalFields,
			colorMagenta, modifiedCount, colorReset,
			colorOrange, userSetCount, colorReset,
			totalFields-userSetCount)
	}
}

// displayFieldDetails displays detailed information about a single field
func displayFieldDetails(path string, field *config.FieldMetadata) {
	// ANSI color codes
	const (
		colorReset  = "\033[0m"
		colorBold   = "\033[1m"
		colorCyan   = "\033[36m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorGray   = "\033[90m"
		colorRed    = "\033[31m"
	)

	fmt.Printf("%s%sCONFIGURATION DETAILS%s\n", colorCyan, colorBold, colorReset)
	fmt.Println(strings.Repeat("━", 80))
	fmt.Println()

	// Path
	fmt.Printf("%sPath:%s      %s\n", colorBlue, colorReset, path)

	// Type
	fmt.Printf("%sType:%s      %s\n", colorBlue, colorReset, field.Type)

	// Description
	if field.Description != "" {
		fmt.Printf("%sDescription:%s\n", colorBlue, colorReset)
		// Word wrap description at 70 characters
		words := strings.Fields(field.Description)
		line := "  "
		for _, word := range words {
			if len(line)+len(word)+1 > 70 {
				fmt.Println(line)
				line = "  " + word
			} else {
				if line == "  " {
					line += word
				} else {
					line += " " + word
				}
			}
		}
		if line != "  " {
			fmt.Println(line)
		}
	}

	// Default value
	if field.Default != "" {
		fmt.Printf("\n%sDefault:%s   %s%s%s\n", colorBlue, colorReset, colorYellow, field.Default, colorReset)
	}

	// Example
	if field.Example != "" {
		fmt.Printf("%sExample:%s   %s%s%s\n", colorBlue, colorReset, colorGreen, field.Example, colorReset)
	}

	// Current value
	fmt.Printf("\n%sCurrent Value:%s\n", colorBlue, colorReset)
	if viper.IsSet(path) {
		val := viper.Get(path)
		switch v := val.(type) {
		case bool:
			if v {
				fmt.Printf("  %s✓ true%s\n", colorGreen, colorReset)
			} else {
				fmt.Printf("  %s✗ false%s\n", colorRed, colorReset)
			}
		case string:
			fmt.Printf("  %s\"%s\"%s\n", colorYellow, v, colorReset)
		case []string:
			fmt.Printf("  %s[\n", colorYellow)
			for _, item := range v {
				fmt.Printf("    \"%s\",\n", item)
			}
			fmt.Printf("  ]%s\n", colorReset)
		case map[string]interface{}:
			data, _ := json.MarshalIndent(v, "  ", "  ")
			fmt.Printf("  %s%s%s\n", colorYellow, string(data), colorReset)
		default:
			fmt.Printf("  %s%v%s\n", colorYellow, v, colorReset)
		}
	} else {
		fmt.Printf("  %s(using default)%s\n", colorGray, colorReset)
	}

	// Deprecated warning
	if field.Deprecated != "" {
		fmt.Printf("\n%s⚠️  DEPRECATED:%s %s\n", colorRed, colorReset, field.Deprecated)
	}

	// Usage examples
	fmt.Printf("\n%sUsage Examples:%s\n", colorBlue, colorReset)
	fmt.Printf("  heimdall config get cli %s\n", path)
	fmt.Printf("  heimdall config set cli %s <value>\n", path)

	// Related fields (children)
	if len(field.Children) > 0 {
		fmt.Printf("\n%sNested Fields:%s\n", colorBlue, colorReset)
		for _, child := range field.Children {
			fmt.Printf("  • %s\n", child)
		}
	}
}

// displayEffectiveConfig displays the effective configuration in tree format
func displayEffectiveConfig(cfg *config.Config) {
	// ANSI color codes
	const (
		colorReset  = "\033[0m"
		colorBold   = "\033[1m"
		colorCyan   = "\033[36m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorGray   = "\033[90m"
	)

	fmt.Printf("%s%sEFFECTIVE CONFIGURATION%s\n", colorCyan, colorBold, colorReset)
	fmt.Println(strings.Repeat("━", 80))
	fmt.Println()

	// Use reflection to display the configuration
	displayConfigStruct(reflect.ValueOf(cfg).Elem(), reflect.TypeOf(cfg).Elem(), "", false)
}

// displayEffectiveConfigWithDiff displays the effective configuration with diff highlighting
func displayEffectiveConfigWithDiff(cfg *config.Config) {
	// ANSI color codes
	const (
		colorReset   = "\033[0m"
		colorBold    = "\033[1m"
		colorCyan    = "\033[36m"
		colorGreen   = "\033[32m"
		colorYellow  = "\033[33m"
		colorBlue    = "\033[34m"
		colorGray    = "\033[90m"
		colorMagenta = "\033[35m"
	)

	fmt.Printf("%s%sEFFECTIVE CONFIGURATION%s %s(customized values highlighted)%s\n",
		colorCyan, colorBold, colorReset, colorGray, colorReset)
	fmt.Println(strings.Repeat("━", 80))
	fmt.Println()

	// Use reflection to display the configuration
	displayConfigStruct(reflect.ValueOf(cfg).Elem(), reflect.TypeOf(cfg).Elem(), "", true)
}

// displayConfigStruct recursively displays a configuration struct
func displayConfigStruct(v reflect.Value, t reflect.Type, prefix string, showDiff bool) {
	// ANSI color codes
	const (
		colorReset   = "\033[0m"
		colorGreen   = "\033[32m"
		colorYellow  = "\033[33m"
		colorBlue    = "\033[34m"
		colorGray    = "\033[90m"
		colorMagenta = "\033[35m"
	)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Tag.Get("mapstructure")
		}
		if jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		// Build full path
		fullPath := jsonName
		if prefix != "" {
			fullPath = prefix + "." + jsonName
		}

		// Check if value is customized (differs from default)
		isCustomized := false
		if showDiff && viper.IsSet(fullPath) {
			isCustomized = true
		}

		// Handle different types
		switch value.Kind() {
		case reflect.Struct:
			// Print struct header
			if isCustomized {
				fmt.Printf("%s%s:%s\n", colorMagenta, jsonName, colorReset)
			} else {
				fmt.Printf("%s%s:%s\n", colorBlue, jsonName, colorReset)
			}
			// Recursively display nested struct
			displayConfigStruct(value, value.Type(), fullPath, showDiff)

		case reflect.Map:
			// Print map header
			if isCustomized {
				fmt.Printf("%s%s:%s %s[map]%s\n", colorMagenta, jsonName, colorReset, colorGray, colorReset)
			} else {
				fmt.Printf("%s%s:%s %s[map]%s\n", colorBlue, jsonName, colorReset, colorGray, colorReset)
			}
			// Display map entries
			iter := value.MapRange()
			for iter.Next() {
				k := iter.Key()
				v := iter.Value()
				fmt.Printf("  %s: %v\n", k, v.Interface())
			}

		case reflect.Slice:
			// Print slice
			if isCustomized {
				fmt.Printf("%s%s:%s %s[%d items]%s\n",
					colorMagenta, jsonName, colorReset, colorYellow, value.Len(), colorReset)
			} else {
				fmt.Printf("%s:%s %s[%d items]%s\n",
					jsonName, colorReset, colorGray, value.Len(), colorReset)
			}

		case reflect.Bool:
			boolVal := value.Bool()
			colorCode := colorGray
			symbol := "✗"
			if boolVal {
				colorCode = colorGreen
				symbol = "✓"
			}
			if isCustomized {
				fmt.Printf("%s%s:%s %s%s %v%s\n",
					colorMagenta, jsonName, colorReset, colorCode, symbol, boolVal, colorReset)
			} else {
				fmt.Printf("%s: %s%s %v%s\n",
					jsonName, colorCode, symbol, boolVal, colorReset)
			}

		case reflect.String:
			strVal := value.String()
			if strVal == "" {
				fmt.Printf("%s: %s(empty)%s\n", jsonName, colorGray, colorReset)
			} else {
				if isCustomized {
					fmt.Printf("%s%s:%s %s\"%s\"%s\n",
						colorMagenta, jsonName, colorReset, colorYellow, strVal, colorReset)
				} else {
					fmt.Printf("%s: \"%s\"\n", jsonName, strVal)
				}
			}

		case reflect.Int, reflect.Int32, reflect.Int64:
			if isCustomized {
				fmt.Printf("%s%s:%s %s%v%s\n",
					colorMagenta, jsonName, colorReset, colorYellow, value.Int(), colorReset)
			} else {
				fmt.Printf("%s: %v\n", jsonName, value.Int())
			}

		case reflect.Float32, reflect.Float64:
			if isCustomized {
				fmt.Printf("%s%s:%s %s%v%s\n",
					colorMagenta, jsonName, colorReset, colorYellow, value.Float(), colorReset)
			} else {
				fmt.Printf("%s: %v\n", jsonName, value.Float())
			}

		default:
			// For other types, just print the value
			fmt.Printf("%s: %v\n", jsonName, value.Interface())
		}
	}
}

// filterByType filters fields by their type
func filterByType(fields map[string]*config.FieldMetadata, filterType string) map[string]*config.FieldMetadata {
	filtered := make(map[string]*config.FieldMetadata)
	for path, field := range fields {
		if field.Type == filterType {
			filtered[path] = field
		}
	}
	return filtered
}

// filterModifiedOnly returns only fields that have been modified from defaults
func filterModifiedOnly(fields map[string]*config.FieldMetadata) map[string]*config.FieldMetadata {
	filtered := make(map[string]*config.FieldMetadata)

	// Get default config for comparison
	defaults := config.GetDefaults()
	defaultsValue := reflect.ValueOf(defaults).Elem()
	defaultsType := reflect.TypeOf(defaults).Elem()

	for path, field := range fields {
		// Skip container objects
		if field.Type == "object" && len(field.Children) > 0 {
			continue
		}

		// Check if the value is set by the user
		if config.IsUserSet(path) {
			// Get current value
			currentValue := viper.Get(path)

			// Get default value using reflection
			defaultValue := getDefaultValueByPath(defaultsValue, defaultsType, path)

			// Compare values
			if !reflect.DeepEqual(currentValue, defaultValue) {
				filtered[path] = field
			}
		}
	}
	return filtered
}

// getDefaultValueByPath retrieves a value from the defaults struct by path
func getDefaultValueByPath(v reflect.Value, t reflect.Type, path string) interface{} {
	parts := strings.Split(path, ".")
	current := v
	currentType := t

	for _, part := range parts {
		if current.Kind() != reflect.Struct {
			return nil
		}

		// Find field by json tag
		found := false
		for i := 0; i < current.NumField(); i++ {
			field := currentType.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = field.Tag.Get("mapstructure")
			}
			jsonName := strings.Split(jsonTag, ",")[0]

			if jsonName == part {
				current = current.Field(i)
				currentType = field.Type
				found = true
				break
			}
		}

		if !found {
			return nil
		}
	}

	return current.Interface()
}

// copyConfigPath copies a configuration path to the clipboard
func copyConfigPath(path string) error {
	// Initialize metadata registry to validate path
	if err := config.InitializeRegistry(); err != nil {
		return fmt.Errorf("failed to initialize metadata: %w", err)
	}

	// Check if path exists
	if _, exists := config.MetadataRegistry.GetFieldMetadata(path); !exists {
		return fmt.Errorf("configuration path '%s' not found", path)
	}

	// Use xclip or pbcopy depending on the platform
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
		}
	default:
		return fmt.Errorf("clipboard copy not supported on %s", runtime.GOOS)
	}

	// Write path to clipboard
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start clipboard command: %w", err)
	}

	if _, err := stdin.Write([]byte(path)); err != nil {
		return fmt.Errorf("failed to write to clipboard: %w", err)
	}

	if err := stdin.Close(); err != nil {
		return fmt.Errorf("failed to close stdin: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("clipboard command failed: %w", err)
	}

	fmt.Printf("✓ Copied '%s' to clipboard\n", path)
	return nil
}

// displayInteractiveConfig displays configuration in an interactive browser mode
func displayInteractiveConfig(fields map[string]*config.FieldMetadata) error {
	// For now, just display with enhanced formatting
	// A full interactive mode would require a TUI library like bubbletea
	fmt.Println("Interactive mode (use arrow keys to navigate, q to quit):")
	fmt.Println("Note: Full interactive mode requires additional TUI implementation")
	fmt.Println()
	displayEnhancedConfigOptions(fields, true)
	return nil
}

// fuzzySearch performs fuzzy string matching
func fuzzySearch(query string, targets []string) []string {
	var matches []string
	query = strings.ToLower(query)

	for _, target := range targets {
		targetLower := strings.ToLower(target)
		// Simple substring matching for now
		if strings.Contains(targetLower, query) {
			matches = append(matches, target)
		}
	}

	return matches
}
