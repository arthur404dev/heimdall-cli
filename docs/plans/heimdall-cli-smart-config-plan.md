# Heimdall-CLI Smart Configuration Management Plan

> **Related:** See [Unified Config System Plan](./unified-config-system-plan.md) for the overarching configuration architecture using the ConfigProvider pattern.

## Executive Summary

This plan outlines the implementation of comprehensive configuration management for heimdall-cli, following a unified approach that distinguishes between two distinct configuration domains while maintaining consistent operations across both:

### Configuration Domains

1. **Heimdall-CLI Configuration** (`~/.config/heimdall-cli/config.toml`)
   - Tool-specific settings and preferences
   - CLI behavior customization
   - Default values and user preferences
   - Managed via `heimdall config` commands

2. **Shell Configuration** (`~/.config/heimdall/shell.json`)
   - Quickshell UI and modules configuration
   - Visual appearance, bar settings, services
   - Read by Quickshell's Heimdall implementation
   - Managed via `heimdall shell config` commands

### Key Objectives
- Clear separation between CLI tool configuration and shell configuration
- Zero-configuration startup with automatic file creation
- Non-destructive property injection for new features
- Intelligent version migration with rollback capabilities
- Preservation of user customizations and comments
- Real-time validation and error recovery
- Atomic file operations to prevent corruption

### Success Metrics
- 100% backward compatibility with existing configurations
- < 50ms configuration validation time
- Zero data loss during migrations
- Automatic recovery from 95% of configuration errors
- Full preservation of user modifications and formatting

## Architecture

### Command Hierarchy (Unified Operations)

```
heimdall
├── config cli                # CLI tool configuration management
│   ├── get <key>            # Get CLI config value
│   ├── set <key> <value>    # Set CLI config value
│   ├── list                 # List all CLI settings
│   ├── reset                # Reset to defaults
│   ├── validate             # Validate configuration
│   ├── backup               # Create backup
│   ├── restore              # Restore from backup
│   └── edit                 # Open in editor
│
├── config shell             # Shell configuration management
│   ├── get <path>           # Get shell config value
│   ├── set <path> <value>   # Set shell config value
│   ├── list                 # List all shell settings
│   ├── reset                # Reset to defaults
│   ├── validate             # Validate shell.json
│   ├── backup               # Create backup
│   ├── restore              # Restore from backup
│   ├── edit                 # Open in editor
│   ├── init                 # Initialize shell.json
│   ├── migrate              # Migrate to new version
│   └── inject               # Inject missing properties
│
├── shell                     # Shell-specific operations
│   ├── reload               # Trigger Quickshell reload
│   └── status               # Show shell status
│
└── [other commands...]
```

**Note:** Both `config cli` and `config shell` share the same unified operations (get, set, list, reset, validate, backup, restore, edit) while maintaining domain-specific behaviors through the ConfigProvider pattern.

### Configuration Locations

```
~/.config/
├── heimdall-cli/           # CLI tool configuration
│   ├── config.toml         # Main CLI config
│   ├── profiles/           # Configuration profiles
│   └── backups/            # CLI config backups
│
└── heimdall/               # Shell configuration
    ├── shell.json          # Main shell config
    ├── backups/            # Shell config backups
    └── templates/          # Shell config templates
```

### Core Components (ConfigProvider Architecture)

> **Architecture Note:** All configuration managers implement the unified `ConfigProvider` interface defined in the [Unified Config System Plan](./unified-config-system-plan.md), ensuring consistent operations across different configuration domains.

#### 1. CLI Configuration Provider
```go
package config

// Implements ConfigProvider interface
type CLIConfigProvider struct {
    configPath string  // ~/.config/heimdall-cli/config.toml
    config     *CLIConfig
    schema     *Schema  // Extracted from struct tags
}

type CLIConfig struct {
    Version  string           `toml:"version"`
    General  GeneralSettings  `toml:"general"`
    Defaults DefaultValues    `toml:"defaults"`
    Behavior BehaviorSettings `toml:"behavior"`
}

// ConfigProvider interface implementation
func (cp *CLIConfigProvider) Get(key string) (interface{}, error)
func (cp *CLIConfigProvider) Set(key string, value interface{}) error
func (cp *CLIConfigProvider) List() map[string]interface{}
func (cp *CLIConfigProvider) Reset() error
func (cp *CLIConfigProvider) Validate() []ValidationError
func (cp *CLIConfigProvider) Backup() (string, error)
func (cp *CLIConfigProvider) Restore(backup string) error
func (cp *CLIConfigProvider) GetSchema() *Schema
```

