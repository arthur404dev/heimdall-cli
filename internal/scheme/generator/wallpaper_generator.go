package generator

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/scheme"
	"github.com/arthur404dev/heimdall-cli/internal/utils/material"
)

// WallpaperGenerator creates complete Heimdall colorschemes from wallpapers
type WallpaperGenerator struct {
	materialGen       *material.Generator
	enhancedExtractor *material.EnhancedExtractor
}

// NewWallpaperGenerator creates a new wallpaper-based scheme generator
func NewWallpaperGenerator() *WallpaperGenerator {
	return &WallpaperGenerator{
		materialGen:       material.NewGenerator(),
		enhancedExtractor: material.NewEnhancedExtractor(),
	}
}

// GenerateFullScheme creates a complete 122-color Heimdall scheme from a Material You scheme
func (g *WallpaperGenerator) GenerateFullScheme(
	materialScheme *material.Scheme,
	wallpaperPath string,
	mode string,
) (*scheme.Scheme, error) {
	// Initialize the scheme
	heimdallScheme := &scheme.Scheme{
		Name:    "material-you",
		Flavour: "wallpaper",
		Mode:    mode,
		Variant: "dynamic",
		Colours: make(map[string]string),
	}

	isDark := mode == "dark"

	// 1. Convert base Material Design colors
	g.addMaterialColors(heimdallScheme, materialScheme)

	// 2. Generate Material Design Fixed variants
	g.addMaterialFixedVariants(heimdallScheme, materialScheme, isDark)

	// 3. Generate ANSI terminal colors
	g.addANSIColors(heimdallScheme, materialScheme, isDark)

	// 4. Generate surface hierarchy (12 levels)
	g.addSurfaceHierarchy(heimdallScheme, materialScheme, isDark)

	// 5. Generate semantic colors (success, warning)
	g.addSemanticColors(heimdallScheme, materialScheme, isDark)

	// 6. Generate theme-specific colors (base, mantle, crust, overlays, subtexts)
	g.addThemeSpecificColors(heimdallScheme, materialScheme, isDark)

	// 7. Generate Catppuccin compatibility colors
	g.addCatppuccinColors(heimdallScheme, materialScheme, isDark)

	// 8. Add duplicate keys for compatibility
	g.addCompatibilityKeys(heimdallScheme)

	// Validate we have all required colors
	if len(heimdallScheme.Colours) < 122 {
		return nil, fmt.Errorf("incomplete scheme generation: only %d colors generated (expected 122)", len(heimdallScheme.Colours))
	}

	return heimdallScheme, nil
}

// addMaterialColors converts base Material Design colors
func (g *WallpaperGenerator) addMaterialColors(scheme *scheme.Scheme, ms *material.Scheme) {
	// Primary colors
	scheme.Colours["primary"] = argbToHex(ms.Primary)
	scheme.Colours["onPrimary"] = argbToHex(ms.OnPrimary)
	scheme.Colours["primaryContainer"] = argbToHex(ms.PrimaryContainer)
	scheme.Colours["onPrimaryContainer"] = argbToHex(ms.OnPrimaryContainer)

	// Secondary colors
	scheme.Colours["secondary"] = argbToHex(ms.Secondary)
	scheme.Colours["onSecondary"] = argbToHex(ms.OnSecondary)
	scheme.Colours["secondaryContainer"] = argbToHex(ms.SecondaryContainer)
	scheme.Colours["onSecondaryContainer"] = argbToHex(ms.OnSecondaryContainer)

	// Tertiary colors
	scheme.Colours["tertiary"] = argbToHex(ms.Tertiary)
	scheme.Colours["onTertiary"] = argbToHex(ms.OnTertiary)
	scheme.Colours["tertiaryContainer"] = argbToHex(ms.TertiaryContainer)
	scheme.Colours["onTertiaryContainer"] = argbToHex(ms.OnTertiaryContainer)

	// Error colors
	scheme.Colours["error"] = argbToHex(ms.Error)
	scheme.Colours["onError"] = argbToHex(ms.OnError)
	scheme.Colours["errorContainer"] = argbToHex(ms.ErrorContainer)
	scheme.Colours["onErrorContainer"] = argbToHex(ms.OnErrorContainer)

	// Background and surface
	scheme.Colours["background"] = argbToHex(ms.Background)
	scheme.Colours["onBackground"] = argbToHex(ms.OnBackground)
	scheme.Colours["surface"] = argbToHex(ms.Surface)
	scheme.Colours["onSurface"] = argbToHex(ms.OnSurface)
	scheme.Colours["surfaceVariant"] = argbToHex(ms.SurfaceVariant)
	scheme.Colours["onSurfaceVariant"] = argbToHex(ms.OnSurfaceVariant)

	// Outline colors
	scheme.Colours["outline"] = argbToHex(ms.Outline)
	scheme.Colours["outlineVariant"] = argbToHex(ms.OutlineVariant)

	// Other colors
	scheme.Colours["shadow"] = argbToHex(ms.Shadow)
	scheme.Colours["scrim"] = argbToHex(ms.Scrim)
	scheme.Colours["inverseSurface"] = argbToHex(ms.InverseSurface)
	scheme.Colours["inverseOnSurface"] = argbToHex(ms.InverseOnSurface)
	scheme.Colours["inversePrimary"] = argbToHex(ms.InversePrimary)

	// Add text and foreground aliases
	scheme.Colours["text"] = argbToHex(ms.OnBackground)
	scheme.Colours["foreground"] = argbToHex(ms.OnBackground)
}

