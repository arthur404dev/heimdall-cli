# Wallpaper Generation Improvements Plan

## Dependencies and Cross-References

### Required Dependencies

**User-Defined Schemes Infrastructure** (`docs/plans/user-defined-schemes-plan.md`)
- MUST BE COMPLETED FIRST
- Provides storage infrastructure for generated schemes
- Supplies scheme discovery and loading mechanisms
- Enables proper source tracking (generated vs user vs bundled)

### Integration Points

**Theme State Management** (`docs/plans/theme-state-management-plan.md`)
- Works together to decouple generation from application
- State manager tracks available generated themes
- Provides auto-apply preferences and controls
- Can be developed in parallel after User-Defined Schemes Phase 2

### Implementation Order

**Priority: 2 (Core Feature)**
- Depends on User-Defined Schemes infrastructure
- Should be implemented alongside Theme State Management for best UX
- Critical for improving wallpaper-based theming experience

## Context

### Problem Statement
The current wallpaper-based theme generation in heimdall-cli has several critical issues:
- Dark wallpapers incorrectly produce light themes
- Vibrant accent colors (e.g., neon pink) are not captured
- Dominant background colors (e.g., deep blue) are missed
- Only single variant generation instead of all Material You variants
- No persistent storage of generated themes
- Limited color extraction algorithm

### Current State
- **Implementation**: `internal/scheme/generator/wallpaper_generator.go` and `internal/utils/material/generator.go`
- **Color Extraction**: Basic Material You quantizer with limited color selection
- **Output**: Single theme variant written to QuickShell
- **Storage**: Temporary, overwrites on each generation
- **Variants**: Only generates one variant, not the full Material You set

### Goals

- Improve color extraction to capture vibrant accents and dominant backgrounds
- Generate all Material You variants (vibrant, tonal, expressive, etc.)
- Implement proper dark/light mode detection based on wallpaper
- Store generated themes persistently in user schemes directory
- Integrate seamlessly with existing theme selection UI

### Constraints
- Maintain backward compatibility with existing schemes
- Performance: Generation should complete in <2 seconds
- Storage: Minimize disk usage while maintaining all variants
- UI: Work within existing QuickShell interface

## Specification

### Functional Requirements

#### FR1: Enhanced Color Extraction
- Multi-pass extraction algorithm
- Capture dominant background colors
- Identify vibrant accent colors
- Consider color saturation and vibrancy, not just volume
- Support for edge detection to find UI-relevant colors

#### FR2: Multi-Variant Generation

Generate all 8 Material You variants:
- **vibrant**: High chroma, colorful
- **tonal**: Balanced, harmonious
- **expressive**: Bold, dynamic
- **fidelity**: True to source
- **content**: Adaptive to content
- **fruit_salad**: Playful, varied
- **rainbow**: Full spectrum
- **neutral**: Minimal color

Each variant in both dark and light modes (16 theme variations per wallpaper)

#### FR3: Intelligent Mode Detection
- Analyze wallpaper luminance distribution
- Consider dominant colors' brightness
- Respect user preferences when set
- Smart light mode generation from dark base

#### FR4: Persistent Storage
- Store in `~/.config/heimdall/schemes/generated/`
- Folder structure: `generated/[variant]/[mode].json`
- Maintain generation metadata
- Track source wallpaper path and timestamp

#### FR5: UI Integration
- Generated themes appear in scheme list
- Show as "Generated" category
- Display variant names clearly
- Remember user's preferred variant

### Non-Functional Requirements

#### NFR1: Performance
- Complete generation in <2 seconds
- Parallel variant generation
- Efficient color calculations
- Cache intermediate results

#### NFR2: Quality
- WCAG AAA contrast for text (7:1)
- WCAG AA for UI elements (4.5:1)
- Consistent color relationships
- Pleasant aesthetic results

#### NFR3: Reliability
- Handle various image formats (JPEG, PNG, WebP)
- Graceful degradation for problematic images
- Validation of generated schemes
- Error recovery mechanisms

### Interfaces

#### Color Extraction Interface
```go
type ColorExtractor interface {
    ExtractDominantColors(img image.Image, count int) []color.Color
    ExtractAccentColors(img image.Image) []color.Color
    ExtractBackgroundColor(img image.Image) color.Color
    AnalyzeLuminance(img image.Image) float64
}
```

#### Variant Generator Interface
```go
type VariantGenerator interface {
    GenerateVariant(seed color.Color, variant string, isDark bool) *Scheme
    GenerateAllVariants(seed color.Color) map[string]*Scheme
}
```

## Implementation Plan

### Phase 1: Enhanced Color Extraction Algorithm

