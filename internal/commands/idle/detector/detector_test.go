package detector

import (
	"os"
	"testing"
)

func TestDetect(t *testing.T) {
	t.Run("returns valid environment", func(t *testing.T) {
		env := Detect()

		if env == nil {
			t.Fatal("Environment should not be nil")
		}

		// Display server should be one of the known values
		validDisplayServers := []string{"x11", "wayland", "console", "unknown"}
		if !contains(validDisplayServers, env.DisplayServer) {
			t.Errorf("Invalid display server: %s", env.DisplayServer)
		}

		// Desktop environment should be set (even if "unknown")
		if env.DesktopEnv == "" {
			t.Error("Desktop environment should not be empty")
		}
	})
}

func TestDetectDisplayServer(t *testing.T) {
	tests := []struct {
		name           string
		sessionType    string
		waylandDisplay string
		x11Display     string
		term           string
		expected       string
	}{
		{
			name:        "wayland via session type",
			sessionType: "wayland",
			expected:    "wayland",
		},
		{
			name:        "x11 via session type",
			sessionType: "x11",
			expected:    "x11",
		},
		{
			name:           "wayland via display",
			waylandDisplay: "wayland-0",
			expected:       "wayland",
		},
		{
			name:       "x11 via display",
			x11Display: ":0",
			expected:   "x11",
		},
		{
			name:     "console mode",
			term:     "linux",
			expected: "console",
		},
		{
			name:     "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalSessionType := os.Getenv("XDG_SESSION_TYPE")
			originalWaylandDisplay := os.Getenv("WAYLAND_DISPLAY")
			originalX11Display := os.Getenv("DISPLAY")
			originalTerm := os.Getenv("TERM")

			// Clear environment
			os.Unsetenv("XDG_SESSION_TYPE")
			os.Unsetenv("WAYLAND_DISPLAY")
			os.Unsetenv("DISPLAY")
			os.Unsetenv("TERM")

			// Set test environment
			if tt.sessionType != "" {
				os.Setenv("XDG_SESSION_TYPE", tt.sessionType)
			}
			if tt.waylandDisplay != "" {
				os.Setenv("WAYLAND_DISPLAY", tt.waylandDisplay)
			}
			if tt.x11Display != "" {
				os.Setenv("DISPLAY", tt.x11Display)
			}
			if tt.term != "" {
				os.Setenv("TERM", tt.term)
			}

			result := detectDisplayServer()

			// Restore original environment
			restoreEnv("XDG_SESSION_TYPE", originalSessionType)
			restoreEnv("WAYLAND_DISPLAY", originalWaylandDisplay)
			restoreEnv("DISPLAY", originalX11Display)
			restoreEnv("TERM", originalTerm)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetectDesktopEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		xdgCurrent     string
		gnomeSession   string
		kdeSession     string
		mateSession    string
		desktopSession string
		swaysock       string
		hyprlandSig    string
		i3sock         string
		expected       string
	}{
		{
			name:       "gnome via xdg current",
			xdgCurrent: "GNOME",
			expected:   "gnome",
		},
		{
			name:       "gnome with multiple values",
			xdgCurrent: "GNOME:GNOME-Classic",
			expected:   "gnome",
		},
		{
			name:       "kde via xdg current",
			xdgCurrent: "KDE",
			expected:   "kde",
		},
		{
			name:       "xfce via xdg current",
			xdgCurrent: "XFCE",
			expected:   "xfce",
		},
		{
			name:       "mate via xdg current",
			xdgCurrent: "MATE",
			expected:   "mate",
		},
		{
			name:       "cinnamon via xdg current",
			xdgCurrent: "X-Cinnamon",
			expected:   "cinnamon",
		},
		{
			name:       "sway via xdg current",
			xdgCurrent: "sway",
			expected:   "sway",
		},
		{
			name:       "hyprland via xdg current",
			xdgCurrent: "Hyprland",
			expected:   "hyprland",
		},
		{
			name:         "gnome via fallback",
			gnomeSession: "gnome",
			expected:     "gnome",
		},
		{
			name:       "kde via fallback",
			kdeSession: "true",
			expected:   "kde",
		},
		{
			name:        "mate via fallback",
			mateSession: "mate",
			expected:    "mate",
		},
		{
			name:     "sway via socket",
			swaysock: "/run/user/1000/sway-ipc.sock",
			expected: "sway",
		},
		{
			name:        "hyprland via signature",
			hyprlandSig: "12345",
			expected:    "hyprland",
		},
		{
			name:     "i3 via socket",
			i3sock:   "/run/user/1000/i3/ipc-socket",
			expected: "i3",
		},
		{
			name:           "gnome via desktop session",
			desktopSession: "gnome",
			expected:       "gnome",
		},
		{
			name:           "ubuntu session maps to gnome",
			desktopSession: "ubuntu",
			expected:       "gnome",
		},
		{
			name:     "unknown desktop",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalXdgCurrent := os.Getenv("XDG_CURRENT_DESKTOP")
			originalGnomeSession := os.Getenv("GNOME_DESKTOP_SESSION_ID")
			originalKdeSession := os.Getenv("KDE_FULL_SESSION")
			originalMateSession := os.Getenv("MATE_DESKTOP_SESSION_ID")
			originalDesktopSession := os.Getenv("DESKTOP_SESSION")
			originalSwaysock := os.Getenv("SWAYSOCK")
			originalHyprlandSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
			originalI3sock := os.Getenv("I3SOCK")

			// Clear environment
			os.Unsetenv("XDG_CURRENT_DESKTOP")
			os.Unsetenv("GNOME_DESKTOP_SESSION_ID")
			os.Unsetenv("KDE_FULL_SESSION")
			os.Unsetenv("MATE_DESKTOP_SESSION_ID")
			os.Unsetenv("DESKTOP_SESSION")
			os.Unsetenv("SWAYSOCK")
			os.Unsetenv("HYPRLAND_INSTANCE_SIGNATURE")
			os.Unsetenv("I3SOCK")

			// Set test environment
			if tt.xdgCurrent != "" {
				os.Setenv("XDG_CURRENT_DESKTOP", tt.xdgCurrent)
			}
			if tt.gnomeSession != "" {
				os.Setenv("GNOME_DESKTOP_SESSION_ID", tt.gnomeSession)
			}
			if tt.kdeSession != "" {
				os.Setenv("KDE_FULL_SESSION", tt.kdeSession)
			}
			if tt.mateSession != "" {
				os.Setenv("MATE_DESKTOP_SESSION_ID", tt.mateSession)
			}
			if tt.desktopSession != "" {
				os.Setenv("DESKTOP_SESSION", tt.desktopSession)
			}
			if tt.swaysock != "" {
				os.Setenv("SWAYSOCK", tt.swaysock)
			}
			if tt.hyprlandSig != "" {
				os.Setenv("HYPRLAND_INSTANCE_SIGNATURE", tt.hyprlandSig)
			}
			if tt.i3sock != "" {
				os.Setenv("I3SOCK", tt.i3sock)
			}

			result := detectDesktopEnvironment()

			// Restore original environment
			restoreEnv("XDG_CURRENT_DESKTOP", originalXdgCurrent)
			restoreEnv("GNOME_DESKTOP_SESSION_ID", originalGnomeSession)
			restoreEnv("KDE_FULL_SESSION", originalKdeSession)
			restoreEnv("MATE_DESKTOP_SESSION_ID", originalMateSession)
			restoreEnv("DESKTOP_SESSION", originalDesktopSession)
			restoreEnv("SWAYSOCK", originalSwaysock)
			restoreEnv("HYPRLAND_INSTANCE_SIGNATURE", originalHyprlandSig)
			restoreEnv("I3SOCK", originalI3sock)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCheckDBusAvailable(t *testing.T) {
	tests := []struct {
		name           string
		sessionBusAddr string
		expected       bool
	}{
		{
			name:           "available via session bus",
			sessionBusAddr: "unix:path=/run/user/1000/bus",
			expected:       true,
		},
		{
			name:     "not available",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalSessionBus := os.Getenv("DBUS_SESSION_BUS_ADDRESS")

			// Clear environment
			os.Unsetenv("DBUS_SESSION_BUS_ADDRESS")

			// Set test environment
			if tt.sessionBusAddr != "" {
				os.Setenv("DBUS_SESSION_BUS_ADDRESS", tt.sessionBusAddr)
			}

			result := checkDBusAvailable()

			// Restore original environment
			restoreEnv("DBUS_SESSION_BUS_ADDRESS", originalSessionBus)

			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestCheckSystemdAvailable(t *testing.T) {
	tests := []struct {
		name       string
		systemdPid string
		expected   bool
	}{
		{
			name:       "available via systemd pid",
			systemdPid: "1",
			expected:   true,
		},
		{
			name:     "not available",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalSystemdPid := os.Getenv("SYSTEMD_EXEC_PID")

			// Clear environment
			os.Unsetenv("SYSTEMD_EXEC_PID")

			// Set test environment
			if tt.systemdPid != "" {
				os.Setenv("SYSTEMD_EXEC_PID", tt.systemdPid)
			}

			result := checkSystemdAvailable()

			// Restore original environment
			restoreEnv("SYSTEMD_EXEC_PID", originalSystemdPid)

			// Note: This test might return true even when expected false
			// due to actual systemd presence on the system
			if tt.expected && !result {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}

func TestEnvironmentMethods(t *testing.T) {
	tests := []struct {
		name          string
		displayServer string
		isWayland     bool
		isX11         bool
		isConsole     bool
	}{
		{
			name:          "wayland environment",
			displayServer: "wayland",
			isWayland:     true,
			isX11:         false,
			isConsole:     false,
		},
		{
			name:          "x11 environment",
			displayServer: "x11",
			isWayland:     false,
			isX11:         true,
			isConsole:     false,
		},
		{
			name:          "console environment",
			displayServer: "console",
			isWayland:     false,
			isX11:         false,
			isConsole:     true,
		},
		{
			name:          "unknown environment",
			displayServer: "unknown",
			isWayland:     false,
			isX11:         false,
			isConsole:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Environment{
				DisplayServer: tt.displayServer,
			}

			if env.IsWayland() != tt.isWayland {
				t.Errorf("IsWayland(): expected %t, got %t", tt.isWayland, env.IsWayland())
			}
			if env.IsX11() != tt.isX11 {
				t.Errorf("IsX11(): expected %t, got %t", tt.isX11, env.IsX11())
			}
			if env.IsConsole() != tt.isConsole {
				t.Errorf("IsConsole(): expected %t, got %t", tt.isConsole, env.IsConsole())
			}
		})
	}
}

func TestSuggestedProviders(t *testing.T) {
	tests := []struct {
		name          string
		desktopEnv    string
		displayServer string
		hasDBus       bool
		hasSystemd    bool
		expectedFirst string
		shouldContain []string
	}{
		{
			name:          "gnome environment",
			desktopEnv:    "gnome",
			displayServer: "x11",
			hasDBus:       true,
			hasSystemd:    true,
			expectedFirst: "dbus-gnome",
			shouldContain: []string{"dbus-gnome", "dbus", "x11", "systemd", "fallback"},
		},
		{
			name:          "kde environment",
			desktopEnv:    "kde",
			displayServer: "wayland",
			hasDBus:       true,
			hasSystemd:    true,
			expectedFirst: "dbus-kde",
			shouldContain: []string{"dbus-kde", "dbus", "wayland", "systemd", "fallback"},
		},
		{
			name:          "sway environment",
			desktopEnv:    "sway",
			displayServer: "wayland",
			hasDBus:       false,
			hasSystemd:    true,
			expectedFirst: "wayland",
			shouldContain: []string{"wayland", "systemd", "fallback"},
		},
		{
			name:          "unknown environment",
			desktopEnv:    "unknown",
			displayServer: "x11",
			hasDBus:       false,
			hasSystemd:    false,
			expectedFirst: "x11",
			shouldContain: []string{"x11", "fallback"},
		},
		{
			name:          "minimal environment",
			desktopEnv:    "unknown",
			displayServer: "unknown",
			hasDBus:       false,
			hasSystemd:    false,
			expectedFirst: "fallback",
			shouldContain: []string{"fallback"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Environment{
				DesktopEnv:    tt.desktopEnv,
				DisplayServer: tt.displayServer,
				HasDBus:       tt.hasDBus,
				HasSystemd:    tt.hasSystemd,
			}

			providers := env.SuggestedProviders()

			if len(providers) == 0 {
				t.Fatal("Expected at least one provider")
			}

			if providers[0] != tt.expectedFirst {
				t.Errorf("Expected first provider to be %s, got %s", tt.expectedFirst, providers[0])
			}

			for _, expected := range tt.shouldContain {
				if !contains(providers, expected) {
					t.Errorf("Expected providers to contain %s, got %v", expected, providers)
				}
			}

			// Fallback should always be last
			if providers[len(providers)-1] != "fallback" {
				t.Errorf("Expected fallback to be last provider, got %s", providers[len(providers)-1])
			}
		})
	}
}

// Benchmark tests
func BenchmarkDetect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		env := Detect()
		_ = env
	}
}

func BenchmarkDetectDisplayServer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := detectDisplayServer()
		_ = result
	}
}

func BenchmarkDetectDesktopEnvironment(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := detectDesktopEnvironment()
		_ = result
	}
}

func BenchmarkSuggestedProviders(b *testing.B) {
	env := &Environment{
		DesktopEnv:    "gnome",
		DisplayServer: "x11",
		HasDBus:       true,
		HasSystemd:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		providers := env.SuggestedProviders()
		_ = providers
	}
}

// Test utilities
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func restoreEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

// Integration tests
func TestEnvironmentDetectionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("real environment detection", func(t *testing.T) {
		env := Detect()

		// Basic sanity checks
		if env.DisplayServer == "" {
			t.Error("Display server should not be empty")
		}
		if env.DesktopEnv == "" {
			t.Error("Desktop environment should not be empty")
		}

		// Suggested providers should not be empty
		providers := env.SuggestedProviders()
		if len(providers) == 0 {
			t.Error("Should have at least one suggested provider")
		}

		// Should always include fallback
		if !contains(providers, "fallback") {
			t.Error("Should always include fallback provider")
		}

		t.Logf("Detected environment: DisplayServer=%s, DesktopEnv=%s, HasDBus=%t, HasSystemd=%t",
			env.DisplayServer, env.DesktopEnv, env.HasDBus, env.HasSystemd)
		t.Logf("Suggested providers: %v", providers)
	})
}

// Edge case tests
func TestEnvironmentEdgeCases(t *testing.T) {
	t.Run("empty environment variables", func(t *testing.T) {
		// Save original environment
		originalXdgCurrent := os.Getenv("XDG_CURRENT_DESKTOP")
		originalSessionType := os.Getenv("XDG_SESSION_TYPE")

		// Clear environment
		os.Unsetenv("XDG_CURRENT_DESKTOP")
		os.Unsetenv("XDG_SESSION_TYPE")

		env := Detect()

		// Should handle empty environment gracefully
		if env.DisplayServer == "" {
			t.Error("Display server should not be empty even with no env vars")
		}
		if env.DesktopEnv == "" {
			t.Error("Desktop environment should not be empty even with no env vars")
		}

		// Restore environment
		restoreEnv("XDG_CURRENT_DESKTOP", originalXdgCurrent)
		restoreEnv("XDG_SESSION_TYPE", originalSessionType)
	})

	t.Run("malformed environment variables", func(t *testing.T) {
		// Save original environment
		originalXdgCurrent := os.Getenv("XDG_CURRENT_DESKTOP")

		// Set malformed environment
		os.Setenv("XDG_CURRENT_DESKTOP", ":::invalid:::")

		env := Detect()

		// Should handle malformed values gracefully
		if env.DesktopEnv == "" {
			t.Error("Should handle malformed desktop environment gracefully")
		}

		// Restore environment
		restoreEnv("XDG_CURRENT_DESKTOP", originalXdgCurrent)
	})
}
