# GTK Theme Implementation Plan

## Context

### Problem Statement
The current GTK theming implementation in Heimdall CLI lacks comprehensive widget coverage, proper asset generation, and live reload capabilities. This limits the ability to provide a complete and seamless theming experience across GTK3 and GTK4 applications.

### Current State
- Basic GTK theme generation exists but with limited widget coverage
- No asset generation system
- No live reload functionality
- Minimal desktop environment integration
- Limited application-specific theming

### Goals
- Complete GTK3/4 widget theming support
- Automated asset generation from color schemes
- Live reload system for development
- Full desktop environment integration
- Application-specific theme variants
- Performance-optimized theme generation

### Constraints
- Must maintain backward compatibility with existing themes
- Should work across different desktop environments
- Must handle both GTK3 and GTK4 simultaneously
- Performance impact should be minimal
- File system operations must be atomic

## Specification

### Functional Requirements
- Generate complete GTK3/4 themes from Heimdall color schemes
- Support all standard GTK widgets and states
- Generate SVG assets dynamically
- Provide live reload during development
- Integrate with desktop environment settings
- Support application-specific overrides

### Non-Functional Requirements
- Theme generation < 100ms
- Live reload latency < 50ms
- Memory usage < 50MB during generation
- Support for 1000+ simultaneous theme applications
- Zero data loss during theme switching

### Interfaces
- CLI commands for theme management
- IPC interface for live reload
- D-Bus integration for desktop environments
- File system watchers for change detection

## Implementation Plan

### Phase 1: Core Infrastructure

**Timeline**: 2 weeks  
**Complexity**: High

#### Objectives
- Establish robust theme generation architecture
- Create modular widget theming system
- Implement basic GTK3/4 compatibility layer

#### Tasks
- [ ] Create theme generator architecture
  - Acceptance: Modular, extensible generator system
  - Test: Unit tests for each generator component
  
- [ ] Implement color mapping system
  - Acceptance: Maps Heimdall colors to GTK color names
  - Test: Validates all color mappings
  
- [ ] Build GTK version compatibility layer
  - Acceptance: Handles GTK3/4 differences transparently
  - Test: Version-specific output validation
  
- [ ] Create theme file structure manager
  - Acceptance: Creates/manages theme directory structure
  - Test: File system operations are atomic
  
- [ ] Implement basic CSS generation
  - Acceptance: Generates valid GTK CSS
  - Test: CSS validation and parsing tests

#### Dependencies
- Existing Heimdall color scheme system
- File system access permissions
- GTK development documentation

### Phase 2: Widget Coverage

**Timeline**: 3 weeks  
**Complexity**: High

#### Objectives
- Complete widget theming for GTK3/4
- Handle all widget states and variants
- Ensure visual consistency

#### Tasks
- [ ] Implement base widget styles
  - Acceptance: All basic widgets themed
  - Test: Visual regression tests
  
- [ ] Add container widget theming
  - Acceptance: Notebooks, panes, frames styled
  - Test: Layout consistency tests
  
- [ ] Create input widget styles
  - Acceptance: Entries, buttons, toggles themed
  - Test: Interaction state tests
  
- [ ] Implement complex widget theming
  - Acceptance: Trees, lists, menus complete
  - Test: Performance with large datasets
  
- [ ] Add state-specific styling
  - Acceptance: Hover, active, disabled states
  - Test: State transition smoothness
  
- [ ] Create focus and selection styles
  - Acceptance: Clear focus indicators
  - Test: Accessibility compliance

#### Dependencies
- Phase 1 completion
- GTK widget documentation
- Accessibility guidelines

### Phase 3: Asset Generation
**Timeline: 2 weeks**
**Complexity: Medium**

#### Objectives
- Generate SVG assets from color schemes
- Create icon variants
- Optimize asset delivery

#### Tasks
- [ ] Build SVG template system
  - Acceptance: Parameterized SVG templates
  - Test: SVG validity and rendering
  
- [ ] Implement asset generator
  - Acceptance: Creates all required assets
  - Test: Asset quality validation
  
- [ ] Create icon colorization system
  - Acceptance: Recolors icons dynamically
  - Test: Color accuracy tests
  
- [ ] Add asset caching layer
  - Acceptance: Caches generated assets
  - Test: Cache invalidation logic
  
- [ ] Implement asset optimization
  - Acceptance: Minimized file sizes
  - Test: Performance benchmarks

#### Dependencies
- SVG manipulation library
- Image processing capabilities
- Phase 2 completion

