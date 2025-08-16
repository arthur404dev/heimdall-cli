package scheme

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
	"github.com/spf13/cobra"
)

// listCommand creates the scheme list subcommand
func listCommand() *cobra.Command {
	var (
		schemeName   string
		flavour      string
		listNames    bool
		listFlavours bool
		listModes    bool
		listVariants bool
		treeView     bool
		showColors   bool
		jsonOutput   bool
		sourceFilter string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available schemes, flavours, or modes",
		Long: `List available color schemes, flavours, or modes.
		
Examples:
  heimdall scheme list                    # Tree view showing scheme structure
  heimdall scheme list -c                 # Tree view with color previews
  heimdall scheme list --json             # List all schemes in JSON format
  heimdall scheme list -n                 # List scheme names only
  heimdall scheme list -f                 # List flavours for current scheme
  heimdall scheme list -m                 # List modes for current scheme/flavour
  heimdall scheme list -v                 # List Material You variants
  heimdall scheme list -s rosepine        # List flavours for rosepine
  heimdall scheme list -s rosepine -f main # List modes for rosepine/main`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manager := scheme.NewManager()

			// Handle tree view (showColors implies treeView)
			if treeView || showColors {
				return listTreeView(manager, showColors, sourceFilter)
			}

			// Handle compatibility flags
			if listNames {
				return listSchemeNames(manager, sourceFilter)
			}

			if listFlavours {
				return listCurrentFlavours(manager, schemeName)
			}

			if listModes {
				return listCurrentModes(manager, schemeName, flavour)
			}

			if listVariants {
				return listMaterialYouVariants()
			}

			// Handle JSON output
			if jsonOutput {
				return listJSONFormat()
			}

			// If no flags, show tree view (default)
			if schemeName == "" && flavour == "" {
				return listTreeView(manager, false, sourceFilter)
			}

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
	cmd.Flags().StringVar(&flavour, "flavour", "", "Flavour name (when used with -s)")
	cmd.Flags().BoolVarP(&listNames, "names", "n", false, "List scheme names only")
	cmd.Flags().BoolVarP(&listFlavours, "flavours", "f", false, "List flavours for current scheme")
	cmd.Flags().BoolVarP(&listModes, "modes", "m", false, "List modes for current scheme/flavour")
	cmd.Flags().BoolVarP(&listVariants, "variants", "v", false, "List Material You variants")
	cmd.Flags().BoolVarP(&treeView, "tree", "t", false, "Display schemes in tree view with structure")
	cmd.Flags().BoolVarP(&showColors, "colors", "c", false, "Show color preview in tree view")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (legacy)")
	cmd.Flags().StringVar(&sourceFilter, "source", "", "Filter by source (bundled, user, generated)")

	return cmd
}

// SchemeOutput represents the output format for JSON compatibility
type SchemeOutput map[string]map[string]interface{}

// listJSONFormat outputs schemes in JSON format including all sources with metadata
func listJSONFormat() error {
	manager := scheme.NewManager()
	output := make(SchemeOutput)

	// Get all schemes from all sources (bundled, user, generated)
	schemes, err := manager.ListSchemes()
	if err != nil {
		return err
	}

	// Process each scheme
	for _, schemeName := range schemes {
		// Skip only the "schemes" directory itself (if it appears)
		if schemeName == "schemes" {
			continue
		}

		// Get source of the scheme
		source := manager.GetSchemeSource(schemeName)
		sourceStr := "bundled" // default
		switch source {
		case scheme.SourceUser:
			sourceStr = "user"
		case scheme.SourceGenerated:
			sourceStr = "generated"
		case scheme.SourceBundled:
			sourceStr = "bundled"
		}

		// Get flavours for this scheme
		flavours, err := manager.ListFlavours(schemeName)
		if err != nil {
			continue // Skip schemes we can't read
		}

		schemeInfo := make(map[string]interface{})
		schemeInfo["source"] = sourceStr

		flavoursData := make(map[string]map[string]string)
		for _, flavourName := range flavours {
			// Get modes for this flavour
			modes, err := manager.ListModes(schemeName, flavourName)
			if err != nil {
				continue
			}

			// For JSON output, we need to get the colors for each mode
			// We'll use dark mode as the default for the JSON format
			modeToUse := "dark"
			if !contains(modes, "dark") && len(modes) > 0 {
				modeToUse = modes[0]
			}

			// Load the scheme
			schemeData, err := manager.LoadScheme(schemeName, flavourName, modeToUse)
			if err != nil {
				continue
			}

			// Extract colors, stripping # prefix for compatibility
			colors := make(map[string]string)
			for key, value := range schemeData.GetColors() {
				// Strip # prefix for JSON format compatibility
				colors[key] = strings.TrimPrefix(value, "#")
			}

			// Add to output only if we have colors
			if len(colors) > 0 {
				flavoursData[flavourName] = colors
			}
		}

		if len(flavoursData) > 0 {
			schemeInfo["flavours"] = flavoursData
			output[schemeName] = schemeInfo
		}
	}

	if err != nil {
		return fmt.Errorf("failed to walk embedded schemes: %w", err)
	}

	// Add dynamic scheme if available
	dynamicColors := readDynamicColors()
	if len(dynamicColors) > 0 {
		dynamicInfo := make(map[string]interface{})
		dynamicInfo["source"] = "generated"
		dynamicInfo["flavours"] = map[string]map[string]string{
			"default": dynamicColors,
		}
		output["dynamic"] = dynamicInfo
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
		// No legacy state to check
		return map[string]string{}
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

	// Try reading from dynamic cache (JSON format)
	dynamicCache := filepath.Join(os.Getenv("HOME"), ".cache", "heimdall", "schemes", "dynamic", "default", "dark.json")
	if colors, err := readMaterialYouJSONFile(dynamicCache); err == nil {
		return colors
	}

	// Try old .txt format for backward compatibility
	dynamicCache = filepath.Join(os.Getenv("HOME"), ".cache", "heimdall", "schemes", "dynamic", "default", "dark.txt")
	if colors, err := readMaterialYouColorFile(dynamicCache); err == nil {
		return colors
	}

	return map[string]string{}
}

// readMaterialYouColorFile reads a Material You format color file (.txt format)
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

// readMaterialYouJSONFile reads a Material You format JSON color file
func readMaterialYouJSONFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var schemeData map[string]interface{}
	if err := json.Unmarshal(content, &schemeData); err != nil {
		return nil, err
	}

	colors := make(map[string]string)

	// Try to get colours (British spelling) or colors (American spelling)
	var colorsMap map[string]interface{}
	if c, ok := schemeData["colours"].(map[string]interface{}); ok {
		colorsMap = c
	} else if c, ok := schemeData["colors"].(map[string]interface{}); ok {
		colorsMap = c
	}

	// Convert colors to string map, stripping # prefix
	if colorsMap != nil {
		for key, value := range colorsMap {
			if colorStr, ok := value.(string); ok {
				// Strip # prefix for consistency
				colors[key] = strings.TrimPrefix(colorStr, "#")
			}
		}
	}

	return colors, nil
}

// listSchemeNames lists all available scheme names
func listSchemeNames(manager *scheme.Manager, sourceFilter string) error {
	schemes, err := manager.ListSchemes()
	if err != nil {
		return err
	}

	sort.Strings(schemes)
	for _, schemeName := range schemes {
		// Apply source filter if specified
		if sourceFilter != "" {
			source := manager.GetSchemeSource(schemeName)
			switch sourceFilter {
			case "bundled":
				if source != scheme.SourceBundled {
					continue
				}
			case "user":
				if source != scheme.SourceUser {
					continue
				}
			case "generated":
				if source != scheme.SourceGenerated {
					continue
				}
			}
		}
		fmt.Println(schemeName)
	}
	return nil
}

// listCurrentFlavours lists flavours for the current scheme or specified scheme
func listCurrentFlavours(manager *scheme.Manager, schemeName string) error {
	if schemeName == "" {
		// Get current scheme
		current, err := manager.GetCurrent()
		if err != nil {
			return err
		}
		schemeName = current.Name
	}

	flavours, err := manager.ListFlavours(schemeName)
	if err != nil {
		return err
	}

	sort.Strings(flavours)
	for _, flavour := range flavours {
		fmt.Println(flavour)
	}
	return nil
}

// listCurrentModes lists modes for the current scheme/flavour or specified scheme/flavour
func listCurrentModes(manager *scheme.Manager, schemeName, flavour string) error {
	if schemeName == "" || flavour == "" {
		// Get current scheme
		current, err := manager.GetCurrent()
		if err != nil {
			return err
		}
		if schemeName == "" {
			schemeName = current.Name
		}
		if flavour == "" {
			flavour = current.Flavour
		}
	}

	modes, err := manager.ListModes(schemeName, flavour)
	if err != nil {
		return err
	}

	sort.Strings(modes)
	for _, mode := range modes {
		fmt.Println(mode)
	}
	return nil
}

// listMaterialYouVariants lists available Material You variants
func listMaterialYouVariants() error {
	variants := []string{
		"tonalspot",
		"neutral",
		"vibrant",
		"expressive",
		"rainbow",
		"fruitsalad",
		"content",
		"monochrome",
	}

	for _, variant := range variants {
		fmt.Println(variant)
	}
	return nil
}

// listTreeView displays schemes in an organized tree structure with optional color previews
func listTreeView(manager *scheme.Manager, showColors bool, sourceFilter string) error {
	schemes, err := manager.ListSchemes()
	if err != nil {
		return fmt.Errorf("failed to list schemes: %w", err)
	}

	// Filter out non-scheme entries
	var validSchemes []string
	for _, s := range schemes {
		// Skip only the "schemes" directory itself (if it appears)
		if s != "schemes" {
			validSchemes = append(validSchemes, s)
		}
	}

	sort.Strings(validSchemes)

	fmt.Printf("\033[36;1mAvailable Color Schemes\033[0m\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	for i, schemeName := range validSchemes {
		// Get source of the scheme
		source := manager.GetSchemeSource(schemeName)

		// Apply source filter if specified
		if sourceFilter != "" {
			switch sourceFilter {
			case "bundled":
				if source != scheme.SourceBundled {
					continue
				}
			case "user":
				if source != scheme.SourceUser {
					continue
				}
			case "generated":
				if source != scheme.SourceGenerated {
					continue
				}
			}
		}

		// Print scheme name with source indicator
		sourceIndicator := ""
		switch source {
		case scheme.SourceUser:
			sourceIndicator = " \033[32m[user]\033[0m"
		case scheme.SourceGenerated:
			sourceIndicator = " \033[33m[generated]\033[0m"
		case scheme.SourceBundled:
			sourceIndicator = " \033[34m[bundled]\033[0m"
		}

		fmt.Printf("\033[35;1m%s\033[0m%s\n", schemeName, sourceIndicator)

		// Get flavours for this scheme
		flavours, err := manager.ListFlavours(schemeName)
		if err != nil {
			continue // Skip if can't get flavours
		}

		sort.Strings(flavours)

		for j, flavourName := range flavours {
			isLastFlavour := j == len(flavours)-1

			// Get modes for this flavour
			modes, err := manager.ListModes(schemeName, flavourName)
			if err != nil {
				// Print flavour without modes
				if isLastFlavour {
					fmt.Printf("  └── %s\n", flavourName)
				} else {
					fmt.Printf("  ├── %s\n", flavourName)
				}
				continue
			}

			sort.Strings(modes)

			// Print flavour
			if isLastFlavour {
				fmt.Printf("  └── \033[34m%s\033[0m", flavourName)
			} else {
				fmt.Printf("  ├── \033[34m%s\033[0m", flavourName)
			}

			// Show color preview if requested
			if showColors && len(modes) > 0 {
				// Load the first mode to get colors
				scheme, err := manager.LoadScheme(schemeName, flavourName, modes[0])
				if err == nil && scheme != nil && scheme.Colours != nil {
					fmt.Printf("  ")
					// Show up to 8 main colors as small blocks
					colorKeys := []string{"base", "text", "red", "green", "yellow", "blue", "pink", "teal"}
					shown := 0
					for _, key := range colorKeys {
						if color, ok := scheme.Colours[key]; ok && shown < 8 {
							// Display small color block
							r, g, b := hexToRGB(color)
							fmt.Printf("\033[48;2;%d;%d;%dm \033[0m", r, g, b)
							shown++
						}
					}
					// If we didn't find named colors, show first 8 colors
					if shown == 0 {
						for _, color := range scheme.Colours {
							if shown >= 8 {
								break
							}
							r, g, b := hexToRGB(color)
							fmt.Printf("\033[48;2;%d;%d;%dm \033[0m", r, g, b)
							shown++
						}
					}
				}
			}

			// Print modes
			if len(modes) > 0 {
				fmt.Printf(" (%s)", strings.Join(modes, ", "))
			}
			fmt.Printf("\n")
		}

		// Add spacing between schemes except for the last one
		if i < len(validSchemes)-1 {
			fmt.Printf("\n")
		}
	}

	return nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
