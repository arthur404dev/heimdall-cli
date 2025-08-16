# Heimdall Template Format Specification

## Overview

Heimdall's template system provides a flexible and powerful way to generate application-specific configuration files from color schemes. Templates support variable substitution, color transformations, conditional logic, and inheritance.

## Template Syntax

### Basic Variable Substitution

The simplest form of template variable is a direct substitution:

```
{{variable}}
```

Example:
```css
background-color: {{background}};
color: {{foreground}};
```

### Hash-Prefixed Variables

For colors that need a hash prefix (when not already present):

```
{{#variable}}
```

Example:
```css
--primary: {{#colour4}};  /* Ensures # prefix */
```

### Property Access

Access specific properties or formats of a color:

```
{{variable.property}}
```

Supported properties:
- `.rgb` - RGB format: `rgb(30, 30, 46)`
- `.rgba` - RGBA format: `rgba(30, 30, 46, 1.0)`
- `.hex` - Hex format with #: `#1e1e2e`
- `.raw` - Raw hex without #: `1e1e2e`
- `.r`, `.g`, `.b` - Individual RGB components: `30`

Example:
```css
background: {{background.rgb}};
border-color: {{outline.rgba}};
```

### Color Transformations

Apply color transformations using pipe syntax:

```
{{variable|function:argument}}
```

Available functions:
- `darken:percent` - Darken by percentage (0-100)
- `lighten:percent` - Lighten by percentage (0-100)
- `saturate:percent` - Increase saturation
- `desaturate:percent` - Decrease saturation
- `alpha:value` - Set alpha channel (0-255)
- `mix:color:ratio` - Mix with another color

Example:
```css
--background-secondary: {{background|darken:5}};
--background-hover: {{background|lighten:10}};
--text-muted: {{foreground|alpha:128}};
--accent-mixed: {{colour4|mix:colour5:50}};
```

### Conditional Blocks

Execute template sections conditionally:

```
{{if condition}}
  content when true
{{else}}
  content when false
{{end}}
```

Available conditions:
- `.dark` - True if dark mode
- `.light` - True if light mode
- Variable existence: `{{if .colours.accent}}`

Example:
```css
{{if .dark}}
  --shadow: rgba(0, 0, 0, 0.5);
{{else}}
  --shadow: rgba(0, 0, 0, 0.2);
{{end}}
```

### Iteration

Iterate over collections:

```
{{range .collection}}
  {{.key}}: {{.value}};
{{end}}
```

Example:
```css
{{range .colours}}
  --color-{{.key}}: {{.value}};
{{end}}
```

### Comments

Template comments (not included in output):

```
{{/* This is a comment */}}
```

### Default Values

Provide fallback values:

```
{{variable|default:fallback}}
```

Example:
```css
--accent: {{accent|default:#89b4fa}};
```

## Standard Variables

### Core Colors

All templates have access to these standard color variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `background` | Main background color | `#1e1e2e` |
| `foreground` | Main text color | `#cdd6f4` |
| `colour0` | Black | `#45475a` |
| `colour1` | Red | `#f38ba8` |
| `colour2` | Green | `#a6e3a1` |
| `colour3` | Yellow | `#f9e2af` |
| `colour4` | Blue | `#89b4fa` |
| `colour5` | Magenta | `#f5c2e7` |
| `colour6` | Cyan | `#94e2d5` |
| `colour7` | White | `#bac2de` |
| `colour8` | Bright Black | `#585b70` |
| `colour9` | Bright Red | `#f38ba8` |
| `colour10` | Bright Green | `#a6e3a1` |
| `colour11` | Bright Yellow | `#f9e2af` |
| `colour12` | Bright Blue | `#89b4fa` |
| `colour13` | Bright Magenta | `#f5c2e7` |
| `colour14` | Bright Cyan | `#94e2d5` |
| `colour15` | Bright White | `#a6adc8` |

### Special Colors

Optional special colors that may be present:

