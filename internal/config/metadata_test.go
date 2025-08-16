package config

import (
	"strings"
	"testing"
)

func TestExtractMetadata(t *testing.T) {
	// Test with default config
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Check that we have metadata
	if len(metadata.Fields) == 0 {
		t.Error("No metadata fields extracted")
	}

	// Check specific fields exist and have descriptions
	testCases := []struct {
		path        string
		shouldExist bool
		hasDesc     bool
	}{
		{"version", true, true},
		{"theme.enableGtk", true, true},
		{"scheme.default", true, true},
		{"wallpaper.directory", true, true},
		{"screenshot.file_format", true, true},
		{"notification.enabled", true, true},
		{"external_tools.grim", true, true},
	}

	for _, tc := range testCases {
		fm, exists := metadata.GetFieldMetadata(tc.path)
		if exists != tc.shouldExist {
			t.Errorf("Field %s: expected exists=%v, got %v", tc.path, tc.shouldExist, exists)
			continue
		}

		if tc.hasDesc && fm.Description == "" {
			t.Errorf("Field %s has no description", tc.path)
		}
	}
}

func TestMetadataRegistry(t *testing.T) {
	// Initialize the registry
	err := InitializeRegistry()
	if err != nil {
		t.Fatalf("Failed to initialize registry: %v", err)
	}

	// Check that registry has fields
	fields := MetadataRegistry.GetAllFields()
	if len(fields) == 0 {
		t.Error("Registry has no fields")
	}

	// Test search functionality
	results := MetadataRegistry.SearchFields("theme")
	if len(results) == 0 {
		t.Error("Search for 'theme' returned no results")
	}

	// Test category extraction
	categories := MetadataRegistry.GetCategories()
	expectedCategories := []string{"version", "theme", "scheme", "wallpaper", "screenshot"}
	for _, expected := range expectedCategories {
		found := false
		for _, cat := range categories {
			if cat == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected category %s not found", expected)
		}
	}
}

func TestFieldTypeExtraction(t *testing.T) {
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Test various field types
	typeTests := []struct {
		path         string
		expectedType string
	}{
		{"version", "string"},
		{"theme.enableGtk", "bool"},
		{"shell.daemon_port", "int"},
		{"wallpaper.threshold", "float"},
		{"shell.args", "[]string"},
		{"wallpaper.extensions", "[]string"},
	}

	for _, tt := range typeTests {
		fm, exists := metadata.GetFieldMetadata(tt.path)
		if !exists {
			t.Errorf("Field %s not found", tt.path)
			continue
		}

		if fm.Type != tt.expectedType {
			t.Errorf("Field %s: expected type %s, got %s", tt.path, tt.expectedType, fm.Type)
		}
	}
}

func TestMetadataCompleteness(t *testing.T) {
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Check for missing descriptions
	missing := metadata.ValidateCompleteness()
	if len(missing) > 0 {
		// We expect some fields might not have descriptions yet
		// but let's make sure critical ones do
		criticalFields := []string{
			"version",
			"theme.enableGtk",
			"scheme.default",
			"wallpaper.directory",
		}

		for _, critical := range criticalFields {
			for _, m := range missing {
				if m == critical {
					t.Errorf("Critical field %s is missing description", critical)
				}
			}
		}
	}
}

func TestDocumentationGeneration(t *testing.T) {
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Generate documentation
	docs := metadata.GenerateDocumentation()

	// Check that documentation contains expected sections
	expectedSections := []string{
		"# Configuration Reference",
		"## version",
		"## theme",
		"## scheme",
		"## wallpaper",
	}

	for _, section := range expectedSections {
		if !strings.Contains(docs, section) {
			t.Errorf("Documentation missing section: %s", section)
		}
	}

	// Check that descriptions are included
	if !strings.Contains(docs, "Theme application settings") {
		t.Error("Documentation missing theme description")
	}
}

func TestJSONSchemaGeneration(t *testing.T) {
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Generate JSON schema
	schema := metadata.GenerateJSONSchema()

	// Check basic schema structure
	if schema["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Error("Invalid schema version")
	}

	if schema["type"] != "object" {
		t.Error("Root type should be object")
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Schema missing properties")
	}

	// Check that version field exists
	if _, exists := properties["version"]; !exists {
		t.Error("Schema missing version property")
	}
}

func TestGetFieldsByPrefix(t *testing.T) {
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Get all theme fields
	themeFields := metadata.GetFieldsByPrefix("theme.")
	if len(themeFields) == 0 {
		t.Error("No theme fields found")
	}

	// Check that all returned fields have the prefix
	for path := range themeFields {
		if !strings.HasPrefix(path, "theme.") {
			t.Errorf("Field %s doesn't have theme. prefix", path)
		}
	}
}

func TestGetFieldsByType(t *testing.T) {
	cfg := getDefaults()
	metadata, err := ExtractMetadata(cfg)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Get all boolean fields
	boolFields := metadata.GetFieldsByType("bool")
	if len(boolFields) == 0 {
		t.Error("No boolean fields found")
	}

	// Check that all returned fields are actually booleans
	for path, fm := range boolFields {
		if fm.Type != "bool" {
			t.Errorf("Field %s is not a boolean: %s", path, fm.Type)
		}
	}
}
