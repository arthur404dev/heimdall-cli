# GTK CSS Templates Plan

## Context

### Problem Statement
The current GTK theming implementation lacks a structured template system for generating CSS files. We need a comprehensive template architecture that supports both GTK3 and GTK4, handles all widget types, and provides consistent color mapping from heimdall colorschemes.

### Current State
- Basic GTK theme application exists but lacks template structure
- No systematic widget coverage
- Limited color variable mapping
- No version-specific handling for GTK3 vs GTK4

### Goals
- Create reusable CSS templates for all GTK widgets
- Establish clear color variable mapping system
- Support both GTK3 and GTK4 with appropriate differences
- Enable application-specific customizations
- Provide validation and optimization in the processing pipeline

### Constraints
- Must maintain compatibility with existing heimdall colorscheme format
- Templates must be maintainable and extensible
- Generated CSS must be performant
- Must handle GTK version differences gracefully

## Specification

### Functional Requirements
- Template system supporting variable substitution
- Complete widget coverage for GTK3 and GTK4
- Color mapping from heimdall schemes to GTK variables
- Application-specific template support
- CSS optimization and validation

### Non-Functional Requirements
- Templates must be human-readable and editable
- Processing must be fast (<100ms for full theme)
- Generated CSS must be valid and optimized
- System must be extensible for new widgets/applications

### Interfaces
- Input: Heimdall colorscheme JSON
- Output: GTK3/GTK4 CSS files
- Template format: CSS with variable placeholders

## CSS Template Architecture

### Template Structure for GTK3 CSS

```css
/* GTK3 Base Template Structure */
/* File: gtk3-base.css.tmpl */

/* Color Definitions */
@define-color theme_bg_color {{background}};
@define-color theme_fg_color {{foreground}};
@define-color theme_base_color {{base}};
@define-color theme_text_color {{text}};
@define-color theme_selected_bg_color {{accent}};
@define-color theme_selected_fg_color {{on_accent}};

/* Widget Defaults */
* {
    background-color: @theme_bg_color;
    color: @theme_fg_color;
}

/* Import widget-specific styles */
@import url("widgets/buttons.css");
@import url("widgets/entries.css");
@import url("widgets/lists.css");
```

### Template Structure for GTK4 CSS

```css
/* GTK4 Base Template Structure */
/* File: gtk4-base.css.tmpl */

/* CSS Variables (GTK4 prefers CSS custom properties) */
:root {
    --theme-bg-color: {{background}};
    --theme-fg-color: {{foreground}};
    --theme-base-color: {{base}};
    --theme-text-color: {{text}};
    --theme-accent-color: {{accent}};
    --theme-accent-fg-color: {{on_accent}};
    
    /* Semantic colors */
    --success-color: {{green}};
    --warning-color: {{yellow}};
    --error-color: {{red}};
}

/* Widget defaults with CSS variables */
window {
    background-color: var(--theme-bg-color);
    color: var(--theme-fg-color);
}
```

### Shared Components vs Version-Specific

#### Shared Components
- Color palette definitions
- Basic geometric properties (padding, margins)
- Font definitions
- Animation timings
- Shadow definitions

#### Version-Specific Components
- **GTK3**: Uses @define-color, specific widget names
- **GTK4**: Uses CSS custom properties, updated widget hierarchy
- **GTK3**: GtkButton, GtkEntry, GtkTreeView
- **GTK4**: button, entry, listview

### Variable System Design
```yaml
variable_types:
  color:
    prefix: "color-"
    format: "hex|rgb|rgba"
  dimension:
    prefix: "dim-"
    format: "px|em|rem"
  timing:
    prefix: "time-"
    format: "ms|s"
  
variable_scopes:
  global:
    - color-background
    - color-foreground
    - dim-border-radius
  widget:
    - button-color-bg
    - button-color-hover
  state:
    - hover-opacity
    - disabled-alpha
```

## Widget Template Catalog

### Core Widget Categories

