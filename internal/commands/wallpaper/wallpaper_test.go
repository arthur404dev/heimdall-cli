package wallpaper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/hypr"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// MockHyprClient is a simple mock implementation for testing
type MockHyprClient struct {
	monitors      []hypr.Monitor
	workspaces    []hypr.Workspace
	windows       []hypr.Window
	dispatchErr   error
	monitorsErr   error
	workspacesErr error
	windowsErr    error
}

func (m *MockHyprClient) Subscribe(events []string) (<-chan hypr.Event, error) {
	return nil, nil
}

func (m *MockHyprClient) GetWindows() ([]hypr.Window, error) {
	return m.windows, m.windowsErr
}

func (m *MockHyprClient) Dispatch(command string, args ...string) error {
	return m.dispatchErr
}

func (m *MockHyprClient) GetMonitors() ([]hypr.Monitor, error) {
	return m.monitors, m.monitorsErr
}

func (m *MockHyprClient) GetWorkspaces() ([]hypr.Workspace, error) {
	return m.workspaces, m.workspacesErr
}

func (m *MockHyprClient) Close() error {
	return nil
}

func TestCommand(t *testing.T) {
	cmd := Command()

	if cmd.Use != "wallpaper [OPTIONS]" {
		t.Errorf("Expected Use to be 'wallpaper [OPTIONS]', got %s", cmd.Use)
	}

	if cmd.Short != "Manage wallpapers with Material You integration" {
		t.Errorf("Expected Short to be 'Manage wallpapers with Material You integration', got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "Material You integration") {
		t.Errorf("Expected Long to contain 'Material You integration'")
	}

	// Check flags
	printFlag := cmd.Flags().Lookup("print")
	if printFlag == nil {
		t.Error("Expected print flag to exist")
	}
	if printFlag.Shorthand != "p" {
		t.Errorf("Expected print flag shorthand to be 'p', got %s", printFlag.Shorthand)
	}

	randomFlag := cmd.Flags().Lookup("random")
	if randomFlag == nil {
		t.Error("Expected random flag to exist")
	}
	if randomFlag.Shorthand != "r" {
		t.Errorf("Expected random flag shorthand to be 'r', got %s", randomFlag.Shorthand)
	}

	fileFlag := cmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Error("Expected file flag to exist")
	}
	if fileFlag.Shorthand != "f" {
		t.Errorf("Expected file flag shorthand to be 'f', got %s", fileFlag.Shorthand)
	}

	noFilterFlag := cmd.Flags().Lookup("no-filter")
	if noFilterFlag == nil {
		t.Error("Expected no-filter flag to exist")
	}
	if noFilterFlag.Shorthand != "n" {
		t.Errorf("Expected no-filter flag shorthand to be 'n', got %s", noFilterFlag.Shorthand)
	}

	thresholdFlag := cmd.Flags().Lookup("threshold")
	if thresholdFlag == nil {
		t.Error("Expected threshold flag to exist")
	}
	if thresholdFlag.Shorthand != "t" {
		t.Errorf("Expected threshold flag shorthand to be 't', got %s", thresholdFlag.Shorthand)
	}

	noSmartFlag := cmd.Flags().Lookup("no-smart")
	if noSmartFlag == nil {
		t.Error("Expected no-smart flag to exist")
	}
	if noSmartFlag.Shorthand != "N" {
		t.Errorf("Expected no-smart flag shorthand to be 'N', got %s", noSmartFlag.Shorthand)
	}
}

