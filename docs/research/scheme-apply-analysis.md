# Scheme Apply Logic Analysis

## Executive Summary

This document analyzes the current scheme application flow in heimdall-cli, identifies critical issues preventing proper theme application for GTK and Kitty, and provides recommendations for fixes. The analysis reveals a fundamental color mapping issue where scheme files use `term0-term15` naming but templates expect `colour0-colour15` naming.

## Current Apply Flow

### 1. Command Entry Point
- `heimdall scheme set <scheme>` calls `applyThemeWithOptions()` in `/internal/commands/scheme/set.go`
- The function loads configuration and determines which applications to theme based on:
  - User-specified `--apps` flag (comma-separated list)
  - Configuration flags in `config.json` (e.g., `enableGtk`, `enableQt`, etc.)
  - Terminal sequences are always applied unless explicitly disabled

### 2. Application Selection Logic
```go
// Valid applications that can be themed
validApps := map[string]bool{
    "btop":      true,
    "discord":   true,
    "fuzzel":    true,
    "gtk":       true,
    "qt":        true,
    "spicetify": true,
    "terminal":  true,
}
```

### 3. Theme Application Process
For each selected application:
1. **Terminal sequences**: Special handling via `ApplyTerminalSequences()`
2. **Other apps**: Standard template processing via `ApplyTheme()`

## File Locations and Output Paths

### Current File Locations
| Application | Output Path | Template Source |
|-------------|-------------|-----------------|
| **GTK** | `~/.config/gtk-3.0/gtk.css`<br>`~/.config/gtk-4.0/gtk.css` | Dedicated GTK handler |
| **Qt** | `~/.config/qt5ct/colors/heimdall.conf`<br>`~/.config/qt6ct/colors/heimdall.conf` | Dedicated Qt handler |
| **Kitty** | `~/.config/kitty/heimdall.conf` | Template: `kitty.conf.tmpl` |
| **Discord** | Multiple client paths (Vesktop, Discord, etc.) | Discord client manager |
| **Btop** | `~/.config/btop/themes/heimdall.theme` | Embedded template |
| **Fuzzel** | `~/.config/fuzzel/fuzzel.ini` | Embedded template |
| **Spicetify** | `~/.config/spicetify/Themes/heimdall/color.ini` | Embedded template |
| **Terminal** | `~/.config/heimdall/sequences.txt` | ANSI sequence generator |

### QuickShell Integration (Working)
- **Primary**: `~/.config/heimdall/scheme.json`
- **State**: `~/.local/state/heimdall/scheme.json`
- **QuickShell**: `~/.local/state/quickshell/user/generated/scheme.json`

QuickShell works because it receives a specially formatted JSON with colors stripped of `#` prefixes and uses the `colours` key (British spelling).

## Critical Issues Identified

### 1. Color Mapping Issue (Primary Problem)
**Root Cause**: Scheme files use `term0-term15` naming but templates expect `colour0-colour15`.

**Evidence**:
- Scheme files contain: `term0`, `term1`, `term2`, etc.
- Templates reference: `{{colour0}}`, `{{colour1}}`, `{{colour2}}`, etc.
- The `GetColors()` method returns raw scheme colors without mapping

**Impact**: Templates receive undefined variables, resulting in literal `{{colour0}}` text in output files.

### 2. GTK Issue Analysis
**Current Implementation**:
- Uses dedicated `GTKHandler` in `/internal/theme/gtk.go`
- Writes to `~/.config/gtk-3.0/gtk.css` and `~/.config/gtk-4.0/gtk.css`
- Generates comprehensive CSS with color variables and widget styling

**Problems**:
1. **Color mapping**: References `colors["colour4"]`, `colors["colour8"]`, etc. but receives `term4`, `term8`
2. **Missing colors**: Many color references return empty strings
3. **No fallback**: No graceful degradation when colors are missing

**Example Issue**:
```go
// In gtk.go line 64
builder.WriteString(fmt.Sprintf("@define-color primary %s;\n", colors["colour4"]))
// colors["colour4"] is empty because scheme has "term4" instead
```

### 3. Kitty Issue Analysis
**Current Implementation**:
- Uses template-based approach with `kitty.conf.tmpl`
- Template contains: `foreground {{foreground}}`, `color0 {{colour0}}`, etc.
- Processed by `SimpleReplacer` for string substitution

**Problems**:
1. **Template variables undefined**: `{{colour0}}` through `{{colour15}}` are not found in color map
2. **Output contains literals**: Kitty config file contains literal `{{colour0}}` text
3. **No color application**: Kitty doesn't recognize the literal template syntax

**Template Content**:
```
# Kitty template expects:
color0 {{colour0}}
color1 {{colour1}}
# But scheme provides:
term0 353434
term1 ac73ff
```

### 4. QuickShell Success Analysis
**Why QuickShell Works**:
1. **Direct color mapping**: Uses raw scheme colors without expecting `colour0-colour15`
2. **Special formatting**: Strips `#` prefixes and uses British spelling (`colours`)
3. **Comprehensive data**: Includes all scheme colors, not just terminal colors
4. **Proper JSON structure**: Well-formed JSON with metadata

