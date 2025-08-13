package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// Schema represents a JSON schema with metadata
type Schema struct {
	Raw         json.RawMessage        `json:"-"`
	Schema      string                 `json:"$schema,omitempty"`
	ID          string                 `json:"$id,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Properties  map[string]*Property   `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Version     string                 `json:"version,omitempty"`
	Metadata    map[string]interface{} `json:"x-heimdall-metadata,omitempty"`
}

// Property represents a schema property
type Property struct {
	Type                 interface{}          `json:"type,omitempty"`
	Description          string               `json:"description,omitempty"`
	Default              interface{}          `json:"default,omitempty"`
	Enum                 []interface{}        `json:"enum,omitempty"`
	Properties           map[string]*Property `json:"properties,omitempty"`
	Items                *Property            `json:"items,omitempty"`
	Minimum              *float64             `json:"minimum,omitempty"`
	Maximum              *float64             `json:"maximum,omitempty"`
	MinLength            *int                 `json:"minLength,omitempty"`
	MaxLength            *int                 `json:"maxLength,omitempty"`
	Pattern              string               `json:"pattern,omitempty"`
	Required             []string             `json:"required,omitempty"`
	AdditionalProperties interface{}          `json:"additionalProperties,omitempty"`
	Format               string               `json:"format,omitempty"`
	Ref                  string               `json:"$ref,omitempty"`
}

// NewSchema creates a new schema from JSON data
func NewSchema(data []byte) (*Schema, error) {
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}
	schema.Raw = json.RawMessage(data)
	return &schema, nil
}

// LoadFromFile loads a schema from a JSON file
func LoadFromFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}
	return NewSchema(data)
}

// LoadFromURL loads a schema from a URL (placeholder for future implementation)
func LoadFromURL(url string) (*Schema, error) {
	// TODO: Implement HTTP fetching
	return nil, fmt.Errorf("URL schema loading not yet implemented")
}

// Validate validates a configuration against the schema
func (s *Schema) Validate(config map[string]interface{}) error {
	if s.Type != "" && s.Type != "object" {
		return fmt.Errorf("root schema must be of type 'object'")
	}

	// Check required fields
	for _, req := range s.Required {
		if _, exists := config[req]; !exists {
			return fmt.Errorf("required field '%s' is missing", req)
		}
	}

	// Validate each property
	for key, value := range config {
		prop, exists := s.Properties[key]
		if !exists {
			// Check if additional properties are allowed
			if s.Properties != nil && len(s.Properties) > 0 {
				// If properties are defined but this key isn't in them, it might be invalid
				// This depends on additionalProperties setting
				continue
			}
		}

		if prop != nil {
			if err := validateValue(value, prop, key); err != nil {
				return err
			}
		}
	}

	return nil
}

// ValidateValue validates a single value against its schema path
func (s *Schema) ValidateValue(path string, value interface{}) error {
	prop, err := s.GetProperty(path)
	if err != nil {
		return err
	}
	return validateValue(value, prop, path)
}

// GetProperty retrieves a property by path (e.g., "appearance.theme")
func (s *Schema) GetProperty(path string) (*Property, error) {
	parts := strings.Split(path, ".")
	props := s.Properties

	for i, part := range parts {
		prop, exists := props[part]
		if !exists {
			return nil, fmt.Errorf("property '%s' not found in schema", strings.Join(parts[:i+1], "."))
		}

		if i == len(parts)-1 {
			return prop, nil
		}

		if prop.Properties == nil {
			return nil, fmt.Errorf("property '%s' is not an object", strings.Join(parts[:i+1], "."))
		}

		props = prop.Properties
	}

	return nil, fmt.Errorf("property '%s' not found", path)
}

// validateValue validates a value against a property schema
func validateValue(value interface{}, prop *Property, path string) error {
	if value == nil {
		// Check if this is allowed (nullable)
		return nil
	}

	// Handle type validation
	if prop.Type != nil {
		if err := validateType(value, prop.Type, path); err != nil {
			return err
		}
	}

	// Check enum values
	if len(prop.Enum) > 0 {
		found := false
		for _, enumVal := range prop.Enum {
			if reflect.DeepEqual(value, enumVal) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value at '%s' must be one of %v, got %v", path, prop.Enum, value)
		}
	}

	// Type-specific validations
	switch v := value.(type) {
	case string:
		if err := validateString(v, prop, path); err != nil {
			return err
		}
	case float64:
		if err := validateNumber(v, prop, path); err != nil {
			return err
		}
	case int:
		if err := validateNumber(float64(v), prop, path); err != nil {
			return err
		}
	case []interface{}:
		if err := validateArray(v, prop, path); err != nil {
			return err
		}
	case map[string]interface{}:
		if err := validateObject(v, prop, path); err != nil {
			return err
		}
	}

	return nil
}