#### 1. Buttons
```css
/* button-template.css.tmpl */
button {
    background-color: {{button_bg|default:surface}};
    color: {{button_fg|default:on_surface}};
    border: 1px solid {{button_border|default:outline}};
    border-radius: {{border_radius|default:4px}};
    padding: {{button_padding|default:6px 12px}};
}

button:hover {
    background-color: {{button_hover_bg|shade:button_bg:1.1}};
}

button:active {
    background-color: {{button_active_bg|shade:button_bg:0.9}};
}

button:disabled {
    opacity: {{disabled_opacity|default:0.5}};
}

button.suggested-action {
    background-color: {{accent}};
    color: {{on_accent}};
}

button.destructive-action {
    background-color: {{error}};
    color: {{on_error}};
}
```

#### 2. Text Entries
```css
/* entry-template.css.tmpl */
entry {
    background-color: {{entry_bg|default:base}};
    color: {{entry_fg|default:text}};
    border: 1px solid {{entry_border|default:outline}};
    border-radius: {{border_radius|default:4px}};
    padding: {{entry_padding|default:4px 8px}};
}

entry:focus {
    border-color: {{accent}};
    box-shadow: 0 0 0 1px {{accent|alpha:0.3}};
}

entry:disabled {
    background-color: {{entry_disabled_bg|shade:entry_bg:0.95}};
    color: {{entry_disabled_fg|alpha:entry_fg:0.5}};
}

entry.error {
    border-color: {{error}};
    color: {{error}};
}
```

#### 3. Lists and Trees
```css
/* list-template.css.tmpl */
treeview,
listview {
    background-color: {{list_bg|default:base}};
    color: {{list_fg|default:text}};
}

treeview:selected,
listview row:selected {
    background-color: {{accent}};
    color: {{on_accent}};
}

treeview:hover,
listview row:hover {
    background-color: {{list_hover_bg|alpha:accent:0.1}};
}

treeview.separator {
    color: {{separator|default:outline}};
}
```

#### 4. Menus and Popovers
```css
/* menu-template.css.tmpl */
menu,
popover {
    background-color: {{menu_bg|default:surface}};
    color: {{menu_fg|default:on_surface}};
    border: 1px solid {{menu_border|default:outline}};
    border-radius: {{border_radius|default:8px}};
    box-shadow: {{menu_shadow|default:0 2px 8px rgba(0,0,0,0.2)}};
}

menuitem:hover {
    background-color: {{menu_hover_bg|alpha:accent:0.1}};
}

menuitem:disabled {
    color: {{menu_disabled_fg|alpha:menu_fg:0.5}};
}
```

#### 5. Toolbars and Headers
```css
/* headerbar-template.css.tmpl */
headerbar {
    background-color: {{headerbar_bg|default:surface}};
    color: {{headerbar_fg|default:on_surface}};
    border-bottom: 1px solid {{headerbar_border|default:outline}};
}

headerbar .title {
    font-weight: bold;
    color: {{headerbar_title|default:on_surface}};
}

headerbar button {
    background-color: transparent;
    border: none;
}

headerbar button:hover {
    background-color: {{headerbar_button_hover|alpha:on_surface:0.1}};
}
```

### Complete Widget List

**Containers**: window, box, grid, stack, notebook, paned  
**Controls**: button, entry, spinbutton, scale, switch, checkbox, radio  
**Display**: label, image, progressbar, levelbar, spinner  
**Lists**: treeview, listview, listbox, flowbox  
**Menus**: menu, menubar, menuitem, popover, popovermenu  
**Dialogs**: dialog, messagedialog, filechooser, colorchooser  
**Bars**: headerbar, actionbar, toolbar, statusbar, infobar  
**Text**: textview, sourceview  
**Special**: calendar, drawingarea, glarea, video

### State Variations
```yaml
states:
  interactive:
    - normal
    - hover
    - active/pressed
    - focused
    - disabled
  selection:
    - selected
    - selected:hover
    - selected:focus
  validation:
    - error
    - warning
    - success
  special:
    - backdrop (unfocused window)
    - osd (on-screen display)
    - touch (touch mode)
```

## Color Variable Mapping