// addMaterialFixedVariants generates Material Design 3 fixed color variants
func (g *WallpaperGenerator) addMaterialFixedVariants(scheme *scheme.Scheme, ms *material.Scheme, isDark bool) {
	primary := scheme.Colours["primary"]
	secondary := scheme.Colours["secondary"]
	tertiary := scheme.Colours["tertiary"]

	if isDark {
		// Dark mode fixed variants
		scheme.Colours["primaryFixed"] = adjustLightness(primary, 15)
		scheme.Colours["primaryFixedDim"] = primary
		scheme.Colours["onPrimaryFixed"] = adjustLightness(primary, -40)
		scheme.Colours["onPrimaryFixedVariant"] = adjustLightness(primary, -20)

		scheme.Colours["secondaryFixed"] = adjustLightness(secondary, 15)
		scheme.Colours["secondaryFixedDim"] = secondary
		scheme.Colours["onSecondaryFixed"] = adjustLightness(secondary, -40)
		scheme.Colours["onSecondaryFixedVariant"] = adjustLightness(secondary, -20)

		scheme.Colours["tertiaryFixed"] = adjustLightness(tertiary, 15)
		scheme.Colours["tertiaryFixedDim"] = tertiary
		scheme.Colours["onTertiaryFixed"] = adjustLightness(tertiary, -40)
		scheme.Colours["onTertiaryFixedVariant"] = adjustLightness(tertiary, -20)
	} else {
		// Light mode fixed variants
		scheme.Colours["primaryFixed"] = adjustLightness(primary, -10)
		scheme.Colours["primaryFixedDim"] = adjustLightness(primary, -5)
		scheme.Colours["onPrimaryFixed"] = adjustLightness(primary, 40)
		scheme.Colours["onPrimaryFixedVariant"] = adjustLightness(primary, 20)

		scheme.Colours["secondaryFixed"] = adjustLightness(secondary, -10)
		scheme.Colours["secondaryFixedDim"] = adjustLightness(secondary, -5)
		scheme.Colours["onSecondaryFixed"] = adjustLightness(secondary, 40)
		scheme.Colours["onSecondaryFixedVariant"] = adjustLightness(secondary, 20)

		scheme.Colours["tertiaryFixed"] = adjustLightness(tertiary, -10)
		scheme.Colours["tertiaryFixedDim"] = adjustLightness(tertiary, -5)
		scheme.Colours["onTertiaryFixed"] = adjustLightness(tertiary, 40)
		scheme.Colours["onTertiaryFixedVariant"] = adjustLightness(tertiary, 20)
	}

	// Add palette key colors
	scheme.Colours["primary_paletteKeyColor"] = scheme.Colours["primaryContainer"]
	scheme.Colours["secondary_paletteKeyColor"] = scheme.Colours["secondaryContainer"]
	scheme.Colours["tertiary_paletteKeyColor"] = scheme.Colours["tertiaryContainer"]
	scheme.Colours["neutral_paletteKeyColor"] = adjustLightness(scheme.Colours["onSurface"], -30)
	scheme.Colours["neutral_variant_paletteKeyColor"] = adjustLightness(scheme.Colours["onSurfaceVariant"], -35)
}

// addANSIColors generates ANSI terminal colors from Material palette
func (g *WallpaperGenerator) addANSIColors(scheme *scheme.Scheme, ms *material.Scheme, isDark bool) {
	// Generate semantically correct ANSI colors
	background := scheme.Colours["background"]
	foreground := scheme.Colours["foreground"]
	primary := scheme.Colours["primary"]
	secondary := scheme.Colours["secondary"]
	tertiary := scheme.Colours["tertiary"]
	error := scheme.Colours["error"]

	// Normal colors (0-7)
	blackAdjust := float64(-10)
	if !isDarkColor(background) {
		blackAdjust = 10
	}
	scheme.Colours["term0"] = adjustLightness(background, blackAdjust) // Black
	scheme.Colours["term1"] = error                                    // Red
	scheme.Colours["term2"] = generateGreen(primary, isDark)           // Green
	scheme.Colours["term3"] = generateYellow(primary, isDark)          // Yellow
	scheme.Colours["term4"] = primary                                  // Blue
	scheme.Colours["term5"] = secondary                                // Magenta
	scheme.Colours["term6"] = tertiary                                 // Cyan
	whiteAdjust := float64(-10)
	if !isDark {
		whiteAdjust = 10
	}
	scheme.Colours["term7"] = adjustLightness(foreground, whiteAdjust) // White

	// Bright colors (8-15)
	for i := 0; i < 8; i++ {
		base := scheme.Colours[fmt.Sprintf("term%d", i)]
		if isDark {
			scheme.Colours[fmt.Sprintf("term%d", i+8)] = adjustLightness(base, 15)
		} else {
			scheme.Colours[fmt.Sprintf("term%d", i+8)] = adjustLightness(base, -15)
		}
	}

	// Also add color0-15 aliases for compatibility
	for i := 0; i < 16; i++ {
		scheme.Colours[fmt.Sprintf("color%d", i)] = scheme.Colours[fmt.Sprintf("term%d", i)]
		scheme.Colours[fmt.Sprintf("colour%d", i)] = scheme.Colours[fmt.Sprintf("term%d", i)]
	}
}

