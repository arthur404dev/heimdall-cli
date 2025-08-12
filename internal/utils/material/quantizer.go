package material

import (
	"image"
	"image/color"
	"math"
	"sort"
)

// QuantizerResult represents the result of color quantization
type QuantizerResult struct {
	Colors map[uint32]int // Map of ARGB colors to their pixel counts
}

// Quantizer performs color quantization on images
type Quantizer struct {
	maxColors int
}

// NewQuantizer creates a new quantizer with the specified maximum colors
func NewQuantizer(maxColors int) *Quantizer {
	if maxColors <= 0 {
		maxColors = 128
	}
	return &Quantizer{maxColors: maxColors}
}

// Quantize performs Wu's color quantization algorithm on an image
func (q *Quantizer) Quantize(img image.Image) *QuantizerResult {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Build color histogram
	histogram := make(map[uint32]int)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			argb := colorToARGB(c)
			histogram[argb]++
		}
	}

	// If we have fewer unique colors than maxColors, return them all
	if len(histogram) <= q.maxColors {
		return &QuantizerResult{Colors: histogram}
	}

	// Perform Wu's quantization
	return q.wuQuantize(histogram, width*height)
}

// wuQuantize implements Wu's color quantization algorithm
func (q *Quantizer) wuQuantize(histogram map[uint32]int, totalPixels int) *QuantizerResult {
	// Create color cube
	cube := newColorCube()

	// Add all colors to the cube
	for argb, count := range histogram {
		r := uint8((argb >> 16) & 0xFF)
		g := uint8((argb >> 8) & 0xFF)
		b := uint8(argb & 0xFF)
		cube.addColor(r, g, b, count)
	}

	// Build 3D color histogram
	cube.computeMoments()

	// Create initial box
	boxes := []*colorBox{cube.createBox(0, 0, 0, 31, 31, 31)}

	// Split boxes until we have maxColors
	for len(boxes) < q.maxColors && len(boxes) < len(histogram) {
		// Find box with maximum variance
		maxVariance := 0.0
		maxIndex := -1

		for i, box := range boxes {
			if box.canSplit() {
				variance := box.variance()
				if variance > maxVariance {
					maxVariance = variance
					maxIndex = i
				}
			}
		}

		if maxIndex == -1 {
			break // No more boxes can be split
		}

		// Split the box
		box1, box2 := boxes[maxIndex].split()
		boxes[maxIndex] = box1
		boxes = append(boxes, box2)
	}

	// Extract colors from boxes
	result := &QuantizerResult{
		Colors: make(map[uint32]int),
	}

	for _, box := range boxes {
		if box.volume() > 0 {
			avgColor := box.averageColor()
			result.Colors[avgColor] = box.weight()
		}
	}

	return result
}

// colorCube represents a 3D histogram for color quantization
type colorCube struct {
	weights  [32][32][32]int
	momentsR [32][32][32]int
	momentsG [32][32][32]int
	momentsB [32][32][32]int
	moments  [32][32][32]float64
}

func newColorCube() *colorCube {
	return &colorCube{}
}

func (c *colorCube) addColor(r, g, b uint8, weight int) {
	// Quantize to 5 bits per channel
	ir := int(r >> 3)
	ig := int(g >> 3)
	ib := int(b >> 3)

	c.weights[ir][ig][ib] += weight
	c.momentsR[ir][ig][ib] += weight * int(r)
	c.momentsG[ir][ig][ib] += weight * int(g)
	c.momentsB[ir][ig][ib] += weight * int(b)
}

func (c *colorCube) computeMoments() {
	// Compute cumulative moments
	for r := 0; r < 32; r++ {
		for g := 0; g < 32; g++ {
			for b := 0; b < 32; b++ {
				weight := c.weights[r][g][b]
				if weight > 0 {
					mr := c.momentsR[r][g][b]
					mg := c.momentsG[r][g][b]
					mb := c.momentsB[r][g][b]
					c.moments[r][g][b] = float64(mr*mr+mg*mg+mb*mb) / float64(weight)
				}
			}
		}
	}
}

func (c *colorCube) createBox(r0, g0, b0, r1, g1, b1 int) *colorBox {
	return &colorBox{
		cube: c,
		r0:   r0, g0: g0, b0: b0,
		r1: r1, g1: g1, b1: b1,
	}
}

// colorBox represents a box in color space
type colorBox struct {
	cube       *colorCube
	r0, g0, b0 int
	r1, g1, b1 int
}

func (b *colorBox) volume() int {
	return (b.r1 - b.r0 + 1) * (b.g1 - b.g0 + 1) * (b.b1 - b.b0 + 1)
}

func (b *colorBox) weight() int {
	weight := 0
	for r := b.r0; r <= b.r1; r++ {
		for g := b.g0; g <= b.g1; g++ {
			for bl := b.b0; bl <= b.b1; bl++ {
				weight += b.cube.weights[r][g][bl]
			}
		}
	}
	return weight
}