#### 2. Shell Configuration Provider
```go
package shell

// Implements ConfigProvider interface
type ShellConfigProvider struct {
    configPath string  // ~/.config/heimdall/shell.json
    validator  *SchemaValidator
    migrator   *VersionMigrator
    injector   *PropertyInjector
    schema     *Schema  // Can be extracted from external sources
}

type ShellConfig struct {
    Version    string                 `json:"version"`
    Metadata   ConfigMetadata         `json:"metadata"`
    System     SystemConfig           `json:"system"`
    Appearance AppearanceConfig       `json:"appearance"`
    Bar        BarConfig              `json:"bar"`
    Modules    ModulesConfig          `json:"modules"`
    Services   ServicesConfig         `json:"services"`
    Commands   CommandsConfig         `json:"commands"`
    Wallpaper  WallpaperConfig        `json:"wallpaper"`
    HotReload  HotReloadConfig        `json:"hotReload"`
    
    // Preserve unknown fields for forward compatibility
    Extra map[string]interface{}   `json:"-"`
}

// ConfigProvider interface implementation
func (sp *ShellConfigProvider) Get(path string) (interface{}, error)
func (sp *ShellConfigProvider) Set(path string, value interface{}) error
func (sp *ShellConfigProvider) List() map[string]interface{}
func (sp *ShellConfigProvider) Reset() error
func (sp *ShellConfigProvider) Validate() []ValidationError
func (sp *ShellConfigProvider) Backup() (string, error)
func (sp *ShellConfigProvider) Restore(backup string) error
func (sp *ShellConfigProvider) GetSchema() *Schema

// Shell-specific extensions
func (sp *ShellConfigProvider) Initialize() error
func (sp *ShellConfigProvider) Migrate() error
func (sp *ShellConfigProvider) InjectDefaults() error
```

**Schema Extraction Note:** The Shell provider can extract schemas from external sources like TypeScript definitions or JSON schemas, enabling dynamic configuration validation without hardcoding structures.

#### 3. Unified Command Implementation
```go
package commands

// Unified config command using ConfigProvider pattern
type ConfigCommand struct {
    providers map[string]config.ConfigProvider
}

// Register providers for different domains
func NewConfigCommand() *ConfigCommand {
    return &ConfigCommand{
        providers: map[string]config.ConfigProvider{
            "cli":   config.NewCLIConfigProvider(),
            "shell": shell.NewShellConfigProvider(),
        },
    }
}

// Unified operations work across all domains
func (c *ConfigCommand) Get(domain, key string) error {
    provider := c.providers[domain]
    return provider.Get(key)
}

func (c *ConfigCommand) Set(domain, key, value string) error {
    provider := c.providers[domain]
    return provider.Set(key, value)
}

func (c *ConfigCommand) List(domain string) error {
    provider := c.providers[domain]
    return provider.List()
}

func (c *ConfigCommand) Validate(domain string) error {
    provider := c.providers[domain]
    return provider.Validate()
}

// Domain-specific extensions
func (c *ConfigCommand) ShellInit() error {
    if sp, ok := c.providers["shell"].(*shell.ShellConfigProvider); ok {
        return sp.Initialize()
    }
    return errors.New("operation not supported for this domain")
}
```

#### 4. Property Injection System
```go
package injection

type PropertyInjector struct {
    defaults  map[string]interface{}
    rules     []InjectionRule
    preserves []string // Paths to never modify
}

type InjectionRule struct {
    Path      string
    Condition func(current interface{}) bool
    Value     interface{}
    Strategy  InjectionStrategy
}

type InjectionStrategy int
const (
    MergeDeep InjectionStrategy = iota  // Recursive merge
    MergeShallow                         // Top-level merge only
    ReplaceIfMissing                     // Only if not exists
    ReplaceIfDefault                     // Only if matches default
    NeverReplace                         // User-locked
)
```

