package material

import "fmt"

// TonalPalette represents a range of tones for a single color
type TonalPalette struct {
	Tones map[int]uint32 // Map of tone (0-100) to ARGB color
}

// Tone returns the color at the specified tone level
func (tp *TonalPalette) Tone(tone int) uint32 {
	if color, ok := tp.Tones[tone]; ok {
		return color
	}

	// If exact tone not found, find closest
	closestTone := 0
	minDiff := 100

	for t := range tp.Tones {
		diff := abs(t - tone)
		if diff < minDiff {
			minDiff = diff
			closestTone = t
		}
	}

	return tp.Tones[closestTone]
}

// Palette represents a complete Material You color palette
type Palette struct {
	Seed           uint32       // Seed color used to generate the palette
	Primary        TonalPalette // Primary color palette
	Secondary      TonalPalette // Secondary color palette
	Tertiary       TonalPalette // Tertiary color palette
	Neutral        TonalPalette // Neutral (grayscale) palette
	NeutralVariant TonalPalette // Neutral variant palette
	Error          TonalPalette // Error color palette
}

// Scheme represents a complete Material You color scheme for light or dark mode
type Scheme struct {
	Seed    uint32   // Seed color
	IsDark  bool     // Whether this is a dark theme
	Palette *Palette // The underlying palette

	// Primary colors
	Primary            uint32
	OnPrimary          uint32
	PrimaryContainer   uint32
	OnPrimaryContainer uint32

	// Secondary colors
	Secondary            uint32
	OnSecondary          uint32
	SecondaryContainer   uint32
	OnSecondaryContainer uint32

	// Tertiary colors
	Tertiary            uint32
	OnTertiary          uint32
	TertiaryContainer   uint32
	OnTertiaryContainer uint32

	// Error colors
	Error            uint32
	OnError          uint32
	ErrorContainer   uint32
	OnErrorContainer uint32

	// Background colors
	Background   uint32
	OnBackground uint32

	// Surface colors
	Surface          uint32
	OnSurface        uint32
	SurfaceVariant   uint32
	OnSurfaceVariant uint32

	// Outline colors
	Outline        uint32
	OutlineVariant uint32

	// Other colors
	Shadow           uint32
	Scrim            uint32
	InverseSurface   uint32
	InverseOnSurface uint32
	InversePrimary   uint32
}

// ToMap converts the scheme to a map for template rendering
func (s *Scheme) ToMap() map[string]string {
	return map[string]string{
		"primary":              argbToHex(s.Primary),
		"on_primary":           argbToHex(s.OnPrimary),
		"primary_container":    argbToHex(s.PrimaryContainer),
		"on_primary_container": argbToHex(s.OnPrimaryContainer),

		"secondary":              argbToHex(s.Secondary),
		"on_secondary":           argbToHex(s.OnSecondary),
		"secondary_container":    argbToHex(s.SecondaryContainer),
		"on_secondary_container": argbToHex(s.OnSecondaryContainer),

		"tertiary":              argbToHex(s.Tertiary),
		"on_tertiary":           argbToHex(s.OnTertiary),
		"tertiary_container":    argbToHex(s.TertiaryContainer),
		"on_tertiary_container": argbToHex(s.OnTertiaryContainer),

		"error":              argbToHex(s.Error),
		"on_error":           argbToHex(s.OnError),
		"error_container":    argbToHex(s.ErrorContainer),
		"on_error_container": argbToHex(s.OnErrorContainer),

		"background":    argbToHex(s.Background),
		"on_background": argbToHex(s.OnBackground),

		"surface":            argbToHex(s.Surface),
		"on_surface":         argbToHex(s.OnSurface),
		"surface_variant":    argbToHex(s.SurfaceVariant),
		"on_surface_variant": argbToHex(s.OnSurfaceVariant),

		"outline":         argbToHex(s.Outline),
		"outline_variant": argbToHex(s.OutlineVariant),

		"shadow":             argbToHex(s.Shadow),
		"scrim":              argbToHex(s.Scrim),
		"inverse_surface":    argbToHex(s.InverseSurface),
		"inverse_on_surface": argbToHex(s.InverseOnSurface),
		"inverse_primary":    argbToHex(s.InversePrimary),
	}
}

// argbToHex converts an ARGB color to hex string
func argbToHex(argb uint32) string {
	r := (argb >> 16) & 0xFF
	g := (argb >> 8) & 0xFF
	b := argb & 0xFF
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
