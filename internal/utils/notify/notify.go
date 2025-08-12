package notify

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Urgency levels for notifications
type Urgency string

const (
	UrgencyLow      Urgency = "low"
	UrgencyNormal   Urgency = "normal"
	UrgencyCritical Urgency = "critical"
)

// Notification represents a desktop notification
type Notification struct {
	Summary   string
	Body      string
	Icon      string
	Urgency   Urgency
	Timeout   time.Duration // in milliseconds, 0 for default
	Category  string
	Transient bool
	ReplaceID uint32
	AppName   string
}

// Notifier handles sending notifications
type Notifier struct {
	command string
}

// NewNotifier creates a new notifier
func NewNotifier() *Notifier {
	// Check for notify-send command
	cmd := "notify-send"
	if _, err := exec.LookPath(cmd); err != nil {
		// Try alternative commands
		if _, err := exec.LookPath("dunstify"); err == nil {
			cmd = "dunstify"
		}
	}

	return &Notifier{
		command: cmd,
	}
}

// Send sends a simple notification
func Send(summary, body string) error {
	n := NewNotifier()
	return n.Send(&Notification{
		Summary: summary,
		Body:    body,
	})
}

// SendUrgent sends an urgent notification
func SendUrgent(summary, body string) error {
	n := NewNotifier()
	return n.Send(&Notification{
		Summary: summary,
		Body:    body,
		Urgency: UrgencyCritical,
	})
}

// Send sends a notification
func (n *Notifier) Send(notif *Notification) error {
	args := []string{}

	// Add app name if specified
	if notif.AppName != "" {
		args = append(args, "-a", notif.AppName)
	} else {
		args = append(args, "-a", "heimdall")
	}

	// Add urgency
	if notif.Urgency != "" {
		args = append(args, "-u", string(notif.Urgency))
	}

	// Add timeout
	if notif.Timeout > 0 {
		args = append(args, "-t", strconv.Itoa(int(notif.Timeout.Milliseconds())))
	}

	// Add icon
	if notif.Icon != "" {
		args = append(args, "-i", notif.Icon)
	}

	// Add category
	if notif.Category != "" {
		args = append(args, "-c", notif.Category)
	}

	// Add transient hint
	if notif.Transient {
		args = append(args, "-h", "int:transient:1")
	}

	// Add replace ID for dunstify
	if notif.ReplaceID > 0 && n.command == "dunstify" {
		args = append(args, "-r", strconv.Itoa(int(notif.ReplaceID)))
	}

	// Add summary and body
	args = append(args, notif.Summary)
	if notif.Body != "" {
		args = append(args, notif.Body)
	}

	// Execute command
	cmd := exec.Command(n.command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to send notification: %w (output: %s)", err, string(output))
	}

	return nil
}

// SendWithID sends a notification and returns its ID (dunstify only)
func (n *Notifier) SendWithID(notif *Notification) (uint32, error) {
	if n.command != "dunstify" {
		// Regular notify-send doesn't support IDs
		err := n.Send(notif)
		return 0, err
	}

	args := []string{"-p"} // Print ID

	// Add app name
	if notif.AppName != "" {
		args = append(args, "-a", notif.AppName)
	} else {
		args = append(args, "-a", "heimdall")
	}

	// Add urgency
	if notif.Urgency != "" {
		args = append(args, "-u", string(notif.Urgency))
	}

	// Add timeout
	if notif.Timeout > 0 {
		args = append(args, "-t", strconv.Itoa(int(notif.Timeout.Milliseconds())))
	}

	// Add icon
	if notif.Icon != "" {
		args = append(args, "-i", notif.Icon)
	}

	// Add replace ID
	if notif.ReplaceID > 0 {
		args = append(args, "-r", strconv.Itoa(int(notif.ReplaceID)))
	}

	// Add summary and body
	args = append(args, notif.Summary)
	if notif.Body != "" {
		args = append(args, notif.Body)
	}

	// Execute command
	cmd := exec.Command(n.command, args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to send notification: %w", err)
	}

	// Parse ID from output
	idStr := strings.TrimSpace(string(output))
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse notification ID: %w", err)
	}

	return uint32(id), nil
}

// Close closes a notification by ID (dunstify only)
func (n *Notifier) Close(id uint32) error {
	if n.command != "dunstify" {
		// Regular notify-send doesn't support closing
		return fmt.Errorf("closing notifications not supported with %s", n.command)
	}

	cmd := exec.Command(n.command, "-C", strconv.Itoa(int(id)))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to close notification: %w", err)
	}

	return nil
}

// CloseAll closes all notifications
func (n *Notifier) CloseAll() error {
	if n.command == "dunstify" {
		cmd := exec.Command("dunstctl", "close-all")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to close all notifications: %w", err)
		}
	} else {
		// Try using gdbus for notify-send
		cmd := exec.Command("gdbus", "call", "--session",
			"--dest", "org.freedesktop.Notifications",
			"--object-path", "/org/freedesktop/Notifications",
			"--method", "org.freedesktop.Notifications.CloseNotification", "0")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to close notifications: %w", err)
		}
	}

	return nil
}

// IsAvailable checks if notification system is available
func IsAvailable() bool {
	n := NewNotifier()
	if n.command == "" {
		return false
	}

	if _, err := exec.LookPath(n.command); err != nil {
		return false
	}

	return true
}
