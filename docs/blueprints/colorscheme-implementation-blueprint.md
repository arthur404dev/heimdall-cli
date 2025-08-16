# Colorscheme Implementation Blueprint

## Pattern Overview

This blueprint provides complete implementation guidance for generating Heimdall-compliant colorschemes from various input sources. The system generates all 122 required color keys while maintaining visual harmony, proper contrast ratios, and semantic consistency.

## Input Analysis and Validation

### Supported Input Types
1. **Minimal Set** (3-5 colors): Background, foreground, 1-3 accent colors
2. **Standard Terminal** (8-16 colors): ANSI colors with optional bright variants
3. **Extended Set** (20+ colors): Includes semantic colors and Material Design tokens
4. **Screenshot/Image**: Extract dominant colors and generate palette
5. **Single Accent + Mode**: One accent color plus light/dark mode preference

### Input Validation
```go
type ColorInput struct {
    Background  string   `json:"background"`
    Foreground  string   `json:"foreground"`
    Accents     []string `json:"accents"`     // Primary colors
    Terminal    []string `json:"terminal"`    // ANSI 0-15
    Mode        string   `json:"mode"`        // "light" or "dark"
    Temperature string   `json:"temperature"` // "cool", "warm", "neutral"
}

func ValidateInput(input ColorInput) error {
    // Validate hex format
    if !isValidHex(input.Background) {
        return fmt.Errorf("invalid background color")
    }
    
    // Check minimum requirements
    if input.Background == "" || input.Foreground == "" {
        return fmt.Errorf("background and foreground required")
    }
    
    // Validate contrast
    contrast := calculateContrast(input.Background, input.Foreground)
    if contrast < 7.0 {
        return fmt.Errorf("insufficient contrast: %.2f (minimum 7.0)", contrast)
    }
    
    return nil
}
```

## Color Inference Engine

### Core Algorithm Structure
```go
type ColorScheme struct {
    // Core colors
    Background string
    Foreground string
    Text       string
    
    // Material Design 3
    Primary              string
    OnPrimary           string
    PrimaryContainer    string
    OnPrimaryContainer  string
    PrimaryFixed        string
    PrimaryFixedDim     string
    OnPrimaryFixed      string
    OnPrimaryFixedVariant string
    
    Secondary           string
    OnSecondary        string
    SecondaryContainer  string
    OnSecondaryContainer string
    SecondaryFixed      string
    SecondaryFixedDim   string
    OnSecondaryFixed    string
    OnSecondaryFixedVariant string
    
    Tertiary            string
    OnTertiary         string
    TertiaryContainer   string
    OnTertiaryContainer string
    TertiaryFixed       string
    TertiaryFixedDim    string
    OnTertiaryFixed     string
    OnTertiaryFixedVariant string
    
    // Surfaces
    Surface                  string
    OnSurface               string
    SurfaceDim              string
    SurfaceBright           string
    SurfaceVariant          string
    OnSurfaceVariant        string
    SurfaceContainerLowest  string
    SurfaceContainerLow     string
    SurfaceContainer        string
    SurfaceContainerHigh    string
    SurfaceContainerHighest string
    SurfaceTint             string
    
    // Additional MD3
    Outline         string
    OutlineVariant  string
    Shadow          string
    Scrim           string
    InverseSurface  string
    InverseOnSurface string
    InversePrimary  string
    
    // Terminal colors
    Term0  string // Black
    Term1  string // Red
    Term2  string // Green
    Term3  string // Yellow
    Term4  string // Blue
    Term5  string // Magenta
    Term6  string // Cyan
    Term7  string // White
    Term8  string // Bright Black
    Term9  string // Bright Red
    Term10 string // Bright Green
    Term11 string // Bright Yellow
    Term12 string // Bright Blue
    Term13 string // Bright Magenta
    Term14 string // Bright Cyan
    Term15 string // Bright White
    
    // Semantic
    Error              string
    OnError           string
    ErrorContainer    string
    OnErrorContainer  string
    Success           string
    OnSuccess         string
    SuccessContainer  string
    OnSuccessContainer string
    
    // Theme-specific
    Base     string
    Mantle   string
    Crust    string
    Overlay0 string
    Overlay1 string
    Overlay2 string
    Subtext0 string
    Subtext1 string
    Surface0 string
    Surface1 string
    Surface2 string
}
```

