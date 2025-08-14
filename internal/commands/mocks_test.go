package commands

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// MockLogger provides a mock implementation of the logger for testing
type MockLogger struct {
	mu       sync.RWMutex
	messages []LogMessage
	level    string
}

// LogMessage represents a logged message
type LogMessage struct {
	Level   string
	Message string
	Args    []interface{}
	Time    time.Time
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		messages: make([]LogMessage, 0),
		level:    "info",
	}
}

// Debug logs a debug message
func (ml *MockLogger) Debug(msg string, args ...interface{}) {
	ml.log("debug", msg, args...)
}

// Info logs an info message
func (ml *MockLogger) Info(msg string, args ...interface{}) {
	ml.log("info", msg, args...)
}

// Warn logs a warning message
func (ml *MockLogger) Warn(msg string, args ...interface{}) {
	ml.log("warn", msg, args...)
}

// Error logs an error message
func (ml *MockLogger) Error(msg string, args ...interface{}) {
	ml.log("error", msg, args...)
}

// SetLevel sets the log level
func (ml *MockLogger) SetLevel(level string) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.level = level
}

// GetMessages returns all logged messages
func (ml *MockLogger) GetMessages() []LogMessage {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	return append([]LogMessage(nil), ml.messages...)
}

// GetMessagesByLevel returns messages of a specific level
func (ml *MockLogger) GetMessagesByLevel(level string) []LogMessage {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	var filtered []LogMessage
	for _, msg := range ml.messages {
		if msg.Level == level {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// Clear clears all logged messages
func (ml *MockLogger) Clear() {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	ml.messages = ml.messages[:0]
}

// HasMessage checks if a message with specific text was logged
func (ml *MockLogger) HasMessage(level, text string) bool {
	ml.mu.RLock()
	defer ml.mu.RUnlock()

	for _, msg := range ml.messages {
		if msg.Level == level && msg.Message == text {
			return true
		}
	}
	return false
}

// log is the internal logging method
func (ml *MockLogger) log(level, msg string, args ...interface{}) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	ml.messages = append(ml.messages, LogMessage{
		Level:   level,
		Message: msg,
		Args:    args,
		Time:    time.Now(),
	})
}

// MockHyprlandClient provides a mock implementation of Hyprland IPC client
type MockHyprlandClient struct {
	isRunning     bool
	version       string
	workspaces    []MockWorkspace
	windows       []MockWindow
	monitors      []MockMonitor
	commandError  error
	commandResult string
}

// MockWorkspace represents a mock workspace
type MockWorkspace struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Monitor string `json:"monitor"`
	Windows int    `json:"windows"`
	Active  bool   `json:"active"`
}

// MockWindow represents a mock window
type MockWindow struct {
	Address   string `json:"address"`
	Title     string `json:"title"`
	Class     string `json:"class"`
	Workspace int    `json:"workspace"`
	Monitor   int    `json:"monitor"`
	Floating  bool   `json:"floating"`
}

// MockMonitor represents a mock monitor
type MockMonitor struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RefreshRate float64 `json:"refreshRate"`
	Focused     bool    `json:"focused"`
}

// NewMockHyprlandClient creates a new mock Hyprland client
func NewMockHyprlandClient() *MockHyprlandClient {
	return &MockHyprlandClient{
		isRunning: false,
		version:   "Hyprland, built from branch  at commit  (props).",
		workspaces: []MockWorkspace{
			{ID: 1, Name: "1", Monitor: "DP-1", Windows: 2, Active: true},
			{ID: 2, Name: "2", Monitor: "DP-1", Windows: 1, Active: false},
		},
		windows: []MockWindow{
			{Address: "0x123", Title: "Terminal", Class: "kitty", Workspace: 1, Monitor: 0, Floating: false},
			{Address: "0x456", Title: "Browser", Class: "firefox", Workspace: 1, Monitor: 0, Floating: false},
		},
		monitors: []MockMonitor{
			{ID: 0, Name: "DP-1", Width: 1920, Height: 1080, RefreshRate: 144.0, Focused: true},
		},
	}
}

// SetRunning sets whether Hyprland is running
func (mhc *MockHyprlandClient) SetRunning(running bool) {
	mhc.isRunning = running
	if running {
		os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "mock-signature")
	} else {
		os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
	}
}

// SetVersion sets the mock version
func (mhc *MockHyprlandClient) SetVersion(version string) {
	mhc.version = version
}

// SetCommandError sets an error to be returned by commands
func (mhc *MockHyprlandClient) SetCommandError(err error) {
	mhc.commandError = err
}

// SetCommandResult sets the result to be returned by commands
func (mhc *MockHyprlandClient) SetCommandResult(result string) {
	mhc.commandResult = result
}

// AddWorkspace adds a mock workspace
func (mhc *MockHyprlandClient) AddWorkspace(ws MockWorkspace) {
	mhc.workspaces = append(mhc.workspaces, ws)
}

// AddWindow adds a mock window
func (mhc *MockHyprlandClient) AddWindow(win MockWindow) {
	mhc.windows = append(mhc.windows, win)
}

