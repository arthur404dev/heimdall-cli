package providers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arthur404dev/heimdall-cli/internal/config/schema"
	"github.com/arthur404dev/heimdall-cli/internal/config/types"
)

// ShellProvider manages shell configuration with external schema support
type ShellProvider struct {
	*BaseProvider
	paths          *types.ConfigPaths
	externalSchema string // Path to external schema file
	outputPath     string // Path where the shell config should be written
}

// NewShellProvider creates a new shell configuration provider
func NewShellProvider(configPath string, paths *types.ConfigPaths) Provider {
	provider := &ShellProvider{
		BaseProvider: NewBaseProvider("shell", configPath),
		paths:        paths,
	}

	// Set default external schema path
	provider.externalSchema = os.Getenv("HEIMDALL_SHELL_SCHEMA")
	if provider.externalSchema == "" {
		// Default to quickshell config location
		provider.externalSchema = filepath.Join(os.Getenv("HOME"), ".config", "quickshell", "config", "default.json")
	}

	// Set default output path
	provider.outputPath = os.Getenv("HEIMDALL_SHELL_OUTPUT")
	if provider.outputPath == "" && paths != nil && paths.OutputPaths != nil {
		provider.outputPath = paths.OutputPaths["shell"]
	}
	if provider.outputPath == "" {
		// Default to quickshell config location
		provider.outputPath = filepath.Join(os.Getenv("HOME"), ".config", "quickshell", "config", "default.json")
	}

	// Try to load external schema
	provider.loadExternalSchema()
	return provider
}

// loadExternalSchema attempts to load schema from external source
func (s *ShellProvider) loadExternalSchema() error {
	// First, try to load from the external schema file
	if s.externalSchema != "" && fileExists(s.externalSchema) {
		data, err := os.ReadFile(s.externalSchema)
		if err != nil {
			return fmt.Errorf("failed to read external schema: %w", err)
		}

		// Check if it's a config file with embedded schema
		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err == nil {
			// Look for $schema field
			if schemaField, ok := config["$schema"].(string); ok && schemaField != "" {
				// This is a config file with schema reference
				// Try to extract schema from the config structure
				if err := s.extractSchemaFromConfig(config); err == nil {
					return nil
				}
			}

			// Look for embedded schema in x-schema or similar fields
			if embeddedSchema, ok := config["x-schema"].(map[string]interface{}); ok {
				schemaData, _ := json.Marshal(embeddedSchema)
				if sch, err := schema.NewSchema(schemaData); err == nil {
					s.SetSchema(sch)
					return nil
				}
			}
		}

		// Try to parse as pure schema
		if sch, err := schema.NewSchema(data); err == nil {
			s.SetSchema(sch)
			return nil
		}
	}

	// Fall back to default schema
	s.initDefaultSchema()
	return nil
}

// extractSchemaFromConfig attempts to extract schema from a config file
func (s *ShellProvider) extractSchemaFromConfig(config map[string]interface{}) error {
	// Build a schema from the config structure
	props := make(map[string]*schema.Property)

	for key, value := range config {
		if key == "$schema" || key == "x-schema" {
			continue
		}
		props[key] = inferProperty(value)
	}

	sch := &schema.Schema{
		Schema:     "http://json-schema.org/draft-07/schema#",
		Title:      "Shell Configuration",
		Type:       "object",
		Properties: props,
	}

	s.SetSchema(sch)
	return nil
}

// inferProperty infers a schema property from a value
func inferProperty(value interface{}) *schema.Property {
	prop := &schema.Property{}

	switch v := value.(type) {
	case bool:
		prop.Type = "boolean"
		prop.Default = v
	case float64:
		if v == float64(int(v)) {
			prop.Type = "integer"
		} else {
			prop.Type = "number"
		}
		prop.Default = v
	case string:
		prop.Type = "string"
		prop.Default = v
	case []interface{}:
		prop.Type = "array"
		if len(v) > 0 {
			prop.Items = inferProperty(v[0])
		}
	case map[string]interface{}:
		prop.Type = "object"
		props := make(map[string]*schema.Property)
		for k, val := range v {
			props[k] = inferProperty(val)
		}
		prop.Properties = props
	case nil:
		prop.Type = []interface{}{"null", "string"} // Allow null or string
	default:
		prop.Type = "string" // Default to string
	}

	return prop
}

// initDefaultSchema initializes a default shell configuration schema
func (s *ShellProvider) initDefaultSchema() {
	schemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title": "Shell Configuration",
		"description": "Configuration for shell integration",
		"type": "object",
		"properties": {
			"version": {
				"type": "string",
				"description": "Configuration version",
				"default": "1.0.0"
			},
			"appearance": {
				"type": "object",
				"description": "Visual appearance settings",
				"properties": {
					"colorScheme": {
						"type": "string",
						"description": "Color scheme identifier",
						"default": "catppuccin-mocha"
					},
					"fontSize": {
						"type": "integer",
						"description": "Base font size in pixels",
						"minimum": 8,
						"maximum": 32,
						"default": 12
					},
					"animations": {
						"type": "boolean",
						"description": "Enable UI animations",
						"default": true
					}
				}
			},
			"bar": {
				"type": "object",
				"description": "Status bar configuration",
				"properties": {
					"position": {
						"type": "string",
						"enum": ["top", "bottom"],
						"default": "top"
					},
					"height": {
						"type": "integer",
						"minimum": 20,
						"maximum": 100,
						"default": 30
					},
					"modules": {
						"type": "array",
						"description": "Active bar modules",
						"items": {
							"type": "object",
							"properties": {
								"name": {
									"type": "string",
									"description": "Module identifier"
								},
								"enabled": {
									"type": "boolean",
									"default": true
								},
								"config": {
									"type": "object",
									"description": "Module-specific configuration",
									"additionalProperties": true
								}
							},
							"required": ["name"]
						},
						"default": []
					}
				}
			}
		}
	}`

	if sch, err := schema.NewSchema([]byte(schemaJSON)); err == nil {
		s.SetSchema(sch)
	}
}

// Save writes the configuration to both internal and external locations
func (s *ShellProvider) Save() error {
	// Save to internal location first
	if err := s.BaseProvider.Save(); err != nil {
		return err
	}

	// If output path is configured and different from config path, save there too
	if s.outputPath != "" && s.outputPath != s.configPath {
		s.mu.RLock()
		data, err := json.MarshalIndent(s.data, "", "  ")
		s.mu.RUnlock()

		if err != nil {
			return fmt.Errorf("failed to marshal config for output: %w", err)
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(s.outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write to output location
		if err := os.WriteFile(s.outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write to output path: %w", err)
		}
	}

	return nil
}

// SetExternalSchema sets the path to the external schema file
func (s *ShellProvider) SetExternalSchema(path string) error {
	s.externalSchema = path
	return s.loadExternalSchema()
}

// SetOutputPath sets the output path for the shell configuration
func (s *ShellProvider) SetOutputPath(path string) {
	s.outputPath = path
}

// GetOutputPath returns the output path for the shell configuration
func (s *ShellProvider) GetOutputPath() string {
	return s.outputPath
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