### HSL Manipulation Functions
```go
type HSL struct {
    H float64 // Hue: 0-360
    S float64 // Saturation: 0-100
    L float64 // Lightness: 0-100
}

func hexToHSL(hex string) HSL {
    r, g, b := hexToRGB(hex)
    
    // Normalize RGB values
    rf := float64(r) / 255.0
    gf := float64(g) / 255.0
    bf := float64(b) / 255.0
    
    max := math.Max(math.Max(rf, gf), bf)
    min := math.Min(math.Min(rf, gf), bf)
    delta := max - min
    
    // Calculate lightness
    l := (max + min) / 2.0
    
    if delta == 0 {
        return HSL{0, 0, l * 100}
    }
    
    // Calculate saturation
    var s float64
    if l < 0.5 {
        s = delta / (max + min)
    } else {
        s = delta / (2.0 - max - min)
    }
    
    // Calculate hue
    var h float64
    switch max {
    case rf:
        h = ((gf - bf) / delta) + (gf < bf ? 6 : 0)
    case gf:
        h = ((bf - rf) / delta) + 2
    case bf:
        h = ((rf - gf) / delta) + 4
    }
    h = h * 60
    
    return HSL{h, s * 100, l * 100}
}

func hslToHex(hsl HSL) string {
    h := hsl.H / 360.0
    s := hsl.S / 100.0
    l := hsl.L / 100.0
    
    var r, g, b float64
    
    if s == 0 {
        r, g, b = l, l, l
    } else {
        var q float64
        if l < 0.5 {
            q = l * (1 + s)
        } else {
            q = l + s - l*s
        }
        p := 2*l - q
        
        r = hueToRGB(p, q, h+1.0/3.0)
        g = hueToRGB(p, q, h)
        b = hueToRGB(p, q, h-1.0/3.0)
    }
    
    return fmt.Sprintf("#%02x%02x%02x", 
        int(r*255), int(g*255), int(b*255))
}

func adjustLightness(hex string, amount float64) string {
    hsl := hexToHSL(hex)
    hsl.L = math.Max(0, math.Min(100, hsl.L + amount))
    return hslToHex(hsl)
}

func adjustSaturation(hex string, amount float64) string {
    hsl := hexToHSL(hex)
    hsl.S = math.Max(0, math.Min(100, hsl.S + amount))
    return hslToHex(hsl)
}

func mixColors(hex1, hex2 string, ratio float64) string {
    hsl1 := hexToHSL(hex1)
    hsl2 := hexToHSL(hex2)
    
    // Mix in HSL space for better perceptual results
    mixed := HSL{
        H: hsl1.H*(1-ratio) + hsl2.H*ratio,
        S: hsl1.S*(1-ratio) + hsl2.S*ratio,
        L: hsl1.L*(1-ratio) + hsl2.L*ratio,
    }
    
    return hslToHex(mixed)
}
```

