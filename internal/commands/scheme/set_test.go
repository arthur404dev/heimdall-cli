package scheme

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
)

func TestSetCommand(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "creates set command with correct properties",
			test: testSetCommandCreation,
		},
		{
			name: "sets scheme with positional arguments",
			test: testSetCommandPositionalArgs,
		},
		{
			name: "sets scheme with flags",
			test: testSetCommandFlags,
		},
		{
			name: "sets random scheme with -r flag",
			test: testSetCommandRandomFlag,
		},
		{
			name: "sets scheme without applying theme",
			test: testSetCommandNoApply,
		},
		{
			name: "sets scheme with notifications",
			test: testSetCommandWithNotifications,
		},
		{
			name: "handles scheme not found error",
			test: testSetCommandSchemeNotFound,
		},
		{
			name: "handles no flavours available error",
			test: testSetCommandNoFlavours,
		},
		{
			name: "handles invalid mode error",
			test: testSetCommandInvalidMode,
		},
		{
			name: "handles theme application error",
			test: testSetCommandThemeError,
		},
		{
			name: "uses default flavour when not specified",
			test: testSetCommandDefaultFlavour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testSetCommandCreation(t *testing.T) {
	// Act
	cmd := setCommand()

	// Assert
	if cmd == nil {
		t.Fatal("Expected command to be created, got nil")
	}
	if cmd.Use != "set [scheme] [flavour] [mode]" {
		t.Errorf("Expected Use to be 'set [scheme] [flavour] [mode]', got '%s'", cmd.Use)
	}
	if cmd.Short != "Set the active color scheme" {
		t.Errorf("Expected Short to be 'Set the active color scheme', got '%s'", cmd.Short)
	}
	if !strings.Contains(cmd.Long, "Set the active color scheme and apply theme") {
		t.Error("Expected Long description to contain 'Set the active color scheme and apply theme'")
	}
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Check flags
	flags := cmd.Flags()
	expectedFlags := []string{"no-apply", "name", "flavour", "mode", "variant", "random", "notify"}
	for _, flagName := range expectedFlags {
		if flags.Lookup(flagName) == nil {
			t.Errorf("Expected --%s flag to be present", flagName)
		}
	}
}

func testSetCommandPositionalArgs(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "mocha",
		Mode:    "dark",
		Colours: map[string]string{"base": "1e1e2e", "text": "cdd6f4"},
	}
	mockManager.AddSchemeData("catppuccin", "mocha", "dark", testScheme)

	cmd := setCommand()
	cmd.SetArgs([]string{"catppuccin", "mocha", "dark"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	originalApplyTheme := applyThemeFunc
	defer func() { applyThemeFunc = originalApplyTheme }()
	applyThemeFunc = func(*scheme.Scheme) error {
		return nil
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify scheme was set
	currentScheme, _ := mockManager.GetCurrent()
	if currentScheme.Name != "catppuccin" {
		t.Errorf("Expected current scheme name to be 'catppuccin', got '%s'", currentScheme.Name)
	}
	if currentScheme.Flavour != "mocha" {
		t.Errorf("Expected current scheme flavour to be 'mocha', got '%s'", currentScheme.Flavour)
	}
	if currentScheme.Mode != "dark" {
		t.Errorf("Expected current scheme mode to be 'dark', got '%s'", currentScheme.Mode)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Scheme set to catppuccin/mocha/dark") {
		t.Error("Expected output to contain success message")
	}
}

func testSetCommandFlags(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	currentScheme := &scheme.Scheme{
		Name:    "gruvbox",
		Flavour: "hard",
		Mode:    "light",
		Variant: "tonalspot",
	}
	mockManager.SetCurrentScheme(currentScheme)

	testScheme := &scheme.Scheme{
		Name:    "catppuccin",
		Flavour: "latte",
		Mode:    "light",
		Colours: map[string]string{"base": "eff1f5", "text": "4c4f69"},
	}
	mockManager.AddSchemeData("catppuccin", "latte", "light", testScheme)

	cmd := setCommand()
	cmd.SetArgs([]string{"-n", "catppuccin", "-f", "latte", "-m", "light", "-v", "vibrant"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	originalApplyTheme := applyThemeFunc
	defer func() { applyThemeFunc = originalApplyTheme }()
	applyThemeFunc = func(*scheme.Scheme) error {
		return nil
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify scheme was set with correct values
	newCurrentScheme, _ := mockManager.GetCurrent()
	if newCurrentScheme.Name != "catppuccin" {
		t.Errorf("Expected current scheme name to be 'catppuccin', got '%s'", newCurrentScheme.Name)
	}
	if newCurrentScheme.Variant != "vibrant" {
		t.Errorf("Expected current scheme variant to be 'vibrant', got '%s'", newCurrentScheme.Variant)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Scheme set to catppuccin/latte/light") {
		t.Error("Expected output to contain success message")
	}
	if !strings.Contains(outputStr, "Variant: vibrant") {
		t.Error("Expected output to contain variant information")
	}
}

func testSetCommandRandomFlag(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.AddScheme("scheme1", []string{"flavour1"})
	mockManager.AddScheme("scheme2", []string{"flavour2"})
	mockManager.AddFlavour("scheme1", "flavour1", []string{"dark", "light"})
	mockManager.AddFlavour("scheme2", "flavour2", []string{"dark"})

	testScheme := &scheme.Scheme{
		Name:    "scheme1",
		Flavour: "flavour1",
		Mode:    "dark",
		Colours: map[string]string{"base": "000000", "text": "ffffff"},
	}
	mockManager.AddSchemeData("scheme1", "flavour1", "dark", testScheme)

	cmd := setCommand()
	cmd.SetArgs([]string{"-r"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	originalApplyTheme := applyThemeFunc
	defer func() { applyThemeFunc = originalApplyTheme }()
	applyThemeFunc = func(*scheme.Scheme) error {
		return nil
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Random scheme set to") {
		t.Error("Expected output to contain random scheme message")
	}
}

func testSetCommandNoApply(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "rosepine",
		Flavour: "main",
		Mode:    "dark",
		Colours: map[string]string{"base": "191724", "text": "e0def4"},
	}
	mockManager.AddSchemeData("rosepine", "main", "dark", testScheme)

	cmd := setCommand()
	cmd.SetArgs([]string{"rosepine", "main", "dark", "--no-apply"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	themeApplied := false
	originalApplyTheme := applyThemeFunc
	defer func() { applyThemeFunc = originalApplyTheme }()
	applyThemeFunc = func(*scheme.Scheme) error {
		themeApplied = true
		return nil
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if themeApplied {
		t.Error("Expected theme not to be applied with --no-apply flag")
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "Scheme set to rosepine/main/dark") {
		t.Error("Expected output to contain success message")
	}
}

func testSetCommandWithNotifications(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "onedark",
		Flavour: "default",
		Mode:    "dark",
		Colours: map[string]string{"base": "282c34", "text": "abb2bf"},
	}
	mockManager.AddSchemeData("onedark", "default", "dark", testScheme)

	cmd := setCommand()
	cmd.SetArgs([]string{"onedark", "default", "dark", "--notify"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	originalApplyTheme := applyThemeFunc
	defer func() { applyThemeFunc = originalApplyTheme }()
	applyThemeFunc = func(*scheme.Scheme) error {
		return nil
	}

	notificationSent := false
	originalSendNotification := sendNotificationFunc
	defer func() { sendNotificationFunc = originalSendNotification }()
	sendNotificationFunc = func(summary, body string) error {
		notificationSent = true
		if !strings.Contains(summary, "Scheme Changed") {
			t.Errorf("Expected notification summary to contain 'Scheme Changed', got '%s'", summary)
		}
		if !strings.Contains(body, "onedark/default/dark") {
			t.Errorf("Expected notification body to contain scheme info, got '%s'", body)
		}
		return nil
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !notificationSent {
		t.Error("Expected notification to be sent with --notify flag")
	}
}

func testSetCommandSchemeNotFound(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.SetError("LoadSchemeWithFallback", &mockError{msg: "scheme not found"})

	cmd := setCommand()
	cmd.SetArgs([]string{"nonexistent", "flavour", "dark"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error for nonexistent scheme")
	}
	if !strings.Contains(err.Error(), "failed to load scheme") {
		t.Errorf("Expected error to contain 'failed to load scheme', got: %v", err)
	}
}

func testSetCommandNoFlavours(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.SetError("ListFlavours", &mockError{msg: "no flavours available"})

	cmd := setCommand()
	cmd.SetArgs([]string{"test-scheme"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error when no flavours available")
	}
	if !strings.Contains(err.Error(), "failed to list flavours") {
		t.Errorf("Expected error to contain 'failed to list flavours', got: %v", err)
	}
}

func testSetCommandInvalidMode(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()

	cmd := setCommand()
	cmd.SetArgs([]string{"test-scheme", "test-flavour", "invalid-mode"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error for invalid mode")
	}
	if !strings.Contains(err.Error(), "invalid mode") {
		t.Errorf("Expected error to contain 'invalid mode', got: %v", err)
	}
	if !strings.Contains(err.Error(), "must be 'dark' or 'light'") {
		t.Errorf("Expected error to contain mode options, got: %v", err)
	}
}

func testSetCommandThemeError(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	testScheme := &scheme.Scheme{
		Name:    "test-scheme",
		Flavour: "test-flavour",
		Mode:    "dark",
		Colours: map[string]string{"base": "000000", "text": "ffffff"},
	}
	mockManager.AddSchemeData("test-scheme", "test-flavour", "dark", testScheme)

	cmd := setCommand()
	cmd.SetArgs([]string{"test-scheme", "test-flavour", "dark"})
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	originalApplyTheme := applyThemeFunc
	defer func() { applyThemeFunc = originalApplyTheme }()
	applyThemeFunc = func(*scheme.Scheme) error {
		return &mockError{msg: "failed to apply theme"}
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err == nil {
		t.Error("Expected error when theme application fails")
	}
	if !strings.Contains(err.Error(), "failed to apply theme") {
		t.Errorf("Expected error to contain theme error message, got: %v", err)
	}
}

func testSetCommandDefaultFlavour(t *testing.T) {
	// Arrange
	mockManager := NewMockSchemeManager()
	mockManager.AddScheme("test-scheme", []string{"default-flavour", "other-flavour"})
	testScheme := &scheme.Scheme{
		Name:    "test-scheme",
		Flavour: "default-flavour",
		Mode:    "dark",
		Colours: map[string]string{"base": "000000", "text": "ffffff"},
	}
	mockManager.AddSchemeData("test-scheme", "default-flavour", "dark", testScheme)

	cmd := setCommand()
	cmd.SetArgs([]string{"test-scheme"}) // No flavour specified
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)

	// Mock dependencies
	originalNewManager := newManagerFunc
	defer func() { newManagerFunc = originalNewManager }()
	newManagerFunc = func() SchemeManagerInterface {
		return mockManager
	}

	originalApplyTheme := applyThemeFunc
	defer func() { applyThemeFunc = originalApplyTheme }()
	applyThemeFunc = func(*scheme.Scheme) error {
		return nil
	}

	// Act
	err := cmd.Execute()

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify default flavour was used
	currentScheme, _ := mockManager.GetCurrent()
	if currentScheme.Flavour != "default-flavour" {
		t.Errorf("Expected default flavour to be used, got '%s'", currentScheme.Flavour)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "test-scheme/default-flavour/dark") {
		t.Error("Expected output to contain scheme with default flavour")
	}
}

// Mock function variables for testing
var (
	applyThemeFunc       = applyTheme
	sendNotificationFunc = func(summary, body string) error { return nil }
)

func TestSetCommandIntegration(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		setupManager   func(*MockSchemeManager)
		expectError    bool
		expectedOutput []string
		themeApplied   bool
	}{
		{
			name: "set scheme with all arguments",
			args: []string{"catppuccin", "mocha", "dark"},
			setupManager: func(m *MockSchemeManager) {
				scheme := &scheme.Scheme{
					Name:    "catppuccin",
					Flavour: "mocha",
					Mode:    "dark",
					Colours: map[string]string{"base": "1e1e2e"},
				}
				m.AddSchemeData("catppuccin", "mocha", "dark", scheme)
			},
			expectedOutput: []string{"Scheme set to catppuccin/mocha/dark"},
			themeApplied:   true,
		},
		{
			name: "set scheme with flags",
			args: []string{"--name", "gruvbox", "--flavour", "hard", "--mode", "light"},
			setupManager: func(m *MockSchemeManager) {
				scheme := &scheme.Scheme{
					Name:    "gruvbox",
					Flavour: "hard",
					Mode:    "light",
					Colours: map[string]string{"base": "fbf1c7"},
				}
				m.AddSchemeData("gruvbox", "hard", "light", scheme)
			},
			expectedOutput: []string{"Scheme set to gruvbox/hard/light"},
			themeApplied:   true,
		},
		{
			name: "set scheme without applying theme",
			args: []string{"rosepine", "main", "dark", "--no-apply"},
			setupManager: func(m *MockSchemeManager) {
				scheme := &scheme.Scheme{
					Name:    "rosepine",
					Flavour: "main",
					Mode:    "dark",
					Colours: map[string]string{"base": "191724"},
				}
				m.AddSchemeData("rosepine", "main", "dark", scheme)
			},
			expectedOutput: []string{"Scheme set to rosepine/main/dark"},
			themeApplied:   false,
		},
		{
			name:        "error when no scheme name provided",
			args:        []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockManager := NewMockSchemeManager()
			if tt.setupManager != nil {
				tt.setupManager(mockManager)
			}

			cmd := setCommand()
			cmd.SetArgs(tt.args)
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)

			// Mock dependencies
			originalNewManager := newManagerFunc
			defer func() { newManagerFunc = originalNewManager }()
			newManagerFunc = func() SchemeManagerInterface {
				return mockManager
			}

			themeApplied := false
			originalApplyTheme := applyThemeFunc
			defer func() { applyThemeFunc = originalApplyTheme }()
			applyThemeFunc = func(*scheme.Scheme) error {
				themeApplied = true
				return nil
			}

			// Act
			err := cmd.Execute()

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if themeApplied != tt.themeApplied {
				t.Errorf("Expected theme applied: %v, got: %v", tt.themeApplied, themeApplied)
			}

			outputStr := output.String()
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't. Output:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestSetCommandArgs(t *testing.T) {
	// Act
	cmd := setCommand()

	// Assert
	// The set command should accept 0-3 arguments
	if cmd.Args == nil {
		return // No args validation set, which is fine
	}

	// Test valid argument counts
	validArgCounts := []int{0, 1, 2, 3}
	for _, count := range validArgCounts {
		args := make([]string, count)
		for i := range args {
			args[i] = "arg"
		}

		err := cmd.Args(cmd, args)
		if err != nil {
			t.Errorf("Expected no error for %d args, got: %v", count, err)
		}
	}

	// Test invalid argument count (more than 3)
	args := []string{"arg1", "arg2", "arg3", "arg4"}
	err := cmd.Args(cmd, args)
	if err == nil {
		t.Error("Expected error for more than 3 arguments")
	}
}
