package hypr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Client represents a Hyprland IPC client
type Client struct {
	socketPath string
	timeout    time.Duration
}

// Workspace represents a Hyprland workspace
type Workspace struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Monitor         string `json:"monitor"`
	MonitorID       int    `json:"monitorID"`
	Windows         int    `json:"windows"`
	HasFullscreen   bool   `json:"hasfullscreen"`
	LastWindow      string `json:"lastwindow"`
	LastWindowTitle string `json:"lastwindowtitle"`
}

// Window represents a Hyprland window
type Window struct {
	Address        string    `json:"address"`
	Mapped         bool      `json:"mapped"`
	Hidden         bool      `json:"hidden"`
	At             [2]int    `json:"at"`
	Size           [2]int    `json:"size"`
	Workspace      Workspace `json:"workspace"`
	Floating       bool      `json:"floating"`
	Monitor        int       `json:"monitor"`
	Class          string    `json:"class"`
	Title          string    `json:"title"`
	InitialClass   string    `json:"initialClass"`
	InitialTitle   string    `json:"initialTitle"`
	PID            int       `json:"pid"`
	Xwayland       bool      `json:"xwayland"`
	Pinned         bool      `json:"pinned"`
	Fullscreen     int       `json:"fullscreen"`
	FullscreenMode int       `json:"fullscreenMode"`
	FakeFullscreen bool      `json:"fakeFullscreen"`
	Grouped        []string  `json:"grouped"`
	Tags           []string  `json:"tags"`
	Swallowing     string    `json:"swallowing"`
	FocusHistoryID int       `json:"focusHistoryID"`
}