### Luminance and Contrast Calculations
```go
func relativeLuminance(hex string) float64 {
    r, g, b := hexToRGB(hex)
    
    // Convert to linear RGB
    rf := gammaCorrect(float64(r) / 255.0)
    gf := gammaCorrect(float64(g) / 255.0)
    bf := gammaCorrect(float64(b) / 255.0)
    
    // Calculate relative luminance
    return 0.2126*rf + 0.7152*gf + 0.0722*bf
}

func gammaCorrect(val float64) float64 {
    if val <= 0.03928 {
        return val / 12.92
    }
    return math.Pow((val+0.055)/1.055, 2.4)
}

func calculateContrast(hex1, hex2 string) float64 {
    l1 := relativeLuminance(hex1)
    l2 := relativeLuminance(hex2)
    
    lighter := math.Max(l1, l2)
    darker := math.Min(l1, l2)
    
    return (lighter + 0.05) / (darker + 0.05)
}

func ensureContrast(bg, fg string, minContrast float64) string {
    currentContrast := calculateContrast(bg, fg)
    if currentContrast >= minContrast {
        return fg
    }
    
    bgLum := relativeLuminance(bg)
    fgHSL := hexToHSL(fg)
    
    // Determine direction to adjust
    if bgLum > 0.5 {
        // Light background - darken foreground
        for fgHSL.L > 0 {
            fgHSL.L -= 1
            newFg := hslToHex(fgHSL)
            if calculateContrast(bg, newFg) >= minContrast {
                return newFg
            }
        }
    } else {
        // Dark background - lighten foreground
        for fgHSL.L < 100 {
            fgHSL.L += 1
            newFg := hslToHex(fgHSL)
            if calculateContrast(bg, newFg) >= minContrast {
                return newFg
            }
        }
    }
    
    return fg // Fallback
}
```

## Material Design 3 Generation

### Primary Color System
```go
func generateMaterialPrimary(baseColor, background string, isDark bool) map[string]string {
    primary := make(map[string]string)
    
    // Base primary color
    primary["primary"] = baseColor
    
    // Generate container - blend with background
    primary["primaryContainer"] = mixColors(baseColor, background, 0.7)
    
    // Generate "on" colors with proper contrast
    primary["onPrimary"] = ensureContrast(baseColor, 
        isDark ? "#ffffff" : "#000000", 4.5)
    primary["onPrimaryContainer"] = ensureContrast(
        primary["primaryContainer"], 
        adjustLightness(baseColor, isDark ? 20 : -20), 3.0)
    
    // Fixed variants (don't change with theme)
    if isDark {
        primary["primaryFixed"] = adjustLightness(baseColor, 15)
        primary["primaryFixedDim"] = baseColor
        primary["onPrimaryFixed"] = adjustLightness(baseColor, -40)
        primary["onPrimaryFixedVariant"] = adjustLightness(baseColor, -20)
    } else {
        primary["primaryFixed"] = adjustLightness(baseColor, -10)
        primary["primaryFixedDim"] = adjustLightness(baseColor, -5)
        primary["onPrimaryFixed"] = adjustLightness(baseColor, 40)
        primary["onPrimaryFixedVariant"] = adjustLightness(baseColor, 20)
    }
    
    // Palette key color for Material You
    primary["primary_paletteKeyColor"] = primary["primaryContainer"]
    
    return primary
}

func generateMaterialSecondary(primaryColor, background string, isDark bool) map[string]string {
    // Generate complementary color
    primaryHSL := hexToHSL(primaryColor)
    
    // Shift hue by 120 degrees for triadic harmony
    secondaryHSL := HSL{
        H: math.Mod(primaryHSL.H + 120, 360),
        S: primaryHSL.S * 0.8, // Slightly less saturated
        L: primaryHSL.L,
    }
    
    secondaryBase := hslToHex(secondaryHSL)
    
    // Apply same generation logic as primary
    return generateMaterialPrimary(secondaryBase, background, isDark)
}

func generateMaterialTertiary(primaryColor, background string, isDark bool) map[string]string {
    // Generate split-complementary color
    primaryHSL := hexToHSL(primaryColor)
    
    // Shift hue by 60 degrees
    tertiaryHSL := HSL{
        H: math.Mod(primaryHSL.H + 60, 360),
        S: primaryHSL.S * 0.9,
        L: primaryHSL.L,
    }
    
    tertiaryBase := hslToHex(tertiaryHSL)
    
    // Apply same generation logic
    return generateMaterialPrimary(tertiaryBase, background, isDark)
}
```

### Tonal Palette Generation
```go
func generateTonalPalette(seedColor string) []string {
    hsl := hexToHSL(seedColor)
    tones := []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99, 100}
    
    palette := make([]string, len(tones))
    for i, tone := range tones {
        // Material Design 3 tonal algorithm
        newHSL := HSL{
            H: hsl.H,
            S: hsl.S * (1 - tone/200), // Reduce saturation as lightness increases
            L: tone,
        }
        palette[i] = hslToHex(newHSL)
    }
    
    return palette
}
```

