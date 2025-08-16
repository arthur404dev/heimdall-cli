# User-Defined Schemes Test Analysis

## Current Test Coverage Status

### Existing Test Files
1. **internal/scheme/manager_test.go**
   - Basic manager functionality tests
   - Tests GetCurrent, SetScheme, ListSchemes, LoadScheme, SaveScheme
   - **DOES NOT** test new user-defined scheme functionality

2. **internal/commands/scheme/** test files:
   - bundled_test.go
   - get_test.go
   - install_test.go
   - list_test.go
   - mocks_test.go
   - scheme_test.go
   - set_test.go

### Missing Tests for User-Defined Schemes

Based on the plan document, the following test requirements are **NOT COMPLETED**:

#### Phase 1: Configuration Infrastructure
- [ ] Config loading with UserPaths field
- [ ] Path expansion for user scheme directories
- [ ] Environment variable parsing (`HEIMDALL_SCHEME_PATHS`)
- [ ] Precedence testing (env vars override config)
- [ ] Path initialization in paths/xdg.go

#### Phase 2: Manager Extension
- [ ] Listing schemes from multiple sources (user + bundled)
- [ ] Deduplication when same scheme exists in multiple sources
- [ ] Load priority (user schemes override bundled)
- [ ] Source tracking (SourceUser, SourceBundled, SourceGenerated)
- [ ] Flavour/mode discovery across sources
- [ ] getUserSchemePaths() helper function
- [ ] GetSchemeSource() method functionality

#### Phase 3: Command Updates
- [ ] Source filtering with `--source` flag in list command
- [ ] Visual indicators ([user], [generated]) in output
- [ ] Getting user schemes with scheme get command
- [ ] Installing to user directory with `--user` flag
- [ ] SaveSchemeToUser() method
- [ ] InstallBundledSchemeToUser() method

#### Phase 4: Validation and Error Handling (Not Started)
- [ ] Invalid JSON structure handling
- [ ] Missing required color keys validation
- [ ] Helpful error messages for malformed schemes
- [ ] Format conversion from .txt to JSON
- [ ] Duplicate scheme conflict resolution

#### Phase 5: Testing and Documentation (Not Started)
- [ ] 80%+ code coverage requirement
- [ ] Integration tests for command workflows
- [ ] Tests with various scheme structures
- [ ] Documentation accuracy validation
- [ ] Template validity testing

## Priority Tasks in Order

### Immediate Priority (Must Complete First)

1. **Create User-Defined Schemes Test Suite**
   - File: `internal/scheme/manager_user_schemes_test.go`
   - Tests for all Phase 1-3 functionality
   - Focus on multi-source support and priority ordering

2. **Create Command Integration Tests**
   - File: `internal/commands/scheme/user_schemes_test.go`
   - Test --source flag filtering
   - Test --user flag for installation
   - Test source indicators in output

3. **Create Config Tests**
   - File: `internal/config/user_paths_test.go`
   - Test UserPaths configuration
   - Test environment variable override
   - Test path expansion

### Next Priority (Phase 4)

4. **Implement Validation Logic**
   - Add validation methods to manager.go
   - Create validation_test.go
   - Test malformed JSON handling
   - Test missing color keys

5. **Implement Migration Helper**
   - Add migration methods for old formats
   - Create migration_test.go
   - Test .txt to JSON conversion

### Final Priority (Phase 5)

6. **Integration Test Suite**
   - End-to-end workflow tests
   - Coverage analysis
   - Performance benchmarks

7. **Documentation Updates**
   - Update README with user scheme guide
   - Create example schemes
   - Document folder structure

## Test Implementation Order

### Step 1: Foundation Tests (TODAY)
```go
// internal/scheme/manager_user_schemes_test.go
- TestGetUserSchemePaths()
- TestGetUserSchemePaths_WithEnvVar()
- TestListSchemes_MultiSource()
- TestListSchemes_Deduplication()
- TestLoadScheme_UserPriority()
- TestGetSchemeSource()
- TestSaveSchemeToUser()
```

### Step 2: Command Tests (TODAY)
```go
// internal/commands/scheme/user_schemes_test.go
- TestListCommand_SourceFilter()
- TestListCommand_SourceIndicators()
- TestGetCommand_UserScheme()
- TestInstallCommand_UserFlag()
- TestInstallBundledSchemeToUser()
```

### Step 3: Config Tests (TODAY)
```go
// internal/config/user_paths_test.go
- TestSchemeConfig_UserPaths()
- TestSchemeConfig_EnvVarOverride()
- TestSchemeConfig_PathExpansion()
```

## Validation Checklist

After implementing tests, verify:
- [ ] All Phase 1-3 test requirements checked off in plan
- [ ] Code coverage > 80% for new functionality
- [ ] All tests pass
- [ ] Integration with existing tests maintained
- [ ] No regression in existing functionality

## Next Actions

1. **Immediate**: Write foundation tests for manager_user_schemes_test.go
2. **Then**: Write command integration tests
3. **Then**: Write config tests
4. **Update**: Check off test requirements in user-defined-schemes-plan.md
5. **Proceed**: Move to Phase 4 implementation after tests pass