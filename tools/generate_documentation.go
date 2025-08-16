//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arthur404dev/heimdall-cli/internal/config"
)

func main() {
	// Initialize the metadata registry
	if err := config.InitializeRegistry(); err != nil {
		log.Fatalf("Failed to initialize metadata registry: %v", err)
	}

	// Validate completeness - fail build if descriptions are missing
	missing := config.MetadataRegistry.ValidateCompleteness()
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "ERROR: The following configuration fields are missing descriptions:\n")
		for _, field := range missing {
			fmt.Fprintf(os.Stderr, "  - %s\n", field)
		}
		fmt.Fprintf(os.Stderr, "\nPlease add 'desc' tags to all configuration fields.\n")
		os.Exit(1)
	}

	// Generate comprehensive configuration reference
	if err := generateConfigReference(); err != nil {
		log.Fatalf("Failed to generate configuration reference: %v", err)
	}

	// Generate JSON schema
	if err := generateJSONSchema(); err != nil {
		log.Fatalf("Failed to generate JSON schema: %v", err)
	}

	// Generate quick reference guide
	if err := generateQuickReference(); err != nil {
		log.Fatalf("Failed to generate quick reference: %v", err)
	}

	fmt.Println("✓ Documentation generation complete")
	fmt.Println("  - docs/CONFIG_REFERENCE.md")
	fmt.Println("  - docs/CONFIG_QUICK_REFERENCE.md")
	fmt.Println("  - docs/examples/config-schema.json")
}

