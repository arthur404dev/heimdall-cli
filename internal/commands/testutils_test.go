package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TestUtilities provides common testing utilities for command tests
type TestUtilities struct {
	t           *testing.T
	tempDir     string
	originalEnv map[string]string
	cleanup     []func()
}

// NewTestUtilities creates a new test utilities instance
func NewTestUtilities(t *testing.T) *TestUtilities {
	return &TestUtilities{
		t:           t,
		originalEnv: make(map[string]string),
		cleanup:     make([]func(), 0),
	}
}

// CreateTempDir creates a temporary directory for testing
func (tu *TestUtilities) CreateTempDir() string {
	if tu.tempDir == "" {
		tu.tempDir = tu.t.TempDir()
	}
	return tu.tempDir
}

// CreateTempFile creates a temporary file with content
func (tu *TestUtilities) CreateTempFile(name, content string) string {
	tempDir := tu.CreateTempDir()
	filePath := filepath.Join(tempDir, name)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		tu.t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		tu.t.Fatalf("Failed to create temp file %s: %v", filePath, err)
	}

	return filePath
}

// SetEnv sets an environment variable and tracks it for cleanup
func (tu *TestUtilities) SetEnv(key, value string) {
	if _, exists := tu.originalEnv[key]; !exists {
		tu.originalEnv[key] = os.Getenv(key)
	}
	os.Setenv(key, value)
}

// UnsetEnv unsets an environment variable and tracks it for cleanup
func (tu *TestUtilities) UnsetEnv(key string) {
	if _, exists := tu.originalEnv[key]; !exists {
		tu.originalEnv[key] = os.Getenv(key)
	}
	os.Unsetenv(key)
}

// AddCleanup adds a cleanup function to be called during cleanup
func (tu *TestUtilities) AddCleanup(fn func()) {
	tu.cleanup = append(tu.cleanup, fn)
}

// Cleanup restores environment and runs cleanup functions
func (tu *TestUtilities) Cleanup() {
	// Restore environment variables
	for key, originalValue := range tu.originalEnv {
		if originalValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, originalValue)
		}
	}

	// Run cleanup functions in reverse order
	for i := len(tu.cleanup) - 1; i >= 0; i-- {
		tu.cleanup[i]()
	}
}

// CaptureOutput captures stdout and stderr from a function
func (tu *TestUtilities) CaptureOutput(fn func()) (stdout, stderr string) {
	// Capture stdout
	oldStdout := os.Stdout
	stdoutR, stdoutW, _ := os.Pipe()
	os.Stdout = stdoutW

	// Capture stderr
	oldStderr := os.Stderr
	stderrR, stderrW, _ := os.Pipe()
	os.Stderr = stderrW

	// Channel to collect output
	stdoutChan := make(chan string)
	stderrChan := make(chan string)

	// Read stdout
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stdoutR)
		stdoutChan <- buf.String()
	}()

	// Read stderr
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, stderrR)
		stderrChan <- buf.String()
	}()

	// Execute function
	fn()

	// Close writers
	stdoutW.Close()
	stderrW.Close()

	// Restore original stdout/stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Get captured output
	stdout = <-stdoutChan
	stderr = <-stderrChan

	return stdout, stderr
}

// MockCommand creates a mock cobra command for testing
type MockCommand struct {
	name        string
	short       string
	long        string
	hidden      bool
	runFunc     func(cmd *cobra.Command, args []string)
	runEFunc    func(cmd *cobra.Command, args []string) error
	flags       map[string]interface{}
	subcommands []*MockCommand
}

// NewMockCommand creates a new mock command
func NewMockCommand(name string) *MockCommand {
	return &MockCommand{
		name:  name,
		flags: make(map[string]interface{}),
	}
}

// WithShort sets the short description
func (mc *MockCommand) WithShort(short string) *MockCommand {
	mc.short = short
	return mc
}

// WithLong sets the long description
func (mc *MockCommand) WithLong(long string) *MockCommand {
	mc.long = long
	return mc
}

// WithHidden sets the hidden flag
func (mc *MockCommand) WithHidden(hidden bool) *MockCommand {
	mc.hidden = hidden
	return mc
}

// WithRun sets the run function
func (mc *MockCommand) WithRun(fn func(cmd *cobra.Command, args []string)) *MockCommand {
	mc.runFunc = fn
	return mc
}

// WithRunE sets the run function that returns an error
func (mc *MockCommand) WithRunE(fn func(cmd *cobra.Command, args []string) error) *MockCommand {
	mc.runEFunc = fn
	return mc
}

// WithStringFlag adds a string flag
func (mc *MockCommand) WithStringFlag(name, defaultValue, usage string) *MockCommand {
	mc.flags[name] = map[string]interface{}{
		"type":    "string",
		"default": defaultValue,
		"usage":   usage,
	}
	return mc
}

