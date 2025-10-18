package util

import (
	"github.com/sirupsen/logrus"
)

// CheckExpectedKeys checks that actual keys are a subset of expected keys
// Returns an error if there are unexpected keys
func CheckExpectedKeys(expected []string, actual map[string]interface{}) error {
	expectedSet := make(map[string]bool)
	for _, key := range expected {
		expectedSet[key] = true
	}

	var unexpected []string
	for key := range actual {
		if !expectedSet[key] {
			unexpected = append(unexpected, key)
		}
	}

	if len(unexpected) > 0 {
		logrus.WithFields(logrus.Fields{
			"expected":   expected,
			"actual":     getKeys(actual),
			"unexpected": unexpected,
		}).Error("Unexpected keys found")
		return NewUnexpectedKeysError(unexpected)
	}

	return nil
}

// getKeys returns all keys from a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
