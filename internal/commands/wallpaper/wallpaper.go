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

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/scheme"
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

	// Create caelestia-compatible JSON output
	output := map[string]interface{}{
		"name":    "dynamic",
		"flavour": "default",
		"mode":    mode,
		"variant": variant,
		"colours": convertMaterialColorsToJSON(materialScheme),
	}

	// Output JSON
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// convertMaterialColorsToJSON converts Material You colors to caelestia JSON format
func convertMaterialColorsToJSON(ms *material.Scheme) map[string]string {
	colors := make(map[string]string)

	// Map Material You colors to caelestia color names
	colorMap := map[string]uint32{
		"primary":              ms.Primary,
		"onPrimary":            ms.OnPrimary,
		"primaryContainer":     ms.PrimaryContainer,
		"onPrimaryContainer":   ms.OnPrimaryContainer,
		"secondary":            ms.Secondary,
		"onSecondary":          ms.OnSecondary,
		"secondaryContainer":   ms.SecondaryContainer,
		"onSecondaryContainer": ms.OnSecondaryContainer,
		"tertiary":             ms.Tertiary,
		"onTertiary":           ms.OnTertiary,
		"tertiaryContainer":    ms.TertiaryContainer,
		"onTertiaryContainer":  ms.OnTertiaryContainer,
		"error":                ms.Error,
		"onError":              ms.OnError,
		"errorContainer":       ms.ErrorContainer,
		"onErrorContainer":     ms.OnErrorContainer,
		"background":           ms.Background,
		"onBackground":         ms.OnBackground,
		"surface":              ms.Surface,
		"onSurface":            ms.OnSurface,
		"surfaceVariant":       ms.SurfaceVariant,
		"onSurfaceVariant":     ms.OnSurfaceVariant,
		"outline":              ms.Outline,
		"outlineVariant":       ms.OutlineVariant,
		"shadow":               ms.Shadow,
		"scrim":                ms.Scrim,
		"inverseSurface":       ms.InverseSurface,
		"inverseOnSurface":     ms.InverseOnSurface,
		"inversePrimary":       ms.InversePrimary,
	}

	for name, argb := range colorMap {
		r := uint8((argb >> 16) & 0xFF)
		g := uint8((argb >> 8) & 0xFF)
		b := uint8(argb & 0xFF)
		colors[name] = fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}

	return colors
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

// generateMaterialYouScheme generates a Material You scheme from the wallpaper
func generateMaterialYouScheme(wallpaperPath string) error {
	logger.Info("Generating Material You scheme from wallpaper")

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

	// Determine mode based on wallpaper
	analyzer := wallpaper.NewAnalyzer()
	mode, err := analyzer.DetermineMode(wallpaperPath)
	if err != nil {
		mode = "dark" // Default to dark
	}

	// Create scheme
	materialScheme, err := generator.GenerateScheme(palette.Seed, mode == "dark")
	if err != nil {
		return fmt.Errorf("failed to generate scheme: %w", err)
	}

	// Convert to our scheme format
	newScheme := &scheme.Scheme{
		Name:    "material-you",
		Flavour: "generated",
		Mode:    mode,
		Variant: "wallpaper",
		Colours: convertMaterialColors(materialScheme),
	}

	// Save and set the scheme
	manager := scheme.NewManager()
	if err := manager.SaveScheme(newScheme); err != nil {
		return fmt.Errorf("failed to save scheme: %w", err)
	}

	if err := manager.SetScheme(newScheme); err != nil {
		return fmt.Errorf("failed to set scheme: %w", err)
	}

	logger.Info("Material You scheme generated and applied")

	return nil
}

// convertMaterialColors converts Material You colors to our format
func convertMaterialColors(ms *material.Scheme) map[string]string {
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