#### 5. Version Migration Engine
```go
package migration

type VersionMigrator struct {
    migrations map[string]Migration
    history    []MigrationRecord
}

type Migration interface {
    FromVersion() string
    ToVersion() string
    Migrate(config map[string]interface{}) error
    Rollback(config map[string]interface{}) error
    Validate(config map[string]interface{}) error
}

type MigrationRecord struct {
    From      string
    To        string
    Timestamp time.Time
    Backup    string
    Success   bool
}
```

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)

#### Tasks
- [ ] Set up Go project structure with unified ConfigProvider pattern
  - Acceptance Criteria:
    - Define `ConfigProvider` interface in `internal/config/provider.go`
    - Separate packages: `internal/config/cli` and `internal/config/shell` implementing the interface
    - Clear interface boundaries between packages
    - No circular dependencies
  - Tests: Unit tests for each module achieving 80% coverage
  - Files: internal/config/provider.go, internal/config/cli/, internal/config/shell/
  - Dependencies: None (foundation phase)
  - **Integration Note:** Implements the ConfigProvider pattern from the unified config system

- [ ] Implement dual configuration path discovery
  - Acceptance Criteria:
    - Correctly resolves `~/.config/heimdall-cli/config.toml` for CLI config
    - Correctly resolves `~/.config/heimdall/shell.json` for shell config
    - Creates directories if they don't exist
    - Respects XDG_CONFIG_HOME if set
  - Tests: Path resolution, XDG compliance, directory creation
  - Files: internal/utils/paths/
  - Dependencies: None

- [ ] Create TOML parser for CLI config
  - Acceptance Criteria:
    - Parses TOML without losing comments
    - Preserves formatting on write operations
    - Handles malformed TOML gracefully with clear errors
    - Supports nested configuration structures
  - Tests: Round-trip parsing, comment preservation, error cases
  - Files: internal/config/parser.go
  - Dependencies: Set up Go project structure

- [ ] Create JSON parser for shell config with comment preservation
  - Acceptance Criteria:
    - Uses JSON5 or custom parser to maintain comments
    - Preserves key ordering and formatting
    - Handles trailing commas and unquoted keys
    - Provides detailed parse error locations
  - Tests: Round-trip parsing, comment preservation, malformed JSON
  - Files: internal/shell/parser.go
  - Dependencies: Set up Go project structure

- [ ] Build atomic file operations wrapper
  - Acceptance Criteria:
    - All writes use temp file + atomic rename pattern
    - Automatic backup before any destructive operation
    - File lock support for concurrent access prevention
    - Rollback capability on write failure
  - Tests: Corruption recovery, concurrent access, rollback scenarios
  - Files: internal/utils/atomic.go
  - Dependencies: Dual configuration path discovery

#### Deliverables
- Dual configuration system foundation
- Separate CLI and shell config packages
- TOML and JSON handling with format preservation
- File operation safety layer

### Phase 2: CLI Configuration Management (Week 2-3)

#### Tasks
- [ ] Implement CLI config schema with automatic extraction
  - Acceptance Criteria:
    - Complete TOML structure with version, general, defaults, and behavior sections
    - All fields have appropriate Go struct tags for schema extraction
    - Support for custom types (e.g., color preferences, key bindings)
    - Schema versioning for future migrations
    - Automatic schema generation from struct tags
  - Tests: Marshaling/unmarshaling, field validation, version handling, schema extraction
  - Files: internal/config/cli/schema.go, internal/config/cli/provider.go
  - Dependencies: Phase 1 - TOML parser, ConfigProvider interface
  - **Integration Note:** CLI provider fully implements ConfigProvider interface with schema extraction

