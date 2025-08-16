# Heimdall CLI Theming System Requirements Analysis

## Executive Summary

This document analyzes the user's requirements for improving the heimdall-cli theming system, focusing on three key areas: user-defined themes infrastructure, generated theme improvements, and theme selection workflow enhancements.

## Current System Analysis

### Existing Infrastructure

#### Theme Storage Locations
- **Bundled themes**: `assets/schemes/` (embedded in binary)
- **User data directory**: `~/.local/share/heimdall/schemes/` (primary user location)
- **Cache directory**: `~/.cache/heimdall/schemes/` (legacy location)
- **State directory**: `~/.local/state/heimdall/scheme.json` (current active scheme)

#### Theme Manager Capabilities
- Loads schemes from multiple sources (embedded assets and filesystem)
- Supports hierarchical structure: `scheme/flavour/mode.json`
- Maintains current scheme state with triple-write for QuickShell integration
- Handles both JSON and legacy TXT format schemes

#### Material You Integration
- Wallpaper-based color extraction using Material Design 3 algorithms
- Automatic mode detection (light/dark) based on wallpaper brightness
- Full 122-color scheme generation from wallpaper
- Immediate application when wallpaper changes (forced, not optional)

## Requirements Analysis

### 1. User-Defined Themes Infrastructure

#### Current Limitations
- No dedicated user themes directory in config path
- User themes mixed with generated/cached themes in data directory
- No configuration option for custom theme paths
- Limited discoverability of user-created themes

#### Required Improvements

##### A. Persistent User Themes Location
```
~/.config/heimdall/schemes/
├── custom-theme/
│   ├── default/
│   │   ├── dark.json
│   │   └── light.json
│   └── vibrant/
│       ├── dark.json
│       └── light.json
└── my-theme/
    └── main/
        └── dark.json
```

**Implementation Requirements:**
- Create new constant in `paths/xdg.go`: `UserSchemesDir = ~/.config/heimdall/schemes`
- Modify `Manager.ListSchemes()` to include user schemes directory
- Ensure user schemes take precedence over bundled themes with same name
- Support for theme inheritance/extension from bundled themes

##### B. Configuration Support
```json
{
  "schemes": {
    "userPaths": [
      "~/.config/heimdall/schemes",
      "~/custom-themes"
    ],
    "searchOrder": ["user", "bundled", "generated"],
    "allowOverrides": true
  }
}
```

**Key Features:**
- Multiple custom paths support
- Configurable search order
- Override control for name conflicts

##### C. Seamless Integration
- User themes appear in `heimdall scheme list` output
- Full support in `heimdall scheme get/set` commands
- Automatic discovery without manual registration
- Support for all existing scheme metadata (name, flavour, mode, variant)

### 2. Generated Theme Improvements

#### Current Issues

##### A. Color Extraction Problems
- **Issue**: Dark wallpapers producing light themes
- **Root Cause**: Mode detection based on average brightness, not dominant colors
- **Example**: Neon pink/blue on black background averages to "light"

##### B. Missing Key Colors
- **Issue**: Vibrant accent colors not captured
- **Root Cause**: Material You quantizer prioritizing volume over vibrancy
- **Example**: Small neon accents ignored in favor of large dark areas

##### C. Limited Variants
- **Issue**: Only single "wallpaper" variant generated
- **Root Cause**: No variant generation logic in current implementation

#### Required Improvements

##### A. Enhanced Color Extraction Algorithm
```go
type ExtractionStrategy struct {
    Mode           string   // "vibrant", "tonal", "expressive", "neutral"
    ColorPriority  string   // "volume", "vibrancy", "contrast"
    AccentBoost    float64  // 0.0-2.0 multiplier for accent colors
    DarkThreshold  float64  // 0.0-1.0 for mode detection
}
```

**Implementation:**
- Multi-pass color extraction (dominant + vibrant + accent)
- Weighted scoring system for color selection
- Separate accent color detection for small but vibrant regions
- Improved mode detection considering color distribution, not just average

##### B. Multiple Variant Generation
```go
type GeneratedScheme struct {
    Base      *Scheme  // Original extracted colors
    Variants  map[string]*Scheme {
        "vibrant":    // Boosted saturation, high contrast
        "tonal":      // Subtle, monochromatic
        "expressive": // Bold, complementary colors
        "neutral":    // Desaturated, professional
        "content":    // Balanced for readability
    }
    Modes     map[string]bool {
        "dark":  true,
        "light": true,
    }
}
```

**Features per Variant:**
- **Vibrant**: +30% saturation, expanded color range
- **Tonal**: Single hue variations, subtle gradients
- **Expressive**: Complementary color generation, artistic
- **Neutral**: -40% saturation, grayscale accents
- **Content**: Optimized contrast ratios for text

##### C. Persistent Generated Themes
```
~/.local/share/heimdall/schemes/
└── generated/
    └── wallpaper-[hash]/
        ├── vibrant/
        │   ├── dark.json
        │   └── light.json
        ├── tonal/
        │   ├── dark.json
        │   └── light.json
        └── metadata.json  # Source wallpaper, generation params
```

