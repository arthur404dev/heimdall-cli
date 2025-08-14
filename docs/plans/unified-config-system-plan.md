# Unified Configuration Management System Plan

## Context

### Problem Statement
The current heimdall-cli configuration system has grown organically, resulting in a monolithic structure that mixes CLI tool settings with shell configuration, making it difficult to extend and maintain. External tools like Quickshell need to provide their own configuration schemas, but there's no mechanism for heimdall-cli to dynamically read and validate against these schemas.

### Current State
- Single monolithic config.go file with all configuration domains mixed together
- Hard-coded Go structs that must be manually updated for schema changes
- No separation between CLI tool configuration and shell/external configurations
- No mechanism to read external JSON schemas from other tools
- Limited extensibility for future configuration types

### Goals
- Create a unified configuration package that manages multiple configuration files
- Enable dynamic schema extraction from external JSON files
- Separate configuration domains (CLI, shell, external tools)
- Provide a consistent command interface across all configuration types
- Support schema versioning and migration
- Maintain backward compatibility

### Constraints
- Must preserve existing configuration data during migration
- Cannot break existing command interfaces
- Must support both JSON and TOML formats
- Performance must remain under 50ms for validation
- Must handle concurrent access safely

## Specification

### Functional Requirements

#### Configuration Management
- FR1: Support multiple configuration files in ~/.config/heimdall/
- FR2: Dynamically read and parse JSON schema files from external sources
- FR3: Generate Go structs from JSON schemas at runtime
- FR4: Validate configurations against their schemas
- FR5: Support schema versioning and migration
- FR6: Provide unified command interface for all configs

#### Schema Processing
- FR7: Extract schemas from external tools' default.json files
- FR8: Convert JSON Schema to Go validation rules
- FR9: Support nested object and array validation
- FR10: Handle schema evolution and backward compatibility

#### Command Operations
- FR11: Provide get/set operations for any configuration domain
- FR12: Support path-based property access (e.g., "shell.appearance.theme")
- FR13: List all available configuration domains
- FR14: Validate configurations before saving
- FR15: Support atomic file operations

### Non-Functional Requirements

#### Performance
- NFR1: Configuration validation < 50ms
- NFR2: Schema parsing < 100ms
- NFR3: File operations must be atomic
- NFR4: Support lazy loading of configurations

#### Reliability
- NFR5: Zero data loss during migrations
- NFR6: Automatic backup before modifications
- NFR7: Rollback capability on failures
- NFR8: Concurrent access protection

#### Usability
- NFR9: Consistent command interface across all configs
- NFR10: Clear error messages with suggestions
- NFR11: Interactive mode for complex operations
- NFR12: Comprehensive help documentation

### Interfaces

#### Command Line Interface
```bash
heimdall config [domain] [operation] [args]
```

#### Configuration Provider Interface
```go
type ConfigProvider interface {
    GetSchema() (*Schema, error)
    Validate(config interface{}) error
    Get(path string) (interface{}, error)
    Set(path string, value interface{}) error
    Migrate(from, to string) error
}
```

#### Schema Registry Interface
```go
type SchemaRegistry interface {
    Register(domain string, schemaPath string) error
    GetSchema(domain string) (*Schema, error)
    ListDomains() []string
    ValidateAgainstSchema(domain string, config interface{}) error
}
```

## Implementation Plan

### Phase 1: Core Architecture
- [x] Design unified config package structure
  - Acceptance criteria: Clear separation of concerns
  - Test requirements: Unit tests for each component
- [x] Implement ConfigManager as central coordinator
  - Acceptance criteria: Manages multiple config providers
  - Test requirements: Integration tests for provider coordination
- [x] Create ConfigProvider interface
  - Acceptance criteria: Extensible for new config types
  - Test requirements: Mock provider tests
- [x] Build Schema registry system
  - Acceptance criteria: Dynamic schema registration
  - Test requirements: Schema CRUD operations

### Phase 2: JSON Schema Support
- [x] Implement JSON Schema parser
  - Acceptance criteria: Full JSON Schema Draft 7 support
  - Test requirements: Schema validation test suite
