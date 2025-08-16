# GTK Asset Generation Plan

## Context

### Problem Statement
The current GTK theming implementation relies solely on CSS color definitions without providing themed graphical assets (checkboxes, radio buttons, switches, etc.). This limits the visual consistency and customization capabilities of GTK applications when using Heimdall themes. Native GTK widgets often fall back to system defaults or hardcoded assets that don't match the applied colorscheme.

### Current State
- GTK theming is handled via CSS generation in `internal/theme/gtk.go`
- Only color variables are defined, no custom assets
- No SVG generation or manipulation capabilities
- Assets default to system theme or GTK defaults
- No HiDPI support for custom assets

### Goals
- Generate complete set of GTK widget assets dynamically from colorschemes
- Support all standard GTK widgets with proper state variations
- Provide HiDPI-ready assets with @2x scaling
- Enable hot-reload for asset changes during development
- Maintain performance through intelligent caching
- Ensure cross-platform compatibility (Linux, BSD)

### Constraints
- Must work with both GTK3 and GTK4
- Assets must be generated in pure Go without external dependencies
- File sizes must be optimized for performance
- Must support both light and dark theme variants
- Generation must be fast enough for real-time updates

## Specification

### Functional Requirements

#### Asset Generation
- Generate SVG assets for all standard GTK widgets
- Support dynamic color injection from Heimdall colorschemes
- Create state variations (normal, hover, active, disabled, checked)
- Generate size variations (16x16, 24x24, 32x32, 48x48)
- Produce @2x versions for HiDPI displays
- Support RTL (right-to-left) layout variants

#### Widget Coverage
- Checkboxes (checked, unchecked, indeterminate)
- Radio buttons (selected, unselected)
- Switches (on, off, sliding states)
- Sliders (horizontal, vertical, marks)
- Progress bars (determinate, indeterminate)
- Spinners (animated frames)
- Arrows (up, down, left, right)
- Close, minimize, maximize buttons
- Expanders (collapsed, expanded)
- Handles (pane separators, resize grips)

#### Integration
- Seamless integration with existing theme system
- Automatic asset path resolution in CSS
- Fallback mechanism for missing assets
- Version management for cache invalidation
- Hot-reload support during development

### Non-Functional Requirements

#### Performance
- Asset generation < 100ms for complete set
- Incremental updates < 10ms per asset
- Memory usage < 50MB during generation
- Cached assets loaded instantly

#### Quality
- Pixel-perfect rendering at all sizes
- Smooth anti-aliasing for curves
- Consistent visual weight across assets
- Accessibility compliance (contrast ratios)

#### Maintainability
- Modular asset templates
- Clear separation of concerns
- Comprehensive test coverage
- Documentation for asset specifications

### Interfaces

#### Input Interface
```go
type AssetGenerationRequest struct {
    Colorscheme map[string]string
    Theme       string // "light" or "dark"
    Scale       int    // 1 or 2 for HiDPI
    Assets      []string // specific assets or "all"
}
```

#### Output Interface
```go
type GeneratedAssets struct {
    Assets    map[string][]byte // asset_name -> SVG content
    Manifest  AssetManifest
    Checksum  string
}

type AssetManifest struct {
    Version   string
    Generated time.Time
    Assets    []AssetEntry
}
```

## Implementation Plan

### Phase 1: SVG Template System

- [ ] Create base SVG template structures
  - Define viewBox standards for each asset type
  - Establish coordinate systems and scaling rules
  - Test SVG compatibility with GTK renderers
  
- [ ] Implement color injection mechanism
  - Parse SVG templates with placeholders
  - Replace color tokens with scheme colors
  - Handle opacity and gradient definitions
  - Test color accuracy across different displays

- [ ] Build path generation for complex shapes
  - Checkbox checkmark path algorithm
  - Radio button dot positioning
  - Switch track and thumb geometry
  - Test path rendering at different scales

### Phase 2: Asset Pipeline Architecture

- [ ] Design asset generation pipeline
  - Input validation and normalization
  - Template selection based on widget type
  - Color mapping and transformation
  - Test pipeline performance benchmarks