func generateConfigReference() error {
	fields := config.MetadataRegistry.GetAllFields()

	// Group fields by category
	categories := make(map[string][]*config.FieldMetadata)
	categoryOrder := []string{}
	categoryDescriptions := map[string]string{
		"version":    "Configuration version management",
		"theme":      "Theme application settings for various applications",
		"scheme":     "Color scheme management and generation",
		"wallpaper":  "Wallpaper management and Material You integration",
		"idle":       "Idle detection and automatic theme switching",
		"discord":    "Discord Rich Presence integration",
		"quickshell": "Quickshell panel and widget theming",
		"paths":      "Custom paths for configuration files",
	}

	// Collect and sort fields
	for path, field := range fields {
		parts := strings.Split(path, ".")
		if len(parts) > 0 {
			category := parts[0]
			if _, exists := categories[category]; !exists {
				categoryOrder = append(categoryOrder, category)
				categories[category] = []*config.FieldMetadata{}
			}
			categories[category] = append(categories[category], field)
		}
	}

	// Sort categories
	sort.Strings(categoryOrder)

	// Build the documentation
	var sb strings.Builder

	// Header
	sb.WriteString("# Heimdall CLI Configuration Reference\n\n")
	sb.WriteString("This document provides a comprehensive reference for all configuration options available in heimdall-cli.\n\n")
	sb.WriteString("## Table of Contents\n\n")

	// Generate TOC
	for _, category := range categoryOrder {
		title := strings.Title(category)
		sb.WriteString(fmt.Sprintf("- [%s Configuration](#%s-configuration)\n", title, category))
	}
	sb.WriteString("\n")

	// Quick start section
	sb.WriteString("## Quick Start\n\n")
	sb.WriteString("Heimdall CLI uses sensible defaults for all configuration options. You only need to create a configuration file if you want to customize the behavior.\n\n")
	sb.WriteString("### Minimal Configuration\n\n")
	sb.WriteString("Create a file at `~/.config/heimdall/config.json` with only the settings you want to change:\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"scheme\": {\n")
	sb.WriteString("    \"default\": \"catppuccin-mocha\"\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")
	sb.WriteString("All other settings will use their default values.\n\n")

	// Configuration sections
	for _, category := range categoryOrder {
		fields := categories[category]

		// Sort fields by path
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Path < fields[j].Path
		})

		// Section header
		title := strings.Title(category)
		sb.WriteString(fmt.Sprintf("## %s Configuration\n\n", title))

		// Category description
		if desc, exists := categoryDescriptions[category]; exists {
			sb.WriteString(fmt.Sprintf("%s\n\n", desc))
		}

		// Build tree structure
		tree := buildFieldTree(fields, category)
		writeFieldTree(&sb, tree, 0)
	}

	// Footer sections
	sb.WriteString("## Default Values\n\n")
	sb.WriteString("To see all default values, run:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("heimdall config defaults --show\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Validation\n\n")
	sb.WriteString("Heimdall CLI validates your configuration on load. To check if your configuration is valid:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("heimdall config validate\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Environment Variables\n\n")
	sb.WriteString("You can override configuration values using environment variables:\n\n")
	sb.WriteString("- `HEIMDALL_CONFIG_PATH`: Override the configuration file location\n")
	sb.WriteString("- `HEIMDALL_SCHEME`: Override the default color scheme\n")
	sb.WriteString("- `HEIMDALL_DEBUG`: Enable debug logging\n\n")

	sb.WriteString("## Examples\n\n")
	sb.WriteString("See the [examples directory](examples/) for various configuration examples:\n\n")
	sb.WriteString("- [Minimal theme configuration](examples/minimal-theme-only.json)\n")
	sb.WriteString("- [Material You wallpaper theming](examples/minimal-material-you.json)\n")
	sb.WriteString("- [Quickshell integration](examples/minimal-quickshell.json)\n")
	sb.WriteString("- [Full configuration with all options](examples/config-full-example.json)\n\n")

	// Write to file
	outputPath := filepath.Join("docs", "CONFIG_REFERENCE.md")
	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

type fieldNode struct {
	Field    *config.FieldMetadata
	Children map[string]*fieldNode
}

func buildFieldTree(fields []*config.FieldMetadata, rootCategory string) *fieldNode {
	root := &fieldNode{
		Children: make(map[string]*fieldNode),
	}

	for _, field := range fields {
		// Skip container objects
		if field.Type == "object" && len(field.Children) > 0 {
			continue
		}

		parts := strings.Split(field.Path, ".")
		if len(parts) == 0 || parts[0] != rootCategory {
			continue
		}

		// Navigate/create the tree
		current := root
		for i := 1; i < len(parts); i++ {
			part := parts[i]
			if _, exists := current.Children[part]; !exists {
				current.Children[part] = &fieldNode{
					Children: make(map[string]*fieldNode),
				}
			}
			current = current.Children[part]
		}
		current.Field = field
	}

	return root
}

func writeFieldTree(sb *strings.Builder, node *fieldNode, depth int) {
	// Sort children for consistent output
	childKeys := make([]string, 0, len(node.Children))
	for key := range node.Children {
		childKeys = append(childKeys, key)
	}
	sort.Strings(childKeys)

	for _, key := range childKeys {
		child := node.Children[key]

		if child.Field != nil {
			// Write field documentation
			writeFieldDoc(sb, child.Field, depth)
		}

		// Recurse for children
		if len(child.Children) > 0 {
			writeFieldTree(sb, child, depth+1)
		}
	}
}

func writeFieldDoc(sb *strings.Builder, field *config.FieldMetadata, depth int) {
	// Field name as header
	headerLevel := "###"
	for i := 0; i < depth; i++ {
		headerLevel += "#"
	}

	sb.WriteString(fmt.Sprintf("%s `%s`\n\n", headerLevel, field.Path))

	// Description
	if field.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", field.Description))
	}

	// Field details table
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| **Type** | `%s` |\n", field.Type))

	if field.Default != "" {
		// Format default value for display
		defaultVal := field.Default
		if field.Type == "bool" || field.Type == "int" || field.Type == "float" {
			sb.WriteString(fmt.Sprintf("| **Default** | `%s` |\n", defaultVal))
		} else {
			sb.WriteString(fmt.Sprintf("| **Default** | `\"%s\"` |\n", defaultVal))
		}
	}

	if field.Required {
		sb.WriteString("| **Required** | Yes |\n")
	}

	if field.Deprecated != "" {
		sb.WriteString(fmt.Sprintf("| **⚠️ Deprecated** | %s |\n", field.Deprecated))
	}

	sb.WriteString("\n")

	// Example if present
	if field.Example != "" {
		sb.WriteString("**Example:**\n\n")
		sb.WriteString("```json\n")

		// Format example based on path
		parts := strings.Split(field.Path, ".")
		indent := ""
		sb.WriteString("{\n")

		for i, part := range parts {
			indent += "  "
			if i < len(parts)-1 {
				sb.WriteString(fmt.Sprintf("%s\"%s\": {\n", indent, part))
			} else {
				// Format value based on type
				if field.Type == "bool" || field.Type == "int" || field.Type == "float" {
					sb.WriteString(fmt.Sprintf("%s\"%s\": %s\n", indent, part, field.Example))
				} else if strings.HasPrefix(field.Type, "[]") {
					sb.WriteString(fmt.Sprintf("%s\"%s\": %s\n", indent, part, field.Example))
				} else {
					sb.WriteString(fmt.Sprintf("%s\"%s\": \"%s\"\n", indent, part, field.Example))
				}
			}
		}

		// Close braces
		for i := len(parts) - 1; i > 0; i-- {
			indent = strings.Repeat("  ", i)
			sb.WriteString(fmt.Sprintf("%s}\n", indent))
		}
		sb.WriteString("}\n")
		sb.WriteString("```\n\n")
	}
}

func generateJSONSchema() error {
	schema := config.MetadataRegistry.GenerateJSONSchema()

	// Add additional schema properties
	schema["$id"] = "https://github.com/heimdall-cli/heimdall-cli/blob/main/docs/examples/config-schema.json"
	schema["additionalProperties"] = false

	// Convert to JSON with indentation
	jsonData, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON schema: %w", err)
	}

	// Write to file
	outputPath := filepath.Join("docs", "examples", "config-schema.json")
	return os.WriteFile(outputPath, jsonData, 0644)
}