// addSurfaceHierarchy generates the complete surface elevation hierarchy
func (g *WallpaperGenerator) addSurfaceHierarchy(scheme *scheme.Scheme, ms *material.Scheme, isDark bool) {
	background := scheme.Colours["background"]
	surface := scheme.Colours["surface"]

	if isDark {
		// Dark mode: progressive lightening for elevation
		scheme.Colours["surfaceDim"] = adjustLightness(surface, -5)
		scheme.Colours["surfaceBright"] = adjustLightness(surface, 15)
		scheme.Colours["surfaceContainerLowest"] = adjustLightness(background, -2)
		scheme.Colours["surfaceContainerLow"] = adjustLightness(background, 3)
		scheme.Colours["surfaceContainer"] = adjustLightness(background, 5)
		scheme.Colours["surfaceContainerHigh"] = adjustLightness(background, 8)
		scheme.Colours["surfaceContainerHighest"] = adjustLightness(background, 12)
		scheme.Colours["surfaceTint"] = scheme.Colours["primary"]

		// Additional surface levels
		scheme.Colours["surface0"] = adjustLightness(background, 5)
		scheme.Colours["surface1"] = adjustLightness(background, 10)
		scheme.Colours["surface2"] = adjustLightness(background, 15)
	} else {
		// Light mode: subtle variations
		scheme.Colours["surfaceDim"] = adjustLightness(surface, -8)
		scheme.Colours["surfaceBright"] = adjustLightness(surface, 5)
		scheme.Colours["surfaceContainerLowest"] = adjustLightness(background, -3)
		scheme.Colours["surfaceContainerLow"] = background
		scheme.Colours["surfaceContainer"] = background
		scheme.Colours["surfaceContainerHigh"] = adjustLightness(background, 1)
		scheme.Colours["surfaceContainerHighest"] = adjustLightness(background, 2)
		scheme.Colours["surfaceTint"] = scheme.Colours["primary"]

		// Additional surface levels
		scheme.Colours["surface0"] = adjustLightness(background, -2)
		scheme.Colours["surface1"] = adjustLightness(background, -4)
		scheme.Colours["surface2"] = adjustLightness(background, -6)
	}
}

// addSemanticColors generates semantic colors (success, warning)
func (g *WallpaperGenerator) addSemanticColors(scheme *scheme.Scheme, ms *material.Scheme, isDark bool) {
	primary := scheme.Colours["primary"]
	background := scheme.Colours["background"]

	// Success colors (green-based)
	successBase := generateGreen(primary, isDark)
	scheme.Colours["success"] = successBase
	successContrast := "#ffffff"
	if isDark {
		successContrast = "#000000"
	}
	scheme.Colours["onSuccess"] = ensureContrast(successBase, successContrast, 4.5)
	scheme.Colours["successContainer"] = mixColors(successBase, background, 0.7)
	scheme.Colours["onSuccessContainer"] = ensureContrast(scheme.Colours["successContainer"], successBase, 3.0)

	// Warning colors (yellow/orange-based) - not in standard Material but useful
	warningBase := generateYellow(primary, isDark)
	scheme.Colours["warning"] = warningBase
	warningContrast := "#ffffff"
	if isDark {
		warningContrast = "#000000"
	}
	scheme.Colours["onWarning"] = ensureContrast(warningBase, warningContrast, 4.5)
	scheme.Colours["warningContainer"] = mixColors(warningBase, background, 0.7)
	scheme.Colours["onWarningContainer"] = ensureContrast(scheme.Colours["warningContainer"], warningBase, 3.0)
}

// addThemeSpecificColors adds theme-specific colors (base, mantle, crust, overlays, subtexts)
func (g *WallpaperGenerator) addThemeSpecificColors(scheme *scheme.Scheme, ms *material.Scheme, isDark bool) {
	background := scheme.Colours["background"]
	foreground := scheme.Colours["foreground"]

	// Base colors
	scheme.Colours["base"] = background
	mantleAdjust := float64(3)
	crustAdjust := float64(6)
	if isDark {
		mantleAdjust = -3
		crustAdjust = -6
	}
	scheme.Colours["mantle"] = adjustLightness(background, mantleAdjust)
	scheme.Colours["crust"] = adjustLightness(background, crustAdjust)

	// Overlay colors (progressive mixing with foreground)
	scheme.Colours["overlay0"] = mixColors(foreground, background, 0.4)
	scheme.Colours["overlay1"] = mixColors(foreground, background, 0.5)
	scheme.Colours["overlay2"] = mixColors(foreground, background, 0.6)

	// Subtext colors
	scheme.Colours["subtext0"] = mixColors(foreground, background, 0.65)
	scheme.Colours["subtext1"] = mixColors(foreground, background, 0.75)
}

