package valet

import (
	"regexp"
	"sync"
)

// RegexCache provides thread-safe caching for compiled regex patterns
type RegexCache struct {
	mu    sync.RWMutex
	cache map[string]*regexp.Regexp
}

var globalRegexCache = &RegexCache{
	cache: make(map[string]*regexp.Regexp),
}

// GetOrCompile returns a cached compiled regex or compiles and caches it
func (rc *RegexCache) GetOrCompile(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, nil
	}

	// Try read lock first for better concurrency
	rc.mu.RLock()
	if cached, ok := rc.cache[pattern]; ok {
		rc.mu.RUnlock()
		return cached, nil
	}
	rc.mu.RUnlock()

	// Compile and cache with write lock
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Double-check after acquiring write lock
	if cached, ok := rc.cache[pattern]; ok {
		return cached, nil
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	rc.cache[pattern] = compiled
	return compiled, nil
}

// GetRegex returns a cached regex, compiling if necessary
// Panics if pattern is invalid (use for known-good patterns)
func GetRegex(pattern string) *regexp.Regexp {
	re, err := globalRegexCache.GetOrCompile(pattern)
	if err != nil {
		panic(err)
	}
	return re
}