- [x] Create schema-to-Go struct generator
  - Acceptance criteria: Runtime struct generation
  - Test requirements: Complex schema conversion tests
- [x] Build validation engine
  - Acceptance criteria: Complete property validation
  - Test requirements: Edge case validation tests
- [x] Add schema versioning support
  - Acceptance criteria: Version compatibility checking
  - Test requirements: Migration path tests

### Phase 3: Configuration Providers
- [x] Implement CLIConfigProvider
  - Acceptance criteria: Manages config.json
  - Test requirements: CRUD operations
- [x] Implement ShellConfigProvider
  - Acceptance criteria: Manages shell.json with external schema
  - Test requirements: Schema extraction tests
- [x] Create ExtensibleConfigProvider
  - Acceptance criteria: Generic provider for future configs
  - Test requirements: Dynamic configuration tests
- [x] Build provider registration system
  - Acceptance criteria: Runtime provider registration
  - Test requirements: Provider lifecycle tests

### Phase 4: Command Structure
- [x] Refactor `heimdall config` command
  - Acceptance criteria: Unified interface for all domains
  - Test requirements: Command parsing tests
- [x] Implement domain-specific subcommands
  - Acceptance criteria: Consistent operations across domains
  - Test requirements: Subcommand execution tests
- [x] Add schema inspection commands
  - Acceptance criteria: View and validate schemas
  - Test requirements: Schema introspection tests
- [ ] Create migration commands
  - Acceptance criteria: Safe schema migrations
  - Test requirements: Migration scenario tests

### Phase 5: Schema Extraction
- [x] Build external schema reader
  - Acceptance criteria: Read schemas from file paths or URLs
  - Test requirements: Various source tests
- [x] Implement schema caching
  - Acceptance criteria: Efficient schema reuse
  - Test requirements: Cache invalidation tests
- [x] Create schema discovery mechanism
  - Acceptance criteria: Auto-discover available schemas
  - Test requirements: Discovery algorithm tests
- [x] Add schema validation
  - Acceptance criteria: Validate schema format
  - Test requirements: Invalid schema handling

### Phase 6: Migration System
- [ ] Design migration interface
  - Acceptance criteria: Clean migration API
  - Test requirements: Migration contract tests
- [ ] Implement version comparison
  - Acceptance criteria: Semantic versioning support
  - Test requirements: Version ordering tests
- [ ] Create migration strategies
  - Acceptance criteria: Multiple migration approaches
  - Test requirements: Strategy selection tests
- [ ] Build rollback mechanism
  - Acceptance criteria: Safe rollback on failure
  - Test requirements: Rollback scenario tests

## Risks and Mitigations

### Technical Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Schema incompatibility | High | Medium | Implement schema version negotiation |
| Performance degradation | Medium | Low | Use caching and lazy loading |
| Data loss during migration | High | Low | Automatic backups, atomic operations |
| Complex schema handling | Medium | Medium | Incremental schema support |

### Implementation Risks
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Breaking changes | High | Medium | Comprehensive testing, gradual rollout |
| External schema changes | Medium | High | Schema version pinning |
| User confusion | Medium | Medium | Clear documentation, intuitive commands |

## Success Metrics

### Technical Metrics
- Schema validation accuracy: 100%
- Migration success rate: > 99%
- Performance benchmarks met: 100%
- Test coverage: > 80%

### User Metrics
- Command consistency score: > 90%
- Error message clarity: > 85% helpful
- Documentation completeness: 100%
- User adoption rate: > 70% in 3 months

## Dev Log

### Session: 2025-08-12 14:30
- Created unified configuration system plan
- Defined extensible architecture with ConfigProvider interface
- Designed JSON Schema integration approach
- Established schema extraction mechanism
- Planned migration strategy
- Next steps: Begin Phase 1 implementation with core architecture

### Session: 2025-08-12 15:45
- Implemented complete unified configuration system
- Created modular architecture with ConfigManager, Providers, and Schema packages
- Built JSON Schema Draft 7 parser and validator
- Implemented CLI and Shell configuration providers
- Added support for external schema loading from Quickshell configs
- Created comprehensive config command with all operations
- Added environment variable support for configuration paths
- Implemented atomic file operations for safe config updates
- Added support for output paths (e.g., writing shell config to Quickshell directory)
- Created documentation for the new system
- Completed Phases 1-5 (except migration commands)
- System is fully functional and ready for use