// addCatppuccinColors adds Catppuccin-style named colors for compatibility
func (g *WallpaperGenerator) addCatppuccinColors(scheme *scheme.Scheme, ms *material.Scheme, isDark bool) {
	// Use existing colors to generate Catppuccin names
	scheme.Colours["rosewater"] = adjustLightness(scheme.Colours["tertiary"], 20)
	scheme.Colours["flamingo"] = mixColors(scheme.Colours["tertiary"], scheme.Colours["secondary"], 0.5)
	scheme.Colours["pink"] = scheme.Colours["tertiary"]
	scheme.Colours["mauve"] = mixColors(scheme.Colours["primary"], scheme.Colours["tertiary"], 0.5)
	scheme.Colours["red"] = scheme.Colours["error"]
	scheme.Colours["maroon"] = adjustLightness(scheme.Colours["error"], -10)
	scheme.Colours["peach"] = generateOrange(scheme.Colours["primary"], scheme.Colours["error"])
	scheme.Colours["yellow"] = scheme.Colours["term3"]
	scheme.Colours["green"] = scheme.Colours["term2"]
	scheme.Colours["teal"] = scheme.Colours["term6"]
	scheme.Colours["sky"] = adjustLightness(scheme.Colours["term6"], 10)
	scheme.Colours["sapphire"] = mixColors(scheme.Colours["term4"], scheme.Colours["term6"], 0.5)
	scheme.Colours["blue"] = scheme.Colours["primary"]
	scheme.Colours["lavender"] = adjustLightness(scheme.Colours["primary"], 10)
}

// addCompatibilityKeys adds duplicate keys for compatibility with different naming conventions
func (g *WallpaperGenerator) addCompatibilityKeys(scheme *scheme.Scheme) {
	// Add cursor color if not present
	if _, exists := scheme.Colours["cursor"]; !exists {
		scheme.Colours["cursor"] = scheme.Colours["primary"]
	}

	// Ensure both on_* and on* formats exist
	keysToCheck := []string{"primary", "secondary", "tertiary", "error", "success", "warning",
		"surface", "background", "surfaceVariant", "primaryContainer", "secondaryContainer",
		"tertiaryContainer", "errorContainer", "successContainer", "warningContainer",
		"primaryFixed", "secondaryFixed", "tertiaryFixed", "primaryFixedVariant",
		"secondaryFixedVariant", "tertiaryFixedVariant"}

	for _, key := range keysToCheck {
		// Check for on* version
		onKey := "on" + strings.Title(key)
		onKeySnake := "on_" + toSnakeCase(key)

		if val, exists := scheme.Colours[onKey]; exists {
			scheme.Colours[onKeySnake] = val
		} else if val, exists := scheme.Colours[onKeySnake]; exists {
			scheme.Colours[onKey] = val
		}
	}
}

// Helper functions

func argbToHex(argb uint32) string {
	r := (argb >> 16) & 0xFF
	g := (argb >> 8) & 0xFF
	b := argb & 0xFF
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func hexToRGB(hex string) (r, g, b uint8) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0
	}

	var rgb uint32
	fmt.Sscanf(hex, "%06x", &rgb)
	r = uint8((rgb >> 16) & 0xFF)
	g = uint8((rgb >> 8) & 0xFF)
	b = uint8(rgb & 0xFF)
	return
}

func rgbToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// HSL represents a color in HSL space
type HSL struct {
	H float64 // Hue: 0-360
	S float64 // Saturation: 0-100
	L float64 // Lightness: 0-100
}

