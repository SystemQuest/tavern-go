package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateMetaField tests schema validation for meta field in request
// Aligned with tavern-py commit 1dcffc6: meta field should accept array of unique strings
func TestValidateMetaField(t *testing.T) {
	tests := []struct {
		name      string
		testSpec  *TestSpec
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid meta field",
			testSpec: &TestSpec{
				TestName: "Test with meta",
				Stages: []Stage{
					{
						Name: "stage1",
						Request: &RequestSpec{
							URL:  "http://example.com",
							Meta: []string{"clear_session_cookies"},
						},
						Response: &ResponseSpec{
							StatusCode: &StatusCode{Single: 200},
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "multiple meta directives",
			testSpec: &TestSpec{
				TestName: "Test with multiple meta",
				Stages: []Stage{
					{
						Name: "stage1",
						Request: &RequestSpec{
							URL:  "http://example.com",
							Meta: []string{"clear_session_cookies", "some_other_directive"},
						},
						Response: &ResponseSpec{
							StatusCode: &StatusCode{Single: 200},
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "empty meta array",
			testSpec: &TestSpec{
				TestName: "Test with empty meta",
				Stages: []Stage{
					{
						Name: "stage1",
						Request: &RequestSpec{
							URL:  "http://example.com",
							Meta: []string{},
						},
						Response: &ResponseSpec{
							StatusCode: &StatusCode{Single: 200},
						},
					},
				},
			},
			shouldErr: false,
		},
		{
			name: "no meta field",
			testSpec: &TestSpec{
				TestName: "Test without meta",
				Stages: []Stage{
					{
						Name: "stage1",
						Request: &RequestSpec{
							URL: "http://example.com",
						},
						Response: &ResponseSpec{
							StatusCode: &StatusCode{Single: 200},
						},
					},
				},
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTest(tt.testSpec)
			if tt.shouldErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
