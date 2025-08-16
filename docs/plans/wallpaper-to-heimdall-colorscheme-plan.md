# Wallpaper to Heimdall Colorscheme Bridge Plan

## Executive Summary

This plan outlines the complete implementation strategy to extend the current wallpaper Material You generator (29 colors) to produce a full Heimdall-compliant colorscheme (122 colors) that seamlessly integrates with QuickShell and all supported applications.

## Current State Analysis

### Existing Implementation
- **Location**: `internal/commands/wallpaper/wallpaper.go`
- **Current Output**: 29 Material Design colors via `convertMaterialColors()`
- **Integration**: Writes to QuickShell at `~/.local/state/quickshell/user/generated/scheme.json`
- **Limitations**:
  - Only maps ~20 Material Design colors
  - Missing 100+ Heimdall-required keys
  - No ANSI terminal colors (term0-15)
  - Incomplete surface hierarchy
  - Missing semantic colors (success, warning)
  - No theme-specific colors (Catppuccin names)

### Required Output
- **Total Keys**: 122 color definitions
- **Categories**:
  - Material Design 3 tokens (60+ colors)
  - ANSI terminal colors (16 colors)
  - Semantic colors (error, success, warning)
  - Surface hierarchy (12 levels)
  - Theme-specific extensions (14 Catppuccin colors)

## Implementation Strategy

### Phase 1: Extend Color Generation Engine

#### 1.1 Create Comprehensive Color Generator
**File**: `internal/scheme/generator/wallpaper_generator.go`

```go
package generator

import (
    "github.com/arthur404dev/heimdall-cli/internal/utils/material"
    "github.com/arthur404dev/heimdall-cli/internal/utils/color"
)

type WallpaperGenerator struct {
    materialGen *material.Generator
}

func (g *WallpaperGenerator) GenerateFullScheme(
    materialScheme *material.Scheme,
    wallpaperPath string,
    mode string,
) (*scheme.Scheme, error) {
    // Generate all 122 colors from Material You base
}
```

#### 1.2 Implement Color Inference Algorithms
Based on the blueprint, implement:
- HSL manipulation functions
- Luminance calculations
- Contrast enforcement
- Tonal palette generation
- Surface hierarchy creation

#### 1.3 ANSI Color Mapping
Generate semantically correct terminal colors:
```go
func generateANSIColors(materialScheme *material.Scheme) map[string]string {
    ansi := make(map[string]string)
    
    // Map Material colors to ANSI semantics
    ansi["term0"] = darken(materialScheme.Background, 0.2)  // Black
    ansi["term1"] = materialScheme.Error                     // Red
    ansi["term2"] = generateGreen(materialScheme.Primary)    // Green
    ansi["term3"] = generateYellow(materialScheme.Primary)   // Yellow
    ansi["term4"] = materialScheme.Primary                   // Blue
    ansi["term5"] = materialScheme.Secondary                 // Magenta
    ansi["term6"] = materialScheme.Tertiary                  // Cyan
    ansi["term7"] = lighten(materialScheme.OnBackground, 0.1) // White
    
    // Generate bright variants
    for i := 0; i < 8; i++ {
        ansi[fmt.Sprintf("term%d", i+8)] = brighten(ansi[fmt.Sprintf("term%d", i)])
    }
    
    return ansi
}
```

### Phase 2: Modify Wallpaper Command

#### 2.1 Update convertMaterialColors Function
**File**: `internal/commands/wallpaper/wallpaper.go`

```go
func convertMaterialColors(ms *material.Scheme) map[string]string {
    generator := NewWallpaperGenerator()
    fullScheme := generator.GenerateFullScheme(ms, wallpaperPath, mode)
    return fullScheme.Colours
}
```

#### 2.2 Enhance generateMaterialYouScheme
```go
func generateMaterialYouScheme(wallpaperPath string) error {
    // ... existing image processing ...
    
    // Generate full Heimdall scheme
    generator := scheme.NewWallpaperGenerator()
    newScheme, err := generator.GenerateFromWallpaper(
        wallpaperPath,
        materialScheme,
        mode,
        variant,
    )
    
    // Validate all 122 colors present
    if err := validateHeimdallScheme(newScheme); err != nil {
        return fmt.Errorf("incomplete scheme generation: %w", err)
    }
    
    // Save and apply
    manager := scheme.NewManager()
    if err := manager.SaveScheme(newScheme); err != nil {
        return err
    }
    
    return manager.SetScheme(newScheme)
}
```

