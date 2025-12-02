package valet

import (
	"regexp"
	"sync"
	"testing"
)

func TestRegexCache_GetOrCompile(t *testing.T) {
	cache := &RegexCache{
		cache: make(map[string]*regexp.Regexp),
	}

	t.Run("compile new pattern", func(t *testing.T) {
		re, err := cache.GetOrCompile(`^\d+$`)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if re == nil {
			t.Error("Expected non-nil regex")
		}
	})

	t.Run("return cached pattern", func(t *testing.T) {
		pattern := `^[a-z]+$`
		re1, _ := cache.GetOrCompile(pattern)
		re2, _ := cache.GetOrCompile(pattern)

		if re1 != re2 {
			t.Error("Expected same cached regex instance")
		}
	})

	t.Run("empty pattern returns nil", func(t *testing.T) {
		re, err := cache.GetOrCompile("")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if re != nil {
			t.Error("Expected nil for empty pattern")
		}
	})

	t.Run("invalid pattern returns error", func(t *testing.T) {
		_, err := cache.GetOrCompile(`[invalid`)
		if err == nil {
			t.Error("Expected error for invalid pattern")
		}
	})
}

func TestRegexCache_Concurrent(t *testing.T) {
	cache := &RegexCache{
		cache: make(map[string]*regexp.Regexp),
	}

	patterns := []string{
		`^\d+$`,
		`^[a-z]+$`,
		`^[A-Z]+$`,
		`^\w+$`,
		`^.+@.+\..+$`,
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pattern := patterns[idx%len(patterns)]
			_, err := cache.GetOrCompile(pattern)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()
}

func TestGlobalRegexCache(t *testing.T) {
	t.Run("uses global cache", func(t *testing.T) {
		pattern := `^test\d+$`
		re1, _ := globalRegexCache.GetOrCompile(pattern)
		re2, _ := globalRegexCache.GetOrCompile(pattern)

		if re1 != re2 {
			t.Error("Expected same cached regex from global cache")
		}
	})
}

func TestGetRegex(t *testing.T) {
	t.Run("valid pattern", func(t *testing.T) {
		re := GetRegex(`^\d+$`)
		if re == nil {
			t.Error("Expected non-nil regex")
		}
	})

	t.Run("invalid pattern panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid pattern")
			}
		}()
		GetRegex(`[invalid`)
	})
}