### Complete Heimdall to GTK Mapping
```yaml
heimdall_to_gtk:
  # Core colors
  background: theme_bg_color
  foreground: theme_fg_color
  base: theme_base_color
  text: theme_text_color
  surface: theme_surface_color
  on_surface: theme_on_surface_color
  
  # Accent colors
  accent: theme_selected_bg_color
  on_accent: theme_selected_fg_color
  
  # Semantic colors
  red: error_color
  green: success_color
  yellow: warning_color
  blue: info_color
  
  # UI elements
  outline: borders
  surface_variant: theme_unfocused_bg_color
  on_surface_variant: theme_unfocused_fg_color
  
  # Special mappings
  color0: terminal_black
  color1: terminal_red
  color2: terminal_green
  color3: terminal_yellow
  color4: terminal_blue
  color5: terminal_magenta
  color6: terminal_cyan
  color7: terminal_white
  color8: terminal_bright_black
  color9: terminal_bright_red
  color10: terminal_bright_green
  color11: terminal_bright_yellow
  color12: terminal_bright_blue
  color13: terminal_bright_magenta
  color14: terminal_bright_cyan
  color15: terminal_bright_white
```

### Semantic Color Definitions
```yaml
semantic_colors:
  success:
    base: "{{green}}"
    light: "{{green|shade:1.2}}"
    dark: "{{green|shade:0.8}}"
    bg: "{{green|alpha:0.1}}"
    
  warning:
    base: "{{yellow}}"
    light: "{{yellow|shade:1.2}}"
    dark: "{{yellow|shade:0.8}}"
    bg: "{{yellow|alpha:0.1}}"
    
  error:
    base: "{{red}}"
    light: "{{red|shade:1.2}}"
    dark: "{{red|shade:0.8}}"
    bg: "{{red|alpha:0.1}}"
    
  info:
    base: "{{blue}}"
    light: "{{blue|shade:1.2}}"
    dark: "{{blue|shade:0.8}}"
    bg: "{{blue|alpha:0.1}}"
```

### Shade Generation Formulas
```go
// Shade generation functions
func GenerateShade(baseColor string, factor float64) string {
    // factor > 1.0 = lighter
    // factor < 1.0 = darker
    
    if factor > 1.0 {
        // Lighten by mixing with white
        return MixColors(baseColor, "#FFFFFF", (factor - 1.0) * 0.5)
    } else {
        // Darken by mixing with black
        return MixColors(baseColor, "#000000", (1.0 - factor) * 0.5)
    }
}

func GenerateAlpha(baseColor string, alpha float64) string {
    // Convert to RGBA with specified alpha
    r, g, b := HexToRGB(baseColor)
    return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alpha)
}
```

### Contrast Ratio Calculations
```go
func CalculateContrast(fg, bg string) float64 {
    // WCAG 2.1 contrast ratio calculation
    l1 := GetRelativeLuminance(fg)
    l2 := GetRelativeLuminance(bg)
    
    lighter := math.Max(l1, l2)
    darker := math.Min(l1, l2)
    
    return (lighter + 0.05) / (darker + 0.05)
}

func EnsureContrast(fg, bg string, minRatio float64) string {
    currentRatio := CalculateContrast(fg, bg)
    if currentRatio >= minRatio {
        return fg
    }
    
    // Adjust foreground to meet minimum contrast
    // Try lightening/darkening until ratio is met
    for factor := 0.1; factor <= 1.0; factor += 0.1 {
        lighter := GenerateShade(fg, 1.0 + factor)
        if CalculateContrast(lighter, bg) >= minRatio {
            return lighter
        }
        
        darker := GenerateShade(fg, 1.0 - factor)
        if CalculateContrast(darker, bg) >= minRatio {
            return darker
        }
    }
    
    // Fallback to maximum contrast
    return CalculateContrast("#FFFFFF", bg) > CalculateContrast("#000000", bg) ? "#FFFFFF" : "#000000"
}
```

## Template Processing Pipeline

### Processing Stages

#### 1. Template Loading
```go
type TemplateLoader struct {
    baseDir     string
    cache       map[string]*Template
    includes    map[string]string
}

func (tl *TemplateLoader) Load(name string) (*Template, error) {
    // Check cache
    if cached, ok := tl.cache[name]; ok {
        return cached, nil
    }
    
    // Load from filesystem
    content, err := os.ReadFile(filepath.Join(tl.baseDir, name))
    if err != nil {
        return nil, err
    }
    
    // Parse template
    tmpl := ParseTemplate(string(content))
    
    // Resolve includes
    tmpl.ResolveIncludes(tl.includes)
    
    // Cache and return
    tl.cache[name] = tmpl
    return tmpl, nil
}
```

