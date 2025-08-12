package color

import (
	"math"
	"testing"
)

func TestNewFromHex(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantErr bool
	}{
		{"Valid hex with #", "#FF0000", 255, 0, 0, false},
		{"Valid hex without #", "00FF00", 0, 255, 0, false},
		{"Valid hex lowercase", "0000ff", 0, 0, 255, false},
		{"Invalid hex length", "FFF", 0, 0, 0, true},
		{"Invalid hex chars", "GGGGGG", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewFromHex(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if c.RGB.R != tt.wantR || c.RGB.G != tt.wantG || c.RGB.B != tt.wantB {
					t.Errorf("NewFromHex() RGB = (%d,%d,%d), want (%d,%d,%d)",
						c.RGB.R, c.RGB.G, c.RGB.B, tt.wantR, tt.wantG, tt.wantB)
				}
			}
		})
	}
}

func TestRGBToHSL(t *testing.T) {
	tests := []struct {
		name    string
		r, g, b uint8
		wantH   float64
		wantS   float64
		wantL   float64
	}{
		{"Pure red", 255, 0, 0, 0, 100, 50},
		{"Pure green", 0, 255, 0, 120, 100, 50},
		{"Pure blue", 0, 0, 255, 240, 100, 50},
		{"White", 255, 255, 255, 0, 0, 100},
		{"Black", 0, 0, 0, 0, 0, 0},
		{"Gray", 128, 128, 128, 0, 0, 50.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rgb := RGB{R: tt.r, G: tt.g, B: tt.b}
			hsl := rgb.ToHSL()

			// Allow small floating point differences
			if math.Abs(hsl.H-tt.wantH) > 1 {
				t.Errorf("HSL.H = %f, want %f", hsl.H, tt.wantH)
			}
			if math.Abs(hsl.S-tt.wantS) > 1 {
				t.Errorf("HSL.S = %f, want %f", hsl.S, tt.wantS)
			}
			if math.Abs(hsl.L-tt.wantL) > 1 {
				t.Errorf("HSL.L = %f, want %f", hsl.L, tt.wantL)
			}
		})
	}
}

func TestHSLToRGB(t *testing.T) {
	tests := []struct {
		name    string
		h, s, l float64
		wantR   uint8
		wantG   uint8
		wantB   uint8
	}{
		{"Pure red", 0, 100, 50, 255, 0, 0},
		{"Pure green", 120, 100, 50, 0, 255, 0},
		{"Pure blue", 240, 100, 50, 0, 0, 255},
		{"White", 0, 0, 100, 255, 255, 255},
		{"Black", 0, 0, 0, 0, 0, 0},
		{"Gray", 0, 0, 50, 128, 128, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hsl := HSL{H: tt.h, S: tt.s, L: tt.l}
			rgb := hsl.ToRGB()

			// Allow small rounding differences
			if abs(int(rgb.R)-int(tt.wantR)) > 1 {
				t.Errorf("RGB.R = %d, want %d", rgb.R, tt.wantR)
			}
			if abs(int(rgb.G)-int(tt.wantG)) > 1 {
				t.Errorf("RGB.G = %d, want %d", rgb.G, tt.wantG)
			}
			if abs(int(rgb.B)-int(tt.wantB)) > 1 {
				t.Errorf("RGB.B = %d, want %d", rgb.B, tt.wantB)
			}
		})
	}
}

func TestLuminance(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		dark bool
	}{
		{"Black", "#000000", true},
		{"White", "#FFFFFF", false},
		{"Dark gray", "#333333", true},
		{"Light gray", "#CCCCCC", false},
		{"Dark blue", "#000080", true},
		{"Light yellow", "#FFFF00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := NewFromHex(tt.hex)
			if c.IsDark() != tt.dark {
				t.Errorf("IsDark() = %v, want %v (luminance: %f)", c.IsDark(), tt.dark, c.Luminance())
			}
		})
	}
}

func TestContrast(t *testing.T) {
	black, _ := NewFromHex("#000000")
	white, _ := NewFromHex("#FFFFFF")

	contrast := Contrast(black, white)
	// WCAG AAA requires contrast ratio of at least 7:1 for normal text
	// Black on white should be 21:1
	if contrast < 20 {
		t.Errorf("Contrast(black, white) = %f, want >= 20", contrast)
	}
}

func TestBlend(t *testing.T) {
	red, _ := NewFromHex("#FF0000")
	blue, _ := NewFromHex("#0000FF")

	// 50% blend should give purple
	blend := Blend(red, blue, 0.5)

	// Should be approximately #800080 (128, 0, 128)
	if abs(int(blend.RGB.R)-128) > 1 {
		t.Errorf("Blend R = %d, want ~128", blend.RGB.R)
	}
	if blend.RGB.G != 0 {
		t.Errorf("Blend G = %d, want 0", blend.RGB.G)
	}
	if abs(int(blend.RGB.B)-128) > 1 {
		t.Errorf("Blend B = %d, want ~128", blend.RGB.B)
	}
}

func TestColorModification(t *testing.T) {
	c, _ := NewFromHex("#808080") // Medium gray

	// Test darkening
	darker := c.Darken(20)
	if darker.HSL.L >= c.HSL.L {
		t.Errorf("Darken failed: L = %f, original = %f", darker.HSL.L, c.HSL.L)
	}

	// Test lightening
	lighter := c.Lighten(20)
	if lighter.HSL.L <= c.HSL.L {
		t.Errorf("Lighten failed: L = %f, original = %f", lighter.HSL.L, c.HSL.L)
	}

	// Test saturation
	c2, _ := NewFromHex("#CC8080") // Desaturated red
	saturated := c2.Saturate(20)
	if saturated.HSL.S <= c2.HSL.S {
		t.Errorf("Saturate failed: S = %f, original = %f", saturated.HSL.S, c2.HSL.S)
	}

	desaturated := c2.Desaturate(20)
	if desaturated.HSL.S >= c2.HSL.S {
		t.Errorf("Desaturate failed: S = %f, original = %f", desaturated.HSL.S, c2.HSL.S)
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
