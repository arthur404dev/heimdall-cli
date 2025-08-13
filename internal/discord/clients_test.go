package discord

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClientManager(t *testing.T) {
	manager := NewClientManager()

	if manager == nil {
		t.Fatal("NewClientManager returned nil")
	}

	clients := manager.GetClients()
	if len(clients) != 6 {
		t.Fatalf("Expected 6 Discord clients, got %d", len(clients))
	}

	// Verify all expected clients are present
	expectedClients := []string{
		"Discord Official",
		"Vesktop",
		"Vencord",
		"BetterDiscord",
		"OpenAsar",
		"Armcord",
	}

	for i, expected := range expectedClients {
		if clients[i].Name != expected {
			t.Errorf("Expected client %d to be %s, got %s", i, expected, clients[i].Name)
		}
	}
}

func TestClientPaths(t *testing.T) {
	manager := NewClientManager()
	clients := manager.GetClients()

	homeDir, _ := os.UserHomeDir()

	expectedPaths := map[string]string{
		"Discord Official": filepath.Join(homeDir, ".config", "discord", "settings.json"),
		"Vesktop":          filepath.Join(homeDir, ".config", "vesktop", "themes", "heimdall.css"),
		"Vencord":          filepath.Join(homeDir, ".config", "Vencord", "themes", "heimdall.css"),
		"BetterDiscord":    filepath.Join(homeDir, ".config", "BetterDiscord", "themes", "heimdall.theme.css"),
		"OpenAsar":         filepath.Join(homeDir, ".config", "discord", "themes", "heimdall.css"),
		"Armcord":          filepath.Join(homeDir, ".config", "armcord", "themes", "heimdall.css"),
	}

	for _, client := range clients {
		expectedPath, exists := expectedPaths[client.Name]
		if !exists {
			t.Errorf("Unexpected client: %s", client.Name)
			continue
		}

		if client.ThemePath != expectedPath {
			t.Errorf("Client %s: expected path %s, got %s", client.Name, expectedPath, client.ThemePath)
		}
	}
}

func TestClientFileFormats(t *testing.T) {
	manager := NewClientManager()
	clients := manager.GetClients()

	expectedFormats := map[string]string{
		"Discord Official": "json",
		"Vesktop":          "css",
		"Vencord":          "css",
		"BetterDiscord":    "css",
		"OpenAsar":         "css",
		"Armcord":          "css",
	}

	for _, client := range clients {
		expectedFormat, exists := expectedFormats[client.Name]
		if !exists {
			t.Errorf("Unexpected client: %s", client.Name)
			continue
		}

		if client.FileFormat != expectedFormat {
			t.Errorf("Client %s: expected format %s, got %s", client.Name, expectedFormat, client.FileFormat)
		}
	}
}

func TestGetTemplate(t *testing.T) {
	cssTemplate := GetTemplate("css")
	betterDiscordTemplate := GetTemplate("betterdiscord")
	defaultTemplate := GetTemplate("unknown")

	if cssTemplate == "" {
		t.Error("CSS template should not be empty")
	}

	if betterDiscordTemplate == "" {
		t.Error("BetterDiscord template should not be empty")
	}

	if defaultTemplate != cssTemplate {
		t.Error("Unknown template type should return CSS template as default")
	}

	// Verify BetterDiscord template has META header
	if !contains(betterDiscordTemplate, "@name Heimdall") {
		t.Error("BetterDiscord template should contain META header with @name")
	}

	// Verify CSS template uses simple placeholder format
	if !contains(cssTemplate, "{{colour4}}") {
		t.Error("CSS template should use simple {{colour4}} placeholder format")
	}
}

func TestApplyThemeToMockClients(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create a test client manager with custom paths
	manager := &ClientManager{
		homeDir: tempDir,
		clients: []Client{
			{
				Name:         "Test Vesktop",
				ConfigPath:   filepath.Join(tempDir, "vesktop"),
				ThemePath:    filepath.Join(tempDir, "vesktop", "themes", "heimdall.css"),
				FileFormat:   "css",
				TemplateType: "css",
			},
		},
	}

	// Test colors
	colors := map[string]string{
		"colour0":    "1a1b26",
		"colour4":    "7aa2f7",
		"background": "1a1b26",
		"foreground": "c0caf5",
	}

	// Get template and apply theme
	template := GetTemplate("css")
	err := manager.ApplyTheme(manager.clients[0], colors, template)

	if err != nil {
		t.Fatalf("Failed to apply theme: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(manager.clients[0].ThemePath); os.IsNotExist(err) {
		t.Error("Theme file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(manager.clients[0].ThemePath)
	if err != nil {
		t.Fatalf("Failed to read theme file: %v", err)
	}

	contentStr := string(content)

	// Verify color substitution worked
	if contains(contentStr, "{{colour4}}") {
		t.Error("Template placeholders were not replaced")
	}

	if !contains(contentStr, "#7aa2f7") {
		t.Error("Color values were not properly substituted")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsInMiddle(s, substr))))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
