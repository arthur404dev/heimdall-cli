# GTK Live Reload System Plan

## Context

### Problem Statement
GTK applications currently require manual restart to apply theme changes, creating a poor developer experience and limiting the ability to preview theme modifications in real-time. This impacts both theme developers and end users who want instant visual feedback when customizing their desktop appearance.

### Current State
- GTK themes are written to static CSS files
- Applications read theme files only at startup
- No mechanism for runtime theme updates
- Manual application restart required for changes
- No integration with heimdall's scheme change events

### Goals
- Enable instant theme updates without application restart
- Integrate with heimdall's existing scheme change pipeline
- Support both GTK3 and GTK4 applications
- Minimize performance impact and resource usage
- Provide fallback mechanisms for unsupported applications

### Constraints
- Must work across different desktop environments
- Cannot modify GTK library internals
- Must handle applications with varying reload capabilities
- Should not break existing theme application methods
- Must be opt-in to avoid unexpected behavior

## Specification

### Functional Requirements
- Detect theme file changes in real-time
- Notify running GTK applications of theme updates
- Trigger theme reload in compatible applications
- Provide force-refresh mechanisms for stubborn apps
- Integrate with heimdall scheme change events
- Support batch updates to avoid reload storms

### Non-Functional Requirements
- Detection latency < 10ms
- Reload trigger latency < 50ms
- CPU usage < 1% during monitoring
- Memory footprint < 10MB for daemon
- Support 100+ simultaneous applications
- Zero data corruption during updates

### Interfaces
- File system monitoring API
- D-Bus service for theme notifications
- XSettings protocol for X11 environments
- GTK Settings API integration
- IPC mechanism for heimdall events
- CLI commands for manual control

## Live Reload Architecture

### Detection Mechanisms

#### File System Monitoring

```go
type ThemeWatcher struct {
    paths    []string           // GTK3/4 config paths
    watcher  *fsnotify.Watcher
    debounce time.Duration      // Prevent rapid fires
    events   chan ThemeChange
}

// Monitored paths
~/.config/gtk-3.0/gtk.css
~/.config/gtk-4.0/gtk.css
~/.config/gtk-3.0/settings.ini
~/.config/gtk-4.0/settings.ini
/usr/share/themes/*/gtk-*/gtk.css
```

#### Heimdall Event Integration

```go
type SchemeChangeListener struct {
    socket   string              // Unix socket path
    handler  func(SchemeEvent)
    filters  []EventFilter       // Only theme-related
}

// Event flow
Scheme Change → Theme Generation → File Write → Reload Trigger
```

#### Change Detection Strategy
- Use inotify on Linux for efficient monitoring
- Implement polling fallback for unsupported systems
- Watch both user and system theme directories
- Monitor parent directories for atomic moves
- Track file checksums to detect actual changes

### Communication Methods

#### D-Bus Service
```xml
<!-- org.heimdall.ThemeReloader -->
<interface name="org.heimdall.ThemeReloader">
    <method name="ReloadTheme">
        <arg name="theme_name" type="s" direction="in"/>
        <arg name="variant" type="s" direction="in"/>
    </method>
    <signal name="ThemeChanged">
        <arg name="theme_name" type="s"/>
        <arg name="variant" type="s"/>
        <arg name="timestamp" type="t"/>
    </signal>
</interface>
```

#### XSettings Daemon Integration
```c
// XSettings properties to update
Net/ThemeName
Net/IconThemeName
Gtk/ThemeName
Gtk/ColorScheme
Gtk/CursorThemeName

// Increment serial number to trigger reload
_XSETTINGS_S0 serial++
```

#### Portal Integration (Flatpak/Snap)
```go
type PortalInterface struct {
    appearance *AppearancePortal
    settings   *SettingsPortal
}

// Update sandboxed applications
portal.appearance.SetColorScheme(scheme)
portal.settings.Changed("gtk-theme", value)
```

### Integration Points

#### Heimdall Scheme Pipeline
```go
// Hook into existing scheme application
func (a *Applier) Apply(scheme *Scheme) error {
    // Generate theme files
    if err := a.generateGTKTheme(scheme); err != nil {
        return err
    }
    
    // Trigger live reload
    if a.liveReloadEnabled {
        return a.reloadManager.TriggerReload()
    }
    
    return nil
}
```