| Variable | Description | Example |
|----------|-------------|---------|
| `cursor` | Cursor color | `#f5e0dc` |
| `cursor_text` | Cursor text color | `#1e1e2e` |
| `selection` | Selection background | `#313244` |
| `selection_text` | Selection text | `#cdd6f4` |

### Metadata Variables

Template metadata available:

| Variable | Description | Example |
|----------|-------------|---------|
| `scheme.name` | Scheme name | `catppuccin` |
| `scheme.flavour` | Scheme variant | `mocha` |
| `scheme.mode` | Color mode | `dark` |
| `scheme.variant` | Optional variant | `blue` |

## Template Examples

### Discord CSS Template

```css
/**
 * @name Heimdall {{scheme.name|title}}
 * @description {{scheme.name}} {{scheme.flavour}} theme
 * @version 1.0.0
 * @author Heimdall CLI
 */

:root {
  /* Primary colors */
  --background-primary: {{background}};
  --background-secondary: {{background|darken:5}};
  --background-tertiary: {{background|darken:10}};
  --background-floating: {{background}};
  
  /* Text colors */
  --text-normal: {{foreground}};
  --text-muted: {{foreground|darken:30}};
  --text-link: {{colour4}};
  
  /* Interactive elements */
  --interactive-normal: {{colour7}};
  --interactive-hover: {{foreground}};
  --interactive-active: {{colour15}};
  --interactive-muted: {{colour8}};
  
  {{if .dark}}
  /* Dark mode specific */
  --elevation-low: 0 1px 0 rgba(0, 0, 0, 0.2);
  --elevation-high: 0 8px 16px rgba(0, 0, 0, 0.24);
  {{else}}
  /* Light mode specific */
  --elevation-low: 0 1px 0 rgba(0, 0, 0, 0.1);
  --elevation-high: 0 8px 16px rgba(0, 0, 0, 0.12);
  {{end}}
}
```

### GTK CSS Template

```css
/* Heimdall GTK Theme - {{scheme.name}} {{scheme.flavour}} */

{{/* Define color variables */}}
@define-color background {{background}};
@define-color foreground {{foreground}};
@define-color primary {{colour4}};
@define-color secondary {{colour5}};
@define-color error {{colour1}};
@define-color warning {{colour3}};
@define-color success {{colour2}};

{{/* Surface colors based on mode */}}
{{if .dark}}
@define-color surface {{background|lighten:5}};
@define-color surface_variant {{background|lighten:10}};
{{else}}
@define-color surface {{background|darken:5}};
@define-color surface_variant {{background|darken:10}};
{{end}}

/* Widget styling */
window {
    background-color: @background;
    color: @foreground;
}

button {
    background-color: @surface;
    color: @foreground;
    border: 1px solid {{colour8|alpha:128}};
}

button:hover {
    background-color: @surface_variant;
}

entry {
    background-color: @surface;
    color: @foreground;
    border-color: {{colour8}};
}
```

### Terminal Sequences Template

```bash
# Heimdall Terminal Color Sequences
# Scheme: {{scheme.name}}/{{scheme.flavour}}/{{scheme.mode}}

# Special colors
printf '\033]10;{{foreground}}\007'  # foreground
printf '\033]11;{{background}}\007'  # background
{{if .cursor}}
printf '\033]12;{{cursor}}\007'      # cursor
{{end}}

# Standard colors
{{range $i, $color := .colours}}
printf '\033]4;{{$i}};{{$color}}\007'  # color{{$i}}
{{end}}
```

### Btop Theme Template

```bash
# Heimdall theme for btop
# {{scheme.name}} {{scheme.flavour}}

# Main colors
theme[main_bg]="{{background}}"
theme[main_fg]="{{foreground}}"

# UI elements
theme[title]="{{foreground}}"
theme[hi_fg]="{{colour4}}"
theme[selected_bg]="{{colour8}}"
theme[selected_fg]="{{colour7}}"

# Graphs
theme[graph_text]="{{foreground}}"

# CPU
theme[cpu_box]="{{colour4}}"
theme[cpu_text]="{{colour7}}"

# Memory
theme[mem_box]="{{colour5}}"
theme[mem_text]="{{colour7}}"

# Network
theme[net_box]="{{colour6}}"
theme[net_text]="{{colour7}}"

# Process
theme[proc_box]="{{background}}"
theme[proc_text]="{{foreground}}"
```

