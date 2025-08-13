package scheme

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestManager_GetCurrent_DefaultScheme(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "heimdall-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	manager := &Manager{
		schemesDir: filepath.Join(tempDir, "schemes"),
		stateDir:   filepath.Join(tempDir, "state"),
	}

	// Test getting current scheme when no state file exists (should return default)
	scheme, err := manager.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	// Verify default scheme
	if scheme.Name != "catppuccin" {
		t.Errorf("Expected default name 'catppuccin', got '%s'", scheme.Name)
	}
	if scheme.Flavour != "mocha" {
		t.Errorf("Expected default flavour 'mocha', got '%s'", scheme.Flavour)
	}
	if scheme.Mode != "dark" {
		t.Errorf("Expected default mode 'dark', got '%s'", scheme.Mode)
	}
	if len(scheme.Colours) == 0 {
		t.Error("Expected default colours to be populated")
	}
}

func TestManager_SetAndGetScheme(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "heimdall-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	manager := &Manager{
		schemesDir: filepath.Join(tempDir, "schemes"),
		stateDir:   filepath.Join(tempDir, "state"),
	}

	// Create test scheme
	testScheme := &Scheme{
		Name:    "test-scheme",
		Flavour: "test-flavour",
		Mode:    "light",
		Variant: "test-variant",
		Colours: map[string]string{
			"base": "ffffff",
			"text": "000000",
		},
	}

	// Set the scheme
	err = manager.SetScheme(testScheme)
	if err != nil {
		t.Fatalf("SetScheme failed: %v", err)
	}

	// Get the scheme back
	retrievedScheme, err := manager.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}

	// Verify the scheme was saved and retrieved correctly
	if retrievedScheme.Name != testScheme.Name {
		t.Errorf("Expected name '%s', got '%s'", testScheme.Name, retrievedScheme.Name)
	}
	if retrievedScheme.Flavour != testScheme.Flavour {
		t.Errorf("Expected flavour '%s', got '%s'", testScheme.Flavour, retrievedScheme.Flavour)
	}
	if retrievedScheme.Mode != testScheme.Mode {
		t.Errorf("Expected mode '%s', got '%s'", testScheme.Mode, retrievedScheme.Mode)
	}
	if retrievedScheme.Variant != testScheme.Variant {
		t.Errorf("Expected variant '%s', got '%s'", testScheme.Variant, retrievedScheme.Variant)
	}
	if retrievedScheme.Colours["base"] != testScheme.Colours["base"] {
		t.Errorf("Expected base colour '%s', got '%s'", testScheme.Colours["base"], retrievedScheme.Colours["base"])
	}
}

func TestManager_ListSchemes(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "heimdall-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	schemesDir := filepath.Join(tempDir, "schemes")

	// Create manager with temp directory
	manager := &Manager{
		schemesDir: schemesDir,
		stateDir:   filepath.Join(tempDir, "state"),
	}

	// Test empty schemes directory
	schemes, err := manager.ListSchemes()
	if err != nil {
		t.Fatalf("ListSchemes failed: %v", err)
	}
	if len(schemes) != 0 {
		t.Errorf("Expected empty schemes list, got %d schemes", len(schemes))
	}

	// Create test scheme directories
	testSchemes := []string{"catppuccin", "gruvbox", "nord"}
	for _, scheme := range testSchemes {
		err := os.MkdirAll(filepath.Join(schemesDir, scheme), 0755)
		if err != nil {
			t.Fatalf("Failed to create scheme directory: %v", err)
		}
	}

	// Test listing schemes
	schemes, err = manager.ListSchemes()
	if err != nil {
		t.Fatalf("ListSchemes failed: %v", err)
	}
	if len(schemes) != len(testSchemes) {
		t.Errorf("Expected %d schemes, got %d", len(testSchemes), len(schemes))
	}

	// Verify all test schemes are present
	schemeMap := make(map[string]bool)
	for _, scheme := range schemes {
		schemeMap[scheme] = true
	}
	for _, expected := range testSchemes {
		if !schemeMap[expected] {
			t.Errorf("Expected scheme '%s' not found in list", expected)
		}
	}
}

func TestManager_LoadScheme(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "heimdall-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	schemesDir := filepath.Join(tempDir, "schemes")

	// Create manager with temp directory
	manager := &Manager{
		schemesDir: schemesDir,
		stateDir:   filepath.Join(tempDir, "state"),
	}

	// Create test scheme file
	schemeDir := filepath.Join(schemesDir, "test-scheme", "test-flavour")
	err = os.MkdirAll(schemeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create scheme directory: %v", err)
	}

	testSchemeData := map[string]interface{}{
		"name":    "test-scheme",
		"flavour": "test-flavour",
		"mode":    "dark",
		"variant": "test-variant",
		"colours": map[string]string{
			"base": "1e1e2e",
			"text": "cdd6f4",
		},
	}

	schemeFile := filepath.Join(schemeDir, "dark.json")
	data, err := json.MarshalIndent(testSchemeData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test scheme: %v", err)
	}

	err = os.WriteFile(schemeFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test scheme file: %v", err)
	}

	// Test loading the scheme
	scheme, err := manager.LoadScheme("test-scheme", "test-flavour", "dark")
	if err != nil {
		t.Fatalf("LoadScheme failed: %v", err)
	}

	// Verify loaded scheme
	if scheme.Name != "test-scheme" {
		t.Errorf("Expected name 'test-scheme', got '%s'", scheme.Name)
	}
	if scheme.Flavour != "test-flavour" {
		t.Errorf("Expected flavour 'test-flavour', got '%s'", scheme.Flavour)
	}
	if scheme.Mode != "dark" {
		t.Errorf("Expected mode 'dark', got '%s'", scheme.Mode)
	}
	if scheme.Variant != "test-variant" {
		t.Errorf("Expected variant 'test-variant', got '%s'", scheme.Variant)
	}
	if scheme.Colours["base"] != "1e1e2e" {
		t.Errorf("Expected base colour '1e1e2e', got '%s'", scheme.Colours["base"])
	}
}

func TestManager_SaveScheme(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "heimdall-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	schemesDir := filepath.Join(tempDir, "schemes")

	// Create manager with temp directory
	manager := &Manager{
		schemesDir: schemesDir,
		stateDir:   filepath.Join(tempDir, "state"),
	}

	// Create test scheme
	testScheme := &Scheme{
		Name:    "saved-scheme",
		Flavour: "saved-flavour",
		Mode:    "light",
		Variant: "saved-variant",
		Colours: map[string]string{
			"base": "ffffff",
			"text": "000000",
		},
	}

	// Save the scheme
	err = manager.SaveScheme(testScheme)
	if err != nil {
		t.Fatalf("SaveScheme failed: %v", err)
	}

	// Verify the file was created
	expectedPath := filepath.Join(schemesDir, "saved-scheme", "saved-flavour", "light.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Expected scheme file was not created: %s", expectedPath)
	}

	// Load the scheme back and verify
	loadedScheme, err := manager.LoadScheme("saved-scheme", "saved-flavour", "light")
	if err != nil {
		t.Fatalf("Failed to load saved scheme: %v", err)
	}

	if loadedScheme.Name != testScheme.Name {
		t.Errorf("Expected name '%s', got '%s'", testScheme.Name, loadedScheme.Name)
	}
	if loadedScheme.Colours["base"] != testScheme.Colours["base"] {
		t.Errorf("Expected base colour '%s', got '%s'", testScheme.Colours["base"], loadedScheme.Colours["base"])
	}
}