- [ ] Create `heimdall config` command structure
  - Acceptance Criteria:
    - Subcommands: get, set, list, reset, edit all functional
    - Consistent command output format
    - Proper error messages with suggestions
    - Help text for each subcommand
  - Tests: Command execution, argument parsing, error handling
  - Files: internal/commands/config/
  - Dependencies: CLI config schema

- [ ] Build CLI config defaults system
  - Acceptance Criteria:
    - Embedded defaults for all configuration fields
    - Environment variable override support
    - Default values documented in help text
    - Reset command restores to defaults
  - Tests: Default generation, override precedence, reset functionality
  - Files: internal/config/defaults.go
  - Dependencies: CLI config schema

- [ ] Implement CLI config validation
  - Acceptance Criteria:
    - Type checking for all fields
    - Range validation for numeric values
    - Path validation for file/directory settings
    - Clear error messages with field paths
  - Tests: Edge cases, invalid inputs, validation messages
  - Files: internal/config/validator.go
  - Dependencies: CLI config schema, defaults system

#### Deliverables
- Complete `heimdall config` command suite
- CLI configuration management system
- Default values and validation
- User-friendly error messages

### Phase 3: Shell Configuration Foundation (Week 3-4)

#### Tasks
- [ ] Define shell config schema with external source extraction
  - Acceptance Criteria:
    - Complete Go structs for all shell.json sections
    - Support for nested objects and arrays
    - Preserve unknown fields for forward compatibility
    - JSON tags for proper serialization
    - Ability to extract schema from TypeScript definitions or JSON Schema files
    - Dynamic schema loading from external sources
  - Tests: Schema completeness, marshaling accuracy, unknown field handling, external schema parsing
  - Files: internal/config/shell/schema.go, internal/config/shell/provider.go, internal/config/shell/extractor.go
  - Dependencies: Phase 1 - JSON parser, ConfigProvider interface
  - **Integration Note:** Shell provider implements ConfigProvider with dynamic schema extraction capabilities

- [ ] Implement shell config validator
  - Acceptance Criteria:
    - Validates against complete schema
    - Provides detailed error messages with JSON paths
    - Checks required fields and value constraints
    - Validates cross-field dependencies
  - Tests: Edge cases, partial configs, error message clarity
  - Files: internal/shell/validator.go
  - Dependencies: Shell config schema structures

- [ ] Create shell config initialization
  - Acceptance Criteria:
    - Creates valid shell.json with all required fields
    - Applies sensible defaults for Quickshell
    - Detects system capabilities (Wayland/X11)
    - Interactive mode for initial setup
  - Tests: Required fields presence, default values, system detection
  - Files: internal/shell/initializer.go
  - Dependencies: Shell config schema, validator

- [ ] Build configuration differ
  - Acceptance Criteria:
    - Deep comparison of nested structures
    - Array comparison with order awareness
    - Generates human-readable diff output
    - Identifies added, removed, and modified fields
  - Tests: Deep comparison, array handling, diff output format
  - Files: internal/shell/differ.go
  - Dependencies: Shell config schema

#### Deliverables
- Complete shell config Go structs
- Schema validation with detailed errors
- Shell config initialization system
- Configuration comparison utilities

### Phase 4: Property Injection System (Week 4-5)

#### Tasks
- [ ] Implement deep merge algorithm
  - Acceptance: Correctly merges nested structures
  - Tests: Complex nesting, array handling

- [ ] Create injection rule engine
  - Acceptance: Applies rules based on conditions
  - Tests: Rule precedence, conflicts

- [ ] Build user preference preservation
  - Acceptance: Never overwrites user changes
  - Tests: Modification detection

- [ ] Develop property discovery system
  - Acceptance: Identifies missing properties
  - Tests: Schema comparison, new fields

#### Deliverables
- Smart property injection without data loss
- User modification detection
- Conditional injection rules
- Missing property identification

### Phase 5: Version Migration (Week 5-6)

#### Tasks
- [ ] Design migration interface
  - Acceptance: Clean migration API
  - Tests: Mock migrations

- [ ] Implement version comparison
  - Acceptance: Semantic versioning support
  - Tests: Version ordering, ranges

