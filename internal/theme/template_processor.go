package theme

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// templateProcessor implements the TemplateProcessor interface
type templateProcessor struct {
	simplePattern *regexp.Regexp
	funcs         template.FuncMap
	validator     *Validator
}

// NewTemplateProcessor creates a new template processor
func NewTemplateProcessor() TemplateProcessor {
	tp := &templateProcessor{
		simplePattern: regexp.MustCompile(`\{\{([^}]+)\}\}`),
		validator:     NewValidator(),
	}

	// Set up template functions for advanced processing
	tp.funcs = template.FuncMap{
		// Color manipulation
		"darken":  tp.darkenColor,
		"lighten": tp.lightenColor,
		"alpha":   tp.addAlpha,
		"hex":     tp.toHex,
		"rgb":     tp.toRGB,
		"rgba":    tp.toRGBA,
		"noHash":  tp.removeHash,

		// String manipulation
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"replace": strings.ReplaceAll,
		"trim":    strings.TrimSpace,

		// Conditionals
		"isDark":  tp.isDark,
		"isLight": tp.isLight,
	}

	return tp
}

// ProcessSimple performs simple {{variable}} replacements
func (tp *templateProcessor) ProcessSimple(templateStr string, colors map[string]string) (string, error) {
	if err := tp.validator.ValidateTemplate(templateStr); err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	result := tp.simplePattern.ReplaceAllStringFunc(templateStr, func(match string) string {
		// Extract variable name
		varName := strings.Trim(match, "{}")
		varName = strings.TrimSpace(varName)

		// Handle special prefixes
		if strings.HasPrefix(varName, "#") {
			// Variable with hash prefix
			varName = strings.TrimPrefix(varName, "#")
			if color, ok := colors[varName]; ok {
				if !strings.HasPrefix(color, "#") {
					return "#" + color
				}
				return color
			}
		}

		// Handle property access (e.g., color.rgb)
		if strings.Contains(varName, ".") {
			parts := strings.SplitN(varName, ".", 2)
			if len(parts) == 2 {
				if color, ok := colors[parts[0]]; ok {
					return tp.processProperty(color, parts[1])
				}
			}
		}

		// Handle function calls (e.g., color|darken:10)
		if strings.Contains(varName, "|") {
			parts := strings.SplitN(varName, "|", 2)
			if len(parts) == 2 {
				if color, ok := colors[parts[0]]; ok {
					return tp.processFunction(color, parts[1])
				}
			}
		}

		// Simple replacement
		if color, ok := colors[varName]; ok {
			return color
		}

		// Return unchanged if not found
		return match
	})

	return result, nil
}

// ProcessAdvanced performs advanced template processing with conditionals
func (tp *templateProcessor) ProcessAdvanced(name, templateStr string, data TemplateData) (string, error) {
	if err := tp.validator.ValidateTemplate(templateStr); err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}

	// Parse template with functions
	tmpl, err := template.New(name).Funcs(tp.funcs).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ValidateTemplate validates template syntax
func (tp *templateProcessor) ValidateTemplate(templateStr string) error {
	return tp.validator.ValidateTemplate(templateStr)
}

// processProperty processes property access like color.rgb
func (tp *templateProcessor) processProperty(color, property string) string {
	switch property {
	case "hex":
		return tp.toHex(color)
	case "rgb":
		return tp.toRGB(color)
	case "rgba":
		return tp.toRGBA(color, 1.0)
	case "noHash":
		return tp.removeHash(color)
	default:
		return color
	}
}

