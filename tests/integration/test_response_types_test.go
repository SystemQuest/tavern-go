package integration

import (
"testing"

"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"github.com/systemquest/tavern-go/pkg/core"
"github.com/systemquest/tavern-go/pkg/schema"
"github.com/systemquest/tavern-go/tests/fixtures"
)

// =============================================================================
// Test Response Types (aligned with tavern-py test_response_types.tavern.yaml)
// =============================================================================

// TestIntegration_ListResponse tests handling of list/array responses
// Aligned with: test_response_types.tavern.yaml - "Match line notation" stage
func TestIntegration_ListResponse(t *testing.T) {
	// Create a mock server that returns a list response
	server := fixtures.NewMockServer(
fixtures.SimpleJSONResponse(200, []interface{}{"a", "b", "c"}),
)
	defer server.Close()

	// Test: Validate exact array match (line notation and JSON notation are equivalent)
	testSpec := &schema.TestSpec{
		TestName: "List response validation",
		Stages: []schema.Stage{
			{
				Name: "Match array",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/fake_list",
					Method: "GET",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body:       []interface{}{"a", "b", "c"},
				},
			},
		},
	}

	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err, "Should validate exact array match")
	assert.Equal(t, 1, server.GetRequestCount())
}

// TestIntegration_ListResponsePartial tests partial list validation
// Extended test: validates tavern-py behavior of partial array matching
func TestIntegration_ListResponsePartial(t *testing.T) {
	// Create a mock server that returns a longer list
	server := fixtures.NewMockServer(
fixtures.SimpleJSONResponse(200, []interface{}{"a", "b", "c", "d", "e"}),
)
	defer server.Close()

	// Test partial validation (only check first 3 elements)
	testSpec := &schema.TestSpec{
		TestName: "Partial list response validation",
		Stages: []schema.Stage{
			{
				Name: "Match partial array",
				Request: &schema.RequestSpec{
					URL:    server.URL + "/fake_list",
					Method: "GET",
				},
				Response: &schema.ResponseSpec{
					StatusCode: 200,
					Body:       []interface{}{"a", "b", "c"}, // Only validate first 3
				},
			},
		},
	}

	runner, err := core.NewRunner(&core.Config{})
	require.NoError(t, err)

	err = runner.RunTest(testSpec)
	assert.NoError(t, err, "Should allow partial array validation")
	assert.Equal(t, 1, server.GetRequestCount())
}
