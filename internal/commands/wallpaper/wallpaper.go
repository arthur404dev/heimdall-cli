package wallpaper

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/scheme"
	"github.com/arthur404dev/heimdall-cli/internal/scheme/generator"
	"github.com/arthur404dev/heimdall-cli/internal/theme"
	"github.com/arthur404dev/heimdall-cli/internal/utils/hypr"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/material"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/arthur404dev/heimdall-cli/internal/utils/wallpaper"
	"github.com/spf13/cobra"
)

// Command creates the wallpaper command with caelestia compatibility
func Command() *cobra.Command {
	var (
		// Caelestia-compatible flags
		printPath string  // -p, --print [PATH] - Extract and print color scheme
		randomDir string  // -r, --random [DIR] - Select random wallpaper from directory
		filePath  string  // -f, --file PATH - Set specific wallpaper by path
		noFilter  bool    // -n, --no-filter - Disable size filtering for random selection
		threshold float64 // -t, --threshold FLOAT - Minimum size ratio for wallpaper selection
		noSmart   bool    // -N, --no-smart - Disable automatic mode/variant detection

		// Legacy flags (deprecated)
		legacyFilter   bool // --filter (deprecated, use --no-filter instead)
		generateScheme bool // -s, --scheme (deprecated, use smart detection)
		info           bool // -i, --info (deprecated, use --print instead)
	)

	cmd := &cobra.Command{
		Use:   "wallpaper [OPTIONS]",
		Short: "Manage wallpapers with Material You integration",
		Long: `Manage wallpapers with smart filtering and Material You integration.

Caelestia-compatible interface:
  - Set wallpapers from file paths
  - Random wallpaper selection with intelligent size filtering
  - Color extraction for dynamic Material You schemes
  - Automatic light/dark mode detection based on wallpaper brightness
  - JSON output for color schemes

Examples:
  heimdall wallpaper                           # Get current wallpaper path
  heimdall wallpaper -f ~/Pictures/sunset.jpg # Set specific wallpaper
  heimdall wallpaper -r                        # Random wallpaper from default directory
  heimdall wallpaper -r ~/Wallpapers          # Random from custom directory
  heimdall wallpaper -p ~/Pictures/test.jpg   # Extract colors without changing wallpaper
  heimdall wallpaper -f ~/Pictures/dark.jpg -N # Set wallpaper without smart mode detection`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			if err := config.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			cfg := config.Get()

			// Handle deprecated flags with warnings
			if legacyFilter {
				logger.Warn("Flag --filter is deprecated, use --no-filter to disable size filtering")
			}
			if generateScheme {
				logger.Warn("Flag --scheme is deprecated, smart mode detection is now automatic")
			}
			if info {
				logger.Warn("Flag --info is deprecated, use --print instead")
				if len(args) > 0 {
					return printColorScheme(args[0])
				}
			}

			// Handle color extraction/printing
			if printPath != "" {
				return printColorScheme(printPath)
			}
			if cmd.Flags().Changed("print") && printPath == "" {
				// -p without argument uses current wallpaper
				return printCurrentWallpaperScheme()
			}

			// Handle file setting
			if filePath != "" {
				return setWallpaper(filePath, !noSmart)
			}

			// Handle random wallpaper selection
			if randomDir != "" {
				return setRandomWallpaperFromDir(cfg, randomDir, !noFilter, threshold, !noSmart)
			}
			if cmd.Flags().Changed("random") && randomDir == "" {
				// -r without argument uses default directory
				defaultDir := cfg.Wallpaper.Directory
				if defaultDir == "" {
					defaultDir = paths.WallpapersDir
				}
				return setRandomWallpaperFromDir(cfg, defaultDir, !noFilter, threshold, !noSmart)
			}

			// No flags provided - return current wallpaper path
			return getCurrentWallpaper()
		},
	}

	// Caelestia-compatible flags
	cmd.Flags().StringVarP(&printPath, "print", "p", "", "Extract and print color scheme from wallpaper (current wallpaper if no path)")
	cmd.Flags().StringVarP(&randomDir, "random", "r", "", "Select random wallpaper from directory (default: ~/Pictures/Wallpapers)")
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "Set specific wallpaper by path")
	cmd.Flags().BoolVarP(&noFilter, "no-filter", "n", false, "Disable size filtering for random selection")
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", 0.8, "Minimum size ratio for wallpaper selection")
	cmd.Flags().BoolVarP(&noSmart, "no-smart", "N", false, "Disable automatic mode/variant detection")

	// Legacy flags (deprecated but maintained for backward compatibility)
	cmd.Flags().BoolVar(&legacyFilter, "filter", false, "Filter by colourfulness (deprecated)")
	cmd.Flags().BoolVarP(&generateScheme, "scheme", "s", false, "Generate Material You scheme (deprecated)")
	cmd.Flags().BoolVarP(&info, "info", "i", false, "Show wallpaper info (deprecated)")

	// Hide deprecated flags from help
	cmd.Flags().MarkHidden("filter")
	cmd.Flags().MarkHidden("scheme")
	cmd.Flags().MarkHidden("info")

	return cmd
}