// AddMonitor adds a mock monitor
func (mhc *MockHyprlandClient) AddMonitor(mon MockMonitor) {
	mhc.monitors = append(mhc.monitors, mon)
}

// IsRunning returns whether Hyprland is running
func (mhc *MockHyprlandClient) IsRunning() bool {
	return mhc.isRunning
}

// GetVersion returns the mock version
func (mhc *MockHyprlandClient) GetVersion() (string, error) {
	if mhc.commandError != nil {
		return "", mhc.commandError
	}
	return mhc.version, nil
}

// SendCommand sends a mock command
func (mhc *MockHyprlandClient) SendCommand(command string) (string, error) {
	if mhc.commandError != nil {
		return "", mhc.commandError
	}
	if mhc.commandResult != "" {
		return mhc.commandResult, nil
	}
	return fmt.Sprintf("Mock response for: %s", command), nil
}

// MockNotifier provides a mock implementation of the notification system
type MockNotifier struct {
	isAvailable   bool
	notifications []MockNotification
	sendError     error
	lastID        uint32
}

// MockNotification represents a mock notification
type MockNotification struct {
	ID       uint32
	Summary  string
	Body     string
	Icon     string
	Urgency  string
	Timeout  time.Duration
	Category string
	AppName  string
	SentAt   time.Time
}

// NewMockNotifier creates a new mock notifier
func NewMockNotifier() *MockNotifier {
	return &MockNotifier{
		isAvailable:   true,
		notifications: make([]MockNotification, 0),
		lastID:        0,
	}
}

// SetAvailable sets whether the notification system is available
func (mn *MockNotifier) SetAvailable(available bool) {
	mn.isAvailable = available
}

// SetSendError sets an error to be returned when sending notifications
func (mn *MockNotifier) SetSendError(err error) {
	mn.sendError = err
}

// IsAvailable returns whether the notification system is available
func (mn *MockNotifier) IsAvailable() bool {
	return mn.isAvailable
}

// Send sends a mock notification
func (mn *MockNotifier) Send(summary, body string) error {
	if mn.sendError != nil {
		return mn.sendError
	}

	mn.lastID++
	notification := MockNotification{
		ID:      mn.lastID,
		Summary: summary,
		Body:    body,
		AppName: "heimdall",
		SentAt:  time.Now(),
	}

	mn.notifications = append(mn.notifications, notification)
	return nil
}

// SendUrgent sends a mock urgent notification
func (mn *MockNotifier) SendUrgent(summary, body string) error {
	if mn.sendError != nil {
		return mn.sendError
	}

	mn.lastID++
	notification := MockNotification{
		ID:      mn.lastID,
		Summary: summary,
		Body:    body,
		Urgency: "critical",
		AppName: "heimdall",
		SentAt:  time.Now(),
	}

	mn.notifications = append(mn.notifications, notification)
	return nil
}

// GetNotifications returns all sent notifications
func (mn *MockNotifier) GetNotifications() []MockNotification {
	return append([]MockNotification(nil), mn.notifications...)
}

// GetLastNotification returns the last sent notification
func (mn *MockNotifier) GetLastNotification() *MockNotification {
	if len(mn.notifications) == 0 {
		return nil
	}
	return &mn.notifications[len(mn.notifications)-1]
}

// Clear clears all notifications
func (mn *MockNotifier) Clear() {
	mn.notifications = mn.notifications[:0]
	mn.lastID = 0
}

// HasNotification checks if a notification with specific summary was sent
func (mn *MockNotifier) HasNotification(summary string) bool {
	for _, notif := range mn.notifications {
		if notif.Summary == summary {
			return true
		}
	}
	return false
}

// MockColorGenerator provides a mock implementation of color generation
type MockColorGenerator struct {
	colors      map[string]string
	generateErr error
}

// NewMockColorGenerator creates a new mock color generator
func NewMockColorGenerator() *MockColorGenerator {
	return &MockColorGenerator{
		colors: map[string]string{
			"#FF6B6B": "#FF8A8A", // Lighter version
			"#4ECDC4": "#6EDDD6", // Lighter version
			"#45B7D1": "#65C7E1", // Lighter version
		},
	}
}

// SetGenerateError sets an error to be returned during color generation
func (mcg *MockColorGenerator) SetGenerateError(err error) {
	mcg.generateErr = err
}

// AddColor adds a color mapping
func (mcg *MockColorGenerator) AddColor(original, lighter string) {
	mcg.colors[original] = lighter
}

// GenerateLighter generates a lighter version of a color
func (mcg *MockColorGenerator) GenerateLighter(hex string) (string, error) {
	if mcg.generateErr != nil {
		return "", mcg.generateErr
	}

	if lighter, exists := mcg.colors[hex]; exists {
		return lighter, nil
	}

	// Default behavior - just return a slightly modified version
	return hex + "AA", nil
}

// MockFileSystem provides a mock file system for testing
type MockFileSystem struct {
	files       map[string][]byte
	directories map[string]bool
	readError   error
	writeError  error
	statError   error
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files:       make(map[string][]byte),
		directories: make(map[string]bool),
	}
}

