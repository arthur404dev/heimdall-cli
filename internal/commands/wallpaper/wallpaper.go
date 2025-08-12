package wallpaper

import (
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
	"github.com/arthur404dev/heimdall-cli/internal/utils/color"
	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
	"github.com/arthur404dev/heimdall-cli/internal/utils/material"
	"github.com/arthur404dev/heimdall-cli/internal/utils/notify"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
	"github.com/arthur404dev/heimdall-cli/internal/utils/wallpaper"
	"github.com/spf13/cobra"
)

// Command creates the wallpaper command
func Command() *cobra.Command {
	var (
		random         bool
		filter         bool
		threshold      float64
		generateScheme bool
		info           bool
	)

	cmd := &cobra.Command{
		Use:   "wallpaper [path]",
		Short: "Manage wallpapers",
		Long: `Manage wallpapers with smart filtering and Material You integration.
		
Features:
  - Set specific wallpaper
  - Random wallpaper selection
  - Colourfulness filtering
  - Material You scheme generation
  - Wallpaper info analysis`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			if err := config.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			cfg := config.Get()

			// Handle wallpaper info
			if info && len(args) > 0 {
				return showWallpaperInfo(args[0])
			}

			// Handle random wallpaper
			if random || len(args) == 0 {
				return setRandomWallpaper(cfg, filter, threshold, generateScheme)
			}

			// Set specific wallpaper
			wallpaperPath := args[0]
			return setWallpaper(wallpaperPath, generateScheme)
		},
	}

	cmd.Flags().BoolVarP(&random, "random", "r", false, "Select random wallpaper")
	cmd.Flags().BoolVarP(&filter, "filter", "f", false, "Filter by colourfulness")
	cmd.Flags().Float64VarP(&threshold, "threshold", "t", 50.0, "Colourfulness threshold")
	cmd.Flags().BoolVarP(&generateScheme, "scheme", "s", false, "Generate Material You scheme")
	cmd.Flags().BoolVarP(&info, "info", "i", false, "Show wallpaper info")

	return cmd
}

// setRandomWallpaper selects and sets a random wallpaper
func setRandomWallpaper(cfg *config.Config, filter bool, threshold float64, generateScheme bool) error {
	wallpaperDir := cfg.Wallpaper.Directory
	if wallpaperDir == "" {
		wallpaperDir = paths.WallpapersDir
	}

	// Expand home directory
	if strings.HasPrefix(wallpaperDir, "~/") {
		home, _ := os.UserHomeDir()
		wallpaperDir = filepath.Join(home, wallpaperDir[2:])
	}

	// Find all image files
	var wallpapers []string
	extensions := []string{".jpg", ".jpeg", ".png", ".webp"}

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

	// Filter by colourfulness if requested
	if filter || cfg.Wallpaper.Filter {
		if threshold == 50.0 && cfg.Wallpaper.Threshold > 0 {
			threshold = cfg.Wallpaper.Threshold
		}

		analyzer := wallpaper.NewAnalyzer()
		var filtered []string

		for _, wp := range wallpapers {
			colourfulness, err := analyzer.AnalyzeColourfulness(wp)
			if err != nil {
				logger.Error("Failed to analyze wallpaper", "path", wp, "error", err)
				continue
			}

			if colourfulness >= threshold {
				filtered = append(filtered, wp)
			}
		}

		if len(filtered) == 0 {
			logger.Warn("No wallpapers passed filter, using all", "threshold", threshold)
		} else {
			wallpapers = filtered
			logger.Info("Filtered wallpapers", "total", len(wallpapers), "threshold", threshold)
		}
	}

	// Select random wallpaper
	rand.Seed(time.Now().UnixNano())
	selected := wallpapers[rand.Intn(len(wallpapers))]

	logger.Info("Selected wallpaper", "path", selected)

	return setWallpaper(selected, generateScheme || cfg.Wallpaper.SmartMode)
}

// setWallpaper sets a specific wallpaper
func setWallpaper(wallpaperPath string, generateScheme bool) error {
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

	// Generate Material You scheme if requested
	if generateScheme {
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
		Metadata: scheme.SchemeMetadata{
			Generated:   true,
			Source:      wallpaperPath,
			Description: "Generated from wallpaper",
		},
		Colors: convertMaterialColors(materialScheme),
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
func convertMaterialColors(ms *material.Scheme) map[string]*color.Color {
	colors := make(map[string]*color.Color)

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
		colors[name] = color.NewFromRGB(r, g, b)
	}

	return colors
}
