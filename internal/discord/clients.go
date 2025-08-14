package discord

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/config"
	"github.com/arthur404dev/heimdall-cli/internal/utils/paths"
)

// Client represents a Discord client configuration
type Client struct {
	Name         string
	ConfigPath   string
	ThemePath    string
	FileFormat   string // "css" or "json"
	TemplateType string // "css" or "settings"
}

// ClientManager manages all Discord clients
type ClientManager struct {
	homeDir string
	clients []Client
}

// NewClientManager creates a new Discord client manager
func NewClientManager() *ClientManager {
	homeDir, _ := os.UserHomeDir()

	// Get configuration for Discord paths
	cfg := config.Get()

	// Build clients list from configuration
	clients := []Client{}

	if cfg != nil && cfg.Theme.Paths.Discord != "" {
		// Use configured paths
		clients = []Client{
			{
				Name:         "Discord Official",
				ConfigPath:   filepath.Dir(cfg.Theme.Paths.Discord),
				ThemePath:    cfg.Theme.Paths.Discord,
				FileFormat:   "css",
				TemplateType: "css",
			},
			{
				Name:         "Vesktop",
				ConfigPath:   filepath.Dir(cfg.Theme.Paths.Vesktop),
				ThemePath:    cfg.Theme.Paths.Vesktop,
				FileFormat:   "css",
				TemplateType: "css",
			},
			{
				Name:         "Vencord",
				ConfigPath:   filepath.Dir(cfg.Theme.Paths.Vencord),
				ThemePath:    cfg.Theme.Paths.Vencord,
				FileFormat:   "css",
				TemplateType: "css",
			},
			{
				Name:         "BetterDiscord",
				ConfigPath:   filepath.Dir(cfg.Theme.Paths.BetterDiscord),
				ThemePath:    cfg.Theme.Paths.BetterDiscord,
				FileFormat:   "css",
				TemplateType: "betterdiscord",
			},
			{
				Name:         "Discord Canary",
				ConfigPath:   filepath.Dir(cfg.Theme.Paths.DiscordCanary),
				ThemePath:    cfg.Theme.Paths.DiscordCanary,
				FileFormat:   "css",
				TemplateType: "css",
			},
			{
				Name:         "Equicord",
				ConfigPath:   filepath.Dir(cfg.Theme.Paths.Equicord),
				ThemePath:    cfg.Theme.Paths.Equicord,
				FileFormat:   "css",
				TemplateType: "css",
			},
		}
	} else {
		// This shouldn't happen if config is properly loaded with defaults
		// But provide fallback just in case
		clients = []Client{
			{
				Name:         "Discord Official",
				ConfigPath:   filepath.Join(homeDir, ".config", "discord"),
				ThemePath:    filepath.Join(homeDir, ".config", "discord", "themes", "heimdall.css"),
				FileFormat:   "css",
				TemplateType: "css",
			},
			{
				Name:         "Vesktop",
				ConfigPath:   filepath.Join(homeDir, ".config", "vesktop"),
				ThemePath:    filepath.Join(homeDir, ".config", "vesktop", "themes", "heimdall.css"),
				FileFormat:   "css",
				TemplateType: "css",
			},
		}
	}

	return &ClientManager{
		homeDir: homeDir,
		clients: clients,
	}
}

// GetClients returns all Discord clients
func (cm *ClientManager) GetClients() []Client {
	return cm.clients
}

// GetDetectedClients returns only clients that are installed/detected
func (cm *ClientManager) GetDetectedClients() []Client {
	var detected []Client

	for _, client := range cm.clients {
		if cm.isClientInstalled(client) {
			detected = append(detected, client)
		}
	}

	return detected
}

// isClientInstalled checks if a Discord client is installed
func (cm *ClientManager) isClientInstalled(client Client) bool {
	// Check if the config directory exists
	if _, err := os.Stat(client.ConfigPath); os.IsNotExist(err) {
		return false
	}

	// For Discord Official, also check for the main executable or config files
	if client.Name == "Discord Official" {
		// Check for common Discord files/directories
		discordFiles := []string{
			filepath.Join(client.ConfigPath, "Local State"),
			filepath.Join(client.ConfigPath, "Preferences"),
		}

		for _, file := range discordFiles {
			if _, err := os.Stat(file); err == nil {
				return true
			}
		}
		return false
	}

	return true
}