#### Event Aggregation
```go
type EventAggregator struct {
    window   time.Duration    // Batch window (e.g., 100ms)
    pending  []ChangeEvent
    timer    *time.Timer
}

// Batch rapid changes into single reload
func (a *EventAggregator) Add(event ChangeEvent) {
    a.pending = append(a.pending, event)
    a.resetTimer()
}
```

## Reload Mechanisms

### XSettings Daemon Integration

#### Implementation
```go
type XSettingsManager struct {
    display  *x11.Display
    window   x11.Window
    atoms    map[string]x11.Atom
}

func (m *XSettingsManager) UpdateTheme(name string) error {
    // Update XSettings properties
    settings := m.getCurrentSettings()
    settings["Net/ThemeName"] = name
    settings["Gtk/ThemeName"] = name
    
    // Increment serial to notify clients
    m.incrementSerial()
    
    // Send PropertyNotify event
    return m.broadcastChange()
}
```

#### Advantages
- Native GTK support
- Works with most GTK applications
- Low latency
- Minimal overhead

#### Limitations
- X11 only (not Wayland native)
- Requires running daemon
- May conflict with DE settings daemon

### DBus Signaling

#### GTK Settings Service
```go
type GTKSettingsService struct {
    conn *dbus.Conn
    path dbus.ObjectPath
}

func (s *GTKSettingsService) EmitThemeChange() error {
    signal := &dbus.Signal{
        Name: "org.gtk.Settings.Changed",
        Body: []interface{}{
            "gtk-theme-name",
            "gtk-application-prefer-dark-theme",
        },
    }
    return s.conn.Emit(s.path, signal)
}
```

#### Application-Specific Signals
```go
// GNOME Shell
gsettings set org.gnome.desktop.interface gtk-theme "NewTheme"

// KDE Plasma
qdbus org.kde.KWin /KWin reconfigure

// XFCE
xfconf-query -c xsettings -p /Net/ThemeName -s "NewTheme"
```

### GTK Settings API Usage

#### Runtime Settings Update
```c
// For GTK3
gtk_settings_get_default()
g_object_set(settings, "gtk-theme-name", theme_name, NULL)
gtk_rc_reset_styles(settings)

// For GTK4
gtk_settings_get_default()
g_object_set(settings, "gtk-theme-name", theme_name, NULL)
// GTK4 auto-reloads on settings change
```

#### CSS Provider Reload
```go
type CSSReloader struct {
    provider *gtk.CSSProvider
    screen   *gdk.Screen
}

func (r *CSSReloader) Reload(cssPath string) error {
    // Remove old provider
    gtk.StyleContext.RemoveProviderForScreen(
        r.screen, r.provider)
    
    // Load new CSS
    r.provider.LoadFromPath(cssPath)
    
    // Re-add provider
    gtk.StyleContext.AddProviderForScreen(
        r.screen, r.provider, PRIORITY_USER)
    
    return nil
}
```

### Application-Specific Triggers

#### Electron/Chromium Apps
```javascript
// Inject via DevTools protocol
chrome.devtools.inspectedWindow.reload({
    ignoreCache: true,
    injectedScript: `
        document.querySelectorAll('link[rel="stylesheet"]')
            .forEach(link => {
                link.href = link.href + '?reload=' + Date.now();
            });
    `
});
```

#### Qt Applications
```go
func reloadQtTheme() error {
    // Update Qt platform theme
    os.Setenv("QT_STYLE_OVERRIDE", "gtk2")
    
    // Send signal to Qt apps
    return exec.Command("qdbus", 
        "org.kde.KWin", "/KWin", "reconfigure").Run()
}
```

#### Java/Swing Applications
```go
func reloadSwingTheme() error {
    // Update look and feel
    props := filepath.Join(os.Getenv("HOME"), 
        ".java/.userPrefs/javax/swing/plaf")
    
    // Write new properties
    return ioutil.WriteFile(props, 
        []byte("laf=com.sun.java.swing.plaf.gtk.GTKLookAndFeel"), 
        0644)
}
```

## Implementation Strategy

### File System Monitoring

