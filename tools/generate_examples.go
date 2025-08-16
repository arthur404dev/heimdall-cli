//go:build ignore
// +build ignore

// This tool generates example configuration files for Heimdall CLI
// Run with: go run tools/generate_examples.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getDefaultConfig returns the default configuration
func getDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"version": "0.2.0",
		"theme": map[string]interface{}{
			"enableTerm":      true,
			"enableHypr":      true,
			"enableDiscord":   true,
			"enableSpicetify": true,
			"enableFuzzel":    true,
			"enableBtop":      true,
			"enableGtk":       true,
			"enableQt":        true,
			"enableKitty":     true,
			"enableAlacritty": false,
			"enableWezterm":   false,
		},
		"shell": map[string]interface{}{
			"command":     "qs",
			"args":        []string{"-c", "heimdall", "-n"},
			"daemon_port": 9999,
			"log_file":    "shell.log",
			"pid_file":    "shell.pid",
			"ipc_timeout": 5,
		},
		"scheme": map[string]interface{}{
			"default":        "rosepine",
			"auto_mode":      true,
			"material_you":   true,
			"user_paths":     []string{"~/.config/heimdall/schemes"},
			"generated_path": "~/.local/share/heimdall/schemes",
		},
		"wallpaper": map[string]interface{}{
			"directory":  "~/Pictures/Wallpapers",
			"filter":     true,
			"threshold":  0.8,
			"smart_mode": true,
			"extensions": []string{".jpg", ".jpeg", ".png", ".webp"},
		},
		"screenshot": map[string]interface{}{
			"directory":           "~/Pictures/Screenshots",
			"file_format":         "png",
			"file_name_pattern":   "screenshot_%Y%m%d_%H%M%S",
			"copy_to_clipboard":   true,
			"open_after_capture":  false,
			"capture_mouse":       false,
			"capture_decorations": true,
			"delay":               0,
			"quality":             100,
		},
		"record": map[string]interface{}{
			"directory":         "~/Videos/Recordings",
			"file_format":       "mp4",
			"file_name_pattern": "recording_%Y%m%d_%H%M%S",
			"fps":               30,
			"quality":           "high",
			"audio":             true,
			"microphone":        false,
			"show_mouse":        true,
		},
		"pip": map[string]interface{}{
			"position":      "bottom-right",
			"size":          "medium",
			"opacity":       0.9,
			"always_on_top": true,
			"border":        true,
			"shadow":        true,
		},
		"idle": map[string]interface{}{
			"timeout":       300,
			"lock_command":  "hyprlock",
			"sleep_command": "systemctl suspend",
			"enable_lock":   true,
			"enable_sleep":  false,
			"warning_time":  30,
		},
		"emoji": map[string]interface{}{
			"picker_command":        "fuzzel-emoji",
			"copy_to_clipboard":     true,
			"close_after_selection": true,
		},
		"clipboard": map[string]interface{}{
			"history_size": 100,
			"persist":      true,
			"sync":         false,
		},
		"toggle": map[string]interface{}{
			"animations":   true,
			"blur":         true,
			"transparency": true,
			"shadows":      true,
		},
	}
}

// generateFullExample generates a complete example config with all defaults
func generateFullExample(outputDir string) error {
	defaults := getDefaultConfig()

	// Marshal with indentation for readability
	data, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal defaults: %w", err)
	}

	outputPath := filepath.Join(outputDir, "config-full-example.json")
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write full example: %w", err)
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

