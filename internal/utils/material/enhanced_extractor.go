package material

import (
	"fmt"
	"image"
	"math"
	"sort"
)

// EnhancedExtractor provides improved color extraction for wallpapers
type EnhancedExtractor struct {
	quantizer *Quantizer
	scorer    *Scorer
}

// NewEnhancedExtractor creates a new enhanced color extractor
func NewEnhancedExtractor() *EnhancedExtractor {
	return &EnhancedExtractor{
		quantizer: NewQuantizer(256), // More colors for better analysis
		scorer:    NewScorer(),
	}
}

// ColorInfo contains detailed information about an extracted color
type ColorInfo struct {
	Color        uint32
	Population   int
	Vibrancy     float64
	Saturation   float64
	Luminance    float64
	IsAccent     bool
	IsBackground bool
}

// ExtractColors performs multi-pass color extraction
func (e *EnhancedExtractor) ExtractColors(img image.Image) (*ExtractedColors, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}

	// Pass 1: Dominant colors by volume
	dominantColors := e.extractDominantColors(img)

	// Pass 2: Vibrant accent colors
	accentColors := e.extractVibrantColors(img)

	// Pass 3: Background color detection
	backgroundColor := e.extractBackgroundColor(img)

	// Pass 4: Edge/contrast colors for UI elements
	edgeColors := e.extractEdgeColors(img)

	// Analyze overall luminance
	avgLuminance := e.analyzeLuminance(img)
	isDark := avgLuminance < 0.5

	// Combine and deduplicate colors
	allColors := e.combineColors(dominantColors, accentColors, edgeColors)

	return &ExtractedColors{
		Dominant:     dominantColors,
		Accents:      accentColors,
		Background:   backgroundColor,
		EdgeColors:   edgeColors,
		AllColors:    allColors,
		IsDark:       isDark,
		AvgLuminance: avgLuminance,
	}, nil
}

// extractDominantColors finds the most prominent colors by volume
func (e *EnhancedExtractor) extractDominantColors(img image.Image) []ColorInfo {
	quantResult := e.quantizer.Quantize(img)

	colors := make([]ColorInfo, 0)
	for argb, count := range quantResult.Colors {
		info := e.analyzeColor(argb, count)
		colors = append(colors, info)
	}

	// Sort by population
	sort.Slice(colors, func(i, j int) bool {
		return colors[i].Population > colors[j].Population
	})

	// Return top 10 dominant colors
	if len(colors) > 10 {
		colors = colors[:10]
	}

	return colors
}

// extractVibrantColors finds high saturation accent colors
func (e *EnhancedExtractor) extractVibrantColors(img image.Image) []ColorInfo {
	bounds := img.Bounds()
	vibrantMap := make(map[uint32]int)

	// Sample image for vibrant colors
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 2 {
			c := img.At(x, y)
			argb := colorToARGB(c)

			// Check if color is vibrant
			info := e.analyzeColor(argb, 1)
			if info.Vibrancy > 0.6 && info.Saturation > 0.5 {
				vibrantMap[argb]++
			}
		}
	}

	// Convert to ColorInfo slice
	colors := make([]ColorInfo, 0)
	for argb, count := range vibrantMap {
		info := e.analyzeColor(argb, count)
		info.IsAccent = true
		colors = append(colors, info)
	}

	// Sort by vibrancy * population
	sort.Slice(colors, func(i, j int) bool {
		scoreI := colors[i].Vibrancy * float64(colors[i].Population)
		scoreJ := colors[j].Vibrancy * float64(colors[j].Population)
		return scoreI > scoreJ
	})

	// Return top 5 vibrant colors
	if len(colors) > 5 {
		colors = colors[:5]
	}

	return colors
}

// extractBackgroundColor finds the most likely background color
func (e *EnhancedExtractor) extractBackgroundColor(img image.Image) ColorInfo {
	bounds := img.Bounds()
	cornerColors := make(map[uint32]int)

	// Sample corners and edges
	sampleSize := bounds.Dx() / 10
	if sampleSize > 50 {
		sampleSize = 50
	}

	// Top-left corner
	for y := bounds.Min.Y; y < bounds.Min.Y+sampleSize && y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Min.X+sampleSize && x < bounds.Max.X; x++ {
			c := img.At(x, y)
			argb := colorToARGB(c)
			cornerColors[argb]++
		}
	}

	// Top-right corner
	for y := bounds.Min.Y; y < bounds.Min.Y+sampleSize && y < bounds.Max.Y; y++ {
		for x := bounds.Max.X - sampleSize; x < bounds.Max.X && x >= bounds.Min.X; x++ {
			c := img.At(x, y)
			argb := colorToARGB(c)
			cornerColors[argb]++
		}
	}

	// Bottom-left corner
	for y := bounds.Max.Y - sampleSize; y < bounds.Max.Y && y >= bounds.Min.Y; y++ {
		for x := bounds.Min.X; x < bounds.Min.X+sampleSize && x < bounds.Max.X; x++ {
			c := img.At(x, y)
			argb := colorToARGB(c)
			cornerColors[argb]++
		}
	}

	// Bottom-right corner
	for y := bounds.Max.Y - sampleSize; y < bounds.Max.Y && y >= bounds.Min.Y; y++ {
		for x := bounds.Max.X - sampleSize; x < bounds.Max.X && x >= bounds.Min.X; x++ {
			c := img.At(x, y)
			argb := colorToARGB(c)
			cornerColors[argb]++
		}
	}

	// Find most common corner color
	var bgColor uint32
	maxCount := 0
	for argb, count := range cornerColors {
		if count > maxCount {
			maxCount = count
			bgColor = argb
		}
	}

	info := e.analyzeColor(bgColor, maxCount)
	info.IsBackground = true
	return info
}