### Phase 3: QuickShell Integration Enhancement

#### 3.1 Update prepareQuickShellFormat
**File**: `internal/scheme/manager.go`

```go
func (m *Manager) prepareQuickShellFormat(scheme *Scheme) map[string]interface{} {
    // Ensure all 122 colors are included
    colours := make(map[string]string)
    
    // Strip # prefix for QuickShell
    for key, value := range scheme.Colours {
        colours[key] = strings.TrimPrefix(value, "#")
    }
    
    // Add display metadata for QuickShell UI
    return map[string]interface{}{
        "name":    scheme.Name,
        "flavour": scheme.Flavour,
        "mode":    scheme.Mode,
        "variant": scheme.Variant,
        "source":  "wallpaper", // Indicate dynamic generation
        "colours": colours,
        "metadata": map[string]interface{}{
            "generated": true,
            "wallpaper": getCurrentWallpaperPath(),
            "timestamp": time.Now().Unix(),
        },
    }
}
```

#### 3.2 Ensure QuickShell Displays "Wallpaper" Option
The scheme should be saved with identifiable metadata:
```json
{
    "name": "material-you",
    "flavour": "wallpaper",
    "mode": "dark",
    "variant": "dynamic",
    "source": "wallpaper"
}
```

### Phase 4: Color Generation Algorithms

#### 4.1 Surface Hierarchy Generation
```go
func generateSurfaceHierarchy(background string, isDark bool) map[string]string {
    surfaces := make(map[string]string)
    
    if isDark {
        // Dark mode: progressive lightening
        surfaces["surfaceContainerLowest"] = adjustLightness(background, -2)
        surfaces["surfaceContainerLow"] = adjustLightness(background, 3)
        surfaces["surfaceContainer"] = adjustLightness(background, 5)
        surfaces["surfaceContainerHigh"] = adjustLightness(background, 8)
        surfaces["surfaceContainerHighest"] = adjustLightness(background, 12)
    } else {
        // Light mode: subtle variations
        surfaces["surfaceContainerLowest"] = adjustLightness(background, -3)
        surfaces["surfaceContainerLow"] = background
        surfaces["surfaceContainer"] = background
        surfaces["surfaceContainerHigh"] = adjustLightness(background, 1)
        surfaces["surfaceContainerHighest"] = adjustLightness(background, 2)
    }
    
    return surfaces
}
```

#### 4.2 Semantic Color Generation
```go
func generateSemanticColors(primary, background string) map[string]string {
    semantic := make(map[string]string)
    
    // Success (green-shifted from primary)
    primaryHSL := hexToHSL(primary)
    successHSL := HSL{
        H: 120, // Green hue
        S: primaryHSL.S,
        L: primaryHSL.L,
    }
    semantic["success"] = hslToHex(successHSL)
    semantic["onSuccess"] = ensureContrast(semantic["success"], background, 4.5)
    semantic["successContainer"] = mixColors(semantic["success"], background, 0.7)
    semantic["onSuccessContainer"] = ensureContrast(semantic["successContainer"], semantic["success"], 3.0)
    
    // Warning (yellow-shifted)
    warningHSL := HSL{
        H: 60, // Yellow hue
        S: primaryHSL.S,
        L: primaryHSL.L,
    }
    semantic["warning"] = hslToHex(warningHSL)
    // ... similar for warning colors
    
    return semantic
}
```

#### 4.3 Theme-Specific Colors (Catppuccin Compatibility)
```go
func generateThemeSpecificColors(scheme *ColorScheme) map[string]string {
    colors := make(map[string]string)
    
    // Generate Catppuccin-style named colors from Material palette
    colors["rosewater"] = lighten(scheme.Tertiary, 0.2)
    colors["flamingo"] = mixColors(scheme.Tertiary, scheme.Secondary, 0.5)
    colors["pink"] = scheme.Tertiary
    colors["mauve"] = mixColors(scheme.Primary, scheme.Tertiary, 0.5)
    colors["red"] = scheme.Error
    colors["maroon"] = darken(scheme.Error, 0.1)
    colors["peach"] = generateOrange(scheme.Primary, scheme.Error)
    colors["yellow"] = scheme.Term3
    colors["green"] = scheme.Term2
    colors["teal"] = scheme.Term6
    colors["sky"] = lighten(scheme.Term6, 0.1)
    colors["sapphire"] = mixColors(scheme.Term4, scheme.Term6, 0.5)
    colors["blue"] = scheme.Primary
    colors["lavender"] = lighten(scheme.Primary, 0.1)
    
    return colors
}
```

