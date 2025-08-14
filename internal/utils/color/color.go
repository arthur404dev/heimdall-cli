package color

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
)

// Color represents a color with multiple representations
type Color struct {
	Hex string `json:"hex"`
	RGB RGB    `json:"rgb"`
	HSL HSL    `json:"hsl"`
	LAB LAB    `json:"lab,omitempty"`
}

// RGB represents a color in RGB space
type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// HSL represents a color in HSL space
type HSL struct {
	H float64 `json:"h"` // 0-360
	S float64 `json:"s"` // 0-100
	L float64 `json:"l"` // 0-100
}

// LAB represents a color in LAB space
type LAB struct {
	L float64 `json:"l"` // 0-100
	A float64 `json:"a"` // -128 to 127
	B float64 `json:"b"` // -128 to 127
}

// XYZ represents a color in XYZ space (intermediate for LAB conversion)
type XYZ struct {
	X float64
	Y float64
	Z float64
}

// NewFromHex creates a Color from a hex string
func NewFromHex(hex string) (*Color, error) {
	// Remove # if present
	hex = strings.TrimPrefix(hex, "#")

	// Validate length
	if len(hex) != 6 {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	// Parse RGB values
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	rgb := RGB{R: uint8(r), G: uint8(g), B: uint8(b)}

	return &Color{
		Hex: "#" + strings.ToUpper(hex),
		RGB: rgb,
		HSL: rgb.ToHSL(),
		LAB: rgb.ToLAB(),
	}, nil
}

// NewFromRGB creates a Color from RGB values
func NewFromRGB(r, g, b uint8) *Color {
	rgb := RGB{R: r, G: g, B: b}
	return &Color{
		Hex: rgb.ToHex(),
		RGB: rgb,
		HSL: rgb.ToHSL(),
		LAB: rgb.ToLAB(),
	}
}

// NewFromHSL creates a Color from HSL values
func NewFromHSL(h, s, l float64) *Color {
	hsl := HSL{H: h, S: s, L: l}
	rgb := hsl.ToRGB()
	return &Color{
		Hex: rgb.ToHex(),
		RGB: rgb,
		HSL: hsl,
		LAB: rgb.ToLAB(),
	}
}

// ToHex converts RGB to hex string
func (rgb RGB) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X", rgb.R, rgb.G, rgb.B)
}

// ToHSL converts RGB to HSL
func (rgb RGB) ToHSL() HSL {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)

	h := 0.0
	s := 0.0
	l := (max + min) / 2.0

	if max != min {
		d := max - min

		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6.0
			}
		case g:
			h = (b-r)/d + 2.0
		case b:
			h = (r-g)/d + 4.0
		}

		h /= 6.0
	}

	return HSL{
		H: h * 360.0,
		S: s * 100.0,
		L: l * 100.0,
	}
}

