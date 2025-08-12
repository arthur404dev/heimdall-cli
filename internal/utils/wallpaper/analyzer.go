package wallpaper

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

// Analyzer analyzes wallpaper images for various properties
type Analyzer struct{}

// NewAnalyzer creates a new wallpaper analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzeColourfulness calculates the colourfulness score of an image
// Based on Hasler and Süsstrunk's metric
func (a *Analyzer) AnalyzeColourfulness(path string) (float64, error) {
	img, err := a.loadImage(path)
	if err != nil {
		return 0, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return 0, fmt.Errorf("invalid image dimensions")
	}

	var rgSum, ybSum float64
	var rgSum2, ybSum2 float64
	pixelCount := 0

	// Sample every nth pixel for performance
	sampleRate := 1
	if width*height > 1000000 { // For images > 1MP
		sampleRate = 2
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, g, b, a := img.At(x, y).RGBA()

			// Skip transparent pixels
			if a == 0 {
				continue
			}

			// Convert to 8-bit
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			// Calculate opponent color space values
			rg := r8 - g8
			yb := 0.5*(r8+g8) - b8

			rgSum += rg
			ybSum += yb
			rgSum2 += rg * rg
			ybSum2 += yb * yb

			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 0, fmt.Errorf("no valid pixels found")
	}

	// Calculate means
	rgMean := rgSum / float64(pixelCount)
	ybMean := ybSum / float64(pixelCount)

	// Calculate standard deviations
	rgStd := math.Sqrt(rgSum2/float64(pixelCount) - rgMean*rgMean)
	ybStd := math.Sqrt(ybSum2/float64(pixelCount) - ybMean*ybMean)

	// Calculate colourfulness metric
	stdRoot := math.Sqrt(rgStd*rgStd + ybStd*ybStd)
	meanRoot := math.Sqrt(rgMean*rgMean + ybMean*ybMean)

	colourfulness := stdRoot + 0.3*meanRoot

	return colourfulness, nil
}

// DetermineMode determines if an image is more suitable for dark or light mode
func (a *Analyzer) DetermineMode(path string) (string, error) {
	img, err := a.loadImage(path)
	if err != nil {
		return "", err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid image dimensions")
	}

	var luminanceSum float64
	pixelCount := 0

	// Sample every nth pixel for performance
	sampleRate := 1
	if width*height > 1000000 { // For images > 1MP
		sampleRate = 3
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, g, b, a := img.At(x, y).RGBA()

			// Skip transparent pixels
			if a == 0 {
				continue
			}

			// Convert to 8-bit
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			// Calculate luminance using ITU-R BT.709 formula
			luminance := 0.2126*r8 + 0.7152*g8 + 0.0722*b8
			luminanceSum += luminance
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return "", fmt.Errorf("no valid pixels found")
	}

	avgLuminance := luminanceSum / float64(pixelCount)

	// Determine mode based on average luminance
	// Dark images (avg < 128) are better for light mode themes
	// Light images (avg >= 128) are better for dark mode themes
	if avgLuminance < 128 {
		return "light", nil
	}
	return "dark", nil
}

// GetDimensions returns the width and height of an image
func (a *Analyzer) GetDimensions(path string) (int, int, error) {
	img, err := a.loadImage(path)
	if err != nil {
		return 0, 0, err
	}

	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}

// AnalyzeDominantColors extracts the dominant colors from an image
func (a *Analyzer) AnalyzeDominantColors(path string, numColors int) ([]uint32, error) {
	img, err := a.loadImage(path)
	if err != nil {
		return nil, err
	}

	// Build color histogram
	colorCount := make(map[uint32]int)
	bounds := img.Bounds()

	// Sample every nth pixel for performance
	sampleRate := 1
	if bounds.Dx()*bounds.Dy() > 500000 {
		sampleRate = 2
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += sampleRate {
		for x := bounds.Min.X; x < bounds.Max.X; x += sampleRate {
			r, g, b, a := img.At(x, y).RGBA()

			// Skip transparent pixels
			if a == 0 {
				continue
			}

			// Convert to 8-bit and quantize to reduce color space
			r8 := uint8((r >> 8) & 0xF8) // 5 bits
			g8 := uint8((g >> 8) & 0xF8) // 5 bits
			b8 := uint8((b >> 8) & 0xF8) // 5 bits

			// Create ARGB color
			argb := uint32(0xFF000000) | uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)
			colorCount[argb]++
		}
	}

	// Sort colors by frequency
	type colorFreq struct {
		color uint32
		count int
	}

	colors := make([]colorFreq, 0, len(colorCount))
	for color, count := range colorCount {
		colors = append(colors, colorFreq{color, count})
	}

	// Sort by count descending
	for i := 0; i < len(colors)-1; i++ {
		for j := i + 1; j < len(colors); j++ {
			if colors[j].count > colors[i].count {
				colors[i], colors[j] = colors[j], colors[i]
			}
		}
	}

	// Return top N colors
	result := make([]uint32, 0, numColors)
	for i := 0; i < numColors && i < len(colors); i++ {
		result = append(result, colors[i].color)
	}

	return result, nil
}

// CalculateContrast calculates the contrast ratio between two colors
func (a *Analyzer) CalculateContrast(color1, color2 uint32) float64 {
	// Extract RGB components
	r1 := float64((color1>>16)&0xFF) / 255.0
	g1 := float64((color1>>8)&0xFF) / 255.0
	b1 := float64(color1&0xFF) / 255.0

	r2 := float64((color2>>16)&0xFF) / 255.0
	g2 := float64((color2>>8)&0xFF) / 255.0
	b2 := float64(color2&0xFF) / 255.0

	// Calculate relative luminance
	l1 := a.relativeLuminance(r1, g1, b1)
	l2 := a.relativeLuminance(r2, g2, b2)

	// Calculate contrast ratio
	if l1 > l2 {
		return (l1 + 0.05) / (l2 + 0.05)
	}
	return (l2 + 0.05) / (l1 + 0.05)
}

// relativeLuminance calculates the relative luminance of a color
func (a *Analyzer) relativeLuminance(r, g, b float64) float64 {
	// Apply gamma correction
	r = a.gammaCorrect(r)
	g = a.gammaCorrect(g)
	b = a.gammaCorrect(b)

	// Calculate luminance using ITU-R BT.709
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// gammaCorrect applies gamma correction to a color channel
func (a *Analyzer) gammaCorrect(channel float64) float64 {
	if channel <= 0.03928 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}

// loadImage loads an image from a file path
func (a *Analyzer) loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// Info represents wallpaper metadata
type Info struct {
	Path          string  `json:"path"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	Colourfulness float64 `json:"colourfulness"`
	Mode          string  `json:"mode"` // "dark" or "light"
	Hash          string  `json:"hash,omitempty"`
}

// Analyze performs a complete analysis of a wallpaper
func (a *Analyzer) Analyze(path string) (*Info, error) {
	info := &Info{
		Path: path,
	}

	// Get dimensions
	width, height, err := a.GetDimensions(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get dimensions: %w", err)
	}
	info.Width = width
	info.Height = height

	// Analyze colourfulness
	colourfulness, err := a.AnalyzeColourfulness(path)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze colourfulness: %w", err)
	}
	info.Colourfulness = colourfulness

	// Determine mode
	mode, err := a.DetermineMode(path)
	if err != nil {
		return nil, fmt.Errorf("failed to determine mode: %w", err)
	}
	info.Mode = mode

	return info, nil
}
