package material

import (
	"math"
	"sort"
)

// Score represents a color's suitability score for theming
type Score struct {
	Color uint32
	Score float64
}

// Scorer calculates scores for colors based on Material You criteria
type Scorer struct {
	targetChroma     float64
	chromaWeight     float64
	populationWeight float64
}

// NewScorer creates a new scorer with default weights
func NewScorer() *Scorer {
	return &Scorer{
		targetChroma:     48.0,
		chromaWeight:     0.7,
		populationWeight: 0.3,
	}
}

// ScoreColors scores a set of colors based on Material You criteria
func (s *Scorer) ScoreColors(colors map[uint32]int) []Score {
	if len(colors) == 0 {
		return nil
	}

	// Calculate total population
	totalPopulation := 0
	for _, count := range colors {
		totalPopulation += count
	}

	if totalPopulation == 0 {
		return nil
	}

	// Score each color
	scores := make([]Score, 0, len(colors))

	for color, count := range colors {
		// Skip very dark or very light colors
		if !isColorSuitable(color) {
			continue
		}

		// Calculate population score (0-100)
		populationScore := 100.0 * float64(count) / float64(totalPopulation)

		// Calculate chroma score (0-100)
		chromaScore := s.calculateChromaScore(color)

		// Combine scores
		finalScore := s.chromaWeight*chromaScore + s.populationWeight*populationScore

		scores = append(scores, Score{
			Color: color,
			Score: finalScore,
		})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores
}

// calculateChromaScore calculates how close a color's chroma is to the target
func (s *Scorer) calculateChromaScore(argb uint32) float64 {
	// Convert to LAB color space for better chroma calculation
	lab := argbToLAB(argb)

	// Calculate chroma in LAB space
	chroma := math.Sqrt(lab.A*lab.A + lab.B*lab.B)

	// Score based on distance from target chroma
	// Maximum score when chroma equals target, decreasing as distance increases
	distance := math.Abs(chroma - s.targetChroma)

	// Use a Gaussian-like scoring function
	score := 100.0 * math.Exp(-distance*distance/1000.0)

	return score
}

// isColorSuitable checks if a color is suitable for theming
func isColorSuitable(argb uint32) bool {
	r := float64((argb >> 16) & 0xFF)
	g := float64((argb >> 8) & 0xFF)
	b := float64(argb & 0xFF)

	// Calculate luminance
	luminance := 0.299*r + 0.587*g + 0.114*b

	// Exclude very dark colors (< 10% luminance)
	if luminance < 25.5 {
		return false
	}

	// Exclude very light colors (> 90% luminance)
	if luminance > 229.5 {
		return false
	}

	// Exclude near-grayscale colors
	maxChannel := math.Max(r, math.Max(g, b))
	minChannel := math.Min(r, math.Min(g, b))

	if maxChannel-minChannel < 15 {
		return false
	}

	return true
}

// LAB represents a color in the LAB color space
type LAB struct {
	L float64 // Lightness (0-100)
	A float64 // Green-Red axis (-128 to 127)
	B float64 // Blue-Yellow axis (-128 to 127)
}

// argbToLAB converts an ARGB color to LAB color space
func argbToLAB(argb uint32) LAB {
	// Extract RGB components
	r := float64((argb>>16)&0xFF) / 255.0
	g := float64((argb>>8)&0xFF) / 255.0
	b := float64(argb&0xFF) / 255.0

	// Convert to linear RGB
	r = gammaExpand(r)
	g = gammaExpand(g)
	b = gammaExpand(b)

	// Convert to XYZ
	x := r*0.4124564 + g*0.3575761 + b*0.1804375
	y := r*0.2126729 + g*0.7151522 + b*0.0721750
	z := r*0.0193339 + g*0.1191920 + b*0.9503041

	// Normalize for D65 illuminant
	x = x / 0.95047
	y = y / 1.00000
	z = z / 1.08883

	// Convert to LAB
	fx := labF(x)
	fy := labF(y)
	fz := labF(z)

	l := 116.0*fy - 16.0
	a := 500.0 * (fx - fy)
	b2 := 200.0 * (fy - fz)

	return LAB{L: l, A: a, B: b2}
}

// gammaExpand applies gamma expansion for sRGB
func gammaExpand(channel float64) float64 {
	if channel <= 0.04045 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}

// labF is the function used in XYZ to LAB conversion
func labF(t float64) float64 {
	const delta = 6.0 / 29.0
	const deltaCubed = delta * delta * delta

	if t > deltaCubed {
		return math.Pow(t, 1.0/3.0)
	}
	return t/(3.0*delta*delta) + 4.0/29.0
}

// FindSeedColor finds the best seed color for Material You theme generation
func FindSeedColor(quantizerResult *QuantizerResult) uint32 {
	if quantizerResult == nil || len(quantizerResult.Colors) == 0 {
		// Return a default color if no colors available
		return 0xFF4285F4 // Material Blue
	}

	scorer := NewScorer()
	scores := scorer.ScoreColors(quantizerResult.Colors)

	if len(scores) == 0 {
		// If no suitable colors found, use the most prominent color
		topColors := quantizerResult.GetTopColors(1)
		if len(topColors) > 0 {
			return topColors[0]
		}
		return 0xFF4285F4 // Material Blue
	}

	// Return the highest scoring color
	return scores[0].Color
}

// FilterSimilarColors removes colors that are too similar to each other
func FilterSimilarColors(colors []uint32, minDistance float64) []uint32 {
	if len(colors) <= 1 {
		return colors
	}

	filtered := []uint32{colors[0]}

	for i := 1; i < len(colors); i++ {
		isSimilar := false
		for _, fc := range filtered {
			if Distance(colors[i], fc) < minDistance {
				isSimilar = true
				break
			}
		}

		if !isSimilar {
			filtered = append(filtered, colors[i])
		}
	}

	return filtered
}
