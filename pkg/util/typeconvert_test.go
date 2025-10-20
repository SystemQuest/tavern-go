package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeConvert_Int(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		variables map[string]interface{}
		expected  interface{}
		wantErr   bool
	}{
		{
			name:      "convert string to int",
			input:     "<<INT>>42",
			variables: map[string]interface{}{},
			expected:  42,
			wantErr:   false,
		},
		{
			name:      "convert formatted string to int",
			input:     "<<INT>>{value}",
			variables: map[string]interface{}{"value": "10"},
			expected:  10,
			wantErr:   false,
		},
		{
			name:      "convert with numeric variable",
			input:     "<<INT>>{value}",
			variables: map[string]interface{}{"value": 20},
			expected:  20,
			wantErr:   false,
		},
		{
			name:      "invalid int conversion",
			input:     "<<INT>>abc",
			variables: map[string]interface{}{},
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatKeys(tt.input, tt.variables)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTypeConvert_Float(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		variables map[string]interface{}
		expected  interface{}
		wantErr   bool
	}{
		{
			name:      "convert string to float",
			input:     "<<FLOAT>>3.14",
			variables: map[string]interface{}{},
			expected:  3.14,
			wantErr:   false,
		},
		{
			name:      "convert formatted string to float",
			input:     "<<FLOAT>>{value}",
			variables: map[string]interface{}{"value": "2.5"},
			expected:  2.5,
			wantErr:   false,
		},
		{
			name:      "convert int string to float",
			input:     "<<FLOAT>>42",
			variables: map[string]interface{}{},
			expected:  42.0,
			wantErr:   false,
		},
		{
			name:      "invalid float conversion",
			input:     "<<FLOAT>>xyz",
			variables: map[string]interface{}{},
			expected:  nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatKeys(tt.input, tt.variables)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTypeConvert_InMap(t *testing.T) {
	input := map[string]interface{}{
		"number":  "<<INT>>{value}",
		"price":   "<<FLOAT>>{cost}",
		"regular": "{name}",
	}

	variables := map[string]interface{}{
		"value": "10",
		"cost":  "19.99",
		"name":  "test",
	}

	result, err := FormatKeys(input, variables)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, 10, resultMap["number"])
	assert.Equal(t, 19.99, resultMap["price"])
	assert.Equal(t, "test", resultMap["regular"])
}

func TestTypeConvert_InArray(t *testing.T) {
	input := []interface{}{
		"<<INT>>5",
		"<<FLOAT>>3.14",
		"regular string",
	}

	variables := map[string]interface{}{}

	result, err := FormatKeys(input, variables)
	require.NoError(t, err)

	resultSlice, ok := result.([]interface{})
	require.True(t, ok)
	require.Len(t, resultSlice, 3)

	assert.Equal(t, 5, resultSlice[0])
	assert.Equal(t, 3.14, resultSlice[1])
	assert.Equal(t, "regular string", resultSlice[2])
}