#### Efficient Watch Strategy
```go
type WatchStrategy struct {
    direct   []string  // Direct file watches
    parent   []string  // Parent directory watches
    recursive bool     // Recursive monitoring
}

func (s *WatchStrategy) Setup() error {
    // Watch specific files for changes
    for _, file := range s.direct {
        s.watcher.Add(file)
    }
    
    // Watch parent dirs for atomic operations
    for _, dir := range s.parent {
        s.watcher.Add(dir)
    }
    
    return nil
}
```

#### Change Detection Pipeline
```go
type ChangePipeline struct {
    stages []ChangeProcessor
}

// Pipeline stages
1. File event received
2. Debounce (aggregate rapid changes)
3. Validate (ensure complete write)
4. Filter (ignore temporary files)
5. Checksum (verify actual change)
6. Classify (determine reload strategy)
7. Execute (trigger appropriate reload)
```

### Event Propagation System

#### Event Bus Architecture
```go
type EventBus struct {
    subscribers map[EventType][]Subscriber
    queue       chan Event
    workers     int
}

type Event struct {
    Type      EventType
    Source    string
    Timestamp time.Time
    Data      interface{}
}

// Event types
ThemeFileChanged
SchemeApplied
ReloadRequested
ReloadCompleted
ReloadFailed
```

#### Priority Queue System
```go
type PriorityQueue struct {
    high   chan ReloadRequest  // User-initiated
    medium chan ReloadRequest  // Scheme changes
    low    chan ReloadRequest  // Auto-detected
}

func (q *PriorityQueue) Process() {
    select {
    case req := <-q.high:
        q.handle(req)
    case req := <-q.medium:
        q.handle(req)
    case req := <-q.low:
        q.handle(req)
    }
}
```

### Debouncing and Throttling

#### Debounce Implementation
```go
type Debouncer struct {
    delay    time.Duration
    timer    *time.Timer
    pending  func()
    mu       sync.Mutex
}

func (d *Debouncer) Trigger(fn func()) {
    d.mu.Lock()
    defer d.mu.Unlock()
    
    d.pending = fn
    
    if d.timer != nil {
        d.timer.Stop()
    }
    
    d.timer = time.AfterFunc(d.delay, func() {
        d.mu.Lock()
        fn := d.pending
        d.pending = nil
        d.mu.Unlock()
        
        if fn != nil {
            fn()
        }
    })
}
```

#### Rate Limiting
```go
type RateLimiter struct {
    rate     int           // Events per second
    burst    int           // Burst capacity
    limiter  *rate.Limiter
}

func (r *RateLimiter) Allow() bool {
    return r.limiter.Allow()
}

// Configuration
defaultLimiter = &RateLimiter{
    rate:  10,  // 10 reloads per second max
    burst: 3,   // Allow 3 instant reloads
}
```

### Error Recovery

#### Retry Logic
```go
type RetryStrategy struct {
    maxAttempts int
    backoff     BackoffStrategy
    timeout     time.Duration
}

func (s *RetryStrategy) Execute(fn func() error) error {
    var lastErr error
    
    for i := 0; i < s.maxAttempts; i++ {
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
            time.Sleep(s.backoff.Next(i))
        }
    }
    
    return fmt.Errorf("failed after %d attempts: %w", 
        s.maxAttempts, lastErr)
}
```

#### Fallback Mechanisms
```go
type FallbackChain struct {
    strategies []ReloadStrategy
}

func (c *FallbackChain) Reload() error {
    for _, strategy := range c.strategies {
        if err := strategy.Reload(); err == nil {
            return nil
        }
        // Log and try next strategy
    }
    return ErrNoWorkingStrategy
}

// Fallback order
1. XSettings update
2. D-Bus signal
3. Settings file touch
4. Application restart
```

## Desktop Environment Integration

### GNOME Integration

#### GSettings Integration
```go
type GNOMEIntegration struct {
    settings *gio.Settings
    shell    *GnomeShellDBus
}

func (g *GNOMEIntegration) ApplyTheme(theme string) error {
    // Update GSettings
    g.settings.Set("gtk-theme", theme)
    
    // Notify GNOME Shell
    g.shell.Eval(`
        Main.loadTheme();
        Main.themeManager.updateTheme();
    `)
    
    // Update window decorations
    g.settings.Set("org.gnome.desktop.wm.preferences", 
        "theme", theme)
    
    return nil
}
```

