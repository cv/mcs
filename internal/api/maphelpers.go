package api

// Type-safe helper functions for extracting values from map[string]interface{}
// These helpers prevent runtime panics from unsafe type assertions.

// getString safely extracts a string value from a map.
func getString(m map[string]any, key string) (string, bool) {
	if m == nil {
		return "", false
	}
	val, exists := m[key]
	if !exists {
		return "", false
	}
	str, ok := val.(string)

	return str, ok
}

// getInt safely extracts an int value from a map.
func getInt(m map[string]any, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	val, exists := m[key]
	if !exists {
		return 0, false
	}
	// Handle both int and float64 (JSON numbers are float64)
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// getFloat64 safely extracts a float64 value from a map.
func getFloat64(m map[string]any, key string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	val, exists := m[key]
	if !exists {
		return 0, false
	}
	f, ok := val.(float64)

	return f, ok
}

// getBool safely extracts a bool value from a map.
func getBool(m map[string]any, key string) (bool, bool) {
	if m == nil {
		return false, false
	}
	val, exists := m[key]
	if !exists {
		return false, false
	}
	b, ok := val.(bool)

	return b, ok
}

// getMap safely extracts a nested map from a map.
func getMap(m map[string]any, key string) (map[string]any, bool) {
	if m == nil {
		return nil, false
	}
	val, exists := m[key]
	if !exists {
		return nil, false
	}
	nested, ok := val.(map[string]any)

	return nested, ok
}

// getSlice safely extracts a slice from a map.
func getSlice(m map[string]any, key string) ([]any, bool) {
	if m == nil {
		return nil, false
	}
	val, exists := m[key]
	if !exists {
		return nil, false
	}
	slice, ok := val.([]any)

	return slice, ok
}

// getMapSlice safely extracts a slice of maps from a map.
func getMapSlice(m map[string]any, key string) ([]map[string]any, bool) {
	if m == nil {
		return nil, false
	}
	val, exists := m[key]
	if !exists {
		return nil, false
	}

	// First check if it's a []interface{}
	slice, ok := val.([]any)
	if !ok {
		return nil, false
	}

	// Convert each element to map[string]interface{}
	result := make([]map[string]any, 0, len(slice))
	for _, item := range slice {
		if mapItem, ok := item.(map[string]any); ok {
			result = append(result, mapItem)
		} else {
			// If any item is not a map, fail the whole operation
			return nil, false
		}
	}

	return result, true
}

// getMapFromSlice safely gets a map from a slice at the given index.
func getMapFromSlice(slice []any, index int) (map[string]any, bool) {
	if slice == nil || index < 0 || index >= len(slice) {
		return nil, false
	}
	m, ok := slice[index].(map[string]any)

	return m, ok
}
