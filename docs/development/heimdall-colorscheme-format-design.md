# Heimdall Colorscheme Format Design

## Table of Contents

1. [Overview](#overview)
2. [Format Structure](#format-structure)
3. [Color Key Categories](#color-key-categories)
4. [Material Design 3 Color System](#material-design-3-color-system)
5. [Terminal Colors](#terminal-colors)
6. [Semantic Colors](#semantic-colors)
7. [Theme-Specific Colors](#theme-specific-colors)
8. [Contrast and Accessibility](#contrast-and-accessibility)
9. [Application Mapping](#application-mapping)
10. [Best Practices](#best-practices)
11. [Porting Guidelines](#porting-guidelines)

## Overview

The Heimdall colorscheme format is a comprehensive, JSON-based color definition system designed to provide consistent theming across multiple applications and environments. It combines Material Design 3 principles with traditional terminal colors and theme-specific extensions to create a flexible yet standardized approach to color management.

### Design Principles

- **Completeness**: Every color needed by any supported application should be definable
- **Consistency**: Same semantic meaning across different applications
- **Compatibility**: Support for legacy terminal colors and modern design systems
- **Extensibility**: Allow theme-specific colors without breaking core functionality
- **Accessibility**: Built-in support for contrast requirements and readability

## Format Structure

### Basic Structure

```json
{
  "name": "theme-name",
  "flavour": "variant-name",
  "mode": "dark|light",
  "colours": {
    // All color definitions
  }
}
```

### Field Definitions

- **name**: The theme family name (e.g., "catppuccin", "gruvbox", "rosepine")
- **flavour**: The specific variant within the theme family (e.g., "mocha", "medium", "main")
- **mode**: Either "dark" or "light" to indicate the theme's brightness
- **colours**: Object containing all color definitions with hex values (including # prefix)

### File Organization

Themes are organized in a hierarchical directory structure:
```
assets/schemes/
├── theme-name/
│   └── flavour-name/
│       ├── dark.json
│       └── light.json
```

## Color Key Categories

The Heimdall format organizes colors into several logical categories, each serving specific purposes:

### 1. Core Colors
Essential colors that form the foundation of any theme:
- `background`: Primary background color
- `foreground`: Primary text/foreground color
- `text`: Alias for foreground (for clarity in some contexts)

### 2. Material Design 3 System
Comprehensive Material Design color tokens for modern UI frameworks

### 3. Terminal Colors
Traditional 16-color terminal palette with multiple naming conventions

### 4. Semantic Colors
Purpose-driven colors for UI states and components

### 5. Theme-Specific Extensions
Custom colors for theme compatibility (e.g., Catppuccin's named colors)

## Material Design 3 Color System

The Material Design 3 color system provides a sophisticated palette for modern applications:

### Primary Colors
```json
{
  "primary": "#89b4fa",              // Main brand color
  "onPrimary": "#29364b",            // Text/icons on primary
  "primaryContainer": "#5f7daf",      // Container using primary
  "onPrimaryContainer": "#1b2331",    // Text/icons on primary container
  "primaryFixed": "#a0c3fb",          // Fixed primary (doesn't change with theme)
  "primaryFixedDim": "#89b4fa",       // Dimmed version of fixed primary
  "onPrimaryFixed": "#364864",        // Text/icons on fixed primary
  "onPrimaryFixedVariant": "#526c96", // Variant text on fixed primary
  "primary_paletteKeyColor": "#5f7daf" // Key color for palette generation
}
```

### Secondary Colors
```json
{
  "secondary": "#94e2d5",
  "onSecondary": "#2c433f",
  "secondaryContainer": "#679e95",
  "onSecondaryContainer": "#b4eae1",
  "secondaryFixed": "#a9e7dd",
  "secondaryFixedDim": "#94e2d5",
  "onSecondaryFixed": "#3b5a55",
  "onSecondaryFixedVariant": "#58877f",
  "secondary_paletteKeyColor": "#679e95"
}
```

### Tertiary Colors
```json
{
  "tertiary": "#f5c2e7",
  "tertiaryContainer": "#ab87a1",
  "onTertiaryContainer": "#f8d4ee",
  "tertiaryFixed": "#f7ceeb",
  "tertiaryFixedDim": "#f5c2e7",
  "onTertiaryFixed": "#624d5c",
  "onTertiaryFixedVariant": "#93748a",
  "tertiary_paletteKeyColor": "#ab87a1"
}
```

### Surface Colors
```json
{
  "surface": "#1e1e2e",                // Default surface
  "onSurface": "#cdd6f4",              // Text/icons on surface
  "surfaceDim": "#1e1e2e",             // Dimmed surface
  "surfaceBright": "#3f3f4d",          // Brightened surface
  "surfaceVariant": "#343442",         // Variant surface
  "onSurfaceVariant": "#b8c0db",       // Text on surface variant
  "surfaceContainerLowest": "#1c1c2b", // Elevation level 0
  "surfaceContainerLow": "#242434",    // Elevation level 1
  "surfaceContainer": "#2b2b3a",       // Elevation level 2
  "surfaceContainerHigh": "#323240",   // Elevation level 3
  "surfaceContainerHighest": "#393947", // Elevation level 4
  "surfaceTint": "#89b4fa"             // Tint color for surfaces
}
```

### Additional Material Colors
```json
{
  "outline": "#61616c",          // Borders and dividers
  "outlineVariant": "#3f3f4d",   // Secondary borders
  "shadow": "#000000",           // Shadow color
  "scrim": "#000000",            // Scrim overlay color
  "inverseSurface": "#cdd6f4",   // Inverted surface
  "inverseOnSurface": "#1e1e2e", // Text on inverted surface
  "inversePrimary": "#445a7d"    // Inverted primary color
}
```

## Terminal Colors

The format supports multiple naming conventions for terminal colors to ensure compatibility:

### Standard Terminal Colors (0-15)
```json
{
  "term0": "#45475a",   // Black
  "term1": "#f38ba8",   // Red
  "term2": "#a6e3a1",   // Green
  "term3": "#f9e2af",   // Yellow
  "term4": "#89b4fa",   // Blue
  "term5": "#f5c2e7",   // Magenta
  "term6": "#94e2d5",   // Cyan
  "term7": "#bac2de",   // White
  "term8": "#585b70",   // Bright Black
  "term9": "#f38ba8",   // Bright Red
  "term10": "#a6e3a1",  // Bright Green
  "term11": "#f9e2af",  // Bright Yellow
  "term12": "#89b4fa",  // Bright Blue
  "term13": "#f5c2e7",  // Bright Magenta
  "term14": "#94e2d5",  // Bright Cyan
  "term15": "#a6adc8"   // Bright White
}
```

### Alternative Terminal Color Names
```json
{
  "color0" through "color15": // Same as term0-15
}
```

### British Spelling Support
For compatibility with applications expecting British spelling:
```json
{
  "colour0" through "colour15": // Same as color0-15
}
```

## Semantic Colors

Semantic colors provide meaning-based color definitions:

### Status Colors
```json
{
  "error": "#f38ba8",
  "onError": "#301b21",
  "errorContainer": "#794554",
  "onErrorContainer": "#f6adc2",
  
  "success": "#a6e3a1",
  "onSuccess": "#314430",
  "successContainer": "#638860",
  "onSuccessContainer": "#c0ebbd"
}
```

### Base Colors (Catppuccin-style)
```json
{
  "base": "#1e1e2e",    // Darkest background
  "mantle": "#181825",  // Slightly lighter than base
  "crust": "#11111b"    // Darkest color in palette
}
```

### Overlay and Subtext Colors
```json
{
  "overlay0": "#6c7086",  // Lightest overlay
  "overlay1": "#7f849c",  // Medium overlay
  "overlay2": "#9399b2",  // Darkest overlay
  "subtext0": "#a6adc8",  // Muted text
  "subtext1": "#bac2de"   // More prominent muted text
}
```

### Surface Levels (Alternative naming)
```json
{
  "surface0": "#313244",  // Lightest surface
  "surface1": "#45475a",  // Medium surface
  "surface2": "#585b70"   // Darkest surface
}
```

## Theme-Specific Colors

### Catppuccin Named Colors
Catppuccin themes include poetic color names for brand consistency:
```json
{
  "rosewater": "#f5e0dc",
  "flamingo": "#f2cdcd",
  "pink": "#f5c2e7",
  "mauve": "#cba6f7",
  "red": "#f38ba8",
  "maroon": "#eba0ac",
  "peach": "#fab387",
  "yellow": "#f9e2af",
  "green": "#a6e3a1",
  "teal": "#94e2d5",
  "sky": "#89dceb",
  "sapphire": "#74c7ec",
  "blue": "#89b4fa",
  "lavender": "#b4befe"
}
```

### Palette Key Colors
Used for Material You dynamic color generation:
```json
{
  "neutral_paletteKeyColor": "#585b70",
  "neutral_variant_paletteKeyColor": "#343442"
}
```

## Contrast and Accessibility

### Contrast Requirements

Based on analysis of curated themes, Heimdall enforces these contrast ratios:

#### Dark Themes
- **Background to Foreground**: Minimum 7:1 (WCAG AAA)
- **Surface to OnSurface**: Minimum 4.5:1 (WCAG AA)
- **Primary to OnPrimary**: Minimum 4.5:1
- **Container to OnContainer**: Minimum 3:1

#### Light Themes
- **Background to Foreground**: Minimum 7:1 (WCAG AAA)
- **Surface to OnSurface**: Minimum 4.5:1 (WCAG AA)
- **Primary to OnPrimary**: Minimum 4.5:1
- **Container to OnContainer**: Minimum 3:1

### Color Relationships

#### Elevation Hierarchy (Dark Mode)
```
background (#1e1e2e) → darkest
  ↓
surfaceContainerLowest (#1c1c2b)
  ↓
surfaceContainerLow (#242434)
  ↓
surfaceContainer (#2b2b3a)
  ↓
surfaceContainerHigh (#323240)
  ↓
surfaceContainerHighest (#393947) → lightest
```

#### Elevation Hierarchy (Light Mode)
```
background (#eff1f5) → lightest
  ↓
surfaceContainerLowest (#e3e4e8)
  ↓
surfaceContainerLow (#eff1f5)
  ↓
surfaceContainer (#eff1f5)
  ↓
surfaceContainerHigh (#f0f2f5)
  ↓
surfaceContainerHighest (#f0f2f6) → darkest
```

### Readability Guidelines

1. **Text on Background**: Use `foreground` for primary text
2. **Muted Text**: Use `subtext0` or `subtext1` for secondary content
3. **Interactive Elements**: Use `primary` for links and buttons
4. **Disabled States**: Use `overlay0` with reduced opacity
5. **Borders**: Use `outline` for primary, `outlineVariant` for subtle

## Application Mapping

### Kitty Terminal
```
foreground → foreground
background → background
color0-15 → color0-15
selection_foreground → background
selection_background → foreground
url_color → color4 (blue)
active_border_color → color4
inactive_border_color → color8
active_tab_background → color5
inactive_tab_background → color0
```

### Discord
```
background → background-primary
surface0 → background-secondary
surface1 → background-tertiary
foreground → text-normal
subtext1 → text-muted
primary → text-link
overlay1 → interactive-normal
foreground → interactive-hover
```

### GTK Applications
```
background → window background
foreground → default text
primary → button background
surface → entry background
outline → border color
error → error states
success → success states
```

### Hyprland
```
primary → active border (no # prefix)
outline → inactive border (no # prefix)
background → background (RGBA format)
```

### QuickShell
All colors exported without # prefix, using British spelling "colours"

### Neovim
```
background → Normal bg
foreground → Normal fg
primary → Function, Keyword
secondary → String
tertiary → Type
error → Error, DiagnosticError
success → DiagnosticOk
surface0-2 → StatusLine gradients
```

## Best Practices

### Color Selection

1. **Maintain Semantic Consistency**
   - Red tones for errors/destructive actions
   - Green tones for success/positive actions
   - Blue tones for primary actions/links
   - Yellow/amber for warnings

2. **Respect Mode Conventions**
   - Dark mode: Dark backgrounds (#000000-#3f3f3f range)
   - Light mode: Light backgrounds (#e0e0e0-#ffffff range)

3. **Surface Elevation**
   - Each elevation level should be visually distinct
   - Maintain 3-5% lightness difference between levels
   - Higher elevation = lighter in dark mode, darker in light mode

4. **Terminal Color Harmony**
   - Ensure terminal colors work well together
   - Maintain sufficient contrast against background
   - Test in common terminal applications (vim, htop, ls)

### Contrast Patterns from Curated Themes

#### Catppuccin Mocha (Dark)
- Background (#1e1e2e) to Foreground (#cdd6f4): 11.2:1
- Uses cooler tones with purple/blue bias
- High saturation for accent colors

#### Gruvbox Medium (Dark)
- Background (#101415) to Foreground (#e0e3e4): 13.8:1
- Warmer, earthy tones
- Lower saturation for comfortable viewing

#### Rosepine Main (Dark)
- Background (#141317) to Foreground (#e5e1e7): 13.5:1
- Muted, sophisticated palette
- Purple and pink accents

#### Catppuccin Latte (Light)
- Background (#eff1f5) to Foreground (#4c4f69): 8.9:1
- Soft, pastel approach to light themes
- Maintains brand colors in light mode

### Color Temperature Guidelines

1. **Cool Themes** (Catppuccin, OneDark)
   - Blue/purple bias in grays
   - Cooler accent colors
   - Good for extended coding sessions

2. **Warm Themes** (Gruvbox)
   - Yellow/brown bias in grays
   - Warmer accent colors
   - Comfortable in low-light environments

3. **Neutral Themes** (Rosepine)
   - Balanced temperature
   - Sophisticated, muted tones
   - Professional appearance

## Porting Guidelines

### Step 1: Analyze Source Theme

1. Identify the color palette structure
2. Determine light/dark mode
3. Map colors to semantic meanings
4. Measure contrast ratios

### Step 2: Map to Heimdall Format

#### Required Minimal Set
```json
{
  "background": "#...",
  "foreground": "#...",
  "term0-15" or "color0-15": "#...",
  "primary": "#...",
  "secondary": "#...",
  "error": "#..."
}
```

#### Extended Material Design Set
Add Material Design 3 colors by:
1. Using primary color as base
2. Generating containers with 30% opacity over background
3. Creating "on" colors with sufficient contrast
4. Adding surface elevation hierarchy

#### Surface Generation Formula
For dark themes:
```
surfaceContainerLowest = darken(background, 2%)
surfaceContainerLow = lighten(background, 3%)
surfaceContainer = lighten(background, 5%)
surfaceContainerHigh = lighten(background, 8%)
surfaceContainerHighest = lighten(background, 12%)
```

For light themes:
```
surfaceContainerLowest = darken(background, 3%)
surfaceContainerLow = background
surfaceContainer = background
surfaceContainerHigh = lighten(background, 1%)
surfaceContainerHighest = lighten(background, 2%)
```

### Step 3: Terminal Color Mapping

Standard mapping for unknown themes:
```
term0/color0 = 20% lightness (dark gray/black)
term1/color1 = error or red accent
term2/color2 = success or green accent
term3/color3 = warning or yellow accent
term4/color4 = primary or blue accent
term5/color5 = secondary or magenta accent
term6/color6 = tertiary or cyan accent
term7/color7 = 70% lightness (light gray)
term8-15 = brighter versions of 0-7
```

### Step 4: Validation

1. **Contrast Check**
   - Verify all text combinations meet WCAG AA
   - Test in both light and dark environments

2. **Application Testing**
   - Apply to kitty terminal
   - Test with syntax highlighting
   - Verify UI elements in GTK/Qt apps

3. **Semantic Verification**
   - Ensure error colors are distinguishable
   - Verify success/warning differentiation
   - Check link visibility

### Step 5: Theme-Specific Extensions

Add theme family colors if part of established system:
- Catppuccin: Add rosewater, flamingo, pink, etc.
- Gruvbox: Add orange, aqua variants
- Dracula: Add comment, current line, etc.

### Example Port: Dracula Theme

Source colors:
```
Background: #282a36
Foreground: #f8f8f2
Comment: #6272a4
Cyan: #8be9fd
Green: #50fa7b
Orange: #ffb86c
Pink: #ff79c6
Purple: #bd93f9
Red: #ff5555
Yellow: #f1fa8c
```

Heimdall format:
```json
{
  "name": "dracula",
  "flavour": "default",
  "mode": "dark",
  "colours": {
    "background": "#282a36",
    "foreground": "#f8f8f2",
    "text": "#f8f8f2",
    "base": "#282a36",
    "mantle": "#21222c",
    "crust": "#191a21",
    
    "primary": "#bd93f9",
    "onPrimary": "#2e2640",
    "primaryContainer": "#6c5a8c",
    "onPrimaryContainer": "#e4d9f4",
    
    "secondary": "#ff79c6",
    "onSecondary": "#4d1f3a",
    "secondaryContainer": "#a64d8a",
    "onSecondaryContainer": "#ffd9ec",
    
    "tertiary": "#8be9fd",
    "tertiaryContainer": "#4a8a99",
    "onTertiaryContainer": "#d9f4f9",
    
    "error": "#ff5555",
    "success": "#50fa7b",
    
    "surface": "#282a36",
    "onSurface": "#f8f8f2",
    "surfaceContainerLowest": "#21222c",
    "surfaceContainerLow": "#2e3040",
    "surfaceContainer": "#353746",
    "surfaceContainerHigh": "#3c3e4d",
    "surfaceContainerHighest": "#434554",
    
    "outline": "#6272a4",
    "outlineVariant": "#44475a",
    
    "term0": "#21222c",
    "term1": "#ff5555",
    "term2": "#50fa7b",
    "term3": "#f1fa8c",
    "term4": "#bd93f9",
    "term5": "#ff79c6",
    "term6": "#8be9fd",
    "term7": "#f8f8f2",
    "term8": "#6272a4",
    "term9": "#ff6e6e",
    "term10": "#69ff94",
    "term11": "#ffffa5",
    "term12": "#d6acff",
    "term13": "#ff92df",
    "term14": "#a4ffff",
    "term15": "#ffffff"
  }
}
```

## Compatibility Layers

### British vs American Spelling
- The format accepts both "colour" and "color" prefixes
- Internally stored with American spelling
- QuickShell export uses British spelling

### Hash Prefix Handling
- All colors stored with # prefix in scheme files
- QuickShell export removes # prefix
- Terminal applications handle both formats

### Format Conversions
- RGB: For CSS (`rgb(r, g, b)`)
- RGBA: For transparency (`rgba(r, g, b, a)`)
- Hex: Standard format (`#rrggbb`)
- No-hash: For specific applications (`rrggbb`)

### Legacy Support
- `term0-15`: Traditional terminal colors
- `color0-15`: Alternative naming
- `colour0-15`: British spelling variant

## Conclusion

The Heimdall colorscheme format provides a comprehensive, well-structured approach to theme management that balances modern design principles with practical compatibility needs. By following these guidelines, theme authors can create consistent, accessible, and beautiful color schemes that work seamlessly across the entire application ecosystem.

Key takeaways:
1. Use Material Design 3 tokens for modern UI consistency
2. Maintain WCAG AA contrast minimums
3. Include all terminal colors for compatibility
4. Follow elevation hierarchy for surface colors
5. Test across multiple applications before release
6. Document theme-specific extensions clearly