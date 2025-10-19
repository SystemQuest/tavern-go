package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatKeys(t *testing.T) {
	variables := map[string]interface{}{
		"name":   "John",
		"age":    30,
		"active": true,
	}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "simple string",
			input:    "Hello {name}",
			expected: "Hello John",
			wantErr:  false,
		},
		{
			name:     "multiple variables",
			input:    "{name} is {age} years old",
			expected: "John is 30 years old",
			wantErr:  false,
		},
		{
			name: "map with variables",
			input: map[string]interface{}{
				"greeting": "Hello {name}",
				"age":      "{age}",
			},
			expected: map[string]interface{}{
				"greeting": "Hello John",
				"age":      "30",
			},
			wantErr: false,
		},
		{
			name:     "missing variable",
			input:    "Hello {missing}",
			expected: nil,
			wantErr:  true,
		},
		{
			name: "array with variables",
			input: []interface{}{
				"{name}",
				"{age}",
				"static",
			},
			expected: []interface{}{
				"John",
				"30",
				"static",
			},
			wantErr: false,
		},
		{
			name: "array with repeated variables",
			input: []interface{}{
				"{name}",
				"{name}",
				"{name}",
			},
			expected: []interface{}{
				"John",
				"John",
				"John",
			},
			wantErr: false,
		},
		{
			name: "nested array in map",
			input: map[string]interface{}{
				"users": []interface{}{
					"{name}",
					"Alice",
				},
				"ages": []interface{}{
					"{age}",
					25,
				},
			},
			expected: map[string]interface{}{
				"users": []interface{}{
					"John",
					"Alice",
				},
				"ages": []interface{}{
					"30",
					25,
				},
			},
			wantErr: false,
		},
		{
			name: "array with maps containing variables",
			input: []interface{}{
				map[string]interface{}{
					"name": "{name}",
					"age":  "{age}",
				},
				map[string]interface{}{
					"greeting": "Hello {name}",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name": "John",
					"age":  "30",
				},
				map[string]interface{}{
					"greeting": "Hello John",
				},
			},
			wantErr: false,
		},
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name: "array with missing variable",
			input: []interface{}{
				"{name}",
				"{missing}",
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatKeys(tt.input, variables)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRecurseAccessKey(t *testing.T) {
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "John",
			"profile": map[string]interface{}{
				"age": 30,
			},
		},
		"items": []interface{}{
			map[string]interface{}{"id": 1, "name": "item1"},
			map[string]interface{}{"id": 2, "name": "item2"},
		},
	}

	tests := []struct {
		name     string
		key      string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "simple key",
			key:      "user.name",
			expected: "John",
			wantErr:  false,
		},
		{
			name:     "nested key",
			key:      "user.profile.age",
			expected: 30,
			wantErr:  false,
		},
		{
			name:     "array access",
			key:      "items.0.name",
			expected: "item1",
			wantErr:  false,
		},
		{
			name:     "missing key",
			key:      "user.missing",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RecurseAccessKey(data, tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDeepMerge(t *testing.T) {
	dst := map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{
			"c": 2,
			"d": 3,
		},
	}

	src := map[string]interface{}{
		"b": map[string]interface{}{
			"d": 4,
			"e": 5,
		},
		"f": 6,
	}

	expected := map[string]interface{}{
		"a": 1,
		"b": map[string]interface{}{
			"c": 2,
			"d": 4,
			"e": 5,
		},
		"f": 6,
	}

	result := DeepMerge(dst, src)
	assert.Equal(t, expected, result)
}

// TestFormatKeys_NestedVariables tests nested variable access with dot notation
// Aligned with tavern-py commit 1b55d6e: supports {tavern.env_vars.VAR_NAME}
func TestFormatKeys_NestedVariables(t *testing.T) {
	variables := map[string]interface{}{
		"tavern": map[string]interface{}{
			"env_vars": map[string]interface{}{
				"TOKEN":     "secret123",
				"API_KEY":   "key456",
				"CI_COMMIT": "abc789",
			},
		},
		"user": map[string]interface{}{
			"name": "John",
			"profile": map[string]interface{}{
				"email": "john@example.com",
			},
		},
	}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "simple nested access",
			input:    "Token: {tavern.env_vars.TOKEN}",
			expected: "Token: secret123",
			wantErr:  false,
		},
		{
			name:     "multiple nested variables",
			input:    "API Key: {tavern.env_vars.API_KEY}, Commit: {tavern.env_vars.CI_COMMIT}",
			expected: "API Key: key456, Commit: abc789",
			wantErr:  false,
		},
		{
			name:     "nested user access",
			input:    "Email: {user.profile.email}",
			expected: "Email: john@example.com",
			wantErr:  false,
		},
		{
			name: "nested in map",
			input: map[string]interface{}{
				"auth":  "Bearer {tavern.env_vars.TOKEN}",
				"email": "{user.profile.email}",
			},
			expected: map[string]interface{}{
				"auth":  "Bearer secret123",
				"email": "john@example.com",
			},
			wantErr: false,
		},
		{
			name:     "missing nested variable",
			input:    "Value: {tavern.env_vars.MISSING}",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid nested path",
			input:    "Value: {tavern.missing.TOKEN}",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "mixed flat and nested",
			input:    "{user.name} has token {tavern.env_vars.TOKEN}",
			expected: "John has token secret123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatKeys(tt.input, variables)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