---

## Technical Specifications

### JSON Schema Format Design

The system will support JSON Schema Draft 7 with extensions for Quickshell-specific metadata:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://quickshell.org/schemas/heimdall/shell/v1.0.0",
  "title": "Heimdall Shell Configuration",
  "description": "Configuration schema for Quickshell's Heimdall implementation",
  "version": "1.0.0",
  "type": "object",
  "required": ["version", "appearance", "bar"],
  "properties": {
    "version": {
      "type": "string",
      "description": "Configuration version",
      "pattern": "^\\d+\\.\\d+\\.\\d+$",
      "default": "1.0.0"
    },
    "appearance": {
      "type": "object",
      "description": "Visual appearance settings",
      "properties": {
        "colorScheme": {
          "type": "string",
          "description": "Color scheme identifier",
          "default": "catppuccin-mocha",
          "enum": ["catppuccin-mocha", "gruvbox-dark", "nord", "dracula"],
          "x-heimdall-mutable": true
        },
        "fontSize": {
          "type": "integer",
          "description": "Base font size in pixels",
          "minimum": 8,
          "maximum": 32,
          "default": 12
        },
        "animations": {
          "type": "boolean",
          "description": "Enable UI animations",
          "default": true,
          "x-heimdall-since": "1.1.0"
        }
      },
      "required": ["colorScheme"],
      "additionalProperties": false
    },
    "bar": {
      "type": "object",
      "description": "Status bar configuration",
      "properties": {
        "position": {
          "type": "string",
          "enum": ["top", "bottom"],
          "default": "top"
        },
        "height": {
          "type": "integer",
          "minimum": 20,
          "maximum": 100,
          "default": 30
        },
        "modules": {
          "type": "array",
          "description": "Active bar modules",
          "items": {
            "type": "object",
            "properties": {
              "name": {
                "type": "string",
                "description": "Module identifier"
              },
              "enabled": {
                "type": "boolean",
                "default": true
              },
              "config": {
                "type": "object",
                "description": "Module-specific configuration",
                "additionalProperties": true
              }
            },
            "required": ["name"]
          },
          "default": []
        }
      }
    }
  },
  "x-heimdall-metadata": {
    "provider": "quickshell",
    "configFile": "shell.json",
    "migrations": {
      "0.9.0": "migrations/0.9.0-to-1.0.0.json"
    },
    "deprecated": {
      "appearance.theme": {
        "since": "1.0.0",
        "use": "appearance.colorScheme"
      }
    }
  }
}
```

### Go Code Architecture

#### ConfigManager - Central Coordinator
```go
package config

import (
    "sync"
    "github.com/arthur404dev/heimdall-cli/internal/config/providers"
    "github.com/arthur404dev/heimdall-cli/internal/config/schema"
)

// ConfigManager coordinates all configuration providers
type ConfigManager struct {
    providers map[string]providers.ConfigProvider
    registry  *schema.Registry
    mu        sync.RWMutex
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
    cm := &ConfigManager{
        providers: make(map[string]providers.ConfigProvider),
        registry:  schema.NewRegistry(),
    }
    
    // Register default providers
    cm.RegisterProvider("cli", providers.NewCLIProvider())
    cm.RegisterProvider("shell", providers.NewShellProvider())
    
    return cm
}

// RegisterProvider adds a new configuration provider
func (cm *ConfigManager) RegisterProvider(domain string, provider providers.ConfigProvider) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()
    
    // Extract and register schema
    schema, err := provider.GetSchema()
    if err != nil {
        return fmt.Errorf("failed to get schema: %w", err)
    }
    
    if err := cm.registry.Register(domain, schema); err != nil {
        return fmt.Errorf("failed to register schema: %w", err)
    }
    
    cm.providers[domain] = provider
    return nil
}

