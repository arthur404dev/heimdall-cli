package providers

import (
	"fmt"
	"sync"

	"github.com/arthur404dev/heimdall-cli/internal/commands/idle/detector"
	"github.com/godbus/dbus/v5"
)

// DBusCookie represents a D-Bus inhibition cookie
type DBusCookie struct {
	conn   *dbus.Conn
	cookie uint32
	iface  string
}

func (c DBusCookie) String() string {
	return fmt.Sprintf("dbus:%s:%d", c.iface, c.cookie)
}

// DBusProvider implements idle prevention using D-Bus
type DBusProvider struct {
	mu           sync.Mutex
	env          *detector.Environment
	conn         *dbus.Conn
	activeCookie *DBusCookie
	interfaces   []dbusInterface
}

// dbusInterface represents a D-Bus interface for idle inhibition
type dbusInterface struct {
	name            string
	service         string
	path            string
	inhibitMethod   string
	uninhibitMethod string
	priority        int
}

var knownInterfaces = []dbusInterface{
	// GNOME Session Manager
	{
		name:            "gnome",
		service:         "org.gnome.SessionManager",
		path:            "/org/gnome/SessionManager",
		inhibitMethod:   "org.gnome.SessionManager.Inhibit",
		uninhibitMethod: "org.gnome.SessionManager.Uninhibit",
		priority:        100,
	},
	// KDE Power Management
	{
		name:            "kde",
		service:         "org.kde.Solid.PowerManagement.PolicyAgent",
		path:            "/org/kde/Solid/PowerManagement/PolicyAgent",
		inhibitMethod:   "org.kde.Solid.PowerManagement.PolicyAgent.AddInhibition",
		uninhibitMethod: "org.kde.Solid.PowerManagement.PolicyAgent.ReleaseInhibition",
		priority:        100,
	},
	// Freedesktop ScreenSaver (generic)
	{
		name:            "freedesktop",
		service:         "org.freedesktop.ScreenSaver",
		path:            "/org/freedesktop/ScreenSaver",
		inhibitMethod:   "org.freedesktop.ScreenSaver.Inhibit",
		uninhibitMethod: "org.freedesktop.ScreenSaver.UnInhibit",
		priority:        50,
	},
	// XFCE Power Manager
	{
		name:            "xfce",
		service:         "org.xfce.PowerManager",
		path:            "/org/xfce/PowerManager",
		inhibitMethod:   "org.xfce.PowerManager.Inhibit",
		uninhibitMethod: "org.xfce.PowerManager.UnInhibit",
		priority:        80,
	},
	// MATE Session Manager
	{
		name:            "mate",
		service:         "org.mate.SessionManager",
		path:            "/org/mate/SessionManager",
		inhibitMethod:   "org.mate.SessionManager.Inhibit",
		uninhibitMethod: "org.mate.SessionManager.Uninhibit",
		priority:        80,
	},
	// Cinnamon Session Manager
	{
		name:            "cinnamon",
		service:         "org.cinnamon.SessionManager",
		path:            "/org/cinnamon/SessionManager",
		inhibitMethod:   "org.cinnamon.SessionManager.Inhibit",
		uninhibitMethod: "org.cinnamon.SessionManager.Uninhibit",
		priority:        80,
	},
	// Freedesktop PowerManagement (older)
	{
		name:            "freedesktop-pm",
		service:         "org.freedesktop.PowerManagement",
		path:            "/org/freedesktop/PowerManagement/Inhibit",
		inhibitMethod:   "org.freedesktop.PowerManagement.Inhibit.Inhibit",
		uninhibitMethod: "org.freedesktop.PowerManagement.Inhibit.UnInhibit",
		priority:        40,
	},
}

// NewDBusProvider creates a new D-Bus provider
func NewDBusProvider() *DBusProvider {
	env := detector.Detect()
	provider := &DBusProvider{
		env:        env,
		interfaces: make([]dbusInterface, 0),
	}

	// Select interfaces based on desktop environment
	provider.selectInterfaces()

	return provider
}

// selectInterfaces chooses which D-Bus interfaces to try based on the environment
func (p *DBusProvider) selectInterfaces() {
	// Add desktop-specific interface first
	switch p.env.DesktopEnv {
	case "gnome":
		for _, iface := range knownInterfaces {
			if iface.name == "gnome" {
				p.interfaces = append(p.interfaces, iface)
				break
			}
		}
	case "kde":
		for _, iface := range knownInterfaces {
			if iface.name == "kde" {
				p.interfaces = append(p.interfaces, iface)
				break
			}
		}
	case "xfce":
		for _, iface := range knownInterfaces {
			if iface.name == "xfce" {
				p.interfaces = append(p.interfaces, iface)
				break
			}
		}
	case "mate":
		for _, iface := range knownInterfaces {
			if iface.name == "mate" {
				p.interfaces = append(p.interfaces, iface)
				break
			}
		}
	case "cinnamon":
		for _, iface := range knownInterfaces {
			if iface.name == "cinnamon" {
				p.interfaces = append(p.interfaces, iface)
				break
			}
		}
	}

	// Add generic interfaces as fallbacks
	for _, iface := range knownInterfaces {
		if iface.name == "freedesktop" || iface.name == "freedesktop-pm" {
			p.interfaces = append(p.interfaces, iface)
		}
	}
}