## Surface Hierarchy Generation

### Dark Mode Surfaces
```go
func generateDarkSurfaces(background string) map[string]string {
    surfaces := make(map[string]string)
    
    bgHSL := hexToHSL(background)
    
    // Base surface equals background
    surfaces["surface"] = background
    
    // Progressive lightening for elevation
    surfaces["surfaceContainerLowest"] = adjustLightness(background, -2)
    surfaces["surfaceContainerLow"] = adjustLightness(background, 3)
    surfaces["surfaceContainer"] = adjustLightness(background, 5)
    surfaces["surfaceContainerHigh"] = adjustLightness(background, 8)
    surfaces["surfaceContainerHighest"] = adjustLightness(background, 12)
    
    // Surface variants
    surfaces["surfaceDim"] = adjustLightness(background, -5)
    surfaces["surfaceBright"] = adjustLightness(background, 15)
    surfaces["surfaceVariant"] = hslToHex(HSL{
        H: bgHSL.H,
        S: bgHSL.S * 1.1,
        L: bgHSL.L + 7,
    })
    
    return surfaces
}
```

### Light Mode Surfaces
```go
func generateLightSurfaces(background string) map[string]string {
    surfaces := make(map[string]string)
    
    bgHSL := hexToHSL(background)
    
    surfaces["surface"] = background
    
    // Progressive darkening for elevation (inverse of dark mode)
    surfaces["surfaceContainerLowest"] = adjustLightness(background, -3)
    surfaces["surfaceContainerLow"] = background
    surfaces["surfaceContainer"] = background
    surfaces["surfaceContainerHigh"] = adjustLightness(background, 1)
    surfaces["surfaceContainerHighest"] = adjustLightness(background, 2)
    
    // Surface variants
    surfaces["surfaceDim"] = adjustLightness(background, -8)
    surfaces["surfaceBright"] = adjustLightness(background, 5)
    surfaces["surfaceVariant"] = hslToHex(HSL{
        H: bgHSL.H,
        S: bgHSL.S * 0.9,
        L: bgHSL.L - 5,
    })
    
    return surfaces
}
```

## Light/Dark Theme Conversion

### LAB Color Space Conversion
```go
type LAB struct {
    L float64 // Lightness: 0-100
    A float64 // Green-Red: -128 to 127
    B float64 // Blue-Yellow: -128 to 127
}

func hexToLAB(hex string) LAB {
    r, g, b := hexToRGB(hex)
    
    // Convert to XYZ
    x, y, z := rgbToXYZ(r, g, b)
    
    // Convert XYZ to LAB
    return xyzToLAB(x, y, z)
}

func labToHex(lab LAB) string {
    // Convert LAB to XYZ
    x, y, z := labToXYZ(lab)
    
    // Convert XYZ to RGB
    r, g, b := xyzToRGB(x, y, z)
    
    return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func invertLightness(hex string) string {
    lab := hexToLAB(hex)
    
    // Invert lightness while preserving color
    lab.L = 100 - lab.L
    
    // Adjust chroma slightly for perceptual consistency
    chromaFactor := lab.L / 50.0 // Reduce chroma at extremes
    if chromaFactor > 1 {
        chromaFactor = 2 - chromaFactor
    }
    lab.A *= chromaFactor
    lab.B *= chromaFactor
    
    return labToHex(lab)
}
```