// generateDocumentedExample generates an example with accompanying documentation
func generateDocumentedExample(outputDir string) error {
	defaults := getDefaultConfig()

	// Generate JSON
	data, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal defaults: %w", err)
	}

	jsonPath := filepath.Join(outputDir, "config-documented.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write documented example: %w", err)
	}

	// Generate accompanying documentation
	var doc strings.Builder
	doc.WriteString("# Configuration Documentation\n\n")
	doc.WriteString("This document describes all configuration options for the `config-documented.json` file.\n\n")

	doc.WriteString("## Theme Settings\n\n")
	doc.WriteString("Controls which applications receive theme updates.\n\n")
	doc.WriteString("- `theme.enableTerm`: Apply themes to terminal emulators\n")
	doc.WriteString("- `theme.enableHypr`: Apply themes to Hyprland\n")
	doc.WriteString("- `theme.enableDiscord`: Apply themes to Discord clients\n")
	doc.WriteString("- `theme.enableSpicetify`: Apply themes to Spotify via Spicetify\n")
	doc.WriteString("- `theme.enableFuzzel`: Apply themes to Fuzzel launcher\n")
	doc.WriteString("- `theme.enableBtop`: Apply themes to btop system monitor\n")
	doc.WriteString("- `theme.enableGtk`: Apply themes to GTK applications\n")
	doc.WriteString("- `theme.enableQt`: Apply themes to Qt applications\n")
	doc.WriteString("- `theme.enableKitty`: Apply themes to Kitty terminal\n")
	doc.WriteString("- `theme.enableAlacritty`: Apply themes to Alacritty terminal\n")
	doc.WriteString("- `theme.enableWezterm`: Apply themes to WezTerm terminal\n\n")

	doc.WriteString("## Shell Integration\n\n")
	doc.WriteString("Settings for Quickshell integration.\n\n")
	doc.WriteString("- `shell.command`: Command to execute shell (default: \"qs\")\n")
	doc.WriteString("- `shell.args`: Arguments for shell command\n")
	doc.WriteString("- `shell.daemon_port`: Port for daemon communication\n")
	doc.WriteString("- `shell.log_file`: Path to log file\n")
	doc.WriteString("- `shell.pid_file`: Path to PID file\n")
	doc.WriteString("- `shell.ipc_timeout`: IPC timeout in seconds\n\n")

	doc.WriteString("## Scheme Settings\n\n")
	doc.WriteString("Color scheme management configuration.\n\n")
	doc.WriteString("- `scheme.default`: Default color scheme to use\n")
	doc.WriteString("- `scheme.auto_mode`: Automatically switch between light/dark variants\n")
	doc.WriteString("- `scheme.material_you`: Generate Material You schemes from wallpapers\n")
	doc.WriteString("- `scheme.user_paths`: Paths to search for user-defined schemes\n")
	doc.WriteString("- `scheme.generated_path`: Path to store generated schemes\n\n")

	doc.WriteString("## Wallpaper Settings\n\n")
	doc.WriteString("Wallpaper management configuration.\n\n")
	doc.WriteString("- `wallpaper.directory`: Directory containing wallpapers\n")
	doc.WriteString("- `wallpaper.filter`: Enable smart filtering based on aspect ratio\n")
	doc.WriteString("- `wallpaper.threshold`: Similarity threshold for filtering (0.0-1.0)\n")
	doc.WriteString("- `wallpaper.smart_mode`: Enable intelligent wallpaper selection\n")
	doc.WriteString("- `wallpaper.extensions`: Supported image file extensions\n\n")

	doc.WriteString("## Screenshot Settings\n\n")
	doc.WriteString("Screenshot capture configuration.\n\n")
	doc.WriteString("- `screenshot.directory`: Directory to save screenshots\n")
	doc.WriteString("- `screenshot.file_format`: Image format (png, jpg, etc.)\n")
	doc.WriteString("- `screenshot.file_name_pattern`: Pattern for filename generation\n")
	doc.WriteString("- `screenshot.copy_to_clipboard`: Copy screenshot to clipboard\n")
	doc.WriteString("- `screenshot.open_after_capture`: Open screenshot after capture\n")
	doc.WriteString("- `screenshot.capture_mouse`: Include mouse cursor in screenshot\n")
	doc.WriteString("- `screenshot.capture_decorations`: Include window decorations\n")
	doc.WriteString("- `screenshot.delay`: Delay before capture in seconds\n")
	doc.WriteString("- `screenshot.quality`: Image quality (1-100)\n\n")

	docPath := filepath.Join(outputDir, "config-documented.md")
	if err := os.WriteFile(docPath, []byte(doc.String()), 0644); err != nil {
		return fmt.Errorf("failed to write documentation: %w", err)
	}

	fmt.Printf("Generated: %s\n", jsonPath)
	fmt.Printf("Generated: %s\n", docPath)
	return nil
}