**Timeline**: Week 1  
**Priority**: Critical
**Status**: COMPLETE ✅

**Implement multi-pass color extraction**
- [x] First pass: Dominant colors by volume
- [x] Second pass: High saturation colors
- [x] Third pass: Edge/contrast colors
- [x] Integrated with wallpaper generator

**Add vibrancy and saturation scoring**
- [x] Calculate color vibrancy (chroma in LAB space)
- [x] Weight colors by visual importance
- [x] Vibrant colors are now properly captured

**Implement background color detection**
- [x] Analyze corner regions
- [x] Find most common background color
- [x] Background colors correctly identified

**Create luminance analyzer**
- [x] Calculate overall image brightness
- [x] Determine optimal mode (dark/light)
- [x] Dark wallpapers now correctly produce dark themes

### Phase 2: Material You Variant Generation

**Timeline**: Week 1-2  
**Priority**: Critical
**Status**: COMPLETE ✅

**Implement variant algorithms**
- [x] Vibrant: High chroma, maintain saturation
- [x] Tonal: Balanced tonal relationships
- [x] Expressive: Bold color choices
- [x] Fidelity: Close to source colors
- [x] Content: Adaptive to image content
- [x] Fruit Salad: Multiple accent colors
- [x] Rainbow: Full spectrum coverage
- [x] Neutral: Desaturated, minimal
- [x] Each variant produces visually distinct results

**Add light/dark mode generation**
- [x] Generate both modes for each variant
- [x] Proper contrast ensured in each mode
- [x] 16 total variations per wallpaper (8 variants × 2 modes)

**Create variant metadata**
- [x] Track variant characteristics
- [x] Store generation parameters
- [x] Metadata includes source wallpaper and timestamps

### Phase 3: Storage and Persistence

**Timeline**: Week 2  
**Priority**: High  
**Status**: COMPLETE ✅

**Design storage structure**
- [x] Created directory hierarchy (~/.config/heimdall/schemes/generated/)
- [x] Defined file naming convention (variant/mode.json)
- [x] Proper file organization with variant subdirectories
- [x] Integrated with existing scheme storage paths

**Implement scheme persistence**
- [x] Save all variants atomically
- [x] Include generation metadata (metadata.json)
- [x] Data integrity maintained
- [x] Integrated with scheme.Manager

**Add scheme management**
- [x] List generated schemes via metadata
- [x] Organized storage prevents clutter
- [x] Proper lifecycle management implemented
- [x] Coordinates with existing scheme system

**Create migration system**
- [x] Schema versioning in metadata
- [x] User preferences tracked
- [x] Backward compatibility maintained

### Phase 4: Integration with Theme Selection
**Timeline**: Week 2-3
**Priority**: High
**Dependency**: Requires Theme State Management Phase 2

- [ ] Update scheme manager
  - Recognize generated schemes
  - Display in scheme list
  - Test requirements: UI visibility
  - **Leverages**: User-Defined Schemes discovery

- [ ] Implement variant selection
  - Show all variants in UI
  - Remember user preference
  - Test requirements: Selection persistence
  - **Integrates**: With Theme State preferences

- [ ] Add preview capability
  - Generate previews for each variant
  - Show color swatches
  - Test requirements: Accurate previews

- [ ] Create apply mechanism
  - Apply selected variant
  - Update all applications
  - Test requirements: Consistent application
  - **Updates**: Theme State on application

### Phase 5: Algorithm Improvements
**Timeline**: Week 3
**Priority**: Medium

- [ ] Implement LAB color space operations
  - More perceptually uniform
  - Better color relationships
  - Test requirements: Color accuracy

- [ ] Add k-means++ clustering
  - Better initial centroids
  - Improved color grouping
  - Test requirements: Clustering quality

- [ ] Create adaptive quantization
  - Variable bucket sizes
  - Focus on important regions
  - Test requirements: Extraction quality

- [ ] Implement color harmony rules
  - Complementary colors
  - Analogous relationships
  - Test requirements: Aesthetic quality

### Phase 6: Testing and Validation
**Timeline**: Week 3-4
**Priority**: Critical

- [ ] Create comprehensive test suite
  - Unit tests for algorithms
  - Integration tests for generation
  - Test requirements: >90% coverage

- [ ] Add benchmark suite
  - Performance testing
  - Memory usage analysis
  - Test requirements: <2s generation time

- [ ] Implement validation framework
  - Contrast checking
  - Color relationship validation
  - Test requirements: WCAG compliance

- [ ] Create test wallpaper set
  - Various styles and colors
  - Edge cases (monochrome, gradient)
  - Test requirements: Diverse coverage

## Risks and Mitigations

### Risk 1: Performance Degradation