// validateType checks if a value matches the expected type
func validateType(value interface{}, expectedType interface{}, path string) error {
	// Handle multiple types (e.g., ["string", "null"])
	switch t := expectedType.(type) {
	case string:
		if !matchesType(value, t) {
			return fmt.Errorf("value at '%s' must be of type %s, got %T", path, t, value)
		}
	case []interface{}:
		matched := false
		for _, typ := range t {
			if typStr, ok := typ.(string); ok && matchesType(value, typStr) {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("value at '%s' must be one of types %v, got %T", path, t, value)
		}
	}
	return nil
}

// matchesType checks if a value matches a JSON schema type
func matchesType(value interface{}, jsonType string) bool {
	switch jsonType {
	case "null":
		return value == nil
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "number":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			return true
		}
		return false
	case "integer":
		switch v := value.(type) {
		case float64:
			return v == float64(int(v))
		case int, int32, int64:
			return true
		}
		return false
	case "string":
		_, ok := value.(string)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return false
	}
}

// validateString validates string-specific constraints
func validateString(value string, prop *Property, path string) error {
	if prop.MinLength != nil && len(value) < *prop.MinLength {
		return fmt.Errorf("string at '%s' must have at least %d characters", path, *prop.MinLength)
	}
	if prop.MaxLength != nil && len(value) > *prop.MaxLength {
		return fmt.Errorf("string at '%s' must have at most %d characters", path, *prop.MaxLength)
	}
	if prop.Pattern != "" {
		matched, err := regexp.MatchString(prop.Pattern, value)
		if err != nil {
			return fmt.Errorf("invalid pattern for '%s': %w", path, err)
		}
		if !matched {
			return fmt.Errorf("string at '%s' does not match pattern %s", path, prop.Pattern)
		}
	}
	return nil
}

// validateNumber validates number-specific constraints
func validateNumber(value float64, prop *Property, path string) error {
	if prop.Minimum != nil && value < *prop.Minimum {
		return fmt.Errorf("number at '%s' must be >= %f", path, *prop.Minimum)
	}
	if prop.Maximum != nil && value > *prop.Maximum {
		return fmt.Errorf("number at '%s' must be <= %f", path, *prop.Maximum)
	}
	return nil
}

// validateArray validates array-specific constraints
func validateArray(value []interface{}, prop *Property, path string) error {
	if prop.Items != nil {
		for i, item := range value {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			if err := validateValue(item, prop.Items, itemPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateObject validates object-specific constraints
func validateObject(value map[string]interface{}, prop *Property, path string) error {
	// Check required fields
	for _, req := range prop.Required {
		if _, exists := value[req]; !exists {
			return fmt.Errorf("required field '%s.%s' is missing", path, req)
		}
	}

	// Validate properties
	if prop.Properties != nil {
		for key, val := range value {
			if subProp, exists := prop.Properties[key]; exists {
				subPath := fmt.Sprintf("%s.%s", path, key)
				if err := validateValue(val, subProp, subPath); err != nil {
					return err
				}
			} else if prop.AdditionalProperties == false {
				return fmt.Errorf("additional property '%s.%s' is not allowed", path, key)
			}
		}
	}

	return nil
}

// ToJSON converts the schema to JSON
func (s *Schema) ToJSON() ([]byte, error) {
	if s.Raw != nil {
		return s.Raw, nil
	}
	return json.MarshalIndent(s, "", "  ")
}

// GetType returns the type of a property as a string
func (p *Property) GetType() string {
	switch t := p.Type.(type) {
	case string:
		return t
	case []interface{}:
		if len(t) > 0 {
			if s, ok := t[0].(string); ok {
				return s
			}
		}
	}
	return "unknown"
}