#### 2. Variable Substitution
```go
type VariableProcessor struct {
    scheme      *ColorScheme
    functions   map[string]VariableFunction
}

func (vp *VariableProcessor) Process(template string) string {
    // Regular expression for variable placeholders
    re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
    
    return re.ReplaceAllStringFunc(template, func(match string) string {
        // Extract variable expression
        expr := strings.Trim(match, "{}")
        
        // Parse expression (variable|function:args)
        parts := strings.Split(expr, "|")
        varName := strings.TrimSpace(parts[0])
        
        // Get base value
        value := vp.scheme.Get(varName)
        
        // Apply functions
        for i := 1; i < len(parts); i++ {
            funcCall := strings.Split(parts[i], ":")
            funcName := funcCall[0]
            funcArgs := funcCall[1:]
            
            if fn, ok := vp.functions[funcName]; ok {
                value = fn(value, funcArgs...)
            }
        }
        
        return value
    })
}
```

#### 3. CSS Optimization
```go
type CSSOptimizer struct {
    mergeSelectors   bool
    removeComments   bool
    minify          bool
}

func (co *CSSOptimizer) Optimize(css string) string {
    // Parse CSS
    stylesheet := ParseCSS(css)
    
    if co.mergeSelectors {
        // Merge duplicate selectors
        stylesheet.MergeDuplicateSelectors()
    }
    
    if co.removeComments {
        // Remove comments
        stylesheet.RemoveComments()
    }
    
    if co.minify {
        // Minify output
        return stylesheet.Minify()
    }
    
    return stylesheet.String()
}
```

#### 4. Validation
```go
type CSSValidator struct {
    gtkVersion  string
    strictMode  bool
}

func (cv *CSSValidator) Validate(css string) []ValidationError {
    errors := []ValidationError{}
    
    // Check for valid CSS syntax
    if syntaxErrors := cv.checkSyntax(css); len(syntaxErrors) > 0 {
        errors = append(errors, syntaxErrors...)
    }
    
    // Check for GTK-specific rules
    if gtkErrors := cv.checkGTKRules(css); len(gtkErrors) > 0 {
        errors = append(errors, gtkErrors...)
    }
    
    // Check for undefined variables
    if varErrors := cv.checkVariables(css); len(varErrors) > 0 {
        errors = append(errors, varErrors...)
    }
    
    return errors
}
```

### Error Handling
```go
type TemplateError struct {
    Template string
    Line     int
    Column   int
    Message  string
    Type     ErrorType
}

type ErrorType int

const (
    SyntaxError ErrorType = iota
    VariableError
    FunctionError
    ValidationError
)

func HandleTemplateError(err TemplateError) {
    // Log error with context
    log.Printf("[%s] Error in template %s at %d:%d: %s",
        err.Type, err.Template, err.Line, err.Column, err.Message)
    
    // Attempt recovery based on error type
    switch err.Type {
    case VariableError:
        // Use default value
    case FunctionError:
        // Skip function application
    case SyntaxError:
        // Skip problematic rule
    }
}
```

## Application-Specific Templates

### GNOME Shell CSS Structure
```css
/* gnome-shell.css.tmpl */

/* Panel */
#panel {
    background-color: {{panel_bg|default:surface}};
    color: {{panel_fg|default:on_surface}};
    height: {{panel_height|default:32px}};
}

#panel .panel-button {
    color: {{panel_button_fg|default:on_surface}};
}

#panel .panel-button:hover {
    background-color: {{panel_button_hover|alpha:on_surface:0.1}};
}

/* Overview */
#overview {
    background-color: {{overview_bg|shade:background:0.9}};
}

.window-clone {
    background-color: {{window_bg|default:surface}};
    border: 2px solid {{window_border|default:outline}};
}

/* Dash */
#dash {
    background-color: {{dash_bg|alpha:surface:0.9}};
    border-radius: {{dash_radius|default:16px}};
}

.dash-item-container .app-icon {
    color: {{dash_icon|default:on_surface}};
}

/* Notifications */
.notification-banner {
    background-color: {{notification_bg|default:surface}};
    color: {{notification_fg|default:on_surface}};
    border-radius: {{notification_radius|default:8px}};
}
```

