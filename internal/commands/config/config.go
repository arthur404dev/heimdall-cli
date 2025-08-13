package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/config/manager"
	"github.com/arthur404dev/heimdall-cli/internal/config/providers"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/spf13/cobra"
)

var (
	mgr *manager.Manager
)

// Command returns the config command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [domain] [operation] [args...]",
		Short: "Manage configuration files",
		Long: `Manage multiple configuration domains with unified interface.
		
Configuration domains are separate config files that can be managed independently.
Each domain can have its own schema and validation rules.

Examples:
  heimdall config list                    # List all configuration domains
  heimdall config cli get theme.enableGtk # Get a specific value
  heimdall config shell set appearance.colorScheme "gruvbox-dark"
  heimdall config all validate            # Validate all configurations`,
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

	// Add 'all' subcommand for operations on all domains
	cmd.AddCommand(allCommand())

	return cmd
}

// listCommand lists all configuration domains
func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			domains := mgr.ListDomains()
			sort.Strings(domains)

			fmt.Println("Available configuration domains:")
			for _, domain := range domains {
				schema, err := mgr.GetSchema(domain)
				desc := ""
				if err == nil && schema != nil {
					desc = schema.Description
					if desc == "" {
						desc = schema.Title
					}
				}

				if desc != "" {
					fmt.Printf("  - %s: %s\n", domain, desc)
				} else {
					fmt.Printf("  - %s\n", domain)
				}
			}

			return nil
		},
	}
}

// getCommand gets a configuration value
func getCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [domain] [path]",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(2),
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
		Args:  cobra.ExactArgs(3),
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
		Use:   "save [domain]",
		Short: "Save a configuration to disk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
		Use:   "load [domain]",
		Short: "Load a configuration from disk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