**QuickShell Format**:
```json
{
  "name": "catppuccin",
  "flavour": "mocha", 
  "mode": "dark",
  "variant": "tonalspot",
  "colours": {
    "term0": "353434",
    "term1": "ac73ff",
    "background": "131317",
    "foreground": "e5e1e7"
  }
}
```

## Application-Specific Requirements

### GTK Applications
**Requirements**:
- CSS color variables (`@define-color`)
- Widget-specific styling (buttons, entries, etc.)
- Proper color contrast ratios
- Support for both GTK3 and GTK4

**Current Issues**:
- Missing color mappings prevent proper variable definition
- Hardcoded color references fail when colors are undefined

### Kitty Terminal
**Requirements**:
- Standard terminal color format (`color0` through `color15`)
- Foreground, background, and cursor colors
- Simple key-value configuration format

**Current Issues**:
- Template variables remain unsubstituted
- Configuration file contains literal template syntax

### Qt Applications  
**Requirements**:
- Qt5ct/Qt6ct color scheme format
- 21-color palette for different widget states
- Proper color role mapping

**Current Issues**:
- Similar color mapping problems as GTK
- Complex color array generation fails with undefined colors

## Recommendations

### 1. Implement Color Mapping Function (Critical)
Create a color mapping function to bridge the gap between scheme naming and template expectations:

```go
// Add to scheme/manager.go
func (s *Scheme) GetMappedColors() map[string]string {
    mapped := make(map[string]string)
    
    // Copy all existing colors
    for k, v := range s.Colours {
        mapped[k] = v
    }
    
    // Map term0-term15 to colour0-colour15
    for i := 0; i < 16; i++ {
        termKey := fmt.Sprintf("term%d", i)
        colourKey := fmt.Sprintf("colour%d", i)
        if color, exists := s.Colours[termKey]; exists {
            mapped[colourKey] = color
        }
    }
    
    // Ensure # prefix for hex colors
    for k, v := range mapped {
        if !strings.HasPrefix(v, "#") && len(v) == 6 {
            mapped[k] = "#" + v
        }
    }
    
    return mapped
}
```

### 2. Update Theme Application Logic
Modify `applyThemeWithOptions()` to use mapped colors:

```go
// In set.go, replace:
colors := s.GetColors()
// With:
colors := s.GetMappedColors()
```

### 3. Add Color Validation and Fallbacks
Implement fallback colors for missing mappings:

```go
func (s *Scheme) GetMappedColorsWithFallbacks() map[string]string {
    colors := s.GetMappedColors()
    
    // Define fallback colors
    fallbacks := map[string]string{
        "colour0": "#000000", // black
        "colour1": "#ff0000", // red
        "colour2": "#00ff00", // green
        // ... etc
    }
    
    // Apply fallbacks for missing colors
    for key, fallback := range fallbacks {
        if _, exists := colors[key]; !exists {
            colors[key] = fallback
        }
    }
    
    return colors
}
```

### 4. Improve Error Handling
Add validation and logging for color mapping issues:

```go
func validateColorMapping(colors map[string]string) error {
    required := []string{"background", "foreground"}
    for i := 0; i < 16; i++ {
        required = append(required, fmt.Sprintf("colour%d", i))
    }
    
    var missing []string
    for _, key := range required {
        if _, exists := colors[key]; !exists {
            missing = append(missing, key)
        }
    }
    
    if len(missing) > 0 {
        return fmt.Errorf("missing required colors: %v", missing)
    }
    
    return nil
}
```

### 5. Update Templates for Consistency
Ensure all templates use consistent color naming:
- Standardize on `colour0-colour15` for terminal colors
- Use `background`, `foreground` for base colors
- Add template validation to catch undefined variables

### 6. Enhance QuickShell Integration
While QuickShell works, improve the integration:
- Add error handling for QuickShell directory creation
- Validate QuickShell JSON format
- Add logging for successful QuickShell updates

## Testing Strategy

### 1. Unit Tests
- Test color mapping function with various scheme formats
- Validate template processing with mapped colors
- Test fallback color application

### 2. Integration Tests
- Test complete scheme application flow
- Verify file creation and content for each application
- Test with different scheme formats (term vs colour naming)

### 3. Manual Testing
- Apply schemes and verify actual application theming
- Test GTK applications (file managers, text editors)
- Test Kitty terminal color changes
- Verify QuickShell continues working

## Implementation Priority

1. **High Priority**: Implement color mapping function
2. **High Priority**: Update theme application to use mapped colors
3. **Medium Priority**: Add color validation and fallbacks
4. **Medium Priority**: Improve error handling and logging
5. **Low Priority**: Template consistency improvements

## Conclusion

The primary issue preventing GTK and Kitty theming is the color naming mismatch between scheme files (`term0-term15`) and templates (`colour0-colour15`). QuickShell works because it uses the raw scheme colors directly. Implementing a color mapping function will resolve the majority of theming issues and bring GTK and Kitty applications in line with the working QuickShell integration.

The recommended fixes are straightforward to implement and will significantly improve the user experience by ensuring consistent theming across all supported applications.