// ApplyTheme applies a theme to a specific Discord client
func (cm *ClientManager) ApplyTheme(client Client, colors map[string]string, template string) error {
	// Ensure the theme directory exists
	themeDir := filepath.Dir(client.ThemePath)
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory %s: %w", themeDir, err)
	}

	switch client.FileFormat {
	case "css":
		return cm.applyCSSTheme(client, colors, template)
	case "json":
		return cm.applyJSONTheme(client, colors)
	default:
		return fmt.Errorf("unsupported file format: %s", client.FileFormat)
	}
}

// applyCSSTheme applies a CSS theme to a Discord client
func (cm *ClientManager) applyCSSTheme(client Client, colors map[string]string, template string) error {
	// Use simple string replacement for CSS template
	rendered := template
	for key, value := range colors {
		placeholder := "{{" + key + "}}"
		rendered = strings.ReplaceAll(rendered, placeholder, value)
	}

	// Write the CSS file
	if err := paths.AtomicWrite(client.ThemePath, []byte(rendered)); err != nil {
		return fmt.Errorf("failed to write CSS theme for %s: %w", client.Name, err)
	}

	return nil
}

// applyJSONTheme applies a JSON theme to Discord Official client
func (cm *ClientManager) applyJSONTheme(client Client, colors map[string]string) error {
	// For Discord Official, we need to modify the settings.json file
	// This is more complex as we need to preserve existing settings

	var settings map[string]interface{}

	// Try to read existing settings
	if data, err := os.ReadFile(client.ThemePath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			// If parsing fails, start with empty settings
			settings = make(map[string]interface{})
		}
	} else {
		// File doesn't exist, start with empty settings
		settings = make(map[string]interface{})
	}

	// Add theme-related settings
	// Discord Official doesn't directly support custom themes, but we can set some appearance options
	settings["theme"] = "dark" // Force dark theme for better color visibility

	// Add a custom CSS injection if the client supports it (some Discord mods do)
	if cssContent := cm.generateDiscordOfficialCSS(colors); cssContent != "" {
		settings["customCSS"] = cssContent
	}

	// Write the updated settings
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings for %s: %w", client.Name, err)
	}

	if err := paths.AtomicWrite(client.ThemePath, data); err != nil {
		return fmt.Errorf("failed to write settings for %s: %w", client.Name, err)
	}

	return nil
}

// generateDiscordOfficialCSS generates CSS for Discord Official client
func (cm *ClientManager) generateDiscordOfficialCSS(colors map[string]string) string {
	// This is a simplified CSS that might work with some Discord mods
	// The official Discord client doesn't support custom CSS directly
	css := `/* Heimdall theme for Discord Official */
/* This requires a Discord mod that supports custom CSS */

:root {
    --background-primary: #{{background}};
    --background-secondary: #{{colour0}};
    --background-tertiary: #{{colour8}};
    --text-normal: #{{foreground}};
    --text-muted: #{{colour7}};
    --interactive-normal: #{{colour4}};
    --interactive-hover: #{{colour5}};
    --interactive-active: #{{colour6}};
    --brand-experiment: #{{colour4}};
}
`

	// Apply color replacements
	for key, value := range colors {
		placeholder := "{{" + key + "}}"
		css = strings.ReplaceAll(css, placeholder, value)
	}

	return css
}

// ApplyThemeToAll applies a theme to all detected Discord clients
func (cm *ClientManager) ApplyThemeToAll(colors map[string]string, cssTemplate, betterDiscordTemplate string) error {
	detected := cm.GetDetectedClients()

	if len(detected) == 0 {
		return fmt.Errorf("no Discord clients detected")
	}

	var errors []string
	successCount := 0

	for _, client := range detected {
		var template string

		// Choose the appropriate template based on client type
		switch client.TemplateType {
		case "betterdiscord":
			template = betterDiscordTemplate
		case "css":
			template = cssTemplate
		case "settings":
			// JSON settings don't use templates
			template = ""
		default:
			template = cssTemplate
		}

		if err := cm.ApplyTheme(client, colors, template); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", client.Name, err))
		} else {
			successCount++
		}
	}

	// Report results
	if len(errors) > 0 {
		if successCount == 0 {
			return fmt.Errorf("failed to apply theme to any Discord clients: %s", strings.Join(errors, "; "))
		}
		// Some succeeded, some failed - log warnings but don't fail
		fmt.Fprintf(os.Stderr, "Warning: failed to apply theme to some Discord clients: %s\n", strings.Join(errors, "; "))
	}

	return nil
}