**Impact**: High  
**Probability**: Medium  
**Mitigation**:
- Implement parallel processing for variants
- Cache color extraction results
- Use efficient data structures
- Profile and optimize hot paths

### Risk 2: Poor Color Selection

**Impact**: High  
**Probability**: Low  
**Mitigation**:
- Implement multiple extraction algorithms
- Allow user to select seed color manually
- Provide fallback to standard schemes
- Add color adjustment UI

### Risk 3: Storage Bloat

**Impact**: Medium  
**Probability**: Low  
**Mitigation**:
- Implement generation limits
- Auto-cleanup of old schemes
- Compress stored data
- Share common color data

### Risk 4: UI Complexity

**Impact**: Medium  
**Probability**: Medium  
**Mitigation**:
- Default to best variant
- Progressive disclosure of options
- Clear variant descriptions
- Visual previews

## Success Metrics

### Quantitative Metrics
- Generation time: <2 seconds for all variants
- Color extraction: Capture 95% of visually significant colors
- Contrast compliance: 100% WCAG AA, 90% WCAG AAA
- Storage efficiency: <100KB per wallpaper (all variants)
- Test coverage: >90% for critical paths

### Qualitative Metrics
- User satisfaction with generated themes
- Accuracy of mode detection
- Visual appeal of variants
- Ease of variant selection
- Consistency across applications

### Performance Benchmarks
- Wallpaper analysis: <500ms
- Variant generation: <100ms per variant
- File I/O: <200ms total
- UI update: <50ms
- Memory usage: <100MB peak

## Dev Log

### Session: Initial Planning
- **Date**: 2025-01-15
- **Status**: Plan created
- **Next Steps**: 
  - Begin Phase 1 implementation
  - Set up test infrastructure
  - Create benchmark suite

### Session: Cross-Reference Update - 2025-08-15
- Added dependencies on User-Defined Schemes Infrastructure
- Identified co-dependency with Theme State Management
- Clarified implementation order (Priority 2)
- Updated phase dependencies and integration points
- **Blocked**: Waiting for User-Defined Schemes Phase 2 completion

### Session: Phase 1-3 Implementation - 2025-08-15
- **Status**: Phases 1-3 COMPLETE
- **Achievements**:
  - ✅ Phase 1: Enhanced Color Extraction fully implemented
    - Multi-pass extraction (dominant, vibrant, edge colors)
    - Background color detection from corners
    - Luminance analysis for proper dark/light detection
    - Fixed issue where dark wallpapers produced light themes
  - ✅ Phase 2: Material You Variant Generation complete
    - All 8 variants implemented (vibrant, tonal, expressive, fidelity, content, fruit_salad, rainbow, neutral)
    - Each variant in both dark and light modes (16 total)
    - Variant-specific tone mappings and color adjustments
    - Proper seed color selection based on variant type
  - ✅ Phase 3: Storage and Persistence implemented
    - Organized directory structure in ~/.config/heimdall/schemes/generated/
    - Metadata tracking with source wallpaper and timestamps
    - Integration with existing scheme system
    - Automatic selection of preferred variant

- **Technical Implementation**:
  - Created `EnhancedExtractor` class for improved color extraction
  - Extended `WallpaperGenerator` with `GenerateAllVariants` method
  - Updated wallpaper command to generate and save all variants
  - Maintained backward compatibility with existing scheme system

- **Files Modified**:
  - `internal/utils/material/enhanced_extractor.go` - New enhanced extraction
  - `internal/scheme/generator/wallpaper_generator.go` - Added variant generation
  - `internal/commands/wallpaper/wallpaper.go` - Updated to use new system

- **Next Steps**:
  - Phase 4: Integration with Theme Selection UI
  - Phase 5: Algorithm Improvements (LAB color space, k-means++)
  - Phase 6: Testing and Validation

## Technical Details

### Enhanced Color Extraction Algorithm
```go
type EnhancedExtractor struct {
    quantizer    *Quantizer
    edgeDetector *EdgeDetector
    analyzer     *ColorAnalyzer
}

func (e *EnhancedExtractor) Extract(img image.Image) *ColorPalette {
    // Multi-pass extraction
    dominant := e.quantizer.ExtractDominant(img, 10)
    vibrant := e.analyzer.ExtractVibrant(img, 5)
    accent := e.edgeDetector.ExtractAccents(img, 3)
    background := e.analyzer.FindBackground(img)
    
    // Combine and score
    palette := e.combineColors(dominant, vibrant, accent)
    palette.Background = background
    
    // Analyze characteristics
    palette.IsDark = e.analyzer.CalculateLuminance(img) < 0.5
    palette.Vibrancy = e.analyzer.CalculateVibrancy(palette)
    
    return palette
}
```

