package util

import "fmt"

// TavernError is the base error type for all Tavern errors
type TavernError struct {
	Message string
	Cause   error
}

func (e *TavernError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *TavernError) Unwrap() error {
	return e.Cause
}

// NewTavernError creates a new TavernError
func NewTavernError(message string, cause error) *TavernError {
	return &TavernError{
		Message: message,
		Cause:   cause,
	}
}

// BadSchemaError represents a schema validation error
type BadSchemaError struct {
	TavernError
}

func NewBadSchemaError(message string, cause error) *BadSchemaError {
	return &BadSchemaError{
		TavernError: TavernError{
			Message: message,
			Cause:   cause,
		},
	}
}

// TestFailError represents a test failure
type TestFailError struct {
	TavernError
	Errors []string
}

func NewTestFailError(message string, errors []string) *TestFailError {
	return &TestFailError{
		TavernError: TavernError{
			Message: message,
		},
		Errors: errors,
	}
}

func (e *TestFailError) Error() string {
	if len(e.Errors) == 0 {
		return e.Message
	}
	return fmt.Sprintf("%s:\n- %s", e.Message, joinErrors(e.Errors))
}

func joinErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	result := errors[0]
	for i := 1; i < len(errors); i++ {
		result += "\n- " + errors[i]
	}
	return result
}

// UnexpectedKeysError represents unexpected keys in request specification
type UnexpectedKeysError struct {
	TavernError
	Keys []string
}

func NewUnexpectedKeysError(keys []string) *UnexpectedKeysError {
	return &UnexpectedKeysError{
		TavernError: TavernError{
			Message: fmt.Sprintf("unexpected keys: %v", keys),
		},
		Keys: keys,
	}
}

// MissingFormatError represents a missing variable in format string
type MissingFormatError struct {
	TavernError
	Variable string
}

func NewMissingFormatError(variable string) *MissingFormatError {
	return &MissingFormatError{
		TavernError: TavernError{
			Message: fmt.Sprintf("missing variable in format: %s", variable),
		},
		Variable: variable,
	}
}