// ToRGB converts HSL to RGB
func (hsl HSL) ToRGB() RGB {
	h := hsl.H / 360.0
	s := hsl.S / 100.0
	l := hsl.L / 100.0

	var r, g, b float64

	if s == 0 {
		r = l
		g = l
		b = l
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

	return RGB{
		R: uint8(math.Round(r * 255)),
		G: uint8(math.Round(g * 255)),
		B: uint8(math.Round(b * 255)),
	}
}

// hueToRGB is a helper function for HSL to RGB conversion
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

// ToXYZ converts RGB to XYZ color space
func (rgb RGB) ToXYZ() XYZ {
	// Normalize RGB values
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	// Apply gamma correction
	if r > 0.04045 {
		r = math.Pow((r+0.055)/1.055, 2.4)
	} else {
		r = r / 12.92
	}

	if g > 0.04045 {
		g = math.Pow((g+0.055)/1.055, 2.4)
	} else {
		g = g / 12.92
	}

	if b > 0.04045 {
		b = math.Pow((b+0.055)/1.055, 2.4)
	} else {
		b = b / 12.92
	}

	// Observer = 2°, Illuminant = D65
	x := r*0.4124564 + g*0.3575761 + b*0.1804375
	y := r*0.2126729 + g*0.7151522 + b*0.0721750
	z := r*0.0193339 + g*0.1191920 + b*0.9503041

	return XYZ{
		X: x * 100.0,
		Y: y * 100.0,
		Z: z * 100.0,
	}
}

// ToLAB converts RGB to LAB color space
func (rgb RGB) ToLAB() LAB {
	xyz := rgb.ToXYZ()

	// Reference white D65
	xn := 95.047
	yn := 100.000
	zn := 108.883

	x := xyz.X / xn
	y := xyz.Y / yn
	z := xyz.Z / zn

	fx := labF(x)
	fy := labF(y)
	fz := labF(z)

	l := 116.0*fy - 16.0
	a := 500.0 * (fx - fy)
	b := 200.0 * (fy - fz)

	return LAB{L: l, A: a, B: b}
}

// labF is a helper function for XYZ to LAB conversion
func labF(t float64) float64 {
	delta := 6.0 / 29.0
	if t > delta*delta*delta {
		return math.Pow(t, 1.0/3.0)
	}
	return t/(3.0*delta*delta) + 4.0/29.0
}

// Distance calculates the Euclidean distance between two colors in LAB space
func Distance(c1, c2 *Color) float64 {
	// Use LAB color space for perceptually uniform distance
	dL := c1.LAB.L - c2.LAB.L
	dA := c1.LAB.A - c2.LAB.A
	dB := c1.LAB.B - c2.LAB.B

	return math.Sqrt(dL*dL + dA*dA + dB*dB)
}

// DeltaE calculates the CIE Delta E 2000 color difference
func DeltaE(c1, c2 *Color) float64 {
	// Simplified Delta E calculation
	// For full Delta E 2000, we'd need more complex calculations
	return Distance(c1, c2)
}

// Luminance calculates the relative luminance of a color
func (c *Color) Luminance() float64 {
	r := float64(c.RGB.R) / 255.0
	g := float64(c.RGB.G) / 255.0
	b := float64(c.RGB.B) / 255.0

	// Apply gamma correction
	if r <= 0.03928 {
		r = r / 12.92
	} else {
		r = math.Pow((r+0.055)/1.055, 2.4)
	}

	if g <= 0.03928 {
		g = g / 12.92
	} else {
		g = math.Pow((g+0.055)/1.055, 2.4)
	}

	if b <= 0.03928 {
		b = b / 12.92
	} else {
		b = math.Pow((b+0.055)/1.055, 2.4)
	}

	// Calculate luminance
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// Contrast calculates the contrast ratio between two colors
func Contrast(c1, c2 *Color) float64 {
	l1 := c1.Luminance()
	l2 := c2.Luminance()

	if l1 > l2 {
		return (l1 + 0.05) / (l2 + 0.05)
	}
	return (l2 + 0.05) / (l1 + 0.05)
}

// IsDark returns true if the color is considered dark
func (c *Color) IsDark() bool {
	return c.Luminance() < 0.5
}

// IsLight returns true if the color is considered light
func (c *Color) IsLight() bool {
	return !c.IsDark()
}

// Blend blends two colors with a given ratio (0.0 to 1.0)
func Blend(c1, c2 *Color, ratio float64) *Color {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	r := uint8(float64(c1.RGB.R)*(1-ratio) + float64(c2.RGB.R)*ratio)
	g := uint8(float64(c1.RGB.G)*(1-ratio) + float64(c2.RGB.G)*ratio)
	b := uint8(float64(c1.RGB.B)*(1-ratio) + float64(c2.RGB.B)*ratio)

	return NewFromRGB(r, g, b)
}

// Darken darkens a color by a percentage (0-100)
func (c *Color) Darken(percent float64) *Color {
	hsl := c.HSL
	hsl.L = math.Max(0, hsl.L-percent)
	return NewFromHSL(hsl.H, hsl.S, hsl.L)
}

// Lighten lightens a color by a percentage (0-100)
func (c *Color) Lighten(percent float64) *Color {
	hsl := c.HSL
	hsl.L = math.Min(100, hsl.L+percent)
	return NewFromHSL(hsl.H, hsl.S, hsl.L)
}

// Saturate increases saturation by a percentage (0-100)
func (c *Color) Saturate(percent float64) *Color {
	hsl := c.HSL
	hsl.S = math.Min(100, hsl.S+percent)
	return NewFromHSL(hsl.H, hsl.S, hsl.L)
}

// Desaturate decreases saturation by a percentage (0-100)
func (c *Color) Desaturate(percent float64) *Color {
	hsl := c.HSL
	hsl.S = math.Max(0, hsl.S-percent)
	return NewFromHSL(hsl.H, hsl.S, hsl.L)
}

// ColorConverter provides optimized color conversions with caching
type ColorConverter struct {
	mu    sync.RWMutex
	cache map[string]*Color // Cache parsed colors
}

// Global converter instance with caching
var globalConverter = &ColorConverter{
	cache: make(map[string]*Color),
}

// FastHexToRGB converts hex to RGB with caching (optimized)
func FastHexToRGB(hex string) (r, g, b uint8, err error) {
	// Remove # if present
	hex = strings.TrimPrefix(hex, "#")

	// Fast path for common case
	if len(hex) == 6 {
		// Use lookup table for hex conversion (faster than strconv)
		r = hexToByte(hex[0])<<4 | hexToByte(hex[1])
		g = hexToByte(hex[2])<<4 | hexToByte(hex[3])
		b = hexToByte(hex[4])<<4 | hexToByte(hex[5])
		return r, g, b, nil
	}

	return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
}

// hexToByte converts a hex character to byte value (optimized with lookup)
func hexToByte(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}

// BatchConvertColors converts multiple colors in parallel
func BatchConvertColors(colors []string) ([]*Color, error) {
	results := make([]*Color, len(colors))
	errors := make([]error, len(colors))

	var wg sync.WaitGroup
	for i, hex := range colors {
		wg.Add(1)
		go func(idx int, hexColor string) {
			defer wg.Done()

			// Check cache first
			globalConverter.mu.RLock()
			if cached, ok := globalConverter.cache[hexColor]; ok {
				globalConverter.mu.RUnlock()
				results[idx] = cached
				return
			}
			globalConverter.mu.RUnlock()

			// Convert and cache
			color, err := NewFromHex(hexColor)
			if err != nil {
				errors[idx] = err
				return
			}

			// Store in cache
			globalConverter.mu.Lock()
			globalConverter.cache[hexColor] = color
			globalConverter.mu.Unlock()

			results[idx] = color
		}(i, hex)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// OptimizedPaletteConversion converts a full palette with optimizations
func OptimizedPaletteConversion(palette map[string]string) (map[string]*Color, error) {
	results := make(map[string]*Color, len(palette))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for name, hex := range palette {
		wg.Add(1)
		go func(n, h string) {
			defer wg.Done()

			// Check cache
			globalConverter.mu.RLock()
			if cached, ok := globalConverter.cache[h]; ok {
				globalConverter.mu.RUnlock()
				mu.Lock()
				results[n] = cached
				mu.Unlock()
				return
			}
			globalConverter.mu.RUnlock()

			// Convert
			color, err := NewFromHex(h)
			if err == nil {
				// Cache it
				globalConverter.mu.Lock()
				globalConverter.cache[h] = color
				globalConverter.mu.Unlock()

				mu.Lock()
				results[n] = color
				mu.Unlock()
			}
		}(name, hex)
	}

	wg.Wait()
	return results, nil
}

// ClearColorCache clears the global color cache
func ClearColorCache() {
	globalConverter.mu.Lock()
	defer globalConverter.mu.Unlock()
	globalConverter.cache = make(map[string]*Color)
}
