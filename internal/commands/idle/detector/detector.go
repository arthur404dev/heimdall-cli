package detector

import (
	"os"
	"strings"
)

// Environment represents the detected desktop environment
type Environment struct {
	DisplayServer  string // "x11", "wayland", "unknown"
	DesktopEnv     string // "gnome", "kde", "xfce", etc.
	SessionType    string // Value of XDG_SESSION_TYPE
	WaylandDisplay string // Value of WAYLAND_DISPLAY
	X11Display     string // Value of DISPLAY
	HasDBus        bool   // Whether D-Bus is available
	HasSystemd     bool   // Whether systemd is available
}

// Detect analyzes the current environment and returns detection results
func Detect() *Environment {
	env := &Environment{
		DisplayServer:  detectDisplayServer(),
		DesktopEnv:     detectDesktopEnvironment(),
		SessionType:    os.Getenv("XDG_SESSION_TYPE"),
		WaylandDisplay: os.Getenv("WAYLAND_DISPLAY"),
		X11Display:     os.Getenv("DISPLAY"),
		HasDBus:        checkDBusAvailable(),
		HasSystemd:     checkSystemdAvailable(),
	}

	return env
}

// detectDisplayServer determines if we're running on X11 or Wayland
func detectDisplayServer() string {
	// Check XDG_SESSION_TYPE first (most reliable)
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	if sessionType == "wayland" {
		return "wayland"
	}
	if sessionType == "x11" {
		return "x11"
	}

	// Check for Wayland display
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return "wayland"
	}

	// Check for X11 display
	if os.Getenv("DISPLAY") != "" {
		return "x11"
	}

	// Check if we're in a TTY/console
	if os.Getenv("TERM") != "" && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		return "console"
	}

	return "unknown"
}

// detectDesktopEnvironment identifies the current desktop environment
func detectDesktopEnvironment() string {
	// Check XDG_CURRENT_DESKTOP (most reliable)
	current := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))

	// Handle multiple values (e.g., "GNOME:GNOME-Classic")
	if current != "" {
		parts := strings.Split(current, ":")
		if len(parts) > 0 {
			current = strings.ToLower(parts[0])
		}
	}

	// Map common desktop environments
	switch current {
	case "gnome", "gnome-classic", "gnome-flashback", "ubuntu":
		return "gnome"
	case "kde", "plasma":
		return "kde"
	case "xfce", "xfce4":
		return "xfce"
	case "mate":
		return "mate"
	case "cinnamon", "x-cinnamon":
		return "cinnamon"
	case "lxde":
		return "lxde"
	case "lxqt":
		return "lxqt"
	case "enlightenment", "e", "e17", "e19", "e20", "e21", "e22", "e23", "e24":
		return "enlightenment"
	case "budgie", "budgie-desktop":
		return "budgie"
	case "deepin", "dde":
		return "deepin"
	case "pantheon":
		return "pantheon"
	case "sway":
		return "sway"
	case "hyprland":
		return "hyprland"
	case "i3":
		return "i3"
	case "awesome":
		return "awesome"
	case "bspwm":
		return "bspwm"
	case "dwm":
		return "dwm"
	case "qtile":
		return "qtile"
	case "openbox":
		return "openbox"
	}

	// Fallback checks using other environment variables
	if os.Getenv("GNOME_DESKTOP_SESSION_ID") != "" {
		return "gnome"
	}
	if os.Getenv("KDE_FULL_SESSION") != "" || os.Getenv("KDE_SESSION_VERSION") != "" {
		return "kde"
	}
	if os.Getenv("MATE_DESKTOP_SESSION_ID") != "" {
		return "mate"
	}

	// Check for window manager specific variables
	if os.Getenv("SWAYSOCK") != "" {
		return "sway"
	}
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		return "hyprland"
	}
	if os.Getenv("I3SOCK") != "" {
		return "i3"
	}

	// Check desktop session
	session := strings.ToLower(os.Getenv("DESKTOP_SESSION"))
	switch session {
	case "gnome", "gnome-classic", "gnome-flashback", "ubuntu":
		return "gnome"
	case "kde", "kde-plasma", "plasma":
		return "kde"
	case "xfce", "xfce4", "xubuntu":
		return "xfce"
	case "mate":
		return "mate"
	case "cinnamon":
		return "cinnamon"
	case "lxde", "lubuntu":
		return "lxde"
	case "lxqt":
		return "lxqt"
	}

	return "unknown"
}

// checkDBusAvailable checks if D-Bus is available
func checkDBusAvailable() bool {
	// Check for session bus
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") != "" {
		return true
	}

	// Check for system bus socket
	if _, err := os.Stat("/var/run/dbus/system_bus_socket"); err == nil {
		return true
	}

	return false
}

// checkSystemdAvailable checks if systemd is available
func checkSystemdAvailable() bool {
	// Check if we're running under systemd
	if os.Getenv("SYSTEMD_EXEC_PID") != "" {
		return true
	}

	// Check for systemd process
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}

	// Check if systemctl exists
	if _, err := os.Stat("/usr/bin/systemctl"); err == nil {
		return true
	}
	if _, err := os.Stat("/bin/systemctl"); err == nil {
		return true
	}

	return false
}

// IsWayland returns true if running on Wayland
func (e *Environment) IsWayland() bool {
	return e.DisplayServer == "wayland"
}

// IsX11 returns true if running on X11
func (e *Environment) IsX11() bool {
	return e.DisplayServer == "x11"
}

// IsConsole returns true if running in console/TTY
func (e *Environment) IsConsole() bool {
	return e.DisplayServer == "console"
}

// SuggestedProviders returns a list of suggested provider names based on the environment
func (e *Environment) SuggestedProviders() []string {
	providers := make([]string, 0)

	// Desktop-specific providers first
	switch e.DesktopEnv {
	case "gnome":
		providers = append(providers, "dbus-gnome")
	case "kde":
		providers = append(providers, "dbus-kde")
	case "xfce":
		providers = append(providers, "dbus-xfce")
	case "mate":
		providers = append(providers, "dbus-mate")
	case "cinnamon":
		providers = append(providers, "dbus-cinnamon")
	case "sway", "hyprland":
		if e.IsWayland() {
			providers = append(providers, "wayland")
		}
	}

	// Generic D-Bus provider
	if e.HasDBus {
		providers = append(providers, "dbus")
	}

	// Display server specific providers
	if e.IsWayland() {
		providers = append(providers, "wayland")
	}
	if e.IsX11() {
		providers = append(providers, "x11")
	}

	// Systemd as fallback
	if e.HasSystemd {
		providers = append(providers, "systemd")
	}

	// Always include fallback
	providers = append(providers, "fallback")

	return providers
}