## Template Inheritance

Templates can extend base templates using the `extends` directive:

```
{{extends "base.tmpl"}}

{{block "content"}}
  Custom content here
{{end}}
```

Base template (`base.tmpl`):
```css
/* Base theme template */
:root {
  {{block "colors"}}
  --background: {{background}};
  --foreground: {{foreground}};
  {{end}}
  
  {{block "content"}}
  /* Extended content goes here */
  {{end}}
}
```

Child template:
```css
{{extends "base.tmpl"}}

{{block "colors"}}
  {{/* Override color block */}}
  --bg: {{background}};
  --fg: {{foreground}};
  --accent: {{colour4}};
{{end}}

{{block "content"}}
  /* Additional styles */
  .custom {
    color: var(--accent);
  }
{{end}}
```

## Custom Templates

### Directory Structure

Custom templates are stored in:
```
~/.config/heimdall/templates/
├── discord/
│   └── custom.css.tmpl
├── gtk/
│   └── custom.css.tmpl
├── terminal/
│   └── custom.sh.tmpl
└── shared/
    └── base.tmpl
```

### Template Naming Convention

Templates follow this naming pattern:
```
[application]/[name].[extension].tmpl
```

Examples:
- `discord/minimal.css.tmpl`
- `gtk/material.css.tmpl`
- `terminal/osc.sh.tmpl`

### Template Selection Priority

1. User-specified template (via `--template` flag)
2. Custom template in `~/.config/heimdall/templates/`
3. Embedded default template

### Template Configuration

Specify custom templates in config:
```json
{
  "theme": {
    "templates": {
      "discord": "minimal",
      "gtk": "material",
      "terminal": "osc"
    }
  }
}
```

## Template Functions Reference

### Color Functions

| Function | Description | Example |
|----------|-------------|---------|
| `darken(color, percent)` | Darken color | `{{darken background 10}}` |
| `lighten(color, percent)` | Lighten color | `{{lighten background 10}}` |
| `saturate(color, percent)` | Increase saturation | `{{saturate colour4 20}}` |
| `desaturate(color, percent)` | Decrease saturation | `{{desaturate colour4 20}}` |
| `alpha(color, value)` | Set alpha channel | `{{alpha foreground 128}}` |
| `mix(color1, color2, ratio)` | Mix two colors | `{{mix colour4 colour5 50}}` |
| `invert(color)` | Invert color | `{{invert background}}` |
| `complement(color)` | Get complement | `{{complement colour4}}` |

### String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `upper(string)` | Convert to uppercase | `{{upper scheme.name}}` |
| `lower(string)` | Convert to lowercase | `{{lower scheme.name}}` |
| `title(string)` | Title case | `{{title scheme.name}}` |
| `replace(string, old, new)` | Replace substring | `{{replace scheme.name "-" "_"}}` |
| `trim(string)` | Trim whitespace | `{{trim " text "}}` |

### Logic Functions

| Function | Description | Example |
|----------|-------------|---------|
| `default(value, fallback)` | Provide default | `{{default accent "#89b4fa"}}` |
| `eq(a, b)` | Check equality | `{{if eq scheme.mode "dark"}}` |
| `ne(a, b)` | Check inequality | `{{if ne scheme.mode "dark"}}` |
| `and(conditions...)` | Logical AND | `{{if and .dark .accent}}` |
| `or(conditions...)` | Logical OR | `{{if or .dark .light}}` |
| `not(condition)` | Logical NOT | `{{if not .dark}}` |

## Template Validation

Templates are validated for:

1. **Syntax correctness**: Valid template syntax
2. **Variable availability**: All referenced variables exist
3. **Function validity**: Functions are called correctly
4. **Type compatibility**: Values match expected types
5. **Circular dependencies**: No infinite loops in inheritance

### Validation Command