#### GNOME Shell Extension
```javascript
// extension.js for live reload
const ThemeReloader = GObject.registerClass(
class ThemeReloader extends GObject.Object {
    _init() {
        this._settings = new Gio.Settings({
            schema_id: 'org.gnome.desktop.interface'
        });
        
        this._settings.connect('changed::gtk-theme', 
            () => this._reloadTheme());
    }
    
    _reloadTheme() {
        // Force GTK theme reload
        St.ThemeContext.get_for_stage(global.stage)
            .set_theme(new St.Theme());
    }
});
```

### KDE/Plasma Integration

#### KWin Integration
```go
type KDEIntegration struct {
    kwin     *KWinDBus
    plasma   *PlasmaDBus
    kconfig  *KConfig
}

func (k *KDEIntegration) ApplyTheme(theme string) error {
    // Update KDE config
    k.kconfig.SetGroup("General")
    k.kconfig.WriteEntry("ColorScheme", theme)
    
    // Reconfigure KWin
    k.kwin.Reconfigure()
    
    // Update Plasma theme
    k.plasma.SetTheme(theme)
    
    // Notify all KDE apps
    return k.broadcastKDEThemeChange()
}
```

#### Plasma Widget Support
```qml
// plasmoid for theme monitoring
PlasmaCore.DataSource {
    id: themeWatcher
    engine: "executable"
    connectedSources: ["inotifywait -m ~/.config/gtk-3.0/gtk.css"]
    
    onNewData: {
        // Reload theme
        PlasmaCore.Theme.themeName = "breeze-dark"
        PlasmaCore.Theme.themeName = "breeze-light"
    }
}
```

### XFCE Integration

#### XFConf Integration
```go
type XFCEIntegration struct {
    xfconf *XFConfClient
    xfwm   *XFWMClient
}

func (x *XFCEIntegration) ApplyTheme(theme string) error {
    // Update xsettings
    x.xfconf.SetProperty("/Net/ThemeName", theme)
    
    // Update window manager theme
    x.xfconf.SetProperty("/general/theme", theme)
    
    // Reload xfwm4
    return x.xfwm.Reload()
}
```

### Generic X11/Wayland Support

#### X11 Implementation
```go
type X11Support struct {
    display *x11.Display
    root    x11.Window
}

func (x *X11Support) BroadcastThemeChange() error {
    // Create client message
    event := x11.ClientMessageEvent{
        Type:   x.atoms["_GTK_THEME_CHANGE"],
        Window: x.root,
        Data:   []byte("reload"),
    }
    
    // Send to all windows
    return x.display.SendEvent(x.root, 
        x11.SubstructureNotifyMask, &event)
}
```

#### Wayland Implementation
```go
type WaylandSupport struct {
    compositor string
    protocol   *WaylandProtocol
}

func (w *WaylandSupport) ReloadTheme() error {
    switch w.compositor {
    case "sway":
        return w.reloadSway()
    case "wayfire":
        return w.reloadWayfire()
    case "river":
        return w.reloadRiver()
    default:
        return w.genericReload()
    }
}

func (w *WaylandSupport) genericReload() error {
    // Use portal API for Wayland
    portal := xdg.Desktop.Portal()
    return portal.Settings.Changed("gtk-theme")
}
```

## Application Support Matrix

### Fully Supported Applications

| Application | Method | Latency | Notes |
|-------------|--------|---------|-------|
| GNOME Terminal | XSettings | <10ms | Native support |
| Nautilus | GSettings | <20ms | GNOME integration |
| Gedit | XSettings | <10ms | GTK3 native |
| Evolution | D-Bus | <30ms | Custom handler |
| Rhythmbox | XSettings | <15ms | GTK3 native |
| Firefox | Custom | <100ms | Requires addon |
| Thunderbird | Custom | <100ms | Requires addon |
| LibreOffice | XSettings | <50ms | Partial support |

### Partially Supported Applications

| Application | Method | Limitations |
|-------------|--------|-------------|
| Chrome/Chromium | DevTools | Requires flag |
| VSCode | Reload command | Manual trigger |
| Slack | Restart | No live reload |
| Discord | CSS injection | Requires mod |
| Steam | Restart | No GTK support |
| GIMP | Restart | Theme cached |

