package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// FormatKeys recursively formats a value with the given variables
func FormatKeys(val interface{}, variables map[string]interface{}) (interface{}, error) {
	switch v := val.(type) {
	case string:
		// Check for type conversion markers
		formatted, err := formatString(v, variables)
		if err != nil {
			return nil, err
		}
		return applyTypeConversion(formatted)
	case map[string]interface{}:
		return formatMap(v, variables)
	case []interface{}:
		return formatSlice(v, variables)
	default:
		return val, nil
	}
}

// formatString replaces variables in a string
// Supports both flat variables {var} and nested variables {a.b.c}
func formatString(s string, variables map[string]interface{}) (string, error) {
	result := s

	// Find all {variable} patterns
	for {
		start := strings.Index(result, "{")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		varPath := result[start+1 : end]

		// Support nested access with dots (e.g., tavern.env_vars.TOKEN)
		var value interface{}
		var ok bool

		if strings.Contains(varPath, ".") {
			// Nested variable access
			value, ok = getNestedValue(variables, varPath)
		} else {
			// Simple variable access
			value, ok = variables[varPath]
		}

		if !ok {
			return "", NewMissingFormatError(varPath)
		}

		// Convert value to string
		var strValue string
		switch v := value.(type) {
		case string:
			strValue = v
		case int, int64, float64, bool:
			strValue = fmt.Sprintf("%v", v)
		default:
			strValue = fmt.Sprintf("%v", v)
		}

		result = result[:start] + strValue + result[end+1:]
	}

	return result, nil
}

// getNestedValue retrieves a value from nested maps using dot notation
// Example: "tavern.env_vars.TOKEN" -> variables["tavern"]["env_vars"]["TOKEN"]
func getNestedValue(variables map[string]interface{}, path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = variables

	for _, key := range keys {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[key]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// applyTypeConversion applies type conversion if the string has a type marker
func applyTypeConversion(s string) (interface{}, error) {
	// Check for !int or !anyint marker
	if strings.HasPrefix(s, "<<INT>>") {
		value := strings.TrimPrefix(s, "<<INT>>")
		// If value is empty, it's a type matcher (!anyint), not a type converter
		// Keep the marker for validation
		if value == "" {
			return s, nil
		}
		return IntConverter(value)
	}

	// Check for !float or !anyfloat marker
	if strings.HasPrefix(s, "<<FLOAT>>") {
		value := strings.TrimPrefix(s, "<<FLOAT>>")
		// If value is empty, it's a type matcher (!anyfloat), not a type converter
		// Keep the marker for validation
		if value == "" {
			return s, nil
		}
		return FloatConverter(value)
	}

	// Check for !str or !anystr marker
	if strings.HasPrefix(s, "<<STR>>") {
		value := strings.TrimPrefix(s, "<<STR>>")
		// If value is empty, it's a type matcher (!anystr), not a type converter
		// Keep the marker for validation
		if value == "" {
			return s, nil
		}
		// For string type, just return the formatted value as string
		return value, nil
	}

	// No type conversion needed
	return s, nil
}

// formatMap recursively formats a map
func formatMap(m map[string]interface{}, variables map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, val := range m {
		formatted, err := FormatKeys(val, variables)
		if err != nil {
			return nil, err
		}
		result[key] = formatted
	}
	return result, nil
}

// formatSlice recursively formats a slice
func formatSlice(s []interface{}, variables map[string]interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(s))
	for i, val := range s {
		formatted, err := FormatKeys(val, variables)
		if err != nil {
			return nil, err
		}
		result[i] = formatted
	}
	return result, nil
}

// RecurseAccessKey accesses a nested key in a map or slice using dot notation.
// It splits the key by dots and recursively traverses the data structure.
//
// Examples:
//
//	Accessing nested dictionary keys:
//	  data := map[string]interface{}{
//	    "user": map[string]interface{}{
//	      "profile": map[string]interface{}{
//	        "name": "John",
//	      },
//	    },
//	  }
//	  result, _ := RecurseAccessKey(data, "user.profile.name")
//	  // result == "John"
//
//	Accessing array elements by index:
//	  data := map[string]interface{}{
//	    "items": []interface{}{"a", "b", "c"},
//	  }
//	  result, _ := RecurseAccessKey(data, "items.1")
//	  // result == "b"
//
//	Combining nested keys and array access:
//	  data := map[string]interface{}{
//	    "users": []interface{}{
//	      map[string]interface{}{"id": 1, "name": "Alice"},
//	      map[string]interface{}{"id": 2, "name": "Bob"},
//	    },
//	  }
//	  result, _ := RecurseAccessKey(data, "users.0.name")
//	  // result == "Alice"
func RecurseAccessKey(data interface{}, key string) (interface{}, error) {
	keys := strings.Split(key, ".")
	return recurseAccessKeyList(data, keys)
}

func recurseAccessKeyList(current interface{}, keys []string) (interface{}, error) {
	if len(keys) == 0 {
		return current, nil
	}

	currentKey := keys[0]
	remainingKeys := keys[1:]

	switch v := current.(type) {
	case map[string]interface{}:
		next, ok := v[currentKey]
		if !ok {
			return nil, fmt.Errorf("key not found: %s", currentKey)
		}
		return recurseAccessKeyList(next, remainingKeys)

	case []interface{}:
		idx, err := strconv.Atoi(currentKey)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", currentKey)
		}
		if idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("index out of range: %d (length: %d)", idx, len(v))
		}
		return recurseAccessKeyList(v[idx], remainingKeys)

	default:
		return nil, fmt.Errorf("cannot access key %s in type %T", currentKey, current)
	}
}

// DeepMerge recursively merges two maps
func DeepMerge(dst, src map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy dst
	for k, v := range dst {
		result[k] = v
	}

	// Merge src
	for k, v := range src {
		if existingVal, ok := result[k]; ok {
			// If both are maps, merge recursively
			if existingMap, ok1 := existingVal.(map[string]interface{}); ok1 {
				if newMap, ok2 := v.(map[string]interface{}); ok2 {
					result[k] = DeepMerge(existingMap, newMap)
					continue
				}
			}
		}
		result[k] = v
	}

	return result
}

// DeepEqual compares two values deeply
func DeepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// ToMap converts an interface{} to map[string]interface{} if possible
func ToMap(val interface{}) (map[string]interface{}, bool) {
	m, ok := val.(map[string]interface{})
	return m, ok
}

// ToSlice converts an interface{} to []interface{} if possible
func ToSlice(val interface{}) ([]interface{}, bool) {
	s, ok := val.([]interface{})
	return s, ok
}