func hexToHSL(hex string) HSL {
	r, g, b := hexToRGB(hex)

	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(math.Max(rf, gf), bf)
	min := math.Min(math.Min(rf, gf), bf)
	delta := max - min

	l := (max + min) / 2.0

	if delta == 0 {
		return HSL{0, 0, l * 100}
	}

	var s float64
	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2.0 - max - min)
	}

	var h float64
	switch max {
	case rf:
		h = ((gf - bf) / delta)
		if gf < bf {
			h += 6
		}
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

	return rgbToHex(uint8(r*255), uint8(g*255), uint8(b*255))
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

func adjustLightness(hex string, amount float64) string {
	hsl := hexToHSL(hex)
	hsl.L = math.Max(0, math.Min(100, hsl.L+amount))
	return hslToHex(hsl)
}

func adjustSaturation(hex string, amount float64) string {
	hsl := hexToHSL(hex)
	hsl.S = math.Max(0, math.Min(100, hsl.S+amount))
	return hslToHex(hsl)
}

func mixColors(hex1, hex2 string, ratio float64) string {
	r1, g1, b1 := hexToRGB(hex1)
	r2, g2, b2 := hexToRGB(hex2)

	r := uint8(float64(r1)*(1-ratio) + float64(r2)*ratio)
	g := uint8(float64(g1)*(1-ratio) + float64(g2)*ratio)
	b := uint8(float64(b1)*(1-ratio) + float64(b2)*ratio)

	return rgbToHex(r, g, b)
}

func isDarkColor(hex string) bool {
	r, g, b := hexToRGB(hex)
	luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return luminance < 128
}

func relativeLuminance(hex string) float64 {
	r, g, b := hexToRGB(hex)

	rf := gammaCorrect(float64(r) / 255.0)
	gf := gammaCorrect(float64(g) / 255.0)
	bf := gammaCorrect(float64(b) / 255.0)

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

func generateGreen(primary string, isDark bool) string {
	primaryHSL := hexToHSL(primary)
	greenHSL := HSL{
		H: 120, // Green hue
		S: primaryHSL.S,
		L: primaryHSL.L,
	}

	// Adjust for visibility
	if isDark && greenHSL.L < 50 {
		greenHSL.L = 50
	} else if !isDark && greenHSL.L > 50 {
		greenHSL.L = 40
	}

	return hslToHex(greenHSL)
}

func generateYellow(primary string, isDark bool) string {
	primaryHSL := hexToHSL(primary)
	yellowHSL := HSL{
		H: 60, // Yellow hue
		S: primaryHSL.S,
		L: primaryHSL.L,
	}

	// Yellow needs special handling for visibility
	if isDark {
		yellowHSL.L = math.Max(60, yellowHSL.L)
	} else {
		yellowHSL.L = math.Min(40, yellowHSL.L)
	}

	return hslToHex(yellowHSL)
}

func generateOrange(primary, error string) string {
	primaryHSL := hexToHSL(primary)
	errorHSL := hexToHSL(error)

	orangeHSL := HSL{
		H: 30, // Orange hue
		S: (primaryHSL.S + errorHSL.S) / 2,
		L: (primaryHSL.L + errorHSL.L) / 2,
	}

	return hslToHex(orangeHSL)
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// MaterialYouVariant represents a Material You color variant
type MaterialYouVariant string

const (
	VariantVibrant    MaterialYouVariant = "vibrant"
	VariantTonal      MaterialYouVariant = "tonal"
	VariantExpressive MaterialYouVariant = "expressive"
	VariantFidelity   MaterialYouVariant = "fidelity"
	VariantContent    MaterialYouVariant = "content"
	VariantFruitSalad MaterialYouVariant = "fruit_salad"
	VariantRainbow    MaterialYouVariant = "rainbow"
	VariantNeutral    MaterialYouVariant = "neutral"
)

// GenerateAllVariants generates all Material You variants from a wallpaper
func (g *WallpaperGenerator) GenerateAllVariants(img image.Image, wallpaperPath string) (map[string]*scheme.Scheme, error) {
	// Extract colors using enhanced extractor
	extractedColors, err := g.enhancedExtractor.ExtractColors(img)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colors: %w", err)
	}

	// Get best seed color
	seedColor := extractedColors.GetBestSeedColor()

	// Generate all variants
	variants := make(map[string]*scheme.Scheme)
	variantTypes := []MaterialYouVariant{
		VariantVibrant, VariantTonal, VariantExpressive, VariantFidelity,
		VariantContent, VariantFruitSalad, VariantRainbow, VariantNeutral,
	}

	for _, variant := range variantTypes {
		// Generate both dark and light modes for each variant
		for _, isDark := range []bool{true, false} {
			mode := "light"
			if isDark {
				mode = "dark"
			}

			// Generate Material scheme for this variant
			materialScheme, err := g.generateVariantScheme(seedColor, extractedColors, variant, isDark)
			if err != nil {
				return nil, fmt.Errorf("failed to generate %s %s variant: %w", variant, mode, err)
			}

			// Convert to full Heimdall scheme
			heimdallScheme, err := g.GenerateFullScheme(materialScheme, wallpaperPath, mode)
			if err != nil {
				return nil, fmt.Errorf("failed to convert %s %s variant: %w", variant, mode, err)
			}

			// Update variant name
			heimdallScheme.Variant = string(variant)

			// Store with key: variant/mode
			key := fmt.Sprintf("%s/%s", variant, mode)
			variants[key] = heimdallScheme
		}
	}

	return variants, nil
}

// generateVariantScheme generates a specific Material You variant
func (g *WallpaperGenerator) generateVariantScheme(
	seedColor uint32,
	extractedColors *material.ExtractedColors,
	variant MaterialYouVariant,
	isDark bool,
) (*material.Scheme, error) {
	// Adjust seed color based on variant
	adjustedSeed := g.adjustSeedForVariant(seedColor, extractedColors, variant)

	// Generate base palette
	palette, err := g.materialGen.GenerateFromColor(adjustedSeed)
	if err != nil {
		return nil, err
	}

	// Apply variant-specific modifications
	g.applyVariantModifications(palette, variant, extractedColors)

	// Generate scheme with variant-specific tone mappings
	scheme := g.generateSchemeWithVariantTones(palette, variant, isDark)

	return scheme, nil
}

// adjustSeedForVariant adjusts the seed color based on the variant type
func (g *WallpaperGenerator) adjustSeedForVariant(
	seedColor uint32,
	extractedColors *material.ExtractedColors,
	variant MaterialYouVariant,
) uint32 {
	switch variant {
	case VariantVibrant:
		// Use most vibrant color
		if len(extractedColors.Accents) > 0 {
			return extractedColors.Accents[0].Color
		}
		return g.boostSaturation(seedColor, 1.3)

	case VariantTonal:
		// Use balanced color
		return seedColor

	case VariantExpressive:
		// Use bold, high-contrast color
		if len(extractedColors.EdgeColors) > 0 {
			return extractedColors.EdgeColors[0].Color
		}
		return g.boostSaturation(seedColor, 1.2)

	case VariantFidelity:
		// Use color closest to source
		if len(extractedColors.Dominant) > 0 {
			return extractedColors.Dominant[0].Color
		}
		return seedColor

	case VariantContent:
		// Adaptive based on content
		return seedColor

	case VariantFruitSalad:
		// Use multiple accent colors - rotate hue for variety
		return g.rotateHue(seedColor, 30)

	case VariantRainbow:
		// Full spectrum - shift hue significantly
		return g.rotateHue(seedColor, 60)

	case VariantNeutral:
		// Desaturated version
		return g.desaturate(seedColor, 0.3)

	default:
		return seedColor
	}
}

// applyVariantModifications applies variant-specific modifications to the palette
func (g *WallpaperGenerator) applyVariantModifications(
	palette *material.Palette,
	variant MaterialYouVariant,
	extractedColors *material.ExtractedColors,
) {
	switch variant {
	case VariantVibrant:
		// Boost saturation across all palettes
		g.boostPaletteSaturation(palette, 1.3)

	case VariantExpressive:
		// Increase contrast between tones
		g.increaseTonesContrast(palette, 1.2)

	case VariantFruitSalad:
		// Add more color variety
		g.addColorVariety(palette, extractedColors)

	case VariantRainbow:
		// Spread colors across spectrum
		g.spreadAcrossSpectrum(palette)

	case VariantNeutral:
		// Reduce saturation significantly
		g.reducePaletteSaturation(palette, 0.3)
	}
}

// generateSchemeWithVariantTones generates a scheme with variant-specific tone mappings
func (g *WallpaperGenerator) generateSchemeWithVariantTones(
	palette *material.Palette,
	variant MaterialYouVariant,
	isDark bool,
) *material.Scheme {
	scheme := &material.Scheme{
		Seed:    palette.Seed,
		IsDark:  isDark,
		Palette: palette,
	}

	// Apply variant-specific tone mappings
	switch variant {
	case VariantVibrant:
		g.applyVibrantTones(scheme, palette, isDark)
	case VariantExpressive:
		g.applyExpressiveTones(scheme, palette, isDark)
	case VariantFidelity:
		g.applyFidelityTones(scheme, palette, isDark)
	default:
		g.applyStandardTones(scheme, palette, isDark)
	}

	return scheme
}

// Tone mapping functions for different variants
func (g *WallpaperGenerator) applyVibrantTones(scheme *material.Scheme, palette *material.Palette, isDark bool) {
	if isDark {
		// Vibrant dark: more saturated, brighter accents
		scheme.Primary = palette.Primary.Tone(85)
		scheme.OnPrimary = palette.Primary.Tone(10)
		scheme.PrimaryContainer = palette.Primary.Tone(35)
		scheme.OnPrimaryContainer = palette.Primary.Tone(95)
	} else {
		// Vibrant light: deeper, more saturated colors
		scheme.Primary = palette.Primary.Tone(35)
		scheme.OnPrimary = palette.Primary.Tone(100)
		scheme.PrimaryContainer = palette.Primary.Tone(85)
		scheme.OnPrimaryContainer = palette.Primary.Tone(5)
	}
	// Apply similar adjustments to other color roles
	g.applyStandardTonesForOtherRoles(scheme, palette, isDark)
}

func (g *WallpaperGenerator) applyExpressiveTones(scheme *material.Scheme, palette *material.Palette, isDark bool) {
	if isDark {
		// Expressive dark: high contrast, bold colors
		scheme.Primary = palette.Primary.Tone(90)
		scheme.OnPrimary = palette.Primary.Tone(5)
		scheme.PrimaryContainer = palette.Primary.Tone(40)
		scheme.OnPrimaryContainer = palette.Primary.Tone(100)
	} else {
		// Expressive light: strong, bold colors
		scheme.Primary = palette.Primary.Tone(30)
		scheme.OnPrimary = palette.Primary.Tone(100)
		scheme.PrimaryContainer = palette.Primary.Tone(80)
		scheme.OnPrimaryContainer = palette.Primary.Tone(0)
	}
	g.applyStandardTonesForOtherRoles(scheme, palette, isDark)
}

func (g *WallpaperGenerator) applyFidelityTones(scheme *material.Scheme, palette *material.Palette, isDark bool) {
	// Fidelity: closer to source colors, less transformation
	if isDark {
		scheme.Primary = palette.Primary.Tone(75)
		scheme.OnPrimary = palette.Primary.Tone(25)
		scheme.PrimaryContainer = palette.Primary.Tone(30)
		scheme.OnPrimaryContainer = palette.Primary.Tone(85)
	} else {
		scheme.Primary = palette.Primary.Tone(45)
		scheme.OnPrimary = palette.Primary.Tone(100)
		scheme.PrimaryContainer = palette.Primary.Tone(90)
		scheme.OnPrimaryContainer = palette.Primary.Tone(15)
	}
	g.applyStandardTonesForOtherRoles(scheme, palette, isDark)
}

func (g *WallpaperGenerator) applyStandardTones(scheme *material.Scheme, palette *material.Palette, isDark bool) {
	// Standard Material You tone mappings
	if isDark {
		scheme.Primary = palette.Primary.Tone(80)
		scheme.OnPrimary = palette.Primary.Tone(20)
		scheme.PrimaryContainer = palette.Primary.Tone(30)
		scheme.OnPrimaryContainer = palette.Primary.Tone(90)
	} else {
		scheme.Primary = palette.Primary.Tone(40)
		scheme.OnPrimary = palette.Primary.Tone(100)
		scheme.PrimaryContainer = palette.Primary.Tone(90)
		scheme.OnPrimaryContainer = palette.Primary.Tone(10)
	}
	g.applyStandardTonesForOtherRoles(scheme, palette, isDark)
}

func (g *WallpaperGenerator) applyStandardTonesForOtherRoles(scheme *material.Scheme, palette *material.Palette, isDark bool) {
	// Apply standard tones for secondary, tertiary, error, etc.
	if isDark {
		scheme.Secondary = palette.Secondary.Tone(80)
		scheme.OnSecondary = palette.Secondary.Tone(20)
		scheme.SecondaryContainer = palette.Secondary.Tone(30)
		scheme.OnSecondaryContainer = palette.Secondary.Tone(90)

		scheme.Tertiary = palette.Tertiary.Tone(80)
		scheme.OnTertiary = palette.Tertiary.Tone(20)
		scheme.TertiaryContainer = palette.Tertiary.Tone(30)
		scheme.OnTertiaryContainer = palette.Tertiary.Tone(90)

		scheme.Error = palette.Error.Tone(80)
		scheme.OnError = palette.Error.Tone(20)
		scheme.ErrorContainer = palette.Error.Tone(30)
		scheme.OnErrorContainer = palette.Error.Tone(90)

		scheme.Background = palette.Neutral.Tone(10)
		scheme.OnBackground = palette.Neutral.Tone(90)
		scheme.Surface = palette.Neutral.Tone(10)
		scheme.OnSurface = palette.Neutral.Tone(90)
		scheme.SurfaceVariant = palette.NeutralVariant.Tone(30)
		scheme.OnSurfaceVariant = palette.NeutralVariant.Tone(80)
		scheme.Outline = palette.NeutralVariant.Tone(60)
		scheme.OutlineVariant = palette.NeutralVariant.Tone(30)
	} else {
		scheme.Secondary = palette.Secondary.Tone(40)
		scheme.OnSecondary = palette.Secondary.Tone(100)
		scheme.SecondaryContainer = palette.Secondary.Tone(90)
		scheme.OnSecondaryContainer = palette.Secondary.Tone(10)

		scheme.Tertiary = palette.Tertiary.Tone(40)
		scheme.OnTertiary = palette.Tertiary.Tone(100)
		scheme.TertiaryContainer = palette.Tertiary.Tone(90)
		scheme.OnTertiaryContainer = palette.Tertiary.Tone(10)

		scheme.Error = palette.Error.Tone(40)
		scheme.OnError = palette.Error.Tone(100)
		scheme.ErrorContainer = palette.Error.Tone(90)
		scheme.OnErrorContainer = palette.Error.Tone(10)

		scheme.Background = palette.Neutral.Tone(99)
		scheme.OnBackground = palette.Neutral.Tone(10)
		scheme.Surface = palette.Neutral.Tone(99)
		scheme.OnSurface = palette.Neutral.Tone(10)
		scheme.SurfaceVariant = palette.NeutralVariant.Tone(90)
		scheme.OnSurfaceVariant = palette.NeutralVariant.Tone(30)
		scheme.Outline = palette.NeutralVariant.Tone(50)
		scheme.OutlineVariant = palette.NeutralVariant.Tone(80)
	}

	scheme.Shadow = 0xFF000000
	scheme.Scrim = 0xFF000000
	scheme.InverseSurface = palette.Neutral.Tone(90)
	scheme.InverseOnSurface = palette.Neutral.Tone(20)
	scheme.InversePrimary = palette.Primary.Tone(40)

	if !isDark {
		scheme.InverseSurface = palette.Neutral.Tone(20)
		scheme.InverseOnSurface = palette.Neutral.Tone(95)
		scheme.InversePrimary = palette.Primary.Tone(80)
	}
}

// Helper functions for color manipulation
func (g *WallpaperGenerator) boostSaturation(argb uint32, factor float64) uint32 {
	hex := argbToHex(argb)
	hsl := hexToHSL(hex)
	hsl.S = math.Min(100, hsl.S*factor)
	newHex := hslToHex(hsl)
	r, gr, b := hexToRGB(newHex)
	return 0xFF000000 | uint32(r)<<16 | uint32(gr)<<8 | uint32(b)
}

func (g *WallpaperGenerator) desaturate(argb uint32, factor float64) uint32 {
	hex := argbToHex(argb)
	hsl := hexToHSL(hex)
	hsl.S = hsl.S * factor
	newHex := hslToHex(hsl)
	r, gr, b := hexToRGB(newHex)
	return 0xFF000000 | uint32(r)<<16 | uint32(gr)<<8 | uint32(b)
}

func (g *WallpaperGenerator) rotateHue(argb uint32, degrees float64) uint32 {
	hex := argbToHex(argb)
	hsl := hexToHSL(hex)
	hsl.H = math.Mod(hsl.H+degrees, 360)
	newHex := hslToHex(hsl)
	r, gr, b := hexToRGB(newHex)
	return 0xFF000000 | uint32(r)<<16 | uint32(gr)<<8 | uint32(b)
}

func (g *WallpaperGenerator) boostPaletteSaturation(palette *material.Palette, factor float64) {
	// Boost saturation for all tones in primary palette
	for tone, color := range palette.Primary.Tones {
		palette.Primary.Tones[tone] = g.boostSaturation(color, factor)
	}
	for tone, color := range palette.Secondary.Tones {
		palette.Secondary.Tones[tone] = g.boostSaturation(color, factor*0.8)
	}
	for tone, color := range palette.Tertiary.Tones {
		palette.Tertiary.Tones[tone] = g.boostSaturation(color, factor*0.6)
	}
}

func (g *WallpaperGenerator) reducePaletteSaturation(palette *material.Palette, factor float64) {
	// Reduce saturation for all tones
	for tone, color := range palette.Primary.Tones {
		palette.Primary.Tones[tone] = g.desaturate(color, factor)
	}
	for tone, color := range palette.Secondary.Tones {
		palette.Secondary.Tones[tone] = g.desaturate(color, factor)
	}
	for tone, color := range palette.Tertiary.Tones {
		palette.Tertiary.Tones[tone] = g.desaturate(color, factor)
	}
}

func (g *WallpaperGenerator) increaseTonesContrast(palette *material.Palette, factor float64) {
	// Increase contrast by making darks darker and lights lighter
	g.adjustPaletteContrast(palette.Primary, factor)
	g.adjustPaletteContrast(palette.Secondary, factor)
	g.adjustPaletteContrast(palette.Tertiary, factor)
}

func (g *WallpaperGenerator) adjustPaletteContrast(tonalPalette material.TonalPalette, factor float64) {
	for tone, color := range tonalPalette.Tones {
		hex := argbToHex(color)
		hsl := hexToHSL(hex)

		// Adjust lightness based on tone
		if tone < 50 {
			// Make darks darker
			hsl.L = hsl.L / factor
		} else {
			// Make lights lighter
			hsl.L = math.Min(100, hsl.L*factor)
		}

		newHex := hslToHex(hsl)
		r, gr, b := hexToRGB(newHex)
		tonalPalette.Tones[tone] = 0xFF000000 | uint32(r)<<16 | uint32(gr)<<8 | uint32(b)
	}
}

func (g *WallpaperGenerator) addColorVariety(palette *material.Palette, extractedColors *material.ExtractedColors) {
	// Use multiple accent colors for variety
	if len(extractedColors.Accents) > 1 {
		// Rotate secondary and tertiary hues more
		palette.Secondary = g.generateTonalPaletteFromColor(extractedColors.Accents[1].Color)
	}
	if len(extractedColors.Accents) > 2 {
		palette.Tertiary = g.generateTonalPaletteFromColor(extractedColors.Accents[2].Color)
	}
}

func (g *WallpaperGenerator) spreadAcrossSpectrum(palette *material.Palette) {
	// Spread colors evenly across the spectrum
	baseSeed := palette.Seed

	// Primary stays as is
	// Secondary at +120 degrees
	palette.Secondary = g.generateTonalPaletteFromColor(g.rotateHue(baseSeed, 120))

	// Tertiary at +240 degrees
	palette.Tertiary = g.generateTonalPaletteFromColor(g.rotateHue(baseSeed, 240))
}

func (g *WallpaperGenerator) generateTonalPaletteFromColor(argb uint32) material.TonalPalette {
	palette := material.TonalPalette{
		Tones: make(map[int]uint32),
	}

	hex := argbToHex(argb)
	baseHSL := hexToHSL(hex)

	// Standard Material You tone values
	tones := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99, 100}

	for _, tone := range tones {
		hsl := baseHSL
		hsl.L = float64(tone)

		// Adjust saturation based on tone
		if tone < 20 {
			hsl.S *= float64(tone) / 20.0
		} else if tone > 80 {
			hsl.S *= float64(100-tone) / 20.0
		}

		newHex := hslToHex(hsl)
		r, gr, b := hexToRGB(newHex)
		palette.Tones[tone] = 0xFF000000 | uint32(r)<<16 | uint32(gr)<<8 | uint32(b)
	}

	return palette
}