// generateMinimalExamples generates minimal configs for common use cases
func generateMinimalExamples(outputDir string) error {
	examples := []struct {
		name        string
		description string
		config      map[string]interface{}
	}{
		{
			name:        "minimal-theme-only.json",
			description: "Minimal config for theme application only",
			config: map[string]interface{}{
				"version": "0.2.0",
				"theme": map[string]interface{}{
					"enableGtk":     true,
					"enableQt":      true,
					"enableDiscord": true,
				},
			},
		},
		{
			name:        "minimal-wallpaper-only.json",
			description: "Minimal config for wallpaper management only",
			config: map[string]interface{}{
				"version": "0.2.0",
				"wallpaper": map[string]interface{}{
					"directory":  "~/Pictures/Wallpapers",
					"filter":     true,
					"smart_mode": true,
				},
			},
		},
		{
			name:        "minimal-scheme-only.json",
			description: "Minimal config for color scheme management",
			config: map[string]interface{}{
				"version": "0.2.0",
				"scheme": map[string]interface{}{
					"default":      "catppuccin-mocha",
					"auto_mode":    true,
					"material_you": false,
				},
			},
		},
		{
			name:        "minimal-terminal-only.json",
			description: "Minimal config for terminal theming only",
			config: map[string]interface{}{
				"version": "0.2.0",
				"theme": map[string]interface{}{
					"enableTerm":  true,
					"enableKitty": true,
				},
			},
		},
		{
			name:        "minimal-material-you.json",
			description: "Minimal config for Material You wallpaper-based theming",
			config: map[string]interface{}{
				"version": "0.2.0",
				"scheme": map[string]interface{}{
					"material_you": true,
				},
				"wallpaper": map[string]interface{}{
					"directory":  "~/Pictures/Wallpapers",
					"smart_mode": true,
				},
			},
		},
		{
			name:        "minimal-quickshell.json",
			description: "Minimal config for Quickshell integration",
			config: map[string]interface{}{
				"version": "0.2.0",
				"shell": map[string]interface{}{
					"command": "qs",
					"args":    []string{"-c", "heimdall", "-n"},
				},
			},
		},
	}

	// Create a README for the minimal examples
	var readme strings.Builder
	readme.WriteString("# Minimal Configuration Examples\n\n")
	readme.WriteString("These minimal configuration files demonstrate common use cases with only the necessary settings.\n")
	readme.WriteString("Heimdall will use default values for any settings not specified.\n\n")

	for _, example := range examples {
		// Generate JSON file
		data, err := json.MarshalIndent(example.config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %w", example.name, err)
		}

		outputPath := filepath.Join(outputDir, example.name)
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", example.name, err)
		}

		fmt.Printf("Generated: %s\n", outputPath)

		// Add to README
		readme.WriteString(fmt.Sprintf("## %s\n\n", example.name))
		readme.WriteString(fmt.Sprintf("%s\n\n", example.description))
		readme.WriteString("```json\n")
		readme.WriteString(string(data))
		readme.WriteString("\n```\n\n")
	}

	// Write README
	readmePath := filepath.Join(outputDir, "MINIMAL_EXAMPLES.md")
	if err := os.WriteFile(readmePath, []byte(readme.String()), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	fmt.Printf("Generated: %s\n", readmePath)
	return nil
}

