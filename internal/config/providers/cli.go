package providers

import (
	"github.com/arthur404dev/heimdall-cli/internal/config/schema"
)

// CLIProvider manages the main CLI configuration
type CLIProvider struct {
	*BaseProvider
}

// NewCLIProvider creates a new CLI configuration provider
func NewCLIProvider(configPath string) Provider {
	provider := &CLIProvider{
		BaseProvider: NewBaseProvider("cli", configPath),
	}

	// Define the schema for CLI configuration
	provider.initSchema()
	return provider
}

// initSchema initializes the CLI configuration schema
func (c *CLIProvider) initSchema() {
	// This is a simplified schema for the CLI config
	// In production, this would be loaded from a JSON schema file
	schemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title": "Heimdall CLI Configuration",
		"type": "object",
		"properties": {
			"version": {
				"type": "string",
				"description": "Configuration version"
			},
			"theme": {
				"type": "object",
				"properties": {
					"enableTerm": {"type": "boolean"},
					"enableHypr": {"type": "boolean"},
					"enableDiscord": {"type": "boolean"},
					"enableSpicetify": {"type": "boolean"},
					"enableFuzzel": {"type": "boolean"},
					"enableBtop": {"type": "boolean"},
					"enableGtk": {"type": "boolean"},
					"enableQt": {"type": "boolean"}
				}
			},
			"scheme": {
				"type": "object",
				"properties": {
					"default": {"type": "string"},
					"autoMode": {"type": "boolean"},
					"materialYou": {"type": "boolean"}
				}
			},
			"wallpaper": {
				"type": "object",
				"properties": {
					"directory": {"type": "string"},
					"filter": {"type": "boolean"},
					"threshold": {"type": "number"},
					"smartMode": {"type": "boolean"},
					"extensions": {
						"type": "array",
						"items": {"type": "string"}
					}
				}
			},
			"screenshot": {
				"type": "object",
				"properties": {
					"directory": {"type": "string"},
					"fileFormat": {"type": "string"},
					"fileNamePattern": {"type": "string"},
					"copyToClipboard": {"type": "boolean"},
					"openWithSwappy": {"type": "boolean"},
					"showNotification": {"type": "boolean"},
					"notificationTimeout": {"type": "integer"},
					"freezeFileName": {"type": "string"}
				}
			},
			"recording": {
				"type": "object",
				"properties": {
					"directory": {"type": "string"},
					"fileFormat": {"type": "string"},
					"fileNamePattern": {"type": "string"},
					"tempFileName": {"type": "string"},
					"showNotification": {"type": "boolean"},
					"audioSource": {"type": "string"}
				}
			},
			"clipboard": {
				"type": "object",
				"properties": {
					"maxEntries": {"type": "integer"},
					"fuzzelPrompt": {"type": "string"},
					"fuzzelArgs": {
						"type": "array",
						"items": {"type": "string"}
					},
					"previewLength": {"type": "integer"},
					"deleteOnSelect": {"type": "boolean"}
				}
			},
			"emoji": {
				"type": "object",
				"properties": {
					"dataDirectory": {"type": "string"},
					"sources": {
						"type": "array",
						"items": {"type": "string"}
					},
					"fuzzelPrompt": {"type": "string"},
					"fuzzelArgs": {
						"type": "array",
						"items": {"type": "string"}
					},
					"copyToClipboard": {"type": "boolean"},
					"typeDirectly": {"type": "boolean"},
					"downloadTimeout": {"type": "integer"}
				}
			},
			"pip": {
				"type": "object",
				"properties": {
					"enabled": {"type": "boolean"},
					"pidFile": {"type": "string"},
					"windowSize": {"type": "string"},
					"windowPosition": {"type": "string"},
					"videoApps": {
						"type": "array",
						"items": {"type": "string"}
					},
					"videoKeywords": {
						"type": "array",
						"items": {"type": "string"}
					},
					"pinWindows": {"type": "boolean"},
					"alwaysOnTop": {"type": "boolean"}
				}
			},
			"notification": {
				"type": "object",
				"properties": {
					"enabled": {"type": "boolean"},
					"provider": {"type": "string"},
					"defaultTimeout": {"type": "integer"},
					"appName": {"type": "string"},
					"defaultUrgency": {"type": "string"}
				}
			},
			"paths": {
				"type": "object",
				"properties": {
					"templates": {"type": "string"},
					"schemes": {"type": "string"},
					"stateDir": {"type": "string"},
					"cacheDir": {"type": "string"},
					"dataDir": {"type": "string"}
				}
			},
			"network": {
				"type": "object",
				"properties": {
					"ipcTimeout": {"type": "integer"},
					"hyprIpcTimeout": {"type": "integer"}
				}
			},
			"external_tools": {
				"type": "object",
				"additionalProperties": {"type": "string"}
			},
			"config_paths": {
				"type": "object",
				"description": "Paths for configuration files",
				"properties": {
					"base_dir": {"type": "string"},
					"file_pattern": {"type": "string"},
					"schema_dir": {"type": "string"},
					"backup_dir": {"type": "string"},
					"output_paths": {
						"type": "object",
						"additionalProperties": {"type": "string"}
					}
				}
			}
		}
	}`

	s, err := schema.NewSchema([]byte(schemaJSON))
	if err == nil {
		c.SetSchema(s)
	}
}