- [ ] Implement batch generation system
  - Parallel asset generation
  - Progress tracking and cancellation
  - Error handling and recovery
  - Test with large asset sets

- [ ] Create caching layer
  - Content-based cache keys
  - Filesystem cache with atomic writes
  - Memory cache for hot assets
  - Test cache hit rates and invalidation

### Phase 3: Core Widget Assets
- [ ] Generate checkbox assets
  - Unchecked states (normal, hover, active, disabled)
  - Checked states (normal, hover, active, disabled)
  - Indeterminate states
  - Test visual consistency across states

- [ ] Generate radio button assets
  - Unselected states (normal, hover, active, disabled)
  - Selected states (normal, hover, active, disabled)
  - Focus ring variations
  - Test circular rendering quality

- [ ] Generate switch assets
  - Off states (normal, hover, active, disabled)
  - On states (normal, hover, active, disabled)
  - Transition states for animation
  - Test switch track and thumb alignment

### Phase 4: Extended Widget Assets
- [ ] Generate slider assets
  - Horizontal track and thumb
  - Vertical track and thumb
  - Scale marks and labels
  - Test precise positioning

- [ ] Generate progress bar assets
  - Determinate progress fill
  - Indeterminate animation frames
  - Different orientations
  - Test smooth progression rendering

- [ ] Generate button assets
  - Window controls (close, minimize, maximize)
  - Navigation arrows
  - Dropdown indicators
  - Test icon clarity at small sizes

### Phase 5: HiDPI and Variants
- [ ] Implement @2x asset generation
  - Scale SVG viewBox appropriately
  - Adjust stroke widths for clarity
  - Optimize for retina displays
  - Test on various DPI settings

- [ ] Create theme variants
  - Light theme optimizations
  - Dark theme adjustments
  - High contrast versions
  - Test readability and accessibility

- [ ] Add RTL support
  - Mirror appropriate assets
  - Adjust directional indicators
  - Test with RTL languages
  - Validate layout correctness

### Phase 6: Integration and Testing
- [ ] Integrate with theme system
  - Update GTKHandler to include assets
  - Modify CSS generation for asset paths
  - Implement fallback mechanisms
  - Test with real GTK applications

- [ ] Add hot-reload support
  - File watcher for asset changes
  - Incremental regeneration
  - CSS cache invalidation
  - Test reload performance

- [ ] Comprehensive testing
  - Visual regression tests
  - Performance benchmarks
  - Cross-platform validation
  - Test with popular GTK apps

## Asset Requirements Analysis

### Complete Asset List

#### Checkbox Assets
- `checkbox-unchecked.svg` - Empty checkbox
- `checkbox-unchecked-hover.svg` - Hover state
- `checkbox-unchecked-active.svg` - Pressed state
- `checkbox-unchecked-disabled.svg` - Disabled empty
- `checkbox-checked.svg` - Checked checkbox
- `checkbox-checked-hover.svg` - Checked hover
- `checkbox-checked-active.svg` - Checked pressed
- `checkbox-checked-disabled.svg` - Disabled checked
- `checkbox-mixed.svg` - Indeterminate state
- `checkbox-mixed-hover.svg` - Indeterminate hover
- `checkbox-mixed-active.svg` - Indeterminate pressed
- `checkbox-mixed-disabled.svg` - Disabled indeterminate

#### Radio Button Assets
- `radio-unchecked.svg` - Empty radio
- `radio-unchecked-hover.svg` - Hover state
- `radio-unchecked-active.svg` - Pressed state
- `radio-unchecked-disabled.svg` - Disabled empty
- `radio-checked.svg` - Selected radio
- `radio-checked-hover.svg` - Selected hover
- `radio-checked-active.svg` - Selected pressed
- `radio-checked-disabled.svg` - Disabled selected

#### Switch Assets
- `switch-off.svg` - Switch in off position
- `switch-off-hover.svg` - Off hover state
- `switch-off-active.svg` - Off pressed state
- `switch-off-disabled.svg` - Disabled off
- `switch-on.svg` - Switch in on position
- `switch-on-hover.svg` - On hover state
- `switch-on-active.svg` - On pressed state
- `switch-on-disabled.svg` - Disabled on
- `switch-slider.svg` - Switch thumb/slider

