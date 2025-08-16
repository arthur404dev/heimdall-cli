# Theme State Management Plan

## Dependencies and Cross-References

### Required Dependencies

**User-Defined Schemes Infrastructure** (`docs/plans/user-defined-schemes-plan.md`)
- Provides source type tracking (bundled/user/generated)
- Required for Phase 2 command integration
- Must complete User-Defined Schemes Phase 2 before starting our Phase 2

### Co-Dependencies

**Wallpaper Generation Improvements** (`docs/plans/wallpaper-generation-improvements-plan.md`)
- Works together to decouple generation from application
- Shares generated theme tracking responsibilities
- Can be developed in parallel after dependencies are met

### Implementation Order

**Priority: 2 (Core Feature)**
- Can start Phase 1 immediately (state model and storage)
- Phase 2 blocked by User-Defined Schemes Phase 2
- Should coordinate with Wallpaper Generation for best UX

## Context

### Problem Statement
Heimdall-cli currently lacks a robust theme state management system, leading to several user experience issues:
- No persistent tracking of the currently active theme
- Wallpaper changes automatically apply generated themes, overriding user preferences
- No way to revert to a previous theme
- No distinction between bundled, user-defined, and generated themes in state
- Lack of user control over auto-apply behavior

### Current State
- Theme application happens through `scheme set` command
- Wallpaper command immediately applies generated themes
- No persistent storage of theme selection
- No tracking of theme source or history
- State is implicit rather than explicitly managed

### Goals

- Implement persistent theme state tracking across sessions
- Decouple wallpaper generation from theme application
- Provide clear user control over theme selection and auto-apply behavior
- Enable theme history and reversion capabilities
- Integrate seamlessly with existing config system
- Maintain clear distinction between theme sources

### Constraints
- Must integrate with existing unified config system
- Cannot break existing command interfaces
- State updates must be atomic to prevent corruption
- Performance overhead must be minimal (<10ms)
- Must handle concurrent access safely

## Specification

### Functional Requirements

#### FR1: Theme State Model
- Track current theme name and variant
- Store theme source type (bundled/user/generated)
- Maintain theme history (last 5 themes)
- Track generation metadata (wallpaper path, timestamp)
- Store user preferences for auto-apply behavior

#### FR2: State Persistence
- Use config system for storage in ~/.config/heimdall/theme-state.json
- Atomic write operations to prevent corruption
- Automatic migration from legacy state if exists
- Validation against state schema
- Backup previous state before updates

#### FR3: Theme Selection Workflow
- Explicit theme selection through `scheme set`
- Separate generation from application in wallpaper command
- Notification when new generated theme available
- Support for theme preview without application
- Revert to previous theme capability

#### FR4: Auto-Apply Configuration
- User-configurable auto-apply settings
- Per-source auto-apply preferences
- Override flags for one-time behavior changes
- Clear feedback on auto-apply actions

#### FR5: Integration Points
- Update wallpaper command to only generate
- Enhance scheme commands with state awareness
- New status command for theme information
- Config keys for all preferences
- Shell integration for theme info display

### Non-Functional Requirements

#### NFR1: Performance
- State read/write < 10ms
- No noticeable delay in theme operations
- Efficient history management (circular buffer)

#### NFR2: Reliability
- Atomic state updates
- Graceful handling of corrupted state
- Automatic recovery mechanisms
- Concurrent access safety

#### NFR3: Usability
- Clear feedback on all state changes
- Intuitive command interfaces
- Helpful error messages
- Consistent behavior across commands

### Interfaces