**Storage Strategy:**
- Hash-based naming from wallpaper path
- Metadata file linking to source wallpaper
- Automatic cleanup of old generated themes (configurable limit)
- Version tracking for regeneration on algorithm improvements

### 3. Theme Selection Workflow

#### Current Issues
- Generated theme immediately applied on wallpaper change
- No user choice in variant selection
- Cannot revert to previous theme after wallpaper change
- Generated themes not visible in scheme list

#### Required Improvements

##### A. Optional Theme Application
```go
type WallpaperChangeEvent struct {
    WallpaperPath   string
    GeneratedThemes []GeneratedScheme
    AutoApply       bool  // User configurable
    PromptUser      bool  // Show selection dialog
}
```

**Workflow Options:**
1. **Manual**: Generate but don't apply (default)
2. **Prompt**: Show picker with preview
3. **Auto**: Apply preferred variant automatically
4. **Smart**: Apply based on time of day/ambient light

##### B. Theme Tracking System
```go
type ThemeState struct {
    Current        *Scheme
    Previous       *Scheme
    Generated      *GeneratedScheme
    LastWallpaper  string
    UserSelected   bool  // true if user explicitly selected
}
```

**Features:**
- Track theme selection source (user vs auto)
- Maintain history for quick switching
- Preserve user choice across wallpaper changes
- Revert capability to previous theme

##### C. Enhanced Selection Interface
```bash
# List all themes including generated
$ heimdall scheme list
Bundled Themes:
  - catppuccin (4 flavours)
  - gruvbox (3 flavours)
  - onedark (1 flavour)

User Themes:
  - my-custom (2 flavours)

Generated Themes:
  - wallpaper-abc123 (5 variants, 2 modes) [CURRENT]
    Source: ~/Pictures/neon-city.jpg
    Generated: 2024-01-15 14:30

# Select specific generated variant
$ heimdall scheme set generated/wallpaper-abc123/vibrant/dark

# Preview without applying
$ heimdall scheme preview generated/wallpaper-abc123/tonal/light

# Show generation options
$ heimdall wallpaper generate --list-variants
Available variants for current wallpaper:
  ✓ vibrant/dark    - High contrast, saturated colors
  ✓ vibrant/light   - Bright, energetic palette
  ✓ tonal/dark     - Subtle, monochromatic
  ✓ tonal/light    - Soft, unified tones
  ✓ expressive/dark - Bold, artistic colors
```

## Implementation Priority

### Phase 1: Foundation (Week 1)
1. Implement user schemes directory in config path
2. Update scheme manager to search multiple locations
3. Add configuration schema for theme paths

### Phase 2: Generation Improvements (Week 2)
1. Enhance color extraction algorithm
2. Implement variant generation system
3. Add persistent storage for generated themes

### Phase 3: Workflow Enhancement (Week 3)
1. Decouple wallpaper change from theme application
2. Implement theme tracking system
3. Add selection UI improvements

### Phase 4: Polish & Testing (Week 4)
1. Add comprehensive tests
2. Update documentation
3. Migration guide for existing users

## Technical Considerations

### Performance
- Lazy loading of theme lists (cache directory scans)
- Parallel variant generation
- Incremental theme application

### Compatibility
- Maintain backward compatibility with existing schemes
- Support migration from old cache locations
- Preserve QuickShell integration

### User Experience
- Clear naming conventions for generated themes
- Visual indicators for theme sources
- Helpful error messages for conflicts

## Configuration Schema

```yaml
# ~/.config/heimdall/config.yaml
schemes:
  # User-defined theme directories
  user_paths:
    - ~/.config/heimdall/schemes
    - ~/Documents/themes
  
  # Search order for theme resolution
  search_order:
    - user      # User-defined themes
    - generated # Wallpaper-generated themes
    - bundled   # Built-in themes
  
  # Generated theme settings
  generation:
    # Variants to generate
    variants:
      - vibrant
      - tonal
      - expressive
    
    # Modes to generate
    modes:
      - dark
      - light
    
    # Auto-apply behavior
    auto_apply: false
    preferred_variant: vibrant
    
    # Storage settings
    max_generated: 10  # Maximum stored generated themes
    cleanup_days: 30   # Remove unused after N days
  
  # Selection behavior
  selection:
    show_generated: true
    show_source: true
    preview_enabled: true
```

## Success Metrics

1. **User Control**: Users can choose when/if to apply generated themes
2. **Theme Persistence**: User themes persist across updates
3. **Color Accuracy**: Generated themes capture wallpaper's key colors
4. **Variant Quality**: Each variant serves distinct use case
5. **Discoverability**: All theme sources easily accessible
6. **Performance**: No regression in theme switching speed

## Conclusion

The proposed improvements address three critical gaps in the current theming system:

1. **Infrastructure**: Dedicated user theme location with full integration
2. **Generation**: Accurate color extraction with multiple useful variants
3. **Workflow**: User control over theme application and selection

These changes will transform heimdall-cli from a tool that forces theme changes to one that empowers users with choice while maintaining the convenience of automatic theme generation.