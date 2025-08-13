package theme

import (
	"testing"
)

func TestSimpleReplacer_ReplaceString(t *testing.T) {
	replacer := NewSimpleReplacer()

	tests := []struct {
		name     string
		template string
		colors   map[string]string
		expected string
	}{
		{
			name:     "single color replacement",
			template: "background: {{colour0}};",
			colors: map[string]string{
				"colour0": "1a1b26",
			},
			expected: "background: 1a1b26;",
		},
		{
			name:     "multiple color replacements",
			template: "color: {{colour1}}; background: {{colour0}};",
			colors: map[string]string{
				"colour0": "1a1b26",
				"colour1": "f7768e",
			},
			expected: "color: f7768e; background: 1a1b26;",
		},
		{
			name:     "no replacements needed",
			template: "static content without placeholders",
			colors: map[string]string{
				"colour0": "1a1b26",
			},
			expected: "static content without placeholders",
		},
		{
			name:     "placeholder not in colors map",
			template: "color: {{colour99}};",
			colors: map[string]string{
				"colour0": "1a1b26",
			},
			expected: "color: {{colour99}};", // Should remain unchanged
		},
		{
			name:     "special colors",
			template: "background: {{background}}; foreground: {{foreground}}; cursor: {{cursor}};",
			colors: map[string]string{
				"background": "1a1b26",
				"foreground": "c0caf5",
				"cursor":     "c0caf5",
			},
			expected: "background: 1a1b26; foreground: c0caf5; cursor: c0caf5;",
		},
		{
			name:     "all 16 standard colors",
			template: "{{colour0}} {{colour1}} {{colour2}} {{colour3}} {{colour4}} {{colour5}} {{colour6}} {{colour7}} {{colour8}} {{colour9}} {{colour10}} {{colour11}} {{colour12}} {{colour13}} {{colour14}} {{colour15}}",
			colors: map[string]string{
				"colour0":  "1a1b26",
				"colour1":  "f7768e",
				"colour2":  "9ece6a",
				"colour3":  "e0af68",
				"colour4":  "7aa2f7",
				"colour5":  "bb9af7",
				"colour6":  "7dcfff",
				"colour7":  "c0caf5",
				"colour8":  "414868",
				"colour9":  "f7768e",
				"colour10": "9ece6a",
				"colour11": "e0af68",
				"colour12": "7aa2f7",
				"colour13": "bb9af7",
				"colour14": "7dcfff",
				"colour15": "c0caf5",
			},
			expected: "1a1b26 f7768e 9ece6a e0af68 7aa2f7 bb9af7 7dcfff c0caf5 414868 f7768e 9ece6a e0af68 7aa2f7 bb9af7 7dcfff c0caf5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacer.ReplaceString(tt.template, tt.colors)
			if result != tt.expected {
				t.Errorf("ReplaceString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSimpleReplacer_ReplaceTemplate(t *testing.T) {
	replacer := NewSimpleReplacer()

	template := `# Heimdall theme
background={{background}}
foreground={{foreground}}
color0={{colour0}}
color1={{colour1}}`

	colors := map[string]string{
		"background": "1a1b26",
		"foreground": "c0caf5",
		"colour0":    "1a1b26",
		"colour1":    "f7768e",
	}

	expected := `# Heimdall theme
background=1a1b26
foreground=c0caf5
color0=1a1b26
color1=f7768e`

	result, err := replacer.ReplaceTemplate(template, colors)
	if err != nil {
		t.Errorf("ReplaceTemplate() error = %v", err)
		return
	}

	if result != expected {
		t.Errorf("ReplaceTemplate() = %v, want %v", result, expected)
	}
}

func TestSimpleReplacer_CaelestiaCompatibility(t *testing.T) {
	// Test that our simple replacer works exactly like caelestia's approach
	replacer := NewSimpleReplacer()

	// Caelestia-style template with simple {{key}} replacements
	template := `/* Discord theme */
:root {
    --primary: {{colour4}};
    --secondary: {{colour5}};
    --background: {{background}};
    --foreground: {{foreground}};
}`

	colors := map[string]string{
		"colour4":    "7aa2f7",
		"colour5":    "bb9af7",
		"background": "1a1b26",
		"foreground": "c0caf5",
	}

	expected := `/* Discord theme */
:root {
    --primary: 7aa2f7;
    --secondary: bb9af7;
    --background: 1a1b26;
    --foreground: c0caf5;
}`

	result, err := replacer.ReplaceTemplate(template, colors)
	if err != nil {
		t.Errorf("ReplaceTemplate() error = %v", err)
		return
	}

	if result != expected {
		t.Errorf("Caelestia compatibility test failed.\nGot:\n%s\nWant:\n%s", result, expected)
	}
}

// Benchmark to ensure simple replacement is fast
func BenchmarkSimpleReplacer_ReplaceString(b *testing.B) {
	replacer := NewSimpleReplacer()
	template := "background: {{background}}; color: {{colour1}}; border: {{colour4}};"
	colors := map[string]string{
		"background": "1a1b26",
		"colour1":    "f7768e",
		"colour4":    "7aa2f7",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replacer.ReplaceString(template, colors)
	}
}
