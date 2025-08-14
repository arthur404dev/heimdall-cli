package providers

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/detector"
)

// X11Cookie represents an X11 inhibition cookie
type X11Cookie struct {
	method string // "xscreensaver", "dpms", or "reset"
}

func (c X11Cookie) String() string {
	return fmt.Sprintf("x11:%s", c.method)
}

// X11Provider implements idle prevention for X11
type X11Provider struct {
	mu              sync.Mutex
	env             *detector.Environment
	active          bool
	stopChan        chan struct{}
	hasXset         bool
	hasXScreensaver bool
}

// NewX11Provider creates a new X11 provider
func NewX11Provider() *X11Provider {
	env := detector.Detect()
	provider := &X11Provider{
		env: env,
	}

	// Check for available tools
	provider.checkTools()

	return provider
}

// checkTools checks which X11 tools are available
func (p *X11Provider) checkTools() {
	// Check for xset
	if _, err := exec.LookPath("xset"); err == nil {
		p.hasXset = true
	}

	// Check for xscreensaver-command
	if _, err := exec.LookPath("xscreensaver-command"); err == nil {
		p.hasXScreensaver = true
	}
}

// Name returns the provider name
func (p *X11Provider) Name() string {
	return "x11"
}

// Available checks if X11 is available
func (p *X11Provider) Available() bool {
	if !p.env.IsX11() {
		return false
	}

	// Need at least one tool available
	return p.hasXset || p.hasXScreensaver
}

// Priority returns the provider priority
func (p *X11Provider) Priority() int {
	// Lower priority than D-Bus providers
	return 50
}

// Inhibit creates an idle inhibition
func (p *X11Provider) Inhibit(reason string) (Cookie, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return X11Cookie{method: p.getActiveMethod()}, nil
	}

	var method string
	var err error

	// Try XScreensaver first
	if p.hasXScreensaver {
		err = p.disableXScreensaver()
		if err == nil {
			method = "xscreensaver"
		}
	}

	// Try DPMS
	if method == "" && p.hasXset {
		err = p.disableDPMS()
		if err == nil {
			method = "dpms"
		}
	}

	// Fallback to idle reset loop
	if method == "" {
		p.startIdleResetLoop()
		method = "reset"
		err = nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to inhibit idle: %w", err)
	}

	p.active = true
	return X11Cookie{method: method}, nil
}

// getActiveMethod returns the currently active inhibition method
func (p *X11Provider) getActiveMethod() string {
	// Check what's currently active
	if p.stopChan != nil {
		return "reset"
	}

	// Check DPMS status
	if p.hasXset {
		cmd := exec.Command("xset", "q")
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "DPMS is Disabled") {
			return "dpms"
		}
	}

	return "xscreensaver"
}

// disableXScreensaver disables the XScreensaver
func (p *X11Provider) disableXScreensaver() error {
	// Deactivate any active screensaver
	cmd := exec.Command("xscreensaver-command", "-deactivate")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Exit the xscreensaver daemon (will be restarted by session manager if needed)
	cmd = exec.Command("xscreensaver-command", "-exit")
	return cmd.Run()
}

// enableXScreensaver re-enables the XScreensaver
func (p *X11Provider) enableXScreensaver() error {
	// Restart xscreensaver daemon
	cmd := exec.Command("xscreensaver", "-no-splash")
	return cmd.Start()
}

// disableDPMS disables DPMS (Display Power Management Signaling)
func (p *X11Provider) disableDPMS() error {
	// Disable DPMS
	cmd := exec.Command("xset", "-dpms")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Also disable screen saver
	cmd = exec.Command("xset", "s", "off")
	return cmd.Run()
}

// enableDPMS re-enables DPMS
func (p *X11Provider) enableDPMS() error {
	// Enable DPMS
	cmd := exec.Command("xset", "+dpms")
	if err := cmd.Run(); err != nil {
		return err
	}

	// Re-enable screen saver with default timeout
	cmd = exec.Command("xset", "s", "on")
	return cmd.Run()
}

// startIdleResetLoop starts a loop that resets the idle timer periodically
func (p *X11Provider) startIdleResetLoop() {
	p.stopChan = make(chan struct{})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Reset idle timer
				if p.hasXset {
					exec.Command("xset", "s", "reset").Run()
				}
				// Move mouse by 0 pixels (also resets idle)
				exec.Command("xdotool", "mousemove_relative", "0", "0").Run()

			case <-p.stopChan:
				return
			}
		}
	}()
}

// stopIdleResetLoop stops the idle reset loop
func (p *X11Provider) stopIdleResetLoop() {
	if p.stopChan != nil {
		close(p.stopChan)
		p.stopChan = nil
	}
}

// Uninhibit releases an idle inhibition
func (p *X11Provider) Uninhibit(cookie Cookie) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	x11Cookie, ok := cookie.(X11Cookie)
	if !ok {
		return fmt.Errorf("invalid cookie type for X11 provider")
	}

	var err error

	switch x11Cookie.method {
	case "xscreensaver":
		err = p.enableXScreensaver()
	case "dpms":
		err = p.enableDPMS()
	case "reset":
		p.stopIdleResetLoop()
	default:
		err = fmt.Errorf("unknown X11 method: %s", x11Cookie.method)
	}

	if err == nil {
		p.active = false
	}

	return err
}

// Status returns whether an inhibition is currently active
func (p *X11Provider) Status() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.active, nil
}