### Bidirectional Theme Conversion
```go
func convertThemeMode(scheme *ColorScheme, targetMode string) *ColorScheme {
    newScheme := &ColorScheme{}
    
    if targetMode == "light" {
        // Dark to light conversion
        newScheme.Background = invertLightness(scheme.Background)
        newScheme.Foreground = invertLightness(scheme.Foreground)
        
        // Ensure proper contrast
        newScheme.Foreground = ensureContrast(
            newScheme.Background, 
            newScheme.Foreground, 
            7.0,
        )
        
        // Convert accent colors with temperature preservation
        newScheme.Primary = preserveTemperature(
            invertLightness(scheme.Primary),
            scheme.Primary,
        )
        
        // Generate new surfaces for light mode
        surfaces := generateLightSurfaces(newScheme.Background)
        // ... apply surfaces
        
    } else {
        // Light to dark conversion
        newScheme.Background = invertLightness(scheme.Background)
        newScheme.Foreground = invertLightness(scheme.Foreground)
        
        // Ensure proper contrast
        newScheme.Foreground = ensureContrast(
            newScheme.Background,
            newScheme.Foreground,
            7.0,
        )
        
        // Generate new surfaces for dark mode
        surfaces := generateDarkSurfaces(newScheme.Background)
        // ... apply surfaces
    }
    
    return newScheme
}

func preserveTemperature(converted, original string) string {
    convLAB := hexToLAB(converted)
    origLAB := hexToLAB(original)
    
    // Preserve the color temperature (a and b channels)
    // while keeping converted lightness
    preserved := LAB{
        L: convLAB.L,
        A: origLAB.A * (convLAB.L / 50.0), // Scale by lightness
        B: origLAB.B * (convLAB.L / 50.0),
    }
    
    return labToHex(preserved)
}
```

## ANSI Color Mapping

### Semantic ANSI Mapping
```go
func generateANSIColors(input ColorInput, scheme *ColorScheme) map[string]string {
    ansi := make(map[string]string)
    
    if len(input.Terminal) >= 8 {
        // Use provided terminal colors
        for i := 0; i < min(16, len(input.Terminal)); i++ {
            ansi[fmt.Sprintf("term%d", i)] = input.Terminal[i]
        }
    } else {
        // Generate from accent colors
        ansi["term0"] = generateBlack(scheme.Background)
        ansi["term1"] = findOrGenerateRed(input.Accents, scheme)
        ansi["term2"] = findOrGenerateGreen(input.Accents, scheme)
        ansi["term3"] = findOrGenerateYellow(input.Accents, scheme)
        ansi["term4"] = findOrGenerateBlue(input.Accents, scheme)
        ansi["term5"] = findOrGenerateMagenta(input.Accents, scheme)
        ansi["term6"] = findOrGenerateCyan(input.Accents, scheme)
        ansi["term7"] = generateWhite(scheme.Foreground)
    }
    
    // Generate bright variants
    for i := 0; i < 8; i++ {
        base := ansi[fmt.Sprintf("term%d", i)]
        ansi[fmt.Sprintf("term%d", i+8)] = generateBrightVariant(base, scheme.Mode == "dark")
    }
    
    return ansi
}

func findOrGenerateRed(accents []string, scheme *ColorScheme) string {
    // Look for red-ish color in accents
    for _, color := range accents {
        hsl := hexToHSL(color)
        if (hsl.H >= 340 || hsl.H <= 20) && hsl.S > 40 {
            return color
        }
    }
    
    // Generate from primary
    primaryHSL := hexToHSL(scheme.Primary)
    redHSL := HSL{
        H: 0, // Pure red
        S: primaryHSL.S,
        L: primaryHSL.L,
    }
    return hslToHex(redHSL)
}

func generateBrightVariant(base string, isDark bool) string {
    if isDark {
        // Lighten for dark themes
        return adjustLightness(base, 15)
    } else {
        // Darken for light themes
        return adjustLightness(base, -15)
    }
}
```

## Complete Implementation Code