// WithBoolFlag adds a boolean flag
func (mc *MockCommand) WithBoolFlag(name string, defaultValue bool, usage string) *MockCommand {
	mc.flags[name] = map[string]interface{}{
		"type":    "bool",
		"default": defaultValue,
		"usage":   usage,
	}
	return mc
}

// WithSubcommand adds a subcommand
func (mc *MockCommand) WithSubcommand(sub *MockCommand) *MockCommand {
	mc.subcommands = append(mc.subcommands, sub)
	return mc
}

// Build creates the actual cobra command
func (mc *MockCommand) Build() *cobra.Command {
	cmd := &cobra.Command{
		Use:    mc.name,
		Short:  mc.short,
		Long:   mc.long,
		Hidden: mc.hidden,
	}

	if mc.runFunc != nil {
		cmd.Run = mc.runFunc
	}

	if mc.runEFunc != nil {
		cmd.RunE = mc.runEFunc
	}

	// Add flags
	for name, config := range mc.flags {
		flagConfig := config.(map[string]interface{})
		switch flagConfig["type"] {
		case "string":
			cmd.Flags().String(name, flagConfig["default"].(string), flagConfig["usage"].(string))
		case "bool":
			cmd.Flags().Bool(name, flagConfig["default"].(bool), flagConfig["usage"].(string))
		}
	}

	// Add subcommands
	for _, sub := range mc.subcommands {
		cmd.AddCommand(sub.Build())
	}

	return cmd
}

// CommandTester provides utilities for testing cobra commands
type CommandTester struct {
	cmd    *cobra.Command
	stdout *bytes.Buffer
	stderr *bytes.Buffer
}

// NewCommandTester creates a new command tester
func NewCommandTester(cmd *cobra.Command) *CommandTester {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	return &CommandTester{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
	}
}

// Execute executes the command with given arguments
func (ct *CommandTester) Execute(args ...string) error {
	ct.cmd.SetArgs(args)
	return ct.cmd.Execute()
}

// ExecuteWithTimeout executes the command with a timeout
func (ct *CommandTester) ExecuteWithTimeout(timeout time.Duration, args ...string) error {
	done := make(chan error, 1)

	go func() {
		done <- ct.Execute(args...)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("command execution timed out after %v", timeout)
	}
}

// GetStdout returns the captured stdout
func (ct *CommandTester) GetStdout() string {
	return ct.stdout.String()
}

// GetStderr returns the captured stderr
func (ct *CommandTester) GetStderr() string {
	return ct.stderr.String()
}

// GetOutput returns both stdout and stderr combined
func (ct *CommandTester) GetOutput() string {
	return ct.stdout.String() + ct.stderr.String()
}

// Reset clears the output buffers
func (ct *CommandTester) Reset() {
	ct.stdout.Reset()
	ct.stderr.Reset()
}

// AssertContains checks if output contains expected string
func (ct *CommandTester) AssertContains(t *testing.T, expected string) {
	output := ct.GetOutput()
	if !strings.Contains(output, expected) {
		t.Errorf("Output should contain: %s\nActual output: %s", expected, output)
	}
}

// AssertNotContains checks if output does not contain string
func (ct *CommandTester) AssertNotContains(t *testing.T, unexpected string) {
	output := ct.GetOutput()
	if strings.Contains(output, unexpected) {
		t.Errorf("Output should not contain: %s\nActual output: %s", unexpected, output)
	}
}

// AssertStdoutContains checks if stdout contains expected string
func (ct *CommandTester) AssertStdoutContains(t *testing.T, expected string) {
	stdout := ct.GetStdout()
	if !strings.Contains(stdout, expected) {
		t.Errorf("Stdout should contain: %s\nActual stdout: %s", expected, stdout)
	}
}

// AssertStderrContains checks if stderr contains expected string
func (ct *CommandTester) AssertStderrContains(t *testing.T, expected string) {
	stderr := ct.GetStderr()
	if !strings.Contains(stderr, expected) {
		t.Errorf("Stderr should contain: %s\nActual stderr: %s", expected, stderr)
	}
}

// ConfigTester provides utilities for testing configuration
type ConfigTester struct {
	originalConfig map[string]interface{}
	tempConfigFile string
}

// NewConfigTester creates a new config tester
func NewConfigTester() *ConfigTester {
	return &ConfigTester{
		originalConfig: make(map[string]interface{}),
	}
}

// SetConfig sets a configuration value and tracks original
func (ct *ConfigTester) SetConfig(key string, value interface{}) {
	if !viper.IsSet(key) {
		ct.originalConfig[key] = nil
	} else {
		ct.originalConfig[key] = viper.Get(key)
	}
	viper.Set(key, value)
}