- [ ] Create migration chain resolver
  - Acceptance: Finds optimal migration path
  - Tests: Multi-step migrations

- [ ] Build rollback mechanism
  - Acceptance: Reverts failed migrations
  - Tests: Rollback scenarios

#### Deliverables
- Version-aware migration system
- Migration path planning
- Automatic backup before migration
- Rollback capability

### Phase 6: Unified Config Commands (Week 6-7)

#### Tasks
- [ ] Implement unified `heimdall config` command tree
  - Acceptance: All subcommands functional for both cli and shell domains
  - Tests: Command parsing, help text, domain routing
  - **Integration Note:** Uses ConfigProvider pattern for unified operations

- [ ] Create unified config operations
  - Acceptance: get, set, list, reset, validate, backup, restore work across domains
  - Tests: Operation execution, error handling, provider selection
  - **Integration Note:** Single command implementation handles all domains through providers

- [ ] Build backup/restore functionality
  - Acceptance: Reliable backup and restoration
  - Tests: Backup integrity, restore accuracy

- [ ] Add interactive shell config editor
  - Acceptance: User-friendly editing experience
  - Tests: Input validation, change tracking

#### Deliverables
- Complete `heimdall shell config` command suite
- Shell configuration management operations
- Backup and restore capabilities
- Interactive configuration editing

### Phase 7: Advanced Features (Week 7-8)

#### Tasks
- [ ] Implement configuration profiles
  - Acceptance: Multiple config management
  - Tests: Profile switching, isolation

- [ ] Add configuration templates
  - Acceptance: Preset configurations
  - Tests: Template application

- [ ] Create configuration linting
  - Acceptance: Best practice suggestions
  - Tests: Performance warnings

- [ ] Build configuration export/import
  - Acceptance: Portable configurations
  - Tests: Format conversions

#### Deliverables
- Profile management system
- Configuration templates
- Linting and suggestions
- Import/export functionality

## Command Examples (Unified Operations)

### CLI Configuration Commands
```bash
# Get CLI configuration value
heimdall config cli get general.editor
# Output: nvim

# Set CLI configuration value
heimdall config cli set general.editor "code"
# Output: Set general.editor to "code"

# List all CLI settings
heimdall config cli list
# Output: 
# general.editor = "code"
# general.color_output = true
# defaults.scheme = "catppuccin-mocha"
# ...

# Validate CLI configuration
heimdall config cli validate
# Output: ✓ Configuration is valid

# Reset CLI config to defaults
heimdall config cli reset
# Output: Configuration reset to defaults

# Backup CLI configuration
heimdall config cli backup
# Output: Backup created: ~/.config/heimdall-cli/backups/config-20250812-143022.toml

# Restore CLI configuration
heimdall config cli restore config-20250812-143022.toml
# Output: Configuration restored from backup

# Edit CLI config in editor
heimdall config cli edit
# Opens ~/.config/heimdall-cli/config.toml in editor
```

### Shell Configuration Commands
```bash
# Initialize shell configuration
heimdall config shell init
# Output: Created shell configuration at ~/.config/heimdall/shell.json

# Get shell configuration value
heimdall config shell get appearance.colorScheme
# Output: "catppuccin-mocha"

# Set shell configuration value
heimdall config shell set appearance.colorScheme "gruvbox-dark"
# Output: Set appearance.colorScheme to "gruvbox-dark"

# List all shell settings
heimdall config shell list
# Output: Complete shell configuration in structured format

# Validate shell configuration
heimdall config shell validate
# Output: ✓ Configuration is valid

# Reset shell config to defaults
heimdall config shell reset
# Output: Configuration reset to defaults

# Backup shell configuration
heimdall config shell backup
# Output: Backup created: ~/.config/heimdall/backups/shell-20250812-143022.json

# Restore shell configuration
heimdall config shell restore shell-20250812-143022.json
# Output: Configuration restored from backup

# Edit shell config in editor
heimdall config shell edit
# Opens ~/.config/heimdall/shell.json in editor

# Shell-specific operations
heimdall config shell inject
# Output: Injected 3 new properties:
#   - modules.newFeature (default: enabled)
#   - services.monitoring (default: false)
#   - appearance.animations (default: true)

heimdall config shell migrate
# Output: Migrating from v1.2.0 to v1.3.0...
#   ✓ Backup created: ~/.config/heimdall/backups/shell-v1.2.0-20250812.json
#   ✓ Migration successful

# Reload shell with new config
heimdall shell reload
# Output: Shell configuration reloaded
```