### Phase 5: Application Consistency

#### 5.1 Terminal Applier Update
Ensure terminal colors are properly applied:
```go
func applyTerminalColors(scheme *scheme.Scheme) error {
    // Generate escape sequences for all 16 ANSI colors
    for i := 0; i < 16; i++ {
        colorKey := fmt.Sprintf("term%d", i)
        if color, ok := scheme.Colours[colorKey]; ok {
            sequence := generateANSISequence(i, color)
            // Apply to terminal
        }
    }
}
```

#### 5.2 Application Template Updates
Ensure all application templates can handle the new color keys:
- Kitty: Map term0-15 to color0-15
- Neovim: Use Material Design tokens for UI
- GTK: Apply surface hierarchy
- Discord: Use semantic colors

### Phase 6: Testing Strategy

#### 6.1 Unit Tests
```go
func TestWallpaperSchemeGeneration(t *testing.T) {
    tests := []struct {
        name           string
        wallpaperPath  string
        expectedColors int
        requiredKeys   []string
    }{
        {
            name:           "Dark wallpaper generates full scheme",
            wallpaperPath:  "testdata/dark_wallpaper.jpg",
            expectedColors: 122,
            requiredKeys: []string{
                "background", "foreground", "primary", "secondary",
                "term0", "term15", "surface", "error", "success",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            scheme := generateFromWallpaper(tt.wallpaperPath)
            assert.Equal(t, tt.expectedColors, len(scheme.Colours))
            
            for _, key := range tt.requiredKeys {
                assert.Contains(t, scheme.Colours, key)
            }
        })
    }
}
```

#### 6.2 Contrast Validation Tests
```go
func TestContrastRequirements(t *testing.T) {
    scheme := generateTestScheme()
    
    // WCAG AAA for main text
    assertContrast(t, scheme.Background, scheme.Foreground, 7.0)
    
    // WCAG AA for UI elements
    assertContrast(t, scheme.Primary, scheme.OnPrimary, 4.5)
    assertContrast(t, scheme.Surface, scheme.OnSurface, 4.5)
    
    // Container contrasts
    assertContrast(t, scheme.PrimaryContainer, scheme.OnPrimaryContainer, 3.0)
}
```

#### 6.3 Integration Tests
- Test wallpaper command generates complete scheme
- Verify QuickShell receives all colors
- Confirm all applications receive consistent theming
- Test mode detection (light/dark) accuracy

### Phase 7: Migration Path

#### 7.1 Backward Compatibility
- Maintain existing scheme format
- Support legacy color names
- Preserve user's custom schemes

#### 7.2 User Migration
```bash
# Regenerate scheme from current wallpaper
heimdall wallpaper --regenerate

# Or set new wallpaper with full generation
heimdall wallpaper -f ~/Pictures/wallpaper.jpg
```

#### 7.3 Configuration Updates
Add configuration for color generation:
```json
{
  "wallpaper": {
    "generateFullScheme": true,
    "colorInference": {
      "ansiMapping": "semantic",
      "contrastEnforcement": true,
      "minContrast": {
        "background": 7.0,
        "surface": 4.5,
        "container": 3.0
      }
    }
  }
}
```

## Implementation Timeline

### Week 1: Core Infrastructure
- [ ] Implement color generation algorithms
- [ ] Create WallpaperGenerator class
- [ ] Add HSL/LAB color space conversions
- [ ] Implement contrast calculations

### Week 2: Integration
- [ ] Update wallpaper command
- [ ] Modify scheme manager
- [ ] Enhance QuickShell format
- [ ] Add validation functions

### Week 3: Testing & Refinement
- [ ] Write comprehensive tests
- [ ] Test with various wallpapers
- [ ] Validate contrast ratios
- [ ] Ensure application consistency

### Week 4: Documentation & Release
- [ ] Update user documentation
- [ ] Create migration guide
- [ ] Add example configurations
- [ ] Release and monitor feedback