func generateQuickReference() error {
	fields := config.MetadataRegistry.GetAllFields()

	var sb strings.Builder

	// Header
	sb.WriteString("# Heimdall CLI Configuration Quick Reference\n\n")
	sb.WriteString("A quick reference guide for common configuration options.\n\n")

	// Common configurations section
	sb.WriteString("## Common Configurations\n\n")

	sb.WriteString("### Set Default Color Scheme\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"scheme\": {\n")
	sb.WriteString("    \"default\": \"catppuccin-mocha\"\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Enable Material You Theming\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"scheme\": {\n")
	sb.WriteString("    \"materialYou\": true\n")
	sb.WriteString("  },\n")
	sb.WriteString("  \"wallpaper\": {\n")
	sb.WriteString("    \"generateMaterialYou\": true\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Disable Specific Applications\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"theme\": {\n")
	sb.WriteString("    \"enableGtk\": false,\n")
	sb.WriteString("    \"enableDiscord\": false\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Configure Idle Detection\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"idle\": {\n")
	sb.WriteString("    \"enabled\": true,\n")
	sb.WriteString("    \"timeout\": 300,\n")
	sb.WriteString("    \"scheme\": \"rosepine\",\n")
	sb.WriteString("    \"wallpaper\": \"/path/to/idle-wallpaper.jpg\"\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	// All options table
	sb.WriteString("## All Configuration Options\n\n")
	sb.WriteString("| Path | Type | Default | Description |\n")
	sb.WriteString("|------|------|---------|-------------|\n")

	// Sort fields by path
	var sortedPaths []string
	for path := range fields {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Strings(sortedPaths)

	for _, path := range sortedPaths {
		field := fields[path]

		// Skip container objects
		if field.Type == "object" && len(field.Children) > 0 {
			continue
		}

		// Format description (truncate if too long)
		desc := field.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}

		// Format default value
		defaultVal := field.Default
		if defaultVal == "" {
			defaultVal = "-"
		} else if len(defaultVal) > 20 {
			defaultVal = defaultVal[:17] + "..."
		}

		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
			field.Path, field.Type, defaultVal, desc))
	}

	sb.WriteString("\n")

	// Commands section
	sb.WriteString("## Useful Commands\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# Show all configuration options\n")
	sb.WriteString("heimdall config list\n\n")
	sb.WriteString("# Search for specific options\n")
	sb.WriteString("heimdall config search theme\n\n")
	sb.WriteString("# Show current effective configuration\n")
	sb.WriteString("heimdall config effective\n\n")
	sb.WriteString("# Show only modified values\n")
	sb.WriteString("heimdall config list --modified\n\n")
	sb.WriteString("# Describe a specific option\n")
	sb.WriteString("heimdall config describe scheme.default\n\n")
	sb.WriteString("# Validate configuration\n")
	sb.WriteString("heimdall config validate\n")
	sb.WriteString("```\n\n")

	// Write to file
	outputPath := filepath.Join("docs", "CONFIG_QUICK_REFERENCE.md")
	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}