// Monitor represents a Hyprland monitor
type Monitor struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Make             string    `json:"make"`
	Model            string    `json:"model"`
	Serial           string    `json:"serial"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	RefreshRate      float64   `json:"refreshRate"`
	X                int       `json:"x"`
	Y                int       `json:"y"`
	ActiveWorkspace  Workspace `json:"activeWorkspace"`
	SpecialWorkspace Workspace `json:"specialWorkspace"`
	Reserved         [4]int    `json:"reserved"`
	Scale            float64   `json:"scale"`
	Transform        int       `json:"transform"`
	Focused          bool      `json:"focused"`
	DPMSStatus       bool      `json:"dpmsStatus"`
	VRR              bool      `json:"vrr"`
	ActivelyTearing  bool      `json:"activelyTearing"`
	Disabled         bool      `json:"disabled"`
	CurrentFormat    string    `json:"currentFormat"`
	AvailableModes   []string  `json:"availableModes"`
}

// Event represents a Hyprland event
type Event struct {
	Type string
	Data string
}

// NewClient creates a new Hyprland IPC client
func NewClient() (*Client, error) {
	// Get Hyprland instance signature
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return nil, fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE not set - is Hyprland running?")
	}

	// Construct socket path
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}

	socketPath := filepath.Join(runtimeDir, "hypr", signature, ".socket.sock")

	// Check if socket exists
	if _, err := os.Stat(socketPath); err != nil {
		return nil, fmt.Errorf("Hyprland socket not found at %s: %w", socketPath, err)
	}

	return &Client{
		socketPath: socketPath,
		timeout:    5 * time.Second,
	}, nil
}

// SendCommand sends a command to Hyprland and returns the response
func (c *Client) SendCommand(command string) (string, error) {
	// Connect to socket
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Hyprland socket: %w", err)
	}
	defer conn.Close()

	// Set deadline
	conn.SetDeadline(time.Now().Add(c.timeout))

	// Send command
	if _, err := conn.Write([]byte(command)); err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	var response strings.Builder
	for scanner.Scan() {
		response.WriteString(scanner.Text())
		response.WriteString("\n")
	}

	if err := scanner.Err(); err != nil && err.Error() != "EOF" {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(response.String()), nil
}

// GetWorkspaces returns all workspaces
func (c *Client) GetWorkspaces() ([]Workspace, error) {
	response, err := c.SendCommand("j/workspaces")
	if err != nil {
		return nil, err
	}

	var workspaces []Workspace
	if err := json.Unmarshal([]byte(response), &workspaces); err != nil {
		return nil, fmt.Errorf("failed to parse workspaces: %w", err)
	}

	return workspaces, nil
}

// GetActiveWorkspace returns the active workspace
func (c *Client) GetActiveWorkspace() (*Workspace, error) {
	response, err := c.SendCommand("j/activeworkspace")
	if err != nil {
		return nil, err
	}

	var workspace Workspace
	if err := json.Unmarshal([]byte(response), &workspace); err != nil {
		return nil, fmt.Errorf("failed to parse active workspace: %w", err)
	}

	return &workspace, nil
}

// GetWindows returns all windows
func (c *Client) GetWindows() ([]Window, error) {
	response, err := c.SendCommand("j/clients")
	if err != nil {
		return nil, err
	}

	var windows []Window
	if err := json.Unmarshal([]byte(response), &windows); err != nil {
		return nil, fmt.Errorf("failed to parse windows: %w", err)
	}

	return windows, nil
}

// GetMonitors returns all monitors
func (c *Client) GetMonitors() ([]Monitor, error) {
	response, err := c.SendCommand("j/monitors")
	if err != nil {
		return nil, err
	}

	var monitors []Monitor
	if err := json.Unmarshal([]byte(response), &monitors); err != nil {
		return nil, fmt.Errorf("failed to parse monitors: %w", err)
	}

	return monitors, nil
}

// Dispatch sends a dispatcher command to Hyprland
func (c *Client) Dispatch(dispatcher string, params ...string) error {
	command := fmt.Sprintf("dispatch %s", dispatcher)
	if len(params) > 0 {
		command += " " + strings.Join(params, " ")
	}

	response, err := c.SendCommand(command)
	if err != nil {
		return err
	}

	// Check for error in response
	if strings.Contains(response, "invalid") || strings.Contains(response, "error") {
		return fmt.Errorf("dispatch failed: %s", response)
	}

	return nil
}

// ToggleSpecialWorkspace toggles a special workspace
func (c *Client) ToggleSpecialWorkspace(name string) error {
	return c.Dispatch("togglespecialworkspace", name)
}

// MoveToWorkspace moves the active window to a workspace
func (c *Client) MoveToWorkspace(workspace string) error {
	return c.Dispatch("movetoworkspace", workspace)
}

// MoveToWorkspaceSilent moves the active window to a workspace without switching to it
func (c *Client) MoveToWorkspaceSilent(workspace string) error {
	return c.Dispatch("movetoworkspacesilent", workspace)
}

// FocusWindow focuses a window by address
func (c *Client) FocusWindow(address string) error {
	return c.Dispatch("focuswindow", fmt.Sprintf("address:%s", address))
}

// CloseWindow closes the active window
func (c *Client) CloseWindow() error {
	return c.Dispatch("killactive")
}

// Reload reloads the Hyprland configuration
func (c *Client) Reload() error {
	_, err := c.SendCommand("reload")
	return err
}

// GetVersion returns the Hyprland version
func (c *Client) GetVersion() (string, error) {
	return c.SendCommand("version")
}

// Subscribe subscribes to Hyprland events
func (c *Client) Subscribe(events []string) (<-chan Event, error) {
	// Get event socket path
	signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if signature == "" {
		return nil, fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE not set")
	}

	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}

	socketPath := filepath.Join(runtimeDir, "hypr", signature, ".socket2.sock")

	// Connect to event socket
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to event socket: %w", err)
	}

	// Create event channel
	eventChan := make(chan Event, 100)

	// Start event reader goroutine
	go func() {
		defer close(eventChan)
		defer conn.Close()

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()

			// Parse event
			parts := strings.SplitN(line, ">>", 2)
			if len(parts) == 2 {
				event := Event{
					Type: parts[0],
					Data: parts[1],
				}

				// Filter events if specified
				if len(events) == 0 || contains(events, event.Type) {
					eventChan <- event
				}
			}
		}
	}()

	return eventChan, nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsRunning checks if Hyprland is running
func IsRunning() bool {
	return os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != ""
}