## Success Criteria

1. **Complete Color Generation**
   - All 122 Heimdall color keys generated
   - Proper contrast ratios maintained
   - Semantic correctness preserved

2. **QuickShell Integration**
   - "Wallpaper" option appears in UI
   - Dynamic updates work seamlessly
   - All colors properly formatted (no # prefix)

3. **Application Consistency**
   - Kitty terminal shows correct colors
   - Neovim theme applies properly
   - GTK/Qt applications themed consistently
   - Discord receives all color values

4. **User Experience**
   - Single command generates complete theme
   - Fast generation (<1 second)
   - Predictable and pleasant results
   - Smooth migration from existing setup

## Risk Mitigation

### Risk 1: Color Generation Quality
**Mitigation**: Implement multiple generation algorithms and allow user selection

### Risk 2: Performance Impact
**Mitigation**: Cache generated schemes, optimize color calculations

### Risk 3: Breaking Changes
**Mitigation**: Maintain backward compatibility, provide migration tools

### Risk 4: QuickShell Compatibility
**Mitigation**: Test extensively with QuickShell, maintain format consistency

## Code Examples

### Complete Generation Function
```go
func GenerateHeimdallFromWallpaper(wallpaperPath string) (*scheme.Scheme, error) {
    // 1. Extract colors from wallpaper
    img, err := loadImage(wallpaperPath)
    if err != nil {
        return nil, err
    }
    
    // 2. Generate Material You palette
    generator := material.NewGenerator()
    palette, err := generator.GenerateFromImage(img)
    if err != nil {
        return nil, err
    }
    
    // 3. Determine mode
    analyzer := wallpaper.NewAnalyzer()
    mode, _ := analyzer.DetermineMode(wallpaperPath)
    
    // 4. Generate Material scheme
    materialScheme, err := generator.GenerateScheme(palette.Seed, mode == "dark")
    if err != nil {
        return nil, err
    }
    
    // 5. Expand to full Heimdall scheme
    heimdallScheme := &scheme.Scheme{
        Name:    "material-you",
        Flavour: "wallpaper",
        Mode:    mode,
        Variant: "dynamic",
        Colours: make(map[string]string),
    }
    
    // 6. Generate all color categories
    heimdallScheme.Colours = mergeColorMaps(
        convertMaterialColors(materialScheme),
        generateANSIColors(materialScheme),
        generateSurfaceHierarchy(materialScheme.Background, mode == "dark"),
        generateSemanticColors(materialScheme.Primary, materialScheme.Background),
        generateThemeSpecificColors(materialScheme),
    )
    
    // 7. Validate completeness
    if len(heimdallScheme.Colours) < 122 {
        return nil, fmt.Errorf("incomplete scheme: only %d colors generated", len(heimdallScheme.Colours))
    }
    
    return heimdallScheme, nil
}
```

### QuickShell Update Function
```go
func updateQuickShellWithWallpaperScheme(scheme *scheme.Scheme, wallpaperPath string) error {
    // Prepare QuickShell format
    quickshellData := map[string]interface{}{
        "name":    "material-you",
        "flavour": "wallpaper",
        "mode":    scheme.Mode,
        "variant": "dynamic",
        "colours": stripHashPrefixes(scheme.Colours),
        "metadata": map[string]interface{}{
            "source":    "wallpaper",
            "wallpaper": wallpaperPath,
            "generated": time.Now().Unix(),
        },
    }
    
    // Write to QuickShell location
    quickshellPath := filepath.Join(
        os.Getenv("HOME"),
        ".local/state/quickshell/user/generated/scheme.json",
    )
    
    return paths.AtomicWriteJSON(quickshellPath, quickshellData)
}
```

## Conclusion

This comprehensive plan provides a clear path to extend the current wallpaper Material You generator to produce complete Heimdall-compliant colorschemes. By implementing the color inference algorithms from the blueprint and ensuring proper integration with QuickShell, users will enjoy a seamless experience where selecting a wallpaper automatically generates and applies a consistent, beautiful theme across all applications.

The phased approach ensures that each component is properly tested before integration, minimizing the risk of breaking changes while maximizing the value delivered to users. The result will be a powerful, user-friendly system that bridges the gap between simple wallpaper selection and comprehensive system theming.