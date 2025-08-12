package material

import (
	"fmt"
	"image"
	"math"

	"github.com/arthur404dev/heimdall-cli/internal/utils/color"
)

// Generator creates Material You color palettes from images
type Generator struct {
	quantizer *Quantizer
	scorer    *Scorer
}

// NewGenerator creates a new Material You generator
func NewGenerator() *Generator {
	return &Generator{
		quantizer: NewQuantizer(128),
		scorer:    NewScorer(),
	}
}

// GenerateFromImage creates a Material You palette from an image
func (g *Generator) GenerateFromImage(img image.Image) (*Palette, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}

	// Quantize the image to extract dominant colors
	quantResult := g.quantizer.Quantize(img)

	// Find the best seed color
	seedColor := FindSeedColor(quantResult)

	// Generate palette from seed color
	return g.GenerateFromColor(seedColor)
}

// GenerateFromColor creates a Material You palette from a seed color
func (g *Generator) GenerateFromColor(seedARGB uint32) (*Palette, error) {
	// Convert seed to our color type
	r := uint8((seedARGB >> 16) & 0xFF)
	gr := uint8((seedARGB >> 8) & 0xFF)
	b := uint8(seedARGB & 0xFF)

	seedColor := color.NewFromRGB(r, gr, b)
	seedHSL := seedColor.HSL

	// Create tonal palettes for each role
	palette := &Palette{
		Seed:    seedARGB,
		Primary: g.generateTonalPalette(seedHSL, 0),
		Secondary: g.generateTonalPalette(
			color.HSL{H: math.Mod(seedHSL.H+60, 360), S: seedHSL.S * 0.7, L: seedHSL.L},
			60,
		),
		Tertiary: g.generateTonalPalette(
			color.HSL{H: math.Mod(seedHSL.H+120, 360), S: seedHSL.S * 0.5, L: seedHSL.L},
			120,
		),
		Neutral:        g.generateNeutralPalette(seedHSL),
		NeutralVariant: g.generateNeutralVariantPalette(seedHSL),
		Error:          g.generateErrorPalette(),
	}

	return palette, nil
}

// generateTonalPalette creates a tonal palette from a base color
func (g *Generator) generateTonalPalette(baseHSL color.HSL, hueShift float64) TonalPalette {
	palette := TonalPalette{
		Tones: make(map[int]uint32),
	}

	// Standard Material You tone values
	tones := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99, 100}

	for _, tone := range tones {
		// Calculate lightness for this tone
		// Tone 0 = black, Tone 100 = white
		lightness := float64(tone)

		// Adjust saturation based on tone
		// Lower saturation at extremes
		saturation := baseHSL.S
		if tone < 20 {
			saturation *= float64(tone) / 20.0
		} else if tone > 80 {
			saturation *= float64(100-tone) / 20.0
		}

		// Create color with adjusted values
		hsl := color.HSL{
			H: math.Mod(baseHSL.H+hueShift, 360),
			S: saturation,
			L: lightness,
		}

		rgb := color.NewFromHSL(hsl.H, hsl.S, hsl.L).RGB
		argb := 0xFF000000 | uint32(rgb.R)<<16 | uint32(rgb.G)<<8 | uint32(rgb.B)

		palette.Tones[tone] = argb
	}

	return palette
}

// generateNeutralPalette creates a neutral (grayscale) palette
func (g *Generator) generateNeutralPalette(seedHSL color.HSL) TonalPalette {
	palette := TonalPalette{
		Tones: make(map[int]uint32),
	}

	tones := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99, 100}

	for _, tone := range tones {
		// Neutral colors have very low saturation
		// Add a tiny bit of the seed hue for warmth
		hsl := color.HSL{
			H: seedHSL.H,
			S: 2.0, // Very low saturation
			L: float64(tone),
		}

		rgb := color.NewFromHSL(hsl.H, hsl.S, hsl.L).RGB
		argb := 0xFF000000 | uint32(rgb.R)<<16 | uint32(rgb.G)<<8 | uint32(rgb.B)

		palette.Tones[tone] = argb
	}

	return palette
}

// generateNeutralVariantPalette creates a neutral variant palette with slight color
func (g *Generator) generateNeutralVariantPalette(seedHSL color.HSL) TonalPalette {
	palette := TonalPalette{
		Tones: make(map[int]uint32),
	}

	tones := []int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99, 100}

	for _, tone := range tones {
		// Neutral variant has slightly more saturation than pure neutral
		hsl := color.HSL{
			H: seedHSL.H,
			S: 8.0, // Low but noticeable saturation
			L: float64(tone),
		}

		rgb := color.NewFromHSL(hsl.H, hsl.S, hsl.L).RGB
		argb := 0xFF000000 | uint32(rgb.R)<<16 | uint32(rgb.G)<<8 | uint32(rgb.B)

		palette.Tones[tone] = argb
	}

	return palette
}

// generateErrorPalette creates the error color palette
func (g *Generator) generateErrorPalette() TonalPalette {
	// Error colors are based on red
	errorHSL := color.HSL{
		H: 0,  // Red hue
		S: 84, // High saturation
		L: 50, // Medium lightness
	}

	return g.generateTonalPalette(errorHSL, 0)
}

// GenerateScheme creates a complete Material You color scheme
func (g *Generator) GenerateScheme(seedARGB uint32, isDark bool) (*Scheme, error) {
	palette, err := g.GenerateFromColor(seedARGB)
	if err != nil {
		return nil, err
	}

	scheme := &Scheme{
		Seed:    seedARGB,
		IsDark:  isDark,
		Palette: palette,
	}

	if isDark {
		// Dark theme tone mappings
		scheme.Primary = palette.Primary.Tone(80)
		scheme.OnPrimary = palette.Primary.Tone(20)
		scheme.PrimaryContainer = palette.Primary.Tone(30)
		scheme.OnPrimaryContainer = palette.Primary.Tone(90)

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

		scheme.Shadow = 0xFF000000
		scheme.Scrim = 0xFF000000

		scheme.InverseSurface = palette.Neutral.Tone(90)
		scheme.InverseOnSurface = palette.Neutral.Tone(20)
		scheme.InversePrimary = palette.Primary.Tone(40)
	} else {
		// Light theme tone mappings
		scheme.Primary = palette.Primary.Tone(40)
		scheme.OnPrimary = palette.Primary.Tone(100)
		scheme.PrimaryContainer = palette.Primary.Tone(90)
		scheme.OnPrimaryContainer = palette.Primary.Tone(10)

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

		scheme.Shadow = 0xFF000000
		scheme.Scrim = 0xFF000000

		scheme.InverseSurface = palette.Neutral.Tone(20)
		scheme.InverseOnSurface = palette.Neutral.Tone(95)
		scheme.InversePrimary = palette.Primary.Tone(80)
	}

	return scheme, nil
}
