package schema

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// testSchema defines the JSON Schema for test specifications
const testSchema = `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["test_name", "stages"],
  "properties": {
    "test_name": {
      "type": "string",
      "description": "Name of the test"
    },
    "includes": {
      "type": "array",
      "description": "Include blocks with variables",
      "items": {
        "type": "object",
        "required": ["name", "description"],
        "properties": {
          "name": {
            "type": "string"
          },
          "description": {
            "type": "string"
          },
          "variables": {
            "type": "object"
          }
        }
      }
    },
    "stages": {
      "type": "array",
      "description": "Test stages",
      "minItems": 1,
      "items": {
        "type": "object",
        "required": ["name", "request", "response"],
        "properties": {
          "name": {
            "type": "string",
            "description": "Stage name"
          },
          "request": {
            "type": "object",
            "required": ["url"],
            "properties": {
              "url": {
                "type": "string"
              },
              "method": {
                "type": "string",
                "enum": ["GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"]
              },
              "headers": {
                "type": "object"
              },
              "json": {},
              "data": {},
              "params": {
                "type": "object"
              },
              "auth": {
                "type": "object"
              },
              "files": {
                "type": "object"
              },
              "cookies": {
                "type": "object"
              },
              "verify": {
                "type": "boolean",
                "description": "Whether to verify SSL certificates (default: true)"
              }
            }
          },
          "response": {
            "type": "object",
            "properties": {
              "status_code": {
                "type": "integer"
              },
              "headers": {
                "type": "object"
              },
              "body": {},
              "cookies": {
                "type": "array",
                "description": "Expected cookie names to verify in response",
                "uniqueItems": true,
                "items": {
                  "type": "string"
                }
              },
              "save": {
                "type": "object",
                "properties": {
                  "body": {
                    "type": "object"
                  },
                  "headers": {
                    "type": "object"
                  },
                  "redirect_query_params": {
                    "type": "object"
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

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