func (b *colorBox) canSplit() bool {
	return b.r1 > b.r0 || b.g1 > b.g0 || b.b1 > b.b0
}

func (b *colorBox) variance() float64 {
	if !b.canSplit() {
		return 0
	}

	weight := b.weight()
	if weight == 0 {
		return 0
	}

	var sumR, sumG, sumB float64
	var sumR2, sumG2, sumB2 float64

	for r := b.r0; r <= b.r1; r++ {
		for g := b.g0; g <= b.g1; g++ {
			for bl := b.b0; bl <= b.b1; bl++ {
				w := float64(b.cube.weights[r][g][bl])
				if w > 0 {
					mr := float64(b.cube.momentsR[r][g][bl])
					mg := float64(b.cube.momentsG[r][g][bl])
					mb := float64(b.cube.momentsB[r][g][bl])

					sumR += mr
					sumG += mg
					sumB += mb
					sumR2 += mr * mr / w
					sumG2 += mg * mg / w
					sumB2 += mb * mb / w
				}
			}
		}
	}

	fw := float64(weight)
	variance := sumR2 - sumR*sumR/fw
	variance += sumG2 - sumG*sumG/fw
	variance += sumB2 - sumB*sumB/fw

	return variance
}

func (b *colorBox) split() (*colorBox, *colorBox) {
	// Find the longest axis
	dr := b.r1 - b.r0
	dg := b.g1 - b.g0
	db := b.b1 - b.b0

	var splitAxis int
	if dr >= dg && dr >= db {
		splitAxis = 0 // Red
	} else if dg >= db {
		splitAxis = 1 // Green
	} else {
		splitAxis = 2 // Blue
	}

	// Find split point
	switch splitAxis {
	case 0: // Red
		mid := (b.r0 + b.r1) / 2
		return b.cube.createBox(b.r0, b.g0, b.b0, mid, b.g1, b.b1),
			b.cube.createBox(mid+1, b.g0, b.b0, b.r1, b.g1, b.b1)
	case 1: // Green
		mid := (b.g0 + b.g1) / 2
		return b.cube.createBox(b.r0, b.g0, b.b0, b.r1, mid, b.b1),
			b.cube.createBox(b.r0, mid+1, b.b0, b.r1, b.g1, b.b1)
	default: // Blue
		mid := (b.b0 + b.b1) / 2
		return b.cube.createBox(b.r0, b.g0, b.b0, b.r1, b.g1, mid),
			b.cube.createBox(b.r0, b.g0, mid+1, b.r1, b.g1, b.b1)
	}
}

func (b *colorBox) averageColor() uint32 {
	if b.weight() == 0 {
		return 0
	}

	var sumR, sumG, sumB, sumW int

	for r := b.r0; r <= b.r1; r++ {
		for g := b.g0; g <= b.g1; g++ {
			for bl := b.b0; bl <= b.b1; bl++ {
				w := b.cube.weights[r][g][bl]
				if w > 0 {
					sumR += b.cube.momentsR[r][g][bl]
					sumG += b.cube.momentsG[r][g][bl]
					sumB += b.cube.momentsB[r][g][bl]
					sumW += w
				}
			}
		}
	}

	if sumW == 0 {
		return 0
	}

	avgR := uint8(sumR / sumW)
	avgG := uint8(sumG / sumW)
	avgB := uint8(sumB / sumW)

	return 0xFF000000 | uint32(avgR)<<16 | uint32(avgG)<<8 | uint32(avgB)
}

// colorToARGB converts a color.Color to ARGB format
func colorToARGB(c color.Color) uint32 {
	r, g, b, a := c.RGBA()
	// Convert from 16-bit to 8-bit
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)
	a8 := uint8(a >> 8)

	return uint32(a8)<<24 | uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)
}

// GetTopColors returns the top N colors by pixel count
func (qr *QuantizerResult) GetTopColors(n int) []uint32 {
	type colorCount struct {
		color uint32
		count int
	}

	var colors []colorCount
	for c, count := range qr.Colors {
		colors = append(colors, colorCount{c, count})
	}

	// Sort by count descending
	sort.Slice(colors, func(i, j int) bool {
		return colors[i].count > colors[j].count
	})

	// Return top N
	result := make([]uint32, 0, n)
	for i := 0; i < n && i < len(colors); i++ {
		result = append(result, colors[i].color)
	}

	return result
}

// Distance calculates the Euclidean distance between two ARGB colors
func Distance(c1, c2 uint32) float64 {
	r1 := float64((c1 >> 16) & 0xFF)
	g1 := float64((c1 >> 8) & 0xFF)
	b1 := float64(c1 & 0xFF)

	r2 := float64((c2 >> 16) & 0xFF)
	g2 := float64((c2 >> 8) & 0xFF)
	b2 := float64(c2 & 0xFF)

	dr := r1 - r2
	dg := g1 - g2
	db := b1 - b2

	return math.Sqrt(dr*dr + dg*dg + db*db)
}