### Workarounds for Non-Supporting Apps

#### CSS Injection Method
```go
func InjectCSS(pid int, css string) error {
    // Attach to process
    proc, err := ptrace.Attach(pid)
    if err != nil {
        return err
    }
    defer proc.Detach()
    
    // Find GTK CSS provider
    provider := proc.FindSymbol("gtk_css_provider_new")
    
    // Inject new CSS
    return proc.Call(provider, css)
}
```

#### Window Manager Refresh
```go
func ForceWindowRefresh(windowID uint32) error {
    // Unmap and remap window
    display := x11.OpenDisplay()
    defer display.Close()
    
    display.UnmapWindow(windowID)
    display.Sync()
    display.MapWindow(windowID)
    
    return nil
}
```

#### Process Restart with State
```go
type AppRestarter struct {
    stateDir string
}

func (r *AppRestarter) RestartWithState(app string) error {
    // Save application state
    state := r.captureState(app)
    r.saveState(state)
    
    // Restart application
    if err := r.stopApp(app); err != nil {
        return err
    }
    
    if err := r.startApp(app); err != nil {
        return err
    }
    
    // Restore state
    return r.restoreState(app, state)
}
```

### Known Limitations

#### Technical Limitations
- Statically linked applications cannot be reloaded
- Snap/Flatpak apps have limited access
- Some apps cache themes in memory
- Custom widgets may not update
- OpenGL/Vulkan rendered UIs unaffected

#### Workaround Strategies
```go
type LimitationHandler struct {
    strategies map[string]Strategy
}

func (h *LimitationHandler) Handle(app Application) error {
    switch app.Type {
    case "snap":
        return h.handleSnap(app)
    case "flatpak":
        return h.handleFlatpak(app)
    case "static":
        return h.requestRestart(app)
    case "cached":
        return h.clearCache(app)
    default:
        return h.defaultStrategy(app)
    }
}
```

## Performance Optimization

### Minimize Reload Latency

#### Parallel Processing
```go
type ParallelReloader struct {
    workers int
    queue   chan ReloadTask
}

func (r *ParallelReloader) Reload(apps []Application) {
    var wg sync.WaitGroup
    
    for _, app := range apps {
        wg.Add(1)
        go func(a Application) {
            defer wg.Done()
            r.reloadApp(a)
        }(app)
    }
    
    wg.Wait()
}
```

#### Caching Strategy
```go
type ThemeCache struct {
    compiled map[string]*CompiledTheme
    mu       sync.RWMutex
}

func (c *ThemeCache) Get(name string) *CompiledTheme {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.compiled[name]
}

// Pre-compile themes for instant switching
func (c *ThemeCache) Precompile(themes []string) {
    for _, theme := range themes {
        compiled := compileTheme(theme)
        c.compiled[theme] = compiled
    }
}
```

### Batch Update Strategies

#### Update Coalescing
```go
type UpdateCoalescer struct {
    window   time.Duration
    pending  map[string]Update
    timer    *time.Timer
}

func (c *UpdateCoalescer) Add(update Update) {
    c.pending[update.ID] = update
    
    if c.timer == nil {
        c.timer = time.AfterFunc(c.window, c.flush)
    }
}

func (c *UpdateCoalescer) flush() {
    updates := make([]Update, 0, len(c.pending))
    for _, u := range c.pending {
        updates = append(updates, u)
    }
    
    c.applyBatch(updates)
    c.pending = make(map[string]Update)
    c.timer = nil
}
```

#### Differential Updates
```go
type DiffUpdater struct {
    previous ThemeState
    current  ThemeState
}

func (u *DiffUpdater) GenerateDiff() []Change {
    changes := []Change{}
    
    // Compare color values
    for key, newVal := range u.current.Colors {
        if oldVal, ok := u.previous.Colors[key]; !ok || oldVal != newVal {
            changes = append(changes, ColorChange{key, newVal})
        }
    }
    
    // Only update changed elements
    return changes
}
```

### Selective Component Updates

