package theme

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Engine is the template processing engine for theming
type Engine struct {
	templates map[string]*template.Template
	funcs     template.FuncMap
}

// NewEngine creates a new template engine
func NewEngine() *Engine {
	e := &Engine{
		templates: make(map[string]*template.Template),
	}

	// Set up template functions
	e.funcs = template.FuncMap{
		// Color manipulation functions
		"hex":     e.toHex,
		"rgb":     e.toRGB,
		"rgba":    e.toRGBA,
		"hsl":     e.toHSL,
		"hsla":    e.toHSLA,
		"lighten": e.lighten,
		"darken":  e.darken,
		"alpha":   e.alpha,

		// String manipulation
		"upper":   strings.ToUpper,
		"lower":   strings.ToLower,
		"title":   strings.Title,
		"replace": strings.ReplaceAll,

		// Conditional helpers
		"isDark":  e.isDark,
		"isLight": e.isLight,
	}

	return e
}

// LoadTemplate loads a template from string
func (e *Engine) LoadTemplate(name, content string) error {
	tmpl, err := template.New(name).Funcs(e.funcs).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	e.templates[name] = tmpl
	return nil
}

// LoadTemplateFile loads a template from a file
func (e *Engine) LoadTemplateFile(name, path string) error {
	tmpl, err := template.New(name).Funcs(e.funcs).ParseFiles(path)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", path, err)
	}

	e.templates[name] = tmpl
	return nil
}

// Render renders a template with the given data
func (e *Engine) Render(name string, data interface{}) (string, error) {
	tmpl, ok := e.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderString renders a template string directly
func (e *Engine) RenderString(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("inline").Funcs(e.funcs).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse inline template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute inline template: %w", err)
	}

	return buf.String(), nil
}

// Template function implementations

// toHex converts a color to hex format
func (e *Engine) toHex(color interface{}) string {
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			return c
		}
		return "#" + c
	case map[string]interface{}:
		if hex, ok := c["hex"].(string); ok {
			return hex
		}
		if rgb, ok := c["rgb"].(map[string]interface{}); ok {
			r := toInt(rgb["r"])
			g := toInt(rgb["g"])
			b := toInt(rgb["b"])
			return fmt.Sprintf("#%02x%02x%02x", r, g, b)
		}
	}
	return "#000000"
}

// toRGB converts a color to RGB format
func (e *Engine) toRGB(color interface{}) string {
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			hex := strings.TrimPrefix(c, "#")
			if len(hex) == 6 {
				r, _ := parseHexByte(hex[0:2])
				g, _ := parseHexByte(hex[2:4])
				b, _ := parseHexByte(hex[4:6])
				return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
			}
		}
	case map[string]interface{}:
		if rgb, ok := c["rgb"].(map[string]interface{}); ok {
			r := toInt(rgb["r"])
			g := toInt(rgb["g"])
			b := toInt(rgb["b"])
			return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
		}
	}
	return "rgb(0, 0, 0)"
}

// toRGBA converts a color to RGBA format with alpha
func (e *Engine) toRGBA(color interface{}, alpha float64) string {
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			hex := strings.TrimPrefix(c, "#")
			if len(hex) == 6 {
				r, _ := parseHexByte(hex[0:2])
				g, _ := parseHexByte(hex[2:4])
				b, _ := parseHexByte(hex[4:6])
				return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alpha)
			}
		}
	case map[string]interface{}:
		if rgb, ok := c["rgb"].(map[string]interface{}); ok {
			r := toInt(rgb["r"])
			g := toInt(rgb["g"])
			b := toInt(rgb["b"])
			return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alpha)
		}
	}
	return fmt.Sprintf("rgba(0, 0, 0, %.2f)", alpha)
}

// toHSL converts a color to HSL format
func (e *Engine) toHSL(color interface{}) string {
	// Implementation would convert to HSL
	// For now, return a placeholder
	return "hsl(0, 0%, 0%)"
}

// toHSLA converts a color to HSLA format with alpha
func (e *Engine) toHSLA(color interface{}, alpha float64) string {
	// Implementation would convert to HSLA
	// For now, return a placeholder
	return fmt.Sprintf("hsla(0, 0%%, 0%%, %.2f)", alpha)
}

// lighten lightens a color by a percentage
func (e *Engine) lighten(color interface{}, percent float64) string {
	// Implementation would lighten the color
	// For now, return the original color
	return e.toHex(color)
}

// darken darkens a color by a percentage
func (e *Engine) darken(color interface{}, percent float64) string {
	// Implementation would darken the color
	// For now, return the original color
	return e.toHex(color)
}

// alpha adds alpha channel to a color
func (e *Engine) alpha(color interface{}, alpha float64) string {
	return e.toRGBA(color, alpha)
}

// isDark checks if a color is dark
func (e *Engine) isDark(color interface{}) bool {
	// Simple luminance check
	switch c := color.(type) {
	case string:
		if strings.HasPrefix(c, "#") {
			hex := strings.TrimPrefix(c, "#")
			if len(hex) == 6 {
				r, _ := parseHexByte(hex[0:2])
				g, _ := parseHexByte(hex[2:4])
				b, _ := parseHexByte(hex[4:6])
				luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
				return luminance < 128
			}
		}
	}
	return false
}

// isLight checks if a color is light
func (e *Engine) isLight(color interface{}) bool {
	return !e.isDark(color)
}

// Helper functions

func parseHexByte(s string) (uint8, error) {
	var b uint8
	_, err := fmt.Sscanf(s, "%02x", &b)
	return b, err
}

func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case uint8:
		return int(val)
	default:
		return 0
	}
}