### Main Generator Function
```go
func GenerateHeimdallScheme(input ColorInput) (*ColorScheme, error) {
    // Validate input
    if err := ValidateInput(input); err != nil {
        return nil, err
    }
    
    scheme := &ColorScheme{
        Background: input.Background,
        Foreground: input.Foreground,
        Text:       input.Foreground,
    }
    
    isDark := input.Mode == "dark" || isColorDark(input.Background)
    
    // Step 1: Generate Material Design colors
    var primaryColor string
    if len(input.Accents) > 0 {
        primaryColor = input.Accents[0]
    } else {
        primaryColor = generatePrimaryFromForeground(input.Foreground, input.Background)
    }
    
    materialPrimary := generateMaterialPrimary(primaryColor, input.Background, isDark)
    applyMaterialColors(scheme, "primary", materialPrimary)
    
    materialSecondary := generateMaterialSecondary(primaryColor, input.Background, isDark)
    applyMaterialColors(scheme, "secondary", materialSecondary)
    
    materialTertiary := generateMaterialTertiary(primaryColor, input.Background, isDark)
    applyMaterialColors(scheme, "tertiary", materialTertiary)
    
    // Step 2: Generate surface hierarchy
    var surfaces map[string]string
    if isDark {
        surfaces = generateDarkSurfaces(input.Background)
    } else {
        surfaces = generateLightSurfaces(input.Background)
    }
    applySurfaces(scheme, surfaces)
    
    // Step 3: Generate ANSI colors
    ansiColors := generateANSIColors(input, scheme)
    applyANSIColors(scheme, ansiColors)
    
    // Step 4: Generate semantic colors
    scheme.Error = scheme.Term1 // Use red
    scheme.OnError = ensureContrast(scheme.Error, scheme.Background, 4.5)
    scheme.ErrorContainer = mixColors(scheme.Error, scheme.Background, 0.7)
    scheme.OnErrorContainer = ensureContrast(scheme.ErrorContainer, scheme.Error, 3.0)
    
    scheme.Success = scheme.Term2 // Use green
    scheme.OnSuccess = ensureContrast(scheme.Success, scheme.Background, 4.5)
    scheme.SuccessContainer = mixColors(scheme.Success, scheme.Background, 0.7)
    scheme.OnSuccessContainer = ensureContrast(scheme.SuccessContainer, scheme.Success, 3.0)
    
    // Step 5: Generate additional colors
    scheme.Outline = mixColors(scheme.Foreground, scheme.Background, 0.3)
    scheme.OutlineVariant = mixColors(scheme.Foreground, scheme.Background, 0.2)
    scheme.Shadow = "#000000"
    scheme.Scrim = "#000000"
    
    scheme.InverseSurface = scheme.Foreground
    scheme.InverseOnSurface = scheme.Background
    scheme.InversePrimary = adjustLightness(primaryColor, isDark ? -30 : 30)
    
    // Step 6: Generate theme-specific colors
    scheme.Base = scheme.Background
    scheme.Mantle = adjustLightness(scheme.Background, isDark ? -3 : 3)
    scheme.Crust = adjustLightness(scheme.Background, isDark ? -6 : 6)
    
    scheme.Overlay0 = mixColors(scheme.Foreground, scheme.Background, 0.4)
    scheme.Overlay1 = mixColors(scheme.Foreground, scheme.Background, 0.5)
    scheme.Overlay2 = mixColors(scheme.Foreground, scheme.Background, 0.6)
    
    scheme.Subtext0 = mixColors(scheme.Foreground, scheme.Background, 0.65)
    scheme.Subtext1 = mixColors(scheme.Foreground, scheme.Background, 0.75)
    
    scheme.Surface0 = surfaces["surfaceContainerLow"]
    scheme.Surface1 = surfaces["surfaceContainer"]
    scheme.Surface2 = surfaces["surfaceContainerHigh"]
    
    // Step 7: Validate all contrasts
    if err := validateSchemeContrasts(scheme); err != nil {
        return nil, err
    }
    
    return scheme, nil
}
```

### Example Usage
```go
func ExampleMinimalInput() {
    input := ColorInput{
        Background: "#212337",
        Foreground: "#ebfafa",
        Accents: []string{
            "#a48cf2", // Primary (blue-purple)
            "#37f499", // Secondary (green)
            "#f1fc79", // Tertiary (yellow)
        },
        Mode: "dark",
    }
    
    scheme, err := GenerateHeimdallScheme(input)
    if err != nil {
        log.Fatal(err)
    }
    
    // Result: Complete 122-color Heimdall scheme
    fmt.Printf("Generated %d colors\n", countColors(scheme))
}

func ExampleTerminalInput() {
    input := ColorInput{
        Background: "#212337",
        Foreground: "#ebfafa",
        Terminal: []string{
            "#212337", // Black
            "#f16c75", // Red
            "#37f499", // Green
            "#f1fc79", // Yellow
            "#a48cf2", // Blue
            "#f265b5", // Magenta
            "#04d1f9", // Cyan
            "#ebfafa", // White
        },
        Mode: "dark",
    }
    
    scheme, err := GenerateHeimdallScheme(input)
    if err != nil {
        log.Fatal(err)
    }
    
    // Uses terminal colors to derive Material Design colors
}
```

