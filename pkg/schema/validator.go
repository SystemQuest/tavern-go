package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

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
	// Validate strict field at test level (aligned with tavern-py commit 3838566)
	if test.Strict != nil {
		if err := test.Strict.Validate(); err != nil {
			return fmt.Errorf("validation failed:\n  - strict: %s", err)
		}
	}

	// Validate strict field at stage level
	for i, stage := range test.Stages {
		if stage.Response != nil && stage.Response.Strict != nil {
			if err := stage.Response.Strict.Validate(); err != nil {
				return fmt.Errorf("validation failed:\n  - stages[%d].response.strict: %s", i, err)
			}
		}
	}

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

	// Custom validation: Check !approx is not used in requests
	// Aligned with tavern-py commit 61065bd: Stop being able to use 'approx' tag in requests
	for i, stage := range test.Stages {
		if stage.Request != nil {
			if err := v.checkApproxInRequest(stage.Request, fmt.Sprintf("stages[%d].request", i)); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkApproxInRequest checks if !approx marker is used in request data
// !approx should only be used in response validation, not in requests
// Aligned with tavern-py commit 61065bd
func (v *Validator) checkApproxInRequest(request *RequestSpec, path string) error {
	if request.JSON != nil {
		if hasApprox(request.JSON) {
			return fmt.Errorf("validation failed:\n  - %s.json: Cannot use '!approx' in request data. !approx is only valid in response.body or mqtt_response.json", path)
		}
	}
	if request.Data != nil {
		if hasApprox(request.Data) {
			return fmt.Errorf("validation failed:\n  - %s.data: Cannot use '!approx' in request data. !approx is only valid in response.body or mqtt_response.json", path)
		}
	}
	return nil
}

// hasApprox recursively checks if a value contains the !approx marker (<<APPROX>>)
func hasApprox(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return strings.Contains(val, "<<APPROX>>")
	case map[string]interface{}:
		for _, value := range val {
			if hasApprox(value) {
				return true
			}
		}
	case []interface{}:
		for _, item := range val {
			if hasApprox(item) {
				return true
			}
		}
	}
	return false
}

// ValidateTest is a convenience function to validate a test
func ValidateTest(test *TestSpec) error {
	validator, err := NewValidator()
	if err != nil {
		return err
	}
	return validator.Validate(test)
}
