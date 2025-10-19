package extension

import (
	"strings"
	"testing"
)

func TestConvertToExtSpec_Valid(t *testing.T) {
	extMap := map[string]interface{}{
		"function": "test:saver",
		"extra_kwargs": map[string]interface{}{
			"pattern": ".*",
			"count":   3,
		},
	}

	ext, err := ConvertToExtSpec(extMap)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if ext.Function != "test:saver" {
		t.Errorf("Expected function='test:saver', got: %s", ext.Function)
	}

	if ext.ExtraKwargs == nil {
		t.Fatal("Expected non-nil extra_kwargs")
	}

	if ext.ExtraKwargs["pattern"] != ".*" {
		t.Errorf("Expected pattern='.*', got: %v", ext.ExtraKwargs["pattern"])
	}

	if ext.ExtraKwargs["count"] != 3 {
		t.Errorf("Expected count=3, got: %v", ext.ExtraKwargs["count"])
	}
}

func TestConvertToExtSpec_MinimalValid(t *testing.T) {
	// Only function is required, extra_kwargs is optional
	extMap := map[string]interface{}{
		"function": "test:minimal",
	}

	ext, err := ConvertToExtSpec(extMap)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if ext.Function != "test:minimal" {
		t.Errorf("Expected function='test:minimal', got: %s", ext.Function)
	}

	if ext.ExtraKwargs != nil {
		t.Errorf("Expected nil extra_kwargs, got: %v", ext.ExtraKwargs)
	}
}

func TestConvertToExtSpec_InvalidType(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		errorMsg string
	}{
		{
			name:     "string instead of map",
			input:    "not a map",
			errorMsg: "$ext must be a map",
		},
		{
			name:     "array instead of map",
			input:    []interface{}{"item1", "item2"},
			errorMsg: "$ext must be a map",
		},
		{
			name:     "nil",
			input:    nil,
			errorMsg: "$ext must be a map",
		},
		{
			name:     "integer",
			input:    123,
			errorMsg: "$ext must be a map",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ConvertToExtSpec(tc.input)

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error containing '%s', got: %v", tc.errorMsg, err)
			}
		})
	}
}

func TestConvertToExtSpec_MissingFunction(t *testing.T) {
	extMap := map[string]interface{}{
		"extra_kwargs": map[string]interface{}{
			"param": "value",
		},
		// "function" field is missing
	}

	_, err := ConvertToExtSpec(extMap)

	if err == nil {
		t.Fatal("Expected error for missing function")
	}

	if !strings.Contains(err.Error(), "$ext.function must be a string") {
		t.Errorf("Expected '$ext.function must be a string' error, got: %v", err)
	}
}

func TestConvertToExtSpec_InvalidFunctionType(t *testing.T) {
	testCases := []struct {
		name         string
		functionVal  interface{}
		errorPattern string
	}{
		{
			name:         "function is integer",
			functionVal:  123,
			errorPattern: "$ext.function must be a string",
		},
		{
			name:         "function is map",
			functionVal:  map[string]interface{}{"key": "value"},
			errorPattern: "$ext.function must be a string",
		},
		{
			name:         "function is array",
			functionVal:  []string{"a", "b"},
			errorPattern: "$ext.function must be a string",
		},
		{
			name:         "function is nil",
			functionVal:  nil,
			errorPattern: "$ext.function must be a string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			extMap := map[string]interface{}{
				"function": tc.functionVal,
			}

			_, err := ConvertToExtSpec(extMap)

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tc.errorPattern) {
				t.Errorf("Expected error containing '%s', got: %v", tc.errorPattern, err)
			}
		})
	}
}

func TestConvertToExtSpec_ExtraKwargsWrongType(t *testing.T) {
	// extra_kwargs should be map[string]interface{}, but we provide string
	// Should not error, just ignore it (since it's optional)
	extMap := map[string]interface{}{
		"function":     "test:saver",
		"extra_kwargs": "not a map",
	}

	ext, err := ConvertToExtSpec(extMap)

	if err != nil {
		t.Fatalf("Expected no error (extra_kwargs is optional), got: %v", err)
	}

	if ext.Function != "test:saver" {
		t.Errorf("Expected function='test:saver', got: %s", ext.Function)
	}

	// Should be nil since type assertion failed
	if ext.ExtraKwargs != nil {
		t.Errorf("Expected nil extra_kwargs (wrong type), got: %v", ext.ExtraKwargs)
	}
}

func TestConvertToExtSpec_EmptyFunction(t *testing.T) {
	// Empty string is technically a valid string, conversion should succeed
	// but ExecuteSaver will reject it later
	extMap := map[string]interface{}{
		"function": "",
	}

	ext, err := ConvertToExtSpec(extMap)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if ext.Function != "" {
		t.Errorf("Expected empty function, got: %s", ext.Function)
	}
}

func TestConvertToExtSpec_ExtraFields(t *testing.T) {
	// Should ignore extra fields that are not part of ExtSpec
	extMap := map[string]interface{}{
		"function":     "test:saver",
		"extra_kwargs": map[string]interface{}{"key": "value"},
		"unknown":      "field",
		"another":      123,
	}

	ext, err := ConvertToExtSpec(extMap)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if ext.Function != "test:saver" {
		t.Errorf("Expected function='test:saver', got: %s", ext.Function)
	}

	if ext.ExtraKwargs == nil {
		t.Fatal("Expected non-nil extra_kwargs")
	}

	if ext.ExtraKwargs["key"] != "value" {
		t.Errorf("Expected key='value', got: %v", ext.ExtraKwargs["key"])
	}
}
