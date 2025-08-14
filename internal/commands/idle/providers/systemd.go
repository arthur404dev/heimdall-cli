package providers

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/detector"
)

// SystemdCookie represents a systemd inhibition cookie
type SystemdCookie struct {
	fd   int
	pid  int
	what string
}

func (c SystemdCookie) String() string {
	return fmt.Sprintf("systemd:fd=%d:pid=%d:what=%s", c.fd, c.pid, c.what)
}

// SystemdProvider implements idle prevention using systemd-inhibit
type SystemdProvider struct {
	mu     sync.Mutex
	env    *detector.Environment
	active bool
	cmd    *exec.Cmd
	pid    int
}

// NewSystemdProvider creates a new systemd provider
func NewSystemdProvider() *SystemdProvider {
	env := detector.Detect()
	return &SystemdProvider{
		env: env,
	}
}

// Name returns the provider name
func (p *SystemdProvider) Name() string {
	return "systemd"
}

// Available checks if systemd-inhibit is available
func (p *SystemdProvider) Available() bool {
	if !p.env.HasSystemd {
		return false
	}

	// Check if systemd-inhibit exists
	_, err := exec.LookPath("systemd-inhibit")
	return err == nil
}

// Priority returns the provider priority
func (p *SystemdProvider) Priority() int {
	// Lower priority - use as fallback
	return 30
}

// Inhibit creates an idle inhibition
func (p *SystemdProvider) Inhibit(reason string) (Cookie, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return SystemdCookie{pid: p.pid, what: "idle:sleep"}, nil
	}

	// Build systemd-inhibit command
	// --what: idle:sleep:shutdown:handle-power-key:handle-suspend-key:handle-hibernate-key:handle-lid-switch
	// We'll use idle:sleep for basic idle/sleep prevention
	args := []string{
		"--what=idle:sleep",
		"--who=heimdall",
		"--why=" + reason,
		"--mode=block",
		"sleep", "infinity", // Keep running indefinitely
	}

	p.cmd = exec.Command("systemd-inhibit", args...)

	// Start the inhibitor process
	if err := p.cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start systemd-inhibit: %w", err)
	}

	p.pid = p.cmd.Process.Pid
	p.active = true

	// Monitor the process in a goroutine
	go func() {
		p.cmd.Wait()
		p.mu.Lock()
		p.active = false
		p.cmd = nil
		p.pid = 0
		p.mu.Unlock()
	}()

	return SystemdCookie{
		pid:  p.pid,
		what: "idle:sleep",
	}, nil
}

// Uninhibit releases an idle inhibition
func (p *SystemdProvider) Uninhibit(cookie Cookie) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	systemdCookie, ok := cookie.(SystemdCookie)
	if !ok {
		return fmt.Errorf("invalid cookie type for systemd provider")
	}

	if !p.active || p.cmd == nil {
		return fmt.Errorf("no active inhibition")
	}

	if p.pid != systemdCookie.pid {
		return fmt.Errorf("cookie PID mismatch")
	}

	// Kill the systemd-inhibit process
	if err := p.cmd.Process.Kill(); err != nil {
		// Try SIGTERM if SIGKILL fails
		p.cmd.Process.Signal(os.Interrupt)
	}

	p.active = false
	p.cmd = nil
	p.pid = 0

	return nil
}

// Status returns whether an inhibition is currently active
func (p *SystemdProvider) Status() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return false, nil
	}

	// Double-check by listing inhibitors
	cmd := exec.Command("systemd-inhibit", "--list")
	output, err := cmd.Output()
	if err != nil {
		return p.active, nil // Fall back to our tracked state
	}

	// Check if our inhibitor is in the list
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "heimdall") && strings.Contains(line, strconv.Itoa(p.pid)) {
			return true, nil
		}
	}

	// Our inhibitor is not in the list, update state
	p.active = false
	p.cmd = nil
	p.pid = 0

	return false, nil
}
