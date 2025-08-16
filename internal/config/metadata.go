package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// FieldMetadata holds metadata extracted from struct tags
type FieldMetadata struct {
	Name        string   // Field name
	Path        string   // Full JSON path (e.g., "theme.enableGtk")
	Type        string   // Field type (bool, string, int, etc.)
	Description string   // Description from 'desc' tag
	Default     string   // Default value from 'default' tag
	Example     string   // Example value from 'example' tag
	Deprecated  string   // Deprecation message from 'deprecated' tag
	JSONName    string   // JSON field name
	Required    bool     // Whether field is required
	Children    []string // Child field paths for nested structs
}

// ConfigMetadata holds all configuration metadata
type ConfigMetadata struct {
	Fields map[string]*FieldMetadata // Map of path to metadata
	mu     sync.RWMutex
}

// MetadataRegistry is the global registry for config metadata
var MetadataRegistry = &ConfigMetadata{
	Fields: make(map[string]*FieldMetadata),
}

// ExtractMetadata extracts metadata from a struct using reflection
func ExtractMetadata(v interface{}) (*ConfigMetadata, error) {
	metadata := &ConfigMetadata{
		Fields: make(map[string]*FieldMetadata),
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	rt := rv.Type()
	extractStructMetadata(rt, rv, "", "", metadata)

	return metadata, nil
}

// extractStructMetadata recursively extracts metadata from struct fields
func extractStructMetadata(rt reflect.Type, rv reflect.Value, parentPath, parentJSONPath string, metadata *ConfigMetadata) {
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Tag.Get("mapstructure")
		}
		if jsonTag == "-" {
			continue
		}

		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		// Build paths
		fieldPath := field.Name
		if parentPath != "" {
			fieldPath = parentPath + "." + field.Name
		}

		jsonPath := jsonName
		if parentJSONPath != "" {
			jsonPath = parentJSONPath + "." + jsonName
		}

		// Extract metadata from tags
		fm := &FieldMetadata{
			Name:        field.Name,
			Path:        jsonPath,
			Type:        getFieldType(field.Type),
			Description: field.Tag.Get("desc"),
			Default:     field.Tag.Get("default"),
			Example:     field.Tag.Get("example"),
			Deprecated:  field.Tag.Get("deprecated"),
			JSONName:    jsonName,
			Required:    strings.Contains(jsonTag, "required"),
			Children:    []string{},
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct && !isTimeType(field.Type) {
			// Add the struct field itself
			metadata.Fields[jsonPath] = fm

			// Extract nested fields
			nestedValue := fieldValue
			if !nestedValue.IsValid() {
				nestedValue = reflect.New(field.Type).Elem()
			}
			extractStructMetadata(field.Type, nestedValue, fieldPath, jsonPath, metadata)

			// Collect children paths
			for path := range metadata.Fields {
				if strings.HasPrefix(path, jsonPath+".") && path != jsonPath {
					fm.Children = append(fm.Children, path)
				}
			}
		} else if field.Type.Kind() == reflect.Map {
			// For maps, just add the field metadata
			metadata.Fields[jsonPath] = fm
		} else if field.Type.Kind() == reflect.Slice {
			// For slices, add the field metadata
			metadata.Fields[jsonPath] = fm
		} else {
			// Regular field
			metadata.Fields[jsonPath] = fm
		}
	}
}

// getFieldType returns a human-readable type name
func getFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Slice:
		elemType := getFieldType(t.Elem())
		return "[]" + elemType
	case reflect.Map:
		keyType := getFieldType(t.Key())
		valueType := getFieldType(t.Elem())
		return "map[" + keyType + "]" + valueType
	case reflect.Struct:
		if isTimeType(t) {
			return "time"
		}
		return "object"
	case reflect.Ptr:
		return getFieldType(t.Elem())
	default:
		return t.String()
	}
}

// isTimeType checks if a type is time.Time or similar
func isTimeType(t reflect.Type) bool {
	return t.PkgPath() == "time" && t.Name() == "Time"
}

// InitializeRegistry initializes the global metadata registry
func InitializeRegistry() error {
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	MetadataRegistry.mu.Lock()
	MetadataRegistry.Fields = metadata.Fields
	MetadataRegistry.mu.Unlock()

	return nil
}

// GetFieldMetadata retrieves metadata for a specific field path
func (cm *ConfigMetadata) GetFieldMetadata(path string) (*FieldMetadata, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	fm, exists := cm.Fields[path]
	return fm, exists
}

// GetAllFields returns all field metadata
func (cm *ConfigMetadata) GetAllFields() map[string]*FieldMetadata {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*FieldMetadata, len(cm.Fields))
	for k, v := range cm.Fields {
		result[k] = v
	}
	return result
}

// GetFieldsByPrefix returns all fields with a given path prefix
func (cm *ConfigMetadata) GetFieldsByPrefix(prefix string) map[string]*FieldMetadata {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*FieldMetadata)
	for path, fm := range cm.Fields {
		if strings.HasPrefix(path, prefix) {
			result[path] = fm
		}
	}
	return result
}