// Get retrieves a configuration value
func (cm *ConfigManager) Get(domain, path string) (interface{}, error) {
    cm.mu.RLock()
    provider, exists := cm.providers[domain]
    cm.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("unknown configuration domain: %s", domain)
    }
    
    return provider.Get(path)
}

// Set updates a configuration value
func (cm *ConfigManager) Set(domain, path string, value interface{}) error {
    cm.mu.RLock()
    provider, exists := cm.providers[domain]
    cm.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("unknown configuration domain: %s", domain)
    }
    
    // Validate against schema
    if err := cm.registry.ValidateValue(domain, path, value); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    return provider.Set(path, value)
}

// ListDomains returns all registered configuration domains
func (cm *ConfigManager) ListDomains() []string {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    
    domains := make([]string, 0, len(cm.providers))
    for domain := range cm.providers {
        domains = append(domains, domain)
    }
    return domains
}
```

#### ConfigProvider Interface
```go
package providers

import "github.com/arthur404dev/heimdall-cli/internal/config/schema"

// ConfigProvider defines the interface for configuration providers
type ConfigProvider interface {
    // GetSchema returns the JSON schema for this provider
    GetSchema() (*schema.Schema, error)
    
    // Load reads the configuration from storage
    Load() error
    
    // Save writes the configuration to storage
    Save() error
    
    // Get retrieves a value by path (e.g., "appearance.theme")
    Get(path string) (interface{}, error)
    
    // Set updates a value by path
    Set(path string, value interface{}) error
    
    // Validate checks the entire configuration
    Validate() error
    
    // Migrate upgrades configuration to a new version
    Migrate(targetVersion string) error
    
    // GetConfigPath returns the file path for this configuration
    GetConfigPath() string
}
```

#### Schema Registry
```go
package schema

import (
    "encoding/json"
    "sync"
)

// Registry manages configuration schemas
type Registry struct {
    schemas map[string]*Schema
    mu      sync.RWMutex
}

// Schema represents a JSON schema with metadata
type Schema struct {
    Raw        json.RawMessage        `json:"-"`
    Properties map[string]Property    `json:"properties"`
    Required   []string              `json:"required"`
    Version    string                `json:"version"`
    Metadata   map[string]interface{} `json:"x-heimdall-metadata"`
}

// Property represents a schema property
type Property struct {
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    Default     interface{}            `json:"default,omitempty"`
    Enum        []interface{}          `json:"enum,omitempty"`
    Properties  map[string]Property    `json:"properties,omitempty"`
    Items       *Property              `json:"items,omitempty"`
    Minimum     *float64               `json:"minimum,omitempty"`
    Maximum     *float64               `json:"maximum,omitempty"`
    Pattern     string                 `json:"pattern,omitempty"`
    Required    []string               `json:"required,omitempty"`
}

// NewRegistry creates a new schema registry
func NewRegistry() *Registry {
    return &Registry{
        schemas: make(map[string]*Schema),
    }
}

// Register adds a schema to the registry
func (r *Registry) Register(domain string, schema *Schema) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    // Validate schema structure
    if err := schema.Validate(); err != nil {
        return fmt.Errorf("invalid schema: %w", err)
    }
    
    r.schemas[domain] = schema
    return nil
}

// LoadFromFile loads a schema from a JSON file
func (r *Registry) LoadFromFile(domain, path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("failed to read schema file: %w", err)
    }
    
    var schema Schema
    if err := json.Unmarshal(data, &schema); err != nil {
        return fmt.Errorf("failed to parse schema: %w", err)
    }
    
    schema.Raw = json.RawMessage(data)
    return r.Register(domain, &schema)
}

// ValidateValue validates a single value against its schema
func (r *Registry) ValidateValue(domain, path string, value interface{}) error {
    r.mu.RLock()
    schema, exists := r.schemas[domain]
    r.mu.RUnlock()
    
    if !exists {
        return fmt.Errorf("schema not found for domain: %s", domain)
    }
    
    // Navigate to the property in the schema
    prop, err := schema.GetProperty(path)
    if err != nil {
        return err
    }
    
    // Validate the value against the property schema
    return validateAgainstProperty(value, prop)
}
```

#### Dynamic Struct Generation
```go
package schema

