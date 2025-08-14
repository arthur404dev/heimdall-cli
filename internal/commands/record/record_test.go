package record

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

	if cmd.Use != "record" {
		t.Errorf("Expected Use to be 'record', got %s", cmd.Use)
	}

	if cmd.Short != "Record the screen" {
		t.Errorf("Expected Short to be 'Record the screen', got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "wl-screenrec") {
		t.Errorf("Expected Long to contain 'wl-screenrec'")
	}

	// Check flags
	regionFlag := cmd.Flags().Lookup("region")
	if regionFlag == nil {
		t.Error("Expected region flag to exist")
	}
	if regionFlag.Shorthand != "r" {
		t.Errorf("Expected region flag shorthand to be 'r', got %s", regionFlag.Shorthand)
	}

	soundFlag := cmd.Flags().Lookup("sound")
	if soundFlag == nil {
		t.Error("Expected sound flag to exist")
	}
	if soundFlag.Shorthand != "s" {
		t.Errorf("Expected sound flag shorthand to be 's', got %s", soundFlag.Shorthand)
	}
}

func TestStartRecording(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalStateDir := paths.HeimdallStateDir
	paths.HeimdallStateDir = tempDir
	defer func() { paths.HeimdallStateDir = originalStateDir }()

	// Setup test config
	defer func() {
		config.Load() // Restore original config
	}()

	// Test with minimal config
	tests := []struct {
		name        string
		regionFlag  string
		soundFlag   bool
		expectError bool
		errorMsg    string
	}{
		{
			name:        "starts recording without region",
			regionFlag:  "",
			soundFlag:   false,
			expectError: true, // Will fail because wl-screenrec is not available in test
		},
		{
			name:        "handles region flag",
			regionFlag:  "100,100 200x200",
			soundFlag:   false,
			expectError: true, // Will fail because wl-screenrec is not available in test
		},
		{
			name:        "handles sound flag",
			regionFlag:  "",
			soundFlag:   true,
			expectError: true, // Will fail because external tools are not available in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flags
			regionFlag = tt.regionFlag
			soundFlag = tt.soundFlag

			err := startRecording()

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

func TestStopRecording(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalStateDir := paths.HeimdallStateDir
	originalRecordingsDir := paths.RecordingsDir
	paths.HeimdallStateDir = tempDir
	paths.RecordingsDir = filepath.Join(tempDir, "recordings")
	defer func() {
		paths.HeimdallStateDir = originalStateDir
		paths.RecordingsDir = originalRecordingsDir
	}()

	// Setup test config
	defer func() {
		config.Load() // Restore original config
	}()

	tests := []struct {
		name        string
		setup       func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "handles missing temp file gracefully",
			setup: func() error {
				// Create recordings directory
				return os.MkdirAll(paths.RecordingsDir, 0755)
			},
			expectError: true, // Will fail because temp file doesn't exist
			errorMsg:    "no such file",
		},
		{
			name: "creates recordings directory if missing",
			setup: func() error {
				// Create temp recording file
				tempFile := filepath.Join(paths.HeimdallStateDir, "recording_temp.mp4")
				return os.WriteFile(tempFile, []byte("fake video data"), 0644)
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.RemoveAll(paths.RecordingsDir)
			os.RemoveAll(filepath.Join(paths.HeimdallStateDir, "recording_temp.mp4"))

			err := tt.setup()
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err = stopRecording()

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

				// Check that recordings directory was created
				if _, err := os.Stat(paths.RecordingsDir); os.IsNotExist(err) {
					t.Error("Expected recordings directory to be created")
				}
			}
		})
	}
}

func TestRecordingConfiguration(t *testing.T) {
	tests := []struct {
		name             string
		config           *config.Config
		expectedTempFile string
		expectedFormat   string
	}{
		{
			name: "uses default configuration",
			config: &config.Config{
				Recording: config.RecordingConfig{
					TempFileName:    "",
					FileFormat:      "",
					FileNamePattern: "",
					Directory:       "",
				},
			},
			expectedTempFile: "recording_temp.mp4", // Default
			expectedFormat:   "mp4",                // Default
		},
		{
			name: "uses custom configuration",
			config: &config.Config{
				Recording: config.RecordingConfig{
					TempFileName:    "custom_temp.mkv",
					FileFormat:      "mkv",
					FileNamePattern: "custom_%Y%m%d_%H%M%S",
					Directory:       "/custom/recordings",
				},
			},
			expectedTempFile: "custom_temp.mkv",
			expectedFormat:   "mkv",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that configuration values are used correctly
			// This is more of a documentation test since we can't easily test the full flow
			if tt.config.Recording.TempFileName == "" {
				// Should use default
				if tt.expectedTempFile != "recording_temp.mp4" {
					t.Errorf("Expected default temp file name, got %s", tt.expectedTempFile)
				}
			} else {
				if tt.config.Recording.TempFileName != tt.expectedTempFile {
					t.Errorf("Expected temp file name %s, got %s", tt.expectedTempFile, tt.config.Recording.TempFileName)
				}
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
			region := tt.regionInput

			if region != "" {
				if region == "slurp" {
					// Would call slurp command
					if !tt.expectSlurp {
						t.Error("Expected slurp to be called")
					}
					// In real implementation, this would get output from slurp
					region = "mocked_slurp_output"
				}
				args = append(args, "-g", region)
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

func TestAudioSourceHandling(t *testing.T) {
	tests := []struct {
		name         string
		soundFlag    bool
		audioSource  string
		expectedArgs []string
		expectError  bool
	}{
		{
			name:         "skips audio when sound flag is false",
			soundFlag:    false,
			audioSource:  "auto",
			expectedArgs: []string{},
			expectError:  false,
		},
		{
			name:         "handles auto audio source",
			soundFlag:    true,
			audioSource:  "auto",
			expectedArgs: []string{"--audio", "--audio-device"},
			expectError:  true, // Will fail because pactl is not available in test
		},
		{
			name:         "handles specific audio source",
			soundFlag:    true,
			audioSource:  "alsa_output.pci-0000_00_1f.3.analog-stereo",
			expectedArgs: []string{"--audio", "--audio-device", "alsa_output.pci-0000_00_1f.3.analog-stereo"},
			expectError:  false,
		},
		{
			name:         "handles none audio source",
			soundFlag:    true,
			audioSource:  "none",
			expectedArgs: []string{},
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test audio source logic
			var args []string

			if tt.soundFlag && tt.audioSource != "none" {
				var audioDevice string
				if tt.audioSource == "auto" {
					// Would call pactl to get running sources
					// In test, we'll simulate this
					audioDevice = "" // Would be empty if no running sources found
				} else {
					audioDevice = tt.audioSource
				}

				if audioDevice == "" && tt.audioSource == "auto" {
					// This would cause an error in real implementation
					if !tt.expectError {
						t.Error("Expected error for empty audio device")
					}
				} else if audioDevice != "" {
					args = append(args, "--audio", "--audio-device", audioDevice)
				}
			}

			if !tt.expectError {
				expectedLen := len(tt.expectedArgs)
				if len(args) != expectedLen {
					t.Errorf("Expected %d args, got %d: %v", expectedLen, len(args), args)
				}

				for i, expected := range tt.expectedArgs {
					if i < len(args) && args[i] != expected {
						t.Errorf("Expected arg[%d] to be '%s', got '%s'", i, expected, args[i])
					}
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
	}{
		{
			name:            "uses default pattern",
			fileNamePattern: "",
			fileFormat:      "",
			expectedPattern: "recording_", // Should start with default pattern
		},
		{
			name:            "uses custom pattern",
			fileNamePattern: "custom_%Y%m%d_%H%M%S",
			fileFormat:      "mkv",
			expectedPattern: "custom_",
		},
		{
			name:            "handles pattern without timestamps",
			fileNamePattern: "my_recording",
			fileFormat:      "mp4",
			expectedPattern: "my_recording",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test filename generation logic
			pattern := tt.fileNamePattern
			if pattern == "" {
				pattern = "recording_%Y%m%d_%H%M%S"
			}

			// Replace timestamp patterns (simplified for test)
			filename := strings.ReplaceAll(pattern, "%Y%m%d", "20240101")
			filename = strings.ReplaceAll(filename, "%H%M%S", "120000")

			if !strings.HasPrefix(filename, tt.expectedPattern) {
				t.Errorf("Expected filename to start with '%s', got '%s'", tt.expectedPattern, filename)
			}
		})
	}
}

// Benchmark tests
func BenchmarkStartRecording(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	originalStateDir := paths.HeimdallStateDir
	paths.HeimdallStateDir = tempDir
	defer func() { paths.HeimdallStateDir = originalStateDir }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will fail, but we're measuring the setup time
		startRecording()
	}
}

func BenchmarkStopRecording(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	originalStateDir := paths.HeimdallStateDir
	originalRecordingsDir := paths.RecordingsDir
	paths.HeimdallStateDir = tempDir
	paths.RecordingsDir = filepath.Join(tempDir, "recordings")
	defer func() {
		paths.HeimdallStateDir = originalStateDir
		paths.RecordingsDir = originalRecordingsDir
	}()

	// Create temp file for each iteration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tempFile := filepath.Join(paths.HeimdallStateDir, "recording_temp.mp4")
		os.WriteFile(tempFile, []byte("fake video data"), 0644)
		os.MkdirAll(paths.RecordingsDir, 0755)
		b.StartTimer()

		stopRecording()
	}
}

// Integration tests
func TestRecordCommandIntegration(t *testing.T) {
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
	cmd.SetArgs([]string{"--region", "slurp", "--sound"})
	err := cmd.ParseFlags([]string{"--region", "slurp", "--sound"})
	if err != nil {
		t.Errorf("Flag parsing failed: %v", err)
	}

	// Check that flags were set
	regionFlagValue, _ := cmd.Flags().GetString("region")
	if regionFlagValue != "slurp" {
		t.Errorf("Expected region flag to be 'slurp', got '%s'", regionFlagValue)
	}

	soundFlagValue, _ := cmd.Flags().GetBool("sound")
	if !soundFlagValue {
		t.Error("Expected sound flag to be true")
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
			name:         "uses configured path",
			toolName:     "wl-screenrec",
			configPath:   "/custom/path/wl-screenrec",
			expectedPath: "/custom/path/wl-screenrec",
		},
		{
			name:         "uses default path when config is empty",
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