### Phase 4: Live Reload System
**Timeline: 2 weeks**
**Complexity: Medium**

#### Objectives
- Enable instant theme updates
- Provide development mode
- Minimize application disruption

#### Tasks
- [ ] Create file system watcher
  - Acceptance: Detects theme changes
  - Test: Change detection accuracy
  
- [ ] Implement IPC communication
  - Acceptance: Notifies applications of changes
  - Test: Message delivery reliability
  
- [ ] Build reload orchestrator
  - Acceptance: Coordinates reload process
  - Test: Concurrent reload handling
  
- [ ] Add incremental update system
  - Acceptance: Updates only changed parts
  - Test: Partial update correctness
  
- [ ] Create development mode CLI
  - Acceptance: Easy-to-use dev commands
  - Test: CLI integration tests

#### Dependencies
- IPC mechanism (D-Bus or custom)
- File system monitoring capability
- Phase 3 completion

### Phase 5: Desktop Integration
**Timeline: 2 weeks**
**Complexity: High**

#### Objectives
- Integrate with desktop environments
- Support system theme settings
- Handle environment-specific features

#### Tasks
- [ ] Implement GNOME integration
  - Acceptance: Works with GNOME settings
  - Test: GNOME-specific features work
  
- [ ] Add KDE Plasma support
  - Acceptance: Integrates with KDE settings
  - Test: KDE color scheme sync
  
- [ ] Create XFCE compatibility
  - Acceptance: XFCE theme management
  - Test: XFCE-specific testing
  
- [ ] Build settings daemon interface
  - Acceptance: Responds to system changes
  - Test: Event handling reliability
  
- [ ] Add display server integration
  - Acceptance: Works with X11 and Wayland
  - Test: Display server compatibility

#### Dependencies
- Desktop environment APIs
- D-Bus access
- Phase 4 completion

### Phase 6: Application-Specific Themes
**Timeline: 1 week**
**Complexity: Low**

#### Objectives
- Support per-application theming
- Create application profiles
- Enable fine-tuned customization

#### Tasks
- [ ] Build application detection system
  - Acceptance: Identifies running applications
  - Test: Detection accuracy
  
- [ ] Create override system
  - Acceptance: Per-app theme overrides
  - Test: Override precedence
  
- [ ] Implement profile manager
  - Acceptance: Saves/loads app profiles
  - Test: Profile persistence
  
- [ ] Add conditional theming
  - Acceptance: Context-aware themes
  - Test: Condition evaluation
  
- [ ] Create app-specific templates
  - Acceptance: Custom app templates
  - Test: Template rendering

#### Dependencies
- Application detection mechanism
- Phase 5 completion

## Development Tasks

### Core Components

#### Theme Generator (`internal/theme/gtk_generator.go`)
```go
type GTKThemeGenerator struct {
    scheme *ColorScheme
    version GTKVersion
    options GeneratorOptions
}

// API Design
Generate() (*Theme, error)
GenerateCSS() (string, error)
GenerateAssets() ([]Asset, error)
```

#### Widget Styler (`internal/theme/gtk_widgets.go`)
```go
type WidgetStyler struct {
    colors ColorMapper
    states StateManager
}

// API Design
StyleWidget(widget Widget) CSS
GetStateStyles(state State) CSS
ApplyModifiers(base CSS, mods []Modifier) CSS
```

#### Asset Generator (`internal/theme/gtk_assets.go`)
```go
type AssetGenerator struct {
    templates []SVGTemplate
    cache Cache
}

// API Design
GenerateAsset(name string, colors ColorScheme) ([]byte, error)
GenerateIconSet(colors ColorScheme) (map[string][]byte, error)
OptimizeSVG(svg []byte) ([]byte, error)
```

#### Live Reload Manager (`internal/theme/gtk_reload.go`)
```go
type ReloadManager struct {
    watcher FileWatcher
    notifier IPCNotifier
}

// API Design
Start() error
Stop() error
RegisterApplication(app Application) error
TriggerReload(changes []Change) error
```

### File Structure
```
internal/theme/
├── gtk/
│   ├── generator.go      # Main generator
│   ├── widgets.go         # Widget styling
│   ├── assets.go          # Asset generation
│   ├── reload.go          # Live reload
│   ├── integration.go     # Desktop integration
│   ├── templates/         # CSS templates
│   │   ├── gtk3/
│   │   └── gtk4/
│   └── assets/           # SVG templates
│       ├── checkboxes/
│       ├── radio/
│       └── arrows/
```

