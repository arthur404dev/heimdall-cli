package theme

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/arthur404dev/heimdall-cli/internal/utils/logger"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	Size       int64       `json:"size"`
	AccessTime time.Time   `json:"access_time"`
	CreateTime time.Time   `json:"create_time"`
	HitCount   int         `json:"hit_count"`
}

// TemplateCache provides caching for parsed templates
type TemplateCache struct {
	mu          sync.RWMutex
	entries     map[string]*CacheEntry
	maxSize     int64 // Maximum cache size in bytes (default 10MB)
	currentSize int64
	ttl         time.Duration
	diskCache   bool
	cacheDir    string
	stats       CacheStats
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits       int64
	Misses     int64
	Evictions  int64
	TotalSize  int64
	EntryCount int
}

// ColorConversionCache caches color conversion results
type ColorConversionCache struct {
	mu      sync.RWMutex
	cache   map[string]map[string]string // [input_color][format] = result
	maxSize int
}

// NewTemplateCache creates a new template cache
func NewTemplateCache(maxSizeMB int, diskCache bool, cacheDir string) *TemplateCache {
	if maxSizeMB <= 0 {
		maxSizeMB = 10 // Default 10MB
	}

	tc := &TemplateCache{
		entries:   make(map[string]*CacheEntry),
		maxSize:   int64(maxSizeMB * 1024 * 1024),
		ttl:       24 * time.Hour, // Default TTL
		diskCache: diskCache,
		cacheDir:  cacheDir,
	}

	if diskCache && cacheDir != "" {
		// Load disk cache if it exists
		tc.loadDiskCache()
	}

	// Start cleanup goroutine
	go tc.cleanupRoutine()

	return tc
}

// Get retrieves a cached template
func (tc *TemplateCache) Get(key string) (interface{}, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	entry, exists := tc.entries[key]
	if !exists {
		tc.stats.Misses++
		return nil, false
	}

	// Check if entry has expired
	if tc.ttl > 0 && time.Since(entry.CreateTime) > tc.ttl {
		tc.stats.Misses++
		return nil, false
	}

	// Update access time and hit count
	entry.AccessTime = time.Now()
	entry.HitCount++
	tc.stats.Hits++

	return entry.Value, true
}

// Set stores a template in the cache
func (tc *TemplateCache) Set(key string, value interface{}, size int64) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// Check if we need to evict entries
	if tc.currentSize+size > tc.maxSize {
		tc.evictLRU(size)
	}

	entry := &CacheEntry{
		Key:        key,
		Value:      value,
		Size:       size,
		AccessTime: time.Now(),
		CreateTime: time.Now(),
		HitCount:   0,
	}

	tc.entries[key] = entry
	tc.currentSize += size
	tc.stats.EntryCount = len(tc.entries)
	tc.stats.TotalSize = tc.currentSize

	// Save to disk if enabled
	if tc.diskCache {
		go tc.saveToDisk(key, entry)
	}

	return nil
}

// Invalidate removes an entry from the cache
func (tc *TemplateCache) Invalidate(key string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if entry, exists := tc.entries[key]; exists {
		tc.currentSize -= entry.Size
		delete(tc.entries, key)
		tc.stats.EntryCount = len(tc.entries)
		tc.stats.TotalSize = tc.currentSize

		// Remove from disk cache
		if tc.diskCache {
			tc.removeFromDisk(key)
		}
	}
}

// InvalidateByPattern removes entries matching a pattern
func (tc *TemplateCache) InvalidateByPattern(pattern string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	for key, entry := range tc.entries {
		if matched, _ := filepath.Match(pattern, key); matched {
			tc.currentSize -= entry.Size
			delete(tc.entries, key)
			if tc.diskCache {
				tc.removeFromDisk(key)
			}
		}
	}

	tc.stats.EntryCount = len(tc.entries)
	tc.stats.TotalSize = tc.currentSize
}

// Clear removes all entries from the cache
func (tc *TemplateCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.entries = make(map[string]*CacheEntry)
	tc.currentSize = 0
	tc.stats.EntryCount = 0
	tc.stats.TotalSize = 0

	// Clear disk cache
	if tc.diskCache && tc.cacheDir != "" {
		os.RemoveAll(tc.cacheDir)
		os.MkdirAll(tc.cacheDir, 0755)
	}
}

// GetStats returns cache statistics
func (tc *TemplateCache) GetStats() CacheStats {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.stats
}

// evictLRU evicts least recently used entries to make space
func (tc *TemplateCache) evictLRU(neededSize int64) {
	// Find LRU entries
	var oldestTime time.Time
	var oldestKey string

	for tc.currentSize+neededSize > tc.maxSize && len(tc.entries) > 0 {
		oldestTime = time.Now()
		oldestKey = ""

		for key, entry := range tc.entries {
			if entry.AccessTime.Before(oldestTime) {
				oldestTime = entry.AccessTime
				oldestKey = key
			}
		}

		if oldestKey != "" {
			if entry, exists := tc.entries[oldestKey]; exists {
				tc.currentSize -= entry.Size
				delete(tc.entries, oldestKey)
				tc.stats.Evictions++

				if tc.diskCache {
					tc.removeFromDisk(oldestKey)
				}
			}
		}
	}
}