#### Slider Assets
- `scale-slider-horizontal.svg` - Horizontal slider thumb
- `scale-slider-vertical.svg` - Vertical slider thumb
- `scale-trough-horizontal.svg` - Horizontal track
- `scale-trough-vertical.svg` - Vertical track
- `scale-highlight-horizontal.svg` - Active portion
- `scale-highlight-vertical.svg` - Active portion vertical

#### Progress Bar Assets
- `progressbar-trough.svg` - Progress bar background
- `progressbar-progress.svg` - Progress fill
- `progressbar-indeterminate-1.svg` - Animation frame 1
- `progressbar-indeterminate-2.svg` - Animation frame 2
- `progressbar-indeterminate-3.svg` - Animation frame 3
- `progressbar-indeterminate-4.svg` - Animation frame 4

#### Arrow Assets
- `arrow-up.svg` - Upward arrow
- `arrow-down.svg` - Downward arrow
- `arrow-left.svg` - Left arrow
- `arrow-right.svg` - Right arrow
- `arrow-up-small.svg` - Small variant
- `arrow-down-small.svg` - Small variant
- `arrow-left-small.svg` - Small variant
- `arrow-right-small.svg` - Small variant

#### Window Control Assets
- `window-close.svg` - Close button
- `window-close-hover.svg` - Close hover
- `window-close-active.svg` - Close pressed
- `window-minimize.svg` - Minimize button
- `window-minimize-hover.svg` - Minimize hover
- `window-minimize-active.svg` - Minimize pressed
- `window-maximize.svg` - Maximize button
- `window-maximize-hover.svg` - Maximize hover
- `window-maximize-active.svg` - Maximize pressed
- `window-restore.svg` - Restore button

#### Miscellaneous Assets
- `expander-collapsed.svg` - Collapsed expander
- `expander-expanded.svg` - Expanded expander
- `handle-horizontal.svg` - Horizontal resize handle
- `handle-vertical.svg` - Vertical resize handle
- `spinner-frame-1.svg` through `spinner-frame-8.svg` - Loading animation

### Size Requirements

**Base sizes**: 16x16, 24x24, 32x32, 48x48 pixels  
**HiDPI sizes**: 32x32, 48x48, 64x64, 96x96 pixels (@2x)  
**Stroke width**: 1px for normal, 2px for @2x  
**Padding**: 2px minimum from edges

### Color Mapping

#### Primary Colors

**Background**: Widget background color  
**Foreground**: Primary content color  
**Border**: Outline and border color  
**Accent**: Primary accent for selections  
**Hover**: Hover state highlight  
**Active**: Active/pressed state color  
**Disabled**: Reduced opacity or muted color

#### State-based Opacity

- Normal: 100% opacity
- Hover: 100% opacity with highlight
- Active: 100% opacity with darker shade
- Disabled: 40% opacity
- Focus: 100% opacity with focus ring

## SVG Generation System

### Template Structure
```xml
<!-- Base checkbox template -->
<svg viewBox="0 0 16 16" xmlns="http://www.w3.org/2000/svg">
  <rect x="2" y="2" width="12" height="12" 
        rx="2" ry="2"
        fill="{{background}}"
        stroke="{{border}}"
        stroke-width="1"/>
  {{#if checked}}
  <path d="M4 8 L7 11 L12 5"
        stroke="{{accent}}"
        stroke-width="2"
        fill="none"
        stroke-linecap="round"
        stroke-linejoin="round"/>
  {{/if}}
</svg>
```

### Color Injection Process
1. Parse template SVG
2. Extract color placeholders
3. Map colorscheme to template variables
4. Apply state-based modifications
5. Generate final SVG string
6. Optimize SVG output

### Dynamic Path Generation
```go
type PathGenerator interface {
    GenerateCheckmark(size int) string
    GenerateRadioDot(size int, selected bool) string
    GenerateSwitchThumb(position float64) string
    GenerateArrow(direction string, size int) string
}
```