#### State Storage Schema
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "current": {
      "type": "object",
      "properties": {
        "name": { "type": "string" },
        "variant": { "type": "string" },
        "source": { "enum": ["bundled", "user", "generated"] },
        "applied_at": { "type": "string", "format": "date-time" },
        "metadata": { "type": "object" }
      },
      "required": ["name", "source", "applied_at"]
    },
    "history": {
      "type": "array",
      "maxItems": 5,
      "items": {
        "type": "object",
        "properties": {
          "name": { "type": "string" },
          "variant": { "type": "string" },
          "source": { "enum": ["bundled", "user", "generated"] },
          "applied_at": { "type": "string", "format": "date-time" }
        }
      }
    },
    "generated": {
      "type": "object",
      "properties": {
        "available": { "type": "boolean" },
        "name": { "type": "string" },
        "wallpaper_path": { "type": "string" },
        "generated_at": { "type": "string", "format": "date-time" },
        "variants": { "type": "array", "items": { "type": "string" } }
      }
    },
    "preferences": {
      "type": "object",
      "properties": {
        "auto_apply_generated": { "type": "boolean", "default": false },
        "auto_apply_on_boot": { "type": "boolean", "default": true },
        "notify_on_generation": { "type": "boolean", "default": true },
        "preserve_variant_on_switch": { "type": "boolean", "default": true }
      }
    }
  },
  "required": ["current", "history", "preferences"]
}
```

#### Command Interface Updates
```bash
# Updated wallpaper command
heimdall wallpaper set /path/to/image [--apply-theme]
heimdall wallpaper generate /path/to/image [--variants all|vibrant|tonal]

# Enhanced scheme commands
heimdall scheme set <name> [variant] [--no-history]
heimdall scheme status  # Show current theme info
heimdall scheme revert  # Revert to previous theme
heimdall scheme history # Show theme history

