package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed tests.schema.json
var testSchema string

// Validator validates test specifications against JSON Schema
type Validator struct {
	schema *gojsonschema.Schema
}

// NewValidator creates a new schema validator
func NewValidator() (*Validator, error) {
	schemaLoader := gojsonschema.NewStringLoader(testSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to load schema: %w", err)
	}

	return &Validator{schema: schema}, nil
}

// Validate validates a test specification
func (v *Validator) Validate(test *TestSpec) error {
	// Convert test to JSON for validation
	testJSON, err := json.Marshal(test)
	if err != nil {
		return fmt.Errorf("failed to marshal test: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(testJSON)
	result, err := v.schema.Validate(documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		errorMsg := "validation failed:\n"
		for _, err := range result.Errors() {
			errorMsg += fmt.Sprintf("  - %s: %s\n", err.Field(), err.Description())
		}
		return fmt.Errorf("%s", errorMsg)
	}

	// Custom validation: Check stage name uniqueness
	stageNames := make(map[string]bool)
	for i, stage := range test.Stages {
		if stageNames[stage.Name] {
			return fmt.Errorf("validation failed:\n  - stages[%d].name: stage name '%s' must be unique", i, stage.Name)
		}
		stageNames[stage.Name] = true
	}

	return nil
}

// ValidateTest is a convenience function to validate a test
func ValidateTest(test *TestSpec) error {
	validator, err := NewValidator()
	if err != nil {
		return err
	}
	return validator.Validate(test)
}