## Validation and Testing

### Contrast Validation
```go
func validateSchemeContrasts(scheme *ColorScheme) error {
    tests := []struct {
        bg       string
        fg       string
        minRatio float64
        name     string
    }{
        {scheme.Background, scheme.Foreground, 7.0, "background-foreground"},
        {scheme.Primary, scheme.OnPrimary, 4.5, "primary-onPrimary"},
        {scheme.Surface, scheme.OnSurface, 4.5, "surface-onSurface"},
        {scheme.Error, scheme.OnError, 4.5, "error-onError"},
    }
    
    for _, test := range tests {
        ratio := calculateContrast(test.bg, test.fg)
        if ratio < test.minRatio {
            return fmt.Errorf("%s contrast %.2f below minimum %.2f",
                test.name, ratio, test.minRatio)
        }
    }
    
    return nil
}
```

### Visual Harmony Testing
```go
func testColorHarmony(scheme *ColorScheme) float64 {
    // Calculate color harmony score
    primary := hexToHSL(scheme.Primary)
    secondary := hexToHSL(scheme.Secondary)
    tertiary := hexToHSL(scheme.Tertiary)
    
    // Check for complementary/triadic relationships
    hueDiff1 := math.Abs(secondary.H - primary.H)
    hueDiff2 := math.Abs(tertiary.H - primary.H)
    
    // Ideal: 120° for triadic, 180° for complementary
    triadicScore := 1.0 - math.Abs(hueDiff1-120)/180
    
    // Check saturation consistency
    saturationVariance := calculateVariance([]float64{
        primary.S, secondary.S, tertiary.S,
    })
    saturationScore := 1.0 - (saturationVariance / 100)
    
    return (triadicScore + saturationScore) / 2
}
```

## Integration Points

### File Structure
```
internal/
  scheme/
    generator/
      generator.go       # Main generation logic
      color_math.go      # Color space conversions
      material.go        # Material Design generation
      ansi.go           # ANSI color mapping
      validator.go      # Contrast and harmony validation
    generator_test.go   # Comprehensive tests
```

### Configuration Requirements
```json
{
  "generator": {
    "minContrast": {
      "background": 7.0,
      "surface": 4.5,
      "container": 3.0
    },
    "defaults": {
      "mode": "dark",
      "temperature": "neutral"
    }
  }
}
```

## Data Structures

### Source Field
The `source` field indicates where a colorscheme originates from:
- `"bundled"` - Schemes that are embedded in the heimdall binary at compile time
- `"user"` - Custom schemes created by users and stored in `~/.config/heimdall/schemes/`
- `"generated"` - Schemes dynamically generated from wallpapers or other sources, stored in `~/.local/share/heimdall/schemes/`

This field is automatically injected if missing to ensure consistent metadata across all scheme sources.

