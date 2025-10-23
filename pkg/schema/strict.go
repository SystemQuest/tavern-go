package schema

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements custom YAML unmarshaling for Strict
// It can handle bool, []string, or nil values
// Note: We store the raw value and defer validation to schema validation phase
// This allows _xfail: verify tests to test invalid strict values
func (s *Strict) UnmarshalYAML(node *yaml.Node) error {
	// Try to unmarshal as bool first
	var boolVal bool
	if err := node.Decode(&boolVal); err == nil {
		s.IsSet = true
		s.AsBool = boolVal
		s.IsList = false
		s.IsLegacy = false
		s.RawValue = boolVal
		return nil
	}

	// Try to unmarshal as []string
	var listVal []string
	if err := node.Decode(&listVal); err == nil {
		s.IsSet = true
		s.IsList = true
		s.AsList = listVal
		s.IsLegacy = false
		s.RawValue = listVal
		// Store raw value for later validation (don't validate here to support _xfail tests)
		return nil
	}

	// Store raw value for other types (string, map, etc.) for validation error
	var rawVal interface{}
	if err := node.Decode(&rawVal); err == nil {
		s.IsSet = true // Mark as set even for invalid types
		s.RawValue = rawVal
	}

	// Don't fail here - let schema validation handle it
	// This allows _xfail: verify tests to work
	return nil
} // Validate checks if the Strict value is valid
// This should be called during schema validation
func (s *Strict) Validate() error {
	if s == nil || !s.IsSet || s.IsLegacy {
		return nil
	}

	if s.IsList {
		// Validate that the list only contains valid response parts
		validParts := map[string]bool{
			"body":                  true,
			"headers":               true,
			"redirect_query_params": true,
		}

		for _, part := range s.AsList {
			if !validParts[part] {
				return fmt.Errorf("invalid strict value: %s (must be one of: body, headers, redirect_query_params)", part)
			}
		}
		return nil
	}

	// For bool values, no validation needed
	if _, ok := s.RawValue.(bool); ok {
		return nil
	}

	// For list values that were already validated
	if _, ok := s.RawValue.([]string); ok {
		return nil
	}

	// Invalid type
	return fmt.Errorf("strict must be either a boolean or a list of strings")
}

// MarshalYAML implements custom YAML marshaling for Strict
func (s *Strict) MarshalYAML() (interface{}, error) {
	if s == nil || s.IsLegacy || !s.IsSet {
		return nil, nil
	}

	if s.IsList {
		return s.AsList, nil
	}

	return s.AsBool, nil
}

// ShouldCheckStrictly returns whether strict checking should be applied for a given response part
// blockName is one of: "body", "headers", "redirect_query_params"
func (s *Strict) ShouldCheckStrictly(blockName string) bool {
	if s == nil || !s.IsSet || s.IsLegacy {
		// Legacy behavior: strict for nested keys, lenient for top-level keys
		return false
	}

	if s.IsList {
		// Check if this block is in the list
		for _, part := range s.AsList {
			if part == blockName {
				return true
			}
		}
		return false
	}

	// Boolean mode
	return s.AsBool
}

// NewStrictFromBool creates a Strict config from a boolean value
func NewStrictFromBool(value bool) *Strict {
	return &Strict{
		IsSet:    true,
		AsBool:   value,
		IsList:   false,
		IsLegacy: false,
	}
}

// NewStrictFromList creates a Strict config from a list of response parts
func NewStrictFromList(parts []string) *Strict {
	return &Strict{
		IsSet:    true,
		IsList:   true,
		AsList:   parts,
		IsLegacy: false,
	}
}

// NewStrictLegacy creates a Strict config with legacy behavior
func NewStrictLegacy() *Strict {
	return &Strict{
		IsSet:    false,
		IsLegacy: true,
	}
}