// GetFieldsByType returns all fields of a specific type
func (cm *ConfigMetadata) GetFieldsByType(fieldType string) map[string]*FieldMetadata {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make(map[string]*FieldMetadata)
	for path, fm := range cm.Fields {
		if fm.Type == fieldType {
			result[path] = fm
		}
	}
	return result
}

// SearchFields searches for fields by description or name
func (cm *ConfigMetadata) SearchFields(query string) map[string]*FieldMetadata {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	query = strings.ToLower(query)
	result := make(map[string]*FieldMetadata)

	for path, fm := range cm.Fields {
		if strings.Contains(strings.ToLower(fm.Name), query) ||
			strings.Contains(strings.ToLower(fm.Description), query) ||
			strings.Contains(strings.ToLower(path), query) {
			result[path] = fm
		}
	}
	return result
}

// GetCategories returns top-level configuration categories
func (cm *ConfigMetadata) GetCategories() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	categories := make(map[string]bool)
	for path := range cm.Fields {
		parts := strings.Split(path, ".")
		if len(parts) > 0 {
			categories[parts[0]] = true
		}
	}

	result := make([]string, 0, len(categories))
	for cat := range categories {
		result = append(result, cat)
	}
	return result
}

// ValidateCompleteness checks if all struct fields have descriptions
func (cm *ConfigMetadata) ValidateCompleteness() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var missing []string
	for path, fm := range cm.Fields {
		// Skip nested struct containers (they don't need descriptions)
		if fm.Type == "object" && len(fm.Children) > 0 {
			continue
		}

		if fm.Description == "" {
			missing = append(missing, path)
		}
	}
	return missing
}

// GenerateDocumentation generates markdown documentation from metadata
func (cm *ConfigMetadata) GenerateDocumentation() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("# Configuration Reference\n\n")
	sb.WriteString("This document describes all available configuration options for heimdall-cli.\n\n")

	// Group by category
	categories := make(map[string][]*FieldMetadata)
	for path, fm := range cm.Fields {
		parts := strings.Split(path, ".")
		if len(parts) > 0 {
			category := parts[0]
			categories[category] = append(categories[category], fm)
		}
	}

	// Write each category
	for category, fields := range categories {
		// Capitalize first letter of category
		categoryTitle := category
		if len(category) > 0 {
			categoryTitle = strings.ToUpper(category[:1]) + category[1:]
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n", categoryTitle))

		for _, fm := range fields {
			// Skip container objects
			if fm.Type == "object" && len(fm.Children) > 0 {
				continue
			}

			sb.WriteString(fmt.Sprintf("### `%s`\n\n", fm.Path))

			if fm.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", fm.Description))
			}

			sb.WriteString(fmt.Sprintf("- **Type:** `%s`\n", fm.Type))

			if fm.Default != "" {
				sb.WriteString(fmt.Sprintf("- **Default:** `%s`\n", fm.Default))
			}

			if fm.Example != "" {
				sb.WriteString(fmt.Sprintf("- **Example:** `%s`\n", fm.Example))
			}

			if fm.Deprecated != "" {
				sb.WriteString(fmt.Sprintf("- **⚠️ Deprecated:** %s\n", fm.Deprecated))
			}

			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// GenerateJSONSchema generates a JSON schema from metadata
func (cm *ConfigMetadata) GenerateJSONSchema() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	schema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"title":       "Heimdall CLI Configuration",
		"description": "Configuration schema for heimdall-cli",
		"type":        "object",
		"properties":  make(map[string]interface{}),
	}

	properties := schema["properties"].(map[string]interface{})

	// Build nested structure
	for path, fm := range cm.Fields {
		parts := strings.Split(path, ".")
		current := properties

		for i, part := range parts {
			if i == len(parts)-1 {
				// Leaf node
				prop := map[string]interface{}{
					"description": fm.Description,
				}

				// Set type
				switch fm.Type {
				case "bool":
					prop["type"] = "boolean"
				case "int", "uint":
					prop["type"] = "integer"
				case "float":
					prop["type"] = "number"
				case "string":
					prop["type"] = "string"
				case "[]string":
					prop["type"] = "array"
					prop["items"] = map[string]string{"type": "string"}
				default:
					if strings.HasPrefix(fm.Type, "[]") {
						prop["type"] = "array"
					} else if strings.HasPrefix(fm.Type, "map") {
						prop["type"] = "object"
					} else {
						prop["type"] = "object"
					}
				}

				// Add default if present
				if fm.Default != "" {
					prop["default"] = fm.Default
				}

				// Add example if present
				if fm.Example != "" {
					prop["examples"] = []string{fm.Example}
				}

				current[part] = prop
			} else {
				// Intermediate node
				if _, exists := current[part]; !exists {
					current[part] = map[string]interface{}{
						"type":       "object",
						"properties": make(map[string]interface{}),
					}
				}
				// Safely navigate to nested properties
				if propMap, ok := current[part].(map[string]interface{}); ok {
					if props, ok := propMap["properties"].(map[string]interface{}); ok {
						current = props
					} else {
						// Create properties if they don't exist
						propMap["properties"] = make(map[string]interface{})
						current = propMap["properties"].(map[string]interface{})
					}
				}
			}
		}
	}

	return schema
}
