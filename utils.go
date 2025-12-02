package valet

import "strings"

// lookupPath traverses nested data using dot notation
func lookupPath(data DataObject, path string) LookupResult {
	if data == nil {
		return LookupResult{nil, false}
	}
	if path == "" {
		return LookupResult{data, true}
	}

	parts := strings.Split(path, ".")
	var current any = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, exists := v[part]
			if !exists {
				return LookupResult{nil, false}
			}
			current = val
		default:
			return LookupResult{nil, false}
		}
	}

	return LookupResult{current, true}
}

// buildFieldPath creates dot-notation path from path slice
func buildFieldPath(path []string) string {
	return strings.Join(path, ".")
}