// processFunction processes function calls like color|darken:10
func (tp *templateProcessor) processFunction(color, function string) string {
	parts := strings.SplitN(function, ":", 2)
	funcName := parts[0]

	switch funcName {
	case "darken":
		if len(parts) == 2 {
			var percent float64
			fmt.Sscanf(parts[1], "%f", &percent)
			return tp.darkenColor(color, percent)
		}
		return tp.darkenColor(color, 10)
	case "lighten":
		if len(parts) == 2 {
			var percent float64
			fmt.Sscanf(parts[1], "%f", &percent)
			return tp.lightenColor(color, percent)
		}
		return tp.lightenColor(color, 10)
	case "alpha":
		if len(parts) == 2 {
			var alpha float64
			fmt.Sscanf(parts[1], "%f", &alpha)
			return tp.addAlpha(color, alpha)
		}
		return tp.addAlpha(color, 0.8)
	case "hex":
		return tp.toHex(color)
	case "rgb":
		return tp.toRGB(color)
	case "noHash":
		return tp.removeHash(color)
	default:
		return color
	}
}

// Color manipulation functions

func (tp *templateProcessor) toHex(color interface{}) string {
	switch c := color.(type) {
	case string:
		if !strings.HasPrefix(c, "#") && len(c) == 6 {
			return "#" + c
		}
		return c
	default:
		return "#000000"
	}
}

func (tp *templateProcessor) toRGB(color interface{}) string {
	hex := tp.toHex(color)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return "rgb(0, 0, 0)"
	}

	var r, g, b int64
	r, _ = strconv.ParseInt(hex[0:2], 16, 64)
	g, _ = strconv.ParseInt(hex[2:4], 16, 64)
	b, _ = strconv.ParseInt(hex[4:6], 16, 64)

	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

func (tp *templateProcessor) toRGBA(color interface{}, alpha float64) string {
	hex := tp.toHex(color)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return fmt.Sprintf("rgba(0, 0, 0, %.2f)", alpha)
	}

	var r, g, b int64
	r, _ = strconv.ParseInt(hex[0:2], 16, 64)
	g, _ = strconv.ParseInt(hex[2:4], 16, 64)
	b, _ = strconv.ParseInt(hex[4:6], 16, 64)

	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alpha)
}

func (tp *templateProcessor) removeHash(color interface{}) string {
	switch c := color.(type) {
	case string:
		return strings.TrimPrefix(c, "#")
	default:
		return "000000"
	}
}

func (tp *templateProcessor) addAlpha(color interface{}, alpha float64) string {
	return tp.toRGBA(color, alpha)
}

func (tp *templateProcessor) darkenColor(color interface{}, percent float64) string {
	hex := tp.toHex(color)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return "#" + hex
	}

	var r, g, b int64
	r, _ = strconv.ParseInt(hex[0:2], 16, 64)
	g, _ = strconv.ParseInt(hex[2:4], 16, 64)
	b, _ = strconv.ParseInt(hex[4:6], 16, 64)

	factor := 1.0 - (percent / 100.0)
	r = int64(float64(r) * factor)
	g = int64(float64(g) * factor)
	b = int64(float64(b) * factor)

	// Clamp values
	if r < 0 {
		r = 0
	}
	if g < 0 {
		g = 0
	}
	if b < 0 {
		b = 0
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func (tp *templateProcessor) lightenColor(color interface{}, percent float64) string {
	hex := tp.toHex(color)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return "#" + hex
	}

	var r, g, b int64
	r, _ = strconv.ParseInt(hex[0:2], 16, 64)
	g, _ = strconv.ParseInt(hex[2:4], 16, 64)
	b, _ = strconv.ParseInt(hex[4:6], 16, 64)

	// Calculate lightening
	factor := percent / 100.0
	r = r + int64(float64(255-r)*factor)
	g = g + int64(float64(255-g)*factor)
	b = b + int64(float64(255-b)*factor)

	// Clamp values
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func (tp *templateProcessor) isDark(color interface{}) bool {
	hex := tp.toHex(color)
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return false
	}

	var r, g, b int64
	r, _ = strconv.ParseInt(hex[0:2], 16, 64)
	g, _ = strconv.ParseInt(hex[2:4], 16, 64)
	b, _ = strconv.ParseInt(hex[4:6], 16, 64)

	// Calculate luminance
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255
	return luminance < 0.5
}

func (tp *templateProcessor) isLight(color interface{}) bool {
	return !tp.isDark(color)
}