#### Component Registry
```go
type ComponentRegistry struct {
    components map[string]Component
    deps       map[string][]string
}

func (r *ComponentRegistry) UpdateComponent(name string, change Change) {
    // Update specific component
    if comp, ok := r.components[name]; ok {
        comp.Update(change)
        
        // Update dependent components
        for _, dep := range r.deps[name] {
            r.components[dep].Refresh()
        }
    }
}
```

#### Smart Invalidation
```go
type InvalidationTracker struct {
    dirty map[string]bool
    deps  DependencyGraph
}

func (t *InvalidationTracker) MarkDirty(component string) {
    t.dirty[component] = true
    
    // Mark dependents as dirty
    for _, dep := range t.deps.GetDependents(component) {
        t.dirty[dep] = true
    }
}

func (t *InvalidationTracker) GetDirtyComponents() []string {
    result := []string{}
    for comp, dirty := range t.dirty {
        if dirty {
            result = append(result, comp)
        }
    }
    return result
}
```

### Resource Management

#### Memory Pool
```go
type MemoryPool struct {
    buffers chan []byte
    size    int
}

func (p *MemoryPool) Get() []byte {
    select {
    case buf := <-p.buffers:
        return buf
    default:
        return make([]byte, p.size)
    }
}

func (p *MemoryPool) Put(buf []byte) {
    select {
    case p.buffers <- buf:
    default:
        // Pool full, let GC handle it
    }
}
```

#### Connection Pooling
```go
type ConnectionPool struct {
    dbus    *DBusPool
    x11     *X11Pool
    wayland *WaylandPool
}

func (p *ConnectionPool) GetDBus() *dbus.Conn {
    return p.dbus.Get()
}

func (p *ConnectionPool) Release(conn interface{}) {
    switch c := conn.(type) {
    case *dbus.Conn:
        p.dbus.Put(c)
    case *x11.Display:
        p.x11.Put(c)
    case *WaylandConn:
        p.wayland.Put(c)
    }
}
```

## Testing and Validation

### Testing Live Reload

#### Unit Tests
```go
func TestFileWatcher(t *testing.T) {
    watcher := NewThemeWatcher()
    events := make(chan ThemeChange, 1)
    
    watcher.Subscribe(events)
    watcher.Start()
    
    // Modify theme file
    os.WriteFile(testThemePath, []byte("test"), 0644)
    
    // Verify event received
    select {
    case event := <-events:
        assert.Equal(t, testThemePath, event.Path)
    case <-time.After(100 * time.Millisecond):
        t.Fatal("No event received")
    }
}

func TestDebouncer(t *testing.T) {
    debouncer := NewDebouncer(50 * time.Millisecond)
    called := 0
    
    // Trigger multiple times
    for i := 0; i < 10; i++ {
        debouncer.Trigger(func() { called++ })
        time.Sleep(10 * time.Millisecond)
    }
    
    // Wait for debounce
    time.Sleep(100 * time.Millisecond)
    
    // Should only be called once
    assert.Equal(t, 1, called)
}
```

#### Integration Tests
```go
func TestGTKApplicationReload(t *testing.T) {
    // Start test GTK application
    app := startTestGTKApp()
    defer app.Stop()
    
    // Get initial theme
    initialTheme := app.GetTheme()
    
    // Trigger reload
    reloader := NewReloadManager()
    reloader.TriggerReload()
    
    // Wait for reload
    time.Sleep(100 * time.Millisecond)
    
    // Verify theme changed
    newTheme := app.GetTheme()
    assert.NotEqual(t, initialTheme, newTheme)
}
```

### Automated Testing Approaches

#### Test Matrix
```yaml
# .github/workflows/reload-tests.yml
strategy:
  matrix:
    gtk-version: [3.24, 4.0, 4.6]
    desktop: [gnome, kde, xfce, none]
    display: [x11, wayland]
    
steps:
  - name: Test Reload
    run: |
      ./test-reload.sh \
        --gtk ${{ matrix.gtk-version }} \
        --desktop ${{ matrix.desktop }} \
        --display ${{ matrix.display }}
```