// CreateTempConfig creates a temporary config file
func (ct *ConfigTester) CreateTempConfig(content string) string {
	tempFile, err := os.CreateTemp("", "heimdall-test-config-*.json")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp config file: %v", err))
	}

	if _, err := tempFile.WriteString(content); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		panic(fmt.Sprintf("Failed to write temp config file: %v", err))
	}

	tempFile.Close()
	ct.tempConfigFile = tempFile.Name()
	return ct.tempConfigFile
}

// LoadTempConfig loads the temporary config file
func (ct *ConfigTester) LoadTempConfig() error {
	if ct.tempConfigFile == "" {
		return fmt.Errorf("no temp config file created")
	}

	viper.SetConfigFile(ct.tempConfigFile)
	return viper.ReadInConfig()
}

// Cleanup restores original configuration
func (ct *ConfigTester) Cleanup() {
	// Restore original config values
	for key, value := range ct.originalConfig {
		if value == nil {
			viper.Set(key, nil)
		} else {
			viper.Set(key, value)
		}
	}

	// Remove temp config file
	if ct.tempConfigFile != "" {
		os.Remove(ct.tempConfigFile)
	}

	// Reset viper
	viper.Reset()
}

// TestAssertions provides custom assertion helpers
type TestAssertions struct {
	t *testing.T
}

// NewTestAssertions creates new test assertions
func NewTestAssertions(t *testing.T) *TestAssertions {
	return &TestAssertions{t: t}
}

// NoError asserts that error is nil
func (ta *TestAssertions) NoError(err error, msgAndArgs ...interface{}) {
	if err != nil {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected no error but got: %v. %v", err, msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected no error but got: %v", err)
		}
	}
}

// Error asserts that error is not nil
func (ta *TestAssertions) Error(err error, msgAndArgs ...interface{}) {
	if err == nil {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected error but got none. %v", msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected error but got none")
		}
	}
}

// Equal asserts that two values are equal
func (ta *TestAssertions) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected != actual {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected %v but got %v. %v", expected, actual, msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected %v but got %v", expected, actual)
		}
	}
}

// NotEqual asserts that two values are not equal
func (ta *TestAssertions) NotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	if expected == actual {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected %v to not equal %v. %v", expected, actual, msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected %v to not equal %v", expected, actual)
		}
	}
}

// Contains asserts that string contains substring
func (ta *TestAssertions) Contains(s, substr string, msgAndArgs ...interface{}) {
	if !strings.Contains(s, substr) {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected '%s' to contain '%s'. %v", s, substr, msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected '%s' to contain '%s'", s, substr)
		}
	}
}

// NotContains asserts that string does not contain substring
func (ta *TestAssertions) NotContains(s, substr string, msgAndArgs ...interface{}) {
	if strings.Contains(s, substr) {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected '%s' to not contain '%s'. %v", s, substr, msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected '%s' to not contain '%s'", s, substr)
		}
	}
}

// True asserts that value is true
func (ta *TestAssertions) True(value bool, msgAndArgs ...interface{}) {
	if !value {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected true but got false. %v", msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected true but got false")
		}
	}
}

// False asserts that value is false
func (ta *TestAssertions) False(value bool, msgAndArgs ...interface{}) {
	if value {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected false but got true. %v", msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected false but got true")
		}
	}
}

// Nil asserts that value is nil
func (ta *TestAssertions) Nil(value interface{}, msgAndArgs ...interface{}) {
	if value != nil {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected nil but got %v. %v", value, msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected nil but got %v", value)
		}
	}
}

// NotNil asserts that value is not nil
func (ta *TestAssertions) NotNil(value interface{}, msgAndArgs ...interface{}) {
	if value == nil {
		if len(msgAndArgs) > 0 {
			ta.t.Errorf("Expected non-nil value. %v", msgAndArgs[0])
		} else {
			ta.t.Errorf("Expected non-nil value")
		}
	}
}

// TestRunner provides utilities for running test suites
type TestRunner struct {
	t     *testing.T
	utils *TestUtilities
}

// NewTestRunner creates a new test runner
func NewTestRunner(t *testing.T) *TestRunner {
	return &TestRunner{
		t:     t,
		utils: NewTestUtilities(t),
	}
}

// Run runs a test with automatic cleanup
func (tr *TestRunner) Run(name string, testFunc func(*TestUtilities)) {
	tr.t.Run(name, func(t *testing.T) {
		utils := NewTestUtilities(t)
		defer utils.Cleanup()
		testFunc(utils)
	})
}

// RunParallel runs tests in parallel
func (tr *TestRunner) RunParallel(tests map[string]func(*TestUtilities)) {
	for name, testFunc := range tests {
		name := name
		testFunc := testFunc
		tr.t.Run(name, func(t *testing.T) {
			t.Parallel()
			utils := NewTestUtilities(t)
			defer utils.Cleanup()
			testFunc(utils)
		})
	}
}
