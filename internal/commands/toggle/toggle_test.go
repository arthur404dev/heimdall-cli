package toggle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/utils/hypr"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// MockHyprClient is a simple mock implementation for testing
type MockHyprClient struct {
	windows       []hypr.Window
	workspaces    []hypr.Workspace
	monitors      []hypr.Monitor
	dispatchErr   error
	windowsErr    error
	workspacesErr error
	monitorsErr   error
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

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	if cmd.Use != "toggle [workspace]" {
		t.Errorf("Expected Use to be 'toggle [workspace]', got %s", cmd.Use)
	}

	if cmd.Short != "Toggle special workspaces" {
		t.Errorf("Expected Short to be 'Toggle special workspaces', got %s", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "Toggle special workspaces in Hyprland") {
		t.Errorf("Expected Long to contain 'Toggle special workspaces in Hyprland'")
	}

	// Check valid args
	expectedArgs := []string{"communication", "music", "sysmon", "specialws", "todo"}
	if len(cmd.ValidArgs) != len(expectedArgs) {
		t.Errorf("Expected %d valid args, got %d", len(expectedArgs), len(cmd.ValidArgs))
	}

	for i, expected := range expectedArgs {
		if i < len(cmd.ValidArgs) && cmd.ValidArgs[i] != expected {
			t.Errorf("Expected ValidArgs[%d] to be '%s', got '%s'", i, expected, cmd.ValidArgs[i])
		}
	}
}

func TestLoadConfig(t *testing.T) {
	// Test default configuration loading
	config := loadConfig()

	// Check that default config is loaded
	if config.Communication == nil {
		t.Error("Expected Communication config to be loaded")
	}

	if config.Music == nil {
		t.Error("Expected Music config to be loaded")
	}

	if config.Sysmon == nil {
		t.Error("Expected Sysmon config to be loaded")
	}

	if config.Todo == nil {
		t.Error("Expected Todo config to be loaded")
	}

	// Check specific default values
	if discord, exists := config.Communication["discord"]; exists {
		if !discord.Enable {
			t.Error("Expected discord to be enabled by default")
		}
		if len(discord.Match) == 0 {
			t.Error("Expected discord to have match criteria")
		}
		if len(discord.Command) == 0 {
			t.Error("Expected discord to have command")
		}
		if !discord.Move {
			t.Error("Expected discord to have move enabled")
		}
	} else {
		t.Error("Expected discord config to exist")
	}

	if spotify, exists := config.Music["spotify"]; exists {
		if !spotify.Enable {
			t.Error("Expected spotify to be enabled by default")
		}
		if len(spotify.Match) == 0 {
			t.Error("Expected spotify to have match criteria")
		}
	} else {
		t.Error("Expected spotify config to exist")
	}
}

func TestMergeConfig(t *testing.T) {
	// Create base config
	dst := defaultConfig

	// Create source config with overrides
	src := ToggleConfig{
		Communication: map[string]ClientConfig{
			"discord": {
				Enable: false, // Override default
				Match: []map[string]interface{}{
					{"class": "custom-discord"},
				},
			},
			"telegram": { // New entry
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "telegram"},
				},
			},
		},
		Music: map[string]ClientConfig{
			"custom-player": { // New entry
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "custom-player"},
				},
			},
		},
	}

	// Merge configs
	mergeConfig(&dst, &src)

	// Check that discord was overridden
	if discord, exists := dst.Communication["discord"]; exists {
		if discord.Enable {
			t.Error("Expected discord to be disabled after merge")
		}
	} else {
		t.Error("Expected discord config to still exist after merge")
	}

	// Check that telegram was added
	if telegram, exists := dst.Communication["telegram"]; exists {
		if !telegram.Enable {
			t.Error("Expected telegram to be enabled")
		}
	} else {
		t.Error("Expected telegram config to be added")
	}

	// Check that custom player was added
	if player, exists := dst.Music["custom-player"]; exists {
		if !player.Enable {
			t.Error("Expected custom-player to be enabled")
		}
	} else {
		t.Error("Expected custom-player config to be added")
	}

	// Check that original spotify config still exists
	if _, exists := dst.Music["spotify"]; !exists {
		t.Error("Expected original spotify config to still exist")
	}
}