### Complete Heimdall Output Format
```json
{
  "name": "generated",
  "flavour": "custom",
  "mode": "dark",
  "source": "generated",
  "colours": {
    "background": "#212337",
    "foreground": "#ebfafa",
    "text": "#ebfafa",
    
    "primary": "#a48cf2",
    "onPrimary": "#2a2640",
    "primaryContainer": "#6b5a9c",
    "onPrimaryContainer": "#e2d9f4",
    
    "secondary": "#37f499",
    "onSecondary": "#0a3d26",
    "secondaryContainer": "#26a366",
    "onSecondaryContainer": "#c8f7e0",
    
    "tertiary": "#f1fc79",
    "onTertiary": "#3d3f1f",
    "tertiaryContainer": "#a6ab56",
    "onTertiaryContainer": "#f8fbd9",
    
    "surface": "#212337",
    "onSurface": "#ebfafa",
    "surfaceContainerLowest": "#1d1f32",
    "surfaceContainerLow": "#26283d",
    "surfaceContainer": "#292b42",
    "surfaceContainerHigh": "#2e3048",
    "surfaceContainerHighest": "#34364f",
    
    "term0": "#212337",
    "term1": "#f16c75",
    "term2": "#37f499",
    "term3": "#f1fc79",
    "term4": "#a48cf2",
    "term5": "#f265b5",
    "term6": "#04d1f9",
    "term7": "#ebfafa",
    "term8": "#3a3c52",
    "term9": "#f4868d",
    "term10": "#5af6aa",
    "term11": "#f4fd95",
    "term12": "#b5a1f5",
    "term13": "#f580c3",
    "term14": "#36d8fb",
    "term15": "#ffffff",
    
    "error": "#f16c75",
    "success": "#37f499",
    "outline": "#6b6d82",
    "shadow": "#000000",
    
    "base": "#212337",
    "mantle": "#1d1f32",
    "crust": "#191b2d",
    "overlay0": "#7a7c8f",
    "overlay1": "#9597a6",
    "overlay2": "#b0b2bd",
    "subtext0": "#c5c7ce",
    "subtext1": "#d8dae1"
  }
}
```

## Validation Checklist

- [ ] All 122 color keys present
- [ ] Source field present and valid ("bundled", "user", or "generated")
- [ ] Background-foreground contrast ≥ 7:1
- [ ] Primary-onPrimary contrast ≥ 4.5:1
- [ ] All container-onContainer contrasts ≥ 3:1
- [ ] Surface hierarchy properly ordered
- [ ] ANSI colors semantically correct
- [ ] Terminal colors visible against background
- [ ] Material Design tokens properly generated
- [ ] Theme-specific colors included
- [ ] All hex values valid format with # prefix
- [ ] Mode correctly identified (light/dark)
- [ ] Color temperature preserved in conversions

## Code Examples

### Complete Generator Implementation
```go
package generator

import (
    "fmt"
    "math"
    "encoding/json"
)

// GenerateFromMinimal creates a full scheme from minimal input
func GenerateFromMinimal(bg, fg string, accent string) (*Scheme, error) {
    input := ColorInput{
        Background: bg,
        Foreground: fg,
        Accents:    []string{accent},
        Mode:       detectMode(bg),
    }
    
    colorScheme, err := GenerateHeimdallScheme(input)
    if err != nil {
        return nil, err
    }
    
    return convertToSchemeFormat(colorScheme), nil
}

// GenerateFromScreenshot extracts colors from an image
func GenerateFromScreenshot(imagePath string) (*Scheme, error) {
    // Extract dominant colors
    colors, err := extractDominantColors(imagePath, 5)
    if err != nil {
        return nil, err
    }
    
    // Identify background (usually most dominant)
    bg := colors[0]
    
    // Find best foreground (highest contrast)
    fg := findBestForeground(bg, colors[1:])
    
    // Use remaining as accents
    input := ColorInput{
        Background: bg,
        Foreground: fg,
        Accents:    colors[1:],
        Mode:       detectMode(bg),
    }
    
    return GenerateHeimdallScheme(input)
}

// ConvertMode converts between light and dark modes
func ConvertMode(scheme *Scheme, targetMode string) (*Scheme, error) {
    colorScheme := parseSchemeToColorScheme(scheme)
    converted := convertThemeMode(colorScheme, targetMode)
    return convertToSchemeFormat(converted), nil
}
```

## References

- Material Design 3 Color System: https://m3.material.io/styles/color
- WCAG Contrast Guidelines: https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum
- LAB Color Space: https://en.wikipedia.org/wiki/CIELAB_color_space
- HSL Color Model: https://en.wikipedia.org/wiki/HSL_and_HSV
- Color Harmony Theory: https://www.colormatters.com/color-and-design/basic-color-theory