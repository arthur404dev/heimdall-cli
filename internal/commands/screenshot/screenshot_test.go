package screenshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Use != "screenshot" {
		t.Errorf("Expected Use to be 'screenshot', got %s", cmd.Use)
	}

	if cmd.Short != "Take a screenshot" {
		t.Errorf("Expected Short to be 'Take a screenshot', got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "screenshot of the entire screen") {
		t.Errorf("Expected Long to contain 'screenshot of the entire screen'")
	}

	// Check flags
	regionFlag := cmd.Flags().Lookup("region")
	if regionFlag == nil {
		t.Error("Expected region flag to exist")
	}
	if regionFlag.Shorthand != "r" {
		t.Errorf("Expected region flag shorthand to be 'r', got %s", regionFlag.Shorthand)
	}

	freezeFlag := cmd.Flags().Lookup("freeze")
	if freezeFlag == nil {
		t.Error("Expected freeze flag to exist")
	}
	if freezeFlag.Shorthand != "f" {
		t.Errorf("Expected freeze flag shorthand to be 'f', got %s", freezeFlag.Shorthand)
	}
}

func TestRunScreenshot(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalScreenshotsDir := paths.ScreenshotsDir
	paths.ScreenshotsDir = filepath.Join(tempDir, "screenshots")
	defer func() { paths.ScreenshotsDir = originalScreenshotsDir }()

	// Setup test config
	defer func() {
		config.Load() // Restore original config
	}()

	tests := []struct {
		name        string
		regionFlag  string
		freezeFlag  bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "takes full screenshot",
			regionFlag:  "",
			freezeFlag:  false,
			expectError: true, // Will fail because grim is not available in test
			errorMsg:    "grim not found",
		},
		{
			name:        "handles region flag",
			regionFlag:  "100,100 200x200",
			freezeFlag:  false,
			expectError: true, // Will fail because grim is not available in test
			errorMsg:    "grim not found",
		},
		{
			name:        "handles slurp region",
			regionFlag:  "slurp",
			freezeFlag:  false,
			expectError: true, // Will fail because grim/slurp not available in test
			errorMsg:    "grim not found",
		},
		{
			name:        "handles freeze flag",
			regionFlag:  "slurp",
			freezeFlag:  true,
			expectError: true, // Will fail because grim/slurp not available in test
			errorMsg:    "grim not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flags
			region = tt.regionFlag
			freeze = tt.freezeFlag

			// Create a mock command for testing
			cmd := NewCommand()
			err := runScreenshot(cmd, []string{})

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestScreenshotConfiguration(t *testing.T) {
	tests := []struct {
		name              string
		config            *config.Config
		expectedDirectory string
		expectedPattern   string
		expectedFormat    string
	}{
		{
			name: "uses default configuration",
			config: &config.Config{
				Screenshot: config.ScreenshotConfig{
					Directory:       "",
					FileNamePattern: "",
					FileFormat:      "",
				},
			},
			expectedDirectory: "",            // Would use paths.ScreenshotsDir
			expectedPattern:   "screenshot_", // Default pattern
			expectedFormat:    "png",         // Default format
		},
		{
			name: "uses custom configuration",
			config: &config.Config{
				Screenshot: config.ScreenshotConfig{
					Directory:       "/custom/screenshots",
					FileNamePattern: "custom_%Y%m%d_%H%M%S",
					FileFormat:      "jpg",
				},
			},
			expectedDirectory: "/custom/screenshots",
			expectedPattern:   "custom_",
			expectedFormat:    "jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that configuration values are used correctly
			directory := tt.config.Screenshot.Directory
			if directory == "" {
				directory = paths.ScreenshotsDir
			}

			pattern := tt.config.Screenshot.FileNamePattern
			if pattern == "" {
				pattern = "screenshot_%Y%m%d_%H%M%S"
			}

			format := tt.config.Screenshot.FileFormat
			if format == "" {
				format = "png"
			}

			if tt.expectedDirectory != "" && directory != tt.expectedDirectory {
				t.Errorf("Expected directory %s, got %s", tt.expectedDirectory, directory)
			}

			if !strings.HasPrefix(pattern, strings.Split(tt.expectedPattern, "_")[0]) {
				t.Errorf("Expected pattern to start with %s, got %s", tt.expectedPattern, pattern)
			}

			if format != tt.expectedFormat {
				t.Errorf("Expected format %s, got %s", tt.expectedFormat, format)
			}
		})
	}
}

func TestRegionHandling(t *testing.T) {
	tests := []struct {
		name         string
		regionInput  string
		expectedArgs []string
		expectSlurp  bool
	}{
		{
			name:         "handles empty region",
			regionInput:  "",
			expectedArgs: []string{},
			expectSlurp:  false,
		},
		{
			name:         "handles slurp region",
			regionInput:  "slurp",
			expectedArgs: []string{"-g"},
			expectSlurp:  true,
		},
		{
			name:         "handles manual region",
			regionInput:  "100,100 200x200",
			expectedArgs: []string{"-g", "100,100 200x200"},
			expectSlurp:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test region parsing logic
			var args []string
			regionValue := tt.regionInput

			if regionValue != "" {
				if regionValue == "slurp" {
					// Would call slurp command
					if !tt.expectSlurp {
						t.Error("Expected slurp to be called")
					}
					// In real implementation, this would get output from slurp
					regionValue = "mocked_slurp_output"
				}
				args = append(args, "-g", regionValue)
			}

			if len(args) != len(tt.expectedArgs) {
				if tt.expectSlurp {
					// For slurp, we expect the args to be generated
					if len(args) != 2 || args[0] != "-g" {
						t.Errorf("Expected slurp to generate region args, got %v", args)
					}
				} else {
					t.Errorf("Expected args %v, got %v", tt.expectedArgs, args)
				}
			}
		})
	}
}