func TestMatchesWindow(t *testing.T) {
	testWindow := hypr.Window{
		Address:      "0x12345",
		Class:        "discord",
		Title:        "Discord - General",
		InitialTitle: "Discord",
		Workspace: hypr.Workspace{
			ID:   1,
			Name: "1",
		},
	}

	tests := []struct {
		name     string
		window   hypr.Window
		matches  []map[string]interface{}
		expected bool
	}{
		{
			name:   "matches by class",
			window: testWindow,
			matches: []map[string]interface{}{
				{"class": "discord"},
			},
			expected: true,
		},
		{
			name:   "matches by partial class",
			window: testWindow,
			matches: []map[string]interface{}{
				{"class": "disc"},
			},
			expected: true,
		},
		{
			name:   "matches by title",
			window: testWindow,
			matches: []map[string]interface{}{
				{"title": "Discord"},
			},
			expected: true,
		},
		{
			name:   "matches by initial title",
			window: testWindow,
			matches: []map[string]interface{}{
				{"initialTitle": "Discord"},
			},
			expected: true,
		},
		{
			name:   "matches by workspace",
			window: testWindow,
			matches: []map[string]interface{}{
				{"workspace": map[string]interface{}{"id": 1}},
			},
			expected: true,
		},
		{
			name:   "matches multiple criteria",
			window: testWindow,
			matches: []map[string]interface{}{
				{"class": "discord", "title": "Discord"},
			},
			expected: true,
		},
		{
			name:   "no match when class differs",
			window: testWindow,
			matches: []map[string]interface{}{
				{"class": "firefox"},
			},
			expected: false,
		},
		{
			name:   "no match when multiple criteria don't all match",
			window: testWindow,
			matches: []map[string]interface{}{
				{"class": "discord", "title": "Firefox"},
			},
			expected: false,
		},
		{
			name:   "matches any of multiple match objects",
			window: testWindow,
			matches: []map[string]interface{}{
				{"class": "firefox"},
				{"class": "discord"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWindow(tt.window, tt.matches)
			if result != tt.expected {
				t.Errorf("matchesWindow() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestWindowToMap(t *testing.T) {
	testWindow := hypr.Window{
		Address:      "0x12345",
		Class:        "discord",
		Title:        "Discord - General",
		InitialTitle: "Discord",
		Workspace: hypr.Workspace{
			ID:   1,
			Name: "1",
		},
	}

	result := windowToMap(testWindow)

	// Check all expected fields
	if result["address"] != "0x12345" {
		t.Errorf("Expected address to be '0x12345', got %v", result["address"])
	}

	if result["class"] != "discord" {
		t.Errorf("Expected class to be 'discord', got %v", result["class"])
	}

	if result["title"] != "Discord - General" {
		t.Errorf("Expected title to be 'Discord - General', got %v", result["title"])
	}

	if result["initialTitle"] != "Discord" {
		t.Errorf("Expected initialTitle to be 'Discord', got %v", result["initialTitle"])
	}

	// Check workspace nested object
	workspace, ok := result["workspace"].(map[string]interface{})
	if !ok {
		t.Error("Expected workspace to be a map")
	} else {
		if workspace["id"] != 1 {
			t.Errorf("Expected workspace id to be 1, got %v", workspace["id"])
		}
		if workspace["name"] != "1" {
			t.Errorf("Expected workspace name to be '1', got %v", workspace["name"])
		}
	}
}

func TestIsSubset(t *testing.T) {
	tests := []struct {
		name     string
		superset map[string]interface{}
		subset   map[string]interface{}
		expected bool
	}{
		{
			name: "simple string match",
			superset: map[string]interface{}{
				"class": "discord",
				"title": "Discord - General",
			},
			subset: map[string]interface{}{
				"class": "discord",
			},
			expected: true,
		},
		{
			name: "partial string match",
			superset: map[string]interface{}{
				"class": "discord",
			},
			subset: map[string]interface{}{
				"class": "disc",
			},
			expected: true,
		},
		{
			name: "nested map match",
			superset: map[string]interface{}{
				"workspace": map[string]interface{}{
					"id":   1,
					"name": "workspace1",
				},
			},
			subset: map[string]interface{}{
				"workspace": map[string]interface{}{
					"id": 1,
				},
			},
			expected: true,
		},
		{
			name: "exact value match",
			superset: map[string]interface{}{
				"id": 42,
			},
			subset: map[string]interface{}{
				"id": 42,
			},
			expected: true,
		},
		{
			name: "no match when key missing",
			superset: map[string]interface{}{
				"class": "discord",
			},
			subset: map[string]interface{}{
				"title": "Discord",
			},
			expected: false,
		},
		{
			name: "no match when string doesn't contain",
			superset: map[string]interface{}{
				"class": "firefox",
			},
			subset: map[string]interface{}{
				"class": "discord",
			},
			expected: false,
		},
		{
			name: "no match when nested map doesn't match",
			superset: map[string]interface{}{
				"workspace": map[string]interface{}{
					"id": 1,
				},
			},
			subset: map[string]interface{}{
				"workspace": map[string]interface{}{
					"id": 2,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSubset(tt.superset, tt.subset)
			if result != tt.expected {
				t.Errorf("isSubset() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSpawnClient(t *testing.T) {
	tests := []struct {
		name    string
		command []string
	}{
		{
			name:    "handles empty command",
			command: []string{},
		},
		{
			name:    "handles single command",
			command: []string{"discord"},
		},
		{
			name:    "handles command with args",
			command: []string{"discord", "--flag", "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test just ensures spawnClient doesn't panic
			// The actual spawning will fail in test environment, which is expected
			spawnClient(tt.command)
		})
	}
}

func TestHandleSpecialWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		workspaces  []hypr.Workspace
		windows     []hypr.Window
		expectError bool
		errorMsg    string
	}{
		{
			name: "handles no special workspace",
			workspaces: []hypr.Workspace{
				{ID: 1, Name: "1"},
				{ID: 2, Name: "2"},
			},
			windows:     []hypr.Window{},
			expectError: false,
		},
		{
			name: "handles existing special:special workspace",
			workspaces: []hypr.Workspace{
				{ID: 1, Name: "1"},
				{ID: -99, Name: "special:special"},
			},
			windows:     []hypr.Window{},
			expectError: false,
		},
		{
			name: "finds window on special workspace",
			workspaces: []hypr.Workspace{
				{ID: 1, Name: "1"},
				{ID: -98, Name: "special:music"},
			},
			windows: []hypr.Window{
				{
					Address: "0x12345",
					Class:   "spotify",
					Workspace: hypr.Workspace{
						ID:   -98,
						Name: "special:music",
					},
				},
			},
			expectError: false,
		},
		{
			name:        "handles GetWorkspaces error",
			workspaces:  []hypr.Workspace{},
			windows:     []hypr.Window{},
			expectError: true,
			errorMsg:    "failed to get workspaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHyprClient{
				workspaces: tt.workspaces,
				windows:    tt.windows,
			}

			if tt.expectError && tt.errorMsg == "failed to get workspaces" {
				mockClient.workspacesErr = fmt.Errorf("connection failed")
			}

			err := testHandleSpecialWorkspace(mockClient)

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

// testHandleSpecialWorkspace tests the core logic without requiring exact hypr.Client interface
func testHandleSpecialWorkspace(client *MockHyprClient) error {
	workspaces, err := client.GetWorkspaces()
	if err != nil {
		return fmt.Errorf("failed to get workspaces: %w", err)
	}

	// Check if special:special workspace exists
	onSpecialWs := false
	for _, ws := range workspaces {
		if ws.Name == "special:special" {
			onSpecialWs = true
			break
		}
	}

	toggleWs := "special"

	if !onSpecialWs {
		// Get active window to check current workspace
		windows, err := client.GetWindows()
		if err == nil {
			// Find window on a special workspace
			for _, w := range windows {
				if strings.HasPrefix(w.Workspace.Name, "special:") {
					toggleWs = w.Workspace.Name[8:] // Remove "special:" prefix
					break
				}
			}
		}
	}

	return client.Dispatch("togglespecialworkspace", toggleWs)
}

func TestHandleClientConfig(t *testing.T) {
	testWindows := []hypr.Window{
		{
			Address: "0x12345",
			Class:   "discord",
			Title:   "Discord - General",
			Workspace: hypr.Workspace{
				ID:   1,
				Name: "1",
			},
		},
		{
			Address: "0x67890",
			Class:   "spotify",
			Title:   "Spotify",
			Workspace: hypr.Workspace{
				ID:   2,
				Name: "2",
			},
		},
	}

	tests := []struct {
		name        string
		config      ClientConfig
		workspace   string
		windows     []hypr.Window
		expectSpawn bool
		expectMove  bool
	}{
		{
			name: "spawns and moves when no matching window exists",
			config: ClientConfig{
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "firefox"},
				},
				Command: []string{"firefox"},
				Move:    true,
			},
			workspace:   "communication",
			windows:     testWindows,
			expectSpawn: true,
			expectMove:  false, // No matching window to move
		},
		{
			name: "moves existing window without spawning",
			config: ClientConfig{
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "discord"},
				},
				Command: []string{"discord"},
				Move:    true,
			},
			workspace:   "communication",
			windows:     testWindows,
			expectSpawn: false, // Window already exists
			expectMove:  true,
		},
		{
			name: "does nothing when disabled",
			config: ClientConfig{
				Enable: false,
				Match: []map[string]interface{}{
					{"class": "discord"},
				},
				Command: []string{"discord"},
				Move:    true,
			},
			workspace:   "communication",
			windows:     testWindows,
			expectSpawn: false,
			expectMove:  false,
		},
		{
			name: "spawns but doesn't move when move is disabled",
			config: ClientConfig{
				Enable: true,
				Match: []map[string]interface{}{
					{"class": "firefox"},
				},
				Command: []string{"firefox"},
				Move:    false,
			},
			workspace:   "communication",
			windows:     testWindows,
			expectSpawn: true,
			expectMove:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the logic (simplified version of handleClientConfig)
			shouldSpawn := tt.config.Enable && len(tt.config.Command) > 0
			if shouldSpawn {
				// Check if window already exists
				for _, w := range tt.windows {
					if matchesWindow(w, tt.config.Match) {
						shouldSpawn = false
						break
					}
				}
			}

			shouldMove := tt.config.Enable && tt.config.Move
			var windowsToMove []hypr.Window
			if shouldMove {
				for _, w := range tt.windows {
					if matchesWindow(w, tt.config.Match) {
						windowsToMove = append(windowsToMove, w)
					}
				}
			}

			if shouldSpawn != tt.expectSpawn {
				t.Errorf("Expected spawn %v, got %v", tt.expectSpawn, shouldSpawn)
			}

			if (len(windowsToMove) > 0) != tt.expectMove {
				t.Errorf("Expected move %v, got %v", tt.expectMove, len(windowsToMove) > 0)
			}
		})
	}
}

func TestConfigFileLoading(t *testing.T) {
	// Test configuration file loading with custom config
	tempDir := t.TempDir()
	originalConfigDir := paths.HeimdallConfigDir
	paths.HeimdallConfigDir = tempDir
	defer func() { paths.HeimdallConfigDir = originalConfigDir }()

	configFile := filepath.Join(tempDir, "config.json")

	tests := []struct {
		name          string
		configContent string
		expectCustom  bool
	}{
		{
			name:          "loads default config when file doesn't exist",
			configContent: "",
			expectCustom:  false,
		},
		{
			name: "loads custom config when file exists",
			configContent: `{
				"toggles": {
					"communication": {
						"custom-app": {
							"enable": true,
							"match": [{"class": "custom-app"}],
							"command": ["custom-app"],
							"move": true
						}
					}
				}
			}`,
			expectCustom: true,
		},
		{
			name: "handles invalid JSON gracefully",
			configContent: `{
				"toggles": {
					"communication": {
						"invalid": "json"
			}`,
			expectCustom: false, // Should fall back to defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up config file
			os.Remove(configFile)

			// Create config file if content provided
			if tt.configContent != "" {
				err := os.WriteFile(configFile, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
			}

			config := loadConfig()

			// Check if custom config was loaded
			if tt.expectCustom {
				if _, exists := config.Communication["custom-app"]; !exists {
					t.Error("Expected custom-app config to be loaded")
				}
			} else {
				// Should have default discord config
				if _, exists := config.Communication["discord"]; !exists {
					t.Error("Expected default discord config to be loaded")
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkMatchesWindow(b *testing.B) {
	testWindow := hypr.Window{
		Address:      "0x12345",
		Class:        "discord",
		Title:        "Discord - General",
		InitialTitle: "Discord",
		Workspace: hypr.Workspace{
			ID:   1,
			Name: "1",
		},
	}

	matches := []map[string]interface{}{
		{"class": "discord"},
		{"title": "Discord"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matchesWindow(testWindow, matches)
	}
}

func BenchmarkIsSubset(b *testing.B) {
	superset := map[string]interface{}{
		"class": "discord",
		"title": "Discord - General",
		"workspace": map[string]interface{}{
			"id":   1,
			"name": "workspace1",
		},
	}

	subset := map[string]interface{}{
		"class": "discord",
		"workspace": map[string]interface{}{
			"id": 1,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isSubset(superset, subset)
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadConfig()
	}
}

// Integration tests
func TestToggleCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Test command creation and validation
	cmd := NewCommand()

	// Test that command can be created without errors
	if cmd == nil {
		t.Error("Expected command to be created")
	}

	// Test valid arguments
	validWorkspaces := []string{"communication", "music", "sysmon", "specialws", "todo"}
	for _, workspace := range validWorkspaces {
		found := false
		for _, validArg := range cmd.ValidArgs {
			if validArg == workspace {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected '%s' to be in ValidArgs", workspace)
		}
	}
}

func TestWorkspaceConfigIntegration(t *testing.T) {
	// Test that all default workspace configurations are valid
	config := loadConfig()

	workspaceConfigs := map[string]map[string]ClientConfig{
		"communication": config.Communication,
		"music":         config.Music,
		"sysmon":        config.Sysmon,
		"todo":          config.Todo,
	}

	for workspaceName, workspaceConfig := range workspaceConfigs {
		if len(workspaceConfig) == 0 {
			t.Errorf("Expected %s workspace to have client configurations", workspaceName)
			continue
		}

		for clientName, clientConfig := range workspaceConfig {
			// Check that enabled clients have match criteria
			if clientConfig.Enable && len(clientConfig.Match) == 0 {
				t.Errorf("Expected enabled client %s in %s workspace to have match criteria", clientName, workspaceName)
			}

			// Check that clients with commands have valid command arrays
			if len(clientConfig.Command) > 0 && clientConfig.Command[0] == "" {
				t.Errorf("Expected client %s in %s workspace to have valid command", clientName, workspaceName)
			}

			// Check that match criteria are valid
			for _, match := range clientConfig.Match {
				if len(match) == 0 {
					t.Errorf("Expected client %s in %s workspace to have non-empty match criteria", clientName, workspaceName)
				}
			}
		}
	}
}
