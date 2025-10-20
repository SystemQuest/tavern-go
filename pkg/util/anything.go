package util

// AnythingType represents a special marker that matches any value in response validation
type AnythingType struct{}

// Anything is a singleton instance used to mark fields that accept any value
var Anything = &AnythingType{}

// String returns a string representation
func (a *AnythingType) String() string {
	return "!anything"
}

// IsAnything checks if a value is the Anything marker
func IsAnything(val interface{}) bool {
	_, ok := val.(*AnythingType)
	return ok
}