### Optimization Strategies
- Remove unnecessary whitespace
- Combine similar paths
- Use CSS classes for repeated styles
- Minimize decimal precision
- Remove default attributes
- Use path shortcuts where possible

## Asset Pipeline Architecture

### Pipeline Stages
```
Input (Colorscheme) 
    ↓
Validation (Check required colors)
    ↓
Template Selection (Choose appropriate templates)
    ↓
Color Processing (Map and transform colors)
    ↓
SVG Generation (Create asset files)
    ↓
Optimization (Minimize file size)
    ↓
Caching (Store for reuse)
    ↓
Output (Asset files + manifest)
```

### Batch Generation
```go
type AssetGenerator struct {
    templates  map[string]*Template
    cache      *AssetCache
    workers    int
}

func (g *AssetGenerator) GenerateBatch(req AssetGenerationRequest) (*GeneratedAssets, error) {
    // Parallel generation with worker pool
    // Progress tracking
    // Error aggregation
}
```

### Caching Strategy
- **Cache Key**: Hash of (colorscheme + theme + scale + asset_type)
- **Cache Location**: `~/.cache/heimdall/gtk-assets/`
- **Cache Structure**:
  ```
  gtk-assets/
  ├── manifest.json
  ├── light/
  │   ├── 1x/
  │   │   └── [asset files]
  │   └── 2x/
  │       └── [asset files]
  └── dark/
      ├── 1x/
      │   └── [asset files]
      └── 2x/
          └── [asset files]
  ```

### Quality Assurance
- Validate SVG syntax
- Check color contrast ratios
- Verify asset dimensions
- Test rendering in GTK
- Compare with reference images

## Implementation Details

### Go Libraries
```go
// SVG manipulation
import (
    "github.com/ajstarks/svgo" // SVG generation
    "github.com/tdewolff/minify/v2" // SVG optimization
)

// Template engine
import (
    "text/template" // Built-in templating
    "github.com/Masterminds/sprig/v3" // Template functions
)

// Color manipulation
import (
    "github.com/lucasb-eyer/go-colorful" // Color operations
)
```

### File Organization
```
internal/theme/assets/
├── generator.go         # Main generator logic
├── templates.go         # SVG templates
├── optimizer.go         # SVG optimization
├── cache.go            # Caching logic
├── templates/
│   ├── checkbox.svg
│   ├── radio.svg
│   ├── switch.svg
│   └── ...
└── generated/          # Output directory
    ├── light/
    └── dark/
```

### Naming Conventions
- **Asset files**: `{widget}-{state}[-{variant}].svg`
- **Template files**: `{widget}-template.svg`
- **Cache files**: `{hash}-{widget}-{state}.svg`
- **Manifest**: `assets-manifest.json`

### Asset Structure in Theme Directory
```
~/.config/gtk-3.0/
├── gtk.css
└── assets/
    ├── checkbox-checked.svg
    ├── checkbox-unchecked.svg
    ├── radio-checked.svg
    └── ...

~/.config/gtk-4.0/
├── gtk.css
└── assets/
    └── [same structure]
```

## Rendering Variations

### Light Theme Adjustments
- Higher contrast borders
- Lighter backgrounds
- Darker foreground elements
- Subtle shadows for depth
- Brighter accent colors

### Dark Theme Adjustments
- Lower contrast borders
- Darker backgrounds
- Lighter foreground elements
- Glow effects for emphasis
- Muted accent colors

### High Contrast Mode
- Pure black/white elements
- Thicker borders (2px minimum)
- No gradients or shadows
- Maximum color separation
- Larger focus indicators

### Compact Sizing
- Reduced padding (1px)
- Smaller corner radius
- Thinner strokes
- Condensed spacing
- Optimized for density

### RTL Support
- Mirror horizontal arrows
- Flip progress directions
- Reverse slider orientations
- Adjust expander directions
- Maintain circular elements

## Integration with Theme System

