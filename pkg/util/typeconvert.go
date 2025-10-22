package util

import (
	"fmt"
	"strconv"
)

// TypeConverter represents a type conversion function
type TypeConverter func(string) (interface{}, error)

// TypeConvertToken wraps a value with a type converter
// Similar to tavern-py's TypeConvertToken class
type TypeConvertToken struct {
	Converter TypeConverter
	Value     interface{}
}

// NewTypeConvertToken creates a new type convert token
func NewTypeConvertToken(converter TypeConverter, value interface{}) *TypeConvertToken {
	return &TypeConvertToken{
		Converter: converter,
		Value:     value,
	}
}

// IntConverter converts a string to int
func IntConverter(s string) (interface{}, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to int: %w", err)
	}
	return val, nil
}

// FloatConverter converts a string to float64
func FloatConverter(s string) (interface{}, error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to float: %w", err)
	}
	return val, nil
}

// BoolConverter converts a string to bool
// Accepts: "1", "t", "T", "true", "TRUE", "True", "0", "f", "F", "false", "FALSE", "False"
// Aligned with tavern-py commit 963bdf6
func BoolConverter(s string) (interface{}, error) {
	val, err := strconv.ParseBool(s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to bool: %w", err)
	}
	return val, nil
}

// IsTypeConvertToken checks if a value is a TypeConvertToken
func IsTypeConvertToken(val interface{}) bool {
	_, ok := val.(*TypeConvertToken)
	return ok
}