// extractEdgeColors finds colors from high-contrast edges
func (e *EnhancedExtractor) extractEdgeColors(img image.Image) []ColorInfo {
	bounds := img.Bounds()
	edgeColors := make(map[uint32]int)

	// Simple edge detection using color differences
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y += 3 {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x += 3 {
			c := img.At(x, y)
			argb := colorToARGB(c)

			// Check neighbors for contrast
			neighbors := []image.Point{
				{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1},
			}

			hasHighContrast := false
			for _, n := range neighbors {
				nc := img.At(n.X, n.Y)
				nargb := colorToARGB(nc)

				if colorDistance(argb, nargb) > 50 {
					hasHighContrast = true
					break
				}
			}

			if hasHighContrast {
				edgeColors[argb]++
			}
		}
	}

	// Convert to ColorInfo slice
	colors := make([]ColorInfo, 0)
	for argb, count := range edgeColors {
		info := e.analyzeColor(argb, count)
		colors = append(colors, info)
	}

	// Sort by population
	sort.Slice(colors, func(i, j int) bool {
		return colors[i].Population > colors[j].Population
	})

	// Return top 5 edge colors
	if len(colors) > 5 {
		colors = colors[:5]
	}

	return colors
}

// analyzeLuminance calculates the average luminance of the image
func (e *EnhancedExtractor) analyzeLuminance(img image.Image) float64 {
	bounds := img.Bounds()
	totalLuminance := 0.0
	pixelCount := 0

	// Sample every 4th pixel for performance
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 4 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 4 {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()

			// Convert to 0-255 range
			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			// Calculate relative luminance
			luminance := (0.2126*r8 + 0.7152*g8 + 0.0722*b8) / 255.0
			totalLuminance += luminance
			pixelCount++
		}
	}

	if pixelCount == 0 {
		return 0.5
	}

	return totalLuminance / float64(pixelCount)
}

// analyzeColor extracts detailed information about a color
func (e *EnhancedExtractor) analyzeColor(argb uint32, population int) ColorInfo {
	r := float64((argb >> 16) & 0xFF)
	g := float64((argb >> 8) & 0xFF)
	b := float64(argb & 0xFF)

	// Calculate luminance
	luminance := (0.2126*r + 0.7152*g + 0.0722*b) / 255.0

	// Calculate saturation in HSL
	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	saturation := 0.0
	if delta != 0 {
		l := (max + min) / 2.0 / 255.0
		if l < 0.5 {
			saturation = delta / (max + min)
		} else {
			saturation = delta / (510.0 - max - min)
		}
	}

	// Calculate vibrancy (combination of saturation and moderate luminance)
	vibrancy := saturation
	if luminance < 0.2 || luminance > 0.8 {
		// Reduce vibrancy for very dark or very light colors
		vibrancy *= 1.0 - math.Abs(0.5-luminance)*2
	}

	return ColorInfo{
		Color:      argb,
		Population: population,
		Vibrancy:   vibrancy,
		Saturation: saturation,
		Luminance:  luminance,
	}
}

// combineColors merges and deduplicates color lists
func (e *EnhancedExtractor) combineColors(colorLists ...[]ColorInfo) []ColorInfo {
	colorMap := make(map[uint32]ColorInfo)

	for _, list := range colorLists {
		for _, info := range list {
			existing, exists := colorMap[info.Color]
			if !exists || info.Population > existing.Population {
				colorMap[info.Color] = info
			}
		}
	}

	// Convert to slice
	result := make([]ColorInfo, 0, len(colorMap))
	for _, info := range colorMap {
		result = append(result, info)
	}

	return result
}

// colorDistance calculates the perceptual distance between two colors
func colorDistance(c1, c2 uint32) float64 {
	r1 := float64((c1 >> 16) & 0xFF)
	g1 := float64((c1 >> 8) & 0xFF)
	b1 := float64(c1 & 0xFF)

	r2 := float64((c2 >> 16) & 0xFF)
	g2 := float64((c2 >> 8) & 0xFF)
	b2 := float64(c2 & 0xFF)

	// Simple Euclidean distance in RGB space
	dr := r1 - r2
	dg := g1 - g2
	db := b1 - b2

	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// ExtractedColors contains the results of color extraction
type ExtractedColors struct {
	Dominant     []ColorInfo
	Accents      []ColorInfo
	Background   ColorInfo
	EdgeColors   []ColorInfo
	AllColors    []ColorInfo
	IsDark       bool
	AvgLuminance float64
}

// GetBestSeedColor finds the best seed color for theme generation
func (ec *ExtractedColors) GetBestSeedColor() uint32 {
	// Prefer vibrant accent colors if available
	if len(ec.Accents) > 0 {
		return ec.Accents[0].Color
	}

	// Otherwise use the most vibrant dominant color
	bestVibrancy := 0.0
	bestColor := uint32(0xFF4285F4) // Default Material Blue

	for _, info := range ec.Dominant {
		if info.Vibrancy > bestVibrancy {
			bestVibrancy = info.Vibrancy
			bestColor = info.Color
		}
	}

	return bestColor
}