### CSS Asset References
```css
/* GTK3 CSS */
checkbutton check {
    -gtk-icon-source: url("assets/checkbox-unchecked.svg");
}

checkbutton check:checked {
    -gtk-icon-source: url("assets/checkbox-checked.svg");
}

/* GTK4 CSS */
checkbutton > check {
    background-image: url("assets/checkbox-unchecked.svg");
}

checkbutton > check:checked {
    background-image: url("assets/checkbox-checked.svg");
}
```

### Fallback Mechanism
1. Check for custom asset in theme directory
2. Fall back to cached generated asset
3. Generate asset on-the-fly if missing
4. Use GTK default as last resort

### Hot Reload Implementation
```go
type AssetWatcher struct {
    watcher *fsnotify.Watcher
    generator *AssetGenerator
    reloadChan chan string
}

func (w *AssetWatcher) Watch() {
    // Monitor colorscheme changes
    // Trigger regeneration
    // Notify GTK to reload CSS
}
```

### Version Management
```json
{
  "version": "1.0.0",
  "colorscheme": "gruvbox",
  "generated": "2024-01-15T10:30:00Z",
  "checksum": "sha256:abc123...",
  "assets": [
    {
      "name": "checkbox-checked.svg",
      "size": 1024,
      "modified": "2024-01-15T10:30:00Z"
    }
  ]
}
```

## Testing Strategy

### Visual Regression Testing
```go
type VisualTest struct {
    Asset    string
    Expected string // Path to reference image
    Tolerance float64
}

func TestAssetGeneration(t *testing.T) {
    // Generate assets
    // Compare with references
    // Report differences
}
```

### Automated Validation
- SVG syntax validation
- Color accuracy tests
- Size verification
- State coverage checks
- Performance benchmarks

### Performance Benchmarks
```go
func BenchmarkAssetGeneration(b *testing.B) {
    // Measure generation time
    // Test memory usage
    // Profile CPU usage
    // Check cache performance
}
```

### Cross-Platform Testing
- Test on different GTK versions (3.24, 4.0+)
- Validate on various distributions
- Check Wayland vs X11 rendering
- Test with different DPI settings
- Verify with screen readers

### Integration Tests
- Test with GNOME applications
- Validate in KDE GTK applications  
- Check custom application themes
- Test with Flatpak applications
- Verify with Electron apps

## Risks and Mitigations

### Risk: SVG Rendering Inconsistencies

**Mitigation**:
- Use simple, well-supported SVG features
- Test across multiple GTK versions
- Provide PNG fallbacks if needed
- Document known limitations

### Risk: Performance Impact

**Mitigation**:
- Implement aggressive caching
- Use worker pools for generation
- Optimize SVG output size
- Lazy-load non-critical assets

### Risk: Color Mapping Complexity

**Mitigation**:
- Define clear color mapping rules
- Provide sensible defaults
- Allow user overrides
- Test with various colorschemes

### Risk: GTK Version Compatibility

**Mitigation**:
- Maintain separate GTK3/GTK4 paths
- Use feature detection
- Provide graceful degradation
- Document version requirements

## Success Metrics

### Performance Metrics
- Asset generation time < 100ms for full set
- Cache hit rate > 90% in normal usage
- Memory usage < 50MB during generation
- File size < 2KB per SVG asset

### Quality Metrics
- 100% widget coverage for standard GTK widgets
- Zero rendering artifacts at any scale
- WCAG AA compliance for contrast ratios
- Pixel-perfect alignment at all sizes

### User Experience Metrics
- Seamless theme switching < 1 second
- No visual glitches during transitions
- Consistent appearance across applications
- Positive user feedback on visual quality

### Technical Metrics
- Test coverage > 80%
- Zero critical bugs in production
- Documentation completeness 100%
- API stability with no breaking changes

## Dev Log

### Session: 2025-08-15 - Initial Planning
- Created comprehensive GTK asset generation plan
- Analyzed current GTK implementation in codebase
- Researched SVG generation approaches
- Defined complete asset requirements
- Established implementation phases
- Next steps: Begin Phase 1 implementation with SVG template system