```bash
# Validate a template
heimdall scheme template validate discord/custom.css.tmpl

# Validate all templates
heimdall scheme template validate --all
```

## Performance Considerations

### Template Caching

Parsed templates are cached in memory for performance:
- Cache size: 10MB maximum
- TTL: 5 minutes
- Invalidation: On file modification

### Processing Optimization

- Simple `{{variable}}` substitutions: < 1ms
- Complex templates with functions: < 10ms
- Large templates (>1000 lines): < 20ms

### Best Practices

1. **Use simple substitution when possible**: `{{variable}}` is faster than `{{variable|function}}`
2. **Minimize nested conditions**: Deep nesting impacts performance
3. **Cache computed values**: Store frequently used transformations
4. **Avoid excessive iterations**: Large `{{range}}` blocks can be slow
5. **Use template inheritance**: Reduce duplication and parsing overhead

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `undefined variable: accent` | Variable not in scheme | Use `{{accent\|default:"#89b4fa"}}` |
| `invalid function: darker` | Function doesn't exist | Use `darken` instead |
| `type mismatch in darken` | Wrong argument type | Ensure percent is a number |
| `circular inheritance` | Template extends itself | Fix inheritance chain |
| `template not found` | Missing template file | Check file path and name |

### Error Messages

Templates provide detailed error messages:
```
Error in discord/custom.css.tmpl at line 15:
  Invalid function call: {{background|darker:10}}
  Did you mean: {{background|darken:10}}?
```

## Migration from Other Systems

### From Caelestia Templates

Caelestia uses a similar but simpler syntax. Key differences:

| Caelestia | Heimdall | Notes |
|-----------|----------|-------|
| `{background}` | `{{background}}` | Double braces |
| `{background.rgb}` | `{{background.rgb}}` | Same property access |
| `{if dark}` | `{{if .dark}}` | Dot prefix for conditions |
| N/A | `{{background\|darken:10}}` | Color functions |

### Conversion Tool

```bash
# Convert Caelestia template to Heimdall format
heimdall migrate template caelestia.tmpl heimdall.tmpl
```

## Template Development

### Testing Templates

```bash
# Test template with a scheme
heimdall scheme template test discord/custom.css.tmpl catppuccin mocha dark

# Test with custom colors
heimdall scheme template test gtk/material.css.tmpl --colors colors.json
```

### Template Debugging

Enable debug output:
```bash
# Show variable substitutions
heimdall scheme set catppuccin mocha dark --template-debug

# Output shows:
# Template: discord/default.css.tmpl
# Variable: background = #1e1e2e
# Variable: foreground = #cdd6f4
# Function: darken(#1e1e2e, 5) = #191926
```

### Template Linting

```bash
# Lint template for issues
heimdall scheme template lint discord/custom.css.tmpl

# Output:
# ✓ Syntax valid
# ⚠ Line 23: Unused variable 'accent'
# ⚠ Line 45: Consider using default for 'cursor'
# ✓ No errors found
```

## Advanced Features

### Dynamic Includes

Include other templates dynamically:
```
{{include "shared/colors.tmpl"}}
```

### Macros

Define reusable template macros:
```
{{define "button-style"}}
  background: {{.bg}};
  color: {{.fg}};
  border: 1px solid {{.border}};
{{end}}

{{template "button-style" dict "bg" background "fg" foreground "border" colour8}}
```

### Custom Functions

Register custom template functions in Go:
```go
funcMap := template.FuncMap{
    "gradient": func(c1, c2 string) string {
        return fmt.Sprintf("linear-gradient(%s, %s)", c1, c2)
    },
}
```

Use in templates:
```css
background: {{gradient background colour4}};
```

## Conclusion

Heimdall's template system provides a powerful and flexible way to generate application-specific theme files from color schemes. With support for variable substitution, color transformations, conditional logic, and inheritance, templates can handle complex theming requirements while remaining maintainable and performant.

For more information, see:
- [Configuration Guide](CONFIGURATION.md)
- [Scheme Command Documentation](heimdall-cli-command-analysis.md)
- [Migration Guide](MIGRATION_FROM_CAELESTIA.md)