// generateWithComments generates a config with inline documentation comments
// Since JSON doesn't support comments, we'll create a JSONC file and a companion MD file
func generateWithComments(outputDir string) error {
	// Create a JSONC (JSON with Comments) version
	var jsonc strings.Builder
	jsonc.WriteString("// Heimdall CLI Configuration File\n")
	jsonc.WriteString("// This file contains all available configuration options with their default values.\n")
	jsonc.WriteString("// You can remove any settings you don't want to customize - Heimdall will use defaults.\n")
	jsonc.WriteString("// Note: This is a JSONC file (JSON with Comments). Remove comments before using as config.json.\n\n")
	jsonc.WriteString("{\n")
	jsonc.WriteString("  // Configuration version for migration\n")
	jsonc.WriteString("  \"version\": \"0.2.0\",\n\n")

	jsonc.WriteString("  // Theme application settings\n")
	jsonc.WriteString("  \"theme\": {\n")
	jsonc.WriteString("    // Apply themes to terminal emulators\n")
	jsonc.WriteString("    \"enableTerm\": true,\n")
	jsonc.WriteString("    // Apply themes to Hyprland window manager\n")
	jsonc.WriteString("    \"enableHypr\": true,\n")
	jsonc.WriteString("    // Apply themes to Discord clients\n")
	jsonc.WriteString("    \"enableDiscord\": true,\n")
	jsonc.WriteString("    // Apply themes to Spotify via Spicetify\n")
	jsonc.WriteString("    \"enableSpicetify\": true,\n")
	jsonc.WriteString("    // Apply themes to Fuzzel launcher\n")
	jsonc.WriteString("    \"enableFuzzel\": true,\n")
	jsonc.WriteString("    // Apply themes to btop system monitor\n")
	jsonc.WriteString("    \"enableBtop\": true,\n")
	jsonc.WriteString("    // Apply themes to GTK applications\n")
	jsonc.WriteString("    \"enableGtk\": true,\n")
	jsonc.WriteString("    // Apply themes to Qt applications\n")
	jsonc.WriteString("    \"enableQt\": true,\n")
	jsonc.WriteString("    // Apply themes to Kitty terminal\n")
	jsonc.WriteString("    \"enableKitty\": true,\n")
	jsonc.WriteString("    // Apply themes to Alacritty terminal\n")
	jsonc.WriteString("    \"enableAlacritty\": false,\n")
	jsonc.WriteString("    // Apply themes to WezTerm terminal\n")
	jsonc.WriteString("    \"enableWezterm\": false\n")
	jsonc.WriteString("  },\n\n")

	jsonc.WriteString("  // Quickshell integration settings\n")
	jsonc.WriteString("  \"shell\": {\n")
	jsonc.WriteString("    // Command to execute shell\n")
	jsonc.WriteString("    \"command\": \"qs\",\n")
	jsonc.WriteString("    // Arguments for shell command\n")
	jsonc.WriteString("    \"args\": [\"-c\", \"heimdall\", \"-n\"],\n")
	jsonc.WriteString("    // Port for daemon communication\n")
	jsonc.WriteString("    \"daemon_port\": 9999,\n")
	jsonc.WriteString("    // Path to log file\n")
	jsonc.WriteString("    \"log_file\": \"shell.log\",\n")
	jsonc.WriteString("    // Path to PID file\n")
	jsonc.WriteString("    \"pid_file\": \"shell.pid\",\n")
	jsonc.WriteString("    // IPC timeout in seconds\n")
	jsonc.WriteString("    \"ipc_timeout\": 5\n")
	jsonc.WriteString("  },\n\n")

	jsonc.WriteString("  // Color scheme settings\n")
	jsonc.WriteString("  \"scheme\": {\n")
	jsonc.WriteString("    // Default color scheme to use\n")
	jsonc.WriteString("    \"default\": \"rosepine\",\n")
	jsonc.WriteString("    // Automatically switch between light/dark variants\n")
	jsonc.WriteString("    \"auto_mode\": true,\n")
	jsonc.WriteString("    // Generate Material You schemes from wallpapers\n")
	jsonc.WriteString("    \"material_you\": true,\n")
	jsonc.WriteString("    // Paths to search for user-defined schemes\n")
	jsonc.WriteString("    \"user_paths\": [\"~/.config/heimdall/schemes\"],\n")
	jsonc.WriteString("    // Path to store generated schemes\n")
	jsonc.WriteString("    \"generated_path\": \"~/.local/share/heimdall/schemes\"\n")
	jsonc.WriteString("  },\n\n")

	jsonc.WriteString("  // Wallpaper management settings\n")
	jsonc.WriteString("  \"wallpaper\": {\n")
	jsonc.WriteString("    // Directory containing wallpapers\n")
	jsonc.WriteString("    \"directory\": \"~/Pictures/Wallpapers\",\n")
	jsonc.WriteString("    // Enable smart filtering based on aspect ratio\n")
	jsonc.WriteString("    \"filter\": true,\n")
	jsonc.WriteString("    // Similarity threshold for filtering (0.0-1.0)\n")
	jsonc.WriteString("    \"threshold\": 0.8,\n")
	jsonc.WriteString("    // Enable intelligent wallpaper selection\n")
	jsonc.WriteString("    \"smart_mode\": true,\n")
	jsonc.WriteString("    // Supported image file extensions\n")
	jsonc.WriteString("    \"extensions\": [\".jpg\", \".jpeg\", \".png\", \".webp\"]\n")
	jsonc.WriteString("  }\n")
	jsonc.WriteString("}\n")

	// Write JSONC file
	jsoncPath := filepath.Join(outputDir, "config-with-comments.jsonc")
	if err := os.WriteFile(jsoncPath, []byte(jsonc.String()), 0644); err != nil {
		return fmt.Errorf("failed to write JSONC: %w", err)
	}

	fmt.Printf("Generated: %s\n", jsoncPath)

	// Also create a clean JSON version
	defaults := getDefaultConfig()
	data, err := json.MarshalIndent(defaults, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal defaults: %w", err)
	}

	jsonPath := filepath.Join(outputDir, "config-default.json")
	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write default JSON: %w", err)
	}

	fmt.Printf("Generated: %s\n", jsonPath)

	return nil
}

func main() {
	outputDir := "docs/examples"

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating example configuration files...")
	fmt.Println()

	// Generate various example configs
	if err := generateFullExample(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate full example: %v\n", err)
		os.Exit(1)
	}

	if err := generateDocumentedExample(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate documented example: %v\n", err)
		os.Exit(1)
	}

	if err := generateMinimalExamples(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate minimal examples: %v\n", err)
		os.Exit(1)
	}

	if err := generateWithComments(outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate commented example: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ All example configuration files generated successfully!")
}