### Nautilus Customizations
```css
/* nautilus.css.tmpl */

/* Sidebar */
.nautilus-window .sidebar {
    background-color: {{sidebar_bg|default:surface_variant}};
}

.nautilus-window .sidebar row:selected {
    background-color: {{accent}};
    color: {{on_accent}};
}

/* Path bar */
.nautilus-window .path-bar button {
    background-color: {{pathbar_bg|default:surface}};
    border: 1px solid {{pathbar_border|default:outline}};
}

/* File view */
.nautilus-window .view {
    background-color: {{view_bg|default:base}};
}

.nautilus-window .view .thumbnail {
    background-color: {{thumbnail_bg|default:surface}};
    border: 1px solid {{thumbnail_border|default:outline}};
}
```

### Terminal Applications
```css
/* terminal.css.tmpl */

/* GNOME Terminal */
terminal-window {
    background-color: {{terminal_bg|default:color0}};
}

terminal-window .terminal-screen {
    background-color: {{terminal_bg|default:color0}};
    color: {{terminal_fg|default:color7}};
}

/* Tabs */
terminal-window notebook tab {
    background-color: {{tab_bg|default:surface}};
    color: {{tab_fg|default:on_surface}};
}

terminal-window notebook tab:checked {
    background-color: {{tab_active_bg|default:base}};
    color: {{tab_active_fg|default:text}};
}
```

### Other Popular GTK Applications
```yaml
application_templates:
  gedit:
    file: "gedit.css.tmpl"
    selectors:
      - ".gedit-window"
      - ".gedit-side-panel"
      - ".gedit-document-panel"
      
  evince:
    file: "evince.css.tmpl"
    selectors:
      - ".evince-window"
      - ".ev-sidebar"
      - ".ev-view"
      
  rhythmbox:
    file: "rhythmbox.css.tmpl"
    selectors:
      - ".rhythmbox-window"
      - ".rb-source-list"
      - ".rb-player-controls"
      
  transmission:
    file: "transmission.css.tmpl"
    selectors:
      - ".transmission-window"
      - ".torrent-list"
      - ".peer-view"
```

## Template Examples

### Button Template - Before Processing
```css
/* button.css.tmpl */
button {
    background-color: {{button_bg|default:surface}};
    color: {{button_fg|default:on_surface}};
    border: 1px solid {{button_border|shade:button_bg:0.8}};
    border-radius: {{radius_small|default:4px}};
    padding: {{padding_medium|default:6px 12px}};
    transition: all {{transition_fast|default:150ms}} ease;
}

button:hover {
    background-color: {{button_bg|shade:1.1}};
    border-color: {{accent|alpha:0.5}};
}

button:active {
    background-color: {{button_bg|shade:0.9}};
}

button.suggested-action {
    background-color: {{accent}};
    color: {{on_accent|contrast:accent:4.5}};
    border-color: {{accent|shade:0.8}};
}
```

### Button Template - After Processing
```css
/* Generated button.css */
button {
    background-color: #2e3440;
    color: #d8dee9;
    border: 1px solid #252a33;
    border-radius: 4px;
    padding: 6px 12px;
    transition: all 150ms ease;
}

button:hover {
    background-color: #333947;
    border-color: rgba(136, 192, 208, 0.5);
}

button:active {
    background-color: #2a2f3a;
}

button.suggested-action {
    background-color: #88c0d0;
    color: #2e3440;
    border-color: #6da9ba;
}
```

### Complex Widget Styling - Notebook
```css
/* notebook.css.tmpl */
notebook {
    background-color: {{notebook_bg|default:background}};
}

notebook > header {
    background-color: {{header_bg|default:surface}};
    border-bottom: 1px solid {{header_border|default:outline}};
}

notebook > header > tabs > tab {
    background-color: {{tab_bg|default:surface_variant}};
    color: {{tab_fg|alpha:on_surface:0.7}};
    border: 1px solid transparent;
    padding: {{tab_padding|default:4px 12px}};
    margin: 0 2px;
}

notebook > header > tabs > tab:hover {
    background-color: {{tab_hover_bg|shade:tab_bg:1.05}};
    color: {{tab_hover_fg|default:on_surface}};
}

notebook > header > tabs > tab:checked {
    background-color: {{tab_active_bg|default:base}};
    color: {{tab_active_fg|default:text}};
    border-color: {{tab_active_border|default:outline}};
    border-bottom-color: {{tab_active_bg|default:base}};
}

notebook > stack {
    background-color: {{stack_bg|default:base}};
    border: 1px solid {{stack_border|default:outline}};
    border-top: none;
}
```