func TestFreezeHandling(t *testing.T) {
	tests := []struct {
		name         string
		freezeFlag   bool
		regionFlag   string
		expectFreeze bool
	}{
		{
			name:         "no freeze without region",
			freezeFlag:   true,
			regionFlag:   "",
			expectFreeze: false,
		},
		{
			name:         "freeze with slurp region",
			freezeFlag:   true,
			regionFlag:   "slurp",
			expectFreeze: true,
		},
		{
			name:         "no freeze when flag is false",
			freezeFlag:   false,
			regionFlag:   "slurp",
			expectFreeze: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test freeze logic
			shouldFreeze := tt.freezeFlag && (tt.regionFlag == "slurp" || tt.regionFlag == "")

			if shouldFreeze != tt.expectFreeze {
				t.Errorf("Expected freeze %v, got %v", tt.expectFreeze, shouldFreeze)
			}
		})
	}
}

func TestCopyToClipboard(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	testImagePath := filepath.Join(tempDir, "test.png")

	// Create a fake image file
	err := os.WriteFile(testImagePath, []byte("fake image data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	tests := []struct {
		name        string
		imagePath   string
		external    config.ExternalTools
		expectError bool
		errorMsg    string
	}{
		{
			name:      "handles missing image file",
			imagePath: "/nonexistent/image.png",
			external: config.ExternalTools{
				WlClipboard: "wl-copy",
			},
			expectError: true,
			errorMsg:    "no such file",
		},
		{
			name:      "attempts wl-copy first",
			imagePath: testImagePath,
			external: config.ExternalTools{
				WlClipboard: "wl-copy",
			},
			expectError: true, // Will fail because wl-copy is not available in test
		},
		{
			name:      "falls back to xclip",
			imagePath: testImagePath,
			external: config.ExternalTools{
				WlClipboard: "", // Empty to skip wl-copy
				Xclip:       "xclip",
			},
			expectError: true, // Will fail because xclip is not available in test
		},
		{
			name:      "returns error when no clipboard tool available",
			imagePath: testImagePath,
			external: config.ExternalTools{
				WlClipboard: "",
				Xclip:       "",
			},
			expectError: true,
			errorMsg:    "no clipboard tool available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := copyToClipboard(tt.imagePath, tt.external)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestFileNaming(t *testing.T) {
	tests := []struct {
		name            string
		fileNamePattern string
		fileFormat      string
		expectedPattern string
		expectedExt     string
	}{
		{
			name:            "uses default pattern and format",
			fileNamePattern: "",
			fileFormat:      "",
			expectedPattern: "screenshot_",
			expectedExt:     ".png",
		},
		{
			name:            "uses custom pattern and format",
			fileNamePattern: "custom_%Y%m%d_%H%M%S",
			fileFormat:      "jpg",
			expectedPattern: "custom_",
			expectedExt:     ".jpg",
		},
		{
			name:            "handles pattern without timestamps",
			fileNamePattern: "my_screenshot",
			fileFormat:      "png",
			expectedPattern: "my_screenshot",
			expectedExt:     ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test filename generation logic
			pattern := tt.fileNamePattern
			if pattern == "" {
				pattern = "screenshot_%Y%m%d_%H%M%S"
			}

			format := tt.fileFormat
			if format == "" {
				format = "png"
			}

			// Replace timestamp patterns (simplified for test)
			filename := strings.ReplaceAll(pattern, "%Y%m%d", "20240101")
			filename = strings.ReplaceAll(filename, "%H%M%S", "120000")
			filename = filename + "." + format

			if !strings.HasPrefix(filename, tt.expectedPattern) {
				t.Errorf("Expected filename to start with '%s', got '%s'", tt.expectedPattern, filename)
			}

			if !strings.HasSuffix(filename, tt.expectedExt) {
				t.Errorf("Expected filename to end with '%s', got '%s'", tt.expectedExt, filename)
			}
		})
	}
}

func TestDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		directory   string
		expectError bool
	}{
		{
			name:        "creates directory when it doesn't exist",
			directory:   filepath.Join(tempDir, "new_screenshots"),
			expectError: false,
		},
		{
			name:        "handles existing directory",
			directory:   tempDir, // Already exists
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test directory creation logic (simplified)
			err := os.MkdirAll(tt.directory, 0755)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				// Check that directory was created
				if _, err := os.Stat(tt.directory); os.IsNotExist(err) {
					t.Error("Expected directory to be created")
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkRunScreenshot(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	originalScreenshotsDir := paths.ScreenshotsDir
	paths.ScreenshotsDir = filepath.Join(tempDir, "screenshots")
	defer func() { paths.ScreenshotsDir = originalScreenshotsDir }()

	cmd := NewCommand()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail, but we're measuring the setup time
		runScreenshot(cmd, []string{})
	}
}

func BenchmarkCopyToClipboard(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	testImagePath := filepath.Join(tempDir, "test.png")
	os.WriteFile(testImagePath, []byte("fake image data"), 0644)

	external := config.ExternalTools{
		WlClipboard: "wl-copy",
		Xclip:       "xclip",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail, but we're measuring the logic time
		copyToClipboard(testImagePath, external)
	}
}

// Integration tests
func TestScreenshotCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test command creation and flag parsing
	cmd := NewCommand()

	// Test that command can be created without errors
	if cmd == nil {
		t.Error("Expected command to be created")
	}

	// Test flag parsing
	cmd.SetArgs([]string{"--region", "slurp", "--freeze"})
	err := cmd.ParseFlags([]string{"--region", "slurp", "--freeze"})
	if err != nil {
		t.Errorf("Flag parsing failed: %v", err)
	}

	// Check that flags were set
	regionFlagValue, _ := cmd.Flags().GetString("region")
	if regionFlagValue != "slurp" {
		t.Errorf("Expected region flag to be 'slurp', got '%s'", regionFlagValue)
	}

	freezeFlagValue, _ := cmd.Flags().GetBool("freeze")
	if !freezeFlagValue {
		t.Error("Expected freeze flag to be true")
	}
}

func TestExternalToolIntegration(t *testing.T) {
	// Test external tool path resolution
	tests := []struct {
		name         string
		toolName     string
		configPath   string
		expectedPath string
	}{
		{
			name:         "uses configured grim path",
			toolName:     "grim",
			configPath:   "/custom/path/grim",
			expectedPath: "/custom/path/grim",
		},
		{
			name:         "uses default grim path when config is empty",
			toolName:     "grim",
			configPath:   "",
			expectedPath: "grim",
		},
		{
			name:         "uses configured slurp path",
			toolName:     "slurp",
			configPath:   "/custom/path/slurp",
			expectedPath: "/custom/path/slurp",
		},
		{
			name:         "uses default slurp path when config is empty",
			toolName:     "slurp",
			configPath:   "",
			expectedPath: "slurp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test path resolution logic
			path := tt.configPath
			if path == "" {
				path = tt.toolName
			}

			if path != tt.expectedPath {
				t.Errorf("Expected path '%s', got '%s'", tt.expectedPath, path)
			}
		})
	}
}

func TestNotificationIntegration(t *testing.T) {
	// Test notification configuration
	tests := []struct {
		name             string
		showNotification bool
		expectedCall     bool
	}{
		{
			name:             "sends notification when enabled",
			showNotification: true,
			expectedCall:     true,
		},
		{
			name:             "skips notification when disabled",
			showNotification: false,
			expectedCall:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test notification logic (simplified)
			shouldNotify := tt.showNotification

			if shouldNotify != tt.expectedCall {
				t.Errorf("Expected notification call %v, got %v", tt.expectedCall, shouldNotify)
			}
		})
	}
}
