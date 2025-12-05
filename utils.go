package valet

import (
	"strconv"
	"strings"
	"sync"
)

// pathCache caches split path results to avoid repeated strings.Split calls
type pathCache struct {
	mu    sync.RWMutex
	cache map[string][]string
}

var globalPathCache = &pathCache{
	cache: make(map[string][]string),
}

// getSplitPath returns cached split path or splits and caches it
func (pc *pathCache) getSplitPath(path string) []string {
	if path == "" {
		return nil
	}

	// Try read lock first for better concurrency
	pc.mu.RLock()
	if cached, ok := pc.cache[path]; ok {
		pc.mu.RUnlock()
		return cached
	}
	pc.mu.RUnlock()

	// Split and cache with write lock
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// Double-check after acquiring write lock
	if cached, ok := pc.cache[path]; ok {
		return cached
	}

	parts := strings.Split(path, ".")
	pc.cache[path] = parts
	return parts
}

// lookupPath traverses nested data using dot notation
func lookupPath(data DataObject, path string) LookupResult {
	if data == nil {
		return LookupResult{nil, false}
	}
	if path == "" {
		return LookupResult{data, true}
	}

	parts := globalPathCache.getSplitPath(path)
	var current any = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, exists := v[part]
			if !exists {
				return LookupResult{nil, false}
			}
			current = val
		case []any:
			// Handle array index access
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return LookupResult{nil, false}
			}
			current = v[idx]
		default:
			return LookupResult{nil, false}
		}
	}

	return LookupResult{current, true}
}