import (
    "reflect"
    "strings"
)

// StructGenerator creates Go structs from JSON schemas
type StructGenerator struct {
    schema *Schema
}

// GenerateStruct creates a dynamic struct from schema
func (sg *StructGenerator) GenerateStruct() reflect.Type {
    fields := sg.generateFields(sg.schema.Properties, "")
    return reflect.StructOf(fields)
}

// generateFields recursively creates struct fields
func (sg *StructGenerator) generateFields(props map[string]Property, prefix string) []reflect.StructField {
    var fields []reflect.StructField
    
    for name, prop := range props {
        field := reflect.StructField{
            Name: toCamelCase(name),
            Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, name)),
        }
        
        switch prop.Type {
        case "string":
            field.Type = reflect.TypeOf("")
        case "integer":
            field.Type = reflect.TypeOf(0)
        case "number":
            field.Type = reflect.TypeOf(0.0)
        case "boolean":
            field.Type = reflect.TypeOf(false)
        case "object":
            if prop.Properties != nil {
                // Create nested struct
                nestedFields := sg.generateFields(prop.Properties, name+".")
                field.Type = reflect.StructOf(nestedFields)
            } else {
                // Generic object
                field.Type = reflect.TypeOf(map[string]interface{}{})
            }
        case "array":
            if prop.Items != nil {
                itemType := sg.getTypeForProperty(*prop.Items)
                field.Type = reflect.SliceOf(itemType)
            } else {
                field.Type = reflect.TypeOf([]interface{}{})
            }
        }
        
        fields = append(fields, field)
    }
    
    return fields
}

// CreateInstance creates an instance of the generated struct
func (sg *StructGenerator) CreateInstance() interface{} {
    structType := sg.GenerateStruct()
    return reflect.New(structType).Interface()
}
```

### Command Examples

#### Basic Configuration Operations
```bash
# List all configuration domains
heimdall config list
# Output:
# Available configuration domains:
#   - cli: CLI tool configuration
#   - shell: Quickshell configuration
#   - theme: Theme settings

# Get a configuration value
heimdall config cli get general.editor
# Output: nvim

heimdall config shell get appearance.colorScheme
# Output: catppuccin-mocha

# Set a configuration value
heimdall config shell set appearance.colorScheme "gruvbox-dark"
# Output: ✓ Set appearance.colorScheme to "gruvbox-dark"

# Validate configuration
heimdall config shell validate
# Output: ✓ Configuration is valid

# View schema for a domain
heimdall config shell schema
# Output: [JSON schema displayed with syntax highlighting]
```

#### Advanced Operations
```bash
# Initialize configuration from schema
heimdall config shell init --schema-url https://quickshell.org/schemas/heimdall/v1.0.0
# Output: ✓ Initialized shell configuration from schema

# Migrate configuration to new version
heimdall config shell migrate --to 1.1.0
# Output:
# Migrating shell configuration from 1.0.0 to 1.1.0...
#   ✓ Backup created: ~/.config/heimdall/backups/shell-1.0.0-20250812.json
#   ✓ Added new property: appearance.animations (default: true)
#   ✓ Migration complete

# Extract schema from external tool
heimdall config register quickshell-modules --schema-file /usr/share/quickshell/schemas/modules.json
# Output: ✓ Registered configuration domain: quickshell-modules

# Batch operations
heimdall config shell batch << EOF
set appearance.fontSize 14
set bar.position bottom
set bar.height 35
EOF
# Output:
# ✓ Set appearance.fontSize to 14
# ✓ Set bar.position to "bottom"  
# ✓ Set bar.height to 35
```

### Migration Strategy

#### Migration Workflow
```go
package migration

// MigrationEngine handles configuration migrations
type MigrationEngine struct {
    migrations map[string][]Migration
    backup     *BackupManager
}

// Migration represents a configuration migration
type Migration struct {
    FromVersion string
    ToVersion   string
    Transform   TransformFunc
    Rollback    RollbackFunc
}

// TransformFunc modifies configuration during migration
type TransformFunc func(config map[string]interface{}) error