// Name returns the provider name
func (p *DBusProvider) Name() string {
	return "dbus"
}

// Available checks if D-Bus is available
func (p *DBusProvider) Available() bool {
	if !p.env.HasDBus {
		return false
	}

	// Try to connect to session bus
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return false
	}
	defer conn.Close()

	// Check if at least one interface is available
	for _, iface := range p.interfaces {
		obj := conn.Object(iface.service, dbus.ObjectPath(iface.path))
		// Try to introspect the object
		var result string
		err := obj.Call("org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&result)
		if err == nil && result != "" {
			return true
		}
	}

	return false
}

// Priority returns the provider priority
func (p *DBusProvider) Priority() int {
	// Higher priority for desktop-specific providers
	switch p.env.DesktopEnv {
	case "gnome", "kde", "xfce", "mate", "cinnamon":
		return 90
	default:
		return 70
	}
}

// Inhibit creates an idle inhibition
func (p *DBusProvider) Inhibit(reason string) (Cookie, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If already inhibited, return existing cookie
	if p.activeCookie != nil {
		return p.activeCookie, nil
	}

	// Connect to session bus
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to D-Bus: %w", err)
	}

	// Try each interface until one works
	for _, iface := range p.interfaces {
		cookie, err := p.tryInhibit(conn, iface, reason)
		if err == nil {
			p.conn = conn
			p.activeCookie = &DBusCookie{
				conn:   conn,
				cookie: cookie,
				iface:  iface.name,
			}
			return p.activeCookie, nil
		}
	}

	conn.Close()
	return nil, fmt.Errorf("no D-Bus interface available for idle inhibition")
}

// tryInhibit attempts to inhibit using a specific interface
func (p *DBusProvider) tryInhibit(conn *dbus.Conn, iface dbusInterface, reason string) (uint32, error) {
	obj := conn.Object(iface.service, dbus.ObjectPath(iface.path))

	var cookie uint32
	var err error

	// Different interfaces have different method signatures
	switch iface.name {
	case "gnome", "mate", "cinnamon":
		// GNOME-style: Inhibit(app_id, toplevel_xid, reason, flags) -> cookie
		// flags: 8 = Inhibit idle, 4 = Inhibit suspend
		err = obj.Call(iface.inhibitMethod, 0, "heimdall", uint32(0), reason, uint32(12)).Store(&cookie)

	case "kde":
		// KDE-style: AddInhibition(type, reason) -> cookie
		// type: 1 = Screen, 2 = Sleep, 4 = Idle
		err = obj.Call(iface.inhibitMethod, 0, uint32(7), reason).Store(&cookie)

	case "xfce":
		// XFCE-style: Inhibit(app_name, reason) -> cookie
		err = obj.Call(iface.inhibitMethod, 0, "heimdall", reason).Store(&cookie)

	case "freedesktop", "freedesktop-pm":
		// Freedesktop-style: Inhibit(app_name, reason) -> cookie
		err = obj.Call(iface.inhibitMethod, 0, "heimdall", reason).Store(&cookie)

	default:
		// Generic attempt
		err = obj.Call(iface.inhibitMethod, 0, "heimdall", reason).Store(&cookie)
	}

	if err != nil {
		return 0, err
	}

	return cookie, nil
}

// Uninhibit releases an idle inhibition
func (p *DBusProvider) Uninhibit(cookie Cookie) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbusCookie, ok := cookie.(*DBusCookie)
	if !ok {
		return fmt.Errorf("invalid cookie type for D-Bus provider")
	}

	if p.conn == nil {
		return fmt.Errorf("no active D-Bus connection")
	}

	// Find the interface that was used
	var targetInterface *dbusInterface
	for _, iface := range p.interfaces {
		if iface.name == dbusCookie.iface {
			targetInterface = &iface
			break
		}
	}

	if targetInterface == nil {
		return fmt.Errorf("unknown interface: %s", dbusCookie.iface)
	}

	obj := p.conn.Object(targetInterface.service, dbus.ObjectPath(targetInterface.path))
	err := obj.Call(targetInterface.uninhibitMethod, 0, dbusCookie.cookie).Err

	if err == nil {
		p.conn.Close()
		p.conn = nil
		p.activeCookie = nil
	}

	return err
}

// Status returns whether an inhibition is currently active
func (p *DBusProvider) Status() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Note: We can't reliably check D-Bus inhibition status without the cookie
	// This is a limitation of the D-Bus screensaver interface
	// The cookie is only valid for the process that created it
	return p.activeCookie != nil, nil
}
