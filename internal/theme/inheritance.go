package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateInheritance manages template inheritance and composition
type TemplateInheritance struct {
	registry      *TemplateRegistry
	maxDepth      int
	resolvedCache map[string]string
}

// NewTemplateInheritance creates a new template inheritance manager
func NewTemplateInheritance(registry *TemplateRegistry) *TemplateInheritance {
	return &TemplateInheritance{
		registry:      registry,
		maxDepth:      3,
		resolvedCache: make(map[string]string),
	}
}

// ResolveTemplate resolves a template with inheritance
func (ti *TemplateInheritance) ResolveTemplate(app, templateName string) (string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", app, templateName)
	if cached, ok := ti.resolvedCache[cacheKey]; ok {
		return cached, nil
	}

	// Get the template content
	content, err := ti.registry.GetTemplate(app, templateName)
	if err != nil {
		return "", err
	}

	// Resolve inheritance
	resolved, err := ti.resolveInheritance(content, 0, make(map[string]bool))
	if err != nil {
		return "", fmt.Errorf("failed to resolve inheritance: %w", err)
	}

	// Cache the result
	ti.resolvedCache[cacheKey] = resolved
	return resolved, nil
}

// resolveInheritance recursively resolves template inheritance
func (ti *TemplateInheritance) resolveInheritance(content string, depth int, visited map[string]bool) (string, error) {
	// Check depth limit
	if depth >= ti.maxDepth {
		return "", fmt.Errorf("maximum inheritance depth (%d) exceeded", ti.maxDepth)
	}

	// Look for extends directive
	if !strings.Contains(content, "{{extends") {
		return content, nil
	}

	// Parse extends directive
	lines := strings.Split(content, "\n")
	var extendsPath string
	var contentStart int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "{{extends") && strings.HasSuffix(trimmed, "}}") {
			// Extract template path
			extendsPath = strings.TrimSpace(trimmed[9 : len(trimmed)-2])
			extendsPath = strings.Trim(extendsPath, `"'`)
			contentStart = i + 1
			break
		}
	}

	if extendsPath == "" {
		return content, nil
	}

	// Check for circular dependency
	if visited[extendsPath] {
		return "", fmt.Errorf("circular dependency detected: %s", extendsPath)
	}
	visited[extendsPath] = true

	// Load parent template
	parentContent, err := ti.loadParentTemplate(extendsPath)
	if err != nil {
		return "", fmt.Errorf("failed to load parent template %s: %w", extendsPath, err)
	}

	// Recursively resolve parent's inheritance
	parentResolved, err := ti.resolveInheritance(parentContent, depth+1, visited)
	if err != nil {
		return "", err
	}

	// Extract blocks from child template
	childContent := strings.Join(lines[contentStart:], "\n")
	childBlocks := ti.extractBlocks(childContent)

	// Replace blocks in parent with child blocks
	result := ti.replaceBlocks(parentResolved, childBlocks)

	return result, nil
}

// extractBlocks extracts named blocks from template content
func (ti *TemplateInheritance) extractBlocks(content string) map[string]string {
	blocks := make(map[string]string)

	// Simple block extraction using regex-like approach
	lines := strings.Split(content, "\n")
	var currentBlock string
	var blockContent []string
	inBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for block start
		if strings.HasPrefix(trimmed, "{{block") && strings.Contains(trimmed, "}}") {
			// Extract block name
			parts := strings.Split(trimmed, `"`)
			if len(parts) >= 2 {
				currentBlock = parts[1]
				inBlock = true
				blockContent = []string{}
				continue
			}
		}

		// Check for block end
		if inBlock && trimmed == "{{end}}" {
			blocks[currentBlock] = strings.Join(blockContent, "\n")
			inBlock = false
			currentBlock = ""
			continue
		}

		// Collect block content
		if inBlock {
			blockContent = append(blockContent, line)
		}
	}

	return blocks
}

// replaceBlocks replaces blocks in parent template with child blocks
func (ti *TemplateInheritance) replaceBlocks(parent string, childBlocks map[string]string) string {
	result := parent

	for blockName, blockContent := range childBlocks {
		// Find and replace the block in parent
		blockStart := fmt.Sprintf("{{block \"%s\"}}", blockName)
		blockEnd := "{{end}}"

		startIdx := strings.Index(result, blockStart)
		if startIdx == -1 {
			continue
		}

		// Find the corresponding end tag
		endIdx := strings.Index(result[startIdx:], blockEnd)
		if endIdx == -1 {
			continue
		}
		endIdx += startIdx + len(blockEnd)

		// Replace the block
		before := result[:startIdx]
		after := result[endIdx:]
		result = before + blockStart + "\n" + blockContent + "\n" + blockEnd + after
	}

	return result
}

// loadParentTemplate loads a parent template from file or registry
func (ti *TemplateInheritance) loadParentTemplate(path string) (string, error) {
	// If path starts with "shared/" or similar, load from custom templates
	if strings.Contains(path, "/") {
		customPath := filepath.Join(ti.registry.customDir, path)
		if !strings.HasSuffix(customPath, ".tmpl") {
			customPath += ".tmpl"
		}

		if content, err := os.ReadFile(customPath); err == nil {
			return string(content), nil
		}
	}

	// Try to load from registry (embedded or custom)
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		app := parts[0]
		templateName := "default"
		if len(parts) > 1 {
			templateName = strings.TrimSuffix(parts[1], ".tmpl")
		}

		return ti.registry.GetTemplate(app, templateName)
	}

	return "", fmt.Errorf("parent template not found: %s", path)
}

// ClearCache clears the resolved template cache
func (ti *TemplateInheritance) ClearCache() {
	ti.resolvedCache = make(map[string]string)
}

// SetMaxDepth sets the maximum inheritance depth
func (ti *TemplateInheritance) SetMaxDepth(depth int) {
	if depth > 0 && depth <= 10 {
		ti.maxDepth = depth
	}
}

// ValidateInheritance validates that a template's inheritance chain is valid
func (ti *TemplateInheritance) ValidateInheritance(content string) error {
	_, err := ti.resolveInheritance(content, 0, make(map[string]bool))
	return err
}

// Example of template inheritance usage:
//
// Base template (shared/base.tmpl):
// ```
// /* Base theme template */
// :root {
//   {{block "colors"}}
//   --background: {{background}};
//   --foreground: {{foreground}};
//   {{end}}
//
//   {{block "custom"}}
//   /* Custom styles go here */
//   {{end}}
// }
// ```
//
// Child template (discord/custom.tmpl):
// ```
// {{extends "shared/base"}}
//
// {{block "colors"}}
//   --bg: {{background}};
//   --fg: {{foreground}};
//   --accent: {{colour4}};
// {{end}}
//
// {{block "custom"}}
//   .custom-class {
//     color: var(--accent);
//   }
// {{end}}
// ```