**Note:** The unified command structure ensures consistent operations (get, set, list, validate, backup, restore, edit) across both CLI and shell configurations, while maintaining domain-specific extensions where needed.

## Property Injection Strategy

### Injection Algorithm
```go
func (i *PropertyInjector) InjectDefaults(config map[string]interface{}) error {
    // 1. Load current schema defaults
    defaults := i.loadDefaults()
    
    // 2. Identify missing properties
    missing := i.findMissingProperties(config, defaults)
    
    // 3. Check user locks
    locked := i.getUserLocks(config)
    
    // 4. Apply injection rules
    for path, value := range missing {
        if i.isUserLocked(path, locked) {
            continue
        }
        
        rule := i.findRule(path)
        if rule.ShouldInject(config, path) {
            i.injectProperty(config, path, value, rule.Strategy)
        }
    }
    
    // 5. Update metadata
    i.updateMetadata(config)
    
    return nil
}
```

### Preservation Rules
1. **Never modify user-customized values**
   - Track original defaults
   - Compare against current values
   - Skip if different from default

2. **Respect user locks**
   - Honor `metadata.userLocked` array
   - Allow explicit preservation markers
   - Support path wildcards

3. **Maintain structural integrity**
   - Preserve object key order
   - Maintain array element order
   - Keep formatting preferences

4. **Handle special cases**
   - Preserve comments (via special parser)
   - Maintain custom properties
   - Keep unknown fields for forward compatibility

## Version Migration Strategy

### Migration Path Resolution
```go
func (m *VersionMigrator) FindMigrationPath(from, to string) []Migration {
    // Build migration graph
    graph := m.buildMigrationGraph()
    
    // Find shortest path
    path := graph.ShortestPath(from, to)
    
    // Validate path
    for _, migration := range path {
        if err := migration.Validate(); err != nil {
            return m.findAlternatePath(from, to)
        }
    }
    
    return path
}
```

### Migration Safety
1. **Pre-migration validation**
   - Schema compatibility check
   - Data integrity verification
   - Dependency resolution

2. **Atomic operations**
   - Create backup before migration
   - Use temporary files
   - Atomic rename on success

3. **Rollback capability**
   - Keep migration history
   - Store rollback information
   - Automatic rollback on failure

4. **Post-migration verification**
   - Validate migrated config
   - Test critical paths
   - Verify data integrity

## Testing Procedures

### Unit Tests
```go
// cli_config_test.go
func TestCLIConfigOperations(t *testing.T) {
    // Test get/set operations
    // Test list functionality
    // Test reset to defaults
    // Test TOML parsing
}

// shell_config_test.go
func TestShellConfigCreation(t *testing.T) {
    // Test automatic creation
    // Test default values
    // Test path resolution
}

func TestPropertyInjection(t *testing.T) {
    // Test missing property detection
    // Test injection strategies
    // Test user lock respect
}

func TestVersionMigration(t *testing.T) {
    // Test version comparison
    // Test migration path finding
    // Test rollback mechanism
}
```

### Integration Tests
```go
// integration_test.go
func TestFullConfigLifecycle(t *testing.T) {
    // Create configs
    // Modify configs
    // Inject properties
    // Migrate versions
    // Validate results
}

func TestCommandExecution(t *testing.T) {
    // Test heimdall config commands
    // Test heimdall shell config commands
    // Test error handling
    // Test help output
}
```

### Scenario Tests
1. **Fresh Installation**
   - No existing configurations
   - Create both CLI and shell configs
   - Verify completeness

2. **Upgrade Scenario**
   - Old version configs
   - Apply migrations
   - Preserve customizations

