package response

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/systemquest/tavern-go/pkg/schema"
)

// TestValidator_AnythingMarkerDebugLog tests that debug logs are generated for !anything matches
func TestValidator_AnythingMarkerDebugLog(t *testing.T) {
	spec := schema.ResponseSpec{
		StatusCode: &schema.StatusCode{Single: 200},
		Body: map[string]interface{}{
			"user.id":   "<<ANYTHING>>",
			"user.name": "John",
			"items": []interface{}{
				"<<ANYTHING>>",
				"fixed",
			},
		},
	}

	validator := NewRestValidator("test", spec, &Config{Variables: map[string]interface{}{}})

	// Capture debug logs
	var buf bytes.Buffer
	validator.logger.SetOutput(&buf)
	validator.logger.SetLevel(logrus.DebugLevel)

	body := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   "uuid-1234-5678",
			"name": "John",
		},
		"items": []interface{}{
			"dynamic-value",
			"fixed",
		},
	}

	bodyBytes, _ := json.Marshal(body)
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer(bodyBytes)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	_, err := validator.Verify(resp)
	assert.NoError(t, err)

	// Check that debug logs contain !anything messages
	logOutput := buf.String()
	assert.Contains(t, logOutput, "matches !anything", "Debug log should mention !anything")
	assert.Contains(t, logOutput, "user.id", "Debug log should mention the key")
	assert.Contains(t, logOutput, "uuid-1234-5678", "Debug log should show actual value")
}