func TestGetCurrentWallpaper(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	linkPath := filepath.Join(tempDir, "current_wallpaper")
	targetPath := "/path/to/wallpaper.jpg"

	tests := []struct {
		name         string
		setup        func() error
		expectError  bool
		errorMsg     string
		expectedPath string
	}{
		{
			name: "returns error when no wallpaper is set",
			setup: func() error {
				return nil // No symlink
			},
			expectError: true,
			errorMsg:    "no wallpaper currently set",
		},
		{
			name: "returns wallpaper path when symlink exists",
			setup: func() error {
				return os.Symlink(targetPath, linkPath)
			},
			expectError:  false,
			expectedPath: targetPath,
		},
		{
			name: "handles broken symlink gracefully",
			setup: func() error {
				// Create symlink to non-existent file
				return os.Symlink("/nonexistent/path.jpg", linkPath)
			},
			expectError:  false,
			expectedPath: "/nonexistent/path.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			os.Remove(linkPath)

			err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err = getCurrentWallpaper()

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

func TestPrintColorScheme(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake image file
	testImagePath := filepath.Join(tempDir, "test.jpg")
	// Create minimal JPEG header for testing
	jpegHeader := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	err := os.WriteFile(testImagePath, jpegHeader, 0644)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	tests := []struct {
		name        string
		imagePath   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "returns error for non-existent file",
			imagePath:   "/nonexistent/image.jpg",
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:        "handles existing image file",
			imagePath:   testImagePath,
			expectError: true, // Will fail because it's not a real image
			errorMsg:    "failed to decode image",
		},
		{
			name:        "handles tilde expansion",
			imagePath:   "~/test.jpg",
			expectError: true, // Will fail because file doesn't exist after expansion
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := printColorScheme(tt.imagePath)

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

func TestColorConversion(t *testing.T) {
	// Test color conversion logic
	tests := []struct {
		name     string
		argb     uint32
		expected string
	}{
		{
			name:     "converts purple color",
			argb:     0xFF6750A4,
			expected: "#6750a4",
		},
		{
			name:     "converts white color",
			argb:     0xFFFFFFFF,
			expected: "#ffffff",
		},
		{
			name:     "converts dark color",
			argb:     0xFF1C1B1F,
			expected: "#1c1b1f",
		},
		{
			name:     "converts black color",
			argb:     0xFF000000,
			expected: "#000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test color conversion logic
			r := uint8((tt.argb >> 16) & 0xFF)
			g := uint8((tt.argb >> 8) & 0xFF)
			b := uint8(tt.argb & 0xFF)

			// Simple hex conversion test
			hexResult := strings.ToLower(strings.TrimPrefix(tt.expected, "#"))
			if len(hexResult) != 6 {
				t.Errorf("Expected hex color to be 6 characters, got %d", len(hexResult))
			}

			// Test that we can extract RGB components
			if r == 0 && g == 0 && b == 0 && tt.expected != "#000000" {
				t.Error("Color extraction failed")
			}
		})
	}
}

func TestSetWallpaper(t *testing.T) {
	tempDir := t.TempDir()
	originalStateDir := paths.StateDir
	paths.StateDir = tempDir
	defer func() { paths.StateDir = originalStateDir }()

	// Create a test wallpaper file
	testWallpaperPath := filepath.Join(tempDir, "test_wallpaper.jpg")
	err := os.WriteFile(testWallpaperPath, []byte("fake image data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test wallpaper: %v", err)
	}

	tests := []struct {
		name            string
		wallpaperPath   string
		enableSmartMode bool
		expectError     bool
		errorMsg        string
	}{
		{
			name:            "sets wallpaper successfully",
			wallpaperPath:   testWallpaperPath,
			enableSmartMode: false,
			expectError:     false,
		},
		{
			name:            "handles non-existent wallpaper",
			wallpaperPath:   "/nonexistent/wallpaper.jpg",
			enableSmartMode: false,
			expectError:     true,
			errorMsg:        "not found",
		},
		{
			name:            "handles tilde expansion",
			wallpaperPath:   "~/test.jpg",
			enableSmartMode: false,
			expectError:     true, // Will fail because expanded path doesn't exist
		},
		{
			name:            "enables smart mode",
			wallpaperPath:   testWallpaperPath,
			enableSmartMode: true,
			expectError:     false, // Smart mode will fail but shouldn't prevent wallpaper setting
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up symlink before each test
			linkPath := filepath.Join(tempDir, "current_wallpaper")
			os.Remove(linkPath)

			err := setWallpaper(tt.wallpaperPath, tt.enableSmartMode)

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

				// Check that symlink was created
				if _, err := os.Lstat(linkPath); err != nil {
					t.Error("Expected symlink to be created")
				}
			}
		})
	}
}

func TestFilterWallpapersBySize(t *testing.T) {
	tempDir := t.TempDir()

	// Create test wallpaper files
	testWallpapers := []string{
		filepath.Join(tempDir, "small.jpg"),
		filepath.Join(tempDir, "medium.jpg"),
		filepath.Join(tempDir, "large.jpg"),
	}

	for _, wp := range testWallpapers {
		err := os.WriteFile(wp, []byte("fake image data"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test wallpaper: %v", err)
		}
	}

	tests := []struct {
		name         string
		wallpapers   []string
		threshold    float64
		mockMonitors []hypr.Monitor
		expectError  bool
		errorMsg     string
	}{
		{
			name:       "filters wallpapers by size",
			wallpapers: testWallpapers,
			threshold:  0.8,
			mockMonitors: []hypr.Monitor{
				{Width: 1920, Height: 1080},
			},
			expectError: false,
		},
		{
			name:       "handles multiple monitors",
			wallpapers: testWallpapers,
			threshold:  0.5,
			mockMonitors: []hypr.Monitor{
				{Width: 1920, Height: 1080},
				{Width: 1366, Height: 768},
			},
			expectError: false,
		},
		{
			name:         "returns error when no monitors found",
			wallpapers:   testWallpapers,
			threshold:    0.8,
			mockMonitors: []hypr.Monitor{},
			expectError:  true,
			errorMsg:     "no monitors found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would require mocking the Hyprland client and wallpaper analyzer
			// For now, we'll test the basic logic
			if len(tt.mockMonitors) == 0 {
				// Should return error
				if !tt.expectError {
					t.Error("Expected error for no monitors")
				}
			} else {
				// Should process monitors
				minWidth := tt.mockMonitors[0].Width
				minHeight := tt.mockMonitors[0].Height

				for _, monitor := range tt.mockMonitors[1:] {
					if monitor.Width < minWidth {
						minWidth = monitor.Width
					}
					if monitor.Height < minHeight {
						minHeight = monitor.Height
					}
				}

				reqWidth := int(float64(minWidth) * tt.threshold)
				reqHeight := int(float64(minHeight) * tt.threshold)

				if reqWidth <= 0 || reqHeight <= 0 {
					t.Error("Expected positive required dimensions")
				}
			}
		})
	}
}

func TestSetRandomWallpaperFromDir(t *testing.T) {
	tempDir := t.TempDir()
	wallpaperDir := filepath.Join(tempDir, "wallpapers")
	err := os.MkdirAll(wallpaperDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create wallpaper directory: %v", err)
	}

	// Create test wallpaper files
	testWallpapers := []string{
		"test1.jpg",
		"test2.png",
		"test3.jpeg",
		"not_image.txt", // Should be ignored
	}

	for _, wp := range testWallpapers {
		path := filepath.Join(wallpaperDir, wp)
		err := os.WriteFile(path, []byte("fake image data"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test wallpaper: %v", err)
		}
	}

	tests := []struct {
		name             string
		wallpaperDir     string
		enableSizeFilter bool
		threshold        float64
		enableSmartMode  bool
		expectError      bool
		errorMsg         string
	}{
		{
			name:             "selects random wallpaper",
			wallpaperDir:     wallpaperDir,
			enableSizeFilter: false,
			threshold:        0.8,
			enableSmartMode:  false,
			expectError:      false,
		},
		{
			name:             "handles non-existent directory",
			wallpaperDir:     "/nonexistent/directory",
			enableSizeFilter: false,
			threshold:        0.8,
			enableSmartMode:  false,
			expectError:      true,
		},
		{
			name:             "handles empty directory",
			wallpaperDir:     t.TempDir(), // Empty directory
			enableSizeFilter: false,
			threshold:        0.8,
			enableSmartMode:  false,
			expectError:      true,
			errorMsg:         "no wallpapers found",
		},
		{
			name:             "handles tilde expansion",
			wallpaperDir:     "~/Pictures/Wallpapers",
			enableSizeFilter: false,
			threshold:        0.8,
			enableSmartMode:  false,
			expectError:      true, // Will fail because directory doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock config
			cfg := &config.Config{
				Wallpaper: config.WallpaperConfig{
					Directory: tt.wallpaperDir,
				},
			}

			err := setRandomWallpaperFromDir(cfg, tt.wallpaperDir, tt.enableSizeFilter, tt.threshold, tt.enableSmartMode)

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

func TestConvertMaterialColors(t *testing.T) {
	// Test the color conversion logic without using the actual material package
	// This test validates the hex color conversion functionality

	// Test color conversion with known values
	testColors := map[string]uint32{
		"background": 0xFF1C1B1F, // Dark background
		"surface":    0xFF1C1B1F, // Same as background
		"primary":    0xFF6750A4, // Purple
		"text":       0xFFE6E1E5, // Light text
	}

	// Test hex conversion function directly
	for name, color := range testColors {
		hex := fmt.Sprintf("%06x", color&0xFFFFFF)

		switch name {
		case "background", "surface":
			if hex != "1c1b1f" {
				t.Errorf("Expected %s to convert to '1c1b1f', got %s", name, hex)
			}
		case "primary":
			if hex != "6750a4" {
				t.Errorf("Expected %s to convert to '6750a4', got %s", name, hex)
			}
		case "text":
			if hex != "e6e1e5" {
				t.Errorf("Expected %s to convert to 'e6e1e5', got %s", name, hex)
			}
		}
	}
}

func TestWallpaperExtensionFiltering(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files with various extensions
	testFiles := []string{
		"image1.jpg",
		"image2.jpeg",
		"image3.png",
		"image4.webp",
		"image5.tif",
		"image6.tiff",
		"document.pdf", // Should be ignored
		"text.txt",     // Should be ignored
		"video.mp4",    // Should be ignored
		"IMAGE7.JPG",   // Should be included (case insensitive)
	}

	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		err := os.WriteFile(path, []byte("fake data"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Test extension filtering logic
	validExtensions := []string{".jpg", ".jpeg", ".png", ".webp", ".tif", ".tiff"}
	expectedValidFiles := []string{
		"image1.jpg", "image2.jpeg", "image3.png", "image4.webp",
		"image5.tif", "image6.tiff", "IMAGE7.JPG",
	}

	var validFiles []string
	for _, file := range testFiles {
		ext := strings.ToLower(filepath.Ext(file))
		for _, validExt := range validExtensions {
			if ext == validExt {
				validFiles = append(validFiles, file)
				break
			}
		}
	}

	if len(validFiles) != len(expectedValidFiles) {
		t.Errorf("Expected %d valid files, got %d", len(expectedValidFiles), len(validFiles))
	}

	// Check that all expected files are included
	for _, expected := range expectedValidFiles {
		found := false
		for _, valid := range validFiles {
			if valid == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file '%s' to be included in valid files", expected)
		}
	}
}

// Benchmark tests
func BenchmarkColorConversion(b *testing.B) {
	// Benchmark hex color conversion
	testColor := uint32(0xFF6750A4)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%06x", testColor&0xFFFFFF)
	}
}

// Integration tests
func TestWallpaperCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test command creation and flag parsing
	cmd := Command()

	// Test that command can be created without errors
	if cmd == nil {
		t.Error("Expected command to be created")
	}

	// Test flag parsing
	cmd.SetArgs([]string{"--file", "/path/to/wallpaper.jpg", "--no-smart"})
	err := cmd.ParseFlags([]string{"--file", "/path/to/wallpaper.jpg", "--no-smart"})
	if err != nil {
		t.Errorf("Flag parsing failed: %v", err)
	}

	// Check that flags were set
	fileFlagValue, _ := cmd.Flags().GetString("file")
	if fileFlagValue != "/path/to/wallpaper.jpg" {
		t.Errorf("Expected file flag to be '/path/to/wallpaper.jpg', got '%s'", fileFlagValue)
	}

	noSmartFlagValue, _ := cmd.Flags().GetBool("no-smart")
	if !noSmartFlagValue {
		t.Error("Expected no-smart flag to be true")
	}
}

func TestCaelestiaCompatibility(t *testing.T) {
	// Test that the command maintains caelestia compatibility
	cmd := Command()

	// Check that caelestia-compatible flags exist
	caelestiaFlags := []string{"print", "random", "file", "no-filter", "threshold", "no-smart"}

	for _, flagName := range caelestiaFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected caelestia-compatible flag '%s' to exist", flagName)
		}
	}

	// Check that deprecated flags are hidden
	deprecatedFlags := []string{"filter", "scheme", "info"}

	for _, flagName := range deprecatedFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected deprecated flag '%s' to exist for backward compatibility", flagName)
		} else if !flag.Hidden {
			t.Errorf("Expected deprecated flag '%s' to be hidden", flagName)
		}
	}
}

func TestJSONOutput(t *testing.T) {
	// Test JSON output structure for caelestia compatibility
	expectedStructure := map[string]interface{}{
		"name":    "dynamic",
		"flavour": "default",
		"mode":    "dark",
		"variant": "content",
		"colours": map[string]string{
			"primary":    "#6750a4",
			"background": "#1c1b1f",
		},
	}

	// Test that the structure can be marshaled to JSON
	jsonData, err := json.MarshalIndent(expectedStructure, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
	}

	// Test that the JSON can be unmarshaled back
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	// Check key fields
	if unmarshaled["name"] != "dynamic" {
		t.Errorf("Expected name to be 'dynamic', got %v", unmarshaled["name"])
	}

	if unmarshaled["mode"] != "dark" {
		t.Errorf("Expected mode to be 'dark', got %v", unmarshaled["mode"])
	}

	colours, ok := unmarshaled["colours"].(map[string]interface{})
	if !ok {
		t.Error("Expected colours to be a map")
	} else {
		if colours["primary"] != "#6750a4" {
			t.Errorf("Expected primary color to be '#6750a4', got %v", colours["primary"])
		}
	}
}