3. **Corruption Recovery**
   - Malformed configurations
   - Missing required fields
   - Automatic repair

4. **Feature Addition**
   - New properties in schema
   - Inject without disruption
   - Maintain user changes

## Quickshell Integration

### Required Updates to Quickshell

#### 1. Update Config.qml (Line 27)
**File:** `~/.config/quickshell/heimdall/Config.qml`  
**Current:**
```qml
path: `${Paths.stringify(Paths.config)}/shell.json`
```
**Change To:**
```qml
path: `${Quickshell.env("HOME")}/.config/heimdall/shell.json`
```

#### 2. Update ConfigEnhanced.qml (Line 22)
**File:** `~/.config/quickshell/heimdall/ConfigEnhanced.qml`  
**Current:**
```qml
readonly property string configPath: `${Paths.stringify(Paths.config)}/shell.json`
```
**Change To:**
```qml
readonly property string configPath: `${Quickshell.env("HOME")}/.config/heimdall/shell.json`
```

### Migration Path with Fallback
```qml
// Graceful fallback implementation
Singleton {
    readonly property string newConfigPath: `${Quickshell.env("HOME")}/.config/heimdall/shell.json`
    readonly property string legacyConfigPath: `${Paths.stringify(Paths.config)}/shell.json`
    
    property string actualPath: {
        let newFile = Qt.createQmlObject(`
            import Quickshell.Io
            FileInfo { path: "${newConfigPath}" }
        `, this);
        
        if (newFile.exists) {
            console.log("Using heimdall-cli managed config at:", newConfigPath);
            return newConfigPath;
        }
        
        console.log("Falling back to legacy config at:", legacyConfigPath);
        return legacyConfigPath;
    }
}
```

## Error Handling

### Error Categories
```go
type ConfigError struct {
    Type     ErrorType
    Path     string
    Message  string
    Severity Severity
    Fix      *SuggestedFix
}

type ErrorType int
const (
    ValidationError ErrorType = iota
    MigrationError
    InjectionError
    IOError
    ParseError
)
```

### Recovery Strategies
1. **Automatic fixes**
   - Missing required fields → inject defaults
   - Invalid types → type coercion
   - Malformed JSON → format recovery

2. **User intervention**
   - Conflicting values → prompt user
   - Ambiguous migrations → user choice
   - Data loss risk → user confirmation

3. **Fallback mechanisms**
   - Use previous backup
   - Load minimal config
   - Start with defaults

## Performance Considerations

### Optimization Targets
- Config load time: < 10ms
- Validation time: < 50ms
- Migration time: < 100ms per version
- Property injection: < 20ms

### Caching Strategy
```go
type ConfigCache struct {
    config    interface{}
    checksum  string
    loadTime  time.Time
    ttl       time.Duration
}

func (c *ConfigCache) IsValid() bool {
    // Check file modification time
    // Verify checksum
    // Check TTL
}
```

## Security Considerations

### File Permissions
- CLI config: 0644 (user read/write, others read)
- Shell config: 0644 (user read/write, others read)
- Backup files: 0600 (user only)
- Directories: 0755 (user full, others read/execute)

### Input Validation
- Sanitize all user inputs
- Validate configuration structures
- Check path traversal attempts
- Limit file sizes

## Documentation Requirements

### User Documentation
- Installation guide
- Command reference for both config types
- Migration guide from legacy locations
- Troubleshooting guide
- Best practices

### Developer Documentation
- API reference for both config managers
- Architecture overview
- Contributing guide
- Testing guide
- Release process

## Risk Mitigation

### Technical Risks
| Risk | Impact | Mitigation |
|------|--------|------------|
| Data loss during migration | High | Automatic backups, rollback capability |
| Config confusion (CLI vs Shell) | Medium | Clear command hierarchy, documentation |
| Breaking changes in Quickshell | Medium | Version detection, compatibility layer |
| Performance degradation | Low | Caching, lazy loading |