// Migrate performs a configuration migration
func (me *MigrationEngine) Migrate(domain string, config map[string]interface{}, targetVersion string) error {
    currentVersion := config["version"].(string)
    
    // Find migration path
    path := me.findMigrationPath(domain, currentVersion, targetVersion)
    if len(path) == 0 {
        return fmt.Errorf("no migration path from %s to %s", currentVersion, targetVersion)
    }
    
    // Create backup
    backupPath, err := me.backup.Create(domain, config)
    if err != nil {
        return fmt.Errorf("failed to create backup: %w", err)
    }
    
    // Apply migrations
    for _, migration := range path {
        if err := migration.Transform(config); err != nil {
            // Rollback on failure
            me.backup.Restore(domain, backupPath)
            return fmt.Errorf("migration failed: %w", err)
        }
    }
    
    // Update version
    config["version"] = targetVersion
    
    return nil
}
```

#### Example Migration
```go
// Migration from 1.0.0 to 1.1.0
var migration_1_0_0_to_1_1_0 = Migration{
    FromVersion: "1.0.0",
    ToVersion:   "1.1.0",
    Transform: func(config map[string]interface{}) error {
        // Add new animations property if not exists
        appearance := config["appearance"].(map[string]interface{})
        if _, exists := appearance["animations"]; !exists {
            appearance["animations"] = true
        }
        
        // Rename deprecated property
        if theme, exists := appearance["theme"]; exists {
            appearance["colorScheme"] = theme
            delete(appearance, "theme")
        }
        
        return nil
    },
    Rollback: func(config map[string]interface{}) error {
        // Remove new property
        appearance := config["appearance"].(map[string]interface{})
        delete(appearance, "animations")
        
        // Restore old property name
        if scheme, exists := appearance["colorScheme"]; exists {
            appearance["theme"] = scheme
            delete(appearance, "colorScheme")
        }
        
        return nil
    },
}
```

### File Management Strategy

#### Atomic Operations
```go
package fileops

import (
    "os"
    "path/filepath"
)

// AtomicWriter provides atomic file write operations
type AtomicWriter struct {
    path string
}

// Write performs an atomic write operation
func (aw *AtomicWriter) Write(data []byte) error {
    // Create temp file in same directory
    dir := filepath.Dir(aw.path)
    temp, err := os.CreateTemp(dir, ".tmp-*")
    if err != nil {
        return err
    }
    tempPath := temp.Name()
    
    // Clean up temp file on any error
    defer func() {
        if err != nil {
            os.Remove(tempPath)
        }
    }()
    
    // Write data to temp file
    if _, err = temp.Write(data); err != nil {
        temp.Close()
        return err
    }
    
    // Sync to disk
    if err = temp.Sync(); err != nil {
        temp.Close()
        return err
    }
    temp.Close()
    
    // Set proper permissions
    if err = os.Chmod(tempPath, 0644); err != nil {
        return err
    }
    
    // Atomic rename
    if err = os.Rename(tempPath, aw.path); err != nil {
        return err
    }
    
    return nil
}
```

#### Configuration Locations
```
~/.config/heimdall/
├── config.json           # Main CLI configuration
├── shell.json           # Shell configuration (Quickshell)
├── themes.json          # Theme configuration
├── modules/             # Module-specific configs
│   ├── bar.json
│   ├── launcher.json
│   └── notifications.json
├── schemas/             # Cached/custom schemas
│   ├── shell-v1.0.0.json
│   └── custom/
├── backups/             # Automatic backups
│   ├── config-20250812-143022.json
│   └── shell-20250812-143022.json
└── migrations/          # Migration scripts
    └── shell/
        └── 1.0.0-to-1.1.0.json
```

This unified configuration management system provides:
1. **Extensibility**: Easy addition of new configuration domains
2. **Schema-driven**: All configurations validated against JSON schemas
3. **Dynamic**: Runtime schema extraction and struct generation
4. **Safe**: Atomic operations, automatic backups, migration support
5. **Consistent**: Unified command interface across all configuration types
6. **Future-proof**: Version management and migration capabilities