#### Stress Testing
```go
func TestReloadStress(t *testing.T) {
    reloader := NewReloadManager()
    
    // Start multiple applications
    apps := make([]*TestApp, 100)
    for i := range apps {
        apps[i] = startTestApp()
        defer apps[i].Stop()
    }
    
    // Trigger rapid reloads
    start := time.Now()
    for i := 0; i < 1000; i++ {
        reloader.TriggerReload()
    }
    
    // Verify performance
    elapsed := time.Since(start)
    assert.Less(t, elapsed, 10*time.Second)
    
    // Verify all apps updated
    for _, app := range apps {
        assert.True(t, app.ThemeUpdated())
    }
}
```

### Performance Benchmarks

#### Latency Benchmarks
```go
func BenchmarkReloadLatency(b *testing.B) {
    reloader := NewReloadManager()
    app := startTestGTKApp()
    defer app.Stop()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        start := time.Now()
        reloader.TriggerReload()
        app.WaitForReload()
        b.ReportMetric(float64(time.Since(start).Microseconds()), "μs/reload")
    }
}

func BenchmarkBatchReload(b *testing.B) {
    reloader := NewReloadManager()
    apps := startTestApps(10)
    defer stopTestApps(apps)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        reloader.BatchReload(apps)
    }
}
```

#### Resource Usage
```go
func BenchmarkMemoryUsage(b *testing.B) {
    reloader := NewReloadManager()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        before := m.Alloc
        
        reloader.TriggerReload()
        
        runtime.ReadMemStats(&m)
        after := m.Alloc
        
        b.ReportMetric(float64(after-before), "bytes/reload")
    }
}
```

### User Experience Metrics

#### Perceived Performance
```go
type UXMetrics struct {
    reloadStart    time.Time
    firstPaint     time.Time
    fullyRendered  time.Time
}

func (m *UXMetrics) Measure() {
    m.reloadStart = time.Now()
    
    // Hook into rendering pipeline
    onFirstPaint(func() {
        m.firstPaint = time.Now()
    })
    
    onFullyRendered(func() {
        m.fullyRendered = time.Now()
        
        // Report metrics
        log.Printf("First paint: %v", m.firstPaint.Sub(m.reloadStart))
        log.Printf("Fully rendered: %v", m.fullyRendered.Sub(m.reloadStart))
    })
}
```

#### Success Rate Tracking
```go
type SuccessTracker struct {
    total     int64
    succeeded int64
    failed    int64
    mu        sync.Mutex
}

func (t *SuccessTracker) Record(success bool) {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    t.total++
    if success {
        t.succeeded++
    } else {
        t.failed++
    }
}

func (t *SuccessTracker) GetRate() float64 {
    t.mu.Lock()
    defer t.mu.Unlock()
    
    if t.total == 0 {
        return 0
    }
    return float64(t.succeeded) / float64(t.total)
}
```

## Risks and Mitigations

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Race conditions during reload | High | Medium | Mutex locks, atomic operations |
| Memory leaks in daemon | High | Low | Regular memory profiling |
| Compatibility breaks | High | Medium | Extensive testing matrix |
| Performance degradation | Medium | Medium | Continuous benchmarking |
| File system watch limits | Medium | Low | Dynamic watch management |
| D-Bus connection failures | Low | Medium | Retry logic, fallbacks |

### Implementation Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Complex integration points | High | High | Modular architecture |
| Desktop environment conflicts | Medium | Medium | Feature detection |
| Application-specific bugs | Low | High | Per-app workarounds |
| User confusion | Medium | Low | Clear documentation |

## Success Metrics

### Technical Metrics
- Detection latency < 10ms (p99)
- Reload completion < 50ms (p95)
- CPU usage < 1% idle
- Memory usage < 10MB
- Success rate > 99%
- Zero data corruption incidents

### User Experience Metrics
- Perceived instant updates (< 100ms)
- No visual glitches during reload
- Seamless theme transitions
- Works with 90% of GTK apps
- Positive user feedback > 4.5/5

### Adoption Metrics
- 50% of users enable live reload
- < 1% disable after trying
- Feature usage grows 10% monthly
- Community contributions increase
- Bug reports decrease over time

## Dev Log

### Session: Initial Planning
- Created comprehensive live reload system plan
- Defined detection and communication architecture
- Specified reload mechanisms for various environments
- Established desktop environment integration strategies
- Created application support matrix
- Defined performance optimization approaches
- Set up testing and validation framework
- Next: Begin implementation of file system monitoring