// cleanupRoutine periodically cleans up expired entries
func (tc *TemplateCache) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		tc.cleanupExpired()
	}
}

// cleanupExpired removes expired entries
func (tc *TemplateCache) cleanupExpired() {
	if tc.ttl <= 0 {
		return
	}

	tc.mu.Lock()
	defer tc.mu.Unlock()

	now := time.Now()
	for key, entry := range tc.entries {
		if now.Sub(entry.CreateTime) > tc.ttl {
			tc.currentSize -= entry.Size
			delete(tc.entries, key)
			tc.stats.Evictions++

			if tc.diskCache {
				tc.removeFromDisk(key)
			}
		}
	}

	tc.stats.EntryCount = len(tc.entries)
	tc.stats.TotalSize = tc.currentSize
}

// Disk cache operations

func (tc *TemplateCache) getCacheFilePath(key string) string {
	hash := md5.Sum([]byte(key))
	filename := hex.EncodeToString(hash[:]) + ".cache"
	return filepath.Join(tc.cacheDir, filename)
}

func (tc *TemplateCache) saveToDisk(key string, entry *CacheEntry) {
	if tc.cacheDir == "" {
		return
	}

	path := tc.getCacheFilePath(key)
	data, err := json.Marshal(entry)
	if err != nil {
		logger.Warn("Failed to marshal cache entry", "key", key, "error", err)
		return
	}

	// Ensure directory exists
	os.MkdirAll(tc.cacheDir, 0755)

	if err := os.WriteFile(path, data, 0644); err != nil {
		logger.Warn("Failed to save cache to disk", "key", key, "error", err)
	}
}

func (tc *TemplateCache) loadFromDisk(key string) (*CacheEntry, error) {
	path := tc.getCacheFilePath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (tc *TemplateCache) removeFromDisk(key string) {
	if tc.cacheDir == "" {
		return
	}

	path := tc.getCacheFilePath(key)
	os.Remove(path)
}

func (tc *TemplateCache) loadDiskCache() {
	if tc.cacheDir == "" {
		return
	}

	// Create cache directory if it doesn't exist
	os.MkdirAll(tc.cacheDir, 0755)

	// Load all cache files
	files, err := filepath.Glob(filepath.Join(tc.cacheDir, "*.cache"))
	if err != nil {
		logger.Warn("Failed to load disk cache", "error", err)
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}

		// Check if entry is still valid
		if tc.ttl > 0 && time.Since(entry.CreateTime) > tc.ttl {
			os.Remove(file)
			continue
		}

		tc.entries[entry.Key] = &entry
		tc.currentSize += entry.Size
	}

	tc.stats.EntryCount = len(tc.entries)
	tc.stats.TotalSize = tc.currentSize
	logger.Info("Loaded disk cache", "entries", len(tc.entries))
}

// NewColorConversionCache creates a new color conversion cache
func NewColorConversionCache(maxSize int) *ColorConversionCache {
	if maxSize <= 0 {
		maxSize = 1000 // Default to 1000 entries
	}

	return &ColorConversionCache{
		cache:   make(map[string]map[string]string),
		maxSize: maxSize,
	}
}

// Get retrieves a cached color conversion
func (ccc *ColorConversionCache) Get(color, format string) (string, bool) {
	ccc.mu.RLock()
	defer ccc.mu.RUnlock()

	if colorMap, exists := ccc.cache[color]; exists {
		if result, ok := colorMap[format]; ok {
			return result, true
		}
	}

	return "", false
}

// Set stores a color conversion result
func (ccc *ColorConversionCache) Set(color, format, result string) {
	ccc.mu.Lock()
	defer ccc.mu.Unlock()

	// Check if we need to evict entries
	if len(ccc.cache) >= ccc.maxSize {
		// Simple eviction: remove first entry
		for k := range ccc.cache {
			delete(ccc.cache, k)
			break
		}
	}

	if _, exists := ccc.cache[color]; !exists {
		ccc.cache[color] = make(map[string]string)
	}

	ccc.cache[color][format] = result
}

// Clear removes all entries from the cache
func (ccc *ColorConversionCache) Clear() {
	ccc.mu.Lock()
	defer ccc.mu.Unlock()

	ccc.cache = make(map[string]map[string]string)
}

// BatchConvert performs batch color conversions with caching
func (ccc *ColorConversionCache) BatchConvert(colors []string, format string, converter func(string) string) map[string]string {
	results := make(map[string]string, len(colors))

	for _, color := range colors {
		// Check cache first
		if cached, ok := ccc.Get(color, format); ok {
			results[color] = cached
			continue
		}

		// Convert and cache
		result := converter(color)
		ccc.Set(color, format, result)
		results[color] = result
	}

	return results
}

// CacheKey generates a cache key for a template
func CacheKey(app, mode string, colors map[string]string) string {
	// Create a deterministic key from the inputs
	h := md5.New()
	h.Write([]byte(app))
	h.Write([]byte(mode))

	// Sort colors for consistent hashing
	for k, v := range colors {
		h.Write([]byte(k))
		h.Write([]byte(v))
	}

	return fmt.Sprintf("%s_%s_%s", app, mode, hex.EncodeToString(h.Sum(nil)))
}