// SetReadError sets an error to be returned when reading files
func (mfs *MockFileSystem) SetReadError(err error) {
	mfs.readError = err
}

// SetWriteError sets an error to be returned when writing files
func (mfs *MockFileSystem) SetWriteError(err error) {
	mfs.writeError = err
}

// SetStatError sets an error to be returned when stating files
func (mfs *MockFileSystem) SetStatError(err error) {
	mfs.statError = err
}

// AddFile adds a file to the mock file system
func (mfs *MockFileSystem) AddFile(path string, content []byte) {
	mfs.files[path] = content
}

// AddDirectory adds a directory to the mock file system
func (mfs *MockFileSystem) AddDirectory(path string) {
	mfs.directories[path] = true
}

// ReadFile reads a file from the mock file system
func (mfs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	if mfs.readError != nil {
		return nil, mfs.readError
	}

	if content, exists := mfs.files[path]; exists {
		return content, nil
	}

	return nil, fmt.Errorf("file not found: %s", path)
}

// WriteFile writes a file to the mock file system
func (mfs *MockFileSystem) WriteFile(path string, content []byte) error {
	if mfs.writeError != nil {
		return mfs.writeError
	}

	mfs.files[path] = content
	return nil
}

// Exists checks if a file or directory exists
func (mfs *MockFileSystem) Exists(path string) bool {
	if mfs.statError != nil {
		return false
	}

	_, fileExists := mfs.files[path]
	_, dirExists := mfs.directories[path]
	return fileExists || dirExists
}

// IsDir checks if a path is a directory
func (mfs *MockFileSystem) IsDir(path string) bool {
	if mfs.statError != nil {
		return false
	}

	return mfs.directories[path]
}

// ListFiles returns all files in the mock file system
func (mfs *MockFileSystem) ListFiles() []string {
	var files []string
	for path := range mfs.files {
		files = append(files, path)
	}
	return files
}

// ListDirectories returns all directories in the mock file system
func (mfs *MockFileSystem) ListDirectories() []string {
	var dirs []string
	for path := range mfs.directories {
		dirs = append(dirs, path)
	}
	return dirs
}

// Clear clears all files and directories
func (mfs *MockFileSystem) Clear() {
	mfs.files = make(map[string][]byte)
	mfs.directories = make(map[string]bool)
}

// MockEnvironment provides utilities for mocking environment variables
type MockEnvironment struct {
	original map[string]string
	current  map[string]string
}

// NewMockEnvironment creates a new mock environment
func NewMockEnvironment() *MockEnvironment {
	return &MockEnvironment{
		original: make(map[string]string),
		current:  make(map[string]string),
	}
}

// Set sets an environment variable
func (me *MockEnvironment) Set(key, value string) {
	if _, exists := me.original[key]; !exists {
		me.original[key] = os.Getenv(key)
	}
	me.current[key] = value
	os.Setenv(key, value)
}

// Unset unsets an environment variable
func (me *MockEnvironment) Unset(key string) {
	if _, exists := me.original[key]; !exists {
		me.original[key] = os.Getenv(key)
	}
	delete(me.current, key)
	os.Unsetenv(key)
}

// Get gets an environment variable
func (me *MockEnvironment) Get(key string) string {
	if value, exists := me.current[key]; exists {
		return value
	}
	return os.Getenv(key)
}

// Restore restores all original environment variables
func (me *MockEnvironment) Restore() {
	for key, originalValue := range me.original {
		if originalValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, originalValue)
		}
	}
	me.current = make(map[string]string)
}

// GetAll returns all current environment variables
func (me *MockEnvironment) GetAll() map[string]string {
	result := make(map[string]string)
	for key, value := range me.current {
		result[key] = value
	}
	return result
}

// TestFixtures provides common test fixtures
type TestFixtures struct {
	SampleConfig    string
	SampleScheme    string
	SampleWallpaper string
}

// NewTestFixtures creates new test fixtures
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{
		SampleConfig: `{
	"theme": "dark",
	"wallpaper": "/path/to/wallpaper.jpg",
	"scheme": "gruvbox",
	"terminal": "kitty",
	"shell": "zsh",
	"verbose": false,
	"debug": false
}`,
		SampleScheme: `{
	"name": "test-scheme",
	"colors": {
		"background": "#1d2021",
		"foreground": "#ebdbb2",
		"cursor": "#ebdbb2",
		"selection": "#504945",
		"color0": "#1d2021",
		"color1": "#cc241d",
		"color2": "#98971a",
		"color3": "#d79921",
		"color4": "#458588",
		"color5": "#b16286",
		"color6": "#689d6a",
		"color7": "#a89984",
		"color8": "#928374",
		"color9": "#fb4934",
		"color10": "#b8bb26",
		"color11": "#fabd2f",
		"color12": "#83a598",
		"color13": "#d3869b",
		"color14": "#8ec07c",
		"color15": "#ebdbb2"
	}
}`,
		SampleWallpaper: "/path/to/sample/wallpaper.jpg",
	}
}
