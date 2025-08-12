package scheme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/heimdall-cli/heimdall/assets/schemes"
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
  heimdall scheme list                    # List all schemes in Caelestia format
  heimdall scheme list -s rosepine        # List flavours for rosepine
  heimdall scheme list -s rosepine -f main # List modes for rosepine/main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no flags, output in Caelestia JSON format
			if schemeName == "" && flavour == "" {
				return listCaelestiaFormat()
			}

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

			return nil
		},
	}

	cmd.Flags().StringVarP(&schemeName, "scheme", "s", "", "Scheme name")
	cmd.Flags().StringVarP(&flavour, "flavour", "f", "", "Flavour name")

	return cmd
}

// CaelestiaSchemeOutput represents the output format for Caelestia compatibility
type CaelestiaSchemeOutput map[string]map[string]map[string]string

// listCaelestiaFormat outputs schemes in Caelestia-compatible JSON format using embedded assets
func listCaelestiaFormat() error {
	output := make(CaelestiaSchemeOutput)

	// Walk through embedded filesystem
	err := fs.WalkDir(schemes.Content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip root and non-.txt files
		if path == "." || d.IsDir() || !strings.HasSuffix(path, ".txt") {
			return nil
		}

		// Parse path: scheme/flavour/mode.txt
		parts := strings.Split(path, "/")
		if len(parts) != 3 {
			return nil
		}

		schemeName := parts[0]
		flavourName := parts[1]

		// Read the file from embedded FS
		data, err := schemes.Content.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse colors
		colors := make(map[string]string)
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			fields := strings.Fields(line)
			if len(fields) == 2 {
				colors[fields[0]] = fields[1]
			}
		}

		// Add to output
		if output[schemeName] == nil {
			output[schemeName] = make(map[string]map[string]string)
		}
		output[schemeName][flavourName] = colors

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk embedded schemes: %w", err)
	}

	// Add dynamic scheme if available
	dynamicColors := readDynamicColors()
	if len(dynamicColors) > 0 {
		output["dynamic"] = map[string]map[string]string{
			"default": dynamicColors,
		}
	}

	// Output as JSON
	jsonData, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal schemes: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// readDynamicColors reads the current dynamic colors if available
func readDynamicColors() map[string]string {
	// Check for dynamic scheme data in Heimdall state
	stateDir := filepath.Join(os.Getenv("HOME"), ".local", "state", "heimdall")
	schemeFile := filepath.Join(stateDir, "scheme.json")

	// Read the current scheme
	content, err := os.ReadFile(schemeFile)
	if err != nil {
		// Also check Caelestia state for migration compatibility
		stateDir = filepath.Join(os.Getenv("HOME"), ".local", "state", "caelestia")
		schemeFile = filepath.Join(stateDir, "scheme.json")
		content, err = os.ReadFile(schemeFile)
		if err != nil {
			return map[string]string{}
		}
	}

	// Parse JSON
	var schemeData map[string]interface{}
	if err := json.Unmarshal(content, &schemeData); err != nil {
		return map[string]string{}
	}

	// Check if it's a dynamic scheme
	if name, ok := schemeData["name"].(string); ok && name == "dynamic" {
		if colors, ok := schemeData["colours"].(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range colors {
				if str, ok := v.(string); ok {
					result[k] = str
				}
			}
			return result
		}
	}

	// Try reading from dynamic cache
	dynamicCache := filepath.Join(os.Getenv("HOME"), ".cache", "heimdall", "schemes", "dynamic", "default", "dark.txt")
	if colors, err := readMaterialYouColorFile(dynamicCache); err == nil {
		return colors
	}

	// Fallback to Caelestia cache for migration
	dynamicCache = filepath.Join(os.Getenv("HOME"), ".cache", "caelestia", "schemes", "dynamic", "default", "dark.txt")
	if colors, err := readMaterialYouColorFile(dynamicCache); err == nil {
		return colors
	}

	return map[string]string{}
}

// readMaterialYouColorFile reads a Material You format color file
func readMaterialYouColorFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	colors := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 2 {
			colors[parts[0]] = parts[1]
		}
	}

	return colors, nil
}