### Operational Risks
| Risk | Impact | Mitigation |
|------|--------|------------|
| User confusion about two configs | Medium | Clear documentation, intuitive commands |
| Support burden | Medium | Comprehensive docs, self-healing |
| Feature creep | Low | Strict scope management |

## Success Metrics

### Technical Metrics
- Zero data loss incidents
- < 1% migration failure rate
- < 100ms operation latency
- 99.9% backward compatibility

### User Metrics
- 90% successful auto-configurations
- < 5% support tickets
- 80% feature adoption rate
- 95% user satisfaction
- Clear understanding of CLI vs Shell config

## Timeline

### Development Schedule
- Week 1-2: Core Infrastructure
- Week 2-3: CLI Configuration Management
- Week 3-4: Shell Configuration Foundation
- Week 4-5: Property Injection
- Week 5-6: Version Migration
- Week 6-7: Shell Config Commands
- Week 7-8: Advanced Features
- Week 8-9: Testing and Documentation
- Week 9-10: Beta Testing
- Week 10-11: Production Release

### Milestones
- [ ] M1: Core infrastructure with dual config support
- [ ] M2: CLI configuration management complete
- [ ] M3: Shell configuration foundation ready
- [ ] M4: Property injection working
- [ ] M5: Version migration implemented
- [ ] M6: Full command suite integrated
- [ ] M7: Beta release ready
- [ ] M8: Production release

## Dev Log

### Session: 2025-08-12 (Initial)
- Created comprehensive implementation plan
- Defined architecture with Go-specific components
- Established property injection strategy
- Designed version migration system
- Set up testing procedures
- Created rollout strategy

### Session: 2025-08-12 (Update 1)
- Updated configuration location from `~/.config/quickshell/heimdall/shell.json` to `~/.config/heimdall/shell.json`
- Clarified ownership: heimdall-cli owns and manages the configuration
- Updated all path references throughout the document
- Added Quickshell Config.qml update requirements
- Added backup directory locations

### Session: 2025-08-12 (Update 2 - Major Restructure)
- **CRITICAL CHANGE**: Separated CLI and Shell configuration management
- Introduced dual configuration system:
  - CLI config at `~/.config/heimdall-cli/config.toml`
  - Shell config at `~/.config/heimdall/shell.json`
- Restructured command hierarchy:
  - `heimdall config` for CLI tool configuration
  - `heimdall shell config` for shell configuration
- Updated all command examples to reflect new hierarchy
- Reorganized implementation phases to build both systems
- Added clear executive summary distinguishing config types
- Maintained all technical details for property injection and migration
- Updated architecture diagrams and component definitions
- Added command implementation examples for both config types

### Session: 2025-08-12 (Update 3 - Unified Architecture)
- **ARCHITECTURAL DECISION**: Adopted unified ConfigProvider pattern
- Added reference to unified-config-system-plan.md at document top
- Updated Executive Summary to mention unified approach
- Revised command structure to show unified operations:
  - `heimdall config cli` for CLI tool config
  - `heimdall config shell` for shell config
  - Consistent operations (get, set, list, validate, backup, restore) across domains
- Updated architecture section to reference ConfigProvider pattern:
  - All configuration managers implement ConfigProvider interface
  - Unified command implementation using provider pattern
  - Domain-specific extensions preserved where needed
- Added note about schema extraction from external sources:
  - CLI provider extracts schema from struct tags
  - Shell provider can extract from TypeScript/JSON Schema files
- Updated implementation phases to integrate with unified system:
  - Phase 1 now includes ConfigProvider interface setup
  - Phase 2 includes automatic schema extraction for CLI
  - Phase 3 includes external schema extraction for shell
  - Phase 6 renamed to "Unified Config Commands"
- Maintained backward compatibility with existing detailed implementation plans

### Next Steps
1. Set up Go project with ConfigProvider interface pattern
2. Implement CLI configuration provider with TOML support and schema extraction
3. Implement shell configuration provider with JSON support and external schema extraction
4. Create unified command structure using provider pattern
5. Build property injection system for shell config
6. Develop version migration engine
7. Create comprehensive test suite for unified operations
8. Document ConfigProvider pattern implementation details