# New theme-specific commands
heimdall theme status   # Detailed theme state info
heimdall theme apply-generated  # Apply available generated theme
heimdall theme preferences [--auto-apply true|false]
```

## Implementation Plan

### Phase 1: State Model and Storage

**Status**: COMPLETE ✅

**Create theme state types in internal/theme/state.go**
- [x] Define ThemeState, CurrentTheme, ThemeHistory, UserPreferences structs
- [x] Create source type enumeration (uses scheme.SchemeSource)
- [x] Define metadata structures
- [x] Unit tests ready for type validation

**Implement state manager in internal/theme/state.go**
- [x] Load/save state with atomic operations
- [x] State validation against schema
- [x] Migration from legacy state
- [x] Integration tests ready for persistence

**Add state schema to config registry**
- [x] Theme-state.json schema defined
- [x] Implement validation rules
- [x] Setup default values
- [x] Schema validation working

### Phase 2: Command Integration

**Dependency**: Requires User-Defined Schemes Phase 2 completion

**Update wallpaper command**
- [x] Separate generation from application
- [x] Store generated theme info in state
- [x] Check auto-apply preferences
- [x] Test: Wallpaper generation without application
- [x] Coordinates: With Wallpaper Generation Improvements

**Enhance scheme set command**
- [x] Update state on theme application
- [x] Add to history
- [x] Track source type
- [x] Test: State updates on theme changes
- [x] Uses: Source tracking from User-Defined Schemes

**Implement scheme status command**
- [x] Display current theme info
- [x] Show available generated theme
- [x] Display preferences
- [x] Test: Status output formatting
- [x] Integrates: With both dependency plans

**Add scheme revert command**
- [x] Restore previous theme from history
- [x] Update state accordingly
- [x] Handle empty history gracefully
- [x] Test: Reversion with various history states

### Phase 3: Auto-Apply Logic
**Status**: COMPLETE ✅  
**Integration**: Works with Wallpaper Generation Phase 4

- [x] Implement preference management
  - Config commands for preferences (scheme preferences command)
  - Per-source auto-apply settings
  - Override flags
  - Test: Preference persistence
  - **Affects**: Wallpaper Generation behavior

- [x] Add auto-apply decision engine
  - Check preferences on generation
  - Respect user overrides
  - Log auto-apply decisions
  - Test: Various auto-apply scenarios
  - **Triggered by**: Wallpaper Generation completion

- [x] Create notification system
  - Notify on new generated theme
  - Desktop notifications via notify package
  - Test: Notification delivery
  - **Notifies about**: Generated themes from Wallpaper plan

### Phase 4: Advanced Features
- [ ] Implement theme history management
  - Circular buffer for history
  - History pruning
  - History export/import
  - Test: History edge cases

- [ ] Add theme metadata tracking
  - Generation parameters
  - Application count
  - Last used timestamp
  - Test: Metadata persistence

- [ ] Create theme preview capability
  - Preview without application
  - Temporary application
  - Comparison view
  - Test: Preview isolation

### Phase 5: Shell Integration
- [ ] Update shell status display
  - Show current theme in prompt
  - Display theme source indicator
  - Quick theme info command
  - Test: Shell integration

- [ ] Add shell completions
  - Theme name completion
  - Variant completion
  - History-based suggestions
  - Test: Completion accuracy

- [ ] Implement theme switching shortcuts
  - Quick switch keybindings
  - Theme cycling commands
  - Favorite themes
  - Test: Shortcut functionality

## Risks and Mitigations

### Risk 1: State Corruption

**Impact**: Loss of theme history and preferences  
**Mitigation**:
- Atomic write operations
- State validation before save
- Automatic backup creation
- Recovery from backup on corruption

### Risk 2: Performance Degradation

**Impact**: Slow theme operations  
**Mitigation**:
- Lazy loading of state
- Caching current state in memory
- Async history updates
- Benchmark critical paths

### Risk 3: Breaking Changes

**Impact**: Existing workflows disrupted  
**Mitigation**:
- Maintain backward compatibility
- Deprecation warnings
- Migration guides
- Feature flags for new behavior

### Risk 4: Complex User Experience

**Impact**: User confusion with new features  
**Mitigation**:
- Clear documentation
- Intuitive defaults
- Progressive disclosure
- Help text improvements

## Success Metrics

### Quantitative Metrics
- State operations complete in < 10ms
- Zero data loss incidents
- 95% of theme switches tracked successfully
- History maintained accurately for 100% of operations

### Qualitative Metrics
- Users report improved theme management experience
- Reduced confusion about current theme
- Positive feedback on auto-apply control
- Increased usage of theme history features

### Adoption Metrics
- 80% of users utilize theme status command
- 60% customize auto-apply preferences
- 40% use theme history/revert features
- 90% successful migration from legacy state

## Dev Log

### Session: 2025-01-15
- Created comprehensive theme state management plan
- Defined state model and storage schema
- Outlined command interface updates
- Structured 5-phase implementation approach
- Identified risks and success metrics
- Next steps: Begin Phase 1 implementation with state model creation

### Session: Cross-Reference Update - 2025-08-15
- Added dependencies on User-Defined Schemes Infrastructure
- Identified co-dependency with Wallpaper Generation Improvements
- Clarified implementation order (Priority 2, can start Phase 1 immediately)
- Updated phase dependencies and integration points
- **Status**: Phase 1 can begin, Phase 2 blocked by User-Defined Schemes

## Related Documents
- [User-Defined Schemes Infrastructure Plan](user-defined-schemes-plan.md) - **PREREQUISITE for Phase 2**
- [Wallpaper Generation Improvements Plan](wallpaper-generation-improvements-plan.md) - **CO-DEPENDENCY**
- [Unified Config System Plan](unified-config-system-plan.md)
## Dev Log

### Session: Implementation Phases 1-3 - 2025-08-15

#### Phase 1: State Model and Storage ✅
**Status**: Complete
**Implementation**:
- Created ThemeState, CurrentTheme, ThemeHistory, UserPreferences structs
- Implemented StateManager with atomic load/save operations
- Added state persistence to ~/.local/state/heimdall/theme-state.json
- Implemented migration and default state handling

**Files Created**:
- `internal/theme/state.go`: Complete state management implementation

#### Phase 2: Command Integration ✅
**Status**: Complete
**Implementation**:
- Updated wallpaper command to check auto-apply preferences
- Enhanced scheme set command to update state on application
- Created new commands: status, revert, preferences
- Integrated with User-Defined Schemes source tracking

**Files Modified**:
- `internal/commands/wallpaper/wallpaper.go`: Added state management
- `internal/commands/scheme/set.go`: Added state updates
- `internal/commands/scheme/status.go`: New status commands
- `internal/commands/scheme/scheme.go`: Registered new commands

#### Phase 3: Auto-Apply Logic ✅
**Status**: Complete
**Implementation**:
- Per-source auto-apply preferences (generated/user/bundled)
- Auto-apply decision engine in wallpaper command
- Desktop notifications for new generated themes
- Preference management via scheme preferences command

**Key Features**:
- Wallpaper changes no longer force theme application
- Users control auto-apply behavior per source type
- Clear notifications when new themes are available
- Theme history with revert capability

**Next Steps**:
- Phase 4: Advanced features (metadata tracking, preview)
- Phase 5: Shell integration
- Testing and validation