// getCurrentWallpaper returns the current wallpaper path
func getCurrentWallpaper() error {
	linkPath := paths.WallpaperLinkPath
	if linkPath == "" {
		linkPath = filepath.Join(paths.StateDir, "current_wallpaper")
	}

	// Check if symlink exists
	if _, err := os.Lstat(linkPath); err != nil {
		return fmt.Errorf("no wallpaper currently set")
	}

	// Read the symlink target
	target, err := os.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("failed to read current wallpaper: %w", err)
	}

	fmt.Println(target)
	return nil
}

// printCurrentWallpaperScheme prints the color scheme of the current wallpaper
func printCurrentWallpaperScheme() error {
	linkPath := paths.WallpaperLinkPath
	if linkPath == "" {
		linkPath = filepath.Join(paths.StateDir, "current_wallpaper")
	}

	// Check if symlink exists
	if _, err := os.Lstat(linkPath); err != nil {
		return fmt.Errorf("no wallpaper currently set")
	}

	// Read the symlink target
	target, err := os.Readlink(linkPath)
	if err != nil {
		return fmt.Errorf("failed to read current wallpaper: %w", err)
	}

	return printColorScheme(target)
}

// printColorScheme extracts and prints the color scheme from a wallpaper in JSON format
func printColorScheme(wallpaperPath string) error {
	// Expand path
	if strings.HasPrefix(wallpaperPath, "~/") {
		home, _ := os.UserHomeDir()
		wallpaperPath = filepath.Join(home, wallpaperPath[2:])
	}

	// Check if file exists
	if _, err := os.Stat(wallpaperPath); err != nil {
		return fmt.Errorf("wallpaper not found: %w", err)
	}

	// Open image
	file, err := os.Open(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to open wallpaper: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate Material You palette
	generator := material.NewGenerator()
	palette, err := generator.GenerateFromImage(img)
	if err != nil {
		return fmt.Errorf("failed to generate palette: %w", err)
	}

	// Determine mode and variant based on wallpaper
	analyzer := wallpaper.NewAnalyzer()
	mode, err := analyzer.DetermineMode(wallpaperPath)
	if err != nil {
		mode = "dark" // Default to dark
	}

	// Determine variant based on colorfulness
	colourfulness, err := analyzer.AnalyzeColourfulness(wallpaperPath)
	if err != nil {
		colourfulness = 15.0 // Default to content
	}

	variant := "content"
	if colourfulness < 10 {
		variant = "neutral"
	} else if colourfulness > 20 {
		variant = "tonalspot"
	}

	// Create scheme
	materialScheme, err := generator.GenerateScheme(palette.Seed, mode == "dark")
	if err != nil {
		return fmt.Errorf("failed to generate scheme: %w", err)
	}

	// Create caelestia-compatible JSON output with full Heimdall scheme
	output := map[string]interface{}{
		"name":    "dynamic",
		"flavour": "default",
		"mode":    mode,
		"variant": variant,
		"colours": convertMaterialColors(materialScheme), // Use the new comprehensive converter
	}

	// Output JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// setRandomWallpaperFromDir selects and sets a random wallpaper from a directory
func setRandomWallpaperFromDir(cfg *config.Config, wallpaperDir string, enableSizeFilter bool, threshold float64, enableSmartMode bool) error {
	// Expand home directory
	if strings.HasPrefix(wallpaperDir, "~/") {
		home, _ := os.UserHomeDir()
		wallpaperDir = filepath.Join(home, wallpaperDir[2:])
	}

	// Find all image files
	var wallpapers []string
	extensions := []string{".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff"}

	err := filepath.Walk(wallpaperDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if it's an image file
		ext := strings.ToLower(filepath.Ext(path))
		for _, validExt := range extensions {
			if ext == validExt {
				wallpapers = append(wallpapers, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan wallpaper directory: %w", err)
	}

	if len(wallpapers) == 0 {
		return fmt.Errorf("no wallpapers found in %s", wallpaperDir)
	}

	// Apply size filtering if enabled
	if enableSizeFilter {
		filtered, err := filterWallpapersBySize(wallpapers, threshold)
		if err != nil {
			logger.Warn("Failed to apply size filtering", "error", err)
		} else if len(filtered) > 0 {
			wallpapers = filtered
			logger.Info("Applied size filtering", "total", len(wallpapers), "threshold", threshold)
		} else {
			logger.Warn("No wallpapers passed size filter, using all", "threshold", threshold)
		}
	}

	// Select random wallpaper
	rand.Seed(time.Now().UnixNano())
	selected := wallpapers[rand.Intn(len(wallpapers))]

	logger.Info("Selected wallpaper", "path", selected)

	return setWallpaper(selected, enableSmartMode)
}

// filterWallpapersBySize filters wallpapers based on monitor size requirements
func filterWallpapersBySize(wallpapers []string, threshold float64) ([]string, error) {
	// Get monitor information via Hyprland IPC
	client, err := hypr.NewClient()
	if err != nil {
		return wallpapers, fmt.Errorf("failed to create Hyprland client: %w", err)
	}

	monitors, err := client.GetMonitors()
	if err != nil {
		return wallpapers, fmt.Errorf("failed to get monitors: %w", err)
	}

	if len(monitors) == 0 {
		return wallpapers, fmt.Errorf("no monitors found")
	}

	// Find the smallest monitor dimensions
	minWidth := monitors[0].Width
	minHeight := monitors[0].Height
	for _, monitor := range monitors[1:] {
		if monitor.Width < minWidth {
			minWidth = monitor.Width
		}
		if monitor.Height < minHeight {
			minHeight = monitor.Height
		}
	}

	// Required minimum dimensions based on threshold
	reqWidth := int(float64(minWidth) * threshold)
	reqHeight := int(float64(minHeight) * threshold)

	logger.Info("Size filtering criteria", "minWidth", reqWidth, "minHeight", reqHeight, "threshold", threshold)

	// Filter wallpapers by size
	var filtered []string
	analyzer := wallpaper.NewAnalyzer()

	for _, wp := range wallpapers {
		width, height, err := analyzer.GetDimensions(wp)
		if err != nil {
			logger.Warn("Failed to get wallpaper dimensions", "path", wp, "error", err)
			continue
		}

		if width >= reqWidth && height >= reqHeight {
			filtered = append(filtered, wp)
		}
	}

	return filtered, nil
}

// setRandomWallpaper selects and sets a random wallpaper (legacy function)
func setRandomWallpaper(cfg *config.Config, filter bool, threshold float64, generateScheme bool) error {
	wallpaperDir := cfg.Wallpaper.Directory
	if wallpaperDir == "" {
		wallpaperDir = paths.WallpapersDir
	}

	// Convert legacy filter logic to new size filtering logic
	// Legacy filter was based on colourfulness, new system uses size filtering
	enableSizeFilter := !filter // Invert logic: old filter=true meant enable colourfulness filter

	// Use legacy threshold for size filtering if provided, otherwise use default
	sizeThreshold := 0.8
	if threshold != 50.0 {
		// Convert colourfulness threshold to size threshold (rough approximation)
		sizeThreshold = 0.5 + (threshold/100.0)*0.5
	}

	// Call the new directory-based function
	return setRandomWallpaperFromDir(cfg, wallpaperDir, enableSizeFilter, sizeThreshold, generateScheme || cfg.Wallpaper.SmartMode)
}

// setWallpaper sets a specific wallpaper
func setWallpaper(wallpaperPath string, enableSmartMode bool) error {
	// Expand path
	if strings.HasPrefix(wallpaperPath, "~/") {
		home, _ := os.UserHomeDir()
		wallpaperPath = filepath.Join(home, wallpaperPath[2:])
	}

	// Check if file exists
	if _, err := os.Stat(wallpaperPath); err != nil {
		return fmt.Errorf("wallpaper not found: %w", err)
	}

	// Create symlink for current wallpaper
	linkPath := paths.WallpaperLinkPath
	if linkPath == "" {
		linkPath = filepath.Join(paths.StateDir, "current_wallpaper")
	}

	// Remove old link if exists
	os.Remove(linkPath)

	// Create new symlink
	if err := os.Symlink(wallpaperPath, linkPath); err != nil {
		logger.Error("Failed to create wallpaper symlink", "error", err)
	}

	// Set wallpaper using hyprpaper
	// First try using hyprctl to set wallpaper
	cmd := exec.Command("hyprctl", "hyprpaper", "wallpaper", fmt.Sprintf(",path=%s", wallpaperPath))
	if err := cmd.Run(); err != nil {
		// Try alternative method using swww if available
		if _, err := exec.LookPath("swww"); err == nil {
			cmd = exec.Command("swww", "img", wallpaperPath)
			if err := cmd.Run(); err != nil {
				logger.Error("Failed to set wallpaper with swww", "error", err)
			}
		} else {
			logger.Error("Failed to set wallpaper", "error", err)
		}
	}

	// Generate Material You scheme if smart mode is enabled
	if enableSmartMode {
		if err := generateMaterialYouScheme(wallpaperPath); err != nil {
			logger.Error("Failed to generate scheme", "error", err)
		}

		// Update state to indicate generated theme is available
		stateManager := theme.NewStateManager()
		stateManager.SetGeneratedAvailable(wallpaperPath, map[string]string{
			"generated_at": time.Now().Format(time.RFC3339),
		})

		// Check if we should auto-apply
		if stateManager.ShouldAutoApply(scheme.SourceGenerated) {
			// Auto-apply the generated theme
			schemeManager := scheme.NewManager()
			prefs := stateManager.GetPreferences()

			// Load the preferred variant
			variant := prefs.PreferredVariant
			if variant == "" {
				variant = "tonal"
			}
			mode := prefs.PreferredMode
			if mode == "" {
				// Use detected mode
				analyzer := wallpaper.NewAnalyzer()
				mode, _ = analyzer.DetermineMode(wallpaperPath)
				if mode == "" {
					mode = "dark"
				}
			}

			// Load and apply the generated scheme
			generatedScheme, err := schemeManager.LoadScheme("generated", variant, mode)
			if err == nil {
				schemeManager.SetScheme(generatedScheme)

				// Update state
				stateManager.SetCurrent(theme.CurrentTheme{
					Name:    "generated",
					Flavour: variant,
					Mode:    mode,
					Source:  scheme.SourceGenerated,
					Metadata: map[string]string{
						"wallpaper": wallpaperPath,
					},
				})

				logger.Info("Auto-applied generated theme", "variant", variant, "mode", mode)
			}
		} else {
			// Notify that new theme is available but not auto-applied
			if stateManager.GetPreferences().NotifyOnGeneration {
				notifier := notify.NewNotifier()
				notifier.Send(&notify.Notification{
					Summary: "New Theme Available",
					Body:    "Generated Material You theme from wallpaper. Use 'heimdall scheme set generated' to apply.",
					Urgency: notify.UrgencyNormal,
				})
			}
		}
	}

	// Send notification
	notifier := notify.NewNotifier()
	notifier.Send(&notify.Notification{
		Summary: "Wallpaper Changed",
		Body:    filepath.Base(wallpaperPath),
		Urgency: notify.UrgencyNormal,
	})

	fmt.Printf("Wallpaper set: %s\n", wallpaperPath)

	return nil
}

// showWallpaperInfo displays information about a wallpaper
func showWallpaperInfo(wallpaperPath string) error {
	// Expand path
	if strings.HasPrefix(wallpaperPath, "~/") {
		home, _ := os.UserHomeDir()
		wallpaperPath = filepath.Join(home, wallpaperPath[2:])
	}

	analyzer := wallpaper.NewAnalyzer()

	info, err := analyzer.Analyze(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to analyze wallpaper: %w", err)
	}

	fmt.Printf("Wallpaper: %s\n", filepath.Base(wallpaperPath))
	fmt.Printf("Dimensions: %dx%d\n", info.Width, info.Height)
	fmt.Printf("Colourfulness: %.2f\n", info.Colourfulness)
	fmt.Printf("Suggested mode: %s\n", info.Mode)

	// Show dominant colors
	colors, err := analyzer.AnalyzeDominantColors(wallpaperPath, 5)
	if err == nil && len(colors) > 0 {
		fmt.Println("Dominant colors:")
		for i, color := range colors {
			r := (color >> 16) & 0xFF
			g := (color >> 8) & 0xFF
			b := color & 0xFF
			fmt.Printf("  %d. #%02x%02x%02x\n", i+1, r, g, b)
		}
	}

	return nil
}

// generateMaterialYouScheme generates all Material You variants from the wallpaper
func generateMaterialYouScheme(wallpaperPath string) error {
	logger.Info("Generating Material You schemes from wallpaper")

	// Open image
	file, err := os.Open(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to open wallpaper: %w", err)
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Generate all Material You variants
	wallpaperGen := generator.NewWallpaperGenerator()
	variants, err := wallpaperGen.GenerateAllVariants(img, wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to generate variants: %w", err)
	}

	// Determine preferred mode based on wallpaper
	analyzer := wallpaper.NewAnalyzer()
	preferredMode, err := analyzer.DetermineMode(wallpaperPath)
	if err != nil {
		preferredMode = "dark" // Default to dark
	}

	// Save all variants to user schemes directory
	manager := scheme.NewManager()
	generatedDir := filepath.Join(paths.ConfigDir, "schemes", "generated")

	// Create generated directory if it doesn't exist
	if err := os.MkdirAll(generatedDir, 0755); err != nil {
		return fmt.Errorf("failed to create generated schemes directory: %w", err)
	}

	// Save metadata
	metadata := map[string]interface{}{
		"version": "1.0",
		"source": map[string]interface{}{
			"wallpaper": wallpaperPath,
			"timestamp": time.Now().Format(time.RFC3339),
		},
		"generation": map[string]interface{}{
			"algorithm":     "enhanced-v2",
			"detected_mode": preferredMode,
		},
		"variants": make(map[string]interface{}),
	}

	// Save each variant
	for key, variantScheme := range variants {
		// Parse variant and mode from key (e.g., "vibrant/dark")
		parts := strings.Split(key, "/")
		if len(parts) != 2 {
			continue
		}
		variant := parts[0]
		mode := parts[1]

		// Create variant directory
		variantDir := filepath.Join(generatedDir, variant)
		if err := os.MkdirAll(variantDir, 0755); err != nil {
			logger.Error("Failed to create variant directory", "variant", variant, "error", err)
			continue
		}

		// Save scheme file
		schemeFile := filepath.Join(variantDir, mode+".json")
		schemeData, err := json.MarshalIndent(variantScheme, "", "  ")
		if err != nil {
			logger.Error("Failed to marshal scheme", "variant", variant, "mode", mode, "error", err)
			continue
		}

		if err := os.WriteFile(schemeFile, schemeData, 0644); err != nil {
			logger.Error("Failed to save scheme", "variant", variant, "mode", mode, "error", err)
			continue
		}

		// Add to metadata
		if _, ok := metadata["variants"].(map[string]interface{})[variant]; !ok {
			metadata["variants"].(map[string]interface{})[variant] = make(map[string]interface{})
		}
		metadata["variants"].(map[string]interface{})[variant].(map[string]interface{})[mode] = map[string]interface{}{
			"path": schemeFile,
		}

		logger.Info("Saved variant", "variant", variant, "mode", mode)
	}

	// Save metadata file
	metadataFile := filepath.Join(generatedDir, "metadata.json")
	metadataData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal metadata", "error", err)
	} else {
		if err := os.WriteFile(metadataFile, metadataData, 0644); err != nil {
			logger.Error("Failed to save metadata", "error", err)
		}
	}

	// Set the preferred variant as active
	// Default to "content" variant in preferred mode
	preferredVariant := "content"
	preferredKey := fmt.Sprintf("%s/%s", preferredVariant, preferredMode)

	var activeScheme *scheme.Scheme

	if preferredScheme, ok := variants[preferredKey]; ok {
		// Use the preferred variant
		activeScheme = preferredScheme
	} else {
		// Fallback: generate a single scheme using the old method
		materialGen := material.NewGenerator()
		palette, err := materialGen.GenerateFromImage(img)
		if err != nil {
			return fmt.Errorf("failed to generate palette: %w", err)
		}

		materialScheme, err := materialGen.GenerateScheme(palette.Seed, preferredMode == "dark")
		if err != nil {
			return fmt.Errorf("failed to generate scheme: %w", err)
		}

		// Convert to our scheme format
		activeScheme = &scheme.Scheme{
			Name:    "material-you",
			Flavour: "generated",
			Mode:    preferredMode,
			Variant: "wallpaper",
			Colours: convertMaterialColors(materialScheme),
		}
	}

	// Save and set the active scheme
	if err := manager.SaveScheme(activeScheme); err != nil {
		logger.Error("Failed to save active scheme", "error", err)
	}

	if err := manager.SetScheme(activeScheme); err != nil {
		logger.Error("Failed to set active scheme", "error", err)
	}

	// Apply theme to all applications (like scheme set does)
	configDir := paths.ConfigDir
	dataDir := paths.DataDir
	applier := theme.NewApplier(configDir, dataDir)

	// Get list of applications to theme
	apps := []string{"gtk", "qt", "kitty", "alacritty", "wezterm", "nvim",
		"discord", "btop", "fuzzel", "spicetify", "hyprland", "waybar",
		"quickshell", "terminal"}

	// Convert colors to map[string]string format (with # prefix)
	colors := make(map[string]string)
	for k, v := range activeScheme.Colours {
		// Ensure colors have # prefix for applications
		if !strings.HasPrefix(v, "#") {
			colors[k] = "#" + v
		} else {
			colors[k] = v
		}
	}

	// Apply to each application
	var errors []string
	kittyThemed := false

	for _, app := range apps {
		if app == "terminal" {
			// Special handling for terminal sequences
			if err := applier.ApplyTerminalSequences(colors, activeScheme.Name); err != nil {
				errors = append(errors, fmt.Sprintf("terminal: %v", err))
				logger.Error("Failed to apply terminal sequences", "error", err)
			} else {
				logger.Info("Applied terminal sequences")
			}
		} else {
			if err := applier.ApplyTheme(app, colors, activeScheme.Mode); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", app, err))
				logger.Error("Failed to apply theme", "app", app, "error", err)
			} else {
				logger.Info("Applied theme", "app", app)
				if app == "kitty" {
					kittyThemed = true
				}
			}
		}
	}

	// Reload kitty instances if kitty was themed
	if kittyThemed {
		if err := theme.ReloadKittyInstances(); err != nil {
			logger.Error("Failed to reload kitty instances", "error", err)
		}
	}

	if len(errors) > 0 {
		logger.Warn("Some applications failed to theme", "errors", strings.Join(errors, ", "))
	}

	logger.Info("Material You scheme generated and applied to all applications")

	return nil
}

// convertMaterialColors converts Material You colors to complete Heimdall format (122 colors)
func convertMaterialColors(ms *material.Scheme) map[string]string {
	// Use the new generator to create a full scheme
	generator := generator.NewWallpaperGenerator()

	// Determine if dark mode based on background luminance
	isDark := isColorDark(argbToHex(ms.Background))
	mode := "dark"
	if !isDark {
		mode = "light"
	}

	// Generate the full scheme
	fullScheme, err := generator.GenerateFullScheme(ms, "", mode)
	if err != nil {
		// Fallback to basic conversion if generation fails
		logger.Error("Failed to generate full scheme, using basic conversion", "error", err)
		return convertMaterialColorsBasic(ms)
	}

	logger.Info("Successfully generated full scheme", "colorCount", len(fullScheme.Colours))

	// Convert to format without # prefix for compatibility
	colors := make(map[string]string)
	for name, hex := range fullScheme.Colours {
		colors[name] = strings.TrimPrefix(hex, "#")
	}

	return colors
}

// convertMaterialColorsBasic is the fallback basic conversion
func convertMaterialColorsBasic(ms *material.Scheme) map[string]string {
	colors := make(map[string]string)

	// Map Material You colors to our color names
	colorMap := map[string]uint32{
		"base":     ms.Background,
		"surface":  ms.Surface,
		"overlay":  ms.SurfaceVariant,
		"text":     ms.OnBackground,
		"subtext0": ms.OnSurfaceVariant,
		"subtext1": ms.OnSurface,

		"primary":      ms.Primary,
		"on_primary":   ms.OnPrimary,
		"secondary":    ms.Secondary,
		"on_secondary": ms.OnSecondary,
		"tertiary":     ms.Tertiary,
		"on_tertiary":  ms.OnTertiary,

		"red":    ms.Error,
		"on_red": ms.OnError,

		"surface0": ms.Surface,
		"surface1": ms.SurfaceVariant,
		"surface2": ms.PrimaryContainer,

		"overlay0": ms.Outline,
		"overlay1": ms.OutlineVariant,
		"overlay2": ms.Shadow,
	}

	for name, argb := range colorMap {
		r := uint8((argb >> 16) & 0xFF)
		g := uint8((argb >> 8) & 0xFF)
		b := uint8(argb & 0xFF)
		colors[name] = fmt.Sprintf("%02x%02x%02x", r, g, b)
	}

	return colors
}

// Helper functions for color conversion
func argbToHex(argb uint32) string {
	r := (argb >> 16) & 0xFF
	g := (argb >> 8) & 0xFF
	b := argb & 0xFF
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func isColorDark(hex string) bool {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return true // Default to dark
	}

	var rgb uint32
	fmt.Sscanf(hex, "%06x", &rgb)
	r := float64((rgb >> 16) & 0xFF)
	g := float64((rgb >> 8) & 0xFF)
	b := float64(rgb & 0xFF)

	// Calculate perceived luminance
	luminance := 0.299*r + 0.587*g + 0.114*b
	return luminance < 128
}
