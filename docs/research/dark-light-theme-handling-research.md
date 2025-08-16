# Dark and Light Theme Handling in Heimdall CLI

## Executive Summary

Heimdall CLI implements a comprehensive dark/light theme system that affects the entire desktop environment. The system uses a three-tier hierarchy (scheme/flavour/mode) where mode specifically determines whether a theme is dark or light. Mode detection from wallpapers uses luminance analysis, and the system integrates with QuickShell, GTK, Qt, and various applications with mode-aware color adjustments.

## 1. Scheme Structure and Mode Definition

### Hierarchical Organization
```
scheme/
├── name (e.g., "catppuccin", "rosepine")
│   ├── flavour (e.g., "mocha", "latte", "main")
│   │   ├── dark.json
│   │   └── light.json
```

### Mode Definition in Scheme Files
Each scheme file explicitly declares its mode:
```json
{
  "name": "catppuccin",
  "flavour": "mocha",
  "mode": "dark",  // Explicitly defined
  "colours": {
    "background": "#1e1e2e",
    "foreground": "#cdd6f4",
    // ... extensive color definitions
  }
}
```

### Color Key Differences Between Modes

**Dark Mode Characteristics:**
- Dark backgrounds (#1e1e2e for Catppuccin Mocha)
- Light foregrounds (#cdd6f4)
- Lower luminance values overall
- Higher contrast against light text

**Light Mode Characteristics:**
- Light backgrounds (#eff1f5 for Catppuccin Latte)
- Dark foregrounds (#4c4f69)
- Higher luminance values overall
- Higher contrast against dark text

## 2. Mode Detection from Wallpapers

### Luminance-Based Detection Algorithm
Located in `internal/utils/wallpaper/analyzer.go`:

```go
func (a *Analyzer) DetermineMode(path string) (string, error) {
    // Calculate average luminance using ITU-R BT.709 formula
    // luminance = 0.2126*r + 0.7152*g + 0.0722*b
    
    avgLuminance := luminanceSum / float64(pixelCount)
    
    // Counter-intuitive but correct logic:
    // Dark images (avg < 128) → light mode (for contrast)
    // Light images (avg >= 128) → dark mode (for contrast)
    if avgLuminance < 128 {
        return "light", nil
    }
    return "dark", nil
}
```

### Smart Mode Detection Flow
1. User sets wallpaper with smart mode enabled
2. Analyzer calculates average luminance
3. Mode determined based on luminance threshold
4. Material You scheme generated with detected mode
5. Theme automatically applied with appropriate mode

## 3. User Preferences and Defaults

### Preference Storage
User preferences stored in theme state (`~/.local/state/heimdall/theme-state.json`):

```go
type UserPreferences struct {
    AutoApplyGenerated bool   `json:"auto_apply_generated"`
    AutoApplyUser      bool   `json:"auto_apply_user"`
    AutoApplyBundled   bool   `json:"auto_apply_bundled"`
    PreferredVariant   string `json:"preferred_variant,omitempty"`
    PreferredMode      string `json:"preferred_mode,omitempty"`  // "dark" or "light"
    NotifyOnGeneration bool   `json:"notify_on_generation"`
}
```

### Default Behavior
- **Default Mode**: "dark" (fallback when not specified)
- **Default Scheme**: catppuccin/mocha/dark
- **Auto Mode**: Enabled by default in config
- **Smart Mode**: Enabled for wallpaper operations

### Setting Preferences
```bash
# Set preferred mode
heimdall scheme preferences --mode=dark

# Set scheme with explicit mode
heimdall scheme set rosepine main dark
heimdall scheme set catppuccin latte light

# Mode defaults to dark if not specified
heimdall scheme set rosepine main  # Uses dark mode
```

## 4. Application-Specific Mode Effects

### GTK Theme Generation (`internal/theme/gtk.go`)

Mode affects color calculations:
```go
func (h *GTKHandler) generateGTKCSS(colors map[string]string, mode string) string {
    // Header includes mode information
    builder.WriteString(fmt.Sprintf("/* Mode: %s */\n\n", mode))
    
    // Color adjustments based on mode
    // Dark mode: lighter containers
    builder.WriteString(fmt.Sprintf("@define-color primary_container %s;\n", 
        h.lighten(colors["background"], 0.1)))
    
    // Light mode would use darker containers
    builder.WriteString(fmt.Sprintf("@define-color surface %s;\n", 
        h.darken(colors["background"], 0.05)))
}
```

### Qt Theme Generation (`internal/theme/qt.go`)

Qt5ct/Qt6ct color arrays adjusted per mode:
```go
func (h *QtHandler) buildColorArray(colors map[string]string, state string) string {
    // Different color arrays for active/disabled/inactive states
    // Mode influences which colors are used for UI elements
    qtColors := []string{
        colors["foreground"],                  // WindowText
        h.lighten(colors["background"], 0.1),  // Button
        // ... 21 total colors
    }
}
```

### Terminal Applications

Terminal colors mapped differently based on mode:
- **Dark Mode**: Uses standard ANSI color mappings
- **Light Mode**: May swap certain colors for better visibility
- Color keys: `term0`-`term15`, `color0`-`color15`

### Discord Integration

Discord themes use mode-specific CSS:
- Dark mode: Dark backgrounds, light text
- Light mode: Light backgrounds, dark text
- Custom CSS generated based on mode

## 5. QuickShell Integration

### Triple-Write Strategy
When setting a scheme, Heimdall writes to three locations:

1. **Config Location**: `~/.config/heimdall/scheme.json`
2. **State Location**: `~/.local/state/heimdall/scheme.json`
3. **QuickShell Location**: `~/.local/state/quickshell/user/generated/scheme.json`

### QuickShell Format Preparation
```go
func (m *Manager) prepareQuickShellFormat(scheme *Scheme) map[string]interface{} {
    // Strip # prefix from colors for QuickShell
    colours := make(map[string]string)
    for key, value := range scheme.Colours {
        colours[key] = strings.TrimPrefix(value, "#")
    }
    
    return map[string]interface{}{
        "name":    scheme.Name,
        "flavour": scheme.Flavour,
        "mode":    scheme.Mode,  // Mode passed to QuickShell
        "variant": scheme.Variant,
        "colours": colours,
    }
}
```

### QuickShell Mode Handling
- QuickShell receives mode information in scheme.json
- Can adjust UI elements based on mode
- Launcher appearance changes with mode
- Window decorations adapt to mode

## 6. Material You Integration

### Mode-Aware Generation
```go
// From wallpaper.go
materialScheme, err := generator.GenerateScheme(palette.Seed, mode == "dark")
```

Material You generates different palettes based on mode:
- **Dark Mode**: Darker base colors, lighter accents
- **Light Mode**: Lighter base colors, darker accents
- Maintains Material Design 3 guidelines

### Variant Selection
Material You variants affected by mode:
- TonalSpot: Standard Material You palette
- Vibrant: Higher saturation
- Expressive: More colorful
- Content: Neutral, content-focused
- Each variant has dark/light versions

## 7. Mode Switching Workflow

### Manual Mode Switch
```bash
# Switch to light mode
heimdall scheme set catppuccin latte light

# Switch to dark mode  
heimdall scheme set catppuccin mocha dark
```

### Automatic Mode Switch (via wallpaper)
1. Set wallpaper with smart mode
2. Luminance analysis determines mode
3. Material You generates mode-appropriate scheme
4. Applications receive mode-specific colors
5. UI updates across the system

### Mode Persistence
- Current mode stored in state files
- Preserved across restarts
- Used as default for new operations
- Can be overridden by user preference

## 8. Color Adjustments per Mode

### Dark Mode Adjustments
```go
// Lighten for containers and surfaces
primary_container = lighten(background, 0.1)
secondary_container = lighten(background, 0.15)

// Darken for depth
surface = darken(background, 0.05)
```

### Light Mode Adjustments
```go
// Darken for containers and surfaces
primary_container = darken(background, 0.1)
secondary_container = darken(background, 0.15)

// Lighten for highlights
surface = lighten(background, 0.05)
```

### Contrast Considerations
- WCAG compliance for text readability
- Minimum contrast ratios maintained
- Automatic adjustments for accessibility
- Mode-specific outline colors

## 9. Configuration and Defaults

### Config File Settings
```json
{
  "scheme": {
    "default": "rosepine",
    "auto_mode": true,      // Enable automatic mode detection
    "material_you": true    // Enable Material You generation
  },
  "wallpaper": {
    "smart_mode": true      // Enable mode detection from wallpaper
  }
}
```

### Theme State Management
```json
{
  "current": {
    "name": "catppuccin",
    "flavour": "mocha",
    "mode": "dark",
    "variant": "tonalspot"
  },
  "preferences": {
    "preferred_mode": "dark",
    "auto_apply_generated": true
  }
}
```

## 10. Implementation Details

### Mode Validation
```go
// From set.go
if mode != "dark" && mode != "light" {
    return fmt.Errorf("invalid mode: %s (must be 'dark' or 'light')", mode)
}
```

### Mode Fallback Chain
1. Explicit user specification
2. User preference from state
3. Wallpaper detection (if smart mode)
4. Default to "dark"

### Mode in Scheme Loading
```go
func (m *Manager) LoadScheme(name, flavour, mode string) (*Scheme, error) {
    // Construct path: scheme/flavour/mode.json
    schemePath := filepath.Join(userPath, name, flavour, mode+".json")
    // Load and parse scheme with mode information
}
```

## Summary

Heimdall CLI's dark/light theme handling is deeply integrated throughout the system:

1. **Structure**: Mode is a fundamental part of the scheme hierarchy
2. **Detection**: Intelligent wallpaper analysis determines appropriate mode
3. **Integration**: All major desktop components respect mode settings
4. **Flexibility**: Users can manually override or use automatic detection
5. **Consistency**: Mode information propagates to all themed applications
6. **Persistence**: Mode preferences are saved and restored
7. **Adaptation**: Colors are adjusted per mode for optimal visibility

The system provides both automatic intelligence and manual control, ensuring users can have a cohesive dark or light theme experience across their entire desktop environment.

## Related Documents
- `docs/plans/gtk-theme-implementation-plan.md` - GTK theming details
- `docs/plans/scheme-sync-implementation-plan.md` - Scheme synchronization
- `docs/quickshell-scheme-requirements-analysis.md` - QuickShell integration
- `docs/research/colorscheme-design-best-practices-research.md` - Color theory

## Dev Log
### Session: 2025-08-15
- Analyzed dark/light mode implementation across codebase
- Documented mode detection algorithm and logic
- Mapped mode effects on various applications
- Detailed QuickShell integration requirements
- Explored Material You mode handling
- Documented color adjustments per mode