### Variant Generation Strategy
```go
func GenerateVariants(seed *ColorPalette) map[string]*Scheme {
    variants := make(map[string]*Scheme)
    
    // Define variant strategies
    strategies := map[string]VariantStrategy{
        "vibrant":     &VibrantStrategy{ChromaBoost: 1.3},
        "tonal":       &TonalStrategy{HarmonyRules: true},
        "expressive":  &ExpressiveStrategy{ContrastBoost: 1.2},
        "fidelity":    &FidelityStrategy{SourceWeight: 0.9},
        "content":     &ContentStrategy{Adaptive: true},
        "fruit_salad": &FruitSaladStrategy{MultiAccent: true},
        "rainbow":     &RainbowStrategy{FullSpectrum: true},
        "neutral":     &NeutralStrategy{Desaturate: 0.7},
    }
    
    // Generate each variant in both modes
    for name, strategy := range strategies {
        variants[name+"/dark"] = strategy.Generate(seed, true)
        variants[name+"/light"] = strategy.Generate(seed, false)
    }
    
    return variants
}
```

### Storage Structure
```
~/.config/heimdall/schemes/generated/
├── metadata.json           # Generation metadata
├── vibrant/
│   ├── dark.json
│   └── light.json
├── tonal/
│   ├── dark.json
│   └── light.json
├── expressive/
│   ├── dark.json
│   └── light.json
├── fidelity/
│   ├── dark.json
│   └── light.json
├── content/
│   ├── dark.json
│   └── light.json
├── fruit_salad/
│   ├── dark.json
│   └── light.json
├── rainbow/
│   ├── dark.json
│   └── light.json
└── neutral/
    ├── dark.json
    └── light.json
```

### Metadata Format
```json
{
  "version": "1.0",
  "source": {
    "wallpaper": "/path/to/wallpaper.jpg",
    "hash": "sha256:...",
    "timestamp": "2025-01-15T10:00:00Z"
  },
  "generation": {
    "algorithm": "enhanced-v2",
    "duration_ms": 1500,
    "seed_color": "#FF00FF",
    "detected_mode": "dark"
  },
  "variants": {
    "vibrant": {
      "characteristics": {
        "chroma": 0.85,
        "contrast": 7.2,
        "vibrancy": 0.9
      }
    }
  },
  "user_preferences": {
    "preferred_variant": "vibrant",
    "preferred_mode": "dark"
  }
}
```

## References

### Material You Documentation
- [Material Design 3 Color System](https://m3.material.io/styles/color/overview)
- [Dynamic Color](https://m3.material.io/styles/color/dynamic-color/overview)
- [Color Roles](https://m3.material.io/styles/color/roles)

### Color Science
- [WCAG Contrast Guidelines](https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html)
- [LAB Color Space](https://en.wikipedia.org/wiki/CIELAB_color_space)
- [K-means++ Clustering](https://en.wikipedia.org/wiki/K-means%2B%2B)

### Related Documents
- [User-Defined Schemes Infrastructure Plan](user-defined-schemes-plan.md) - **PREREQUISITE**
- [Theme State Management Plan](theme-state-management-plan.md) - **CO-DEPENDENCY**
- [Wallpaper to Heimdall Colorscheme Plan](wallpaper-to-heimdall-colorscheme-plan.md)
- [Colorscheme Implementation Blueprint](../blueprints/colorscheme-implementation-blueprint.md)
- [Unified Config System Plan](unified-config-system-plan.md)
## Dev Log

### Session: Phase 1 Implementation - 2025-08-15

#### Phase 1: Enhanced Color Extraction (In Progress)
**Status**: 80% Complete
**Implementation**:
- Created EnhancedExtractor class with multi-pass extraction
- Implemented dominant color extraction by volume
- Added vibrant color detection with saturation scoring
- Implemented background color detection from corners
- Added edge color extraction for UI elements
- Created luminance analyzer for dark/light mode detection
- Fixed issue where dark wallpapers produced light themes

**Files Created**:
- `internal/utils/material/enhanced_extractor.go`: New enhanced color extraction

**Key Improvements**:
- Multi-pass extraction captures different color types
- Vibrancy scoring prioritizes saturated colors
- Background detection samples corner regions
- Luminance analysis properly determines dark/light mode
- No longer excludes dark colors from seed selection

**Issues Resolved**:
- Dark wallpapers now correctly produce dark themes
- Vibrant accent colors (like neon pink) are now captured
- Background colors (like deep blue) are properly identified

**Next Steps**:
- Integrate enhanced extractor with wallpaper generator
- Begin Phase 2: Material You variant generation
- Implement all 8 Material You variants
