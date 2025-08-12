package screenshot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/heimdall-cli/heimdall/internal/config"
	"github.com/heimdall-cli/heimdall/internal/utils/logger"
	"github.com/heimdall-cli/heimdall/internal/utils/notify"
	"github.com/heimdall-cli/heimdall/internal/utils/paths"
	"github.com/spf13/cobra"
)

var (
	region string
	freeze bool
)

// NewCommand creates the screenshot command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screenshot",
		Short: "Take a screenshot",
		Long:  `Take a screenshot of the entire screen or a selected region.`,
		RunE:  runScreenshot,
	}

	cmd.Flags().StringVarP(&region, "region", "r", "", "Take a screenshot of a region (use 'slurp' or provide geometry)")
	cmd.Flags().BoolVarP(&freeze, "freeze", "f", false, "Freeze the screen while selecting a region")

	return cmd
}

func runScreenshot(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	external := cfg.External
	screenshotCfg := cfg.Screenshot

	// Use configured directory or fallback to default
	screenshotDir := screenshotCfg.Directory
	if screenshotDir == "" {
		screenshotDir = paths.ScreenshotsDir
	}

	// Ensure screenshots directory exists
	if err := paths.EnsureDir(screenshotDir); err != nil {
		return fmt.Errorf("failed to create screenshots directory: %w", err)
	}

	// Generate filename with configured pattern
	filePattern := screenshotCfg.FileNamePattern
	if filePattern == "" {
		filePattern = "screenshot_%Y%m%d_%H%M%S"
	}
	// Replace timestamp patterns
	filename := strings.ReplaceAll(filePattern, "%Y%m%d", time.Now().Format("20060102"))
	filename = strings.ReplaceAll(filename, "%H%M%S", time.Now().Format("150405"))
	filename = fmt.Sprintf("%s.%s", filename, screenshotCfg.FileFormat)
	outputPath := filepath.Join(screenshotDir, filename)

	// Check if grim is available
	grimPath := external.Grim
	if grimPath == "" {
		grimPath = "grim"
	}
	if _, err := exec.LookPath(grimPath); err != nil {
		return fmt.Errorf("grim not found: %w", err)
	}

	// Build grim command
	grimArgs := []string{}

	// Handle region selection
	if region != "" {
		if region == "slurp" || region == "" {
			// Use slurp for region selection
			slurpPath := external.Slurp
			if slurpPath == "" {
				slurpPath = "slurp"
			}

			if _, err := exec.LookPath(slurpPath); err != nil {
				return fmt.Errorf("slurp not found: %w", err)
			}

			// Handle freeze option
			if freeze {
				// Take a temporary screenshot first
				cacheDir := filepath.Join(paths.HeimdallCacheDir, "screenshots")
				if err := paths.EnsureDir(cacheDir); err != nil {
					return fmt.Errorf("failed to create cache directory: %w", err)
				}
				tempFile := filepath.Join(cacheDir, screenshotCfg.FreezeFileName)

				// Capture current screen
				freezeCmd := exec.Command(grimPath, tempFile)
				if err := freezeCmd.Run(); err != nil {
					return fmt.Errorf("failed to capture freeze frame: %w", err)
				}
				defer os.Remove(tempFile)

				// TODO: Display frozen image while selecting
				// This would require a more complex implementation with a viewer
				logger.Warn("Freeze option not fully implemented yet")
			}

			// Run slurp to get region
			slurpCmd := exec.Command(slurpPath)
			output, err := slurpCmd.Output()
			if err != nil {
				// User cancelled selection
				logger.Info("Screenshot cancelled")
				return nil
			}

			region = strings.TrimSpace(string(output))
			if region == "" {
				logger.Info("No region selected")
				return nil
			}
		}

		// Add region to grim arguments
		grimArgs = append(grimArgs, "-g", region)
	}

	// Add output path
	grimArgs = append(grimArgs, outputPath)

	// Take screenshot
	logger.Debug("Taking screenshot", "command", grimPath, "args", grimArgs)
	grimCmd := exec.Command(grimPath, grimArgs...)
	if err := grimCmd.Run(); err != nil {
		return fmt.Errorf("failed to take screenshot: %w", err)
	}

	logger.Info("Screenshot saved", "path", outputPath)

	// Copy to clipboard if configured
	if screenshotCfg.CopyToClipboard {
		if err := copyToClipboard(outputPath, external); err != nil {
			logger.Warn("Failed to copy to clipboard", "error", err)
		}
	}

	// Send notification if configured
	if screenshotCfg.ShowNotification && notify.IsAvailable() {
		notif := &notify.Notification{
			Summary: "Screenshot captured",
			Body:    fmt.Sprintf("Saved to %s", filename),
			Icon:    outputPath,
			Timeout: screenshotCfg.GetNotificationTimeout(),
		}

		if err := notify.NewNotifier().Send(notif); err != nil {
			logger.Warn("Failed to send notification", "error", err)
		}
	}

	// Open with swappy if configured and available
	if screenshotCfg.OpenWithSwappy {
		swappyPath := external.Swappy
		if swappyPath == "" {
			swappyPath = "swappy"
		}

		if _, err := exec.LookPath(swappyPath); err == nil {
			// Launch swappy in background
			swappyCmd := exec.Command(swappyPath, "-f", outputPath)
			if err := swappyCmd.Start(); err != nil {
				logger.Warn("Failed to open with swappy", "error", err)
			}
		}
	}

	return nil
}

func copyToClipboard(imagePath string, external config.ExternalTools) error {
	// Try wl-copy first
	wlCopyPath := external.WlClipboard
	if wlCopyPath == "" {
		wlCopyPath = "wl-copy"
	}

	if _, err := exec.LookPath(wlCopyPath); err == nil {
		cmd := exec.Command(wlCopyPath, "-t", "image/png")

		// Open image file
		file, err := os.Open(imagePath)
		if err != nil {
			return err
		}
		defer file.Close()

		cmd.Stdin = file
		return cmd.Run()
	}

	// Try xclip as fallback
	xclipPath := external.Xclip
	if xclipPath == "" {
		xclipPath = "xclip"
	}

	if _, err := exec.LookPath(xclipPath); err == nil {
		cmd := exec.Command(xclipPath, "-selection", "clipboard", "-t", "image/png", "-i", imagePath)
		return cmd.Run()
	}

	return fmt.Errorf("no clipboard tool available")
}