### Advanced Color Functions
```css
/* advanced-colors.css.tmpl */

/* Using contrast function */
.high-contrast-text {
    color: {{foreground|contrast:background:7.0}};
}

/* Using mix function */
.subtle-accent {
    background-color: {{accent|mix:background:0.1}};
}

/* Using shade with conditions */
.adaptive-surface {
    background-color: {{surface|shade:is_dark?0.95:1.05}};
}

/* Chaining multiple functions */
.complex-color {
    background-color: {{accent|shade:1.2|alpha:0.3|mix:surface:0.5}};
}
```

## Implementation Plan

### Phase 1: Core Template System
- [ ] Design template file structure
  - Define directory hierarchy
  - Create naming conventions
  - Set up template inheritance system
- [ ] Implement variable processor
  - Basic variable substitution
  - Default value handling
  - Function system (shade, alpha, mix, contrast)
- [ ] Create base templates
  - GTK3 base template
  - GTK4 base template
  - Color definition templates

### Phase 2: Widget Templates
- [ ] Implement core widget templates
  - Buttons (all variants)
  - Entries and text inputs
  - Lists and trees
  - Menus and popovers
- [ ] Add state variations
  - Hover, active, disabled states
  - Focus indicators
  - Selection states
- [ ] Create specialized widgets
  - Headers and toolbars
  - Notebooks and stacks
  - Dialogs and windows

### Phase 3: Color System
- [ ] Implement color mapping
  - Heimdall to GTK variable mapping
  - Semantic color generation
  - Shade calculation algorithms
- [ ] Add contrast management
  - WCAG contrast calculation
  - Automatic contrast adjustment
  - Accessibility validation
- [ ] Create color functions
  - Mix, shade, alpha functions
  - Conditional color selection
  - Color space conversions

### Phase 4: Processing Pipeline
- [ ] Build template loader
  - File loading and caching
  - Include resolution
  - Template inheritance
- [ ] Implement CSS optimizer
  - Selector merging
  - Dead code elimination
  - Minification options
- [ ] Add validation system
  - CSS syntax validation
  - GTK-specific rule checking
  - Variable reference validation

### Phase 5: Application Templates
- [ ] Create GNOME Shell templates
  - Panel styling
  - Overview and dash
  - Notifications
- [ ] Add application-specific templates
  - Nautilus file manager
  - Terminal applications
  - Text editors (gedit, etc.)
- [ ] Implement template discovery
  - Auto-detection of installed apps
  - Dynamic template loading
  - User override system

### Phase 6: Testing and Documentation
- [ ] Create test suite
  - Unit tests for processors
  - Integration tests for pipeline
  - Visual regression tests
- [ ] Write documentation
  - Template authoring guide
  - Variable reference
  - Function documentation
- [ ] Add examples
  - Complete theme examples
  - Custom template tutorials
  - Migration guides

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| GTK version incompatibilities | High | Maintain separate template sets, version detection |
| Performance with large templates | Medium | Implement caching, lazy loading, optimization |
| Color contrast issues | High | Automatic contrast adjustment, validation tools |
| Application-specific quirks | Medium | Extensive testing, community feedback, override system |
| Template maintenance burden | Medium | Clear documentation, automated testing, modular design |

## Success Metrics

- **Coverage**: 100% of standard GTK widgets styled
- **Performance**: <100ms processing time for complete theme
- **Compatibility**: Works with GTK 3.24+ and GTK 4.0+
- **Quality**: All generated CSS passes validation
- **Accessibility**: All color combinations meet WCAG AA standards
- **Maintainability**: New widgets can be added in <30 minutes
- **Adoption**: Successfully themes 20+ popular GTK applications

## Dev Log

### Session: Initial Planning
- Created comprehensive GTK CSS template plan
- Defined template architecture for GTK3 and GTK4
- Established widget catalog with all major GTK widgets
- Designed color variable mapping system
- Outlined processing pipeline with optimization and validation
- Created application-specific template structures
- Provided concrete examples of template processing
- Set up 6-phase implementation roadmap