### New CLI Commands
```bash
heimdall scheme gtk generate    # Generate GTK theme
heimdall scheme gtk watch       # Live reload mode
heimdall scheme gtk validate    # Validate theme
heimdall scheme gtk preview     # Preview in test window
heimdall scheme gtk profile     # Manage app profiles
```

## Migration Strategy

### Transition Approach
1. **Parallel Implementation**: New system runs alongside old
2. **Feature Flag Control**: Gradual rollout with flags
3. **Compatibility Layer**: Translates old configs to new
4. **Deprecation Warnings**: Notify users of changes
5. **Migration Tool**: Automated config migration

### Backward Compatibility
- Maintain old theme format support for 6 months
- Provide conversion utilities
- Keep legacy CLI commands with deprecation notices
- Support old configuration keys with warnings

### User Migration Path
1. **Phase 1**: Optional opt-in to new system
2. **Phase 2**: New system default, old available
3. **Phase 3**: Old system deprecated
4. **Phase 4**: Old system removed

### Data Migration
```bash
# Migration command
heimdall migrate gtk-themes

# What it does:
- Backs up existing themes
- Converts to new format
- Validates conversion
- Updates configuration
```

## Testing Strategy

### Unit Tests
```go
// Generator tests
TestColorMapping()
TestWidgetGeneration()
TestCSSValidity()
TestAssetGeneration()

// Coverage target: 90%
```

### Integration Tests
```go
// End-to-end tests
TestFullThemeGeneration()
TestLiveReload()
TestDesktopIntegration()
TestApplicationTheming()

// Coverage target: 80%
```

### Visual Regression Testing
- Screenshot comparison for each widget
- Automated visual diff detection
- Test matrix: GTK versions × color schemes × widgets
- CI/CD integration for PR validation

### Performance Benchmarks
```go
// Benchmark targets
BenchmarkThemeGeneration()  // < 100ms
BenchmarkAssetGeneration()  // < 50ms
BenchmarkLiveReload()       // < 50ms
BenchmarkMemoryUsage()      // < 50MB
```

### Test Infrastructure
```yaml
# .github/workflows/gtk-tests.yml
- Unit tests on every commit
- Integration tests on PR
- Visual regression on merge
- Performance tests weekly
```

## Rollout Plan

### Feature Flags
```json
{
  "gtk_new_generator": false,
  "gtk_live_reload": false,
  "gtk_asset_generation": false,
  "gtk_desktop_integration": false,
  "gtk_app_profiles": false
}
```

### Gradual Enablement
1. **Week 1-2**: Internal testing, dogfooding
2. **Week 3-4**: Beta users (opt-in)
3. **Week 5-6**: 10% rollout
4. **Week 7-8**: 50% rollout
5. **Week 9-10**: 100% rollout

### Documentation Requirements
- [ ] User guide for new GTK theming
- [ ] Migration guide from old system
- [ ] API documentation for developers
- [ ] Troubleshooting guide
- [ ] Video tutorials for common tasks

### User Communication
- [ ] Blog post announcing new features
- [ ] Discord announcement and Q&A
- [ ] GitHub release notes
- [ ] Email to registered users
- [ ] Social media updates

### Monitoring and Metrics
- Generation success rate
- Performance metrics (p50, p95, p99)
- User adoption rate
- Bug report frequency
- User satisfaction surveys

## Risks and Mitigations

### Technical Risks
| Risk | Impact | Mitigation |
|------|--------|------------|
| GTK version incompatibility | High | Extensive testing matrix |
| Performance degradation | Medium | Continuous benchmarking |
| Asset generation failures | Low | Fallback to defaults |
| Live reload instability | Medium | Feature flag control |

### User Experience Risks
| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking changes | High | Compatibility layer |
| Learning curve | Medium | Comprehensive docs |
| Migration issues | Medium | Automated tooling |
| Feature parity | Low | Phased approach |

## Success Metrics

### Technical Metrics
- Theme generation time < 100ms (p95)
- Live reload latency < 50ms (p95)
- Test coverage > 85%
- Zero critical bugs in production
- 99.9% generation success rate

### User Metrics
- 80% adoption within 3 months
- < 5% rollback rate
- User satisfaction > 4.5/5
- Support ticket reduction by 30%
- Active usage increase by 50%

### Business Metrics
- Increased user retention
- Positive community feedback
- Contributor engagement increase
- Feature request completion

## Dev Log

### Session: Initial Planning
- Created comprehensive implementation plan
- Defined six implementation phases
- Established testing and rollout strategies
- Set clear success metrics
- Next: Begin Phase 1 implementation